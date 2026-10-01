package migrations

import (
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
)

var projectManagementPermissions = []string{
	"manage_collection_types",
	"manage_asset_types",
	"manage_dependency_types",
	"manage_statuses",
	"manage_tags",
	"manage_workflows",
	"manage_integrations",
	"manage_project_settings",
	"manage_roles",
}

// MigrateV2_3 adds explicit project management permissions.
func MigrateV2_3(db *sqlx.DB, schema string) error {
	for _, permission := range projectManagementPermissions {
		if err := utils.AddColumnIfNotExist(db, "role", permission, "BOOLEAN", "0", false); err != nil {
			return err
		}
	}

	_, err := db.Exec(`
		UPDATE role SET
			manage_collection_types = 1,
			manage_asset_types = 1,
			manage_dependency_types = 1,
			manage_statuses = 1,
			manage_tags = 1,
			manage_workflows = 1,
			manage_integrations = 1,
			manage_project_settings = 1,
			manage_roles = 1
		WHERE lower(name) = 'admin'
	`)
	return err
}
