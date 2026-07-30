package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"clustta/internal/base_service"
	"clustta/internal/error_service"
	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

type Dependency struct {
	Id       string `json:"id"`
	TypeId   string `json:"type_id"`
	TypeName string `json:"type_name"`
}
type AssetTags struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func CreateAssetFast(
	tx *sqlx.Tx, id, name, assetTypeId, collectionId string, isResource bool, description, template_file_path string, previewId, userId, comment, checkpointGroupId, assetPath, statusId string,
	callback func(int, int, string, string),
) error {
	if id == "" {
		id = uuid.New().String()
	}
	if checkpointGroupId == "" {
		checkpointGroupId = uuid.New().String()
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("asset name cannot be empty")
	}

	if collectionId != "" {
		parent := models.Collection{}
		err := base_service.Get(tx, "collection", collectionId, &parent)
		if err != nil {
			return err
		}
	}

	extension := filepath.Ext(template_file_path)

	if template_file_path == "" {
		return errors.New("template file does not exist")
	}

	chunkSequence, err := StoreFileChunks(tx, template_file_path, callback)
	if err != nil {
		return err
	}
	checksum, err := utils.GenerateXXHashChecksum(template_file_path)
	if err != nil {
		return err
	}
	fileInfo, err := os.Stat(template_file_path)
	if err != nil {
		return err
	}
	timeModified := int(fileInfo.ModTime().Unix())
	fileSize := int(fileInfo.Size())

	// Skip 0KB files and log them
	if fileSize == 0 {
		fmt.Printf("Skipping 0KB file: %s\n", template_file_path)
		return nil
	}

	params := map[string]any{
		"id":            id,
		"created_at":    utils.GetCurrentTime(),
		"name":          name,
		"description":   description,
		"extension":     extension,
		"asset_type_id": assetTypeId,
		"collection_id": collectionId,
		"is_resource":   isResource,
		"status_id":     statusId,
		"pointer":       "",
		"is_link":       false,
		"preview_id":    previewId,
	}
	err = base_service.Create(tx, "asset", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return error_service.ErrAssetExists
			}
		}
		return err
	}

	author_id := userId
	if comment == "" {
		comment = "Asset created"
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return err
	}
	assetFilePath := filepath.Join(rootFolder, assetPath)

	// fmt.Printf("Before CreateNewAssetCheckpoint - extension: %s, template_file_path: %s, assetPath: %s\n, assetFilePath: %s\n", extension, template_file_path, assetPath, assetFilePath)
	err = CreateNewAssetCheckpoint(tx, id, comment, chunkSequence, checksum, timeModified, fileSize, assetFilePath, author_id, "", checkpointGroupId, func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		// fmt.Printf("Error creating new asset checkpoint for asset %s: %v\n", id, err)
		fmt.Printf("Error creating new asset checkpoint for asset %s: %v\n", assetFilePath, err)
		return err
	}
	return nil
}

func CreateAsset(
	tx *sqlx.Tx, id, name, assetTypeId, collectionId string, isResource bool,
	templateId, description, template_file_path string, tags []string, pointer string, isLink bool, previewId, userId, comment, checkpointGroupId string,
	callback func(int, int, string, string),
) (models.Asset, error) {
	if checkpointGroupId == "" {
		checkpointGroupId = uuid.New().String()
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return models.Asset{}, errors.New("asset name cannot be empty")
	}

	if isLink && !utils.IsValidPointer(pointer) {
		return models.Asset{}, errors.New("invalid pointer, path does not exist")
	}

	if !isLink && templateId == "" && template_file_path == "" {
		return models.Asset{}, errors.New("template not provided")
	}

	if collectionId != "" {
		parent := models.Collection{}
		err := base_service.Get(tx, "collection", collectionId, &parent)
		if err != nil {
			return models.Asset{}, err
		}
	}

	asset := models.Asset{}

	chunkSequence := ""
	checksum := ""
	timeModified := 0
	fileSize := 0
	extension := ""
	if !isLink {
		if template_file_path == "" {
			if templateId == "" {
				return models.Asset{}, errors.New("no template provided")
			}
			template, err := GetTemplate(tx, templateId)
			if err != nil {
				return models.Asset{}, err
			}

			// if !utils.NonCaseSensitiveContains(tags, template.Name) {
			// 	tags = append(tags, template.Name)
			// }
			extension = template.Extension

			conditions := map[string]interface{}{
				"name":          name,
				"collection_id": collectionId,
				"extension":     extension,
			}
			err = base_service.GetBy(tx, "asset", conditions, &asset)
			if err == nil {
				if asset.Trashed {
					return models.Asset{}, error_service.ErrAssetExistsInTrash
				} else {
					return models.Asset{}, error_service.ErrAssetExists
				}
			}

			chunkSequence = template.Chunks
			checksum = template.XxhashChecksum
			fileSize = template.FileSize
			timeModified = int(utils.GetEpochTime())
		} else {
			if template_file_path == "" {
				return models.Asset{}, errors.New("template file does not exist")
			}
			extension = filepath.Ext(template_file_path)

			conditions := map[string]interface{}{
				"name":          name,
				"collection_id": collectionId,
				"extension":     extension,
			}
			err := base_service.GetBy(tx, "asset", conditions, &asset)
			if err == nil {
				if asset.Trashed {
					return models.Asset{}, error_service.ErrAssetExistsInTrash
				} else {
					return models.Asset{}, error_service.ErrAssetExists
				}
			}

			Sequence, err := StoreFileChunks(tx, template_file_path, callback)
			chunkSequence = Sequence
			if err != nil {
				return models.Asset{}, err
			}
			checksum, err = utils.GenerateXXHashChecksum(template_file_path)
			if err != nil {
				return models.Asset{}, err
			}
			fileInfo, err := os.Stat(template_file_path)
			if err != nil {
				return models.Asset{}, err
			}
			timeModified = int(fileInfo.ModTime().Unix())
			fileSize = int(fileInfo.Size())
		}
	} else {
		conditions := map[string]interface{}{
			"name":          name,
			"collection_id": collectionId,
		}
		err := base_service.GetBy(tx, "asset", conditions, &asset)
		if err == nil {
			if asset.Trashed {
				return models.Asset{}, error_service.ErrAssetExistsInTrash
			} else {
				return models.Asset{}, error_service.ErrAssetExists
			}
		}
	}
	status, err := GetStatusByShortName(tx, "todo")
	if err != nil {
		return models.Asset{}, err
	}
	statusId := status.Id
	params := map[string]any{
		"id":            id,
		"created_at":    utils.GetCurrentTime(),
		"name":          name,
		"description":   description,
		"extension":     extension,
		"asset_type_id": assetTypeId,
		"collection_id": collectionId,
		"is_resource":   isResource,
		"status_id":     statusId,
		"pointer":       pointer,
		"is_link":       isLink,
		"preview_id":    previewId,
	}
	err = base_service.Create(tx, "asset", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return models.Asset{}, error_service.ErrAssetExists
			}
		}
		return models.Asset{}, err
	}
	asset, err = GetAssetByName(tx, name, collectionId, extension)
	if err != nil {
		return models.Asset{}, err
	}
	// tx.Get(&asset, "SELECT * FROM asset WHERE name = ?", name)

	// if isLink {
	// 	if !utils.NonCaseSensitiveContains(tags, "link") {
	// 		tags = append(tags, "link")
	// 	}
	// }
	// if !isLink && pointer != "" {
	// 	if !utils.NonCaseSensitiveContains(tags, "tracked") {
	// 		tags = append(tags, "tracked")
	// 	}
	// }

	if len(tags) > 0 {
		for _, tag := range tags {
			err = AddTagToAsset(tx, asset.Id, tag)
			if err != nil {
				return models.Asset{}, err
			}
			asset.Tags = append(asset.Tags, tag)
		}
	}
	if !isLink {
		author_id := userId
		if comment == "" {
			comment = "Asset created"
		}
		_, err = CreateCheckpoint(tx, asset.Id, comment, chunkSequence, checksum, timeModified, fileSize, asset.FilePath, author_id, "", checkpointGroupId, func(i1, i2 int, s1, s2 string) {})
		if err != nil {
			return models.Asset{}, err
		}
	}
	return asset, nil
}

