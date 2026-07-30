package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"clustta/internal/base_service"
	"clustta/internal/error_service"
	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

func TopologicalSortOld(collections []models.Collection) ([]models.Collection, error) {

	// Create maps for easy lookup
	idToCollection := make(map[string]models.Collection)
	fmt.Println("idToCollection: ", idToCollection)
	dependencyCount := make(map[string]int)
	fmt.Println("Dependency count: ", dependencyCount)
	children := make(map[string][]string)
	fmt.Println("Children: ", children)

	// Build maps
	for _, e := range collections {
		idToCollection[e.Id] = e
		if e.ParentId != "" {
			dependencyCount[e.Id]++
			children[e.ParentId] = append(children[e.ParentId], e.Id)
		}
	}

	// Start with nodes that have no dependencies (root nodes)
	var queue []string
	for _, e := range collections {
		if dependencyCount[e.Id] == 0 {
			queue = append(queue, e.Id)
		}
	}

	var result []models.Collection
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		result = append(result, idToCollection[id])

		for _, childId := range children[id] {
			dependencyCount[childId]--
			if dependencyCount[childId] == 0 {
				queue = append(queue, childId)
			}
		}
	}
	fmt.Println("Collections: ", collections)
	fmt.Println("Result: ", result)

	if len(result) != len(collections) {
		fmt.Println("cycle detected or missing parents: ", len(result)-len(collections))
		// return nil, fmt.Errorf("cycle detected or missing parents")
	}

	return result, nil
}

func TopologicalSort(collections []models.Collection) ([]models.Collection, error) {
	idToCollection := make(map[string]models.Collection)
	dependencyCount := make(map[string]int)
	children := make(map[string][]string)

	for _, e := range collections {
		idToCollection[e.Id] = e
		if e.ParentId != "" {
			dependencyCount[e.Id]++
			children[e.ParentId] = append(children[e.ParentId], e.Id)
		}
	}

	collectionIds := make(map[string]struct{})
	for _, e := range collections {
		collectionIds[e.Id] = struct{}{}
	}

	var queue []string
	for _, e := range collections {
		_, parentExists := collectionIds[e.ParentId]
		if dependencyCount[e.Id] == 0 || (e.ParentId != "" && !parentExists) {
			queue = append(queue, e.Id)
		}
	}

	processed := make(map[string]bool)
	var result []models.Collection
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		result = append(result, idToCollection[id])
		processed[id] = true

		for _, childId := range children[id] {
			dependencyCount[childId]--
			if dependencyCount[childId] == 0 {
				queue = append(queue, childId)
			}
		}
	}

	for _, e := range collections {
		if !processed[e.Id] {
			result = append(result, e)
		}
	}

	if len(result) != len(collections) {
		fmt.Println("cycle detected or missing parents: ", len(result)-len(collections))
		// return nil, fmt.Errorf("cycle detected or missing parents")
	}

	return result, nil
}

func CreateCollectionFast(
	tx *sqlx.Tx, id string, name, description, collection_type_id, parent_id, previewId string, isShared bool,
) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("collection name cannot be empty")
	}

	// if parent_id != "" {
	// 	parent := models.Collection{}
	// 	err := base_service.Get(tx, "collection", parent_id, &parent)
	// 	if err != nil {
	// 		return models.Collection{}, errors.New("parent collection not found")
	// 	}
	// }

	params := map[string]any{
		"id":                 id,
		"created_at":         utils.GetCurrentTime(),
		"name":               name,
		"description":        description,
		"collection_type_id": collection_type_id,
		"parent_id":          parent_id,
		"preview_id":         previewId,
		"is_shared":          isShared,
	}
	err := base_service.Create(tx, "collection", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return error_service.ErrCollectionExists
			}
		}
		return err
	}
	return nil
}

