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
	EntityTypeMappings []TypeMapping
	TaskTypeMappings   []TypeMapping
	RoleMappings       []TypeMapping
}

// FetchLocalTypes retrieves all types from the local project database.
func FetchLocalTypes(tx *sqlx.Tx) (statuses []models.Status, entityTypes []models.EntityType, taskTypes []models.TaskType, roles []models.Role, err error) {
	err = tx.Select(&statuses, "SELECT * FROM status")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to fetch local statuses: %w", err)
	}

	err = tx.Select(&entityTypes, "SELECT * FROM entity_type")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to fetch local entity types: %w", err)
	}

	err = tx.Select(&taskTypes, "SELECT * FROM task_type")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to fetch local task types: %w", err)
	}

	err = tx.Select(&roles, "SELECT * FROM role")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to fetch local roles: %w", err)
	}

	return statuses, entityTypes, taskTypes, roles, nil
}

// BuildTypeMappings creates mappings between local and remote type IDs by matching names.
func BuildTypeMappings(
	localStatuses []models.Status, remoteStatuses []models.Status,
	localEntityTypes []models.EntityType, remoteEntityTypes []models.EntityType,
	localTaskTypes []models.TaskType, remoteTaskTypes []models.TaskType,
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

	// Build entity type mappings
	for _, local := range localEntityTypes {
		for _, remote := range remoteEntityTypes {
			if local.Name == remote.Name {
				mappings.EntityTypeMappings = append(mappings.EntityTypeMappings, TypeMapping{
					LocalID:  local.Id,
					RemoteID: remote.Id,
					Name:     local.Name,
				})
				break
			}
		}
	}

	// Build task type mappings
	for _, local := range localTaskTypes {
		for _, remote := range remoteTaskTypes {
			if local.Name == remote.Name {
				mappings.TaskTypeMappings = append(mappings.TaskTypeMappings, TypeMapping{
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
	// Remap status IDs in tasks
	for _, m := range mappings.StatusMappings {
		_, err := tx.Exec("UPDATE task SET status_id = ? WHERE status_id = ?", m.RemoteID, m.LocalID)
		if err != nil {
			return fmt.Errorf("failed to remap status_id %s to %s: %w", m.LocalID, m.RemoteID, err)
		}
	}

	// Remap entity type IDs in entities
	for _, m := range mappings.EntityTypeMappings {
		_, err := tx.Exec("UPDATE entity SET entity_type_id = ? WHERE entity_type_id = ?", m.RemoteID, m.LocalID)
		if err != nil {
			return fmt.Errorf("failed to remap entity_type_id %s to %s: %w", m.LocalID, m.RemoteID, err)
		}
	}

	// Remap task type IDs in tasks
	for _, m := range mappings.TaskTypeMappings {
		_, err := tx.Exec("UPDATE task SET task_type_id = ? WHERE task_type_id = ?", m.RemoteID, m.LocalID)
		if err != nil {
			return fmt.Errorf("failed to remap task_type_id %s to %s: %w", m.LocalID, m.RemoteID, err)
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
func ReplaceTypeTables(tx *sqlx.Tx, remoteStatuses []models.Status, remoteEntityTypes []models.EntityType, remoteTaskTypes []models.TaskType, remoteRoles []models.Role) error {
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

	// Clear and replace entity types
	_, err = tx.Exec("DELETE FROM entity_type")
	if err != nil {
		return fmt.Errorf("failed to clear entity_type table: %w", err)
	}
	for _, et := range remoteEntityTypes {
		_, err := tx.Exec(
			"INSERT INTO entity_type (id, name, icon, synced, mtime) VALUES (?, ?, ?, 1, ?)",
			et.Id, et.Name, et.Icon, utils.GetEpochTime(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert entity_type %s: %w", et.Name, err)
		}
	}

	// Clear and replace task types
	_, err = tx.Exec("DELETE FROM task_type")
	if err != nil {
		return fmt.Errorf("failed to clear task_type table: %w", err)
	}
	for _, tt := range remoteTaskTypes {
		_, err := tx.Exec(
			"INSERT INTO task_type (id, name, icon, synced, mtime) VALUES (?, ?, ?, 1, ?)",
			tt.Id, tt.Name, tt.Icon, utils.GetEpochTime(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert task_type %s: %w", tt.Name, err)
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
				view_entity, create_entity, update_entity, delete_entity,
				view_task, create_task, update_task, delete_task,
				view_template, create_template, update_template, delete_template,
				view_checkpoint, create_checkpoint, delete_checkpoint,
				pull_chunk, assign_task, unassign_task,
				add_user, remove_user, change_role,
				change_status, set_done_task, set_retake_task, view_done_task, manage_dependencies
			) VALUES (?, ?, 1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			r.Id, r.Name, utils.GetEpochTime(),
			r.ViewEntity, r.CreateEntity, r.UpdateEntity, r.DeleteEntity,
			r.ViewTask, r.CreateTask, r.UpdateTask, r.DeleteTask,
			r.ViewTemplate, r.CreateTemplate, r.UpdateTemplate, r.DeleteTemplate,
			r.ViewCheckpoint, r.CreateCheckpoint, r.DeleteCheckpoint,
			r.PullChunk, r.AssignTask, r.UnassignTask,
			r.AddUser, r.RemoveUser, r.ChangeRole,
			r.ChangeStatus, r.SetDoneTask, r.SetRetakeTask, r.ViewDoneTask, r.ManageDependencies,
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
		"entity",
		"entity_assignee",
		"task",
		"task_checkpoint",
		"task_dependency",
		"entity_dependency",
		"dependency_type",
		"user",
		"role",
		"template",
		"workflow",
		"workflow_link",
		"workflow_entity",
		"workflow_task",
		"tag",
		"task_tag",
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
func FetchRemoteProjectTypes(remoteUrl string, userId string) ([]models.Status, []models.EntityType, []models.TaskType, []models.Role, error) {
	data, err := FetchData(remoteUrl, userId)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to fetch remote project data: %w", err)
	}

	return data.Statuses, data.EntityTypes, data.TaskTypes, data.Roles, nil
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
	localStatuses, localEntityTypes, localTaskTypes, localRoles, err := FetchLocalTypes(tx)
	if err != nil {
		return err
	}

	// Fetch remote types
	remoteStatuses, remoteEntityTypes, remoteTaskTypes, remoteRoles, err := FetchRemoteProjectTypes(remoteUrl, userId)
	if err != nil {
		return err
	}

	// Build mappings
	mappings := BuildTypeMappings(
		localStatuses, remoteStatuses,
		localEntityTypes, remoteEntityTypes,
		localTaskTypes, remoteTaskTypes,
		localRoles, remoteRoles,
	)

	// Remap IDs in tasks and entities
	err = RemapProjectIds(tx, mappings)
	if err != nil {
		return err
	}

	// Replace type tables with remote types
	err = ReplaceTypeTables(tx, remoteStatuses, remoteEntityTypes, remoteTaskTypes, remoteRoles)
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
