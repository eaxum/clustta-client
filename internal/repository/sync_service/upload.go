package sync_service

import (
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

const projectPreviewConfigName = "project_preview"

// TypeMapping holds the mapping between local and remote type IDs.
type TypeMapping struct {
	LocalID  string
	RemoteID string
	Name     string
}

// TypeMappings holds all type mappings for a project upload.
type TypeMappings struct {
	StatusMappings         []TypeMapping
	CollectionTypeMappings []TypeMapping
	AssetTypeMappings      []TypeMapping
	RoleMappings           []TypeMapping
}

// FetchLocalTypes retrieves all types from the local project database.
func FetchLocalTypes(tx *sqlx.Tx) (statuses []models.Status, collectionTypes []models.CollectionType, assetTypes []models.AssetType, roles []models.Role, err error) {
	err = tx.Select(&statuses, "SELECT * FROM status")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to fetch local statuses: %w", err)
	}

	err = tx.Select(&collectionTypes, "SELECT * FROM collection_type")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to fetch local collection types: %w", err)
	}

	err = tx.Select(&assetTypes, "SELECT * FROM asset_type")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to fetch local asset types: %w", err)
	}

	err = tx.Select(&roles, "SELECT * FROM role")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to fetch local roles: %w", err)
	}

	return statuses, collectionTypes, assetTypes, roles, nil
}

// BuildTypeMappings creates mappings between local and remote type IDs by matching names.
func BuildTypeMappings(
	localStatuses []models.Status, remoteStatuses []models.Status,
	localCollectionTypes []models.CollectionType, remoteCollectionTypes []models.CollectionType,
	localAssetTypes []models.AssetType, remoteAssetTypes []models.AssetType,
	localRoles []models.Role, remoteRoles []models.Role,
) TypeMappings {
	mappings := TypeMappings{}

	// Build status mappings
	for _, local := range localStatuses {
		for _, remote := range remoteStatuses {
			if strings.EqualFold(local.Name, remote.Name) {
				mappings.StatusMappings = append(mappings.StatusMappings, TypeMapping{
					LocalID:  local.Id,
					RemoteID: remote.Id,
					Name:     local.Name,
				})
				break
			}
		}
	}

	// Build collection type mappings
	for _, local := range localCollectionTypes {
		for _, remote := range remoteCollectionTypes {
			if strings.EqualFold(local.Name, remote.Name) {
				mappings.CollectionTypeMappings = append(mappings.CollectionTypeMappings, TypeMapping{
					LocalID:  local.Id,
					RemoteID: remote.Id,
					Name:     local.Name,
				})
				break
			}
		}
	}

	// Build asset type mappings
	for _, local := range localAssetTypes {
		for _, remote := range remoteAssetTypes {
			if strings.EqualFold(local.Name, remote.Name) {
				mappings.AssetTypeMappings = append(mappings.AssetTypeMappings, TypeMapping{
					LocalID:  local.Id,
					RemoteID: remote.Id,
					Name:     local.Name,
				})
				break
			}
		}
	}

	// Build role mappings
	for _, local := range localRoles {
		for _, remote := range remoteRoles {
			if strings.EqualFold(local.Name, remote.Name) {
				mappings.RoleMappings = append(mappings.RoleMappings, TypeMapping{
					LocalID:  local.Id,
					RemoteID: remote.Id,
					Name:     local.Name,
				})
				break
			}
		}
	}

	return mappings
}

// RemapProjectIds updates all type IDs in the local database to match remote IDs.
func RemapProjectIds(tx *sqlx.Tx, mappings TypeMappings) error {
	// Remap status IDs in assets
	for _, m := range mappings.StatusMappings {
		_, err := tx.Exec("UPDATE asset SET status_id = ? WHERE status_id = ?", m.RemoteID, m.LocalID)
		if err != nil {
			return fmt.Errorf("failed to remap status_id %s to %s: %w", m.LocalID, m.RemoteID, err)
		}
	}

	// Remap collection type IDs in collections
	for _, m := range mappings.CollectionTypeMappings {
		for _, table := range []string{"workflow_collection", "workflow_link"} {
			if _, err := tx.Exec("UPDATE "+table+" SET collection_type_id = ? WHERE collection_type_id = ?", m.RemoteID, m.LocalID); err != nil {
				return fmt.Errorf("failed to remap %s collection type: %w", table, err)
			}
		}
		_, err := tx.Exec("UPDATE collection SET collection_type_id = ? WHERE collection_type_id = ?", m.RemoteID, m.LocalID)
		if err != nil {
			return fmt.Errorf("failed to remap collection_type_id %s to %s: %w", m.LocalID, m.RemoteID, err)
		}
	}

	// Remap asset type IDs in assets
	for _, m := range mappings.AssetTypeMappings {
		if _, err := tx.Exec("UPDATE workflow_asset SET asset_type_id = ? WHERE asset_type_id = ?", m.RemoteID, m.LocalID); err != nil {
			return fmt.Errorf("failed to remap workflow asset type: %w", err)
		}
		_, err := tx.Exec("UPDATE asset SET asset_type_id = ? WHERE asset_type_id = ?", m.RemoteID, m.LocalID)
		if err != nil {
			return fmt.Errorf("failed to remap asset_type_id %s to %s: %w", m.LocalID, m.RemoteID, err)
		}
	}

	// Remap role IDs in users
	for _, m := range mappings.RoleMappings {
		_, err := tx.Exec("UPDATE user SET role_id = ? WHERE role_id = ?", m.RemoteID, m.LocalID)
		if err != nil {
			return fmt.Errorf("failed to remap role_id %s to %s: %w", m.LocalID, m.RemoteID, err)
		}
	}

	return nil
}