func CreateCollection(
	tx *sqlx.Tx, id string, name, description, collection_type_id, parent_id, previewId string, isShared bool,
) (models.Collection, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.Collection{}, errors.New("collection name cannot be empty")
	}

	if parent_id != "" {
		parent := models.Collection{}
		err := base_service.Get(tx, "collection", parent_id, &parent)
		if err != nil {
			return models.Collection{}, errors.New("parent collection not found")
		}
	}

	conditions := map[string]any{
		"parent_id": parent_id,
		"name":      name,
	}
	collection := models.Collection{}
	err := base_service.GetBy(tx, "collection", conditions, &collection)
	if err == nil {
		if collection.Trashed {
			return models.Collection{}, error_service.ErrCollectionExistsInTrash
		} else {
			return models.Collection{}, error_service.ErrCollectionExists
		}
	}

	params := map[string]any{
		"id":                 id,
		"created_at":         utils.GetCurrentTime(),
		"name":               name,
		"description":        description,
		"collection_type_id": collection_type_id,
		"parent_id":          parent_id,
		"preview_id":         previewId,
		"is_shared":          isShared,
	}
	err = base_service.Create(tx, "collection", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return models.Collection{}, error_service.ErrCollectionExists
			}
		}
		return models.Collection{}, err
	}
	collection, err = GetCollectionByName(tx, name, parent_id)
	if err != nil {
		return models.Collection{}, err
	}
	return collection, nil
}

func AddCollection(
	tx *sqlx.Tx, id string, name, description, collection_type_id, parent_id, previewId string, isShared bool,
) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("collection name cannot be empty")
	}

	// if parent_id != "" {
	// 	parent := models.Collection{}
	// 	err := base_service.Get(tx, "collection", parent_id, &parent)
	// 	if err != nil {
	// 		return models.Collection{}, errors.New("parent collection not found")
	// 	}
	// }

	params := map[string]any{
		"id":                 id,
		"created_at":         utils.GetCurrentTime(),
		"name":               name,
		"description":        description,
		"collection_type_id": collection_type_id,
		"parent_id":          parent_id,
		"preview_id":         previewId,
		"is_shared":          isShared,
	}
	err := base_service.Create(tx, "collection", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return error_service.ErrCollectionExists
			}
		}
		return err
	}
	return nil
}

func GetSimpleCollections(tx *sqlx.Tx) ([]models.Collection, error) {
	collections := []models.Collection{}

	query := "SELECT * FROM collection"

	err := tx.Select(&collections, query)
	if err != nil && err == sql.ErrNoRows {
		return []models.Collection{}, nil
	} else if err != nil {
		return []models.Collection{}, err
	}
	return collections, nil
}

// GetSimpleCollection returns a single collection row without JOINs.
func GetSimpleCollection(tx *sqlx.Tx, id string) (models.Collection, error) {
	collection := models.Collection{}
	err := tx.Get(&collection, "SELECT * FROM collection WHERE id = ?", id)
	if err != nil {
		return models.Collection{}, err
	}
	return collection, nil
}

func GetCollection(tx *sqlx.Tx, id string) (models.Collection, error) {
	collection := models.Collection{}

	query := "SELECT * FROM full_collection WHERE id = ?"

	err := tx.Get(&collection, query, id)
	if err != nil && err == sql.ErrNoRows {
		return models.Collection{}, error_service.ErrCollectionNotFound
	} else if err != nil {
		return models.Collection{}, err
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return collection, err
	}
	collectionFilePath, err := utils.BuildCollectionPath(rootFolder, collection.CollectionPath)
	if err != nil {
		return collection, err
	}
	collection.FilePath = collectionFilePath

	if collection.AssigneeIdsRaw != "[]" {
		assigneeIds := []string{}
		err = json.Unmarshal([]byte(collection.AssigneeIdsRaw), &assigneeIds)
		if err != nil {
			return collection, err
		}
		collection.AssigneeIds = assigneeIds
	} else {
		collection.AssigneeIds = []string{} // Ensure it's initialized as an empty slice
	}
	return collection, nil
}

