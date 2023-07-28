package jobs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXxx(t *testing.T) {
	toLayout := []uint32{8192, 8192, 1203}
	bisectHaystack := make([]uint32, 0, len(toLayout)+1)
	bisectHaystack = append(bisectHaystack, 0)
	for _, x := range toLayout {
		bisectHaystack = append(bisectHaystack, bisectHaystack[len(bisectHaystack)-1]+x)
	}

	bisectPinpoint := func(needle uint32) (int, uint32) {
		i, j := 0, len(bisectHaystack)
		for i < j {
			m := (i + j) / 2
			if bisectHaystack[m] > needle {
				j = m
			} else {
				i = m + 1
			}
		}
		// bisectHaystack[i] is the first number > needle, so the needle falls into i-1 th block
		blkIdx := i - 1
		rows := needle - bisectHaystack[blkIdx]
		return blkIdx, rows
	}

	var blk int
	var idx uint32
	blk, idx = bisectPinpoint(0)
	require.Equal(t, 0, blk)
	require.Equal(t, 0, int(idx))

	blk, idx = bisectPinpoint(1)
	require.Equal(t, 0, blk)
	require.Equal(t, 1, int(idx))

	blk, idx = bisectPinpoint(8191)
	require.Equal(t, 0, blk)
	require.Equal(t, 8191, int(idx))

	blk, idx = bisectPinpoint(8192)
	require.Equal(t, 1, blk)
	require.Equal(t, 0, int(idx))

	blk, idx = bisectPinpoint(8193)
	require.Equal(t, 1, blk)
	require.Equal(t, 1, int(idx))

	blk, idx = bisectPinpoint(16383)
	require.Equal(t, 1, blk)
	require.Equal(t, 8191, int(idx))

	blk, idx = bisectPinpoint(16384)
	require.Equal(t, 2, blk)
	require.Equal(t, 0, int(idx))
	blk, idx = bisectPinpoint(16385)
	require.Equal(t, 2, blk)
	require.Equal(t, 1, int(idx))

	blk, idx = bisectPinpoint(17000)
	require.Equal(t, 2, blk)
	require.Equal(t, 616, int(idx))
}
