package sync_service

import (
	"clustta/internal/utils"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// ChangeChild represents a nested sub-change belonging to a parent asset or collection.
type ChangeChild struct {
	ID          string `json:"id"`
	ParentID    string `json:"parent_id"`
	RefID       string `json:"ref_id,omitempty"`
	Source      string `json:"source"`
	Description string `json:"description"`
	ChangeType  string `json:"change_type"`
}

// ChangeSummaryItem represents a single unsynced change for the changelog UI.
type ChangeSummaryItem struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Icon       string        `json:"icon,omitempty"`
	Extension  string        `json:"extension,omitempty"`
	Source     string        `json:"source"`
	ChangeType string        `json:"change_type"`
	Mtime      int           `json:"mtime"`
	Children   []ChangeChild `json:"children,omitempty"`
}

// ChangeSummary groups all pending changes by category for frontend display.
type ChangeSummary struct {
	Assets      []ChangeSummaryItem `json:"assets"`
	Collections []ChangeSummaryItem `json:"collections"`
	Other       []ChangeSummaryItem `json:"other"`
	TotalCount  int                 `json:"total_count"`
}

// changeSummaryRow is used internally for scanning asset/collection queries that include created_at.
type changeSummaryRow struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	Icon      string `db:"icon"`
	Extension string `db:"extension"`
	Source    string `db:"source"`
	Mtime     int    `db:"mtime"`
	CreatedAt string `db:"created_at"`
	Trashed   bool   `db:"trashed"`
}

// childRow is used internally for scanning child table queries.
type childRow struct {
	ID          string `db:"id"`
	ParentID    string `db:"parent_id"`
	RefID       string `db:"ref_id"`
	Source      string `db:"source"`
	Description string `db:"description"`
	Trashed     bool   `db:"trashed"`
}

// classifyChangeType returns "deleted" if trashed, "new" if created after last sync, otherwise "modified".
func classifyChangeType(trashed bool, createdAt string, lastSyncTime int64) string {
	if trashed {
		return "deleted"
	}
	createdEpoch, err := utils.RFC3339ToEpoch(createdAt)
	if err != nil {
		return "modified"
	}
	if createdEpoch > lastSyncTime {
		return "new"
	}
	return "modified"
}

// classifyCheckpointChangeType returns "deleted" if trashed, otherwise "added".
func classifyCheckpointChangeType(trashed bool) string {
	if trashed {
		return "deleted"
	}
	return "added"
}

// lookupAssetInfo returns the name, type icon, and extension for an asset by ID.
func lookupAssetInfo(tx *sqlx.Tx, assetID string) (string, string, string) {
	var info struct {
		Name      string `db:"name"`
		Icon      string `db:"icon"`
		Extension string `db:"extension"`
	}
	err := tx.Get(&info,
		`SELECT a.name, COALESCE(at.icon, '') AS icon, a.extension
		 FROM asset a LEFT JOIN asset_type at ON a.asset_type_id = at.id WHERE a.id = ?`, assetID)
	if err != nil {
		return assetID, "", ""
	}
	return info.Name, info.Icon, info.Extension
}

// lookupCollectionInfo returns the name and type icon for a collection by ID.
func lookupCollectionInfo(tx *sqlx.Tx, collectionID string) (string, string) {
	var info struct {
		Name string `db:"name"`
		Icon string `db:"icon"`
	}
	err := tx.Get(&info,
		`SELECT c.name, COALESCE(ct.icon, '') AS icon
		 FROM collection c LEFT JOIN collection_type ct ON c.collection_type_id = ct.id WHERE c.id = ?`, collectionID)
	if err != nil {
		return collectionID, ""
	}
	return info.Name, info.Icon
}