func GetCollectionChildren(tx *sqlx.Tx, id string) ([]models.Collection, error) {
	collections := []models.Collection{}

	query := "SELECT * FROM full_collection WHERE parent_id = ? AND trashed = 0 ORDER BY name"

	err := tx.Select(&collections, query, id)
	if err != nil && err == sql.ErrNoRows {
		return collections, nil
	} else if err != nil {
		return []models.Collection{}, err
	}
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return collections, err
	}
	for i, collection := range collections {
		collectionFilePath, err := utils.BuildCollectionPath(rootFolder, collection.CollectionPath)
		if err != nil {
			return collections, err
		}
		collections[i].FilePath = collectionFilePath

		if collection.AssigneeIdsRaw != "[]" {
			assigneeIds := []string{}
			err = json.Unmarshal([]byte(collection.AssigneeIdsRaw), &assigneeIds)
			if err != nil {
				return collections, err
			}
			collections[i].AssigneeIds = assigneeIds
		} else {
			collections[i].AssigneeIds = []string{} // Ensure it's initialized as an empty slice
		}
	}
	return collections, nil
}

func GetCollectionAssets(tx *sqlx.Tx, id string) ([]models.Asset, error) {
	assets := []models.Asset{}

	query := "SELECT * FROM full_asset WHERE collection_id = ? AND trashed = 0 ORDER BY name"

	err := tx.Select(&assets, query, id)
	if err != nil && err == sql.ErrNoRows {
		return assets, nil
	} else if err != nil {
		return []models.Asset{}, err
	}

	statuses, err := GetStatuses(tx)
	if err != nil {
		return assets, err
	}
	statusesMap := map[string]models.Status{}
	for _, status := range statuses {
		statusesMap[status.Id] = status
	}

	qoutedAssetIds := make([]string, len(assets))
	for i, asset := range assets {
		qoutedAssetIds[i] = fmt.Sprintf("\"%s\"", asset.Id)
	}
	// assetsCheckpointQuery := fmt.Sprintf("SELECT * FROM asset_checkpoint WHERE asset_id IN (%s) AND trashed = 0 ORDER BY created_at DESC", strings.Join(qoutedAssetIds, ","))

	// // checkpointQuery := "SELECT * FROM asset_checkpoint WHERE trashed = 0 ORDER BY created_at DESC"
	// assetsCheckpoints := []models.Checkpoint{}
	// tx.Select(&assetsCheckpoints, assetsCheckpointQuery)

	// assetCheckpoints := map[string][]models.Checkpoint{}
	// for _, assetCheckpoint := range assetsCheckpoints {
	// 	assetCheckpoints[assetCheckpoint.AssetId] = append(assetCheckpoints[assetCheckpoint.AssetId], assetCheckpoint)
	// }

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
		// assets[i].Checkpoints = assetCheckpoints[asset.Id]
		// fileStatus, err := GetAssetFileStatus(&assets[i], assetCheckpoints[asset.Id])
		// if err != nil {
		// 	return assets, err
		// }
		assets[i].FileStatus = "fetchable"
	}

	return assets, nil
}

func GetCollectionByPath(tx *sqlx.Tx, collectionPath string) (models.Collection, error) {
	collection := models.Collection{}

	query := "SELECT * FROM full_collection WHERE collection_path = ?"

	err := tx.Get(&collection, query, collectionPath)
	if err != nil && err == sql.ErrNoRows {
		return models.Collection{}, error_service.ErrCollectionNotFound
	} else if err != nil {
		return models.Collection{}, err
	}
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return collection, err
	}
	collectionFilePath, err := utils.BuildCollectionPath(rootFolder, collection.CollectionPath)
	if err != nil {
		return collection, err
	}
	collection.FilePath = collectionFilePath
	if collection.AssigneeIdsRaw != "[]" {
		assigneeIds := []string{}
		err = json.Unmarshal([]byte(collection.AssigneeIdsRaw), &assigneeIds)
		if err != nil {
			return collection, err
		}
		collection.AssigneeIds = assigneeIds
	} else {
		collection.AssigneeIds = []string{} // Ensure it's initialized as an empty slice
	}
	return collection, nil
}