func AddAsset(
	tx *sqlx.Tx, id string, createdAt string, name string, assetTypeId, collectionId string,
	statusId string, extension string, description string, tags []string,
	pointer string, isLink bool, assignee_id string, previewId string,
) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("asset name cannot be empty")
	}

	if isLink && !utils.IsValidPointer(pointer) {
		return errors.New("invalid pointer, path does not exist")
	}

	params := map[string]interface{}{
		"id":            id,
		"created_at":    createdAt,
		"name":          name,
		"description":   description,
		"extension":     extension,
		"asset_type_id": assetTypeId,
		"collection_id": collectionId,
		"status_id":     statusId,
		"pointer":       pointer,
		"is_link":       isLink,
		"assignee_id":   assignee_id,
		"preview_id":    previewId,
	}
	err := base_service.Create(tx, "asset", params)
	if err != nil {
		return err
	}
	return nil
}

func GetAsset(tx *sqlx.Tx, id string) (models.Asset, error) {
	asset := models.Asset{}

	query := "SELECT * FROM full_asset WHERE id = ?"

	err := tx.Get(&asset, query, id)
	if err != nil && err == sql.ErrNoRows {
		return models.Asset{}, error_service.ErrAssetNotFound
	} else if err != nil {
		return models.Asset{}, err
	}

	if asset.TagsRaw != "[]" {
		assetTags := []AssetTags{}
		err = json.Unmarshal([]byte(asset.TagsRaw), &assetTags)
		if err != nil {
			return asset, err
		}
		for _, assetTag := range assetTags {
			asset.Tags = append(asset.Tags, assetTag.Name)
		}
	} else {
		asset.Tags = []string{} // Ensure it's initialized as an empty slice
	}

	if asset.CollectionDependenciesRaw != "[]" {
		collectionDependencies := []Dependency{}
		err = json.Unmarshal([]byte(asset.CollectionDependenciesRaw), &collectionDependencies)
		if err != nil {
			return asset, err
		}
		for _, collectionDependency := range collectionDependencies {
			asset.CollectionDependencies = append(asset.CollectionDependencies, collectionDependency.Id)
		}
	} else {
		asset.CollectionDependencies = []string{} // Ensure it's initialized as an empty slice
	}

	if asset.DependenciesRaw != "[]" {
		assetDependencies := []Dependency{}
		err = json.Unmarshal([]byte(asset.DependenciesRaw), &assetDependencies)
		if err != nil {
			return asset, err
		}
		for _, assetDependency := range assetDependencies {
			asset.Dependencies = append(asset.Dependencies, assetDependency.Id)
		}
	} else {
		asset.Dependencies = []string{} // Ensure it's initialized as an empty slice
	}

	status, err := GetStatus(tx, asset.StatusId)
	if err != nil {
		return asset, err
	}
	asset.Status = status
	asset.StatusShortName = status.ShortName

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return asset, err
	}

	assetFilePath, err := utils.BuildAssetPath(rootFolder, asset.CollectionPath, asset.Name, asset.Extension)
	if err != nil {
		return asset, err
	}
	asset.FilePath = assetFilePath
	checkpoints, err := GetCheckpoints(tx, asset.Id, false)
	if err != nil {
		return asset, err
	}
	asset.Checkpoints = checkpoints
	fileStatus, err := GetAssetFileStatus(&asset, checkpoints)
	if err != nil {
		return asset, err
	}
	asset.FileStatus = fileStatus
	return asset, nil
}

func GetSimpleAssets(tx *sqlx.Tx) ([]models.Asset, error) {
	assets := []models.Asset{}

	query := "SELECT * FROM asset"

	err := tx.Select(&assets, query)
	if err != nil && err == sql.ErrNoRows {
		return []models.Asset{}, nil
	} else if err != nil {
		return []models.Asset{}, err
	}
	return assets, nil
}
func GetSimpleAsset(tx *sqlx.Tx, id string) (models.Asset, error) {
	asset := models.Asset{}

	query := "SELECT * FROM asset WHERE id = ?"

	err := tx.Get(&asset, query, id)
	if err != nil && err == sql.ErrNoRows {
		return models.Asset{}, error_service.ErrAssetNotFound
	} else if err != nil {
		return models.Asset{}, err
	}
	return asset, nil
}

func GetAssetByName(tx *sqlx.Tx, name, collectionId string, extension string) (models.Asset, error) {
	asset := models.Asset{}
	query := "SELECT * FROM full_asset WHERE name = ? AND collection_id = ? AND extension = ?"

	err := tx.Get(&asset, query, name, collectionId, extension)
	if err != nil && err == sql.ErrNoRows {
		return models.Asset{}, error_service.ErrAssetNotFound
	} else if err != nil {
		return models.Asset{}, err
	}

	if asset.TagsRaw != "[]" {
		assetTags := []AssetTags{}
		err = json.Unmarshal([]byte(asset.TagsRaw), &assetTags)
		if err != nil {
			return asset, err
		}
		for _, assetTag := range assetTags {
			asset.Tags = append(asset.Tags, assetTag.Name)
		}
	} else {
		asset.Tags = []string{} // Ensure it's initialized as an empty slice
	}

	if asset.CollectionDependenciesRaw != "[]" {
		collectionDependencies := []Dependency{}
		err = json.Unmarshal([]byte(asset.CollectionDependenciesRaw), &collectionDependencies)
		if err != nil {
			return asset, err
		}
		for _, collectionDependency := range collectionDependencies {
			asset.CollectionDependencies = append(asset.CollectionDependencies, collectionDependency.Id)
		}
	} else {
		asset.CollectionDependencies = []string{} // Ensure it's initialized as an empty slice
	}
	if asset.DependenciesRaw != "[]" {
		assetDependencies := []Dependency{}
		err = json.Unmarshal([]byte(asset.DependenciesRaw), &assetDependencies)
		if err != nil {
			return asset, err
		}
		for _, assetDependency := range assetDependencies {
			asset.Dependencies = append(asset.Dependencies, assetDependency.Id)
		}
	} else {
		asset.Dependencies = []string{} // Ensure it's initialized as an empty slice
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return asset, err
	}

	assetFilePath, err := utils.BuildAssetPath(rootFolder, asset.CollectionPath, asset.Name, asset.Extension)
	if err != nil {
		return asset, err
	}
	checkpoints, err := GetCheckpoints(tx, asset.Id, false)
	if err != nil {
		return asset, err
	}
	asset.Checkpoints = checkpoints
	asset.FilePath = assetFilePath
	fileStatus, err := GetAssetFileStatus(&asset, checkpoints)
	if err != nil {
		return asset, err
	}
	asset.FileStatus = fileStatus
	return asset, nil
}

