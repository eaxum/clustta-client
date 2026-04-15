package sync_service

import (
	"bytes"
	"clustta/internal/auth_service"
	"clustta/internal/chunk_service"
	"clustta/internal/constants"
	"clustta/internal/repository"
	"clustta/internal/repository/repositorypb"
	"clustta/internal/studio_service"
	"clustta/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/DataDog/zstd"
	"google.golang.org/protobuf/proto"
)

// legacyServerVersion is the version threshold below which the server uses legacy table names.
const legacyServerVersion = "0.4.25"

// shouldUseLegacyNames checks the server version and returns true if tomb table names
// need to be remapped to legacy names (task/entity instead of asset/collection).
func shouldUseLegacyNames(remoteUrl string) bool {
	if !utils.IsValidURL(remoteUrl) {
		return false
	}
	version, err := studio_service.GetServerVersion(remoteUrl)
	if err != nil || version == "" {
		return true // default to legacy if we can't determine
	}
	return version != legacyServerVersion
}

func PushData(ctx context.Context, projectPath, remoteUrl string, userId string, callback func(int, int, string, string)) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = repository.ClearTrash(tx)
	if err != nil {
		return err
	}

	data, err := LoadChangedData(tx)
	if err != nil {
		return err
	}
	if data.IsEmpty() {
		return nil
	}

	pdData := repositorypb.ProjectData{
		ProjectPreview:      data.ProjectPreview,
		CollectionTypes:     repository.ToPbCollectionTypes(data.CollectionTypes),
		Collections:         repository.ToPbCollections(data.Collections),
		CollectionAssignees: repository.ToPbCollectionAssignees(data.CollectionAssignees),

		AssetTypes:             repository.ToPbAssetTypes(data.AssetTypes),
		Assets:                 repository.ToPbAssets(data.Assets),
		AssetCheckpoints:       repository.ToPbCheckpoints(data.AssetCheckpoints),
		AssetDependencies:      repository.ToPbAssetDependencies(data.AssetDependencies),
		CollectionDependencies: repository.ToPbCollectionDependencies(data.CollectionDependencies),

		Statuses:        repository.ToPbStatuses(data.Statuses),
		DependencyTypes: repository.ToPbDependencyTypes(data.DependencyTypes),

		Users: repository.ToPbUsers(data.Users),
		Roles: repository.ToPbRoles(data.Roles),

		Templates: repository.ToPbTemplates(data.Templates),

		Workflows:           repository.ToPbWorkflows(data.Workflows),
		WorkflowLinks:       repository.ToPbWorkflowLinks(data.WorkflowLinks),
		WorkflowCollections: repository.ToPbWorkflowCollections(data.WorkflowCollections),
		WorkflowAssets:      repository.ToPbWorkflowAssets(data.WorkflowAssets),

		Tags:      repository.ToPbTags(data.Tags),
		AssetTags: repository.ToPbAssetTags(data.AssetTags),

		Tomb: repository.ToPbTombs(data.Tombs),

		IntegrationProjects:           repository.ToPbIntegrationProjects(data.IntegrationProjects),
		IntegrationCollectionMappings: repository.ToPbIntegrationCollectionMappings(data.IntegrationCollectionMappings),
		IntegrationAssetMappings:      repository.ToPbIntegrationAssetMappings(data.IntegrationAssetMappings),
	}

	// if shouldUseLegacyNames(remoteUrl) {
	// 	repository.RemapTombsToLegacyNames(pdData.Tomb)
	// }

	dataByte, err := proto.Marshal(&pdData)
	if err != nil {
		return err
	}

	compressedData, err := zstd.CompressLevel(nil, dataByte, 3)
	if err != nil {
		return err
	}

	chunks := []string{}
	chunkSet := make(map[string]bool)
	for _, AssetCheckpoint := range data.AssetCheckpoints {
		chunksString := AssetCheckpoint.Chunks
		chunkHashes := strings.Split(chunksString, ",")
		for _, chunkHash := range chunkHashes {
			if !chunkSet[chunkHash] {
				chunkSet[chunkHash] = true
				chunks = append(chunks, chunkHash)
			}
		}

	}
	for _, Template := range data.Templates {
		chunksString := Template.Chunks
		chunkHashes := strings.Split(chunksString, ",")
		for _, chunkHash := range chunkHashes {
			if !chunkSet[chunkHash] {
				chunkSet[chunkHash] = true
				chunks = append(chunks, chunkHash)
			}
		}

	}

	remoteMissingChunks, err := FetchMissingChunks(ctx, remoteUrl, userId, chunks)
	if err != nil {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if len(remoteMissingChunks) > 0 {
		remoteMissingChunksInfo, err := chunk_service.GetChunksInfo(tx, remoteMissingChunks)
		if err != nil {
			return err
		}
		err = chunk_service.PushChunksBatch(ctx, tx, remoteUrl, userId, remoteMissingChunksInfo, callback)
		if err != nil {
			return err
		}
	}

	previewIds := []string{}
	previewSet := make(map[string]bool)
	if data.ProjectPreview != "" && !previewSet[data.ProjectPreview] {
		previewSet[data.ProjectPreview] = true
		previewIds = append(previewIds, data.ProjectPreview)
	}
	for _, asset := range data.Assets {
		if asset.PreviewId != "" && !previewSet[asset.PreviewId] {
			previewSet[asset.PreviewId] = true
			previewIds = append(previewIds, asset.PreviewId)
		}
	}
	for _, collection := range data.Collections {
		if collection.PreviewId != "" && !previewSet[collection.PreviewId] {
			previewSet[collection.PreviewId] = true
			previewIds = append(previewIds, collection.PreviewId)
		}
	}
	for _, assetCheckpoint := range data.AssetCheckpoints {
		if assetCheckpoint.PreviewId != "" && !previewSet[assetCheckpoint.PreviewId] {
			previewSet[assetCheckpoint.PreviewId] = true
			previewIds = append(previewIds, assetCheckpoint.PreviewId)
		}
	}

	remoteMissingPreviews, err := FetchMissingPreviews(ctx, remoteUrl, userId, previewIds)
	if err != nil {
		return err
	}

	if len(remoteMissingPreviews) > 0 {
		err = repository.PushPreviews(tx, remoteUrl, userId, remoteMissingPreviews, callback)
		if err != nil {
			return err
		}
	}

	if utils.IsValidURL(remoteUrl) {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		dataUrl := remoteUrl + "/data"

		req, err := http.NewRequestWithContext(ctx, "POST", dataUrl, bytes.NewBuffer(compressedData))
		if err != nil {
			return err
		}
		req.Header.Set("Clustta-Agent", constants.USER_AGENT)
		auth_service.AttachBearerToken(req)

		client := &http.Client{
			Timeout: 10 * time.Minute,
		}
		response, err := client.Do(req)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		responseCode := response.StatusCode
		switch responseCode {
		case 200:
			err = utils.SetTablesToSynced(tx, ProjectTables)
			if err != nil {
				return err
			}
			err = tx.Commit()
			if err != nil {
				return err
			}
			return nil
		case 409:
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return fmt.Errorf("failed to read conflict response: %w", err)
			}

			var result WriteResult
			if err := json.Unmarshal(body, &result); err != nil {
				return fmt.Errorf("failed to parse conflict response: %w", err)
			}

			return &SyncConflictError{Conflicts: result.Conflicts}
		default:
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return err
			}
			return errors.New(string(body))
		}
	} else if utils.FileExists(remoteUrl) {
		db, err := utils.OpenDb(remoteUrl)
		if err != nil {
			return err
		}
		defer db.Close()
		remoteTx, err := db.Beginx()
		if err != nil {
			return err
		}
		err = WriteProjectData(remoteTx, data, true)
		if err != nil {
			return err
		}
		err = remoteTx.Commit()
		if err != nil {
			remoteTx.Rollback()
			return err
		}

		err = utils.SetTablesToSynced(tx, ProjectTables)
		if err != nil {
			return err
		}
		err = tx.Commit()
		if err != nil {
			return err
		}

		return nil
	} else {
		return fmt.Errorf("invalid url:%s", remoteUrl)
	}
}