func GetOrCreateCollection(tx *sqlx.Tx, collectionPath string) (models.Collection, []models.Collection, error) {
	parts := strings.Split(collectionPath, "/")
	var collectionPaths []string
	var current string

	if collectionPath == "" {
		return models.Collection{}, []models.Collection{}, nil
	}

	for _, part := range parts {
		if current == "" {
			current = part
		} else {
			current = path.Join(current, part)
		}
		collectionPaths = append(collectionPaths, current)
	}

	collectionType, err := GetCollectionTypeByName(tx, "generic")
	if err != nil {
		return models.Collection{}, []models.Collection{}, err
	}

	newCollections := []models.Collection{}
	prevCollection := models.Collection{}
	for _, curentCollectionPath := range collectionPaths {
		collection, err := GetCollectionByPath(tx, curentCollectionPath)
		if err == nil {
			prevCollection = collection
			continue
		}
		collectionName := filepath.Base(curentCollectionPath)
		collection, err = CreateCollection(tx, "", collectionName, "", collectionType.Id, prevCollection.Id, "", false)
		if err != nil {
			return collection, newCollections, err
		}
		newCollections = append(newCollections, collection)
		prevCollection = collection
	}

	return prevCollection, newCollections, nil
}

func GetCollectionByName(tx *sqlx.Tx, name string, parentId string) (models.Collection, error) {
	collection := models.Collection{}
	query := "SELECT * FROM full_collection WHERE name = ? AND parent_id = ?"

	err := tx.Get(&collection, query, name, parentId)
	if err != nil && err == sql.ErrNoRows {
		return models.Collection{}, error_service.ErrCollectionNotFound
	} else if err != nil {
		return models.Collection{}, err
	}
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return collection, err
	}
	collectionFilePath, err := utils.BuildCollectionPath(rootFolder, collection.CollectionPath)
	if err != nil {
		return collection, err
	}
	collection.FilePath = collectionFilePath
	if collection.AssigneeIdsRaw != "[]" {
		assigneeIds := []string{}
		err = json.Unmarshal([]byte(collection.AssigneeIdsRaw), &assigneeIds)
		if err != nil {
			return collection, err
		}
		collection.AssigneeIds = assigneeIds
	} else {
		collection.AssigneeIds = []string{} // Ensure it's initialized as an empty slice
	}
	return collection, nil
}

func GetCollections(tx *sqlx.Tx, withDeleted bool) ([]models.Collection, error) {
	collections := []models.Collection{}
	queryWhereClause := ""
	if !withDeleted {
		queryWhereClause = "WHERE trashed = 0"
	}
	query := fmt.Sprintf("SELECT * FROM full_collection %s", queryWhereClause)

	err := tx.Select(&collections, query)
	if err != nil {
		return collections, err
	}
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return collections, err
	}
	for i, collection := range collections {
		collectionFilePath, err := utils.BuildCollectionPath(rootFolder, collection.CollectionPath)
		if err != nil {
			return collections, err
		}
		collections[i].FilePath = collectionFilePath

		if collection.AssigneeIdsRaw != "[]" {
			assigneeIds := []string{}
			err = json.Unmarshal([]byte(collection.AssigneeIdsRaw), &assigneeIds)
			if err != nil {
				return collections, err
			}
			collections[i].AssigneeIds = assigneeIds
		} else {
			collections[i].AssigneeIds = []string{} // Ensure it's initialized as an empty slice
		}
	}

	return collections, nil
}