func GetAssetByPath(tx *sqlx.Tx, assetPath, extension string) (models.Asset, error) {
	asset := models.Asset{}
	query := "SELECT * FROM full_asset WHERE asset_path = ? AND extension = ?"

	err := tx.Get(&asset, query, assetPath, extension)
	if err != nil && err == sql.ErrNoRows {
		return models.Asset{}, error_service.ErrAssetNotFound
	} else if err != nil {
		return models.Asset{}, err
	}

	if asset.TagsRaw != "[]" {
		assetTags := []AssetTags{}
		err = json.Unmarshal([]byte(asset.TagsRaw), &assetTags)
		if err != nil {
			return asset, err
		}
		for _, assetTag := range assetTags {
			asset.Tags = append(asset.Tags, assetTag.Name)
		}
	} else {
		asset.Tags = []string{} // Ensure it's initialized as an empty slice
	}

	if asset.CollectionDependenciesRaw != "[]" {
		collectionDependencies := []Dependency{}
		err = json.Unmarshal([]byte(asset.CollectionDependenciesRaw), &collectionDependencies)
		if err != nil {
			return asset, err
		}
		for _, collectionDependency := range collectionDependencies {
			asset.CollectionDependencies = append(asset.CollectionDependencies, collectionDependency.Id)
		}
	} else {
		asset.CollectionDependencies = []string{} // Ensure it's initialized as an empty slice
	}
	if asset.DependenciesRaw != "[]" {
		assetDependencies := []Dependency{}
		err = json.Unmarshal([]byte(asset.DependenciesRaw), &assetDependencies)
		if err != nil {
			return asset, err
		}
		for _, assetDependency := range assetDependencies {
			asset.Dependencies = append(asset.Dependencies, assetDependency.Id)
		}
	} else {
		asset.Dependencies = []string{} // Ensure it's initialized as an empty slice
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return asset, err
	}
	assetFilePath, err := utils.BuildAssetPath(rootFolder, asset.CollectionPath, asset.Name, asset.Extension)
	if err != nil {
		return asset, err
	}
	checkpoints, err := GetCheckpoints(tx, asset.Id, false)
	if err != nil {
		return asset, err
	}
	asset.Checkpoints = checkpoints
	asset.FilePath = assetFilePath
	fileStatus, err := GetAssetFileStatus(&asset, checkpoints)
	if err != nil {
		return asset, err
	}
	asset.FileStatus = fileStatus
	return asset, nil
}

func GetAssets(tx *sqlx.Tx, withDeleted bool) ([]models.Asset, error) {
	assets := []models.Asset{}
	queryWhereClause := ""
	if !withDeleted {
		queryWhereClause = "WHERE trashed = 0"
	}
	query := fmt.Sprintf("SELECT * FROM full_asset %s ORDER BY asset_path", queryWhereClause)

	err := tx.Select(&assets, query)
	if err != nil {
		return assets, err
	}

	statuses, err := GetStatuses(tx)
	if err != nil {
		return assets, err
	}
	statusesMap := map[string]models.Status{}
	for _, status := range statuses {
		statusesMap[status.Id] = status
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
		return assets, err
	}

	for i, asset := range assets {
		status := statusesMap[asset.StatusId]
		assets[i].Status = status
		assets[i].StatusShortName = status.ShortName

		if assets[i].TagsRaw != "[]" {
			assetTags := []AssetTags{}
			err = json.Unmarshal([]byte(asset.TagsRaw), &assetTags)
			if err != nil {
				return assets, err
			}
			for _, assetTag := range assetTags {
				assets[i].Tags = append(assets[i].Tags, assetTag.Name)
			}
		} else {
			assets[i].Tags = []string{} // Ensure it's initialized as an empty slice
		}

		if asset.CollectionDependenciesRaw != "[]" {
			collectionDependencies := []Dependency{}
			err = json.Unmarshal([]byte(asset.CollectionDependenciesRaw), &collectionDependencies)
			if err != nil {
				return assets, err
			}
			for _, collectionDependency := range collectionDependencies {
				assets[i].CollectionDependencies = append(assets[i].CollectionDependencies, collectionDependency.Id)
			}
		} else {
			assets[i].CollectionDependencies = []string{} // Ensure it's initialized as an empty slice
		}
		if asset.DependenciesRaw != "[]" {
			assetDependencies := []Dependency{}
			err = json.Unmarshal([]byte(asset.DependenciesRaw), &assetDependencies)
			if err != nil {
				return assets, err
			}
			for _, assetDependency := range assetDependencies {
				assets[i].Dependencies = append(assets[i].Dependencies, assetDependency.Id)
			}
		} else {
			assets[i].Dependencies = []string{} // Ensure it's initialized as an empty slice
		}

		assetFilePath, err := utils.BuildAssetPath(rootFolder, asset.CollectionPath, asset.Name, asset.Extension)
		if err != nil {
			return assets, err
		}
		assets[i].FilePath = assetFilePath
		assets[i].Checkpoints = assetCheckpoints[asset.Id]

		// fileStatus, err := GetAssetFileStatus(&assets[i], assetCheckpoints[asset.Id])
		// if err != nil {
		// 	return assets, err
		// }
		// assets[i].FileStatus = fileStatus
	}
	return assets, nil
}

func OLDGetUserAssets(tx *sqlx.Tx, userId string) ([]models.Asset, error) {
	assets := []models.Asset{}
	// query := "SELECT * FROM full_asset WHERE trashed = 0 AND assignee_id = ?"
	query := fmt.Sprintf(`
		WITH RECURSIVE 
			-- First get all collection dependencies recursively
			collection_dependencies AS (
				-- Base case: Get direct collection dependencies
				SELECT 
					ed.asset_id,
					e.id as collection_id,
					0 as collection_depth,
					CAST(e.id AS TEXT) as collection_path
				FROM collection_dependency ed
				JOIN collection e ON ed.dependency_id = e.id
				WHERE e.trashed = 0

				UNION ALL

				-- Recursive case: Get child collections
				SELECT 
					ed.asset_id,
					e.id as collection_id,
					ed.collection_depth + 1,
					ed.collection_path || ',' || e.id
				FROM collection e
				JOIN collection_dependencies ed ON e.parent_id = ed.collection_id
				WHERE e.trashed = 0
				AND e.id NOT IN (
					SELECT value 
					FROM json_each('["' || REPLACE(ed.collection_path, ',', '","') || '"]')
				)
			),
			
			-- Then get all assets and their dependencies
			asset_dependencies AS (
				-- Base case: Get all assets directly assigned to the user
				SELECT 
					t.*,
					0 as dependency_level,
					0 as is_dependency,  -- Directly assigned assets
					t.id as root_asset_id,  -- Track the original assigned asset
					CAST(t.id AS TEXT) as dependency_path,  -- Track the path to prevent cycles
					NULL as via_collection  -- Track if asset came through collection dependency
				FROM full_asset t
				WHERE t.assignee_id = '%s' 
				AND t.trashed = 0

				UNION ALL

				-- Get assets through collection dependencies
				SELECT 
					t.*,
					td.dependency_level,
					1 as is_dependency,
					td.root_asset_id,
					td.dependency_path || ',' || t.id,
					ed.collection_id as via_collection
				FROM full_asset t
				JOIN collection_dependencies ed ON t.collection_id = ed.collection_id
				JOIN asset_dependencies td ON ed.asset_id = td.id
				WHERE t.trashed = 0
				AND t.id NOT IN (
					SELECT value 
					FROM json_each('["' || REPLACE(td.dependency_path, ',', '","') || '"]')
				)

				UNION ALL

				-- Get direct asset dependencies
				SELECT 
					t.*,
					td.dependency_level + 1,
					1 as is_dependency,
					td.root_asset_id,
					td.dependency_path || ',' || t.id,
					NULL as via_collection
				FROM full_asset t
				JOIN asset_dependency dep ON t.id = dep.dependency_id
				JOIN asset_dependencies td ON dep.asset_id = td.id
				WHERE t.trashed = 0
				AND t.id NOT IN (
					SELECT value 
					FROM json_each('["' || REPLACE(td.dependency_path, ',', '","') || '"]')
				)
			),
			
			-- Select the highest-level occurrence of each asset
			ranked_dependencies AS (
				SELECT 
					*,
					ROW_NUMBER() OVER (
						PARTITION BY id 
						ORDER BY 
							-- Prefer directly assigned assets
							CASE WHEN assignee_id = '%s' THEN 0 ELSE 1 END,
							-- Then prefer the shortest dependency chain
							dependency_level,
							-- For equal levels, prefer alphabetical order of root asset
							root_asset_id
					) as rank
				FROM asset_dependencies
			)
		SELECT 
			id,
			created_at,
			mtime,
			name,
			description,
			extension,
			is_resource,
			is_link,
			pointer,
			status_id,
			asset_type_id,
			collection_id,
			assignee_id,
			assigner_id,
			preview_id,
			trashed,
			synced,
			asset_type_icon,
			asset_type_name,
			collection_name,
			preview_extension,
			preview,
			COALESCE(collection_path, '') as collection_path, -- Handle NULL
			asset_path,
			assignee_name,
			assignee_email,
			assigner_name,
			assigner_email,
			tags,
			dependencies,
			collection_dependencies,
			dependency_level,
			is_dependency
		FROM ranked_dependencies
		WHERE rank = 1
		ORDER BY 
			dependency_level,
			name;
	`, userId, userId)

	err := tx.Select(&assets, query)
	if err != nil {
		return assets, err
	}

	statuses, err := GetStatuses(tx)
	if err != nil {
		return assets, err
	}
	statusesMap := map[string]models.Status{}
	for _, status := range statuses {
		statusesMap[status.Id] = status
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
		return assets, err
	}
	for i, asset := range assets {
		status := statusesMap[asset.StatusId]
		assets[i].StatusShortName = status.ShortName
		assets[i].Status = status

		if assets[i].TagsRaw != "[]" {
			assetTags := []AssetTags{}
			err = json.Unmarshal([]byte(asset.TagsRaw), &assetTags)
			if err != nil {
				return assets, err
			}
			for _, assetTag := range assetTags {
				assets[i].Tags = append(assets[i].Tags, assetTag.Name)
			}
		} else {
			assets[i].Tags = []string{} // Ensure it's initialized as an empty slice
		}

		if asset.CollectionDependenciesRaw != "[]" {
			collectionDependencies := []Dependency{}
			err = json.Unmarshal([]byte(asset.CollectionDependenciesRaw), &collectionDependencies)
			if err != nil {
				return assets, err
			}
			for _, collectionDependency := range collectionDependencies {
				assets[i].CollectionDependencies = append(assets[i].CollectionDependencies, collectionDependency.Id)
			}
		} else {
			assets[i].CollectionDependencies = []string{} // Ensure it's initialized as an empty slice
		}
		if asset.DependenciesRaw != "[]" {
			assetDependencies := []Dependency{}
			err = json.Unmarshal([]byte(asset.DependenciesRaw), &assetDependencies)
			if err != nil {
				return assets, err
			}
			for _, assetDependency := range assetDependencies {
				assets[i].Dependencies = append(assets[i].Dependencies, assetDependency.Id)
			}
		} else {
			assets[i].Dependencies = []string{} // Ensure it's initialized as an empty slice
		}

		assetFilePath, err := utils.BuildAssetPath(rootFolder, asset.CollectionPath, asset.Name, asset.Extension)
		if err != nil {
			return assets, err
		}
		assets[i].FilePath = assetFilePath
		assets[i].Checkpoints = assetCheckpoints[asset.Id]

		// fileStatus := "normal"
		fileStatus, err := GetAssetFileStatus(&assets[i], assetCheckpoints[asset.Id])
		if err != nil {
			return assets, err
		}
		assets[i].FileStatus = fileStatus
	}

	return assets, nil
}

