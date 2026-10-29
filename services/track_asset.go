package services

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"clustta/internal/error_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/transfer"
	"clustta/internal/utils"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	initialCheckpointComment = "v0001"
	trackChunkProgressStart  = 10
	trackChunkProgressRange  = 75
)

type TrackAssetOptions struct {
	FilePath            string
	ContentPath         string
	Message             string
	PreviewPath         string
	GroupID             string
	SourceCheckpointID  string
	AssignToCurrentUser bool
}

type TrackAssetResult struct {
	Asset      models.Asset
	Checkpoint models.Checkpoint
	Created    bool
}

// TrackAsset adopts one existing project file and creates its initial checkpoint.
func TrackAsset(
	projectPath string,
	options TrackAssetOptions,
	progress func(int),
) (TrackAssetResult, error) {
	releaseTransfer, err := transfer.Exclusive(projectPath)
	if err != nil {
		return TrackAssetResult{}, err
	}
	defer releaseTransfer()
	return trackAsset(projectPath, options, progress)
}

func trackAsset(
	projectPath string,
	options TrackAssetOptions,
	progress func(int),
) (TrackAssetResult, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return TrackAssetResult{}, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return TrackAssetResult{}, err
	}
	defer tx.Rollback()

	user, role, err := activeAssetRole(tx)
	if err != nil {
		return TrackAssetResult{}, err
	}
	if err = authorizeTrackAssetRole(role, options.AssignToCurrentUser); err != nil {
		return TrackAssetResult{}, err
	}

	workingDirectory, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return TrackAssetResult{}, err
	}
	relativePath, absolutePath, err := trackAssetPaths(workingDirectory, options.FilePath)
	if err != nil {
		return TrackAssetResult{}, err
	}
	fileInfo, err := os.Stat(absolutePath)
	if err != nil {
		return TrackAssetResult{}, fmt.Errorf("asset file is not available: %w", err)
	}
	if !fileInfo.Mode().IsRegular() {
		return TrackAssetResult{}, errors.New("asset path must identify a regular file")
	}
	if fileInfo.Size() == 0 {
		return TrackAssetResult{}, errors.New("asset file cannot be empty")
	}

	collectionID, err := resolveTrackAssetCollection(tx, relativePath, user, role)
	if err != nil {
		return TrackAssetResult{}, err
	}
	assetName := strings.TrimSuffix(path.Base(relativePath), path.Ext(relativePath))
	extension := path.Ext(relativePath)
	existingAsset, err := repository.GetAssetByName(tx, assetName, collectionID, extension)
	if err == nil {
		if options.AssignToCurrentUser {
			if existingAsset.AssigneeId != "" && existingAsset.AssigneeId != user.Id {
				return TrackAssetResult{}, errors.New("file is already tracked and is assigned to another user")
			}
			if existingAsset.AssigneeId == "" {
				if err = repository.ToggleIsTask(tx, existingAsset.Id, true); err != nil {
					return TrackAssetResult{}, err
				}
				if err = repository.AssignAsset(tx, existingAsset.Id, user.Id); err != nil {
					return TrackAssetResult{}, err
				}
				existingAsset, err = repository.GetAsset(tx, existingAsset.Id)
				if err != nil {
					return TrackAssetResult{}, err
				}
			}
		}
		checkpoint, checkpointErr := repository.GetLatestCheckpoint(tx, existingAsset.Id)
		if checkpointErr != nil {
			return TrackAssetResult{}, checkpointErr
		}
		if err = tx.Commit(); err != nil {
			return TrackAssetResult{}, err
		}
		return TrackAssetResult{
			Asset:      existingAsset,
			Checkpoint: checkpoint,
		}, nil
	}
	if !errors.Is(err, error_service.ErrAssetNotFound) {
		return TrackAssetResult{}, err
	}

	previewID := ""
	if options.PreviewPath != "" {
		preview, previewErr := repository.CreatePreview(tx, options.PreviewPath)
		if previewErr != nil {
			return TrackAssetResult{}, previewErr
		}
		previewID = preview.Hash
	}

	message := strings.TrimSpace(options.Message)
	if message == "" {
		message = initialCheckpointComment
	}
	groupID := options.GroupID
	if groupID == "" {
		groupID = uuid.NewString()
	}
	assetID := uuid.NewString()
	assetType, err := repository.GetAssetTypeByName(tx, "generic")
	if err != nil {
		return TrackAssetResult{}, err
	}
	status, err := repository.GetStatusByShortName(tx, "todo")
	if err != nil {
		return TrackAssetResult{}, err
	}
	callback := func(current, total int, _, _ string) {
		if progress == nil || total == 0 {
			return
		}
		progress(trackChunkProgressStart + current*trackChunkProgressRange/total)
	}
	contentPath := options.ContentPath
	if contentPath == "" {
		contentPath = absolutePath
	}
	if err = repository.CreateAssetFast(
		tx,
		assetID,
		assetName,
		assetType.Id,
		collectionID,
		!options.AssignToCurrentUser,
		"",
		contentPath,
		previewID,
		user.Id,
		message,
		groupID,
		relativePath,
		status.Id,
		callback,
	); err != nil {
		return TrackAssetResult{}, err
	}
	if options.AssignToCurrentUser {
		if err = repository.AssignAsset(tx, assetID, user.Id); err != nil {
			return TrackAssetResult{}, err
		}
	}

	checkpoint, err := repository.GetLatestCheckpoint(tx, assetID)
	if err != nil {
		return TrackAssetResult{}, err
	}
	if options.SourceCheckpointID != "" {
		if err = authorizeCheckpointSourceTx(tx, options.SourceCheckpointID); err != nil {
			return TrackAssetResult{}, err
		}
		if err = repository.UpdateCheckpoint(
			tx,
			checkpoint.Id,
			checkpoint.Comment,
			&options.SourceCheckpointID,
		); err != nil {
			return TrackAssetResult{}, err
		}
		checkpoint, err = repository.GetCheckpoint(tx, checkpoint.Id)
		if err != nil {
			return TrackAssetResult{}, err
		}
	}
	asset, err := repository.GetAsset(tx, assetID)
	if err != nil {
		return TrackAssetResult{}, err
	}
	if err = tx.Commit(); err != nil {
		return TrackAssetResult{}, err
	}
	return TrackAssetResult{
		Asset:      asset,
		Checkpoint: checkpoint,
		Created:    true,
	}, nil
}