func OLDGetUserCollections(tx *sqlx.Tx, userId string) ([]models.Collection, error) {
	collections := []models.Collection{}
	// query := "SELECT * FROM full_collection WHERE trashed = 0"
	query := fmt.Sprintf(`
		WITH RECURSIVE 
		-- First get all assets and their dependencies
		asset_dependencies AS (
			-- Base case: directly assigned assets
			SELECT 
				id,
				collection_id
			FROM asset
			WHERE assignee_id = '%s' 
			AND trashed = 0
			
			UNION
			
			-- Recursive case: asset dependencies
			SELECT 
				t.id,
				t.collection_id
			FROM asset t
			JOIN asset_dependency dep ON t.id = dep.dependency_id
			JOIN asset_dependencies td ON dep.asset_id = td.id
			WHERE t.trashed = 0
		),
		
		-- Get collection dependencies for all assets
		collection_dependencies AS (
			-- Base case: direct collection dependencies from assets
			SELECT 
				ed.dependency_id as id
			FROM collection_dependency ed
			JOIN asset_dependencies td ON ed.asset_id = td.id
			WHERE ed.dependency_id != ''
			
			UNION
			
			-- Include original asset collections
			SELECT DISTINCT collection_id as id
			FROM asset_dependencies
			WHERE collection_id != ''
		),
		
		-- Get full collection hierarchy (both parents and children) for all relevant collections
		collection_hierarchy_full AS (
			-- Base case: collections from assets and collection dependencies
			SELECT 
				e.id,
				e.parent_id,
				e.name,
				0 as level,
				CAST(e.id AS TEXT) as hierarchy_path
			FROM collection e
			LEFT JOIN collection_dependencies ed ON e.id = ed.id
			WHERE e.trashed = 0
			AND (ed.id IS NOT NULL OR e.is_shared = 1)
			
			UNION ALL
			
			-- Recursive case upward: get parents
			SELECT 
				e.id,
				e.parent_id,
				e.name,
				ehf.level - 1,
				ehf.hierarchy_path || ',' || e.id
			FROM collection e
			JOIN collection_hierarchy_full ehf ON e.id = ehf.parent_id
			WHERE e.trashed = 0
			AND e.id != ''
			AND e.id NOT IN (
				SELECT value 
				FROM json_each('["' || REPLACE(ehf.hierarchy_path, ',', '","') || '"]')
			)
			
			UNION ALL
			
			-- Recursive case downward: get children
			SELECT 
				e.id,
				e.parent_id,
				e.name,
				ehf.level + 1,
				ehf.hierarchy_path || ',' || e.id
			FROM collection e
			JOIN collection_hierarchy_full ehf ON e.parent_id = ehf.id
			WHERE e.trashed = 0
			AND e.is_shared = 1
			AND e.id NOT IN (
				SELECT value 
				FROM json_each('["' || REPLACE(ehf.hierarchy_path, ',', '","') || '"]')
			)
		)
		-- Final select with collection details
		SELECT DISTINCT
			e.*,
			et.name AS collection_type_name,
			et.icon AS collection_type_icon,
			p.preview AS preview,
			COALESCE(eh.collection_path, '') AS collection_path, -- Ensure no NULL values
			CASE 
				WHEN ed.id IS NOT NULL THEN true 
				ELSE false 
			END as is_dependency
		FROM collection_hierarchy_full ehf
		JOIN collection e ON ehf.id = e.id
		JOIN collection_type et ON e.collection_type_id = et.id
		LEFT JOIN preview p ON e.preview_id = p.hash
		LEFT JOIN collection_hierarchy eh ON e.id = eh.id
		LEFT JOIN collection_dependencies ed ON e.id = ed.id
		ORDER BY 
			eh.collection_path;
	`, userId)

	err := tx.Select(&collections, query)
	if err != nil {
		return collections, err
	}
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return collections, err
	}
	for i, collection := range collections {

		collectionFilePath, err := utils.BuildCollectionPath(rootFolder, collection.CollectionPath)
		if err != nil {
			return collections, err
		}
		collections[i].FilePath = collectionFilePath
	}

	return collections, nil
}

func getAllCollectionChildren(collectionId string, allCollections []models.Collection) []string {
	children := []string{}
	for _, collection := range allCollections {
		if collection.ParentId == collectionId {
			children = append(children, collection.Id)
			// Recursively get children of the current child
			children = append(children, getAllCollectionChildren(collection.Id, allCollections)...)
		}
	}
	return children
}

func getCollectionParents(collectionId string, allCollections []models.Collection) []string {
	parents := []string{}
	for _, collection := range allCollections {
		if collection.Id == collectionId && collection.ParentId != "" {
			parents = append(parents, collection.ParentId)
			// Recursively get parents of the current parent
			parents = append(parents, getCollectionParents(collection.ParentId, allCollections)...)
		}
	}
	return parents
}

func getAssetCollections(asset models.Asset, allCollections []models.Collection) []string {
	parents := []string{}
	parents = append(parents, asset.CollectionId)
	// Recursively get parent collections
	parents = append(parents, getCollectionParents(asset.CollectionId, allCollections)...)
	return parents
}

