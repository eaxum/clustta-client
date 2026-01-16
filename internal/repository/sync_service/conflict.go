package sync_service

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// ConflictInfo contains details about a sync conflict (matches server response)
type ConflictInfo struct {
	Type       string `json:"type"`        // "entity" or "task"
	LocalId    string `json:"local_id"`    // ID client tried to push
	ExistingId string `json:"existing_id"` // ID that already exists on server
	Name       string `json:"name"`        // The conflicting name
	ParentId   string `json:"parent_id"`   // Parent entity ID (or entity_id for tasks)
	Extension  string `json:"extension"`   // For tasks only
}

// WriteResult contains the result of a write operation (matches server response)
type WriteResult struct {
	Success   bool           `json:"success"`
	Conflicts []ConflictInfo `json:"conflicts,omitempty"`
}

// SyncConflictError is returned when conflicts are detected during push
type SyncConflictError struct {
	Conflicts []ConflictInfo
}

func (e *SyncConflictError) Error() string {
	return fmt.Sprintf("sync conflict: %d items need to be merged", len(e.Conflicts))
}

// ResolveConflicts updates local entity and task IDs to match the server's existing IDs.
// This merges the local items with the server's items by adopting the server's IDs.
func ResolveConflicts(tx *sqlx.Tx, conflicts []ConflictInfo) error {
	for _, conflict := range conflicts {
		switch conflict.Type {
		case "entity":
			err := remapEntityId(tx, conflict.LocalId, conflict.ExistingId)
			if err != nil {
				return fmt.Errorf("failed to remap entity %s: %w", conflict.LocalId, err)
			}
			fmt.Printf("[RESOLVE] Entity '%s': remapped ID %s -> %s\n",
				conflict.Name, conflict.LocalId, conflict.ExistingId)

		case "task":
			err := remapTaskId(tx, conflict.LocalId, conflict.ExistingId)
			if err != nil {
				return fmt.Errorf("failed to remap task %s: %w", conflict.LocalId, err)
			}
			fmt.Printf("[RESOLVE] Task '%s': remapped ID %s -> %s\n",
				conflict.Name, conflict.LocalId, conflict.ExistingId)
		}
	}
	return nil
}

// remapEntityId changes an entity's ID and updates all foreign key references
func remapEntityId(tx *sqlx.Tx, oldId, newId string) error {
	now := time.Now().Unix()

	_, err := tx.Exec(`UPDATE entity SET parent_id = ?, mtime = ?, synced = 0 WHERE parent_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update child entities: %w", err)
	}

	_, err = tx.Exec(`UPDATE task SET entity_id = ?, mtime = ?, synced = 0 WHERE entity_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update tasks: %w", err)
	}

	_, err = tx.Exec(`UPDATE entity_assignee SET entity_id = ?, mtime = ?, synced = 0 WHERE entity_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update entity_assignee: %w", err)
	}

	_, err = tx.Exec(`UPDATE entity_dependency SET dependency_id = ?, mtime = ?, synced = 0 WHERE dependency_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update entity_dependency: %w", err)
	}

	_, err = tx.Exec(`UPDATE workflow_entity SET entity_type_id = ?, mtime = ?, synced = 0 WHERE entity_type_id = ?`,
		newId, now, oldId)

	_, err = tx.Exec(`UPDATE entity SET id = ?, mtime = ?, synced = 0 WHERE id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update entity id: %w", err)
	}

	return nil
}

// remapTaskId changes a task's ID and updates all foreign key references
func remapTaskId(tx *sqlx.Tx, oldId, newId string) error {
	now := time.Now().Unix()

	_, err := tx.Exec(`UPDATE task_checkpoint SET task_id = ?, mtime = ?, synced = 0 WHERE task_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update checkpoints: %w", err)
	}

	_, err = tx.Exec(`UPDATE task_dependency SET task_id = ?, mtime = ?, synced = 0 WHERE task_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update task_dependency (task_id): %w", err)
	}

	_, err = tx.Exec(`UPDATE task_dependency SET dependency_id = ?, mtime = ?, synced = 0 WHERE dependency_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update task_dependency (dependency_id): %w", err)
	}

	_, err = tx.Exec(`UPDATE entity_dependency SET task_id = ?, mtime = ?, synced = 0 WHERE task_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update entity_dependency: %w", err)
	}

	_, err = tx.Exec(`UPDATE task_tag SET task_id = ?, mtime = ?, synced = 0 WHERE task_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update task_tag: %w", err)
	}

	_, err = tx.Exec(`UPDATE task SET id = ?, mtime = ?, synced = 0 WHERE id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update task id: %w", err)
	}

	return nil
}