func OLDGetUserAssetsX(tx *sqlx.Tx, userId string) ([]models.Asset, error) {
	// Get all assets assigned to the user
	assignedAssetIds := []string{}
	query := "SELECT id FROM asset WHERE assignee_id = ? AND trashed = 0"
	err := tx.Select(&assignedAssetIds, query, userId)
	if err != nil {
		return nil, err
	}

	// Get all assets that are dependencies of the assigned assets and dependencies of those assets recursively
	dependencyAssetIds := []string{}
	query = `
		WITH RECURSIVE asset_dependencies AS (
			-- Base case: Get direct dependencies
			SELECT 
				td.dependency_id, 
				0 AS dependency_level, 
				CAST(td.dependency_id AS TEXT) AS dependency_path
			FROM asset_dependency td
			JOIN asset t ON td.asset_id = t.id
			WHERE t.assignee_id = ? AND t.trashed = 0

			UNION ALL

			-- Recursive case: Get dependencies of dependencies
			SELECT 
				td.dependency_id, 
				td2.dependency_level + 1 AS dependency_level, 
				td2.dependency_path || ',' || td.dependency_id AS dependency_path
			FROM asset_dependency td
			JOIN asset_dependencies td2 ON td.asset_id = td2.dependency_id
			WHERE td.dependency_id NOT IN (
				SELECT value
				FROM json_each('["' || REPLACE(td2.dependency_path, ',', '","') || '"]')
			)
		)
		SELECT DISTINCT t.id
		FROM asset t
		JOIN asset_dependencies td ON t.id = td.dependency_id
	`
	err = tx.Select(&dependencyAssetIds, query, userId)
	if err != nil {
		return nil, err
	}

	// Get all assets that are dependencies of the assigned assets through collection dependencies
	collectionAssetsDependencyIds := []string{}
	query = `
		WITH RECURSIVE collection_dependencies AS (
			SELECT ed.asset_id, e.id as collection_id, 0 as collection_depth, CAST(e.id AS TEXT) as collection_path
			FROM collection_dependency ed
			JOIN collection e ON ed.dependency_id = e.id
			WHERE e.trashed = 0

			UNION ALL

			SELECT ed.asset_id, e.id as collection_id, ed.collection_depth + 1, ed.collection_path || ',' || e.id
			FROM collection e
			JOIN collection_dependencies ed ON e.parent_id = ed.collection_id
			WHERE e.trashed = 0
			AND e.id NOT IN (
				SELECT value
				FROM json_each('["' || REPLACE(ed.collection_path, ',', '","') || '"]')
			)
		)
		SELECT DISTINCT t.id
		FROM asset t
		JOIN collection_dependencies ed ON t.collection_id = ed.collection_id
		JOIN asset_dependency td ON t.id = td.dependency_id
		WHERE td.asset_id IN (SELECT id FROM asset WHERE assignee_id = ? AND trashed = 0)
		AND t.trashed = 0
	`
	err = tx.Select(&collectionAssetsDependencyIds, query, userId)
	if err != nil {
		return nil, err
	}

	// Combine all asset IDs
	allAssetIds := make(map[string]struct{})
	for _, id := range assignedAssetIds {
		allAssetIds[id] = struct{}{}
	}
	for _, id := range dependencyAssetIds {
		allAssetIds[id] = struct{}{}
	}
	for _, id := range collectionAssetsDependencyIds {
		allAssetIds[id] = struct{}{}
	}

	// Convert map keys to slice
	assetIds := make([]string, 0, len(allAssetIds))
	for id := range allAssetIds {
		assetIds = append(assetIds, id)
	}

	// Get all assets with the collected IDs
	query = `
		SELECT * FROM full_asset WHERE id IN (SELECT value FROM json_each(?)) AND trashed = 0
	`
	assets := []models.Asset{}
	jsonAssetIds, err := json.Marshal(assetIds)
	if err != nil {
		return nil, err
	}
	err = tx.Select(&assets, query, jsonAssetIds)
	if err != nil {
		return nil, err
	}

	statuses, err := GetStatuses(tx)
	if err != nil {
		return assets, err
	}
	statusesMap := map[string]models.Status{}
	for _, status := range statuses {
		statusesMap[status.Id] = status
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
		return assets, err
	}
	for i, asset := range assets {
		status := statusesMap[asset.StatusId]
		assets[i].StatusShortName = status.ShortName
		assets[i].Status = status

		if assets[i].TagsRaw != "[]" {
			assetTags := []AssetTags{}
			err = json.Unmarshal([]byte(asset.TagsRaw), &assetTags)
			if err != nil {
				return assets, err
			}
			for _, assetTag := range assetTags {
				assets[i].Tags = append(assets[i].Tags, assetTag.Name)
			}
		} else {
			assets[i].Tags = []string{} // Ensure it's initialized as an empty slice
		}

		if asset.CollectionDependenciesRaw != "[]" {
			collectionDependencies := []Dependency{}
			err = json.Unmarshal([]byte(asset.CollectionDependenciesRaw), &collectionDependencies)
			if err != nil {
				return assets, err
			}
			for _, collectionDependency := range collectionDependencies {
				assets[i].CollectionDependencies = append(assets[i].CollectionDependencies, collectionDependency.Id)
			}
		} else {
			assets[i].CollectionDependencies = []string{} // Ensure it's initialized as an empty slice
		}
		if asset.DependenciesRaw != "[]" {
			assetDependencies := []Dependency{}
			err = json.Unmarshal([]byte(asset.DependenciesRaw), &assetDependencies)
			if err != nil {
				return assets, err
			}
			for _, assetDependency := range assetDependencies {
				assets[i].Dependencies = append(assets[i].Dependencies, assetDependency.Id)
			}
		} else {
			assets[i].Dependencies = []string{} // Ensure it's initialized as an empty slice
		}

		assetFilePath, err := utils.BuildAssetPath(rootFolder, asset.CollectionPath, asset.Name, asset.Extension)
		if err != nil {
			return assets, err
		}
		assets[i].FilePath = assetFilePath
		assets[i].Checkpoints = assetCheckpoints[asset.Id]

		// fileStatus := "normal"
		fileStatus, err := GetAssetFileStatus(&assets[i], assetCheckpoints[asset.Id])
		if err != nil {
			return assets, err
		}
		assets[i].FileStatus = fileStatus
	}

	return assets, nil
}

