package migrations

import (
	_ "embed"

	"github.com/jmoiron/sqlx"
)

//go:embed sql/v2_1.sql
var v2_1SQL string

// MigrateV2_1 adds client-only pending path tracking.
func MigrateV2_1(tx *sqlx.Tx, _ string) error {
	_, err := tx.Exec(v2_1SQL)
	return err
}
