package repository

import (
	"os"

	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func GetAssetFileStatus(asset *models.Asset, checkpoints []models.Checkpoint) (string, error) {
	filePath := asset.GetFilePath()
	if asset.IsLink && utils.IsValidPointer(asset.Pointer) {
		return "normal", nil
	}
	if asset.IsLink && !utils.IsValidPointer(asset.Pointer) {
		return "missing", nil
	}

	_, err := os.Stat(filePath)
	isMissing := os.IsNotExist(err)

	// isMissing := !utils.FileExists(filePath)

	if isMissing && len(checkpoints) == 0 {
		return "missing", nil
	} else if isMissing && len(checkpoints) != 0 {
		return "fetchable", nil
	}

	fileHash, err := utils.GenerateXXHashChecksum(filePath)
	if err != nil {
		return "", err
	}
	for i, checkpoint := range checkpoints {
		if fileHash == checkpoint.XXHashChecksum {
			if i == 0 {
				return "normal", nil
			} else {
				return "outdated", nil

			}
		}
	}
	return "modified", nil
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