func getAllDependencies(assetId string, depth int, allAssetDependencies []models.AssetDependency) []string {
	if depth > 20 {
		return []string{}
	}
	dependencies := []string{}
	for _, assetDependency := range allAssetDependencies {
		if assetDependency.AssetId == assetId {
			dependencies = append(dependencies, assetDependency.DependencyId)
			dependencies = append(dependencies, getAllDependencies(assetDependency.DependencyId, depth+1, allAssetDependencies)...)
		}
	}
	return dependencies
}

func getAllCollectionAssets(collectionId string, allAssets []models.Asset, allCollections []models.Collection) []string {
	dependencies := []string{}
	for _, asset := range allAssets {
		if asset.CollectionId == collectionId {
			dependencies = append(dependencies, asset.Id)
		}
	}
	for _, collection := range allCollections {
		if collection.ParentId == collectionId {
			dependencies = append(dependencies, getAllCollectionAssets(collection.Id, allAssets, allCollections)...)
		}
	}
	return dependencies
}

func GetUserAssets(tx *sqlx.Tx, userId string) ([]models.Asset, error) {
	// Get all assets assigned to the user
	assignedAssetIds := []string{}
	query := "SELECT id FROM asset WHERE assignee_id = ? AND trashed = 0"
	err := tx.Select(&assignedAssetIds, query, userId)
	if err != nil {
		return nil, err
	}

	// Get all assets id
	allAssetsId := []string{}
	query = `SELECT id FROM asset WHERE trashed = 0`
	err = tx.Select(&allAssetsId, query)
	if err != nil {
		return nil, err
	}

	allAssetInfo := []models.Asset{}
	query = `SELECT id, collection_id FROM asset WHERE trashed = 0`
	err = tx.Select(&allAssetInfo, query)
	if err != nil {
		return nil, err
	}

	allCollectionInfo := []models.Collection{}
	query = `SELECT id, parent_id FROM collection WHERE trashed = 0`
	err = tx.Select(&allCollectionInfo, query)
	if err != nil {
		return nil, err
	}

	// All asset dependencies records
	allAssetDependencies := []models.AssetDependency{}
	query = `SELECT asset_id, dependency_id FROM asset_dependency`
	err = tx.Select(&allAssetDependencies, query)
	if err != nil {
		return nil, err
	}

	// All collection dependencies records
	allCollectionDependencies := []models.CollectionDependency{}
	query = `SELECT asset_id, dependency_id FROM collection_dependency`
	err = tx.Select(&allCollectionDependencies, query)
	if err != nil {
		return nil, err
	}

	// Get all assigned collections for the user
	assignedCollectionsIds := []string{}
	query = "select collection_id from collection_assignee where assignee_id = ?"
	err = tx.Select(&assignedCollectionsIds, query, userId)
	if err != nil {
		return nil, err
	}

	sharedCollections := []string{}
	query = "select id from collection where is_shared = 1"
	err = tx.Select(&sharedCollections, query)
	if err != nil {
		return nil, err
	}

	// process all user assets and their dependencies recursively
	dependencies := map[string]struct{}{}
	for _, assetId := range assignedAssetIds {
		dependencies[assetId] = struct{}{}
		for _, dependency := range getAllDependencies(assetId, 0, allAssetDependencies) {
			dependencies[dependency] = struct{}{}
		}
		for _, collectionDependency := range allCollectionDependencies {
			if collectionDependency.AssetId == assetId {
				collectionAssets := getAllCollectionAssets(collectionDependency.DependencyId, allAssetInfo, allCollectionInfo)
				for _, collectionAsset := range collectionAssets {
					dependencies[collectionAsset] = struct{}{}
				}
			}
		}
	}

	for _, collectionId := range assignedCollectionsIds {
		collectionAssets := getAllCollectionAssets(collectionId, allAssetInfo, allCollectionInfo)
		for _, collectionAsset := range collectionAssets {
			dependencies[collectionAsset] = struct{}{}
		}
	}

	for _, collectionId := range sharedCollections {
		collectionAssets := getAllCollectionAssets(collectionId, allAssetInfo, allCollectionInfo)
		for _, collectionAsset := range collectionAssets {
			dependencies[collectionAsset] = struct{}{}
		}
	}

	// Convert map keys to slice
	assetIds := make([]string, 0, len(dependencies))
	for id := range dependencies {
		assetIds = append(assetIds, id)
	}

	// Get all assets with the collected IDs
	query = `
		SELECT * FROM full_asset WHERE id IN (SELECT value FROM json_each(?)) AND trashed = 0
	`
	assets := []models.Asset{}
	jsonAssetIds, err := json.Marshal(assetIds)
	if err != nil {
		return nil, err
	}
	err = tx.Select(&assets, query, jsonAssetIds)
	if err != nil {
		return nil, err
	}

	statuses, err := GetStatuses(tx)
	if err != nil {
		return assets, err
	}
	statusesMap := map[string]models.Status{}
	for _, status := range statuses {
		statusesMap[status.Id] = status
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
		return assets, err
	}
	for i, asset := range assets {
		status := statusesMap[asset.StatusId]
		assets[i].StatusShortName = status.ShortName
		assets[i].Status = status

		if assets[i].TagsRaw != "[]" {
			assetTags := []AssetTags{}
			err = json.Unmarshal([]byte(asset.TagsRaw), &assetTags)
			if err != nil {
				return assets, err
			}
			for _, assetTag := range assetTags {
				assets[i].Tags = append(assets[i].Tags, assetTag.Name)
			}
		} else {
			assets[i].Tags = []string{} // Ensure it's initialized as an empty slice
		}

		if asset.CollectionDependenciesRaw != "[]" {
			collectionDependencies := []Dependency{}
			err = json.Unmarshal([]byte(asset.CollectionDependenciesRaw), &collectionDependencies)
			if err != nil {
				return assets, err
			}
			for _, collectionDependency := range collectionDependencies {
				assets[i].CollectionDependencies = append(assets[i].CollectionDependencies, collectionDependency.Id)
			}
		} else {
			assets[i].CollectionDependencies = []string{} // Ensure it's initialized as an empty slice
		}
		if asset.DependenciesRaw != "[]" {
			assetDependencies := []Dependency{}
			err = json.Unmarshal([]byte(asset.DependenciesRaw), &assetDependencies)
			if err != nil {
				return assets, err
			}
			for _, assetDependency := range assetDependencies {
				assets[i].Dependencies = append(assets[i].Dependencies, assetDependency.Id)
			}
		} else {
			assets[i].Dependencies = []string{} // Ensure it's initialized as an empty slice
		}

		assetFilePath, err := utils.BuildAssetPath(rootFolder, asset.CollectionPath, asset.Name, asset.Extension)
		if err != nil {
			return assets, err
		}
		assets[i].FilePath = assetFilePath
		assets[i].Checkpoints = assetCheckpoints[asset.Id]

		// fileStatus := "normal"
		fileStatus, err := GetAssetFileStatus(&assets[i], assetCheckpoints[asset.Id])
		if err != nil {
			return assets, err
		}
		assets[i].FileStatus = fileStatus
	}

	return assets, nil
}

