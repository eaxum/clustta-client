package services

import (
	"bytes"
	"clustta/internal/auth_service"
	"clustta/internal/constants"
	"clustta/internal/error_service"
	"clustta/internal/ignore"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/repositorypb"
	"clustta/internal/utils"
	"clustta/output"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"compress/zlib"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/wailsapp/wails/v3/pkg/application"
	"google.golang.org/protobuf/proto"
)

type AssetService struct {
}

type assetAction string

const (
	assetActionCreate             assetAction = "create_asset"
	assetActionUpdate             assetAction = "update_asset"
	assetActionDelete             assetAction = "delete_asset"
	assetActionAssign             assetAction = "assign_asset"
	assetActionUnassign           assetAction = "unassign_asset"
	assetActionManageDependencies assetAction = "manage_dependencies"
	assetActionChangeStatus       assetAction = "change_status"
)

func activeAssetRole(tx *sqlx.Tx) (models.User, models.Role, error) {
	activeUser, err := auth_service.GetActiveUser()
	if err != nil {
		return models.User{}, models.Role{}, err
	}
	user, err := repository.GetUser(tx, activeUser.Id)
	if err != nil {
		return models.User{}, models.Role{}, err
	}
	role, err := repository.GetRole(tx, user.RoleId)
	if err != nil {
		return models.User{}, models.Role{}, err
	}
	return user, role, nil
}

func roleAllowsAssetAction(role models.Role, action assetAction) bool {
	switch action {
	case assetActionCreate:
		return role.CreateAsset
	case assetActionUpdate:
		return role.UpdateAsset
	case assetActionDelete:
		return role.DeleteAsset
	case assetActionAssign:
		return role.AssignAsset
	case assetActionUnassign:
		return role.UnassignAsset
	case assetActionManageDependencies:
		return role.ManageDependencies
	case assetActionChangeStatus:
		return role.ChangeStatus
	default:
		return false
	}
}

// authorizeAssetActionTx requires both the role capability and, for nested
// assets, either direct assignment or modify scope inherited from the parent.
func authorizeAssetActionTx(tx *sqlx.Tx, action assetAction, assetIds []string) error {
	user, role, err := activeAssetRole(tx)
	if err != nil {
		return err
	}
	if !roleAllowsAssetAction(role, action) {
		return fmt.Errorf("user does not have %s permission", action)
	}
	if role.Name == "admin" {
		return nil
	}
	allowedCollectionIds, err := repository.GetUserCanModifyCollectionIds(tx, user.Id)
	if err != nil {
		return err
	}
	for _, assetId := range assetIds {
		asset, err := repository.GetAsset(tx, assetId)
		if err != nil {
			return err
		}
		if action != assetActionCreate && asset.AssigneeId == user.Id {
			continue
		}
		if asset.CollectionId == "" {
			continue
		}
		if _, allowed := allowedCollectionIds[asset.CollectionId]; !allowed {
			return fmt.Errorf("user cannot %s asset outside assigned collection scope", action)
		}
	}
	return nil
}

func authorizeAssetCollectionTx(tx *sqlx.Tx, action assetAction, collectionId string) error {
	user, role, err := activeAssetRole(tx)
	if err != nil {
		return err
	}
	if !roleAllowsAssetAction(role, action) {
		return fmt.Errorf("user does not have %s permission", action)
	}
	if role.Name == "admin" {
		return nil
	}
	if collectionId == "" {
		return fmt.Errorf("user cannot %s assets at project root", action)
	}
	allowedCollectionIds, err := repository.GetUserCanModifyCollectionIds(tx, user.Id)
	if err != nil {
		return err
	}
	if _, allowed := allowedCollectionIds[collectionId]; !allowed {
		return fmt.Errorf("user cannot %s assets outside assigned collection scope", action)
	}
	return nil
}

func authorizeAssetAction(projectPath string, action assetAction, assetIds []string) error {
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
	return authorizeAssetActionTx(tx, action, assetIds)
}

type ChangedFiles struct {
	Modifieds  []string `json:"modified"`
	Untrackeds []string `json:"untracked"`
}

type AssetStateItem struct {
	AssetId     string `json:"asset_id"`     // asset ID for filtering
	AssetPath   string `json:"asset_path"`   // for checkpoints: "path/to/file"
	DisplayPath string `json:"display_path"` // for UI: "path/to/file.blend"
}

type AssetsStates struct {
	Modifieds []AssetStateItem `json:"modified"`
	Fetchable []AssetStateItem `json:"fetchable"`
	Outdated  []AssetStateItem `json:"outdated"`
}

type ModifiedAssets struct {
	Modified  []AssetStateItem `json:"modified"`
	Untracked []string         `json:"untracked"`
}

func (t *AssetService) GetAssetCount(projectPath string) (int, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return 0, err
	}
	defer dbConn.Close()

	var count int
	query := "SELECT COUNT(*) FROM full_asset"

	err = dbConn.Get(&count, query)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (t *AssetService) GetAssetByID(projectPath, assetId string) (models.Asset, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Asset{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Asset{}, err
	}
	defer tx.Rollback()
	return repository.GetAsset(tx, assetId)
}

// GetAssetByPath retrieves an asset by its asset_path and extension.
// Returns the asset or an error if not found.
func (t *AssetService) GetAssetByPath(projectPath, assetPath, extension string) (models.Asset, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Asset{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Asset{}, err
	}
	defer tx.Rollback()
	return repository.GetAssetByPath(tx, assetPath, extension)
}

func (t *AssetService) CreateAsset(projectPath, name, description, assetTypeId, collectionId string, isResource bool, templateId, templateFilePath, pointer string, isLink bool, tags []string, previewPath, comment string) (models.Asset, error) {
	app := application.Get()
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Asset{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Asset{}, err
	}
	defer tx.Rollback()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return models.Asset{}, err
	}
	if err := authorizeAssetCollectionTx(tx, assetActionCreate, collectionId); err != nil {
		return models.Asset{}, err
	}

	previewId := ""
	if previewPath != "" {
		preview, err := repository.CreatePreview(tx, previewPath)
		if err != nil {
			tx.Rollback()
			return models.Asset{}, err
		}
		previewId = preview.Hash
	}
	callBack := func(current int, total int, message string, extraMessage string) {
		progress := output.ProgressReport{
			Title:         "Creating Assets for Collection",
			Message:       name,
			Percentage:    float64(current) / float64(total) * 99,
			Current:       1,
			Total:         2,
			OperationType: "write", // Write operation - creates database records
		}
		app.Event.Emit("progress-update", progress)
	}

	createdAsset, err := repository.CreateAsset(
		tx,
		"",
		name,
		assetTypeId,
		collectionId,
		isResource,
		templateId,
		description,
		templateFilePath,
		tags,
		pointer,
		isLink,
		previewId,
		user.Id,
		comment,
		uuid.New().String(),
		callBack,
	)
	if err != nil {
		tx.Rollback()
		return models.Asset{}, err
	}

	createdAsset, err = repository.GetAsset(tx, createdAsset.Id)
	if err != nil {
		return models.Asset{}, err
	}

	// Read simple asset for write-through before commit (pointer/link assets only)
	var simpleAsset models.Asset
	if createdAsset.Pointer != "" {
		simpleAsset, _ = repository.GetSimpleAsset(tx, createdAsset.Id)
	}

	err = tx.Commit()
	if err != nil {
		return models.Asset{}, err
	}

	// Write through pointer/link assets (no binary data)
	if createdAsset.Pointer != "" && simpleAsset.Id != "" {
		enqueueAssetWriteThrough(projectPath, simpleAsset)
	}

	if templateFilePath == "" {
		return createdAsset, nil
	} else {
		progress := output.ProgressReport{
			Title:         "Creating Assets for Collection",
			Message:       "Asset",
			Percentage:    100,
			Current:       1,
			Total:         1,
			OperationType: "write",
		}
		app.Event.Emit("progress-update", progress)
	}
	return createdAsset, nil
}

