package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
)

const (
	pendingAsset      = "asset"
	pendingCollection = "collection"
)

type pendingPathUpdate struct {
	EntityType       string `db:"entity_type"`
	EntityId         string `db:"entity_id"`
	CurrentLocalPath string `db:"current_local_path"`
}

type completedPathMove struct {
	oldPath string
	newPath string
}

// RecordPendingPathUpdates compares incoming path metadata with current rows.
func RecordPendingPathUpdates(tx *sqlx.Tx, incomingCollections []models.Collection, incomingAssets []models.Asset, requireNewer bool) error {
	workingDirectory, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return err
	}

	collections := []models.Collection{}
	incomingCollectionIds := make([]string, len(incomingCollections))
	for i, collection := range incomingCollections {
		incomingCollectionIds[i] = collection.Id
	}
	encodedCollectionIds, err := json.Marshal(incomingCollectionIds)
	if err != nil {
		return err
	}
	collectionIds := string(encodedCollectionIds)
	if collectionIds != "[]" {
		query := `SELECT id, mtime, name, parent_id, collection_path FROM collection
			WHERE trashed = 0 AND id IN (SELECT value FROM json_each(?))`
		if err := tx.Select(&collections, query, collectionIds); err != nil {
			return err
		}
	}
	collectionIndex := make(map[string]models.Collection, len(collections))
	for _, collection := range collections {
		collectionIndex[collection.Id] = collection
	}

	for _, incoming := range incomingCollections {
		current, exists := collectionIndex[incoming.Id]
		if !exists || (requireNewer && current.MTime >= incoming.MTime) {
			continue
		}
		if current.Name == incoming.Name && current.ParentId == incoming.ParentId {
			continue
		}
		currentPath, err := utils.BuildCollectionPath(workingDirectory, current.CollectionPath)
		if err != nil {
			return err
		}
		if !utils.DirExists(currentPath) {
			continue
		}
		if err := recordCollectionSubtreePaths(tx, workingDirectory, current.Id); err != nil {
			return err
		}
	}

	assets := []models.Asset{}
	incomingAssetIds := make([]string, len(incomingAssets))
	for i, asset := range incomingAssets {
		incomingAssetIds[i] = asset.Id
	}
	encodedAssetIds, err := json.Marshal(incomingAssetIds)
	if err != nil {
		return err
	}
	assetIds := string(encodedAssetIds)
	assetQuery := `SELECT asset.id, asset.mtime, asset.name, asset.extension, asset.collection_id,
		asset.is_link, IFNULL(collection.collection_path, '') AS collection_path
		FROM asset LEFT JOIN collection ON collection.id = asset.collection_id
		WHERE asset.trashed = 0 AND asset.id IN (SELECT value FROM json_each(?))`
	if assetIds != "[]" {
		if err := tx.Select(&assets, assetQuery, assetIds); err != nil {
			return err
		}
	}
	assetIndex := make(map[string]models.Asset, len(assets))
	for _, asset := range assets {
		assetIndex[asset.Id] = asset
	}
	for _, incoming := range incomingAssets {
		current, exists := assetIndex[incoming.Id]
		if !exists || current.IsLink || (requireNewer && current.MTime >= incoming.MTime) {
			continue
		}
		if current.Name == incoming.Name && current.Extension == incoming.Extension && current.CollectionId == incoming.CollectionId {
			continue
		}
		currentPath, err := utils.BuildAssetPath(workingDirectory, current.CollectionPath, current.Name, current.Extension)
		if err != nil {
			return err
		}
		if utils.FileExists(currentPath) {
			if err := preservePendingPath(tx, pendingAsset, current.Id, currentPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func recordCollectionSubtreePaths(tx *sqlx.Tx, workingDirectory, collectionId string) error {
	collections := []models.Collection{}
	collectionQuery := `WITH RECURSIVE subtree(id, collection_path) AS (
		SELECT id, collection_path FROM collection WHERE id = ? AND trashed = 0
		UNION ALL
		SELECT collection.id, collection.collection_path FROM collection
		JOIN subtree ON collection.parent_id = subtree.id WHERE collection.trashed = 0
	) SELECT id, collection_path FROM subtree`
	if err := tx.Select(&collections, collectionQuery, collectionId); err != nil {
		return err
	}
	collectionIds := make([]string, 0, len(collections))
	for _, collection := range collections {
		collectionIds = append(collectionIds, collection.Id)
		path, err := utils.BuildCollectionPath(workingDirectory, collection.CollectionPath)
		if err != nil {
			return err
		}
		if utils.DirExists(path) {
			if err := preservePendingPath(tx, pendingCollection, collection.Id, path); err != nil {
				return err
			}
		}
	}

	encodedCollectionIds, err := json.Marshal(collectionIds)
	if err != nil {
		return err
	}
	assets := []models.Asset{}
	query := `SELECT asset.id, asset.name, asset.extension, collection.collection_path
		FROM asset JOIN collection ON collection.id = asset.collection_id
		WHERE asset.trashed = 0 AND asset.is_link = 0
		AND asset.collection_id IN (SELECT value FROM json_each(?))`
	if err := tx.Select(&assets, query, string(encodedCollectionIds)); err != nil {
		return err
	}
	for _, asset := range assets {
		path, err := utils.BuildAssetPath(workingDirectory, asset.CollectionPath, asset.Name, asset.Extension)
		if err != nil {
			return err
		}
		if utils.FileExists(path) {
			if err := preservePendingPath(tx, pendingAsset, asset.Id, path); err != nil {
				return err
			}
		}
	}
	return nil
}

func preservePendingPath(tx *sqlx.Tx, entityType, entityId, localPath string) error {
	existingPath := ""
	err := tx.Get(&existingPath,
		"SELECT current_local_path FROM pending_path_update WHERE entity_type = ? AND entity_id = ?",
		entityType, entityId,
	)
	if err == nil && (utils.FileExists(existingPath) || utils.DirExists(existingPath)) {
		return nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	_, err = tx.Exec(`INSERT INTO pending_path_update (entity_type, entity_id, current_local_path)
		VALUES (?, ?, ?) ON CONFLICT(entity_type, entity_id)
		DO UPDATE SET current_local_path = excluded.current_local_path`,
		entityType, entityId, filepath.Clean(localPath),
	)
	return err
}

// ReconcilePendingPathUpdates removes candidates not changed by the merge.
func ReconcilePendingPathUpdates(tx *sqlx.Tx) error {
	bindings := []pendingPathUpdate{}
	if err := tx.Select(&bindings, "SELECT entity_type, entity_id, current_local_path FROM pending_path_update"); err != nil {
		return err
	}
	for _, binding := range bindings {
		desiredPath, err := desiredEntityPath(tx, binding.EntityType, binding.EntityId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if err := deletePendingPath(tx, binding.EntityType, binding.EntityId); err != nil {
					return err
				}
				continue
			}
			return err
		}
		pathExists := utils.FileExists(binding.CurrentLocalPath) || utils.DirExists(binding.CurrentLocalPath)
		if pathsEqual(binding.CurrentLocalPath, desiredPath) || !pathExists {
			if err := deletePendingPath(tx, binding.EntityType, binding.EntityId); err != nil {
				return err
			}
		}
	}
	return nil
}

// ApplyAssetPathUpdate renames a fetched asset to its current metadata path.
func ApplyAssetPathUpdate(tx *sqlx.Tx, assetId string) error {
	ancestorId, err := topmostPendingCollection(tx, assetId)
	if err != nil {
		return err
	}
	if ancestorId != "" {
		return ApplyCollectionPathUpdate(tx, ancestorId)
	}
	completedMoves := []completedPathMove{}
	if err := applySinglePendingPath(tx, pendingAsset, assetId, &completedMoves); err != nil {
		return rollbackPathMoves(completedMoves, err)
	}
	return nil
}

// ApplyCollectionPathUpdate renames a collection and affected descendants.
func ApplyCollectionPathUpdate(tx *sqlx.Tx, collectionId string) error {
	rootId, err := topmostPendingCollectionForCollection(tx, collectionId)
	if err != nil {
		return err
	}
	if rootId != "" {
		collectionId = rootId
	}

	completedMoves := []completedPathMove{}
	collectionIds := []string{}
	if collectionId == "" {
		if err := tx.Select(&collectionIds, "SELECT id FROM collection ORDER BY length(collection_path)"); err != nil {
			return err
		}
	} else {
		collectionIds, err = collectionSubtreeIds(tx, collectionId)
		if err != nil {
			return err
		}
	}
	for _, id := range collectionIds {
		if err := applySinglePendingPath(tx, pendingCollection, id, &completedMoves); err != nil {
			return rollbackPathMoves(completedMoves, err)
		}
	}
	for _, id := range collectionIds {
		assetIds := []string{}
		if err := tx.Select(&assetIds, `SELECT p.entity_id FROM pending_path_update p
			JOIN asset ON asset.id = p.entity_id
			WHERE p.entity_type = 'asset' AND asset.collection_id = ?`, id); err != nil {
			return rollbackPathMoves(completedMoves, err)
		}
		for _, assetId := range assetIds {
			if err := applySinglePendingPath(tx, pendingAsset, assetId, &completedMoves); err != nil {
				return rollbackPathMoves(completedMoves, err)
			}
		}
	}
	if collectionId == "" {
		assetIds := []string{}
		if err := tx.Select(&assetIds, `SELECT p.entity_id FROM pending_path_update p
			JOIN asset ON asset.id = p.entity_id
			WHERE p.entity_type = 'asset' AND asset.collection_id = ''`); err != nil {
			return rollbackPathMoves(completedMoves, err)
		}
		for _, assetId := range assetIds {
			if err := applySinglePendingPath(tx, pendingAsset, assetId, &completedMoves); err != nil {
				return rollbackPathMoves(completedMoves, err)
			}
		}
	}
	return nil
}

func rollbackPathMoves(completedMoves []completedPathMove, cause error) error {
	rollbackErrors := []string{}
	for i := len(completedMoves) - 1; i >= 0; i-- {
		move := completedMoves[i]
		if err := utils.RenamePathCaseSafe(move.newPath, move.oldPath); err != nil {
			rollbackErrors = append(rollbackErrors, err.Error())
		}
	}
	if len(rollbackErrors) > 0 {
		return fmt.Errorf("%w; rollback failed: %s", cause, strings.Join(rollbackErrors, "; "))
	}
	return cause
}

func applySinglePendingPath(tx *sqlx.Tx, entityType, entityId string, completedMoves *[]completedPathMove) error {
	binding := pendingPathUpdate{}
	err := tx.Get(&binding, `SELECT entity_type, entity_id, current_local_path FROM pending_path_update
		WHERE entity_type = ? AND entity_id = ?`, entityType, entityId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	desiredPath, err := desiredEntityPath(tx, entityType, entityId)
	if err != nil {
		return err
	}
	if pathsEqual(binding.CurrentLocalPath, desiredPath) {
		return deletePendingPath(tx, entityType, entityId)
	}
	workingDirectory, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return err
	}
	if err := validatePendingMove(workingDirectory, binding.CurrentLocalPath, desiredPath); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(desiredPath), os.ModePerm); err != nil {
		return err
	}
	if err := utils.RenamePathCaseSafe(binding.CurrentLocalPath, desiredPath); err != nil {
		return err
	}
	if completedMoves != nil {
		*completedMoves = append(*completedMoves, completedPathMove{oldPath: binding.CurrentLocalPath, newPath: desiredPath})
	}
	if entityType == pendingCollection {
		if err := rebasePendingPaths(tx, binding.CurrentLocalPath, desiredPath); err != nil {
			return err
		}
	}
	return deletePendingPath(tx, entityType, entityId)
}

func desiredEntityPath(tx *sqlx.Tx, entityType, entityId string) (string, error) {
	workingDirectory, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return "", err
	}
	if entityType == pendingCollection {
		collection := models.Collection{}
		if err := tx.Get(&collection, "SELECT collection_path FROM collection WHERE id = ?", entityId); err != nil {
			return "", err
		}
		return utils.BuildCollectionPath(workingDirectory, collection.CollectionPath)
	}
	asset := models.Asset{}
	query := `SELECT asset.name, asset.extension, IFNULL(collection.collection_path, '') AS collection_path
		FROM asset LEFT JOIN collection ON collection.id = asset.collection_id WHERE asset.id = ?`
	if err := tx.Get(&asset, query, entityId); err != nil {
		return "", err
	}
	return utils.BuildAssetPath(workingDirectory, asset.CollectionPath, asset.Name, asset.Extension)
}

func collectionSubtreeIds(tx *sqlx.Tx, collectionId string) ([]string, error) {
	collections := []models.Collection{}
	query := `WITH RECURSIVE subtree(id, collection_path) AS (
		SELECT id, collection_path FROM collection WHERE id = ?
		UNION ALL
		SELECT collection.id, collection.collection_path FROM collection JOIN subtree ON collection.parent_id = subtree.id
	) SELECT id, collection_path FROM subtree ORDER BY length(collection_path)`
	if err := tx.Select(&collections, query, collectionId); err != nil {
		return nil, err
	}
	ids := make([]string, len(collections))
	for i, collection := range collections {
		ids[i] = collection.Id
	}
	return ids, nil
}

func topmostPendingCollection(tx *sqlx.Tx, assetId string) (string, error) {
	collectionId := ""
	if err := tx.Get(&collectionId, "SELECT collection_id FROM asset WHERE id = ?", assetId); err != nil {
		return "", err
	}
	return topmostPendingCollectionForCollection(tx, collectionId)
}

func topmostPendingCollectionForCollection(tx *sqlx.Tx, collectionId string) (string, error) {
	if collectionId == "" {
		return "", nil
	}
	ids := []string{}
	query := `WITH RECURSIVE ancestors(id, parent_id, depth) AS (
		SELECT id, parent_id, 0 FROM collection WHERE id = ?
		UNION ALL
		SELECT collection.id, collection.parent_id, ancestors.depth + 1
		FROM collection JOIN ancestors ON ancestors.parent_id = collection.id
	) SELECT ancestors.id FROM ancestors JOIN pending_path_update p
		ON p.entity_type = 'collection' AND p.entity_id = ancestors.id ORDER BY depth DESC`
	if err := tx.Select(&ids, query, collectionId); err != nil {
		return "", err
	}
	if len(ids) == 0 {
		return "", nil
	}
	return ids[0], nil
}

// GetCollectionPathUpdateRoot returns the highest pending ancestor affected.
func GetCollectionPathUpdateRoot(tx *sqlx.Tx, collectionId string) (string, error) {
	rootId, err := topmostPendingCollectionForCollection(tx, collectionId)
	if err != nil || rootId != "" {
		return rootId, err
	}
	return collectionId, nil
}

func rebasePendingPaths(tx *sqlx.Tx, oldRoot, newRoot string) error {
	bindings := []pendingPathUpdate{}
	if err := tx.Select(&bindings, "SELECT entity_type, entity_id, current_local_path FROM pending_path_update"); err != nil {
		return err
	}
	for _, binding := range bindings {
		relative, err := filepath.Rel(oldRoot, binding.CurrentLocalPath)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			continue
		}
		rebasedPath := filepath.Join(newRoot, relative)
		if _, err := tx.Exec(`UPDATE pending_path_update SET current_local_path = ?
			WHERE entity_type = ? AND entity_id = ?`, rebasedPath, binding.EntityType, binding.EntityId); err != nil {
			return err
		}
	}
	return nil
}

func validatePendingMove(workingDirectory, oldPath, newPath string) error {
	if !pathWithin(workingDirectory, oldPath) || !pathWithin(workingDirectory, newPath) {
		return errors.New("path update is outside the project working directory")
	}
	if !utils.FileExists(oldPath) && !utils.DirExists(oldPath) {
		return fmt.Errorf("local path does not exist: %s", oldPath)
	}
	caseOnlyRename := strings.EqualFold(filepath.Clean(oldPath), filepath.Clean(newPath))
	if !caseOnlyRename && (utils.FileExists(newPath) || utils.DirExists(newPath)) {
		return fmt.Errorf("target path already exists: %s", newPath)
	}
	return nil
}

func pathWithin(root, path string) bool {
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func pathsEqual(left, right string) bool {
	return filepath.Clean(left) == filepath.Clean(right)
}

func deletePendingPath(tx *sqlx.Tx, entityType, entityId string) error {
	_, err := tx.Exec("DELETE FROM pending_path_update WHERE entity_type = ? AND entity_id = ?", entityType, entityId)
	return err
}