// This function is meant for the get user collections to process collections proper
func GetUserAssetsMinimal(tx *sqlx.Tx, userId string) ([]models.Asset, error) {
	// Get all assets assigned to the user
	assignedAssetIds := []string{}
	query := "SELECT id FROM asset WHERE assignee_id = ? AND trashed = 0"
	err := tx.Select(&assignedAssetIds, query, userId)
	if err != nil {
		return nil, err
	}

	// Get all assets id
	allAssetsId := []string{}
	query = `SELECT id FROM asset WHERE trashed = 0`
	err = tx.Select(&allAssetsId, query)
	if err != nil {
		return nil, err
	}

	allAssetInfo := []models.Asset{}
	query = `SELECT id, collection_id FROM asset WHERE trashed = 0`
	err = tx.Select(&allAssetInfo, query)
	if err != nil {
		return nil, err
	}

	allCollectionInfo := []models.Collection{}
	query = `SELECT id, parent_id FROM collection WHERE trashed = 0`
	err = tx.Select(&allCollectionInfo, query)
	if err != nil {
		return nil, err
	}

	// All asset dependencies records
	allAssetDependencies := []models.AssetDependency{}
	query = `SELECT asset_id, dependency_id FROM asset_dependency`
	err = tx.Select(&allAssetDependencies, query)
	if err != nil {
		return nil, err
	}

	// All collection dependencies records
	allCollectionDependencies := []models.CollectionDependency{}
	query = `SELECT asset_id, dependency_id FROM collection_dependency`
	err = tx.Select(&allCollectionDependencies, query)
	if err != nil {
		return nil, err
	}

	// Get all assigned collections for the user
	assignedCollectionsIds := []string{}
	query = "select collection_id from collection_assignee where assignee_id = ?"
	err = tx.Select(&assignedCollectionsIds, query, userId)
	if err != nil {
		return nil, err
	}

	sharedCollections := []string{}
	query = "select id from collection where is_shared = 1"
	err = tx.Select(&sharedCollections, query, userId)
	if err != nil {
		return nil, err
	}

	// process all user assets and their dependencies recursively
	dependencies := map[string]struct{}{}
	for _, assetId := range assignedAssetIds {
		dependencies[assetId] = struct{}{}
		for _, dependency := range getAllDependencies(assetId, 0, allAssetDependencies) {
			dependencies[dependency] = struct{}{}
		}
		for _, collectionDependency := range allCollectionDependencies {
			if collectionDependency.AssetId == assetId {
				collectionAssets := getAllCollectionAssets(collectionDependency.DependencyId, allAssetInfo, allCollectionInfo)
				for _, collectionAsset := range collectionAssets {
					dependencies[collectionAsset] = struct{}{}
				}
			}
		}

		for _, collectionId := range assignedCollectionsIds {
			collectionAssets := getAllCollectionAssets(collectionId, allAssetInfo, allCollectionInfo)
			for _, collectionAsset := range collectionAssets {
				dependencies[collectionAsset] = struct{}{}
			}
		}

		for _, collectionId := range sharedCollections {
			collectionAssets := getAllCollectionAssets(collectionId, allAssetInfo, allCollectionInfo)
			for _, collectionAsset := range collectionAssets {
				dependencies[collectionAsset] = struct{}{}
			}
		}
	}

	// Convert map keys to slice
	assetIds := make([]string, 0, len(dependencies))
	for id := range dependencies {
		assetIds = append(assetIds, id)
	}

	// Get all assets with the collected IDs
	query = `
		SELECT id, collection_id FROM asset WHERE id IN (SELECT value FROM json_each(?)) AND trashed = 0
	`
	assets := []models.Asset{}
	jsonAssetIds, err := json.Marshal(assetIds)
	if err != nil {
		return nil, err
	}
	err = tx.Select(&assets, query, jsonAssetIds)
	if err != nil {
		return nil, err
	}
	return assets, nil
}

func GetDeletedAssets(tx *sqlx.Tx) ([]models.Asset, error) {
	assets := []models.Asset{}

	collections, err := GetCollections(tx, true)
	if err != nil {
		return assets, err
	}
	collectionsMap := map[string]models.Collection{}
	for _, collection := range collections {
		collectionsMap[collection.Id] = collection
	}

	statuses, err := GetStatuses(tx)
	if err != nil {
		return assets, err
	}
	statusesMap := map[string]models.Status{}
	for _, status := range statuses {
		statusesMap[status.Id] = status
	}

	tags, err := GetTags(tx)
	if err != nil {
		return assets, err
	}
	tagsMap := map[string]models.Tag{}
	for _, tag := range tags {
		tagsMap[tag.Id] = tag
	}

	query := "SELECT asset_id, GROUP_CONCAT(tag_id) FROM asset_tag GROUP BY asset_id"
	assetsTags := map[string][]models.Tag{}
	tx.Select(&assetsTags, query)
	for assetId, tagIds := range assetsTags {
		assetsTags[assetId] = []models.Tag{}
		for _, tagId := range tagIds {
			assetsTags[assetId] = append(assetsTags[assetId], tagsMap[tagId.Id])
		}
	}

	conditions := map[string]interface{}{
		"trashed": 1,
	}
	err = base_service.GetAllBy(tx, "asset", conditions, &assets)
	if err != nil {
		return assets, err
	}
	// tx.Select(&assets, "SELECT * FROM asset WHERE trashed = 0")
	newAssets := []models.Asset{}
	//TODO Investigate why for loop does not update data
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return assets, err
	}
	for _, asset := range assets {
		collection := collectionsMap[asset.CollectionId]
		asset.CollectionName = collection.Name

		status := statusesMap[asset.StatusId]
		asset.StatusShortName = status.ShortName

		assetTags := assetsTags[asset.Id]
		asset.Tags = []string{}
		for _, tag := range assetTags {
			asset.Tags = append(asset.Tags, tag.Name)
		}

		assetFilePath, err := utils.BuildAssetPath(rootFolder, asset.CollectionPath, asset.Name, asset.Extension)
		if err != nil {
			return assets, err
		}
		asset.FilePath = assetFilePath
		newAssets = append(newAssets, asset)
	}

	return newAssets, nil
}

