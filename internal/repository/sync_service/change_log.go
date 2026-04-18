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
	Assets      []ChangeSummaryItem `json:"assets"`
	Collections   []ChangeSummaryItem `json:"collections"`
	Other      []ChangeSummaryItem `json:"other"`
	TotalCount int                 `json:"total_count"`
}

// LoadChangeSummary returns a lightweight summary of all unsynced rows grouped by category.
// This is much cheaper than LoadChangedData as it only fetches id, name, and mtime.
func LoadChangeSummary(tx *sqlx.Tx) (ChangeSummary, error) {
	summary := ChangeSummary{}

	// Assets (modified or newly created)
	assetItems := []ChangeSummaryItem{}
	err := tx.Select(&assetItems, "SELECT id, name, 'asset' AS source, mtime FROM asset WHERE synced = 0")
	if err != nil && err != sql.ErrNoRows {
		return summary, err
	}
	for i := range assetItems {
		assetItems[i].ChangeType = "modified"
	}
	summary.Assets = assetItems

	// Collections (modified or newly created)
	collectionItems := []ChangeSummaryItem{}
	err = tx.Select(&collectionItems, "SELECT id, name, 'collection' AS source, mtime FROM collection WHERE synced = 0")
	if err != nil && err != sql.ErrNoRows {
		return summary, err
	}
	for i := range collectionItems {
		collectionItems[i].ChangeType = "modified"
	}
	summary.Collections = collectionItems

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
		case "asset":
			summary.Assets = append(summary.Assets, item)
		case "collection":
			summary.Collections = append(summary.Collections, item)
		default:
			summary.Other = append(summary.Other, item)
		}
	}

	// Other changes — collection_assignee, asset_dependency, collection_dependency, asset_checkpoint, asset_tag
	otherQueries := []struct {
		query  string
		source string
	}{
		{"SELECT id, collection_id AS name, 'collection_assignee' AS source, mtime FROM collection_assignee WHERE synced = 0", "collection_assignee"},
		{"SELECT id, asset_id AS name, 'asset_dependency' AS source, mtime FROM asset_dependency WHERE synced = 0", "asset_dependency"},
		{"SELECT id, asset_id AS name, 'collection_dependency' AS source, mtime FROM collection_dependency WHERE synced = 0", "collection_dependency"},
		{"SELECT id, asset_id AS name, 'asset_checkpoint' AS source, mtime FROM asset_checkpoint WHERE synced = 0", "asset_checkpoint"},
		{"SELECT id, asset_id AS name, 'asset_tag' AS source, mtime FROM asset_tag WHERE synced = 0", "asset_tag"},
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

	summary.TotalCount = len(summary.Assets) + len(summary.Collections) + len(summary.Other)
	return summary, nil
}

