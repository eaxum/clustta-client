package sync_service

import (
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// TypeMapping holds the mapping between local and remote type IDs.
type TypeMapping struct {
	LocalID  string
	RemoteID string
	Name     string
}

// TypeMappings holds all type mappings for a project upload.
type TypeMappings struct {
	StatusMappings     []TypeMapping
	CollectionTypeMappings []TypeMapping
	AssetTypeMappings   []TypeMapping
	RoleMappings       []TypeMapping
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
			if local.Name == remote.Name {
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
			if local.Name == remote.Name {
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
			if local.Name == remote.Name {
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
			if local.Name == remote.Name {
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
		_, err := tx.Exec("UPDATE collection SET collection_type_id = ? WHERE collection_type_id = ?", m.RemoteID, m.LocalID)
		if err != nil {
			return fmt.Errorf("failed to remap collection_type_id %s to %s: %w", m.LocalID, m.RemoteID, err)
		}
	}

	// Remap asset type IDs in assets
	for _, m := range mappings.AssetTypeMappings {
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

// ReplaceTypeTables replaces local type tables with remote types.
func ReplaceTypeTables(tx *sqlx.Tx, remoteStatuses []models.Status, remoteCollectionTypes []models.CollectionType, remoteAssetTypes []models.AssetType, remoteRoles []models.Role) error {
	// Clear and replace statuses
	_, err := tx.Exec("DELETE FROM status")
	if err != nil {
		return fmt.Errorf("failed to clear status table: %w", err)
	}
	for _, s := range remoteStatuses {
		_, err := tx.Exec(
			"INSERT INTO status (id, name, short_name, color, synced, mtime) VALUES (?, ?, ?, ?, 1, ?)",
			s.Id, s.Name, s.ShortName, s.Color, utils.GetEpochTime(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert status %s: %w", s.Name, err)
		}
	}

	// Clear and replace collection types
	_, err = tx.Exec("DELETE FROM collection_type")
	if err != nil {
		return fmt.Errorf("failed to clear collection_type table: %w", err)
	}
	for _, et := range remoteCollectionTypes {
		_, err := tx.Exec(
			"INSERT INTO collection_type (id, name, icon, synced, mtime) VALUES (?, ?, ?, 1, ?)",
			et.Id, et.Name, et.Icon, utils.GetEpochTime(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert collection_type %s: %w", et.Name, err)
		}
	}

	// Clear and replace asset types
	_, err = tx.Exec("DELETE FROM asset_type")
	if err != nil {
		return fmt.Errorf("failed to clear asset_type table: %w", err)
	}
	for _, tt := range remoteAssetTypes {
		_, err := tx.Exec(
			"INSERT INTO asset_type (id, name, icon, synced, mtime) VALUES (?, ?, ?, 1, ?)",
			tt.Id, tt.Name, tt.Icon, utils.GetEpochTime(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert asset_type %s: %w", tt.Name, err)
		}
	}

	// Clear and replace roles
	_, err = tx.Exec("DELETE FROM role")
	if err != nil {
		return fmt.Errorf("failed to clear role table: %w", err)
	}
	for _, r := range remoteRoles {
		_, err := tx.Exec(
			`INSERT INTO role (id, name, synced, mtime, 
				view_collection, create_collection, update_collection, delete_collection,
				view_asset, create_asset, update_asset, delete_asset,
				view_template, create_template, update_template, delete_template,
				view_checkpoint, create_checkpoint, delete_checkpoint,
				pull_chunk, assign_asset, unassign_asset,
				add_user, remove_user, change_role,
				change_status, set_done_asset, set_retake_asset, view_done_asset, manage_dependencies
			) VALUES (?, ?, 1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			r.Id, r.Name, utils.GetEpochTime(),
			r.ViewCollection, r.CreateCollection, r.UpdateCollection, r.DeleteCollection,
			r.ViewAsset, r.CreateAsset, r.UpdateAsset, r.DeleteAsset,
			r.ViewTemplate, r.CreateTemplate, r.UpdateTemplate, r.DeleteTemplate,
			r.ViewCheckpoint, r.CreateCheckpoint, r.DeleteCheckpoint,
			r.PullChunk, r.AssignAsset, r.UnassignAsset,
			r.AddUser, r.RemoveUser, r.ChangeRole,
			r.ChangeStatus, r.SetDoneAsset, r.SetRetakeAsset, r.ViewDoneAsset, r.ManageDependencies,
		)
		if err != nil {
			return fmt.Errorf("failed to insert role %s: %w", r.Name, err)
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
	_, err = tx.Exec("INSERT OR REPLACE INTO config (name, value, mtime) VALUES ('working_directory', ?, ?)", workingDirectory, mtime)
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
		"asset_dependency",
		"collection_dependency",
		"dependency_type",
		"user",
		"role",
		"template",
		"workflow",
		"workflow_link",
		"workflow_collection",
		"workflow_asset",
		"tag",
		"asset_tag",
	}

	for _, table := range tables {
		_, err := tx.Exec(fmt.Sprintf("UPDATE %s SET synced = 0", table))
		if err != nil {
			// Table might not exist or be empty, continue
			continue
		}
	}

	// Also mark project preview as unsynced
	_, err := tx.Exec("UPDATE config SET synced = 0 WHERE name = 'project_preview'")
	if err != nil {
		// Ignore if doesn't exist
	}

	return nil
}

// FetchRemoteProjectTypes fetches the type configurations from a remote project.
func FetchRemoteProjectTypes(remoteUrl string, userId string) ([]models.Status, []models.CollectionType, []models.AssetType, []models.Role, error) {
	data, err := FetchData(remoteUrl, userId)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to fetch remote project data: %w", err)
	}

	return data.Statuses, data.CollectionTypes, data.AssetTypes, data.Roles, nil
}

// PrepareProjectForUpload prepares a local .clst file for upload to a remote studio.
// It remaps all type IDs to match the remote project and updates config accordingly.
func PrepareProjectForUpload(projectPath string, remoteProjectInfo repository.ProjectInfo, remoteUrl string, workingDirectory string, userId string) error {
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

	// Fetch local types
	localStatuses, localCollectionTypes, localAssetTypes, localRoles, err := FetchLocalTypes(tx)
	if err != nil {
		return err
	}

	// Fetch remote types
	remoteStatuses, remoteCollectionTypes, remoteAssetTypes, remoteRoles, err := FetchRemoteProjectTypes(remoteUrl, userId)
	if err != nil {
		return err
	}

	// Build mappings
	mappings := BuildTypeMappings(
		localStatuses, remoteStatuses,
		localCollectionTypes, remoteCollectionTypes,
		localAssetTypes, remoteAssetTypes,
		localRoles, remoteRoles,
	)

	// Remap IDs in assets and collections
	err = RemapProjectIds(tx, mappings)
	if err != nil {
		return err
	}

	// Replace type tables with remote types
	err = ReplaceTypeTables(tx, remoteStatuses, remoteCollectionTypes, remoteAssetTypes, remoteRoles)
	if err != nil {
		return err
	}

	// Update config
	err = UpdateProjectConfig(tx, remoteProjectInfo.Id, remoteUrl, workingDirectory)
	if err != nil {
		return err
	}

	// Mark all tables as unsynced
	err = MarkAllTablesUnsynced(tx)
	if err != nil {
		return err
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
