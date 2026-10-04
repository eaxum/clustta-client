package migrations

import (
	_ "embed"

	"github.com/jmoiron/sqlx"
)

//go:embed sql/v1_7.sql
var v1_7SQL string

// MigrateV1_7 adds integration tables.
func MigrateV1_7(tx *sqlx.Tx, _ string) error {
	_, err := tx.Exec(v1_7SQL)
	return err
}