func GetAssetsByCollectionId(tx *sqlx.Tx, collectionId string) ([]models.Asset, error) {
	assets := []models.Asset{}

	statuses, err := GetStatuses(tx)
	if err != nil {
		return assets, err
	}
	statusesMap := map[string]models.Status{}
	for _, status := range statuses {
		statusesMap[status.Id] = status
	}

	tags, err := GetTags(tx)
	if err != nil {
		return assets, err
	}
	tagsMap := map[string]models.Tag{}
	for _, tag := range tags {
		tagsMap[tag.Id] = tag
	}

	query := "SELECT * FROM asset_tag"
	assetsTags := []models.AssetTag{}
	tx.Select(&assetsTags, query)

	assetTags := map[string][]models.Tag{}
	for _, assetTag := range assetsTags {
		assetTags[assetTag.AssetId] = append(assetTags[assetTag.AssetId], tagsMap[assetTag.TagId])
	}

	checkpointQuery := "SELECT * FROM asset_checkpoint WHERE trashed = 0 ORDER BY created_at DESC"
	assetsCheckpoints := []models.Checkpoint{}
	tx.Select(&assetsCheckpoints, checkpointQuery)

	assetCheckpoints := map[string][]models.Checkpoint{}
	for _, assetCheckpoint := range assetsCheckpoints {
		assetCheckpoints[assetCheckpoint.AssetId] = append(assetCheckpoints[assetCheckpoint.AssetId], assetCheckpoint)
	}

	conditions := map[string]interface{}{
		"trashed":       0,
		"collection_id": collectionId,
	}
	err = base_service.GetAllBy(tx, "full_asset", conditions, &assets)
	if err != nil {
		return assets, err
	}
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return assets, err
	}
	for i, asset := range assets {
		status := statusesMap[asset.StatusId]
		assets[i].StatusShortName = status.ShortName

		assetTags := assetTags[asset.Id]
		assets[i].Tags = []string{}
		for _, tag := range assetTags {
			assets[i].Tags = append(assets[i].Tags, tag.Name)
		}
		assetFilePath, err := utils.BuildAssetPath(rootFolder, asset.CollectionPath, asset.Name, asset.Extension)
		if err != nil {
			return assets, err
		}
		assets[i].FilePath = assetFilePath

		// fileStatus := "normal"
		fileStatus, err := GetAssetFileStatus(&assets[i], assetCheckpoints[asset.Id])
		if err != nil {
			return assets, err
		}
		assets[i].FileStatus = fileStatus
	}

	return assets, nil
}

func GetAssetsByTagId(tx *sqlx.Tx, tagId string) []models.Asset {
	assets := []models.Asset{}
	tx.Select(&assets, "SELECT * FROM asset WHERE id IN (SELECT asset_id FROM asset_tag WHERE tag_id = ?)", tagId)
	return assets
}

func GetAssetsByTagName(tx *sqlx.Tx, tagName string) []models.Asset {
	assets := []models.Asset{}
	tx.Select(&assets, "SELECT * FROM asset WHERE id IN (SELECT asset_id FROM asset_tag WHERE tag_id IN (SELECT id FROM tag WHERE name = ?))", tagName)
	return assets
}