// UpdateProjectConfig updates the config table with remote project settings.
func UpdateProjectConfig(tx *sqlx.Tx, projectId, remoteUrl, workingDirectory string) error {
	mtime := utils.GetEpochTime()

	// Update or insert project_id
	_, err := tx.Exec("INSERT OR REPLACE INTO config (name, value, mtime) VALUES ('project_id', ?, ?)", projectId, mtime)
	if err != nil {
		return fmt.Errorf("failed to update project_id: %w", err)
	}

	// Update or insert remote
	_, err = tx.Exec("INSERT OR REPLACE INTO config (name, value, mtime) VALUES ('remote', ?, ?)", remoteUrl, mtime)
	if err != nil {
		return fmt.Errorf("failed to update remote: %w", err)
	}

	// Clear sync_token to force full sync
	_, err = tx.Exec("INSERT OR REPLACE INTO config (name, value, mtime) VALUES ('sync_token', '', ?)", mtime)
	if err != nil {
		return fmt.Errorf("failed to clear sync_token: %w", err)
	}

	// Update working directory
	_, err = tx.Exec("INSERT OR REPLACE INTO config (name, value, mtime) VALUES ('working_dir', ?, ?)", workingDirectory, mtime)
	if err != nil {
		return fmt.Errorf("failed to update working_directory: %w", err)
	}

	return nil
}

// MarkAllTablesUnsynced sets synced = 0 for all data tables to ensure full sync.
func MarkAllTablesUnsynced(tx *sqlx.Tx) error {
	tables := []string{
		"collection",
		"collection_assignee",
		"asset",
		"asset_checkpoint",
		"asset_checkpoint_tag",
		"asset_dependency",
		"collection_dependency",
		"dependency_type",
		"status",
		"collection_type",
		"asset_type",
		"user",
		"role",
		"template",
		"workflow",
		"workflow_link",
		"workflow_collection",
		"workflow_asset",
		"tag",
		"asset_tag",
		"integration_project",
		"integration_collection_mapping",
		"integration_asset_mapping",
	}

	for _, table := range tables {
		_, err := tx.Exec(fmt.Sprintf("UPDATE %s SET synced = 0", table))
		if err != nil {
			return fmt.Errorf("failed to mark %s unsynced: %w", table, err)
		}
	}

	configNames := append([]string{projectPreviewConfigName}, repository.SyncableProjectConfigNames...)
	query, args, err := sqlx.In("UPDATE config SET synced = 0 WHERE name IN (?)", configNames)
	if err != nil {
		return err
	}
	_, err = tx.Exec(tx.Rebind(query), args...)
	return err
}

// FetchRemoteProjectTypes fetches the type configurations from a remote project.
func FetchRemoteProjectTypes(remoteUrl string, userId string) ([]models.Status, []models.CollectionType, []models.AssetType, []models.Role, error) {
	data, err := FetchData(remoteUrl, userId)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to fetch remote project data: %w", err)
	}

	return data.Statuses, data.CollectionTypes, data.AssetTypes, data.Roles, nil
}

