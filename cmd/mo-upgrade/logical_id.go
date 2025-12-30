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
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

var (
	host     string
	port     int
	user     string
	password string
	autoYes  bool
	dryRun   bool
)

var logicalIDCmd = &cobra.Command{
	Use:   "logical-id",
	Short: "Upgrade rel_logical_id for mo_tables",
	Long: `Upgrade rel_logical_id for mo_tables.

This command performs the following steps:
  1. Flush mo_tables to disk
  2. Merge all objects into one
  3. Wait for merge completion
  4. Verify rel_logical_id is populated
  5. Sync index table data

Example:
  mo-tool upgrade logical-id -H 127.0.0.1 -P 6001 -u dump -p 111`,
	RunE: runLogicalIDUpgrade,
}

func init() {
	logicalIDCmd.Flags().StringVarP(&host, "host", "H", "127.0.0.1", "MO server host")
	logicalIDCmd.Flags().IntVarP(&port, "port", "P", 6001, "MO server port")
	logicalIDCmd.Flags().StringVarP(&user, "user", "u", "dump", "MO username")
	logicalIDCmd.Flags().StringVarP(&password, "password", "p", "111", "MO password")
	logicalIDCmd.Flags().BoolVarP(&autoYes, "yes", "y", false, "Auto confirm all prompts")
	logicalIDCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without executing")
}

type upgradeContext struct {
	db *sql.DB
}

func newUpgradeContext() (*upgradeContext, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/mo_catalog", user, password, host, port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping: %w", err)
	}
	return &upgradeContext{db: db}, nil
}

func (c *upgradeContext) close() {
	if c.db != nil {
		c.db.Close()
	}
}

func runLogicalIDUpgrade(cmd *cobra.Command, args []string) error {
	printHeader("MO Logical ID Upgrade Tool")
	if dryRun {
		fmt.Printf("%s[DRY-RUN MODE]%s - No changes will be made\n", colorYellow, colorReset)
	}
	fmt.Printf("Connecting to %s:%d as %s...\n", host, port, user)

	ctx, err := newUpgradeContext()
	if err != nil {
		return err
	}
	defer ctx.close()
	printSuccess("Connected successfully")

	// Step 1: Check current status
	printStep(1, "Checking current status")
	nullCount, err := ctx.getNullLogicalIDCount()
	if err != nil {
		return err
	}
	objectCount, err := ctx.getObjectCount()
	if err != nil {
		return err
	}
	fmt.Printf("  - Objects in mo_tables: %d\n", objectCount)
	fmt.Printf("  - Rows with NULL rel_logical_id: %d\n", nullCount)

	// 核心目标：确保 rel_logical_id 都有值
	// merge 是手段，通过 merge 触发 rel_logical_id 的填充
	// object = 0 说明还没刷盘，需要继续 flush-merge 流程
	if nullCount == 0 && objectCount > 0 {
		printSuccess("All rel_logical_id values are already populated")
		return ctx.ensureIndexTableSync()
	}

	if dryRun {
		fmt.Println("\n[DRY-RUN] Would perform the following steps:")
		fmt.Println("  2. Flush mo_tables to disk")
		fmt.Println("  3. Merge all objects into one")
		fmt.Println("  4. Wait for merge completion")
		fmt.Println("  5. Verify rel_logical_id is populated")
		fmt.Println("  6. Sync index table")
		fmt.Println("  7. Final verification")
		return nil
	}

	if !confirm("Proceed with upgrade?") {
		fmt.Println("Upgrade cancelled.")
		return nil
	}

	// Step 2: Flush mo_tables
	printStep(2, "Flushing mo_tables to disk")
	if err := ctx.flushMoTables(); err != nil {
		return err
	}
	printSuccess("Flush completed")

	// Step 3: Merge objects
	printStep(3, "Merging objects")
	if err := ctx.mergeObjects(); err != nil {
		return err
	}
	printSuccess("Merge triggered")

	// Step 4: Wait for merge completion
	printStep(4, "Waiting for merge completion")
	if err := ctx.waitForMerge(); err != nil {
		return err
	}
	printSuccess("Merge completed")

	// Step 5: Verify rel_logical_id
	printStep(5, "Verifying rel_logical_id")
	nullCount, err = ctx.getNullLogicalIDCount()
	if err != nil {
		return err
	}
	if nullCount > 0 {
		return fmt.Errorf("still have %d rows with NULL rel_logical_id", nullCount)
	}
	printSuccess("All rel_logical_id values are populated")

	// Step 6: Sync index table
	printStep(6, "Syncing index table")
	if err := ctx.syncIndexTable(); err != nil {
		return err
	}
	printSuccess("Index table synced")

	// Final verification
	printStep(7, "Final verification")
	if err := ctx.verifyIndexTable(); err != nil {
		return err
	}

	printHeader("Upgrade completed successfully!")
	return nil
}