// DuplicateAsset duplicates a asset to the same or a different collection.
// If targetCollectionId is empty, the asset is duplicated in the same collection as the source.
func (t *AssetService) DuplicateAsset(projectPath, sourceAssetId, targetCollectionId string) (models.Asset, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Asset{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Asset{}, err
	}
	defer tx.Rollback()

	// Get the source asset
	sourceAsset, err := repository.GetAsset(tx, sourceAssetId)
	if err != nil {
		return models.Asset{}, err
	}

	// Determine target collection: use provided targetCollectionId or fall back to source's collection
	destinationCollectionId := targetCollectionId
	if destinationCollectionId == "" {
		destinationCollectionId = sourceAsset.CollectionId
	}
	if err := authorizeAssetCollectionTx(tx, assetActionCreate, destinationCollectionId); err != nil {
		return models.Asset{}, err
	}

	// Generate unique name by checking for conflicts in the destination collection
	baseName := sourceAsset.Name
	if destinationCollectionId == sourceAsset.CollectionId {
		baseName = sourceAsset.Name + "-duplicate"
	}
	newName := baseName
	counter := 1

	// Check for name conflicts in the destination collection
	for {
		_, err := repository.GetAssetByName(tx, newName, destinationCollectionId, sourceAsset.Extension)
		if err != nil {
			// Asset with this name doesn't exist, so we can use it
			break
		}
		// Asset exists, try with number suffix
		if destinationCollectionId == sourceAsset.CollectionId {
			newName = fmt.Sprintf("%s-%d", baseName, counter)
		} else {
			newName = fmt.Sprintf("%s (%d)", baseName, counter)
		}
		counter++
	}

	// Generate new ID
	newAssetId := uuid.New().String()

	// Create the duplicate asset using AddAsset (simpler than CreateAsset)
	err = repository.AddAsset(
		tx,
		newAssetId,
		utils.GetCurrentTime(),
		newName,
		sourceAsset.AssetTypeId,
		destinationCollectionId,
		sourceAsset.StatusId,
		sourceAsset.Extension,
		sourceAsset.Description,
		sourceAsset.Tags,
		sourceAsset.Pointer,
		sourceAsset.IsLink,
		"", // No assignee for duplicated asset
		sourceAsset.PreviewId,
	)
	if err != nil {
		return models.Asset{}, err
	}

	// Copy tags from source asset
	if len(sourceAsset.Tags) > 0 {
		for _, tag := range sourceAsset.Tags {
			err = repository.AddTagToAsset(tx, newAssetId, tag)
			if err != nil {
				return models.Asset{}, err
			}
		}
	}

	// Duplicate the most recent checkpoint if it exists
	latestCheckpoint, err := repository.GetLatestCheckpoint(tx, sourceAssetId)
	if err != nil && err.Error() != "no checkpoints" {
		return models.Asset{}, err
	}

	if latestCheckpoint.Id != "" {
		// Generate new checkpoint ID and group ID
		newCheckpointId := uuid.New().String()
		newGroupId := uuid.New().String()

		// Duplicate the latest checkpoint for the new asset
		err = repository.AddCheckpoint(
			tx,
			newCheckpointId,
			utils.GetEpochTime(),
			newAssetId, // Link to the new duplicated asset
			latestCheckpoint.XXHashChecksum,
			latestCheckpoint.TimeModified,
			latestCheckpoint.FileSize,
			latestCheckpoint.Comment,
			latestCheckpoint.Chunks,
			latestCheckpoint.AuthorUID,
			latestCheckpoint.PreviewId,
			false, // Not synced initially
		)
		if err != nil {
			return models.Asset{}, err
		}

		// Update the checkpoint with the new group ID
		_, err = tx.Exec("UPDATE asset_checkpoint SET group_id = ? WHERE id = ?", newGroupId, newCheckpointId)
		if err != nil {
			return models.Asset{}, err
		}
	}

	// Get the newly created asset
	duplicatedAsset, err := repository.GetAsset(tx, newAssetId)
	if err != nil {
		return models.Asset{}, err
	}

	err = tx.Commit()
	if err != nil {
		return models.Asset{}, err
	}

	return duplicatedAsset, nil
}