// RemapTypeTableIds updates the primary key IDs in type tables to match the remote project.
// Custom types that only exist locally are preserved.
func RemapTypeTableIds(tx *sqlx.Tx, mappings TypeMappings) error {
	for _, m := range mappings.StatusMappings {
		_, err := tx.Exec("UPDATE status SET id = ? WHERE id = ?", m.RemoteID, m.LocalID)
		if err != nil {
			return fmt.Errorf("failed to remap status id %s to %s: %w", m.LocalID, m.RemoteID, err)
		}
	}

	for _, m := range mappings.CollectionTypeMappings {
		_, err := tx.Exec("UPDATE collection_type SET id = ? WHERE id = ?", m.RemoteID, m.LocalID)
		if err != nil {
			return fmt.Errorf("failed to remap collection_type id %s to %s: %w", m.LocalID, m.RemoteID, err)
		}
	}

	for _, m := range mappings.AssetTypeMappings {
		_, err := tx.Exec("UPDATE asset_type SET id = ? WHERE id = ?", m.RemoteID, m.LocalID)
		if err != nil {
			return fmt.Errorf("failed to remap asset_type id %s to %s: %w", m.LocalID, m.RemoteID, err)
		}
	}

	for _, m := range mappings.RoleMappings {
		_, err := tx.Exec("UPDATE role SET id = ? WHERE id = ?", m.RemoteID, m.LocalID)
		if err != nil {
			return fmt.Errorf("failed to remap role id %s to %s: %w", m.LocalID, m.RemoteID, err)
		}
	}

	return nil
}

// PrepareProjectForCloudUpload preserves custom types while aligning remote IDs.
func PrepareProjectForCloudUpload(projectPath string, projectId string, remoteUrl string, workingDirectory string, userId string) error {
	return prepareProjectForRemoteUpload(projectPath, projectId, remoteUrl, workingDirectory, userId)
}

func prepareProjectForRemoteUpload(projectPath, projectId, remoteUrl, workingDirectory, userId string) error {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return fmt.Errorf("failed to open project database: %w", err)
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Fetch local and remote types
	localStatuses, localCollectionTypes, localAssetTypes, localRoles, err := FetchLocalTypes(tx)
	if err != nil {
		return err
	}

	remoteStatuses, remoteCollectionTypes, remoteAssetTypes, remoteRoles, err := FetchRemoteProjectTypes(remoteUrl, userId)
	if err != nil {
		return err
	}
	for _, remote := range remoteCollectionTypes {
		if err := validateUploadIcon(tx, "collection_type", remote.Name, remote.Icon); err != nil {
			return err
		}
	}
	for _, remote := range remoteAssetTypes {
		if err := validateUploadIcon(tx, "asset_type", remote.Name, remote.Icon); err != nil {
			return err
		}
	}

	// Build mappings between local and remote types by name
	mappings := BuildTypeMappings(
		localStatuses, remoteStatuses,
		localCollectionTypes, remoteCollectionTypes,
		localAssetTypes, remoteAssetTypes,
		localRoles, remoteRoles,
	)

	// Remap FK references in assets, collections, and users
	err = RemapProjectIds(tx, mappings)
	if err != nil {
		return err
	}

	// Remap PKs in type tables to match remote IDs (preserves custom types)
	err = RemapTypeTableIds(tx, mappings)
	if err != nil {
		return err
	}
	if err := validateUploadReferences(tx); err != nil {
		return err
	}

	// Update config with remote project settings
	err = UpdateProjectConfig(tx, projectId, remoteUrl, workingDirectory)
	if err != nil {
		return err
	}

	// Mark all tables as unsynced so the next sync pushes everything
	err = MarkAllTablesUnsynced(tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// PrepareProjectForUpload shares non-destructive preparation with cloud studio uploads.
func PrepareProjectForUpload(projectPath string, remoteProjectInfo repository.ProjectInfo, remoteUrl string, workingDirectory string, userId string) error {
	return prepareProjectForRemoteUpload(projectPath, remoteProjectInfo.Id, remoteUrl, workingDirectory, userId)
}

func validateUploadIcon(tx *sqlx.Tx, table, remoteName, remoteIcon string) error {
	var names []string
	if err := tx.Select(&names, "SELECT name FROM "+table+" WHERE icon = ? AND name != ? COLLATE NOCASE", remoteIcon, remoteName); err != nil {
		return err
	}
	if len(names) != 0 {
		return fmt.Errorf("%s %q uses icon %q already used by remote type %q; choose a different icon before uploading", table, names[0], remoteIcon, remoteName)
	}
	return nil
}

func validateUploadReferences(tx *sqlx.Tx) error {
	for _, reference := range []struct{ table, column, target string }{
		{"collection", "collection_type_id", "collection_type"},
		{"asset", "asset_type_id", "asset_type"},
		{"asset", "status_id", "status"},
		{"user", "role_id", "role"},
		{"workflow_collection", "collection_type_id", "collection_type"},
		{"workflow_link", "collection_type_id", "collection_type"},
		{"workflow_asset", "asset_type_id", "asset_type"},
	} {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s r WHERE NOT EXISTS (SELECT 1 FROM %s t WHERE t.id = r.%s)", reference.table, reference.target, reference.column)
		if err := tx.Get(&count, query); err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("cannot upload: %d rows in %s reference missing %s records", count, reference.table, reference.target)
		}
	}
	return nil
}
