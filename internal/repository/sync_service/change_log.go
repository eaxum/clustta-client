package sync_service

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// ChangeSummaryItem represents a single unsynced change for the changelog UI.
type ChangeSummaryItem struct {
	ID         string `db:"id" json:"id"`
	Name       string `db:"name" json:"name"`
	Source     string `db:"source" json:"source"`
	ChangeType string `json:"change_type"`
	Mtime      int    `db:"mtime" json:"mtime"`
}

// ChangeSummary groups all pending changes by category for frontend display.
type ChangeSummary struct {
	Tasks      []ChangeSummaryItem `json:"tasks"`
	Entities   []ChangeSummaryItem `json:"entities"`
	Other      []ChangeSummaryItem `json:"other"`
	TotalCount int                 `json:"total_count"`
}

// LoadChangeSummary returns a lightweight summary of all unsynced rows grouped by category.
// This is much cheaper than LoadChangedData as it only fetches id, name, and mtime.
func LoadChangeSummary(tx *sqlx.Tx) (ChangeSummary, error) {
	summary := ChangeSummary{}

	// Tasks (modified or newly created)
	taskItems := []ChangeSummaryItem{}
	err := tx.Select(&taskItems, "SELECT id, name, 'task' AS source, mtime FROM task WHERE synced = 0")
	if err != nil && err != sql.ErrNoRows {
		return summary, err
	}
	for i := range taskItems {
		taskItems[i].ChangeType = "modified"
	}
	summary.Tasks = taskItems

	// Entities (modified or newly created)
	entityItems := []ChangeSummaryItem{}
	err = tx.Select(&entityItems, "SELECT id, name, 'entity' AS source, mtime FROM entity WHERE synced = 0")
	if err != nil && err != sql.ErrNoRows {
		return summary, err
	}
	for i := range entityItems {
		entityItems[i].ChangeType = "modified"
	}
	summary.Entities = entityItems

	// Tombs (deleted items) — categorize by table_name
	type tombRow struct {
		ID        string `db:"id"`
		TableName string `db:"table_name"`
		Mtime     int    `db:"mtime"`
	}
	tombItems := []tombRow{}
	err = tx.Select(&tombItems, "SELECT id, table_name, mtime FROM tomb WHERE synced = 0")
	if err != nil && err != sql.ErrNoRows {
		return summary, err
	}
	for _, tomb := range tombItems {
		item := ChangeSummaryItem{
			ID:         tomb.ID,
			Name:       tomb.ID,
			Source:     tomb.TableName,
			ChangeType: "deleted",
			Mtime:      tomb.Mtime,
		}
		switch tomb.TableName {
		case "task":
			summary.Tasks = append(summary.Tasks, item)
		case "entity":
			summary.Entities = append(summary.Entities, item)
		default:
			summary.Other = append(summary.Other, item)
		}
	}

	// Other changes — entity_assignee, task_dependency, entity_dependency, task_checkpoint, task_tag
	otherQueries := []struct {
		query  string
		source string
	}{
		{"SELECT id, entity_id AS name, 'entity_assignee' AS source, mtime FROM entity_assignee WHERE synced = 0", "entity_assignee"},
		{"SELECT id, task_id AS name, 'task_dependency' AS source, mtime FROM task_dependency WHERE synced = 0", "task_dependency"},
		{"SELECT id, task_id AS name, 'entity_dependency' AS source, mtime FROM entity_dependency WHERE synced = 0", "entity_dependency"},
		{"SELECT id, task_id AS name, 'task_checkpoint' AS source, mtime FROM task_checkpoint WHERE synced = 0", "task_checkpoint"},
		{"SELECT id, task_id AS name, 'task_tag' AS source, mtime FROM task_tag WHERE synced = 0", "task_tag"},
		{"SELECT id, username AS name, 'user' AS source, mtime FROM user WHERE synced = 0", "user"},
		{"SELECT id, name, 'role' AS source, mtime FROM role WHERE synced = 0", "role"},
		{"SELECT id, name, 'status' AS source, mtime FROM status WHERE synced = 0", "status"},
		{"SELECT id, name, 'workflow' AS source, mtime FROM workflow WHERE synced = 0", "workflow"},
	}
	for _, q := range otherQueries {
		items := []ChangeSummaryItem{}
		err = tx.Select(&items, q.query)
		if err != nil && err != sql.ErrNoRows {
			return summary, err
		}
		for i := range items {
			items[i].ChangeType = "modified"
		}
		summary.Other = append(summary.Other, items...)
	}

	summary.TotalCount = len(summary.Tasks) + len(summary.Entities) + len(summary.Other)
	return summary, nil
}

