package sync_service

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// ConflictInfo contains details about a sync conflict (matches server response)
type ConflictInfo struct {
	Type       string `json:"type"`        // "collection" or "asset"
	LocalId    string `json:"local_id"`    // ID client tried to push
	ExistingId string `json:"existing_id"` // ID that already exists on server
	Name       string `json:"name"`        // The conflicting name
	ParentId   string `json:"parent_id"`   // Parent collection ID (or collection_id for assets)
	Extension  string `json:"extension"`   // For assets only
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

// ResolveConflicts updates local collection and asset IDs to match the server's existing IDs.
// This merges the local items with the server's items by adopting the server's IDs.
func ResolveConflicts(tx *sqlx.Tx, conflicts []ConflictInfo) error {
	for _, conflict := range conflicts {
		switch conflict.Type {
		case "collection":
			err := remapCollectionId(tx, conflict.LocalId, conflict.ExistingId)
			if err != nil {
				return fmt.Errorf("failed to remap collection %s: %w", conflict.LocalId, err)
			}
			fmt.Printf("[RESOLVE] Collection '%s': remapped ID %s -> %s\n",
				conflict.Name, conflict.LocalId, conflict.ExistingId)

		case "asset":
			err := remapAssetId(tx, conflict.LocalId, conflict.ExistingId)
			if err != nil {
				return fmt.Errorf("failed to remap asset %s: %w", conflict.LocalId, err)
			}
			fmt.Printf("[RESOLVE] Asset '%s': remapped ID %s -> %s\n",
				conflict.Name, conflict.LocalId, conflict.ExistingId)
		}
	}
	return nil
}

// remapCollectionId changes an collection's ID and updates all foreign key references
func remapCollectionId(tx *sqlx.Tx, oldId, newId string) error {
	now := time.Now().Unix()

	_, err := tx.Exec(`UPDATE collection SET parent_id = ?, mtime = ?, synced = 0 WHERE parent_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update child collections: %w", err)
	}

	_, err = tx.Exec(`UPDATE asset SET collection_id = ?, mtime = ?, synced = 0 WHERE collection_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update assets: %w", err)
	}

	_, err = tx.Exec(`UPDATE collection_assignee SET collection_id = ?, mtime = ?, synced = 0 WHERE collection_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update collection_assignee: %w", err)
	}

	_, err = tx.Exec(`UPDATE collection_dependency SET dependency_id = ?, mtime = ?, synced = 0 WHERE dependency_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update collection_dependency: %w", err)
	}

	_, err = tx.Exec(`UPDATE workflow_collection SET collection_type_id = ?, mtime = ?, synced = 0 WHERE collection_type_id = ?`,
		newId, now, oldId)

	_, err = tx.Exec(`UPDATE collection SET id = ?, mtime = ?, synced = 0 WHERE id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update collection id: %w", err)
	}
	if _, err = tx.Exec(`UPDATE pending_path_update SET entity_id = ?
		WHERE entity_type = 'collection' AND entity_id = ?`, newId, oldId); err != nil {
		return fmt.Errorf("failed to update pending collection path: %w", err)
	}

	return nil
}

// remapAssetId changes a asset's ID and updates all foreign key references
func remapAssetId(tx *sqlx.Tx, oldId, newId string) error {
	now := time.Now().Unix()

	_, err := tx.Exec(`UPDATE asset_checkpoint SET asset_id = ?, mtime = ?, synced = 0 WHERE asset_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update checkpoints: %w", err)
	}

	_, err = tx.Exec(`UPDATE asset_dependency SET asset_id = ?, mtime = ?, synced = 0 WHERE asset_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update asset_dependency (asset_id): %w", err)
	}

	_, err = tx.Exec(`UPDATE asset_dependency SET dependency_id = ?, mtime = ?, synced = 0 WHERE dependency_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update asset_dependency (dependency_id): %w", err)
	}

	_, err = tx.Exec(`UPDATE collection_dependency SET asset_id = ?, mtime = ?, synced = 0 WHERE asset_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update collection_dependency: %w", err)
	}

	_, err = tx.Exec(`UPDATE asset_tag SET asset_id = ?, mtime = ?, synced = 0 WHERE asset_id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update asset_tag: %w", err)
	}

	_, err = tx.Exec(`UPDATE asset SET id = ?, mtime = ?, synced = 0 WHERE id = ?`,
		newId, now, oldId)
	if err != nil {
		return fmt.Errorf("failed to update asset id: %w", err)
	}
	if _, err = tx.Exec(`UPDATE pending_path_update SET entity_id = ?
		WHERE entity_type = 'asset' AND entity_id = ?`, newId, oldId); err != nil {
		return fmt.Errorf("failed to update pending asset path: %w", err)
	}

	return nil
}
