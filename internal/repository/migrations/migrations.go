package migrations

import (
	"clustta/internal/compatibility"
	"clustta/internal/utils"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// LatestVersion is the current schema version after all migrations.
const LatestVersion = "2.2"

// Migration defines a single schema migration step.
type Migration struct {
	Version     string
	Description string
	Up          func(db *sqlx.DB, schema string) error
}

// All returns the ordered list of migrations.
func All() []Migration {
	return []Migration{
		{Version: "1.2", Description: "Rename checkpoint column, add columns, remap icons", Up: MigrateV1_2},
		{Version: "1.3", Description: "Set default working directory", Up: MigrateV1_3},
		{Version: "1.4", Description: "Add checkpoint grouping", Up: MigrateV1_4},
		{Version: "1.5", Description: "Add collection paths", Up: MigrateV1_5},
		{Version: "1.6", Description: "Add collection path update trigger", Up: MigrateV1_6},
		{Version: "1.7", Description: "Add integration tables", Up: MigrateV1_7},
		{Version: "1.8", Description: "Rename task/entity to asset/collection", Up: MigrateV1_8},
		{Version: "1.9", Description: "Add manage_share_links permission", Up: MigrateV1_9},
		{Version: "2.0", Description: "Add project storage tables", Up: MigrateV2_0},
		{Version: "2.1", Description: "Add pending local path updates", Up: MigrateV2_1},
		{Version: "2.2", Description: "Add versioned dependencies, checkpoint tags, and sources", Up: MigrateV2_2},
	}
}

// RunMigrations applies all pending migrations to the database.
func RunMigrations(db *sqlx.DB, currentVersion string, schema string) error {
	comparison, err := compatibility.CompareVersions(currentVersion, LatestVersion)
	if err != nil {
		return err
	}
	if comparison > 0 {
		return fmt.Errorf("project schema %s is newer than supported schema %s", currentVersion, LatestVersion)
	}
	// Earlier migrations reapply the latest schema before the final migration runs.
	for _, table := range []string{"asset_checkpoint", "task_checkpoint"} {
		exists, err := utils.TableExists(db, table)
		if err != nil {
			return err
		}
		if exists {
			if err = utils.AddColumnIfNotExist(db, table, "source_checkpoint_id", "TEXT", "", true); err != nil {
				return err
			}
		}
	}
	for _, m := range All() {
		shouldRun := false
		if m.Version == "1.2" {
			shouldRun = currentVersion == "1.2"
		} else {
			comparison, err := compatibility.CompareVersions(currentVersion, m.Version)
			if err != nil {
				return err
			}
			shouldRun = comparison < 0
		}

		if shouldRun {
			if err := m.Up(db, schema); err != nil {
				return err
			}
		}
	}

	// Re-apply schema to ensure views, triggers, and indexes are current.
	err = utils.CreateSchema(db, schema)
	if err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = utils.SetProjectVersion(tx, LatestVersion)
	if err != nil {
		return err
	}

	return tx.Commit()
}