// DiscardTaskChanges reverts a single task and its related rows to the server state.
// It deletes the local task and related unsynced rows, then re-inserts from server data.
func DiscardTaskChanges(tx *sqlx.Tx, serverData ProjectData, taskId string) error {
	// Delete local task and related rows
	_, err := tx.Exec("DELETE FROM task_dependency WHERE task_id = ? OR dependency_id = ?", taskId, taskId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM entity_dependency WHERE task_id = ?", taskId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM task_tag WHERE task_id = ?", taskId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM task_checkpoint WHERE task_id = ?", taskId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM task WHERE id = ?", taskId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM tomb WHERE id = ? AND table_name = 'task'", taskId)
	if err != nil {
		return err
	}

	// Re-insert from server data
	for _, task := range serverData.Tasks {
		if task.Id == taskId {
			_, err = tx.Exec(`INSERT INTO task (id, mtime, created_at, name, description, extension, is_resource, is_link, pointer, status_id, task_type_id, entity_id, assignee_id, assigner_id, preview_id, trashed, synced) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,1)`,
				task.Id, task.MTime, task.CreatedAt, task.Name, task.Description, task.Extension, task.IsResource, task.IsLink, task.Pointer, task.StatusId, task.TaskTypeId, task.EntityId, task.AssigneeId, task.AssignerId, task.PreviewId, task.Trashed)
			if err != nil {
				return err
			}
			break
		}
	}

	// Re-insert related checkpoints from server
	for _, cp := range serverData.TasksCheckpoints {
		if cp.TaskId == taskId {
			_, err = tx.Exec(`INSERT OR IGNORE INTO task_checkpoint (id, mtime, created_at, task_id, xxhash_checksum, time_modified, file_size, comment, chunks, author_id, preview_id, group_id, trashed, synced) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,1)`,
				cp.Id, cp.MTime, cp.CreatedAt, cp.TaskId, cp.XXHashChecksum, cp.TimeModified, cp.FileSize, cp.Comment, cp.Chunks, cp.AuthorUID, cp.PreviewId, cp.GroupId, cp.Trashed)
			if err != nil {
				return err
			}
		}
	}

	// Re-insert related task dependencies from server
	for _, dep := range serverData.TaskDependencies {
		if dep.TaskId == taskId || dep.DependencyId == taskId {
			_, err = tx.Exec(`INSERT OR IGNORE INTO task_dependency (id, mtime, task_id, dependency_id, dependency_type_id, synced) VALUES (?,?,?,?,?,1)`,
				dep.Id, dep.MTime, dep.TaskId, dep.DependencyId, dep.DependencyTypeId)
			if err != nil {
				return err
			}
		}
	}

	// Re-insert related entity dependencies from server
	for _, dep := range serverData.EntityDependencies {
		if dep.TaskId == taskId {
			_, err = tx.Exec(`INSERT OR IGNORE INTO entity_dependency (id, mtime, task_id, dependency_id, dependency_type_id, synced) VALUES (?,?,?,?,?,1)`,
				dep.Id, dep.MTime, dep.TaskId, dep.DependencyId, dep.DependencyTypeId)
			if err != nil {
				return err
			}
		}
	}

	// Re-insert related task tags from server
	for _, tt := range serverData.TasksTags {
		if tt.TaskId == taskId {
			_, err = tx.Exec(`INSERT OR IGNORE INTO task_tag (id, mtime, task_id, tag_id, synced) VALUES (?,?,?,?,1)`,
				tt.Id, tt.MTime, tt.TaskId, tt.TagId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DiscardEntityChanges reverts a single entity and its related rows to the server state.
func DiscardEntityChanges(tx *sqlx.Tx, serverData ProjectData, entityId string) error {
	// Delete local entity and related rows
	_, err := tx.Exec("DELETE FROM entity_assignee WHERE entity_id = ?", entityId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM entity WHERE id = ?", entityId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM tomb WHERE id = ? AND table_name = 'entity'", entityId)
	if err != nil {
		return err
	}

	// Re-insert from server data
	for _, entity := range serverData.Entities {
		if entity.Id == entityId {
			_, err = tx.Exec(`INSERT INTO entity (id, mtime, created_at, name, description, entity_type_id, parent_id, trashed, preview_id, is_library, synced) VALUES (?,?,?,?,?,?,?,?,?,?,1)`,
				entity.Id, entity.MTime, entity.CreatedAt, entity.Name, entity.Description, entity.EntityTypeId, entity.ParentId, entity.Trashed, entity.PreviewId, entity.IsLibrary)
			if err != nil {
				return err
			}
			break
		}
	}

	// Re-insert entity assignees from server
	for _, ea := range serverData.EntityAssignees {
		if ea.EntityId == entityId {
			_, err = tx.Exec(`INSERT OR IGNORE INTO entity_assignee (id, mtime, entity_id, assignee_id, assigner_id, synced) VALUES (?,?,?,?,?,1)`,
				ea.Id, ea.MTime, ea.EntityId, ea.AssigneeId, ea.AssignerId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DiscardAllChanges resets all unsynced rows to the server state using selective replacement.
// This is more efficient than ClearLocalDataDrop as it only touches dirty rows.
func DiscardAllChanges(tx *sqlx.Tx, serverData ProjectData) error {
	// Collect all unsynced task IDs
	unsyncedTaskIds := []string{}
	err := tx.Select(&unsyncedTaskIds, "SELECT id FROM task WHERE synced = 0")
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Collect all unsynced entity IDs
	unsyncedEntityIds := []string{}
	err = tx.Select(&unsyncedEntityIds, "SELECT id FROM entity WHERE synced = 0")
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Collect tomb IDs for tasks and entities
	type tombRow struct {
		ID        string `db:"id"`
		TableName string `db:"table_name"`
	}
	tombItems := []tombRow{}
	err = tx.Select(&tombItems, "SELECT id, table_name FROM tomb WHERE synced = 0")
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	for _, tomb := range tombItems {
		switch tomb.TableName {
		case "task":
			unsyncedTaskIds = append(unsyncedTaskIds, tomb.ID)
		case "entity":
			unsyncedEntityIds = append(unsyncedEntityIds, tomb.ID)
		}
	}

	// Discard each task
	for _, taskId := range unsyncedTaskIds {
		err = DiscardTaskChanges(tx, serverData, taskId)
		if err != nil {
			return err
		}
	}

	// Discard each entity
	for _, entityId := range unsyncedEntityIds {
		err = DiscardEntityChanges(tx, serverData, entityId)
		if err != nil {
			return err
		}
	}

	// Clear remaining unsynced rows in other tables by resetting to synced
	otherTables := []string{
		"entity_assignee", "task_dependency", "entity_dependency",
		"task_checkpoint", "task_tag", "user", "role", "status",
		"dependency_type", "task_type", "entity_type",
		"template", "workflow", "workflow_link", "workflow_entity", "workflow_task", "tag",
	}
	for _, table := range otherTables {
		_, err = tx.Exec("DELETE FROM " + table + " WHERE synced = 0")
		if err != nil {
			return err
		}
	}

	// Clear all tombs
	_, err = tx.Exec("DELETE FROM tomb")
	if err != nil {
		return err
	}

	// Re-write all server data for other tables using WriteProjectData
	// The tasks and entities are already handled above, so we write the rest
	err = writeOtherServerData(tx, serverData)
	if err != nil {
		return err
	}

	return nil
}

// writeOtherServerData inserts non-task/entity server data for tables cleared during discard.
func writeOtherServerData(tx *sqlx.Tx, data ProjectData) error {
	for _, user := range data.Users {
		_, err := tx.Exec(`INSERT OR IGNORE INTO user (id, mtime, added_at, first_name, last_name, username, email, photo, role_id, synced) VALUES (?,?,?,?,?,?,?,?,?,1)`,
			user.Id, user.MTime, user.AddedAt, user.FirstName, user.LastName, user.Username, user.Email, user.Photo, user.RoleId)
		if err != nil {
			return err
		}
	}
	for _, role := range data.Roles {
		_, err := tx.Exec(`INSERT OR IGNORE INTO role (id, name, mtime, view_entity, create_entity, update_entity, delete_entity, view_task, create_task, update_task, delete_task, view_template, create_template, update_template, delete_template, view_checkpoint, create_checkpoint, delete_checkpoint, pull_chunk, assign_task, unassign_task, add_user, remove_user, change_role, change_status, set_done_task, set_retake_task, view_done_task, manage_dependencies, synced) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,1)`,
			role.Id, role.Name, role.MTime, role.ViewEntity, role.CreateEntity, role.UpdateEntity, role.DeleteEntity, role.ViewTask, role.CreateTask, role.UpdateTask, role.DeleteTask, role.ViewTemplate, role.CreateTemplate, role.UpdateTemplate, role.DeleteTemplate, role.ViewCheckpoint, role.CreateCheckpoint, role.DeleteCheckpoint, role.PullChunk, role.AssignTask, role.UnassignTask, role.AddUser, role.RemoveUser, role.ChangeRole, role.ChangeStatus, role.SetDoneTask, role.SetRetakeTask, role.ViewDoneTask, role.ManageDependencies)
		if err != nil {
			return err
		}
	}
	for _, status := range data.Statuses {
		_, err := tx.Exec(`INSERT OR IGNORE INTO status (id, mtime, name, short_name, color, synced) VALUES (?,?,?,?,?,1)`,
			status.Id, status.MTime, status.Name, status.ShortName, status.Color)
		if err != nil {
			return err
		}
	}
	for _, tag := range data.Tags {
		_, err := tx.Exec(`INSERT OR IGNORE INTO tag (id, mtime, name, synced) VALUES (?,?,?,1)`,
			tag.Id, tag.MTime, tag.Name)
		if err != nil {
			return err
		}
	}
	for _, tt := range data.TasksTags {
		_, err := tx.Exec(`INSERT OR IGNORE INTO task_tag (id, mtime, task_id, tag_id, synced) VALUES (?,?,?,?,1)`,
			tt.Id, tt.MTime, tt.TaskId, tt.TagId)
		if err != nil {
			return err
		}
	}
	for _, dt := range data.DependencyTypes {
		_, err := tx.Exec(`INSERT OR IGNORE INTO dependency_type (id, mtime, name, synced) VALUES (?,?,?,1)`,
			dt.Id, dt.MTime, dt.Name)
		if err != nil {
			return err
		}
	}
	for _, et := range data.EntityTypes {
		_, err := tx.Exec(`INSERT OR IGNORE INTO entity_type (id, mtime, name, icon, synced) VALUES (?,?,?,?,1)`,
			et.Id, et.MTime, et.Name, et.Icon)
		if err != nil {
			return err
		}
	}
	for _, tt := range data.TaskTypes {
		_, err := tx.Exec(`INSERT OR IGNORE INTO task_type (id, mtime, name, icon, synced) VALUES (?,?,?,?,1)`,
			tt.Id, tt.MTime, tt.Name, tt.Icon)
		if err != nil {
			return err
		}
	}
	for _, ea := range data.EntityAssignees {
		_, err := tx.Exec(`INSERT OR IGNORE INTO entity_assignee (id, mtime, entity_id, assignee_id, assigner_id, synced) VALUES (?,?,?,?,?,1)`,
			ea.Id, ea.MTime, ea.EntityId, ea.AssigneeId, ea.AssignerId)
		if err != nil {
			return err
		}
	}
	for _, dep := range data.TaskDependencies {
		_, err := tx.Exec(`INSERT OR IGNORE INTO task_dependency (id, mtime, task_id, dependency_id, dependency_type_id, synced) VALUES (?,?,?,?,?,1)`,
			dep.Id, dep.MTime, dep.TaskId, dep.DependencyId, dep.DependencyTypeId)
		if err != nil {
			return err
		}
	}
	for _, dep := range data.EntityDependencies {
		_, err := tx.Exec(`INSERT OR IGNORE INTO entity_dependency (id, mtime, task_id, dependency_id, dependency_type_id, synced) VALUES (?,?,?,?,?,1)`,
			dep.Id, dep.MTime, dep.TaskId, dep.DependencyId, dep.DependencyTypeId)
		if err != nil {
			return err
		}
	}
	for _, cp := range data.TasksCheckpoints {
		_, err := tx.Exec(`INSERT OR IGNORE INTO task_checkpoint (id, mtime, created_at, task_id, xxhash_checksum, time_modified, file_size, comment, chunks, author_id, preview_id, group_id, trashed, synced) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,1)`,
			cp.Id, cp.MTime, cp.CreatedAt, cp.TaskId, cp.XXHashChecksum, cp.TimeModified, cp.FileSize, cp.Comment, cp.Chunks, cp.AuthorUID, cp.PreviewId, cp.GroupId, cp.Trashed)
		if err != nil {
			return err
		}
	}
	for _, tmpl := range data.Templates {
		_, err := tx.Exec(`INSERT OR IGNORE INTO template (id, mtime, name, extension, xxhash_checksum, file_size, chunks, trashed, synced) VALUES (?,?,?,?,?,?,?,?,1)`,
			tmpl.Id, tmpl.MTime, tmpl.Name, tmpl.Extension, tmpl.XxhashChecksum, tmpl.FileSize, tmpl.Chunks, tmpl.Trashed)
		if err != nil {
			return err
		}
	}
	for _, wf := range data.Workflows {
		_, err := tx.Exec(`INSERT OR IGNORE INTO workflow (id, name, mtime, synced) VALUES (?,?,?,1)`,
			wf.Id, wf.Name, wf.MTime)
		if err != nil {
			return err
		}
	}
	for _, wfl := range data.WorkflowLinks {
		_, err := tx.Exec(`INSERT OR IGNORE INTO workflow_link (id, name, entity_type_id, workflow_id, linked_workflow_id, mtime, synced) VALUES (?,?,?,?,?,?,1)`,
			wfl.Id, wfl.Name, wfl.EntityTypeId, wfl.WorkflowId, wfl.LinkedWorkflowId, wfl.MTime)
		if err != nil {
			return err
		}
	}
	for _, wfe := range data.WorkflowEntities {
		_, err := tx.Exec(`INSERT OR IGNORE INTO workflow_entity (id, name, workflow_id, entity_type_id, mtime, synced) VALUES (?,?,?,?,?,1)`,
			wfe.Id, wfe.Name, wfe.WorkflowId, wfe.EntityTypeId, wfe.MTime)
		if err != nil {
			return err
		}
	}
	for _, wft := range data.WorkflowTasks {
		_, err := tx.Exec(`INSERT OR IGNORE INTO workflow_task (id, name, workflow_id, task_type_id, is_resource, template_id, pointer, is_link, mtime, synced) VALUES (?,?,?,?,?,?,?,?,?,1)`,
			wft.Id, wft.Name, wft.WorkflowId, wft.TaskTypeId, wft.IsResource, wft.TemplateId, wft.Pointer, wft.IsLink, wft.MTime)
		if err != nil {
			return err
		}
	}
	return nil
}
