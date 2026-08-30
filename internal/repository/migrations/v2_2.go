package migrations

import (
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
)

// MigrateV2_2 adds checkpoint groups and versioned dependency selectors.
func MigrateV2_2(db *sqlx.DB, schema string) error {
	_, err := db.Exec(`
		DROP VIEW IF EXISTS full_asset;
		DROP VIEW IF EXISTS asset_dependencies;
		DROP TRIGGER IF EXISTS asset_dependency_selector_insert;
		DROP TRIGGER IF EXISTS asset_dependency_selector_update;
		DROP TRIGGER IF EXISTS checkpoint_tag_dependency_delete;
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
	if err := utils.AddColumnIfNotExist(db, "asset_dependency", "checkpoint_tag_id", "TEXT", "", true); err != nil {
		return err
	}
	if err := utils.CreateSchema(db, schema); err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE asset_checkpoint
		SET group_id = lower(hex(randomblob(16)))
		WHERE group_id = ''
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT OR IGNORE INTO checkpoint_group (
			id, mtime, created_at, group_type, finalized, synced
		)
		SELECT
			group_id,
			unixepoch(),
			MIN(created_at),
			CASE WHEN COUNT(DISTINCT CASE WHEN trashed = 0 THEN asset_id END) >= 2
				THEN 'multi' ELSE 'single' END,
			1,
			0
		FROM asset_checkpoint
		GROUP BY group_id
	`)
	if err != nil {
		return err
	}

	return tx.Commit()
}
