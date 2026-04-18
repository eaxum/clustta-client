package migrations

import (
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
)

// MigrateV1_9 adds the manage_share_links permission and renames is_library to is_shared.
func MigrateV1_9(db *sqlx.DB, schema string) error {
	err := utils.AddColumnIfNotExist(db, "role", "manage_share_links", "BOOLEAN", "0", false)
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE role SET manage_share_links = 1 WHERE name IN ('admin', 'supervisor') AND manage_share_links = 0`)
	if err != nil {
		return err
	}
	err = utils.RenameColumn(db, "collection", "is_library", "is_shared")
	return err
}
