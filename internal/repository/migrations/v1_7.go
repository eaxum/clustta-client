package migrations

import (
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
)

// MigrateV1_7 re-applies the schema to add integration tables.
func MigrateV1_7(db *sqlx.DB, schema string) error {
	return utils.CreateSchema(db, schema)
}
