package migrations

import (
	_ "embed"

	"github.com/jmoiron/sqlx"
)

//go:embed sql/v1_6.sql
var v1_6SQL string

// MigrateV1_6 recreates the entity_path_update trigger.
func MigrateV1_6(db *sqlx.DB, _ string) error {
	_, err := db.Exec(v1_6SQL)
	return err
}
