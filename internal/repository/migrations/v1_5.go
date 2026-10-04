package migrations

import (
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
)

// MigrateV1_5 adds entity_path column, creates an index, and backfills paths.
func MigrateV1_5(tx *sqlx.Tx, _ string) error {
	err := utils.AddColumnIfNotExist(tx, "entity", "entity_path", "TEXT", "", false)
	if err != nil {
		return err
	}

	_, err = tx.Exec("CREATE INDEX IF NOT EXISTS idx_entity_path ON entity(entity_path);")
	if err != nil {
		return err
	}

	type collectionPath struct {
		Id   string `db:"id"`
		Path string `db:"entity_path"`
	}
	var paths []collectionPath
	err = tx.Select(&paths, "SELECT id, entity_path FROM entity_hierarchy")
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE entity SET entity_path = ? WHERE id = ?")
	if err != nil {
		return err
	}

	for _, p := range paths {
		_, err := stmt.Exec(p.Path, p.Id)
		if err != nil {
			return err
		}
	}

	return nil
}
