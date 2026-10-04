package migrations

import (
	_ "embed"

	"github.com/jmoiron/sqlx"
)

//go:embed sql/v2_0.sql
var v2_0SQL string

// MigrateV2_0 adds the server-owned project storage tables.
func MigrateV2_0(tx *sqlx.Tx, _ string) error {
	_, err := tx.Exec(v2_0SQL)
	return err
}