// CopyAssetToProject copies an asset from one project to another, including metadata, checkpoints, chunks, and previews.
// If targetCollectionId is empty, the asset is copied to the root of the target project.
// If copyAllCheckpoints is false, only the latest checkpoint is copied.
func (t *AssetService) CopyAssetToProject(sourceProjectPath, sourceAssetId, targetProjectPath, targetCollectionId string, copyAllCheckpoints bool) (models.Asset, error) {
	app := application.Get()

	// Open source database
	sourceDbConn, err := utils.OpenDb(sourceProjectPath)
	if err != nil {
		return models.Asset{}, fmt.Errorf("failed to open source project: %w", err)
	}
	defer sourceDbConn.Close()

	sourceTx, err := sourceDbConn.Beginx()
	if err != nil {
		return models.Asset{}, err
	}
	defer sourceTx.Rollback()

	// Open target database
	targetDbConn, err := utils.OpenDb(targetProjectPath)
	if err != nil {
		return models.Asset{}, fmt.Errorf("failed to open target project: %w", err)
	}
	defer targetDbConn.Close()

	targetTx, err := targetDbConn.Beginx()
	if err != nil {
		return models.Asset{}, err
	}
	defer targetTx.Rollback()
	if err := authorizeAssetCollectionTx(targetTx, assetActionCreate, targetCollectionId); err != nil {
		return models.Asset{}, err
	}

	// Get the source asset
	sourceAsset, err := repository.GetAsset(sourceTx, sourceAssetId)
	if err != nil {
		return models.Asset{}, fmt.Errorf("failed to get source asset: %w", err)
	}

	// Emit progress
	progress := output.ProgressReport{
		Title:         "Copying Asset",
		Message:       sourceAsset.Name,
		Percentage:    10,
		OperationType: "write",
	}
	app.Event.Emit("progress-update", progress)

	// Map asset type - find equivalent in target project by name, or use "Generic"
	var targetAssetTypeId string
	if sourceAsset.AssetTypeName != "" {
		targetAssetType, err := repository.GetAssetTypeByName(targetTx, sourceAsset.AssetTypeName)
		if err == nil {
			targetAssetTypeId = targetAssetType.Id
		}
	}
	if targetAssetTypeId == "" {
		// Fall back to Generic
		genericAssetType, err := repository.GetAssetTypeByName(targetTx, "Generic")
		if err == nil {
			targetAssetTypeId = genericAssetType.Id
		} else {
			// Create Generic if it doesn't exist
			genericAssetType, err = repository.GetOrCreateAssetType(targetTx, "Generic", "")
			if err != nil {
				return models.Asset{}, fmt.Errorf("failed to get or create Generic asset type: %w", err)
			}
			targetAssetTypeId = genericAssetType.Id
		}
	}

	// Map status - find equivalent in target project or use first available
	var targetStatusId string
	if sourceAsset.StatusShortName != "" {
		targetStatus, err := repository.GetStatusByShortName(targetTx, sourceAsset.StatusShortName)
		if err == nil {
			targetStatusId = targetStatus.Id
		}
	}
	if targetStatusId == "" {
		// Get first available status
		statuses, err := repository.GetStatuses(targetTx)
		if err == nil && len(statuses) > 0 {
			targetStatusId = statuses[0].Id
		}
	}

	// Generate unique name by checking for conflicts in target project
	baseName := sourceAsset.Name
	newName := baseName
	counter := 1

	for {
		_, err := repository.GetAssetByName(targetTx, newName, targetCollectionId, sourceAsset.Extension)
		if err != nil {
			// Asset with this name doesn't exist, we can use it
			break
		}
		// Asset exists, try with number suffix
		newName = fmt.Sprintf("%s-%d", baseName, counter)
		counter++
	}

	progress.Percentage = 20
	progress.Message = "Creating asset in target project"
	app.Event.Emit("progress-update", progress)

	// Generate new asset ID
	newAssetId := uuid.New().String()

	// Copy preview if exists
	var newPreviewId string
	if sourceAsset.PreviewId != "" {
		sourcePreview, err := repository.GetPreview(sourceTx, sourceAsset.PreviewId)
		if err == nil {
			// Add preview to target project if it doesn't exist
			err = repository.AddPreview(targetTx, sourcePreview.Hash, sourcePreview.Preview, sourcePreview.Extension)
			if err == nil {
				newPreviewId = sourcePreview.Hash
			}
		}
	}

	// Create the asset in target project
	err = repository.AddAsset(
		targetTx,
		newAssetId,
		utils.GetCurrentTime(),
		newName,
		targetAssetTypeId,
		targetCollectionId,
		targetStatusId,
		sourceAsset.Extension,
		sourceAsset.Description,
		[]string{}, // Don't copy tags (they may not exist in target project)
		"",         // No pointer (file path will be different in target project)
		false,      // Not a link
		"",         // No assignee
		newPreviewId,
	)
	if err != nil {
		return models.Asset{}, fmt.Errorf("failed to create asset in target project: %w", err)
	}

	progress.Percentage = 30
	progress.Message = "Copying checkpoints"
	app.Event.Emit("progress-update", progress)

	// Get checkpoints to copy
	var checkpointsToCopy []models.Checkpoint
	if copyAllCheckpoints {
		checkpointsToCopy, err = repository.GetCheckpoints(sourceTx, sourceAssetId, false)
		if err != nil && err.Error() != "no checkpoints" {
			return models.Asset{}, fmt.Errorf("failed to get checkpoints: %w", err)
		}
	} else {
		latestCheckpoint, err := repository.GetLatestCheckpoint(sourceTx, sourceAssetId)
		if err != nil && err.Error() != "no checkpoints" {
			return models.Asset{}, fmt.Errorf("failed to get latest checkpoint: %w", err)
		}
		if latestCheckpoint.Id != "" {
			checkpointsToCopy = []models.Checkpoint{latestCheckpoint}
		}
	}

	// Copy each checkpoint
	totalCheckpoints := len(checkpointsToCopy)
	for i, checkpoint := range checkpointsToCopy {
		progressPercent := 30 + (50 * (i + 1) / max(totalCheckpoints, 1))
		progress.Percentage = float64(progressPercent)
		progress.Message = fmt.Sprintf("Copying checkpoint %d of %d", i+1, totalCheckpoints)
		app.Event.Emit("progress-update", progress)

		// Copy checkpoint preview if exists
		var checkpointPreviewId string
		if checkpoint.PreviewId != "" {
			checkpointPreview, err := repository.GetPreview(sourceTx, checkpoint.PreviewId)
			if err == nil {
				err = repository.AddPreview(targetTx, checkpointPreview.Hash, checkpointPreview.Preview, checkpointPreview.Extension)
				if err == nil {
					checkpointPreviewId = checkpointPreview.Hash
				}
			}
		}

		// Copy chunks for this checkpoint
		if checkpoint.Chunks != "" {
			chunkHashes := strings.Split(checkpoint.Chunks, ",")
			for _, chunkHash := range chunkHashes {
				if chunkHash == "" {
					continue
				}
				// Check if chunk already exists in target
				var existingHash string
				targetTx.Get(&existingHash, "SELECT hash FROM chunk WHERE hash = ?", chunkHash)
				if existingHash != "" {
					continue // Chunk already exists
				}

				// Get chunk from source
				var chunk struct {
					Hash string `db:"hash"`
					Data []byte `db:"data"`
					Size int    `db:"size"`
				}
				err := sourceTx.Get(&chunk, "SELECT hash, data, size FROM chunk WHERE hash = ?", chunkHash)
				if err != nil {
					continue // Skip if chunk not found
				}

				// Insert chunk into target
				_, err = targetTx.Exec("INSERT INTO chunk (hash, data, size) VALUES (?, ?, ?)",
					chunk.Hash, chunk.Data, chunk.Size)
				if err != nil {
					// Ignore duplicate errors
					continue
				}
			}
		}

		// Generate new checkpoint ID and group ID
		newCheckpointId := uuid.New().String()
		newGroupId := uuid.New().String()

		// Add checkpoint to target project
		err = repository.AddCheckpoint(
			targetTx,
			newCheckpointId,
			utils.GetEpochTime(),
			newAssetId,
			checkpoint.XXHashChecksum,
			checkpoint.TimeModified,
			checkpoint.FileSize,
			checkpoint.Comment,
			checkpoint.Chunks,
			checkpoint.AuthorUID,
			checkpointPreviewId,
			false, // Not synced
		)
		if err != nil {
			return models.Asset{}, fmt.Errorf("failed to add checkpoint: %w", err)
		}

		// Update the checkpoint with the group ID
		_, err = targetTx.Exec("UPDATE asset_checkpoint SET group_id = ? WHERE id = ?", newGroupId, newCheckpointId)
		if err != nil {
			return models.Asset{}, fmt.Errorf("failed to update checkpoint group: %w", err)
		}
	}

	progress.Percentage = 90
	progress.Message = "Finalizing"
	app.Event.Emit("progress-update", progress)

	// Get the newly created asset
	newAsset, err := repository.GetAsset(targetTx, newAssetId)
	if err != nil {
		return models.Asset{}, fmt.Errorf("failed to get created asset: %w", err)
	}

	// Commit target transaction
	err = targetTx.Commit()
	if err != nil {
		return models.Asset{}, fmt.Errorf("failed to commit changes: %w", err)
	}

	progress.Percentage = 100
	progress.Message = "Complete"
	app.Event.Emit("progress-update", progress)

	return newAsset, nil
}