func GetUserCollections(tx *sqlx.Tx, userAssetInfos []models.Asset, userId string) ([]models.Collection, error) {
	// Get all assigned collections for the user
	assignedCollectionsIds := []string{}
	query := "select collection_id from collection_assignee where assignee_id = ?"
	err := tx.Select(&assignedCollectionsIds, query, userId)
	if err != nil {
		return nil, err
	}

	sharedCollections := []string{}
	query = "select id from collection where is_shared = 1"
	err = tx.Select(&sharedCollections, query, userId)
	if err != nil {
		return nil, err
	}

	allCollectionInfo := []models.Collection{}
	query = `SELECT id, parent_id FROM collection WHERE trashed = 0`
	err = tx.Select(&allCollectionInfo, query)
	if err != nil {
		return nil, err
	}

	userAssetCollectionsIds := []string{}
	for _, userAsset := range userAssetInfos {
		userAssetCollectionsIds = append(userAssetCollectionsIds, getAssetCollections(userAsset, allCollectionInfo)...)
	}

	//process all user collections
	canModifyCollectionsIds := map[string]struct{}{}
	userCollectionsIds := map[string]struct{}{}
	for _, collectionId := range assignedCollectionsIds {
		userCollectionsIds[collectionId] = struct{}{}
		canModifyCollectionsIds[collectionId] = struct{}{}
		for _, parent := range getCollectionParents(collectionId, allCollectionInfo) {
			userCollectionsIds[parent] = struct{}{}
		}
		for _, child := range getAllCollectionChildren(collectionId, allCollectionInfo) {
			userCollectionsIds[child] = struct{}{}
			canModifyCollectionsIds[child] = struct{}{}
		}
	}
	for _, collectionId := range sharedCollections {
		userCollectionsIds[collectionId] = struct{}{}
		for _, parent := range getCollectionParents(collectionId, allCollectionInfo) {
			userCollectionsIds[parent] = struct{}{}
		}
		for _, child := range getAllCollectionChildren(collectionId, allCollectionInfo) {
			userCollectionsIds[child] = struct{}{}
		}
	}
	for _, collectionId := range userAssetCollectionsIds {
		userCollectionsIds[collectionId] = struct{}{}
	}

	collectionsIds := make([]string, 0, len(userCollectionsIds))
	for id := range userCollectionsIds {
		collectionsIds = append(collectionsIds, id)
	}

	// Get all collections with the collected IDs
	query = `
		SELECT * FROM full_collection WHERE id IN (SELECT value FROM json_each(?)) AND trashed = 0
	`
	collections := []models.Collection{}
	jsonCollectionIds, err := json.Marshal(collectionsIds)
	if err != nil {
		return nil, err
	}
	err = tx.Select(&collections, query, jsonCollectionIds)
	if err != nil {
		return nil, err
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return collections, err
	}
	for i, collection := range collections {
		collectionFilePath, err := utils.BuildCollectionPath(rootFolder, collection.CollectionPath)
		if err != nil {
			return collections, err
		}
		collections[i].FilePath = collectionFilePath
		if collection.AssigneeIdsRaw != "[]" {
			assigneeIds := []string{}
			err = json.Unmarshal([]byte(collection.AssigneeIdsRaw), &assigneeIds)
			if err != nil {
				return collections, err
			}
			collections[i].AssigneeIds = assigneeIds
		} else {
			collections[i].AssigneeIds = []string{} // Ensure it's initialized as an empty slice
		}

		// Add CanModify to collection if in the canModifyCollectionsIds map
		if _, exists := canModifyCollectionsIds[collection.Id]; exists {
			collections[i].CanModify = true
		} else {
			collections[i].CanModify = false
		}
	}

	return collections, nil
}

func GetDeletedCollections(tx *sqlx.Tx) ([]models.Collection, error) {
	collections := []models.Collection{}

	conditions := map[string]any{
		"trashed": 1,
	}
	err := base_service.GetAllBy(tx, "full_collection", conditions, &collections)
	if err != nil {
		return collections, err
	}
	// tx.Select(&collections, "SELECT * FROM full_collection WHERE trashed = 0")
	newCollections := []models.Collection{}
	//TODO Investigate why for loop does not update data
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return collections, err
	}
	for _, collection := range collections {
		collectionFilePath, err := utils.BuildCollectionPath(rootFolder, collection.CollectionPath)
		if err != nil {
			return collections, err
		}
		collection.FilePath = collectionFilePath
		newCollections = append(newCollections, collection)
	}

	return newCollections, nil
}

