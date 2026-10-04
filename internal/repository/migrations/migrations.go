package migrations

import (
	"clustta/internal/compatibility"
	"clustta/internal/utils"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

// LatestVersion is the current schema version after all migrations.
const LatestVersion = "2.2"

// Migration defines a single schema migration step.
type Migration struct {
	Version     string
	Description string
	Up          func(tx *sqlx.Tx, schema string) error
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
		{Version: "2.2", Description: "Add versioned dependencies and project management permissions", Up: MigrateV2_2},
	}
}

// RunMigrations applies all pending migrations to the database.
func RunMigrations(db *sqlx.DB, currentVersion string, schema string) error {
	currentVersion = normalizeLegacyVersion(currentVersion)
	comparison, err := compatibility.CompareVersions(currentVersion, LatestVersion)
	if err != nil {
		return err
	}
	if comparison > 0 {
		return fmt.Errorf("project schema %s is newer than supported schema %s", currentVersion, LatestVersion)
	}
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

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
			if err := m.Up(tx, schema); err != nil {
				return fmt.Errorf("failed to migrate project schema to %s: %w", m.Version, err)
			}
		}
	}

	// Re-apply schema to ensure views, triggers, and indexes are current.
	err = utils.CreateSchemaTx(tx, schema)
	if err != nil {
		return err
	}
	var integrity string
	if err = tx.Get(&integrity, "PRAGMA quick_check"); err != nil {
		return fmt.Errorf("failed to validate migrated project: %w", err)
	}
	if integrity != "ok" {
		return fmt.Errorf("migrated project failed integrity check: %s", integrity)
	}

	err = utils.SetProjectVersion(tx, LatestVersion)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func normalizeLegacyVersion(version string) string {
	if strings.Contains(version, ".") {
		return version
	}
	number, err := strconv.Atoi(version)
	if err != nil || number < 0 || strconv.Itoa(number) != version {
		return version
	}
	return version + ".0"
}
