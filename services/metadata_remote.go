package services

import (
	"bytes"
	"clustta/internal/auth_service"
	"clustta/internal/constants"
	"clustta/internal/repository/models"
	repositorysync "clustta/internal/repository/sync_service"
	"clustta/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"io"
	"net/http"
	"time"
)

type MetadataUpdateResult struct {
	RemoteApplied bool `json:"remote_applied"`
	RequiresSync  bool `json:"requires_sync"`
}

// FetchResult identifies the assets whose working files were successfully
// restored by a fetch operation.
type FetchResult struct {
	RestoredAssetIds []string `json:"restored_asset_ids"`
}

type metadataTransportError struct {
	err error
}

func (e *metadataTransportError) Error() string { return e.err.Error() }
func (e *metadataTransportError) Unwrap() error { return e.err }

// IsMetadataTransportFailure reports whether no usable HTTP response was
// received. Only these failures may safely fall back to a local mutation.
func IsMetadataTransportFailure(err error) bool {
	var transportErr *metadataTransportError
	return errors.As(err, &transportErr)
}

type assetPatch struct {
	Id          string  `json:"id"`
	StatusId    *string `json:"status_id,omitempty"`
	AssigneeId  *string `json:"assignee_id,omitempty"`
	IsTask      *bool   `json:"is_task,omitempty"`
	AssetTypeId *string `json:"asset_type_id,omitempty"`
}
type assetPatchResponse struct {
	Assets            []models.Asset `json:"assets"`
	PreviousSyncToken *string        `json:"previous_sync_token"`
	SyncToken         string         `json:"sync_token"`
}
type collectionPatch struct {
	Id                string   `json:"id"`
	IsShared          *bool    `json:"is_shared,omitempty"`
	CollectionTypeId  *string  `json:"collection_type_id,omitempty"`
	AddAssigneeIds    []string `json:"add_assignee_ids,omitempty"`
	RemoveAssigneeIds []string `json:"remove_assignee_ids,omitempty"`
}
type collectionPatchResponse struct {
	Collections         []models.Collection         `json:"collections"`
	CollectionAssignees []models.CollectionAssignee `json:"collection_assignees"`
	PreviousSyncToken   *string                     `json:"previous_sync_token"`
	SyncToken           string                      `json:"sync_token"`
}

type typePutRequest struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type assetTypePutResponse struct {
	AssetType         models.AssetType `json:"asset_type"`
	PreviousSyncToken *string          `json:"previous_sync_token"`
	SyncToken         string           `json:"sync_token"`
}

type collectionTypePutResponse struct {
	CollectionType    models.CollectionType `json:"collection_type"`
	PreviousSyncToken *string               `json:"previous_sync_token"`
	SyncToken         string                `json:"sync_token"`
}

// PatchMetadataRemote sends an authenticated typed metadata PATCH request.
// It is exported so the sibling services_test package can verify the wire contract.
func PatchMetadataRemote(remoteURL, path string, payload, result any) error {
	return metadataRemoteRequest(http.MethodPatch, remoteURL, path, payload, result)
}

func metadataRemoteRequest(method, remoteURL, path string, payload, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, remoteURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	auth_service.AttachBearerToken(req)
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return &metadataTransportError{err: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		message, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remote metadata update failed (%d): %s", resp.StatusCode, string(message))
	}
	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}
	return nil
}

func putAssetTypeRemote(remoteURL, id, name, icon string) (assetTypePutResponse, error) {
	var out assetTypePutResponse
	err := metadataRemoteRequest(http.MethodPut, remoteURL, "/asset-types/"+id, typePutRequest{Name: name, Icon: icon}, &out)
	return out, err
}

func putCollectionTypeRemote(remoteURL, id, name, icon string) (collectionTypePutResponse, error) {
	var out collectionTypePutResponse
	err := metadataRemoteRequest(http.MethodPut, remoteURL, "/collection-types/"+id, typePutRequest{Name: name, Icon: icon}, &out)
	return out, err
}
func patchAssetsRemote(remoteURL string, patches []assetPatch) (assetPatchResponse, error) {
	var out assetPatchResponse
	err := PatchMetadataRemote(remoteURL, "/assets", map[string]any{"assets": patches}, &out)
	return out, err
}
func patchCollectionsRemote(remoteURL string, patches []collectionPatch) (collectionPatchResponse, error) {
	var out collectionPatchResponse
	err := PatchMetadataRemote(remoteURL, "/collections", map[string]any{"collections": patches}, &out)
	return out, err
}