func DeleteCollection(tx *sqlx.Tx, collectionId string, removeFromDir bool, recycle bool) error {
	collection, err := GetCollection(tx, collectionId)
	if err != nil {
		return err
	}
	if recycle {
		err = base_service.MarkAsDeleted(tx, "collection", collectionId)
		if err != nil {
			return err
		}
	} else {
		err = base_service.Delete(tx, "collection", collectionId)
		if err != nil {
			return err
		}
	}
	if removeFromDir {
		err := os.RemoveAll(collection.FilePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateCollection(tx *sqlx.Tx, collectionId string, name string, tags []string) (models.Collection, error) {
	name = strings.TrimSpace(name)
	oldCollection, err := GetCollection(tx, collectionId)
	if err != nil {
		return models.Collection{}, err
	}

	newCollectionName := oldCollection.Name

	if (name != "") && (name != oldCollection.Name) {
		newCollectionName = name
	}

	params := map[string]any{
		"name": newCollectionName,
	}
	err = base_service.Update(tx, "collection", collectionId, params)
	if err != nil {
		return models.Collection{}, err
	}
	err = base_service.UpdateMtime(tx, "collection", collectionId, utils.GetEpochTime())
	if err != nil {
		return models.Collection{}, err
	}
	collection, err := GetCollection(tx, collectionId)
	if err != nil {
		return models.Collection{}, err
	}

	if oldCollection.FilePath != collection.FilePath && utils.DirExists(oldCollection.FilePath) {
		newCollectionFolderDir := filepath.Dir(collection.FilePath)
		err := os.MkdirAll(newCollectionFolderDir, os.ModePerm)
		if err != nil {
			return models.Collection{}, err
		}

		err = utils.RenamePathCaseSafe(oldCollection.FilePath, collection.FilePath)
		if err != nil {
			return models.Collection{}, err
		}
	}

	return collection, nil
}

func RenameCollection(tx *sqlx.Tx, collectionId string, name string) (models.Collection, error) {
	name = strings.TrimSpace(name)
	oldCollection, err := GetCollection(tx, collectionId)
	if err != nil {
		return models.Collection{}, err
	}
	newCollectionName := oldCollection.Name

	if (name != "") && (name != oldCollection.Name) {
		newCollectionName = name
	}

	err = base_service.Rename(tx, "collection", collectionId, newCollectionName)
	if err != nil {
		return models.Collection{}, err
	}
	err = base_service.UpdateMtime(tx, "collection", collectionId, utils.GetEpochTime())
	if err != nil {
		return models.Collection{}, err
	}
	collection, err := GetCollection(tx, collectionId)
	if err != nil {
		return models.Collection{}, err
	}

	if oldCollection.FilePath != collection.FilePath && utils.DirExists(oldCollection.FilePath) {

		newCollectionFolderDir := filepath.Dir(collection.FilePath)

		err := os.MkdirAll(newCollectionFolderDir, os.ModePerm)
		if err != nil {
			return models.Collection{}, err
		}

		err = utils.RenamePathCaseSafe(oldCollection.FilePath, collection.FilePath)
		if err != nil {
			return models.Collection{}, err
		}
	}
	return collection, nil
}

func ChangeParent(tx *sqlx.Tx, collectionId string, parentId string) error {
	oldCollection, err := GetCollection(tx, collectionId)
	if err != nil {
		return err
	}
	params := map[string]any{
		"parent_id": parentId,
	}
	err = base_service.Update(tx, "collection", collectionId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "collection", collectionId, utils.GetEpochTime())
	if err != nil {
		return err
	}

	collection, err := GetCollection(tx, collectionId)
	if err != nil {
		return err
	}
	if oldCollection.FilePath != collection.FilePath && utils.DirExists(oldCollection.FilePath) {
		newCollectionFolderDir := filepath.Dir(collection.FilePath)
		err := os.MkdirAll(newCollectionFolderDir, os.ModePerm)
		if err != nil {
			return err
		}
		err = utils.RenamePathCaseSafe(oldCollection.FilePath, collection.FilePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func ChangeCollectionType(tx *sqlx.Tx, collectionId string, collectionTypeId string) error {
	params := map[string]any{
		"collection_type_id": collectionTypeId,
	}
	err := base_service.Update(tx, "collection", collectionId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "collection", collectionId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func ChangeIsShared(tx *sqlx.Tx, collectionId string, isShared bool) error {
	params := map[string]any{
		"is_shared": isShared,
	}
	err := base_service.Update(tx, "collection", collectionId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "collection", collectionId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func UpdateCollectionPreview(tx *sqlx.Tx, collectionId string, previewPath string) error {
	_, err := GetCollection(tx, collectionId)
	if err != nil {
		return err
	}
	preview, err := CreatePreview(tx, previewPath)
	if err != nil {
		return err
	}
	params := map[string]any{
		"preview_id": preview.Hash,
	}
	err = base_service.Update(tx, "collection", collectionId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "collection", collectionId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func AddAssignee(tx *sqlx.Tx, id, collectionId, userId string) error {
	params := map[string]any{
		"id":            id,
		"collection_id": collectionId,
		"assignee_id":   userId,
	}
	err := base_service.Create(tx, "collection_assignee", params)
	if err != nil {
		return err
	}
	return nil
}

func GetAssignee(tx *sqlx.Tx, id string) (models.CollectionAssignee, error) {
	assignee := models.CollectionAssignee{}
	err := base_service.Get(tx, "collection_assignee", id, &assignee)
	if err != nil {
		return assignee, err
	}
	return assignee, nil
}

func AssignCollection(tx *sqlx.Tx, collectionId, userId string) error {
	params := map[string]any{
		"collection_id": collectionId,
		"assignee_id":   userId,
	}
	err := base_service.Create(tx, "collection_assignee", params)
	if err != nil {
		return err
	}
	return nil
}

func UnAssignCollection(tx *sqlx.Tx, collectionId, userId string) error {
	conditions := map[string]any{
		"collection_id": collectionId,
		"assignee_id":   userId,
	}
	err := base_service.DeleteBy(tx, "collection_assignee", conditions)
	if err != nil {
		return err
	}
	return nil
}

// GetUserCanModifyCollectionIds returns the set of collection ids the user has
// modify access to via direct assignment or recursive descendants.
func GetUserCanModifyCollectionIds(tx *sqlx.Tx, userId string) (map[string]struct{}, error) {
	assigned := []string{}
	if err := tx.Select(&assigned, "SELECT collection_id FROM collection_assignee WHERE assignee_id = ?", userId); err != nil {
		return nil, err
	}
	all := []models.Collection{}
	if err := tx.Select(&all, "SELECT id, parent_id FROM collection WHERE trashed = 0"); err != nil {
		return nil, err
	}
	set := map[string]struct{}{}
	for _, id := range assigned {
		set[id] = struct{}{}
		for _, child := range getAllCollectionChildren(id, all) {
			set[child] = struct{}{}
		}
	}
	return set, nil
}

// IsUserAssignedToCollectionOrAncestor checks if a user is assigned to the given collection
// or any of its parent collections recursively up to the root.
func IsUserAssignedToCollectionOrAncestor(tx *sqlx.Tx, collectionId, userId string) (bool, error) {
	// Check the collection itself first.
	directQuery := `SELECT EXISTS(SELECT 1 FROM collection_assignee WHERE collection_id = ? AND assignee_id = ?)`
	var directMatch bool
	err := tx.Get(&directMatch, directQuery, collectionId, userId)
	if err != nil {
		return false, err
	}
	if directMatch {
		return true, nil
	}

	// Check ancestor collections recursively.
	ancestorQuery := `
		WITH RECURSIVE ancestors AS (
			SELECT parent_id FROM collection WHERE id = ? AND parent_id != ''
			UNION ALL
			SELECT e.parent_id FROM collection e
			JOIN ancestors a ON e.id = a.parent_id
			WHERE a.parent_id != ''
		)
		SELECT EXISTS(
			SELECT 1 FROM collection_assignee ea
			JOIN ancestors a ON ea.collection_id = a.parent_id
			WHERE ea.assignee_id = ?
		)
	`
	var ancestorMatch bool
	err = tx.Get(&ancestorMatch, ancestorQuery, collectionId, userId)
	if err != nil {
		return false, err
	}
	return ancestorMatch, nil
}
