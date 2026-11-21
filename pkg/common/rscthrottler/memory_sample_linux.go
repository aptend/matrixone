// Copyright 2026 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux

package rscthrottler

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
)

func readCgroupMemory(sample *memorySample) error {
	mode := cgroups.Mode()
	if mode == cgroups.Unavailable {
		if _, statErr := os.Stat("/sys/fs/cgroup"); errors.Is(statErr, os.ErrNotExist) {
			return errNoMemoryController
		}
		return moerr.NewInternalErrorNoCtx("cannot determine cgroup mount state")
	}
	if mode == cgroups.Unified {
		_, group, e := cgroups.ParseCgroupFileUnified("/proc/self/cgroup")
		if e != nil {
			return e
		}
		if e = cgroup2.VerifyGroupPath(group); e != nil {
			return e
		}
		mount, e := readMemoryCgroupMount("/sys/fs/cgroup", "cgroup2", "")
		if e != nil {
			return e
		}
		path, e := resolveMemoryCgroupPath(group, mount)
		if e != nil {
			return e
		}
		return readUnifiedMemoryHierarchy(sample, path, mount.mountpoint)
	}
	groups, e := cgroup1.ParseCgroupFile("/proc/self/cgroup")
	if e != nil {
		return e
	}
	subsystems, e := cgroup1.SingleSubsystem(cgroup1.Default, cgroup1.Memory)()
	if e != nil {
		return legacyMemoryError(e)
	}
	mountpoint, e := legacyMemoryControllerPath(subsystems)
	if e != nil {
		return e
	}
	mount, e := readMemoryCgroupMount(mountpoint, "cgroup", "memory")
	if e != nil {
		return e
	}
	sample.limit, sample.current, sample.inactiveFile, sample.cgroupHeadroom, e = readLegacyMemory(groups[string(cgroup1.Memory)], mount, cgroup1.PidPath(os.Getpid()), func() ([]cgroup1.Subsystem, error) { return subsystems, nil })
	return e
}

// readLegacyMemory retains the existing controller and Stat implementation.
// Selected metadata and the hierarchy are explicit only at this private boundary
// so tests can exercise a real temporary controller without mounting cgroupfs.
func readLegacyMemory(group string, mount memoryCgroupMount, path cgroup1.Path, hierarchy cgroup1.Hierarchy) (limit, current, inactive int64, headroom *int64, err error) {
	subsystems, e := hierarchy()
	if e != nil {
		return 0, 0, 0, nil, legacyMemoryError(e)
	}
	mountpoint, e := legacyMemoryControllerPath(subsystems)
	if e != nil {
		return 0, 0, 0, nil, e
	}
	if mount.mountpoint != mountpoint {
		return 0, 0, 0, nil, moerr.NewInternalErrorNoCtx("memory cgroup mount does not match selected controller")
	}
	directory, e := resolveMemoryCgroupPath(group, mount)
	if e != nil {
		return 0, 0, 0, nil, e
	}
	relative, e := filepath.Rel(mountpoint, directory)
	if e != nil {
		return 0, 0, 0, nil, e
	}
	memoryPath := "/"
	if relative != "." {
		memoryPath += relative
	}
	// The dependency mishandles private-root equality and escaped roots.
	// Use the same decoded, validated mapping for every Memory path.
	legacyPath := path
	path = func(name cgroup1.Name) (string, error) {
		if name == cgroup1.Memory {
			return memoryPath, nil
		}
		return legacyPath(name)
	}
	cg, e := cgroup1.Load(path, cgroup1.WithHiearchy(func() ([]cgroup1.Subsystem, error) { return subsystems, nil }))
	if e != nil {
		return 0, 0, 0, nil, legacyMemoryError(e)
	}
	stats, e := cg.Stat(cgroup1.IgnoreNotExist)
	if e != nil {
		return 0, 0, 0, nil, e
	}
	m := stats.GetMemory()
	if m == nil {
		return 0, 0, 0, nil, moerr.NewInternalErrorNoCtx("incomplete cgroup memory accounting")
	}
	if m.GetUsage() == nil || m.GetUsage().GetUsage() == 0 {
		return 0, 0, 0, nil, moerr.NewInternalErrorNoCtx("incomplete cgroup memory usage")
	}
	inactive = max(0, memoryInt(m.GetTotalInactiveFile())-memoryInt(m.GetTotalDirty()))
	inactive = max(0, inactive-memoryInt(m.GetTotalWriteback()))
	current = memoryInt(m.GetUsage().GetUsage())
	limit, headroom, err = readLegacyMemoryHeadroom(directory, mount, legacyLimit(m.GetHierarchicalMemoryLimit()), current, inactive)
	return
}

func legacyLimit(value uint64) int64 {
	// The v1 kernel encodes unlimited as LONG_MAX rounded down to a page.
	if value >= uint64(math.MaxInt64)-uint64(os.Getpagesize()-1) {
		return math.MaxInt64
	}
	return int64(value)
}

func readLegacyMemoryValue(dir, file string) (uint64, error) {
	data, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
}

