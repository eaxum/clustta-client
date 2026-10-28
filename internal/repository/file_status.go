package repository

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const fileStatusCacheLimit = 1024

type fileStatusCacheEntry struct {
	fileInfo            os.FileInfo
	fileSize            int64
	modifiedAt          int64
	checkpointSignature string
	status              string
}

var fileStatusCache = struct {
	sync.RWMutex
	entries map[string]fileStatusCacheEntry
}{entries: map[string]fileStatusCacheEntry{}}

func GetAssetFileStatus(asset *models.Asset, checkpoints []models.Checkpoint) (string, error) {
	return getAssetFileStatus(asset, checkpoints, utils.GenerateXXHashChecksum)
}

func getAssetFileStatus(
	asset *models.Asset,
	checkpoints []models.Checkpoint,
	generateChecksum func(string) (string, error),
) (string, error) {
	filePath := asset.GetFilePath()
	if asset.IsLink && utils.IsValidPointer(asset.Pointer) {
		return "normal", nil
	}
	if asset.IsLink && !utils.IsValidPointer(asset.Pointer) {
		return "missing", nil
	}

	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		if len(checkpoints) == 0 {
			return "missing", nil
		}
		return "fetchable", nil
	}
	if err != nil {
		return "", err
	}

	cacheKey := fileStatusCacheKey(filePath)
	checkpointSignature := fileStatusCheckpointSignature(checkpoints)
	if status, found := cachedFileStatus(
		cacheKey,
		fileInfo,
		fileInfo.ModTime().UnixNano(),
		checkpointSignature,
	); found {
		return status, nil
	}

	fileHash, err := generateChecksum(filePath)
	if err != nil {
		return "", err
	}
	status := "modified"
	for i, checkpoint := range checkpoints {
		if fileHash != checkpoint.XXHashChecksum {
			continue
		}
		status = "outdated"
		if i == 0 {
			status = "normal"
		}
		break
	}
	storeFileStatus(cacheKey, fileStatusCacheEntry{
		fileInfo:            fileInfo,
		fileSize:            fileInfo.Size(),
		modifiedAt:          fileInfo.ModTime().UnixNano(),
		checkpointSignature: checkpointSignature,
		status:              status,
	})
	return status, nil
}

func fileStatusCacheKey(filePath string) string {
	cleanPath, err := filepath.Abs(filepath.Clean(filePath))
	if err != nil {
		cleanPath = filepath.Clean(filePath)
	}
	if runtime.GOOS == "windows" {
		return strings.ToLower(cleanPath)
	}
	return cleanPath
}

func fileStatusCheckpointSignature(checkpoints []models.Checkpoint) string {
	var signature strings.Builder
	for _, checkpoint := range checkpoints {
		signature.WriteString(checkpoint.Id)
		signature.WriteByte(0)
		signature.WriteString(checkpoint.XXHashChecksum)
		signature.WriteByte(0)
	}
	return signature.String()
}

func cachedFileStatus(
	cacheKey string,
	fileInfo os.FileInfo,
	modifiedAt int64,
	checkpointSignature string,
) (string, bool) {
	fileStatusCache.RLock()
	defer fileStatusCache.RUnlock()
	entry, found := fileStatusCache.entries[cacheKey]
	if !found || !os.SameFile(entry.fileInfo, fileInfo) ||
		entry.fileSize != fileInfo.Size() ||
		entry.modifiedAt != modifiedAt ||
		entry.checkpointSignature != checkpointSignature {
		return "", false
	}
	return entry.status, true
}

func storeFileStatus(cacheKey string, entry fileStatusCacheEntry) {
	fileStatusCache.Lock()
	defer fileStatusCache.Unlock()
	_, exists := fileStatusCache.entries[cacheKey]
	if !exists && len(fileStatusCache.entries) >= fileStatusCacheLimit {
		fileStatusCache.entries = map[string]fileStatusCacheEntry{}
	}
	fileStatusCache.entries[cacheKey] = entry
}

func GetAssetState(tx *sqlx.Tx, assetId string) (string, error) {
	// Fetch checkpoints for this asset
	checkpointQuery := "SELECT * FROM asset_checkpoint WHERE asset_id = ? AND trashed = 0 ORDER BY created_at DESC"
	checkpoints := []models.Checkpoint{}
	err := tx.Select(&checkpoints, checkpointQuery, assetId)
	if err != nil {
		return "", err
	}

	// Fetch the asset
	assetQuery := "SELECT * FROM full_asset WHERE id = ?"
	asset := models.Asset{}
	err = tx.Get(&asset, assetQuery, assetId)
	if err != nil {
		return "", err
	}

	// Get root folder for building file path
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return "", err
	}

	// Build file path
	assetFilePath, err := utils.BuildAssetPath(rootFolder, asset.CollectionPath, asset.Name, asset.Extension)
	if err != nil {
		return "", err
	}
	asset.FilePath = assetFilePath

	// Compute status
	status, err := GetAssetFileStatus(&asset, checkpoints)
	if err != nil {
		return "", err
	}

	return status, nil
}

func GetFilesStatus(tx *sqlx.Tx, assetIds []string) (map[string]string, error) {
	assetFilesStatus := map[string]string{}

	checkpointQuery := "SELECT * FROM asset_checkpoint WHERE trashed = 0 ORDER BY created_at DESC"
	assetsCheckpoints := []models.Checkpoint{}
	tx.Select(&assetsCheckpoints, checkpointQuery)

	assetCheckpoints := map[string][]models.Checkpoint{}
	for _, assetCheckpoint := range assetsCheckpoints {
		assetCheckpoints[assetCheckpoint.AssetId] = append(assetCheckpoints[assetCheckpoint.AssetId], assetCheckpoint)
	}

	for _, assetId := range assetIds {
		query := `SELECT 
			t.id,
			t.name,
			t.description,
			t.created_at,
			t.mtime,
			t.extension,
			t.is_link,
			t.pointer,
			t.status_id,
			t.collection_path
		FROM 
			full_asset t
		WHERE 
			t.id = ?;`
		// t.trashed = 0;`
		asset := models.Asset{}
		err := tx.Get(&asset, query, assetId)
		if err != nil {
			return assetFilesStatus, err
		}

		// assetFilePath, err := utils.BuildAssetPath(tx, asset.CollectionPath, asset.Name, asset.Extension)
		// if err != nil {
		// 	return assetFilesStatus, err
		// }
		// asset.FilePath = assetFilePath

		status, err := GetAssetFileStatus(&asset, assetCheckpoints[asset.Id])
		if err != nil {
			return assetFilesStatus, err
		}
		assetFilesStatus[assetId] = status
	}
	return assetFilesStatus, nil
}
