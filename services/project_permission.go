package services

import (
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func requireProjectPermissionForPath(projectPath string, permission projectPermission) error {
	db, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	return requireProjectPermission(tx, permission)
}

type projectPermission string

const (
	permissionManageCollectionTypes projectPermission = "manage_collection_types"
	permissionManageAssetTypes      projectPermission = "manage_asset_types"
	permissionManageTags            projectPermission = "manage_tags"
	permissionManageWorkflows       projectPermission = "manage_workflows"
	permissionManageIntegrations    projectPermission = "manage_integrations"
	permissionManageProjectSettings projectPermission = "manage_project_settings"
	permissionManageRoles           projectPermission = "manage_roles"
)

func requireProjectPermission(tx *sqlx.Tx, permission projectPermission) error {
	_, role, err := activeAssetRole(tx)
	if err != nil {
		return err
	}
	if hasProjectPermission(role, permission) {
		return nil
	}
	return fmt.Errorf("user does not have %s permission", permission)
}

func hasProjectPermission(role models.Role, permission projectPermission) bool {
	switch permission {
	case permissionManageCollectionTypes:
		return role.ManageCollectionTypes
	case permissionManageAssetTypes:
		return role.ManageAssetTypes
	case permissionManageTags:
		return role.ManageTags
	case permissionManageWorkflows:
		return role.ManageWorkflows
	case permissionManageIntegrations:
		return role.ManageIntegrations
	case permissionManageProjectSettings:
		return role.ManageProjectSettings
	case permissionManageRoles:
		return role.ManageRoles
	default:
		return false
	}
}