func readLegacyMemoryHeadroom(leaf string, mount memoryCgroupMount, reportedLimit, leafCurrent, leafInactive int64) (int64, *int64, error) {
	limit, headroom := int64(math.MaxInt64), int64(math.MaxInt64)
	for dir := leaf; ; dir = filepath.Dir(dir) {
		value, err := readLegacyMemoryValue(dir, "memory.limit_in_bytes")
		if errors.Is(err, os.ErrNotExist) && dir == mount.mountpoint && mount.root == "/" {
			// The global cgroup root has no effective memory limit.
			break
		}
		if err != nil {
			return 0, nil, err
		}
		if ownLimit := legacyLimit(value); ownLimit != math.MaxInt64 {
			current, inactive := leafCurrent, leafInactive
			if dir != leaf {
				usage, err := readLegacyMemoryValue(dir, "memory.usage_in_bytes")
				if err != nil {
					return 0, nil, err
				}
				current = memoryInt(usage)
				stat, err := os.ReadFile(filepath.Join(dir, "memory.stat"))
				if err != nil {
					return 0, nil, err
				}
				inactive, err = reclaimableFileBytes(stat, "total_inactive_file", "total_dirty", "total_writeback")
				if err != nil {
					return 0, nil, err
				}
			}
			limit = min(limit, ownLimit)
			headroom = min(headroom, max(0, ownLimit-max(0, current-inactive)))
		}
		if dir == mount.mountpoint {
			break
		}
	}
	// A lower hierarchical limit can reveal a controller above the mount.
	// Its usage cannot be inferred from the leaf, so deny new commitments.
	if reportedLimit > 0 && reportedLimit < limit {
		return 0, nil, moerr.NewInternalErrorNoCtx("cgroup v1 memory limit is outside the visible hierarchy")
	}
	if headroom == math.MaxInt64 {
		return limit, nil, nil
	}
	return limit, &headroom, nil
}

func legacyMemoryControllerPath(subsystems []cgroup1.Subsystem) (string, error) {
	if len(subsystems) != 1 || subsystems[0].Name() != cgroup1.Memory {
		return "", moerr.NewInternalErrorNoCtx("invalid memory cgroup hierarchy")
	}
	controller, ok := subsystems[0].(interface{ Path(string) string })
	if !ok {
		return "", moerr.NewInternalErrorNoCtx("memory cgroup controller has no path")
	}
	return controller.Path("/"), nil
}

func legacyMemoryError(err error) error {
	if errors.Is(err, cgroup1.ErrMountPointNotExist) || errors.Is(err, cgroup1.ErrMemoryNotSupported) || errors.Is(err, cgroup1.ErrNoCgroupMountDestination) {
		return errNoMemoryController
	}
	return err
}

// path has already been resolved inside mountpoint. Ancestors above the visible
// mount can still constrain the process, but cannot be sampled here.
func readUnifiedMemoryHierarchy(sample *memorySample, path, mountpoint string) error {
	limit, headroom := int64(math.MaxInt64), int64(math.MaxInt64)
	var observed bool
	for dir := path; ; dir = filepath.Dir(dir) {
		maxBytes, current, inactive, err := readUnifiedMemory(dir)
		if err == nil {
			observed = true
			if dir == path {
				sample.current, sample.inactiveFile = current, inactive
			}
			if maxBytes < math.MaxInt64 {
				limit = min(limit, maxBytes)
				headroom = min(headroom, max(0, maxBytes-max(0, current-inactive)))
			}
		} else if !errors.Is(err, errNoMemoryController) {
			return err
		}
		if dir == mountpoint {
			break
		}
	}
	if !observed {
		return errNoMemoryController
	}
	sample.limit = limit
	if headroom != math.MaxInt64 {
		sample.cgroupHeadroom = &headroom
	}
	return nil
}

// cgroup2.Manager.Stat masks missing memory.stat and scalar read failures as
// zero. Read the three required accounting files explicitly so an incomplete
// sample cannot release container pressure or erase a previously known limit.
func readUnifiedMemory(path string) (limit, current, inactive int64, err error) {
	names := []string{"memory.max", "memory.current", "memory.stat"}
	var data [3][]byte
	var readErrors [3]error
	for i, name := range names {
		data[i], readErrors[i] = os.ReadFile(filepath.Join(path, name))
	}
	// The global root can have memory.stat without controller limit or usage
	// files. A missing whole controller does not end the ancestor walk.
	if errors.Is(readErrors[0], os.ErrNotExist) && errors.Is(readErrors[1], os.ErrNotExist) &&
		(readErrors[2] == nil || errors.Is(readErrors[2], os.ErrNotExist)) {
		info, e := os.Stat(path)
		if e == nil && info.IsDir() {
			return 0, 0, 0, errNoMemoryController
		}
	}
	if err = errors.Join(readErrors[:]...); err != nil {
		return 0, 0, 0, err
	}
	value := strings.TrimSpace(string(data[0]))
	if value == "max" {
		limit = math.MaxInt64
	} else {
		n, e := strconv.ParseUint(value, 10, 64)
		if e != nil {
			return 0, 0, 0, e
		}
		limit = memoryInt(n)
	}
	n, e := strconv.ParseUint(strings.TrimSpace(string(data[1])), 10, 64)
	if e != nil {
		return 0, 0, 0, e
	}
	current = memoryInt(n)
	inactive, err = reclaimableFileBytes(data[2], "inactive_file", "file_dirty", "file_writeback")
	return
}

func reclaimableFileBytes(data []byte, inactiveKey, dirtyKey, writebackKey string) (int64, error) {
	fields := strings.Fields(string(data))
	if len(fields) == 0 || len(fields)%2 != 0 {
		return 0, moerr.NewInternalErrorNoCtx("incomplete cgroup memory.stat")
	}
	values := make(map[string]int64)
	for i := 0; i < len(fields); i += 2 {
		n, e := strconv.ParseUint(fields[i+1], 10, 64)
		if e != nil {
			return 0, e
		}
		values[fields[i]] = memoryInt(n)
	}
	// Older kernels may omit dirty/writeback counters. In that case retain all
	// file pages in the pressure estimate rather than inventing reclaim credit.
	raw, hasInactive := values[inactiveKey]
	dirty, hasDirty := values[dirtyKey]
	writeback, hasWriteback := values[writebackKey]
	if hasInactive && hasDirty && hasWriteback {
		return max(0, max(0, raw-dirty)-writeback), nil
	}
	return 0, nil
}