// LoadChangeSummary returns a lightweight summary of all unsynced rows grouped by category.
// Asset and collection children (checkpoints, dependencies, tags, assignees) are nested
// under their parent item. "Other" only includes templates.
func LoadChangeSummary(tx *sqlx.Tx) (ChangeSummary, error) {
	summary := ChangeSummary{}

	// Get last sync time to distinguish new vs modified items
	lastSyncTime, err := utils.GetLastSyncTime(tx)
	if err != nil {
		lastSyncTime = 0
	}

	// --- Assets ---
	assetRows := []changeSummaryRow{}
	err = tx.Select(&assetRows,
		`SELECT a.id, a.name, COALESCE(at.icon, '') AS icon, a.extension, 'asset' AS source, a.mtime, a.created_at, a.trashed
		 FROM asset a LEFT JOIN asset_type at ON a.asset_type_id = at.id WHERE a.synced = 0`)
	if err != nil && err != sql.ErrNoRows {
		return summary, err
	}
	assetMap := make(map[string]*ChangeSummaryItem, len(assetRows))
	for _, row := range assetRows {
		item := ChangeSummaryItem{
			ID: row.ID, Name: row.Name, Icon: row.Icon, Extension: row.Extension, Source: row.Source, Mtime: row.Mtime,
			ChangeType: classifyChangeType(row.Trashed, row.CreatedAt, lastSyncTime),
		}
		assetMap[row.ID] = &item
	}

	// Batch-query all asset child tables and attach to parents
	// Checkpoints (have trashed field)
	checkpointRows := []childRow{}
	err = tx.Select(&checkpointRows,
		`SELECT ac.id, ac.asset_id AS parent_id, '' AS ref_id, 'asset_checkpoint' AS source, ac.created_at AS description, ac.trashed
		 FROM asset_checkpoint ac WHERE ac.synced = 0`)
	if err != nil && err != sql.ErrNoRows {
		return summary, err
	}
	for _, row := range checkpointRows {
		child := ChangeChild{
			ID: row.ID, ParentID: row.ParentID, Source: row.Source,
			Description: row.Description,
			ChangeType:  classifyCheckpointChangeType(row.Trashed),
		}
		if parent, ok := assetMap[row.ParentID]; ok {
			parent.Children = append(parent.Children, child)
		} else {
			name, icon, ext := lookupAssetInfo(tx, row.ParentID)
			container := &ChangeSummaryItem{
				ID: row.ParentID, Name: name, Icon: icon, Extension: ext, Source: "asset",
				ChangeType: "modified",
				Children:   []ChangeChild{child},
			}
			assetMap[row.ParentID] = container
		}
	}

	// Dependencies and tags (no trashed field — live rows are "added")
	depQueries := []string{
		`SELECT ad.id, ad.asset_id AS parent_id, ad.dependency_id AS ref_id, 'asset_dependency' AS source,
		 COALESCE(a2.name, ad.dependency_id) AS description, 0 AS trashed
		 FROM asset_dependency ad LEFT JOIN asset a2 ON ad.dependency_id = a2.id WHERE ad.synced = 0`,
		`SELECT at2.id, at2.asset_id AS parent_id, at2.tag_id AS ref_id, 'asset_tag' AS source,
		 COALESCE(t.name, at2.tag_id) AS description, 0 AS trashed
		 FROM asset_tag at2 LEFT JOIN tag t ON at2.tag_id = t.id WHERE at2.synced = 0`,
		`SELECT cd.id, cd.asset_id AS parent_id, cd.dependency_id AS ref_id, 'collection_dependency' AS source,
		 COALESCE(c2.name, cd.dependency_id) AS description, 0 AS trashed
		 FROM collection_dependency cd LEFT JOIN collection c2 ON cd.dependency_id = c2.id WHERE cd.synced = 0`,
	}
	for _, q := range depQueries {
		rows := []childRow{}
		err = tx.Select(&rows, q)
		if err != nil && err != sql.ErrNoRows {
			return summary, err
		}
		for _, row := range rows {
			child := ChangeChild{
				ID: row.ID, ParentID: row.ParentID, RefID: row.RefID, Source: row.Source,
				Description: row.Description,
				ChangeType:  "added",
			}
			if parent, ok := assetMap[row.ParentID]; ok {
				parent.Children = append(parent.Children, child)
			} else {
				name, icon, ext := lookupAssetInfo(tx, row.ParentID)
				container := &ChangeSummaryItem{
					ID: row.ParentID, Name: name, Icon: icon, Extension: ext, Source: "asset",
					ChangeType: "modified",
					Children:   []ChangeChild{child},
				}
				assetMap[row.ParentID] = container
			}
		}
	}
	for _, item := range assetMap {
		summary.Assets = append(summary.Assets, *item)
	}

	// --- Collections ---
	collectionRows := []changeSummaryRow{}
	err = tx.Select(&collectionRows,
		`SELECT c.id, c.name, COALESCE(ct.icon, '') AS icon, '' AS extension, 'collection' AS source, c.mtime, c.created_at, c.trashed
		 FROM collection c LEFT JOIN collection_type ct ON c.collection_type_id = ct.id WHERE c.synced = 0`)
	if err != nil && err != sql.ErrNoRows {
		return summary, err
	}
	collectionMap := make(map[string]*ChangeSummaryItem, len(collectionRows))
	for _, row := range collectionRows {
		item := ChangeSummaryItem{
			ID: row.ID, Name: row.Name, Icon: row.Icon, Source: row.Source, Mtime: row.Mtime,
			ChangeType: classifyChangeType(row.Trashed, row.CreatedAt, lastSyncTime),
		}
		collectionMap[row.ID] = &item
	}

	// Batch-query all collection child tables and attach to parents
	collectionChildQueries := []string{
		`SELECT ca.id, ca.collection_id AS parent_id, ca.assignee_id AS ref_id, 'collection_assignee' AS source,
		 COALESCE(u.first_name || ' ' || u.last_name, ca.assignee_id) AS description, 0 AS trashed
		 FROM collection_assignee ca LEFT JOIN user u ON ca.assignee_id = u.id WHERE ca.synced = 0`,
	}
	for _, q := range collectionChildQueries {
		rows := []childRow{}
		err = tx.Select(&rows, q)
		if err != nil && err != sql.ErrNoRows {
			return summary, err
		}
		for _, row := range rows {
			child := ChangeChild{
				ID: row.ID, ParentID: row.ParentID, RefID: row.RefID, Source: row.Source,
				Description: row.Description,
				ChangeType:  "added",
			}
			if parent, ok := collectionMap[row.ParentID]; ok {
				parent.Children = append(parent.Children, child)
			} else {
				name, icon := lookupCollectionInfo(tx, row.ParentID)
				container := &ChangeSummaryItem{
					ID: row.ParentID, Name: name, Icon: icon, Source: "collection",
					ChangeType: "modified",
					Children:   []ChangeChild{child},
				}
				collectionMap[row.ParentID] = container
			}
		}
	}
	for _, item := range collectionMap {
		summary.Collections = append(summary.Collections, *item)
	}

	// --- Tombstones ---
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
		switch tomb.TableName {
		case "asset":
			if _, ok := assetMap[tomb.ID]; !ok {
				summary.Assets = append(summary.Assets, ChangeSummaryItem{
					ID: tomb.ID, Name: tomb.ID, Source: "asset",
					ChangeType: "deleted", Mtime: tomb.Mtime,
				})
			}
		case "collection":
			if _, ok := collectionMap[tomb.ID]; !ok {
				summary.Collections = append(summary.Collections, ChangeSummaryItem{
					ID: tomb.ID, Name: tomb.ID, Source: "collection",
					ChangeType: "deleted", Mtime: tomb.Mtime,
				})
			}
		case "template":
			summary.Other = append(summary.Other, ChangeSummaryItem{
				ID: tomb.ID, Name: tomb.ID, Source: "template",
				ChangeType: "deleted", Mtime: tomb.Mtime,
			})
		}
	}

	// --- Other: Templates only ---
	type templateRow struct {
		ID      string `db:"id"`
		Name    string `db:"name"`
		Mtime   int    `db:"mtime"`
		Trashed bool   `db:"trashed"`
	}
	templateRows := []templateRow{}
	err = tx.Select(&templateRows, "SELECT id, name, mtime, trashed FROM template WHERE synced = 0")
	if err != nil && err != sql.ErrNoRows {
		return summary, err
	}
	for _, row := range templateRows {
		changeType := "modified"
		if row.Trashed {
			changeType = "deleted"
		}
		summary.Other = append(summary.Other, ChangeSummaryItem{
			ID: row.ID, Name: row.Name, Source: "template",
			ChangeType: changeType, Mtime: row.Mtime,
		})
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