func authorizeTrackAssetRole(role models.Role, assignToCurrentUser bool) error {
	permissions := []struct {
		allowed bool
		name    string
	}{
		{allowed: role.CreateAsset, name: "create_asset"},
		{allowed: role.CreateCheckpoint, name: "create_checkpoint"},
	}
	for _, permission := range permissions {
		if !permission.allowed {
			return fmt.Errorf("user does not have %s permission", permission.name)
		}
	}
	if assignToCurrentUser && !role.AssignAsset {
		return errors.New("user does not have assign_asset permission")
	}
	return nil
}

func resolveTrackAssetCollection(
	tx *sqlx.Tx,
	relativePath string,
	user models.User,
	role models.Role,
) (string, error) {
	collections := []models.Collection{}
	if err := tx.Select(&collections, "SELECT id, collection_path FROM full_collection"); err != nil {
		return "", err
	}
	collectionIDs := make(map[string]string, len(collections))
	for _, collection := range collections {
		collectionIDs[collection.CollectionPath] = collection.Id
	}

	collectionPaths := utils.GetCollectionPaths(utils.GetParent(relativePath))
	parentID := ""
	firstMissingIndex := len(collectionPaths)
	for index, collectionPath := range collectionPaths {
		collectionID, exists := collectionIDs[collectionPath]
		if !exists {
			firstMissingIndex = index
			break
		}
		parentID = collectionID
	}
	if err := authorizeTrackAssetCollectionScope(tx, user, role, parentID); err != nil {
		return "", err
	}
	if firstMissingIndex == len(collectionPaths) {
		return parentID, nil
	}
	if !role.CreateCollection {
		return "", errors.New("user does not have create_collection permission")
	}
	collectionType, err := repository.GetCollectionTypeByName(tx, "generic")
	if err != nil {
		return "", err
	}
	for _, collectionPath := range collectionPaths[firstMissingIndex:] {
		collectionID := uuid.NewString()
		collectionName := path.Base(strings.TrimSuffix(collectionPath, "/"))
		if err = repository.CreateCollectionFast(
			tx,
			collectionID,
			collectionName,
			"",
			collectionType.Id,
			parentID,
			"",
			false,
		); err != nil {
			return "", err
		}
		parentID = collectionID
	}
	return parentID, nil
}

func authorizeTrackAssetCollectionScope(
	tx *sqlx.Tx,
	user models.User,
	role models.Role,
	collectionID string,
) error {
	if role.Name == "admin" {
		return nil
	}
	if collectionID == "" {
		return errors.New("user cannot create assets at project root")
	}
	allowedCollectionIDs, err := repository.GetUserCanModifyCollectionIds(tx, user.Id)
	if err != nil {
		return err
	}
	if _, allowed := allowedCollectionIDs[collectionID]; !allowed {
		return errors.New("user cannot create assets outside assigned collection scope")
	}
	return nil
}

func trackAssetPaths(workingDirectory, filePath string) (string, string, error) {
	workingDirectory = canonicalTrackPath(workingDirectory)
	filePath = canonicalTrackPath(filePath)
	relativePath, err := filepath.Rel(workingDirectory, filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve asset path: %w", err)
	}
	if relativePath == "." || relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return "", "", errors.New("asset file must be inside the project working directory")
	}
	return filepath.ToSlash(relativePath), filePath, nil
}

func canonicalTrackPath(filePath string) string {
	absolutePath, err := filepath.Abs(filepath.Clean(filePath))
	if err != nil {
		absolutePath = filepath.Clean(filePath)
	}
	resolvedPath, err := filepath.EvalSymlinks(absolutePath)
	if err == nil {
		absolutePath = resolvedPath
	}
	return absolutePath
}