func DeleteAsset(tx *sqlx.Tx, assetId string, removeFromDir bool, recycle bool) error {
	asset, err := GetAsset(tx, assetId)
	if err != nil {
		return err
	}
	if recycle {
		err = base_service.MarkAsDeleted(tx, "asset", assetId)
		if err != nil {
			return err
		}
	} else {
		err = DeleteCheckpoints(tx, assetId)
		if err != nil {
			return err
		}
		RemoveAllTagsFromAsset(tx, assetId)
		err = base_service.Delete(tx, "asset", assetId)
		if err != nil {
			return err
		}
	}
	if removeFromDir {
		err := os.RemoveAll(asset.FilePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteCollectionAssets(tx *sqlx.Tx, collectionId string, removeFromDir bool) error {
	assets, err := GetAssetsByCollectionId(tx, collectionId)
	if err != nil {
		return err
	}
	for _, asset := range assets {
		err := DeleteAsset(tx, asset.Id, removeFromDir, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateStatus(tx *sqlx.Tx, assetId string, statusId string) error {
	params := map[string]interface{}{
		"status_id": statusId,
	}
	err := base_service.Update(tx, "asset", assetId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "asset", assetId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

// UpdateStatusOnly sets the status_id without bumping mtime.
// Used after a remote-confirmed status change to avoid flagging the row as dirty.
func UpdateStatusOnly(tx *sqlx.Tx, assetId string, statusId string) error {
	params := map[string]interface{}{
		"status_id": statusId,
	}
	return base_service.Update(tx, "asset", assetId, params)
}

func ChangeCollection(tx *sqlx.Tx, assetId string, collectionId string) error {
	oldAsset, err := GetAsset(tx, assetId)
	if err != nil {
		return err
	}

	params := map[string]interface{}{
		"collection_id": collectionId,
	}
	err = base_service.Update(tx, "asset", assetId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "asset", assetId, utils.GetEpochTime())
	if err != nil {
		return err
	}

	asset, err := GetAsset(tx, assetId)
	if err != nil {
		return err
	}

	if oldAsset.FilePath != asset.FilePath && utils.FileExists(oldAsset.FilePath) {
		newFilePath := asset.FilePath

		newAssetFolder := filepath.Dir(newFilePath)

		err := os.MkdirAll(newAssetFolder, os.ModePerm)
		if err != nil {
			return err
		}

		err = utils.RenamePathCaseSafe(oldAsset.FilePath, asset.FilePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func ChangeAssetType(tx *sqlx.Tx, assetId string, assetTypeId string) error {
	params := map[string]any{
		"asset_type_id": assetTypeId,
	}
	err := base_service.Update(tx, "asset", assetId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "asset", assetId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func ToggleIsTask(tx *sqlx.Tx, assetId string, isTask bool) error {
	params := map[string]any{
		"is_resource": !isTask,
	}
	err := base_service.Update(tx, "asset", assetId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "asset", assetId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func BulkToggleIsTask(tx *sqlx.Tx, assetIds []string, isTask bool) error {
	if len(assetIds) == 0 {
		return nil
	}

	assetIdsJSON, err := json.Marshal(assetIds)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`UPDATE asset
		SET is_resource = ?, mtime = ?
		WHERE id IN (SELECT value FROM json_each(?)) AND trashed = 0`,
		!isTask,
		utils.GetEpochTime(),
		string(assetIdsJSON),
	)
	return err
}

func UpdateAsset(tx *sqlx.Tx, assetId string, name, assetTypeId string, isResource bool, pointer string, tags []string) (models.Asset, error) {
	name = strings.TrimSpace(name)
	oldAsset, err := GetAsset(tx, assetId)
	if err != nil {
		return models.Asset{}, err
	}

	newAssetName := oldAsset.Name

	if (name != "") && (name != oldAsset.Name) {
		newAssetName = name
	}

	params := map[string]interface{}{
		"name":          newAssetName,
		"pointer":       pointer,
		"is_resource":   isResource,
		"asset_type_id": assetTypeId,
	}
	err = base_service.Update(tx, "asset", assetId, params)
	if err != nil {
		return models.Asset{}, err
	}
	err = base_service.UpdateMtime(tx, "asset", assetId, utils.GetEpochTime())
	if err != nil {
		return models.Asset{}, err
	}
	asset, err := GetAsset(tx, assetId)
	if err != nil {
		return models.Asset{}, err
	}

	// Reconcile tags by diffing current against desired instead of wiping all rows.
	// This keeps stable asset_tag ids and avoids (asset_id, tag_id) sync collisions.
	existingTags, err := GetAssetTags(tx, assetId)
	if err != nil {
		return models.Asset{}, err
	}

	desiredTags := make(map[string]bool, len(tags))
	for _, tag := range tags {
		desiredTags[tag] = true
	}

	existingTagNames := make(map[string]bool, len(existingTags))
	for _, existingTag := range existingTags {
		existingTagNames[existingTag.Name] = true
		if !desiredTags[existingTag.Name] {
			err = RemoveTagFromAsset(tx, assetId, existingTag.Id)
			if err != nil {
				return models.Asset{}, err
			}
		}
	}

	asset.Tags = []string{}
	for _, tag := range tags {
		if !existingTagNames[tag] {
			err = AddTagToAsset(tx, asset.Id, tag)
			if err != nil {
				return models.Asset{}, err
			}
			existingTagNames[tag] = true
		}
		asset.Tags = append(asset.Tags, tag)
	}

	if oldAsset.FilePath != asset.FilePath && utils.FileExists(oldAsset.FilePath) {
		newFilePath := asset.FilePath

		newAssetFolder := filepath.Dir(newFilePath)

		err := os.MkdirAll(newAssetFolder, os.ModePerm)
		if err != nil {
			return models.Asset{}, err
		}

		err = utils.RenamePathCaseSafe(oldAsset.FilePath, asset.FilePath)
		if err != nil {
			return models.Asset{}, err
		}
	}

	return asset, nil
}

func UpdateSyncAsset(tx *sqlx.Tx, assetId string, name, collectionId, assetTypeId, assigneeId, assignerId, statusId, previewId string, isResource, isLink bool, pointer string, tags []string) error {
	name = strings.TrimSpace(name)
	oldAsset, err := GetAsset(tx, assetId)
	if err != nil {
		return err
	}

	newAssetName := oldAsset.Name

	if (name != "") && (name != oldAsset.Name) {
		newAssetName = name
	}

	params := map[string]any{
		"name":          newAssetName,
		"is_resource":   isResource,
		"is_link":       isLink,
		"pointer":       pointer,
		"asset_type_id": assetTypeId,
		"assignee_id":   assigneeId,
		"assigner_id":   assignerId,
		"collection_id": collectionId,
		"status_id":     statusId,
		"preview_id":    previewId,
	}
	err = base_service.Update(tx, "asset", assetId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "asset", assetId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	asset, err := GetSimpleAsset(tx, assetId)
	if err != nil {
		return err
	}

	err = RemoveAllTagsFromAsset(tx, assetId)
	if err != nil {
		return err
	}
	if asset.IsLink {
		if !utils.NonCaseSensitiveContains(tags, "link") {
			tags = append(tags, "link")
		}
	}
	if !asset.IsLink && asset.Pointer != "" {
		if !utils.NonCaseSensitiveContains(tags, "tracked") {
			tags = append(tags, "tracked")
		}
	}

	if len(tags) > 0 {
		for _, tag := range tags {
			err = AddTagToAsset(tx, asset.Id, tag)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func RenameAsset(tx *sqlx.Tx, assetId, name string) (models.Asset, error) {
	name = strings.TrimSpace(name)
	oldAsset, err := GetAsset(tx, assetId)
	if err != nil {
		return models.Asset{}, err
	}

	newAssetName := oldAsset.Name

	if (name != "") && (name != oldAsset.Name) {
		newAssetName = name
	}

	err = base_service.Rename(tx, "asset", assetId, newAssetName)
	if err != nil {
		return models.Asset{}, err
	}
	err = base_service.UpdateMtime(tx, "asset", assetId, utils.GetEpochTime())
	if err != nil {
		return models.Asset{}, err
	}
	asset, err := GetAsset(tx, assetId)
	if err != nil {
		return models.Asset{}, err
	}

	if oldAsset.FilePath != asset.FilePath && utils.FileExists(oldAsset.FilePath) {
		newFilePath := asset.FilePath

		newAssetFolder := filepath.Dir(newFilePath)

		err := os.MkdirAll(newAssetFolder, os.ModePerm)
		if err != nil {
			return models.Asset{}, err
		}

		err = utils.RenamePathCaseSafe(oldAsset.FilePath, asset.FilePath)
		if err != nil {
			return models.Asset{}, err
		}
	}

	return asset, nil
}

func ToggleIsResource(tx *sqlx.Tx, assetId string, isResource bool) error {
	params := map[string]interface{}{
		"is_resource": isResource,
	}
	err := base_service.Update(tx, "asset", assetId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "asset", assetId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func ToggleIsResourceM(tx *sqlx.Tx, assetIds []string, isResource bool) error {
	for _, assetId := range assetIds {
		err := ToggleIsResource(tx, assetId, isResource)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateAssignation(tx *sqlx.Tx, assetId string, assigneeId, assignerId string) error {
	params := map[string]interface{}{
		"assignee_id": assigneeId,
		"assigner_id": assignerId,
	}
	err := base_service.Update(tx, "asset", assetId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "asset", assetId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func AssignAsset(tx *sqlx.Tx, assetId string, userId string) error {
	params := map[string]interface{}{
		"assignee_id": userId,
	}
	err := base_service.Update(tx, "asset", assetId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "asset", assetId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func UnAssignAsset(tx *sqlx.Tx, assetId string) error {
	params := map[string]interface{}{
		"assignee_id": "",
	}
	err := base_service.Update(tx, "asset", assetId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "asset", assetId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func UnAssignAssets(tx *sqlx.Tx, assetIds []string) error {
	if len(assetIds) == 0 {
		return nil
	}

	currentTime := utils.GetEpochTime()
	params := map[string]interface{}{
		"assignee_id": "",
	}

	for _, assetId := range assetIds {
		err := base_service.Update(tx, "asset", assetId, params)
		if err != nil {
			return err
		}
		err = base_service.UpdateMtime(tx, "asset", assetId, currentTime)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateAssetPointer(tx *sqlx.Tx, assetId string, pointer string) (models.Asset, error) {
	if !utils.FileExists(pointer) {
		return models.Asset{}, errors.New("invalid pointer, path does not exist")
	}
	params := map[string]interface{}{
		"pointer": pointer,
	}
	err := base_service.Update(tx, "asset", assetId, params)
	if err != nil {
		return models.Asset{}, err
	}
	err = base_service.UpdateMtime(tx, "asset", assetId, utils.GetEpochTime())
	if err != nil {
		return models.Asset{}, err
	}
	asset, err := GetAsset(tx, assetId)
	if err != nil {
		return models.Asset{}, err
	}

	return asset, nil
}

// GetAssetAssets gets all assets where is_resource is false with minimal fields for UI display
func GetAssetAssets(tx *sqlx.Tx) ([]models.Asset, error) {
	assets := []models.Asset{}
	queryWhereClause := "WHERE is_resource = 0 AND trashed = 0"

	query := fmt.Sprintf(`
		SELECT 
			id,
			name,
			asset_type_icon,
			assignee_id,
			preview,
			status_id,
			is_resource,
			asset_type_id,
			extension,
			tags
		FROM full_asset %s ORDER BY name`, queryWhereClause)

	err := tx.Select(&assets, query)
	if err != nil {
		return assets, err
	}
	for i := range assets {
		if assets[i].TagsRaw != "" && assets[i].TagsRaw != "[]" {
			assetTags := []AssetTags{}
			if err := json.Unmarshal([]byte(assets[i].TagsRaw), &assetTags); err != nil {
				return assets, err
			}
			for _, t := range assetTags {
				assets[i].Tags = append(assets[i].Tags, t.Name)
			}
		} else {
			assets[i].Tags = []string{}
		}
	}
	return assets, nil
}

// GetCollectionDescendantAssets returns assets located anywhere
// under the given collection's subtree, with the same minimal fields as GetAssetAssets.
func GetCollectionDescendantAssets(tx *sqlx.Tx, collectionId string, includeResources bool) ([]models.Asset, error) {
	assets := []models.Asset{}

	var collectionPath string
	err := tx.Get(&collectionPath, "SELECT collection_path FROM collection_hierarchy WHERE id = ?", collectionId)
	if err != nil {
		return assets, err
	}

	resourceClause := "AND is_resource = 0"
	if includeResources {
		resourceClause = ""
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			name,
			asset_type_icon,
			assignee_id,
			preview,
			status_id,
			is_resource,
			asset_type_id,
			asset_path,
			extension,
			tags
		FROM full_asset
		WHERE trashed = 0 AND collection_path LIKE ? %s
		ORDER BY name`, resourceClause)

	err = tx.Select(&assets, query, collectionPath+"%")
	if err != nil {
		return assets, err
	}

	statuses, err := GetStatuses(tx)
	if err != nil {
		return assets, err
	}
	statusesMap := map[string]models.Status{}
	for _, status := range statuses {
		statusesMap[status.Id] = status
	}

	for i := range assets {
		status := statusesMap[assets[i].StatusId]
		assets[i].Status = status
		assets[i].StatusShortName = status.ShortName

		if assets[i].TagsRaw != "" && assets[i].TagsRaw != "[]" {
			assetTags := []AssetTags{}
			if err := json.Unmarshal([]byte(assets[i].TagsRaw), &assetTags); err != nil {
				return assets, err
			}
			for _, t := range assetTags {
				assets[i].Tags = append(assets[i].Tags, t.Name)
			}
		} else {
			assets[i].Tags = []string{}
		}
	}
	return assets, nil
}