// DiscardAssetChanges reverts a single asset and its related rows to the server state.
// It deletes the local asset and related unsynced rows, then re-inserts from server data.
func DiscardAssetChanges(tx *sqlx.Tx, serverData ProjectData, assetId string) error {
	// Delete local asset and related rows
	_, err := tx.Exec("DELETE FROM asset_dependency WHERE asset_id = ? OR dependency_id = ?", assetId, assetId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM collection_dependency WHERE asset_id = ?", assetId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM asset_tag WHERE asset_id = ?", assetId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM asset_checkpoint WHERE asset_id = ?", assetId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM asset WHERE id = ?", assetId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM tomb WHERE id = ? AND table_name = 'asset'", assetId)
	if err != nil {
		return err
	}

	// Re-insert from server data
	for _, asset := range serverData.Assets {
		if asset.Id == assetId {
			_, err = tx.Exec(`INSERT INTO asset (id, mtime, created_at, name, description, extension, is_resource, is_link, pointer, status_id, asset_type_id, collection_id, assignee_id, assigner_id, preview_id, trashed, synced) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,1)`,
				asset.Id, asset.MTime, asset.CreatedAt, asset.Name, asset.Description, asset.Extension, asset.IsResource, asset.IsLink, asset.Pointer, asset.StatusId, asset.AssetTypeId, asset.CollectionId, asset.AssigneeId, asset.AssignerId, asset.PreviewId, asset.Trashed)
			if err != nil {
				return err
			}
			break
		}
	}

	// Re-insert related checkpoints from server
	for _, cp := range serverData.AssetCheckpoints {
		if cp.AssetId == assetId {
			_, err = tx.Exec(`INSERT OR IGNORE INTO asset_checkpoint (id, mtime, created_at, asset_id, xxhash_checksum, time_modified, file_size, comment, chunks, author_id, preview_id, group_id, trashed, synced) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,1)`,
				cp.Id, cp.MTime, cp.CreatedAt, cp.AssetId, cp.XXHashChecksum, cp.TimeModified, cp.FileSize, cp.Comment, cp.Chunks, cp.AuthorUID, cp.PreviewId, cp.GroupId, cp.Trashed)
			if err != nil {
				return err
			}
		}
	}

	// Re-insert related asset dependencies from server
	for _, dep := range serverData.AssetDependencies {
		if dep.AssetId == assetId || dep.DependencyId == assetId {
			_, err = tx.Exec(`INSERT OR IGNORE INTO asset_dependency (id, mtime, asset_id, dependency_id, dependency_type_id, synced) VALUES (?,?,?,?,?,1)`,
				dep.Id, dep.MTime, dep.AssetId, dep.DependencyId, dep.DependencyTypeId)
			if err != nil {
				return err
			}
		}
	}

	// Re-insert related collection dependencies from server
	for _, dep := range serverData.CollectionDependencies {
		if dep.AssetId == assetId {
			_, err = tx.Exec(`INSERT OR IGNORE INTO collection_dependency (id, mtime, asset_id, dependency_id, dependency_type_id, synced) VALUES (?,?,?,?,?,1)`,
				dep.Id, dep.MTime, dep.AssetId, dep.DependencyId, dep.DependencyTypeId)
			if err != nil {
				return err
			}
		}
	}

	// Re-insert related asset tags from server
	for _, tt := range serverData.AssetTags {
		if tt.AssetId == assetId {
			_, err = tx.Exec(`INSERT OR IGNORE INTO asset_tag (id, mtime, asset_id, tag_id, synced) VALUES (?,?,?,?,1)`,
				tt.Id, tt.MTime, tt.AssetId, tt.TagId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DiscardCollectionChanges reverts a single collection and its related rows to the server state.
func DiscardCollectionChanges(tx *sqlx.Tx, serverData ProjectData, collectionId string) error {
	// Delete local collection and related rows
	_, err := tx.Exec("DELETE FROM collection_assignee WHERE collection_id = ?", collectionId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM collection WHERE id = ?", collectionId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM tomb WHERE id = ? AND table_name = 'collection'", collectionId)
	if err != nil {
		return err
	}

	// Re-insert from server data
	for _, collection := range serverData.Collections {
		if collection.Id == collectionId {
			_, err = tx.Exec(`INSERT INTO collection (id, mtime, created_at, name, description, collection_type_id, parent_id, trashed, preview_id, is_shared, synced) VALUES (?,?,?,?,?,?,?,?,?,?,1)`,
				collection.Id, collection.MTime, collection.CreatedAt, collection.Name, collection.Description, collection.CollectionTypeId, collection.ParentId, collection.Trashed, collection.PreviewId, collection.IsShared)
			if err != nil {
				return err
			}
			break
		}
	}

	// Re-insert collection assignees from server
	for _, ea := range serverData.CollectionAssignees {
		if ea.CollectionId == collectionId {
			_, err = tx.Exec(`INSERT OR IGNORE INTO collection_assignee (id, mtime, collection_id, assignee_id, assigner_id, synced) VALUES (?,?,?,?,?,1)`,
				ea.Id, ea.MTime, ea.CollectionId, ea.AssigneeId, ea.AssignerId)
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
	// Collect all unsynced asset IDs
	unsyncedAssetIds := []string{}
	err := tx.Select(&unsyncedAssetIds, "SELECT id FROM asset WHERE synced = 0")
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Collect all unsynced collection IDs
	unsyncedCollectionIds := []string{}
	err = tx.Select(&unsyncedCollectionIds, "SELECT id FROM collection WHERE synced = 0")
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Collect tomb IDs for assets and collections
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
		case "asset":
			unsyncedAssetIds = append(unsyncedAssetIds, tomb.ID)
		case "collection":
			unsyncedCollectionIds = append(unsyncedCollectionIds, tomb.ID)
		}
	}

	// Discard each asset
	for _, assetId := range unsyncedAssetIds {
		err = DiscardAssetChanges(tx, serverData, assetId)
		if err != nil {
			return err
		}
	}

	// Discard each collection
	for _, collectionId := range unsyncedCollectionIds {
		err = DiscardCollectionChanges(tx, serverData, collectionId)
		if err != nil {
			return err
		}
	}

	// Clear remaining unsynced rows in other tables by resetting to synced
	otherTables := []string{
		"collection_assignee", "asset_dependency", "collection_dependency",
		"asset_checkpoint", "asset_tag", "user", "role", "status",
		"dependency_type", "asset_type", "collection_type",
		"template", "workflow", "workflow_link", "workflow_collection", "workflow_asset", "tag",
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
	// The assets and collections are already handled above, so we write the rest
	err = writeOtherServerData(tx, serverData)
	if err != nil {
		return err
	}

	return nil
}

// writeOtherServerData inserts non-asset/collection server data for tables cleared during discard.
func writeOtherServerData(tx *sqlx.Tx, data ProjectData) error {
	for _, user := range data.Users {
		_, err := tx.Exec(`INSERT OR IGNORE INTO user (id, mtime, added_at, first_name, last_name, username, email, photo, role_id, synced) VALUES (?,?,?,?,?,?,?,?,?,1)`,
			user.Id, user.MTime, user.AddedAt, user.FirstName, user.LastName, user.Username, user.Email, user.Photo, user.RoleId)
		if err != nil {
			return err
		}
	}
	for _, role := range data.Roles {
		_, err := tx.Exec(`INSERT OR IGNORE INTO role (id, name, mtime, view_collection, create_collection, update_collection, delete_collection, view_asset, create_asset, update_asset, delete_asset, view_template, create_template, update_template, delete_template, view_checkpoint, create_checkpoint, delete_checkpoint, pull_chunk, assign_asset, unassign_asset, add_user, remove_user, change_role, change_status, set_done_asset, set_retake_asset, view_done_asset, manage_dependencies, synced) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,1)`,
			role.Id, role.Name, role.MTime, role.ViewCollection, role.CreateCollection, role.UpdateCollection, role.DeleteCollection, role.ViewAsset, role.CreateAsset, role.UpdateAsset, role.DeleteAsset, role.ViewTemplate, role.CreateTemplate, role.UpdateTemplate, role.DeleteTemplate, role.ViewCheckpoint, role.CreateCheckpoint, role.DeleteCheckpoint, role.PullChunk, role.AssignAsset, role.UnassignAsset, role.AddUser, role.RemoveUser, role.ChangeRole, role.ChangeStatus, role.SetDoneAsset, role.SetRetakeAsset, role.ViewDoneAsset, role.ManageDependencies)
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
	for _, tt := range data.AssetTags {
		_, err := tx.Exec(`INSERT OR IGNORE INTO asset_tag (id, mtime, asset_id, tag_id, synced) VALUES (?,?,?,?,1)`,
			tt.Id, tt.MTime, tt.AssetId, tt.TagId)
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
	for _, et := range data.CollectionTypes {
		_, err := tx.Exec(`INSERT OR IGNORE INTO collection_type (id, mtime, name, icon, synced) VALUES (?,?,?,?,1)`,
			et.Id, et.MTime, et.Name, et.Icon)
		if err != nil {
			return err
		}
	}
	for _, tt := range data.AssetTypes {
		_, err := tx.Exec(`INSERT OR IGNORE INTO asset_type (id, mtime, name, icon, synced) VALUES (?,?,?,?,1)`,
			tt.Id, tt.MTime, tt.Name, tt.Icon)
		if err != nil {
			return err
		}
	}
	for _, ea := range data.CollectionAssignees {
		_, err := tx.Exec(`INSERT OR IGNORE INTO collection_assignee (id, mtime, collection_id, assignee_id, assigner_id, synced) VALUES (?,?,?,?,?,1)`,
			ea.Id, ea.MTime, ea.CollectionId, ea.AssigneeId, ea.AssignerId)
		if err != nil {
			return err
		}
	}
	for _, dep := range data.AssetDependencies {
		_, err := tx.Exec(`INSERT OR IGNORE INTO asset_dependency (id, mtime, asset_id, dependency_id, dependency_type_id, synced) VALUES (?,?,?,?,?,1)`,
			dep.Id, dep.MTime, dep.AssetId, dep.DependencyId, dep.DependencyTypeId)
		if err != nil {
			return err
		}
	}
	for _, dep := range data.CollectionDependencies {
		_, err := tx.Exec(`INSERT OR IGNORE INTO collection_dependency (id, mtime, asset_id, dependency_id, dependency_type_id, synced) VALUES (?,?,?,?,?,1)`,
			dep.Id, dep.MTime, dep.AssetId, dep.DependencyId, dep.DependencyTypeId)
		if err != nil {
			return err
		}
	}
	for _, cp := range data.AssetCheckpoints {
		_, err := tx.Exec(`INSERT OR IGNORE INTO asset_checkpoint (id, mtime, created_at, asset_id, xxhash_checksum, time_modified, file_size, comment, chunks, author_id, preview_id, group_id, trashed, synced) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,1)`,
			cp.Id, cp.MTime, cp.CreatedAt, cp.AssetId, cp.XXHashChecksum, cp.TimeModified, cp.FileSize, cp.Comment, cp.Chunks, cp.AuthorUID, cp.PreviewId, cp.GroupId, cp.Trashed)
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
		_, err := tx.Exec(`INSERT OR IGNORE INTO workflow_link (id, name, collection_type_id, workflow_id, linked_workflow_id, mtime, synced) VALUES (?,?,?,?,?,?,1)`,
			wfl.Id, wfl.Name, wfl.CollectionTypeId, wfl.WorkflowId, wfl.LinkedWorkflowId, wfl.MTime)
		if err != nil {
			return err
		}
	}
	for _, wfe := range data.WorkflowCollections {
		_, err := tx.Exec(`INSERT OR IGNORE INTO workflow_collection (id, name, workflow_id, collection_type_id, mtime, synced) VALUES (?,?,?,?,?,1)`,
			wfe.Id, wfe.Name, wfe.WorkflowId, wfe.CollectionTypeId, wfe.MTime)
		if err != nil {
			return err
		}
	}
	for _, wft := range data.WorkflowAssets {
		_, err := tx.Exec(`INSERT OR IGNORE INTO workflow_asset (id, name, workflow_id, asset_type_id, is_resource, template_id, pointer, is_link, mtime, synced) VALUES (?,?,?,?,?,?,?,?,?,1)`,
			wft.Id, wft.Name, wft.WorkflowId, wft.AssetTypeId, wft.IsResource, wft.TemplateId, wft.Pointer, wft.IsLink, wft.MTime)
		if err != nil {
			return err
		}
	}
	return nil
}
