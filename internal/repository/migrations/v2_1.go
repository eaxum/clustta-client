package migrations

import (
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
)

// MigrateV2_1 adds client-only pending path tracking.
func MigrateV2_1(db *sqlx.DB, schema string) error {
	return utils.CreateSchema(db, schema)
}