// PushPartialData pushes a pre-built ProjectData to the server without loading from the DB.
// It marks only the specified row IDs as synced on success. On failure it returns silently
// so the rows stay unsynced and get picked up by the next full sync.
func PushPartialData(projectPath, remoteUrl, userId string, data ProjectData, syncTargets map[string][]string) error {
	if data.IsEmpty() {
		return nil
	}
	if !utils.IsValidURL(remoteUrl) {
		return nil
	}

	pdData := repositorypb.ProjectData{
		ProjectPreview:      data.ProjectPreview,
		CollectionTypes:     repository.ToPbCollectionTypes(data.CollectionTypes),
		Collections:         repository.ToPbCollections(data.Collections),
		CollectionAssignees: repository.ToPbCollectionAssignees(data.CollectionAssignees),

		AssetTypes:             repository.ToPbAssetTypes(data.AssetTypes),
		Assets:                 repository.ToPbAssets(data.Assets),
		AssetCheckpoints:       repository.ToPbCheckpoints(data.AssetCheckpoints),
		AssetDependencies:      repository.ToPbAssetDependencies(data.AssetDependencies),
		CollectionDependencies: repository.ToPbCollectionDependencies(data.CollectionDependencies),

		Statuses:        repository.ToPbStatuses(data.Statuses),
		DependencyTypes: repository.ToPbDependencyTypes(data.DependencyTypes),

		Users: repository.ToPbUsers(data.Users),
		Roles: repository.ToPbRoles(data.Roles),

		Templates: repository.ToPbTemplates(data.Templates),

		Workflows:           repository.ToPbWorkflows(data.Workflows),
		WorkflowLinks:       repository.ToPbWorkflowLinks(data.WorkflowLinks),
		WorkflowCollections: repository.ToPbWorkflowCollections(data.WorkflowCollections),
		WorkflowAssets:      repository.ToPbWorkflowAssets(data.WorkflowAssets),

		Tags:      repository.ToPbTags(data.Tags),
		AssetTags: repository.ToPbAssetTags(data.AssetTags),

		Tomb: repository.ToPbTombs(data.Tombs),

		IntegrationProjects:           repository.ToPbIntegrationProjects(data.IntegrationProjects),
		IntegrationCollectionMappings: repository.ToPbIntegrationCollectionMappings(data.IntegrationCollectionMappings),
		IntegrationAssetMappings:      repository.ToPbIntegrationAssetMappings(data.IntegrationAssetMappings),
	}

	if shouldUseLegacyNames(remoteUrl) {
		repository.RemapTombsToLegacyNames(pdData.Tomb)
	}

	dataByte, err := proto.Marshal(&pdData)
	if err != nil {
		return err
	}
	compressedData, err := zstd.CompressLevel(nil, dataByte, 3)
	if err != nil {
		return err
	}

	dataUrl := remoteUrl + "/data"
	req, err := http.NewRequest("POST", dataUrl, bytes.NewBuffer(compressedData))
	if err != nil {
		return err
	}
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	auth_service.AttachBearerToken(req)

	client := &http.Client{Timeout: 30 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("write-through push returned %d", response.StatusCode)
	}

	// Mark only the specific rows as synced
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for table, ids := range syncTargets {
		err = utils.SetRowsSynced(tx, table, ids)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// PushAssetData loads a single asset and its checkpoints, uploads their chunks and previews,
// then pushes the metadata to the server. On success it marks only the pushed rows as synced.
func PushAssetData(projectPath, remoteUrl, userId, assetId string, callback func(int, int, string, string)) error {
	if !utils.IsValidURL(remoteUrl) {
		return fmt.Errorf("invalid remote URL: %s", remoteUrl)
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	data, err := LoadAssetData(tx, assetId)
	if err != nil {
		return err
	}
	if data.IsEmpty() {
		return nil
	}

	// Collect chunk hashes from checkpoints
	chunks := []string{}
	chunkSet := make(map[string]bool)
	for _, cp := range data.AssetCheckpoints {
		if cp.Chunks == "" {
			continue
		}
		for _, hash := range strings.Split(cp.Chunks, ",") {
			if !chunkSet[hash] {
				chunkSet[hash] = true
				chunks = append(chunks, hash)
			}
		}
	}

	// Upload missing chunks
	if len(chunks) > 0 {
		remoteMissing, err := FetchMissingChunks(context.Background(), remoteUrl, userId, chunks)
		if err != nil {
			return err
		}
		if len(remoteMissing) > 0 {
			chunkInfos, err := chunk_service.GetChunksInfo(tx, remoteMissing)
			if err != nil {
				return err
			}
			err = chunk_service.PushChunksBatch(context.Background(), tx, remoteUrl, userId, chunkInfos, callback)
			if err != nil {
				return err
			}
		}
	}

	// Collect and upload missing previews
	previewIds := []string{}
	previewSet := make(map[string]bool)
	for _, asset := range data.Assets {
		if asset.PreviewId != "" && !previewSet[asset.PreviewId] {
			previewSet[asset.PreviewId] = true
			previewIds = append(previewIds, asset.PreviewId)
		}
	}
	for _, cp := range data.AssetCheckpoints {
		if cp.PreviewId != "" && !previewSet[cp.PreviewId] {
			previewSet[cp.PreviewId] = true
			previewIds = append(previewIds, cp.PreviewId)
		}
	}

	if len(previewIds) > 0 {
		remoteMissingPreviews, err := FetchMissingPreviews(context.Background(), remoteUrl, userId, previewIds)
		if err != nil {
			return err
		}
		if len(remoteMissingPreviews) > 0 {
			err = repository.PushPreviews(tx, remoteUrl, userId, remoteMissingPreviews, callback)
			if err != nil {
				return err
			}
		}
	}

	// Serialize and push metadata
	pdData := repositorypb.ProjectData{
		Assets:           repository.ToPbAssets(data.Assets),
		AssetCheckpoints: repository.ToPbCheckpoints(data.AssetCheckpoints),
	}

	dataByte, err := proto.Marshal(&pdData)
	if err != nil {
		return err
	}
	compressedData, err := zstd.CompressLevel(nil, dataByte, 3)
	if err != nil {
		return err
	}

	dataUrl := remoteUrl + "/data"
	req, err := http.NewRequest("POST", dataUrl, bytes.NewBuffer(compressedData))
	if err != nil {
		return err
	}
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	auth_service.AttachBearerToken(req)

	client := &http.Client{Timeout: 5 * time.Minute}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case 200:
		// Mark only the specific rows as synced
		syncTargets := map[string][]string{}
		for _, t := range data.Assets {
			syncTargets["asset"] = append(syncTargets["asset"], t.Id)
		}
		for _, cp := range data.AssetCheckpoints {
			syncTargets["asset_checkpoint"] = append(syncTargets["asset_checkpoint"], cp.Id)
		}
		for table, ids := range syncTargets {
			err = utils.SetRowsSynced(tx, table, ids)
			if err != nil {
				return err
			}
		}
		return tx.Commit()
	case 409:
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("failed to read conflict response: %w", err)
		}
		var result WriteResult
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("failed to parse conflict response: %w", err)
		}
		return &SyncConflictError{Conflicts: result.Conflicts}
	default:
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		return errors.New(string(body))
	}
}