// ChangeStatus updates asset statuses on the remote server first, then locally.
// If no remote is configured, falls back to local-only update.
func (t *AssetService) ChangeStatus(projectPath string, assetIds []string, statusId string) (MetadataUpdateResult, error) {
	if err := authorizeAssetAction(projectPath, assetActionChangeStatus, assetIds); err != nil {
		return MetadataUpdateResult{}, err
	}
	remoteURL, err := utils.ResolveProjectRemoteURL(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, fmt.Errorf("failed to resolve project remote: %w", err)
	}

	// Remote-first: push to the typed asset mutation endpoint, then apply canonical rows locally.
	remoteConfirmed := false
	var remoteFailure error
	var remoteAssets []models.Asset
	if remoteURL != "" {
		patches := make([]assetPatch, 0, len(assetIds))
		for _, assetId := range assetIds {
			id := statusId
			patches = append(patches, assetPatch{Id: assetId, StatusId: &id})
		}
		response, err := patchAssetsRemote(remoteURL, patches)
		if err != nil {
			if !IsMetadataTransportFailure(err) {
				return MetadataUpdateResult{}, fmt.Errorf("remote status update failed: %w", err)
			}
			remoteFailure = err
		} else {
			remoteAssets = response.Assets
			remoteConfirmed = true
		}
	}

	// Apply locally after remote confirmation (or directly if no remote)
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	defer tx.Rollback()

	if !remoteConfirmed {
		for _, assetId := range assetIds {
			err = repository.UpdateStatus(tx, assetId, statusId)
			if err != nil {
				return MetadataUpdateResult{}, err
			}
		}
	}
	if remoteConfirmed {
		if err = applyCanonicalAssets(tx, remoteAssets); err != nil {
			return MetadataUpdateResult{}, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return MetadataUpdateResult{}, err
	}

	// Push status to external integration (non-blocking)
	go func() {
		integrationSvc := &IntegrationService{}
		if pushErr := integrationSvc.PushToIntegration(projectPath, assetIds, "", "", ""); pushErr != nil {
			log.Printf("integration status push failed: %v", pushErr)
		}
	}()

	return MetadataUpdateResult{RemoteApplied: remoteConfirmed, RequiresSync: remoteFailure != nil}, nil
}

// postStatusChangeRemote sends a status change request to the remote server.
func postStatusChangeRemote(remoteURL string, assetIds []string, statusId string) error {
	type assetStatus struct {
		AssetId  string `json:"asset_id"`
		StatusId string `json:"status_id"`
	}
	assets := make([]assetStatus, len(assetIds))
	for i, id := range assetIds {
		assets[i] = assetStatus{AssetId: id, StatusId: statusId}
	}
	payload := struct {
		Assets []assetStatus `json:"assets"`
	}{Assets: assets}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", remoteURL+"/status", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	auth_service.AttachBearerToken(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ChangeAssetCollection moves one or more assets to a different collection.
// Checks for name+extension conflicts in the target collection before moving.
// Returns an error if any asset would conflict or if the operation fails.
func (t *AssetService) ChangeAssetCollection(projectPath string, assetIds []string, collectionId string) error {
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
	if err := authorizeAssetActionTx(tx, assetActionUpdate, assetIds); err != nil {
		return err
	}
	if err := authorizeAssetCollectionTx(tx, assetActionUpdate, collectionId); err != nil {
		return err
	}

	var conflicts []string
	for _, assetId := range assetIds {
		asset, err := repository.GetAsset(tx, assetId)
		if err != nil {
			return err
		}
		if asset.CollectionId == collectionId {
			continue
		}
		_, err = repository.GetAssetByName(tx, asset.Name, collectionId, asset.Extension)
		if err == nil {
			conflicts = append(conflicts, asset.Name+asset.Extension)
		} else if err != error_service.ErrAssetNotFound {
			return err
		}
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("assets with the same name and extension already exist in the target collection: %s", strings.Join(conflicts, ", "))
	}

	var movedAssets []models.Asset
	for _, assetId := range assetIds {
		repository.ChangeCollection(tx, assetId, collectionId)
	}
	for _, assetId := range assetIds {
		asset, err := repository.GetSimpleAsset(tx, assetId)
		if err == nil {
			movedAssets = append(movedAssets, asset)
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	for _, asset := range movedAssets {
		enqueueAssetWriteThrough(projectPath, asset)
	}
	return nil
}

// MoveAssetsToCollection moves one or more assets to a different collection.
// Updates the database and moves the physical files if they exist on disk.
// Checks for name+extension conflicts in the target collection before moving.
// Returns an error if any asset would conflict or if the operation fails.
func (t *AssetService) MoveAssetsToCollection(projectPath string, assetIds []string, targetCollectionId string) error {
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

	if err := authorizeAssetActionTx(tx, assetActionUpdate, assetIds); err != nil {
		return err
	}
	_, role, err := activeAssetRole(tx)
	if err != nil {
		return err
	}
	if targetCollectionId == "" && role.Name != "admin" {
		return errors.New("assets can only be moved to an assigned collection")
	}
	if targetCollectionId != "" {
		if err := authorizeAssetCollectionTx(tx, assetActionUpdate, targetCollectionId); err != nil {
			return err
		}
	}

	// Get target directory path
	var targetDir string
	if targetCollectionId == "" {
		// Moving to root - use project's working directory
		targetDir, err = utils.GetProjectWorkingDir(tx)
		if err != nil {
			return err
		}
	} else {
		// Moving to a collection - get its file path
		targetCollection, err := repository.GetCollection(tx, targetCollectionId)
		if err != nil {
			return err
		}
		targetDir = targetCollection.FilePath
	}

	// Check for name+extension conflicts in target collection
	var conflicts []string
	var assetsToMove []models.Asset
	for _, assetId := range assetIds {
		asset, err := repository.GetAsset(tx, assetId)
		if err != nil {
			return err
		}
		// Skip if already in target collection
		if asset.CollectionId == targetCollectionId {
			continue
		}
		// Check if a asset with same name+extension exists in target collection
		_, err = repository.GetAssetByName(tx, asset.Name, targetCollectionId, asset.Extension)
		if err == nil {
			// Asset exists - this is a conflict
			conflicts = append(conflicts, asset.Name+asset.Extension)
		} else if err != error_service.ErrAssetNotFound {
			// Some other error occurred
			return err
		}
		assetsToMove = append(assetsToMove, asset)
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("assets with the same name and extension already exist in the target collection: %s", strings.Join(conflicts, ", "))
	}

	// No conflicts - proceed with move
	// Ensure target directory exists
	os.MkdirAll(targetDir, os.ModePerm)

	// Move files and update database
	for _, asset := range assetsToMove {
		// Calculate new file path
		fileName := asset.Name + asset.Extension
		newFilePath := filepath.Join(targetDir, fileName)

		// Move file if it exists on disk
		if asset.FilePath != "" {
			if _, err := os.Stat(asset.FilePath); err == nil {
				err = os.Rename(asset.FilePath, newFilePath)
				if err != nil {
					return fmt.Errorf("failed to move file %s: %w", asset.Name, err)
				}
			}
		}

		// Update database
		repository.ChangeCollection(tx, asset.Id, targetCollectionId)
	}

	// Read updated assets for write-through before commit
	var movedSimpleAssets []models.Asset
	for _, asset := range assetsToMove {
		simpleAsset, err := repository.GetSimpleAsset(tx, asset.Id)
		if err == nil {
			movedSimpleAssets = append(movedSimpleAssets, simpleAsset)
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	for _, asset := range movedSimpleAssets {
		enqueueAssetWriteThrough(projectPath, asset)
	}
	return nil
}

func (t *AssetService) DeleteAsset(projectPath, assetId string, removeFiles bool) error {
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
	if err := authorizeAssetActionTx(tx, assetActionDelete, []string{assetId}); err != nil {
		return err
	}

	err = repository.DeleteAsset(tx, assetId, removeFiles, true)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (t *AssetService) UpdateAsset(projectPath, assetId, name, assetTypeId string, isResource bool, pointer string, tags []string) (models.Asset, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Asset{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Asset{}, err
	}
	defer tx.Rollback()
	if err := authorizeAssetActionTx(tx, assetActionUpdate, []string{assetId}); err != nil {
		return models.Asset{}, err
	}

	updatedAsset, err := repository.UpdateAsset(tx, assetId, name, assetTypeId, isResource, pointer, tags)
	if err != nil {
		tx.Rollback()
		return models.Asset{}, err
	}
	err = tx.Commit()
	if err != nil {
		return models.Asset{}, err
	}
	return updatedAsset, nil
}

func (t *AssetService) ChangeAssetType(projectPath, assetId, assetTypeId string) error {
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
	if err := authorizeAssetActionTx(tx, assetActionUpdate, []string{assetId}); err != nil {
		return err
	}

	err = repository.ChangeAssetType(tx, assetId, assetTypeId)
	if err != nil {
		return err
	}
	asset, err := repository.GetSimpleAsset(tx, assetId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	enqueueAssetWriteThrough(projectPath, asset)
	return nil
}

func (t *AssetService) ToggleIsTask(projectPath, assetId string, isTask bool) (MetadataUpdateResult, error) {
	return t.BulkToggleIsTask(projectPath, []string{assetId}, isTask)
}

func (t *AssetService) BulkToggleIsTask(projectPath string, assetIds []string, isTask bool) (MetadataUpdateResult, error) {
	if len(assetIds) == 0 {
		return MetadataUpdateResult{}, nil
	}
	if err := authorizeAssetAction(projectPath, assetActionUpdate, assetIds); err != nil {
		return MetadataUpdateResult{}, err
	}
	remoteURL, err := utils.ResolveProjectRemoteURL(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, fmt.Errorf("failed to resolve project remote: %w", err)
	}
	var remoteFailure error
	if remoteURL != "" {
		patches := make([]assetPatch, 0, len(assetIds))
		for _, assetId := range assetIds {
			value := isTask
			patches = append(patches, assetPatch{Id: assetId, IsTask: &value})
		}
		response, err := patchAssetsRemote(remoteURL, patches)
		if err != nil {
			if !IsMetadataTransportFailure(err) {
				return MetadataUpdateResult{}, err
			}
			remoteFailure = err
		} else {
			db, err := utils.OpenDb(projectPath)
			if err != nil {
				return MetadataUpdateResult{}, err
			}
			defer db.Close()
			tx, err := db.Beginx()
			if err != nil {
				return MetadataUpdateResult{}, err
			}
			defer tx.Rollback()
			if err = applyCanonicalAssets(tx, response.Assets); err != nil {
				return MetadataUpdateResult{}, err
			}
			if err = tx.Commit(); err != nil {
				return MetadataUpdateResult{}, err
			}
			return MetadataUpdateResult{RemoteApplied: true}, nil
		}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	defer tx.Rollback()

	err = repository.BulkToggleIsTask(tx, assetIds, isTask)
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	assets, err := repository.GetSimpleAssetsByIds(tx, assetIds)
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	err = tx.Commit()
	if err != nil {
		return MetadataUpdateResult{}, err
	}

	if remoteFailure == nil {
		go enqueueAssetsWriteThrough(projectPath, assets)
	}
	return MetadataUpdateResult{RequiresSync: remoteFailure != nil}, nil
}

func (t *AssetService) RenameAsset(projectPath, assetId, name string) (models.Asset, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Asset{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Asset{}, err
	}
	defer tx.Rollback()
	if err := authorizeAssetActionTx(tx, assetActionUpdate, []string{assetId}); err != nil {
		return models.Asset{}, err
	}

	updatedAsset, err := repository.RenameAsset(tx, assetId, name)
	if err != nil {
		tx.Rollback()
		return models.Asset{}, err
	}
	simpleAsset, err := repository.GetSimpleAsset(tx, assetId)
	if err != nil {
		return models.Asset{}, err
	}
	err = tx.Commit()
	if err != nil {
		return models.Asset{}, err
	}
	enqueueAssetWriteThrough(projectPath, simpleAsset)
	return updatedAsset, nil
}

func (t *AssetService) AddPreview(projectPath, assetId, previewPath string) (models.Asset, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Asset{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Asset{}, err
	}
	defer tx.Rollback()

	preview, err := repository.CreatePreview(tx, previewPath)
	if err != nil {
		tx.Rollback()
		return models.Asset{}, err
	}
	err = repository.SetCollectionPreview(tx, assetId, "asset", preview.Hash)
	if err != nil {
		tx.Rollback()
		return models.Asset{}, err
	}
	updatedAsset, err := repository.GetAsset(tx, assetId)
	if err != nil {
		return models.Asset{}, err
	}

	err = tx.Commit()
	if err != nil {
		return models.Asset{}, err
	}

	return updatedAsset, nil
}

// AssignAsset assigns a asset to a user.
// If the asset is a resource (is_resource == true), it will be converted to a asset first.
func (t *AssetService) AssignAsset(projectPath, assetId, userId string) (MetadataUpdateResult, error) {
	return t.AssignAssets(projectPath, []string{assetId}, userId)
}

// AssignAssets assigns assets atomically; remote projects use one PATCH request.
func (t *AssetService) AssignAssets(projectPath string, assetIds []string, userId string) (MetadataUpdateResult, error) {
	if len(assetIds) == 0 {
		return MetadataUpdateResult{}, nil
	}
	if err := authorizeAssetAction(projectPath, assetActionAssign, assetIds); err != nil {
		return MetadataUpdateResult{}, err
	}
	remoteURL, err := utils.ResolveProjectRemoteURL(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, fmt.Errorf("failed to resolve project remote: %w", err)
	}
	var remoteFailure error
	if remoteURL != "" {
		uid := userId
		patches := make([]assetPatch, 0, len(assetIds))
		for _, assetId := range assetIds {
			patches = append(patches, assetPatch{Id: assetId, AssigneeId: &uid})
		}
		response, err := patchAssetsRemote(remoteURL, patches)
		if err != nil {
			if !IsMetadataTransportFailure(err) {
				return MetadataUpdateResult{}, err
			}
			remoteFailure = err
		} else {
			db, err := utils.OpenDb(projectPath)
			if err != nil {
				return MetadataUpdateResult{}, err
			}
			defer db.Close()
			tx, err := db.Beginx()
			if err != nil {
				return MetadataUpdateResult{}, err
			}
			defer tx.Rollback()
			if err = applyCanonicalAssets(tx, response.Assets); err != nil {
				return MetadataUpdateResult{}, err
			}
			if err = tx.Commit(); err != nil {
				return MetadataUpdateResult{}, err
			}
			return MetadataUpdateResult{RemoteApplied: true}, nil
		}
	}
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	defer tx.Rollback()

	assets := make([]models.Asset, 0, len(assetIds))
	for _, assetId := range assetIds {
		asset, err := repository.GetAsset(tx, assetId)
		if err != nil {
			return MetadataUpdateResult{}, err
		}
		if asset.IsResource {
			if err = repository.ToggleIsTask(tx, assetId, true); err != nil {
				return MetadataUpdateResult{}, err
			}
		}
		if err = repository.AssignAsset(tx, assetId, userId); err != nil {
			return MetadataUpdateResult{}, err
		}
		asset, err = repository.GetSimpleAsset(tx, assetId)
		if err != nil {
			return MetadataUpdateResult{}, err
		}
		assets = append(assets, asset)
	}
	err = tx.Commit()
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	if remoteFailure == nil {
		enqueueAssetsWriteThrough(projectPath, assets)
	}
	return MetadataUpdateResult{RequiresSync: remoteFailure != nil}, nil
}
func (t *AssetService) UnassignAsset(projectPath, assetId string) (MetadataUpdateResult, error) {
	if err := authorizeAssetAction(projectPath, assetActionUnassign, []string{assetId}); err != nil {
		return MetadataUpdateResult{}, err
	}
	remoteURL, err := utils.ResolveProjectRemoteURL(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, fmt.Errorf("failed to resolve project remote: %w", err)
	}
	var remoteFailure error
	if remoteURL != "" {
		empty := ""
		response, err := patchAssetsRemote(remoteURL, []assetPatch{{Id: assetId, AssigneeId: &empty}})
		if err != nil {
			if !IsMetadataTransportFailure(err) {
				return MetadataUpdateResult{}, err
			}
			remoteFailure = err
		} else {
			db, err := utils.OpenDb(projectPath)
			if err != nil {
				return MetadataUpdateResult{}, err
			}
			defer db.Close()
			tx, err := db.Beginx()
			if err != nil {
				return MetadataUpdateResult{}, err
			}
			defer tx.Rollback()
			if err = applyCanonicalAssets(tx, response.Assets); err != nil {
				return MetadataUpdateResult{}, err
			}
			if err = tx.Commit(); err != nil {
				return MetadataUpdateResult{}, err
			}
			return MetadataUpdateResult{RemoteApplied: true}, nil
		}
	}
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	defer tx.Rollback()

	err = repository.UnAssignAsset(tx, assetId)
	if err != nil {
		tx.Rollback()
		return MetadataUpdateResult{}, err
	}
	asset, err := repository.GetSimpleAsset(tx, assetId)
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	err = tx.Commit()
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	if remoteFailure == nil {
		enqueueAssetWriteThrough(projectPath, asset)
	}
	return MetadataUpdateResult{RequiresSync: remoteFailure != nil}, nil
}
func (t *AssetService) UnassignAssets(projectPath string, assetIds []string) (MetadataUpdateResult, error) {
	if len(assetIds) == 0 {
		return MetadataUpdateResult{}, nil
	}
	if err := authorizeAssetAction(projectPath, assetActionUnassign, assetIds); err != nil {
		return MetadataUpdateResult{}, err
	}
	remoteURL, err := utils.ResolveProjectRemoteURL(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, fmt.Errorf("failed to resolve project remote: %w", err)
	}
	var remoteFailure error
	if remoteURL != "" {
		empty := ""
		patches := make([]assetPatch, 0, len(assetIds))
		for _, assetId := range assetIds {
			patches = append(patches, assetPatch{Id: assetId, AssigneeId: &empty})
		}
		response, err := patchAssetsRemote(remoteURL, patches)
		if err != nil {
			if !IsMetadataTransportFailure(err) {
				return MetadataUpdateResult{}, err
			}
			remoteFailure = err
		} else {
			db, err := utils.OpenDb(projectPath)
			if err != nil {
				return MetadataUpdateResult{}, err
			}
			defer db.Close()
			tx, err := db.Beginx()
			if err != nil {
				return MetadataUpdateResult{}, err
			}
			defer tx.Rollback()
			if err = applyCanonicalAssets(tx, response.Assets); err != nil {
				return MetadataUpdateResult{}, err
			}
			if err = tx.Commit(); err != nil {
				return MetadataUpdateResult{}, err
			}
			return MetadataUpdateResult{RemoteApplied: true}, nil
		}
	}
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	defer tx.Rollback()

	err = repository.UnAssignAssets(tx, assetIds)
	if err != nil {
		tx.Rollback()
		return MetadataUpdateResult{}, err
	}
	var unassignedAssets []models.Asset
	for _, id := range assetIds {
		asset, err := repository.GetSimpleAsset(tx, id)
		if err == nil {
			unassignedAssets = append(unassignedAssets, asset)
		}
	}
	err = tx.Commit()
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	if remoteFailure == nil {
		for _, asset := range unassignedAssets {
			enqueueAssetWriteThrough(projectPath, asset)
		}
	}
	return MetadataUpdateResult{RequiresSync: remoteFailure != nil}, nil
}
func (t *AssetService) AssetFileStatus(projectPath, assetId string) (string, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return "", err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	asset, err := repository.GetAsset(tx, assetId)
	if err != nil {
		return "", err
	}
	return asset.FileStatus, nil
}
func (t *AssetService) AssetFilesStatus(projectPath string, assetIds []string) (map[string]string, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return map[string]string{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return map[string]string{}, err
	}
	defer tx.Rollback()

	filesStatus, err := repository.GetFilesStatus(tx, assetIds)
	if err != nil {
		return map[string]string{}, err
	}
	return filesStatus, nil
}

func (t *AssetService) GetAssetState(projectPath string, assetId string) (string, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return "", err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	assetState, err := repository.GetAssetState(tx, assetId)
	if err != nil {
		return "", err
	}
	return assetState, nil
}

func (t *AssetService) ToggleIsResource(projectPath string, assetIds []string, isResource bool) error {
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

	err = repository.ToggleIsResourceM(tx, assetIds, isResource)
	if err != nil {
		return err
	}
	return nil
}

func (t *AssetService) RevealAsset(projectPath, assetId string) error {
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

	asset, err := repository.GetAsset(tx, assetId)
	if err != nil {
		return err
	}
	utils.RevealInExplorer(asset.GetFilePath())
	return nil
}

// dependencies

// ResolveBuildDependencies returns the asset and every asset transitively
// reachable through its asset and collection dependencies. Used by the
// "Build with dependencies" action to expand the full revert set on the
// backend instead of relying on the partial frontend store.
func (t *AssetService) ResolveBuildDependencies(projectPath, assetId string) ([]string, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return nil, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	return repository.ResolveBuildDependencies(tx, assetId)
}

func (t *AssetService) AddCollectionDependency(projectPath, assetId, dependencyId, dependencyTypeId string) (models.AssetDependency, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.AssetDependency{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.AssetDependency{}, err
	}
	defer tx.Rollback()
	if err := authorizeAssetActionTx(tx, assetActionManageDependencies, []string{assetId}); err != nil {
		return models.AssetDependency{}, err
	}
	collectionDependency, err := repository.AddCollectionDependency(tx, "", assetId, dependencyId, dependencyTypeId)
	if err != nil {
		return models.AssetDependency{}, err
	}
	err = tx.Commit()
	if err != nil {
		return models.AssetDependency{}, err
	}
	enqueueCollectionDependencyWriteThrough(projectPath, collectionDependency)
	return collectionDependency, nil
}
func (t *AssetService) RemoveCollectionDependency(projectPath, assetId, dependencyId string) error {
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
	if err := authorizeAssetActionTx(tx, assetActionManageDependencies, []string{assetId}); err != nil {
		return err
	}
	// Read the collection dependency row ID before deletion for tomb lookup
	dep, err := repository.GetCollectionDependencyByKeys(tx, assetId, dependencyId)
	if err != nil {
		return err
	}
	err = repository.RemoveCollectionDependency(tx, assetId, dependencyId)
	if err != nil {
		return err
	}
	tomb, err := repository.GetTomb(tx, dep.Id)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	enqueueTombWriteThrough(projectPath, tomb)
	return nil
}
func (t *AssetService) AddAssetDependency(projectPath, assetId, dependencyId, dependencyTypeId string) (models.AssetDependency, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.AssetDependency{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.AssetDependency{}, err
	}
	defer tx.Rollback()
	if err := authorizeAssetActionTx(tx, assetActionManageDependencies, []string{assetId}); err != nil {
		return models.AssetDependency{}, err
	}

	assetDependency, err := repository.AddDependency(tx, "", assetId, dependencyId, dependencyTypeId)
	if err != nil {
		return models.AssetDependency{}, err
	}
	err = tx.Commit()
	if err != nil {
		return models.AssetDependency{}, err
	}
	enqueueDependencyWriteThrough(projectPath, assetDependency)
	return assetDependency, nil
}
func (t *AssetService) RemoveAssetDependency(projectPath, assetId, dependencyId string) error {
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
	if err := authorizeAssetActionTx(tx, assetActionManageDependencies, []string{assetId}); err != nil {
		return err
	}
	err = repository.RemoveAssetDependency(tx, assetId, dependencyId)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
func (t *AssetService) GetAssetDependencies2(projectPath string, assetIds []string) ([]models.Asset, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Asset{}, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Asset{}, err
	}
	defer tx.Rollback()

	if len(assetIds) == 0 {
		return []models.Asset{}, nil
	}

	assets := []models.Asset{}
	quotedAssetIds := make([]string, len(assetIds))
	for i, id := range assetIds {
		quotedAssetIds[i] = fmt.Sprintf("'%s'", id)
	}

	assetsQuery := fmt.Sprintf(` SELECT * FROM full_asset  WHERE id IN (%s) AND trashed = 0 ORDER BY name `, strings.Join(quotedAssetIds, ","))

	err = tx.Select(&assets, assetsQuery)
	if err != nil && err != sql.ErrNoRows {
		return []models.Asset{}, err
	}

	err = tx.Commit()
	if err != nil {
		return assets, err
	}

	return assets, nil
}
func (t *AssetService) GetAssetDependencies(projectPath string, assetIds []string) ([]interface{}, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []interface{}{}, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return []interface{}{}, err
	}
	defer tx.Rollback()

	// If no asset IDs provided, return empty result
	if len(assetIds) == 0 {
		return []interface{}{}, nil
	}

	result := []interface{}{}

	// Get the asset records
	assets := []models.Asset{}
	quotedAssetIds := make([]string, len(assetIds))
	for i, id := range assetIds {
		quotedAssetIds[i] = fmt.Sprintf("'%s'", id)
	}

	assetsQuery := fmt.Sprintf(`
		SELECT * FROM full_asset 
		WHERE id IN (%s) AND trashed = 0 
		ORDER BY name
	`, strings.Join(quotedAssetIds, ","))

	err = tx.Select(&assets, assetsQuery)
	if err != nil && err != sql.ErrNoRows {
		return []interface{}{}, err
	}

	// Add assets to result
	for _, asset := range assets {
		result = append(result, asset)
	}

	// Find IDs that didn't match any assets
	foundAssetIds := make(map[string]bool)
	for _, asset := range assets {
		foundAssetIds[asset.Id] = true
	}

	missingIds := []string{}
	for _, id := range assetIds {
		if !foundAssetIds[id] {
			missingIds = append(missingIds, id)
		}
	}

	// Get collections for missing IDs
	if len(missingIds) > 0 {
		collections := []models.Collection{}
		quotedMissingIds := make([]string, len(missingIds))
		for i, id := range missingIds {
			quotedMissingIds[i] = fmt.Sprintf("'%s'", id)
		}

		collectionsQuery := fmt.Sprintf(`
			SELECT * FROM full_collection 
			WHERE id IN (%s) AND trashed = 0 
			ORDER BY name
		`, strings.Join(quotedMissingIds, ","))

		err = tx.Select(&collections, collectionsQuery)
		if err != nil && err != sql.ErrNoRows {
			return []interface{}{}, err
		}

		for _, collection := range collections {
			result = append(result, collection)
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return result, err
	}

	return result, nil
}

func (t *AssetService) GetRecursiveDependencies(projectPath string, assetId string, maxDepth int) ([]interface{}, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []interface{}{}, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return []interface{}{}, err
	}
	defer tx.Rollback()

	// Get all asset dependencies records
	allAssetDependencies := []models.AssetDependency{}
	query := `SELECT asset_id, dependency_id FROM asset_dependency`
	err = tx.Select(&allAssetDependencies, query)
	if err != nil {
		return []interface{}{}, err
	}

	// Get all collection dependencies records
	allCollectionDependencies := []models.CollectionDependency{}
	query = `SELECT asset_id, dependency_id FROM collection_dependency`
	err = tx.Select(&allCollectionDependencies, query)
	if err != nil {
		return []interface{}{}, err
	}

	// Get all asset info for checking asset existence
	allAssetInfo := []models.Asset{}
	query = `SELECT id FROM asset WHERE trashed = 0`
	err = tx.Select(&allAssetInfo, query)
	if err != nil {
		return []interface{}{}, err
	}

	// Get all collection info for checking collection existence
	allCollectionInfo := []models.Collection{}
	query = `SELECT id FROM collection WHERE trashed = 0`
	err = tx.Select(&allCollectionInfo, query)
	if err != nil {
		return []interface{}{}, err
	}

	// Track dependencies with their depth and parent information
	type DependencyInfo struct {
		ID       string
		Depth    int
		ParentID string
	}

	result := []interface{}{}
	dependenciesMap := make(map[string]DependencyInfo) // track dependency info

	// Helper function to collect dependencies recursively
	var collectDependencies func(string, int, string)
	collectDependencies = func(currentAssetId string, currentDepth int, parentAssetId string) {
		if currentDepth >= maxDepth {
			return
		}

		// If we encounter the original assetId, skip collecting it or its dependencies
		if currentAssetId == assetId && currentDepth > 0 {
			return
		}

		// Get direct asset dependencies
		for _, assetDep := range allAssetDependencies {
			if assetDep.AssetId == currentAssetId {
				depId := assetDep.DependencyId
				if depId == assetId {
					// Skip collecting the original assetId as a dependency
					continue
				}
				if existing, exists := dependenciesMap[depId]; !exists || currentDepth+1 < existing.Depth {
					dependenciesMap[depId] = DependencyInfo{
						ID:       depId,
						Depth:    currentDepth + 1,
						ParentID: currentAssetId, // The current asset is the parent of this dependency
					}
					collectDependencies(depId, currentDepth+1, currentAssetId)
				}
			}
		}

		// Get collection dependencies (collections only, no child traversal)
		for _, collectionDep := range allCollectionDependencies {
			if collectionDep.AssetId == currentAssetId {
				collectionId := collectionDep.DependencyId
				if collectionId == assetId {
					// Skip collecting the original assetId as a dependency
					continue
				}
				if existing, exists := dependenciesMap[collectionId]; !exists || currentDepth+1 < existing.Depth {
					dependenciesMap[collectionId] = DependencyInfo{
						ID:       collectionId,
						Depth:    currentDepth + 1,
						ParentID: currentAssetId, // The current asset is the parent of this collection dependency
					}
				}
			}
		}
	}

	// Start recursive collection from the given asset
	collectDependencies(assetId, 0, "")

	// Get all dependency IDs
	dependencyIds := make([]string, 0, len(dependenciesMap))
	for depId := range dependenciesMap {
		dependencyIds = append(dependencyIds, depId)
	}

	if len(dependencyIds) == 0 {
		return result, nil
	}

	// Fetch asset objects
	assets := []models.Asset{}
	quotedAssetIds := make([]string, 0)

	for _, depId := range dependencyIds {
		// Check if this ID corresponds to a asset
		for _, asset := range allAssetInfo {
			if asset.Id == depId {
				quotedAssetIds = append(quotedAssetIds, fmt.Sprintf("'%s'", depId))
				break
			}
		}
	}

	if len(quotedAssetIds) > 0 {
		assetsQuery := fmt.Sprintf(`
			SELECT * FROM full_asset 
			WHERE id IN (%s) AND trashed = 0 
			ORDER BY name
		`, strings.Join(quotedAssetIds, ","))

		err = tx.Select(&assets, assetsQuery)
		if err != nil && err != sql.ErrNoRows {
			return []interface{}{}, err
		}

		// Add depth and parent information to assets
		for _, asset := range assets {
			depInfo := dependenciesMap[asset.Id]
			assetWithDepth := map[string]interface{}{
				"asset":    asset,
				"name":     asset.Name,
				"depth":    depInfo.Depth,
				"parentId": depInfo.ParentID,
				"type":     "asset",
			}
			result = append(result, assetWithDepth)
		}
	}

	// Fetch collection objects
	collections := []models.Collection{}
	quotedCollectionIds := make([]string, 0)

	for _, depId := range dependencyIds {
		// Check if this ID corresponds to an collection
		for _, collection := range allCollectionInfo {
			if collection.Id == depId {
				quotedCollectionIds = append(quotedCollectionIds, fmt.Sprintf("'%s'", depId))
				break
			}
		}
	}

	if len(quotedCollectionIds) > 0 {
		collectionsQuery := fmt.Sprintf(`
			SELECT * FROM full_collection 
			WHERE id IN (%s) AND trashed = 0 
			ORDER BY name
		`, strings.Join(quotedCollectionIds, ","))

		err = tx.Select(&collections, collectionsQuery)
		if err != nil && err != sql.ErrNoRows {
			return []interface{}{}, err
		}

		// Add depth and parent information to collections
		for _, collection := range collections {
			depInfo := dependenciesMap[collection.Id]
			collectionWithDepth := map[string]interface{}{
				"collection": collection,
				"depth":      depInfo.Depth,
				"parentId":   depInfo.ParentID,
				"type":       "collection",
			}
			result = append(result, collectionWithDepth)
		}
	}

	err = tx.Commit()
	if err != nil {
		return result, err
	}

	return result, nil
}

func (t *AssetService) GetAssets(projectPath string) ([]models.Asset, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return []models.Asset{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Asset{}, err
	}
	defer tx.Rollback()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return []models.Asset{}, err
	}
	userData, err := repository.GetUser(tx, user.Id)
	if err != nil {
		return []models.Asset{}, err
	}
	userRole, err := repository.GetRole(tx, userData.RoleId)
	if err != nil {
		return []models.Asset{}, err
	}
	if userRole.ViewAsset {
		start := time.Now()
		assets, err := repository.GetAssets(tx, true)
		if err != nil {
			return []models.Asset{}, err
		}
		elapsed := time.Since(start)
		fmt.Printf("GetAssets operation took %s\n", elapsed)
		return assets, nil
	} else {
		assets, err := repository.GetUserAssets(tx, user.Id)
		if err != nil {
			return []models.Asset{}, err
		}
		return assets, nil
	}
}

// GetAssetAssets gets all assets where is_resource is false with minimal fields for UI display
func (t *AssetService) GetAssetAssets(projectPath string) ([]models.Asset, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return []models.Asset{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Asset{}, err
	}
	defer tx.Rollback()

	assets, err := repository.GetAssetAssets(tx)
	if err != nil {
		return []models.Asset{}, err
	}
	return assets, nil
}

// GetCollectionDescendantAssets returns assets located anywhere under the given collection's subtree.
// includeResources controls whether resource assets are included alongside task assets.
func (t *AssetService) GetCollectionDescendantAssets(projectPath, collectionId string, includeResources bool) ([]models.Asset, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return []models.Asset{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Asset{}, err
	}
	defer tx.Rollback()

	assets, err := repository.GetCollectionDescendantAssets(tx, collectionId, includeResources)
	if err != nil {
		return []models.Asset{}, err
	}
	return assets, nil
}

func (t *AssetService) TestData() string {
	return "test"
}

func (t *AssetService) GetAssetsPB(projectPath string) ([]byte, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return []byte{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []byte{}, err
	}
	defer tx.Rollback()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return []byte{}, err
	}
	userData, err := repository.GetUser(tx, user.Id)
	if err != nil {
		return []byte{}, err
	}
	userRole, err := repository.GetRole(tx, userData.RoleId)
	if err != nil {
		return []byte{}, err
	}
	if userRole.ViewAsset {
		start := time.Now()
		assets, err := repository.GetAssets(tx, true)
		if err != nil {
			return []byte{}, err
		}

		pbAssets := repository.ToPbFullAssets(assets)
		pbAssetsList := &repositorypb.FullAssetList{FullAssets: pbAssets}
		pbAssetsBytes, err := proto.Marshal(pbAssetsList)
		if err != nil {
			return []byte{}, err
		}

		elapsed := time.Since(start)
		fmt.Printf("GetAssets operation took %s\n", elapsed)

		//zlib compression
		compressedData := bytes.NewBuffer(nil)
		writer := zlib.NewWriter(compressedData)
		_, err = writer.Write(pbAssetsBytes)
		if err != nil {
			return []byte{}, err
		}
		err = writer.Close()
		if err != nil {
			return []byte{}, err
		}
		compressedBytes := compressedData.Bytes()

		return compressedBytes, nil
	} else {
		assets, err := repository.GetUserAssets(tx, user.Id)
		if err != nil {
			return []byte{}, err
		}

		pbAssets := repository.ToPbFullAssets(assets)
		pbAssetsList := &repositorypb.FullAssetList{FullAssets: pbAssets}
		pbAssetsBytes, err := proto.Marshal(pbAssetsList)
		if err != nil {
			return []byte{}, err
		}

		compressedData := bytes.NewBuffer(nil)
		writer := zlib.NewWriter(compressedData)
		_, err = writer.Write(pbAssetsBytes)
		if err != nil {
			return []byte{}, err
		}
		err = writer.Close()
		if err != nil {
			return []byte{}, err
		}
		compressedBytes := compressedData.Bytes()
		return compressedBytes, nil
	}
}

// asset types
func (t *AssetService) GetAssetTypes(projectPath string) ([]models.AssetType, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return []models.AssetType{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.AssetType{}, err
	}
	defer tx.Rollback()

	assetTypes, err := repository.GetAssetTypes(tx)
	if err != nil {
		return []models.AssetType{}, err
	}
	return assetTypes, nil
}

func (t *AssetService) DeleteAssetType(projectPath, id string) error {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = repository.DeleteAssetType(tx, id)
	if err != nil {
		return err
	}
	tomb, err := repository.GetTomb(tx, id)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	enqueueTombWriteThrough(projectPath, tomb)
	return nil
}

func (t *AssetService) CreateAssetType(projectPath, name, icon string) (models.AssetType, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return models.AssetType{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.AssetType{}, err
	}
	defer tx.Rollback()

	assetTypes, err := repository.CreateAssetType(tx, "", name, icon)
	if err != nil {
		return models.AssetType{}, err
	}

	err = tx.Commit()
	if err != nil {
		return models.AssetType{}, err
	}
	enqueueAssetTypeWriteThrough(projectPath, assetTypes)
	return assetTypes, nil
}

func (t *AssetService) UpdateAssetType(projectPath, id, name, icon string) (models.AssetType, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return models.AssetType{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.AssetType{}, err
	}
	defer tx.Rollback()

	assetType, err := repository.UpdateAssetType(tx, id, name, icon)
	if err != nil {
		return models.AssetType{}, err
	}

	err = tx.Commit()
	if err != nil {
		return models.AssetType{}, err
	}
	enqueueAssetTypeWriteThrough(projectPath, assetType)
	return assetType, nil
}

func (t *AssetService) GetAssetsStates(projectPath, projectWorkingDir string, ignoreList []string) (AssetsStates, error) {
	assetsStates := AssetsStates{
		Modifieds: []AssetStateItem{},
		Fetchable: []AssetStateItem{},
		Outdated:  []AssetStateItem{},
	}

	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return assetsStates, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return assetsStates, err
	}
	defer tx.Rollback()

	if !utils.DirExists(projectWorkingDir) {
		assets := []models.Asset{}
		query := "SELECT asset_path, extension FROM full_asset WHERE trashed = 0 ORDER BY asset_path"
		err = tx.Select(&assets, query)
		if err != nil {
			return assetsStates, err
		}
		for _, asset := range assets {
			displayPath := asset.AssetPath
			if asset.Extension != "" {
				displayPath = asset.AssetPath + asset.Extension
			}
			assetsStates.Fetchable = append(assetsStates.Fetchable, AssetStateItem{
				AssetId:     asset.Id,
				AssetPath:   asset.AssetPath,
				DisplayPath: displayPath,
			})
		}
		return assetsStates, nil // No untracked items if the collection folder does not exist
	}

	assets := []models.Asset{}
	query := "SELECT * FROM full_asset WHERE trashed = 0 ORDER BY asset_path"

	err = tx.Select(&assets, query)
	if err != nil {
		return assetsStates, err
	}

	checkpointQuery := "SELECT * FROM asset_checkpoint WHERE trashed = 0 ORDER BY created_at DESC"
	assetsCheckpoints := []models.Checkpoint{}
	tx.Select(&assetsCheckpoints, checkpointQuery)

	assetCheckpoints := map[string][]models.Checkpoint{}
	for _, assetCheckpoint := range assetsCheckpoints {
		assetCheckpoints[assetCheckpoint.AssetId] = append(assetCheckpoints[assetCheckpoint.AssetId], assetCheckpoint)
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return assetsStates, err
	}

	for i, asset := range assets {
		assetFilePath, err := utils.BuildAssetPath(rootFolder, asset.CollectionPath, asset.Name, asset.Extension)
		if err != nil {
			return assetsStates, err
		}
		assets[i].FilePath = assetFilePath
		assets[i].Checkpoints = assetCheckpoints[asset.Id]

		fileStatus, err := repository.GetAssetFileStatus(&assets[i], assetCheckpoints[asset.Id])
		if err != nil {
			return assetsStates, err
		}
		if fileStatus == "modified" {
			displayPath := asset.AssetPath
			if asset.Extension != "" {
				displayPath = asset.AssetPath + asset.Extension
			}
			assetsStates.Modifieds = append(assetsStates.Modifieds, AssetStateItem{
				AssetId:     asset.Id,
				AssetPath:   asset.AssetPath,
				DisplayPath: displayPath,
			})
		} else if fileStatus == "outdated" {
			displayPath := asset.AssetPath
			if asset.Extension != "" {
				displayPath = asset.AssetPath + asset.Extension
			}
			assetsStates.Outdated = append(assetsStates.Outdated, AssetStateItem{
				AssetId:     asset.Id,
				AssetPath:   asset.AssetPath,
				DisplayPath: displayPath,
			})
		} else if fileStatus == "fetchable" {
			displayPath := asset.AssetPath
			if asset.Extension != "" {
				displayPath = asset.AssetPath + asset.Extension
			}
			assetsStates.Fetchable = append(assetsStates.Fetchable, AssetStateItem{
				AssetId:     asset.Id,
				AssetPath:   asset.AssetPath,
				DisplayPath: displayPath,
			})
		}
	}

	return assetsStates, nil
}

func (t *AssetService) GetUntrackedFiles(projectPath, projectWorkingDir string, ignoreList []string) ([]string, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return []string{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []string{}, err
	}
	defer tx.Rollback()

	untrackedFiles := []string{}
	if !utils.DirExists(projectWorkingDir) {
		return untrackedFiles, nil
	}

	clusttaIgnore := ignore.CompileIgnoreLines(ignoreList...)

	// Pre-process tracked items into maps for O(1) lookup
	absoluteTrackedFiles := make(map[string]bool)

	assets := []models.Asset{}
	query := "SELECT asset_path, extension FROM full_asset WHERE trashed = 0 ORDER BY asset_path"

	err = tx.Select(&assets, query)
	if err != nil {
		return untrackedFiles, err
	}

	for _, asset := range assets {
		absoluteAssetFilePath, err := filepath.Abs(filepath.Join(projectWorkingDir, asset.AssetPath+asset.Extension))
		if err != nil {
			return untrackedFiles, err
		}
		// absoluteAssetFilePath = utils.NormalizePath(absoluteAssetFilePath)
		absoluteTrackedFiles[absoluteAssetFilePath] = true
	}

	err = filepath.WalkDir(projectWorkingDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		if d.IsDir() {
			if strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}
		} else {
			if strings.HasPrefix(filepath.Base(path), ".") {
				return nil
			}
			relativePath, err := filepath.Rel(projectWorkingDir, path)
			if err != nil {
				return err
			}
			relativePath = utils.NormalizePath(relativePath)
			if !absoluteTrackedFiles[path] && !clusttaIgnore.MatchesPath(relativePath) {
				untrackedFiles = append(untrackedFiles, "/"+relativePath)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return untrackedFiles, nil
}

// GetSiblingAssetNames returns the names of all assets in the same collection with the given extension.
// Used for client-side name validation to avoid duplicate asset names.
func (t *AssetService) GetSiblingAssetNames(projectPath, collectionId, extension string) ([]string, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []string{}, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return []string{}, err
	}
	defer tx.Rollback()

	type assetName struct {
		Name string `db:"name"`
	}
	names := []assetName{}
	err = tx.Select(&names, "SELECT name FROM asset WHERE collection_id = ? AND extension = ?", collectionId, extension)
	if err != nil {
		return []string{}, err
	}

	result := make([]string, len(names))
	for i, n := range names {
		result[i] = n.Name
	}
	return result, nil
}
