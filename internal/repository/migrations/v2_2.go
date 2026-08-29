package migrations

import (
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
)

// MigrateV2_2 adds explicit checkpoint groups and backfills existing group IDs.
func MigrateV2_2(db *sqlx.DB, schema string) error {
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
