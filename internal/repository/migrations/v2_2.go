package migrations

import (
	"clustta/internal/utils"
	"github.com/jmoiron/sqlx"
)

// MigrateV2_2 adds checkpoint tags and versioned dependency selectors.
func MigrateV2_2(db *sqlx.DB, schema string) error {
	_, err := db.Exec(`
		DROP VIEW IF EXISTS full_asset;
		DROP VIEW IF EXISTS asset_dependencies;
		DROP TRIGGER IF EXISTS asset_dependency_selector_insert;
		DROP TRIGGER IF EXISTS asset_dependency_selector_update;
		DROP TRIGGER IF EXISTS asset_checkpoint_tag_dependency_delete;
	`)
	if err != nil {
		return err
	}

	if err := utils.AddColumnIfNotExist(db, "asset_dependency", "resolution_mode", "TEXT", "'floating'", false); err != nil {
		return err
	}
	if err := utils.AddColumnIfNotExist(db, "asset_dependency", "checkpoint_id", "TEXT", "", true); err != nil {
		return err
	}
	if err := utils.AddColumnIfNotExist(db, "asset_dependency", "asset_checkpoint_tag_id", "TEXT", "", true); err != nil {
		return err
	}
	if err := utils.CreateSchema(db, schema); err != nil {
		return err
	}

	return nil
}
