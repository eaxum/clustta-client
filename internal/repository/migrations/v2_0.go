package migrations

import (
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
)

// MigrateV2_0 adds the server-owned project storage tables.
func MigrateV2_0(db *sqlx.DB, schema string) error {
	return utils.CreateSchema(db, schema)
}