func applyCanonicalAssets(tx *sqlx.Tx, assets []models.Asset) (bool, error) {
	ids := make([]string, 0, len(assets))
	for _, asset := range assets {
		ids = append(ids, asset.Id)
	}
	clean, err := metadataRowsAreSynced(tx, "asset", ids)
	if err != nil || !clean {
		return clean, err
	}
	for _, asset := range assets {
		_, err = tx.Exec(`UPDATE asset SET mtime=?, name=?, description=?, is_resource=?, status_id=?, asset_type_id=?, collection_id=?, assignee_id=?, assigner_id=?, is_link=?, pointer=?, preview_id=?, trashed=? WHERE id=?`, asset.MTime, asset.Name, asset.Description, asset.IsResource, asset.StatusId, asset.AssetTypeId, asset.CollectionId, asset.AssigneeId, asset.AssignerId, asset.IsLink, asset.Pointer, asset.PreviewId, asset.Trashed, asset.Id)
		if err != nil {
			return false, err
		}
	}
	if err = utils.SetRowsSynced(tx, "asset", ids); err != nil {
		return false, err
	}
	return true, nil
}

func applyCanonicalCollections(tx *sqlx.Tx, collections []models.Collection) (bool, error) {
	ids := make([]string, 0, len(collections))
	for _, collection := range collections {
		ids = append(ids, collection.Id)
	}
	clean, err := metadataRowsAreSynced(tx, "collection", ids)
	if err != nil || !clean {
		return clean, err
	}
	for _, collection := range collections {
		_, err = tx.Exec(`UPDATE collection SET mtime=?, name=?, description=?, collection_type_id=?, parent_id=?, preview_id=?, is_shared=?, trashed=? WHERE id=?`, collection.MTime, collection.Name, collection.Description, collection.CollectionTypeId, collection.ParentId, collection.PreviewId, collection.IsShared, collection.Trashed, collection.Id)
		if err != nil {
			return false, err
		}
	}
	if err = utils.SetRowsSynced(tx, "collection", ids); err != nil {
		return false, err
	}
	return true, nil
}

func metadataRowsAreSynced(tx *sqlx.Tx, table string, ids []string) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	query, args, err := sqlx.In(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id IN (?) AND synced=0", table), ids)
	if err != nil {
		return false, err
	}
	var count int
	if err = tx.Get(&count, tx.Rebind(query), args...); err != nil {
		return false, err
	}
	return count == 0, nil
}

// applyReturnedSyncToken advances the local token only when the response is
// continuous with the local state and no pending local data remains.
func applyReturnedSyncToken(tx *sqlx.Tx, previous *string, next string) error {
	if previous == nil || next == "" {
		return nil
	}
	unsynced, err := repositorysync.IsUnsynced(tx)
	if err != nil {
		return err
	}
	current, err := utils.GetProjectSyncToken(tx)
	if err != nil {
		return err
	}
	if unsynced {
		return nil
	}
	if current != *previous {
		return nil
	}
	if err = utils.SetProjectSyncToken(tx, next); err != nil {
		return err
	}
	return nil
}

func applyCanonicalAssetType(tx *sqlx.Tx, assetType models.AssetType) error {
	_, err := tx.Exec(`INSERT INTO asset_type(id,mtime,name,icon,synced) VALUES(?,?,?,?,1)
		ON CONFLICT(id) DO UPDATE SET mtime=excluded.mtime,name=excluded.name,icon=excluded.icon,synced=1`,
		assetType.Id, assetType.MTime, assetType.Name, assetType.Icon)
	if err != nil {
		return err
	}
	return utils.SetRowsSynced(tx, "asset_type", []string{assetType.Id})
}

func applyCanonicalCollectionType(tx *sqlx.Tx, collectionType models.CollectionType) error {
	_, err := tx.Exec(`INSERT INTO collection_type(id,mtime,name,icon,synced) VALUES(?,?,?,?,1)
		ON CONFLICT(id) DO UPDATE SET mtime=excluded.mtime,name=excluded.name,icon=excluded.icon,synced=1`,
		collectionType.Id, collectionType.MTime, collectionType.Name, collectionType.Icon)
	if err != nil {
		return err
	}
	return utils.SetRowsSynced(tx, "collection_type", []string{collectionType.Id})
}

func applyCanonicalCollectionAssignees(tx *sqlx.Tx, assignees []models.CollectionAssignee) error {
	ids := make([]string, 0, len(assignees))
	for _, assignee := range assignees {
		result, err := tx.Exec(`UPDATE collection_assignee SET mtime=?, assigner_id=? WHERE collection_id=? AND assignee_id=?`, assignee.MTime, assignee.AssignerId, assignee.CollectionId, assignee.AssigneeId)
		if err != nil {
			return err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if count == 0 {
			if _, err = tx.Exec(`INSERT INTO collection_assignee(id,mtime,collection_id,assignee_id,assigner_id,synced) VALUES(?,?,?,?,?,1)`, assignee.Id, assignee.MTime, assignee.CollectionId, assignee.AssigneeId, assignee.AssignerId); err != nil {
				return err
			}
		}
		var id string
		if err = tx.Get(&id, `SELECT id FROM collection_assignee WHERE collection_id=? AND assignee_id=?`, assignee.CollectionId, assignee.AssigneeId); err != nil {
			return err
		}
		ids = append(ids, id)
	}
	return utils.SetRowsSynced(tx, "collection_assignee", ids)
}
