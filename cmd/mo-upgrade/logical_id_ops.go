// Copyright 2024 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mo_upgrade

import (
	"fmt"
	"strings"
	"time"
)

func (c *upgradeContext) getNullLogicalIDCount() (int, error) {
	var count int
	err := c.db.QueryRow("SELECT count(*) FROM mo_catalog.mo_tables WHERE rel_logical_id IS NULL").Scan(&count)
	return count, err
}

func (c *upgradeContext) getObjectCount() (int, error) {
	var count int
	err := c.db.QueryRow("SELECT count(*) FROM metadata_scan('mo_catalog.mo_tables', 'relname')").Scan(&count)
	return count, err
}

func (c *upgradeContext) flushMoTables() error {
	_, err := c.db.Exec("SELECT mo_ctl('dn', 'flush', 'mo_catalog.mo_tables')")
	return err
}

func (c *upgradeContext) mergeObjects() error {
	// Get merge SQL
	var mergeSQL string
	err := c.db.QueryRow(`
		SELECT concat(
			"SELECT mo_ctl('dn', 'inspect', 'merge trigger -t 1.2 --kind objects --objects ", 
			group_concat(object_name), 
			"');"
		) FROM metadata_scan('mo_catalog.mo_tables', 'relname') g
	`).Scan(&mergeSQL)
	if err != nil {
		return fmt.Errorf("failed to generate merge SQL: %w", err)
	}

	// Extract the inner SQL (remove outer SELECT and quotes)
	mergeSQL = strings.TrimPrefix(mergeSQL, "SELECT ")
	mergeSQL = strings.TrimSuffix(mergeSQL, ";")

	fmt.Printf("  Executing merge command...\n")
	_, err = c.db.Exec(mergeSQL)
	if err != nil {
		return fmt.Errorf("failed to execute merge: %w", err)
	}
	return nil
}

func (c *upgradeContext) waitForMerge() error {
	maxWait := 5 * time.Minute
	interval := 3 * time.Second
	start := time.Now()

	for {
		count, err := c.getObjectCount()
		if err != nil {
			return fmt.Errorf("failed to check object count: %w", err)
		}

		elapsed := time.Since(start)
		fmt.Printf("\r  Objects remaining: %d (elapsed: %s)    ", count, elapsed.Round(time.Second))

		if count == 1 {
			fmt.Println()
			return nil
		}

		if elapsed > maxWait {
			fmt.Println()
			return fmt.Errorf("merge timeout after %s, still have %d objects", maxWait, count)
		}

		time.Sleep(interval)
	}
}

func (c *upgradeContext) syncIndexTable() error {
	// Clear index table
	fmt.Println("  Clearing index table...")
	_, err := c.db.Exec("DELETE FROM mo_catalog.__mo_index_unique_mo_tables_logical_id WHERE __mo_index_idx_col > 0")
	if err != nil {
		return fmt.Errorf("failed to clear index table: %w", err)
	}

	// Import data from mo_tables
	fmt.Println("  Importing data from mo_tables...")
	_, err = c.db.Exec(`
		INSERT INTO mo_catalog.__mo_index_unique_mo_tables_logical_id 
		SELECT rel_logical_id, __mo_cpkey_col FROM mo_catalog.mo_tables
	`)
	if err != nil {
		return fmt.Errorf("failed to import data: %w", err)
	}

	return nil
}

func (c *upgradeContext) verifyIndexTable() error {
	var moTablesCount, indexCount int

	err := c.db.QueryRow("SELECT count(*) FROM mo_catalog.mo_tables").Scan(&moTablesCount)
	if err != nil {
		return fmt.Errorf("failed to count mo_tables: %w", err)
	}

	err = c.db.QueryRow("SELECT count(*) FROM mo_catalog.__mo_index_unique_mo_tables_logical_id").Scan(&indexCount)
	if err != nil {
		return fmt.Errorf("failed to count index table: %w", err)
	}

	fmt.Printf("  - mo_tables rows: %d\n", moTablesCount)
	fmt.Printf("  - index table rows: %d\n", indexCount)

	if moTablesCount != indexCount {
		return fmt.Errorf("row count mismatch: mo_tables=%d, index=%d", moTablesCount, indexCount)
	}

	printSuccess("Row counts match")
	return nil
}

func (c *upgradeContext) ensureIndexTableSync() error {
	var moTablesCount, indexCount int

	err := c.db.QueryRow("SELECT count(*) FROM mo_catalog.mo_tables").Scan(&moTablesCount)
	if err != nil {
		return err
	}

	err = c.db.QueryRow("SELECT count(*) FROM mo_catalog.__mo_index_unique_mo_tables_logical_id").Scan(&indexCount)
	if err != nil {
		return err
	}

	fmt.Printf("  - mo_tables rows: %d\n", moTablesCount)
	fmt.Printf("  - index table rows: %d\n", indexCount)

	if moTablesCount == indexCount {
		printSuccess("Index table is already in sync")
		return nil
	}

	fmt.Printf("  Index table needs sync (diff: %d)\n", moTablesCount-indexCount)
	if !confirm("Sync index table?") {
		return nil
	}

	return c.syncIndexTable()
}
