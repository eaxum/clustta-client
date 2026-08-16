package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/error_service"
	"clustta/internal/ignore"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/sync_service"
	"clustta/internal/utils"
	"clustta/output"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type CollectionItems struct {
	Assets           []models.Asset               `json:"assets"`
	Collections      []models.Collection          `json:"collections"`
	UntrackedFiles   []models.UntrackedAsset      `json:"untracked_assets"`
	UntrackedFolders []models.UntrackedCollection `json:"untracked_collections"`
}

type CollectionStateFlags struct {
	HasUntracked     bool `json:"has_untracked"`
	HasModified      bool `json:"has_modified"`
	HasOutdated      bool `json:"has_outdated"`
	HasFetchable     bool `json:"has_fetchable"`
	HasRenamePending bool `json:"has_rename_pending"`
}

type CollectionChildrenState struct {
	ModifiedAssets      []models.Asset               `json:"modified_assets"`
	OutdatedAssets      []models.Asset               `json:"outdated_assets"`
	FetchableAssets     []models.Asset               `json:"fetchable_assets"`
	NormalAssets        []models.Asset               `json:"normal_assets"`
	RenamePendingAssets []models.Asset               `json:"rename_pending_assets"`
	UntrackedFiles      []models.UntrackedAsset      `json:"untracked_files"`
	UntrackedFolders    []models.UntrackedCollection `json:"untracked_folders"`
}

type ItemsForCheckpoint struct {
	ModifiedAssets []models.Asset          `json:"modified_assets"`
	UntrackedFiles []models.UntrackedAsset `json:"untracked_files"`
}

type ItemsForUpdate struct {
	OutdatedAssets []models.Asset `json:"outdated_assets"`
}

type PurgeRecursiveUntrackedItemsResult struct {
	DeletedFiles   int      `json:"deleted_files"`
	DeletedFolders int      `json:"deleted_folders"`
	Skipped        int      `json:"skipped"`
	Errors         []string `json:"errors"`
}

type CollectionService struct {
}

func authorizeCreateCollectionTx(tx *sqlx.Tx, parentId string) error {
	activeUser, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}
	user, err := repository.GetUser(tx, activeUser.Id)
	if err != nil {
		return err
	}
	role, err := repository.GetRole(tx, user.RoleId)
	if err != nil {
		return err
	}
	if !role.CreateCollection {
		return errors.New("user does not have create_collection permission")
	}
	if role.Name == "admin" {
		return nil
	}
	if parentId == "" {
		return errors.New("user cannot create collections at project root")
	}
	canModifyIds, err := repository.GetUserCanModifyCollectionIds(tx, user.Id)
	if err != nil {
		return err
	}
	if _, allowed := canModifyIds[parentId]; !allowed {
		return errors.New("user cannot create collections outside assigned collection scope")
	}
	return nil
}

func authorizeUpdateCollectionTx(tx *sqlx.Tx, collectionId string) error {
	activeUser, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}
	user, err := repository.GetUser(tx, activeUser.Id)
	if err != nil {
		return err
	}
	if !user.Role.UpdateCollection {
		return errors.New("user does not have update_collection permission")
	}
	if user.Role.Name == "admin" {
		return nil
	}
	canModifyIds, err := repository.GetUserCanModifyCollectionIds(tx, user.Id)
	if err != nil {
		return err
	}
	if _, allowed := canModifyIds[collectionId]; !allowed {
		return errors.New("user cannot update collection outside assigned collection scope")
	}
	return nil
}

func authorizeCollectionPathUpdateTx(tx *sqlx.Tx, collectionId string) error {
	activeUser, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}
	user, err := repository.GetUser(tx, activeUser.Id)
	if err != nil {
		return err
	}
	role, err := repository.GetRole(tx, user.RoleId)
	if err != nil {
		return err
	}
	if !role.PullChunk {
		return errors.New("user does not have pull_chunk permission")
	}
	if role.Name == "admin" {
		return nil
	}
	canModifyIds, err := repository.GetUserCanModifyCollectionIds(tx, user.Id)
	if err != nil {
		return err
	}
	if _, allowed := canModifyIds[collectionId]; !allowed {
		return errors.New("user cannot update local paths outside assigned collection scope")
	}
	return nil
}

type untrackedCanModifyContext struct {
	allowAll            bool
	canModifyIds        map[string]struct{}
	collectionIdsByPath map[string]string
}

// stampCollectionsCanModify sets CanModify on each collection based on the
// active user's admin role or collaborator scope.
func (e *CollectionService) stampCollectionsCanModify(tx *sqlx.Tx, collections []models.Collection) error {
	if len(collections) == 0 {
		return nil
	}
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}
	userData, err := repository.GetUser(tx, user.Id)
	if err != nil {
		return err
	}
	role, err := repository.GetRole(tx, userData.RoleId)
	if err != nil {
		return err
	}
	if role.Name == "admin" {
		for i := range collections {
			collections[i].CanModify = true
		}
		return nil
	}
	canModify, err := repository.GetUserCanModifyCollectionIds(tx, user.Id)
	if err != nil {
		return err
	}
	for i := range collections {
		if _, ok := canModify[collections[i].Id]; ok {
			collections[i].CanModify = true
		}
	}
	return nil
}

// stampCollectionCanModify sets CanModify on a single collection.
func (e *CollectionService) stampCollectionCanModify(tx *sqlx.Tx, c *models.Collection) error {
	list := []models.Collection{*c}
	if err := e.stampCollectionsCanModify(tx, list); err != nil {
		return err
	}
	c.CanModify = list[0].CanModify
	return nil
}

func (e *CollectionService) buildUntrackedCanModifyContext(tx *sqlx.Tx, includePathResolver bool) (untrackedCanModifyContext, error) {
	context := untrackedCanModifyContext{
		canModifyIds: map[string]struct{}{},
	}

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return context, err
	}
	userData, err := repository.GetUser(tx, user.Id)
	if err != nil {
		return context, err
	}
	role, err := repository.GetRole(tx, userData.RoleId)
	if err != nil {
		return context, err
	}
	context.allowAll = role.Name == "admin"

	if !context.allowAll {
		context.canModifyIds, err = repository.GetUserCanModifyCollectionIds(tx, user.Id)
		if err != nil {
			return context, err
		}
	}

	if includePathResolver {
		context.collectionIdsByPath = map[string]string{}
		type trackedCollectionPath struct {
			Id             string `db:"id"`
			CollectionPath string `db:"collection_path"`
		}
		trackedCollections := []trackedCollectionPath{}
		if err := tx.Select(&trackedCollections, "SELECT id, collection_path FROM full_collection WHERE trashed = 0"); err != nil {
			return context, err
		}
		for _, collection := range trackedCollections {
			context.collectionIdsByPath[normalizeCollectionLookupPath(collection.CollectionPath)] = collection.Id
		}
	}

	return context, nil
}

func normalizeCollectionLookupPath(path string) string {
	path = utils.NormalizePath(path)
	path = strings.Trim(path, "/")
	if path == "." {
		return ""
	}
	return path
}

func parentLookupPath(path string) string {
	path = normalizeCollectionLookupPath(path)
	if path == "" {
		return ""
	}
	parent := normalizeCollectionLookupPath(filepath.Dir(path))
	if parent == "." {
		return ""
	}
	return parent
}

func (context untrackedCanModifyContext) canModifyCollection(collectionId string) bool {
	if context.allowAll {
		return true
	}
	if collectionId == "" {
		return false
	}
	_, ok := context.canModifyIds[collectionId]
	return ok
}

func (context untrackedCanModifyContext) canModifyPath(path string, fallbackCollectionId string) bool {
	lookupPath := normalizeCollectionLookupPath(path)
	if context.collectionIdsByPath != nil {
		for {
			if collectionId, ok := context.collectionIdsByPath[lookupPath]; ok {
				return context.canModifyCollection(collectionId)
			}
			if lookupPath == "" {
				break
			}
			lookupPath = parentLookupPath(lookupPath)
		}
	}
	return context.canModifyCollection(fallbackCollectionId)
}

// GetCollectionCount returns the total number of collections in the project.
// Returns the count or an error if the operation fails.
func (t *CollectionService) GetCollectionCount(projectPath string) (int, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return 0, err
	}
	defer dbConn.Close()

	var count int
	query := "SELECT COUNT(*) FROM full_collection"

	err = dbConn.Get(&count, query)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// CreateCollection creates a new collection in the project.
// Returns the created collection or an error if the operation fails.
func (e *CollectionService) CreateCollection(projectPath, name, description, collectionTypeId, parentId, previewPath string, isShared bool) (models.Collection, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Collection{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Collection{}, err
	}
	defer tx.Rollback()
	if err := authorizeCreateCollectionTx(tx, parentId); err != nil {
		return models.Collection{}, err
	}

	previewId := ""
	if previewPath != "" {
		preview, err := repository.CreatePreview(tx, previewPath)
		if err != nil {
			tx.Rollback()
			return models.Collection{}, err
		}
		previewId = preview.Hash
	}

	createdCollection, err := repository.CreateCollection(
		tx,
		"",
		name,
		description,
		collectionTypeId,
		parentId,
		previewId,
		isShared,
	)
	if err != nil {
		tx.Rollback()
		return models.Collection{}, err
	}
	err = tx.Commit()
	if err != nil {
		return models.Collection{}, err
	}
	return createdCollection, nil
}

// RenameCollection renames an existing collection.
// Returns the updated collection or an error if the operation fails.
func (e *CollectionService) RenameCollection(projectPath, collectionId, newName string) (models.Collection, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Collection{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Collection{}, err
	}
	defer tx.Rollback()

	updatedCollection, err := repository.RenameCollection(tx, collectionId, newName)
	if err != nil {
		tx.Rollback()
		return models.Collection{}, err
	}

	err = tx.Commit()
	if err != nil {
		return models.Collection{}, err
	}
	return updatedCollection, nil
}

// ApplyPathUpdate applies pending remote path changes to a collection tree.
func (e *CollectionService) ApplyPathUpdate(projectPath, collectionId string) error {
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
	pathUpdateRoot, err := repository.GetCollectionPathUpdateRoot(tx, collectionId)
	if err != nil {
		return err
	}
	if err := authorizeCollectionPathUpdateTx(tx, pathUpdateRoot); err != nil {
		return err
	}
	if err := repository.ApplyCollectionPathUpdate(tx, collectionId); err != nil {
		return err
	}
	return tx.Commit()
}

// CreateCollections creates multiple collection collections in bulk.
// Currently a stub implementation for future batch creation functionality.
func (e *CollectionService) CreateCollections(projectPath, name, description, collectionTypeId, parentId string) ([]models.Collection, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Collection{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Collection{}, err
	}
	defer tx.Rollback()

	return []models.Collection{}, nil
}

// DeleteCollection removes a collection from the project.
// Optionally removes associated files if removeFiles is true.
func (e *CollectionService) DeleteCollection(projectPath, collectionId string, removeFiles bool) error {
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

	err = repository.DeleteCollection(tx, collectionId, removeFiles, true)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// GetCollections retrieves collections based on user permissions.
// Returns all collections or only user-accessible collections based on role.
func (e *CollectionService) GetCollections(projectPath string) ([]models.Collection, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Collection{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Collection{}, err
	}
	defer tx.Rollback()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return []models.Collection{}, err
	}
	userData, err := repository.GetUser(tx, user.Id)
	if err != nil {
		return []models.Collection{}, err
	}
	userRole, err := repository.GetRole(tx, userData.RoleId)
	if err != nil {
		return []models.Collection{}, err
	}

	if userRole.ViewAsset {
		collections, err := repository.GetCollections(tx, true)
		if err != nil {
			return []models.Collection{}, err
		}
		if err := e.stampCollectionsCanModify(tx, collections); err != nil {
			return []models.Collection{}, err
		}
		return collections, nil
	} else {
		userAssetInfo, err := repository.GetUserAssetsMinimal(tx, user.Id)
		if err != nil {
			return []models.Collection{}, err
		}

		collections, err := repository.GetUserCollections(tx, userAssetInfo, user.Id)
		if err != nil {
			return []models.Collection{}, err
		}
		return collections, err
	}
}

// GetCollectionChildren retrieves all children of a collection including tracked and untracked items.
// Returns separate lists for assets, collections, and untracked items.
func (e *CollectionService) GetCollectionChildren(projectPath, collectionId, projectWorkingDir, collectionFolderPath string, ignoreList []string, isUntracked bool) (CollectionItems, error) {
	children := CollectionItems{
		Assets:           make([]models.Asset, 0),
		Collections:      make([]models.Collection, 0),
		UntrackedFiles:   make([]models.UntrackedAsset, 0),
		UntrackedFolders: make([]models.UntrackedCollection, 0),
	}
	if collectionId == "root" {
		collectionId = ""
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return children, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return children, err
	}
	defer tx.Rollback()
	canModifyContext, err := e.buildUntrackedCanModifyContext(tx, isUntracked)
	if err != nil {
		return children, err
	}

	collectionTrackFolders := []string{}
	collectionTrackFiles := []string{}
	if !isUntracked {
		if collectionId == "root" {
			collectionId = ""
		}

		collections, err := repository.GetCollectionChildren(tx, collectionId)
		if err != nil {
			return children, err
		}
		if err := e.stampCollectionsCanModify(tx, collections); err != nil {
			return children, err
		}

		assets, err := repository.GetCollectionAssets(tx, collectionId)
		if err != nil {
			return children, err
		}
		collections, assets, err = filterChildrenByViewPermission(tx, collections, assets)
		if err != nil {
			return children, err
		}
		children.Collections = collections
		children.Assets = assets

		for _, child := range collections {
			if child.LocalPath != "" {
				collectionTrackFolders = append(collectionTrackFolders, filepath.Base(filepath.Clean(child.LocalPath)))
				continue
			}
			collectionTrackFolders = append(collectionTrackFolders, child.Name)
		}

		for _, child := range assets {
			if child.LocalPath != "" {
				collectionTrackFiles = append(collectionTrackFiles, filepath.Base(child.LocalPath))
				continue
			}
			collectionTrackFiles = append(collectionTrackFiles, child.Name+child.Extension)
		}

		if collectionId != "" {
			collection, err := repository.GetCollection(tx, collectionId)
			if err != nil {
				return children, err
			}
			collectionFolderPath = collection.GetFilePath()
		}
		pendingFolders, pendingFiles, err := pendingNamesInFolder(tx, collectionFolderPath)
		if err != nil {
			return children, err
		}
		collectionTrackFolders = append(collectionTrackFolders, pendingFolders...)
		collectionTrackFiles = append(collectionTrackFiles, pendingFiles...)
	}

	if !utils.DirExists(collectionFolderPath) {
		return children, nil
	}

	absoluteCollectionFolderPath, err := filepath.Abs(collectionFolderPath)
	if err != nil {
		return children, err
	}

	relativeCollectionFolderPath, err := filepath.Rel(projectWorkingDir, absoluteCollectionFolderPath)
	if err != nil {
		return children, err
	}
	relativeCollectionFolderPath = utils.NormalizePath(relativeCollectionFolderPath)

	clusttaIgnore := ignore.CompileIgnoreLines(ignoreList...)

	entries, err := os.ReadDir(absoluteCollectionFolderPath)
	if err != nil {
		return children, err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(absoluteCollectionFolderPath, entry.Name())
		relativePath := utils.NormalizePath(filepath.Join(relativeCollectionFolderPath, entry.Name()))
		parentId := collectionId
		if parentId == "root" {
			parentId = ""
		}
		if entry.IsDir() {
			if slices.Contains(collectionTrackFolders, entry.Name()) {
				continue
			}
			if !clusttaIgnore.MatchesPath(relativePath) {
				untrackedFolder := models.UntrackedCollection{
					Id:             utils.GetMD5Hash(entryPath),
					Name:           entry.Name(),
					FilePath:       entryPath,
					CollectionPath: "/" + relativePath + "/",
					ItemPath:       "/" + relativePath + "/",
					ParentId:       parentId,
				}
				untrackedFolder.CanModify = canModifyContext.canModifyPath(parentLookupPath(untrackedFolder.ItemPath), untrackedFolder.ParentId)
				children.UntrackedFolders = append(children.UntrackedFolders, untrackedFolder)
			}
		} else {
			if slices.Contains(collectionTrackFiles, entry.Name()) {
				continue
			}

			if !clusttaIgnore.MatchesPath(relativePath) {
				assetName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
				untrackedFile := models.UntrackedAsset{
					Id:             utils.GetMD5Hash(entryPath),
					Name:           assetName,
					FilePath:       entryPath,
					AssetPath:      "/" + relativePath,
					CollectionId:   parentId,
					CollectionPath: "/" + relativeCollectionFolderPath + "/",
					Extension:      filepath.Ext(entry.Name()),
					ItemPath:       "/" + relativePath + "/",
					AssetTypeIcon:  "generic",
				}
				untrackedFile.CanModify = canModifyContext.canModifyPath(untrackedFile.CollectionPath, untrackedFile.CollectionId)
				children.UntrackedFiles = append(children.UntrackedFiles, untrackedFile)
			}
		}
	}

	return children, nil
}

func filterChildrenByViewPermission(tx *sqlx.Tx, collections []models.Collection, assets []models.Asset) ([]models.Collection, []models.Asset, error) {
	activeUser, err := auth_service.GetActiveUser()
	if err != nil {
		return nil, nil, err
	}
	user, err := repository.GetUser(tx, activeUser.Id)
	if err != nil {
		return nil, nil, err
	}
	role, err := repository.GetRole(tx, user.RoleId)
	if err != nil {
		return nil, nil, err
	}
	if role.ViewAsset {
		return collections, assets, nil
	}

	visibleAssets, err := repository.GetUserAssetsMinimal(tx, user.Id)
	if err != nil {
		return nil, nil, err
	}
	visibleCollections, err := repository.GetUserCollections(tx, visibleAssets, user.Id)
	if err != nil {
		return nil, nil, err
	}

	filteredCollections, filteredAssets := filterVisibleChildren(collections, assets, visibleCollections, visibleAssets)
	return filteredCollections, filteredAssets, nil
}

const pendingPathCollectionEntity = "collection"

func pendingNamesInFolder(tx *sqlx.Tx, folderPath string) ([]string, []string, error) {
	pendingPaths := []struct {
		EntityType string `db:"entity_type"`
		LocalPath  string `db:"current_local_path"`
	}{}
	if err := tx.Select(&pendingPaths, "SELECT entity_type, current_local_path FROM pending_path_update"); err != nil {
		return nil, nil, err
	}
	folders := []string{}
	files := []string{}
	for _, pendingPath := range pendingPaths {
		if !sameFilesystemPath(filepath.Dir(pendingPath.LocalPath), folderPath) {
			continue
		}
		name := filepath.Base(filepath.Clean(pendingPath.LocalPath))
		if pendingPath.EntityType == pendingPathCollectionEntity {
			folders = append(folders, name)
		} else {
			files = append(files, name)
		}
	}
	return folders, files, nil
}

func sameFilesystemPath(left, right string) bool {
	left = filepath.Clean(left)
	right = filepath.Clean(right)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}

func filterVisibleChildren(collections []models.Collection, assets []models.Asset, visibleCollections []models.Collection, visibleAssets []models.Asset) ([]models.Collection, []models.Asset) {
	visibleCollectionIds := make(map[string]struct{}, len(visibleCollections))
	for _, collection := range visibleCollections {
		visibleCollectionIds[collection.Id] = struct{}{}
	}
	filteredCollections := make([]models.Collection, 0, len(collections))
	for _, collection := range collections {
		if _, visible := visibleCollectionIds[collection.Id]; visible {
			filteredCollections = append(filteredCollections, collection)
		}
	}

	visibleAssetIds := make(map[string]struct{}, len(visibleAssets))
	for _, asset := range visibleAssets {
		visibleAssetIds[asset.Id] = struct{}{}
	}
	filteredAssets := make([]models.Asset, 0, len(assets))
	for _, asset := range assets {
		if _, visible := visibleAssetIds[asset.Id]; visible {
			filteredAssets = append(filteredAssets, asset)
		}
	}

	return filteredCollections, filteredAssets
}

// GetRecursiveUntrackedAssets returns all untracked files under a collection folder.
// Tracked asset files are excluded, but tracked collection folders are still scanned
// so untracked files inside child collections are included.
func (e *CollectionService) GetRecursiveUntrackedAssets(projectPath, collectionId, projectWorkingDir, collectionFolderPath string, ignoreList []string) ([]models.UntrackedAsset, error) {
	untrackedAssets := make([]models.UntrackedAsset, 0)

	if collectionId == "root" {
		collectionId = ""
	}
	if collectionFolderPath == "" {
		collectionFolderPath = projectWorkingDir
	}
	if !utils.DirExists(collectionFolderPath) {
		return untrackedAssets, nil
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return untrackedAssets, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return untrackedAssets, err
	}
	defer tx.Rollback()
	canModifyContext, err := e.buildUntrackedCanModifyContext(tx, true)
	if err != nil {
		return untrackedAssets, err
	}

	type trackedAssetFile struct {
		AssetPath string `db:"asset_path"`
		Extension string `db:"extension"`
		LocalPath string `db:"local_path"`
	}
	trackedAssets := []trackedAssetFile{}
	err = tx.Select(&trackedAssets, "SELECT asset_path, extension, local_path FROM full_asset WHERE trashed = 0")
	if err != nil {
		return untrackedAssets, err
	}

	absoluteTrackedFiles := make(map[string]bool, len(trackedAssets))
	for _, asset := range trackedAssets {
		if asset.LocalPath != "" {
			absoluteTrackedFiles[filepath.Clean(asset.LocalPath)] = true
			continue
		}
		absoluteAssetPath, err := filepath.Abs(filepath.Join(projectWorkingDir, asset.AssetPath+asset.Extension))
		if err != nil {
			return untrackedAssets, err
		}
		absoluteTrackedFiles[absoluteAssetPath] = true
	}

	collectionIdForPath := func(relativePath string) string {
		relativePath = normalizeCollectionLookupPath(relativePath)
		for {
			if id, ok := canModifyContext.collectionIdsByPath[relativePath]; ok {
				return id
			}
			if relativePath == "" {
				break
			}
			relativePath = parentLookupPath(relativePath)
		}
		return collectionId
	}

	rootedDirPath := func(relativePath string) string {
		relativePath = strings.Trim(utils.NormalizePath(relativePath), "/")
		if relativePath == "" || relativePath == "." {
			return "/"
		}
		return "/" + relativePath + "/"
	}

	absoluteCollectionFolderPath, err := filepath.Abs(collectionFolderPath)
	if err != nil {
		return untrackedAssets, err
	}

	clusttaIgnore := ignore.CompileIgnoreLines(ignoreList...)
	err = filepath.WalkDir(absoluteCollectionFolderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink != 0 {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		relativePath, err := filepath.Rel(projectWorkingDir, path)
		if err != nil {
			return err
		}
		relativePath = utils.NormalizePath(relativePath)

		isRoot := path == absoluteCollectionFolderPath
		if d.IsDir() {
			if !isRoot && strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}
			if !isRoot && clusttaIgnore.MatchesPath(relativePath) {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		if absoluteTrackedFiles[path] || clusttaIgnore.MatchesPath(relativePath) {
			return nil
		}

		parentRelativePath := utils.NormalizePath(filepath.Dir(relativePath))
		if parentRelativePath == "." {
			parentRelativePath = ""
		}
		assetName := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		untrackedFile := models.UntrackedAsset{
			Id:             utils.GetMD5Hash(path),
			Name:           assetName,
			FilePath:       path,
			AssetPath:      "/" + relativePath,
			CollectionId:   collectionIdForPath(parentRelativePath),
			CollectionPath: rootedDirPath(parentRelativePath),
			Extension:      filepath.Ext(d.Name()),
			ItemPath:       "/" + relativePath + "/",
			AssetTypeIcon:  "generic",
		}
		untrackedFile.CanModify = canModifyContext.canModifyPath(untrackedFile.CollectionPath, untrackedFile.CollectionId)
		untrackedAssets = append(untrackedAssets, untrackedFile)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return untrackedAssets, nil
}

// PurgeRecursiveUntrackedItems deletes untracked files under a collection folder.
// Ignore rules are intentionally not applied: ignored files are still untracked
// and should be removed by an explicit purge.
func (e *CollectionService) PurgeRecursiveUntrackedItems(projectPath, collectionId, projectWorkingDir, collectionFolderPath string) (PurgeRecursiveUntrackedItemsResult, error) {
	result := PurgeRecursiveUntrackedItemsResult{
		Errors: []string{},
	}

	if collectionId == "root" {
		collectionId = ""
	}
	if collectionFolderPath == "" {
		collectionFolderPath = projectWorkingDir
	}
	if !utils.DirExists(collectionFolderPath) {
		return result, nil
	}

	absoluteProjectWorkingDir, err := filepath.Abs(projectWorkingDir)
	if err != nil {
		return result, err
	}
	absoluteTargetPath, err := filepath.Abs(collectionFolderPath)
	if err != nil {
		return result, err
	}
	relativeTarget, err := filepath.Rel(absoluteProjectWorkingDir, absoluteTargetPath)
	if err != nil {
		return result, err
	}
	if relativeTarget == ".." || strings.HasPrefix(relativeTarget, ".."+string(filepath.Separator)) || filepath.IsAbs(relativeTarget) {
		return result, fmt.Errorf("purge target must be inside the project working directory")
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return result, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return result, err
	}
	defer tx.Rollback()

	type trackedAssetFile struct {
		AssetPath string `db:"asset_path"`
		Extension string `db:"extension"`
		LocalPath string `db:"local_path"`
	}
	trackedAssets := []trackedAssetFile{}
	err = tx.Select(&trackedAssets, "SELECT asset_path, extension, local_path FROM full_asset WHERE trashed = 0")
	if err != nil {
		return result, err
	}

	pathKey := func(path string) string {
		cleanedPath := filepath.Clean(path)
		if runtime.GOOS == "windows" {
			return strings.ToLower(cleanedPath)
		}
		return cleanedPath
	}

	absoluteTrackedFiles := make(map[string]bool, len(trackedAssets))
	for _, asset := range trackedAssets {
		if asset.LocalPath != "" {
			absoluteTrackedFiles[pathKey(asset.LocalPath)] = true
			continue
		}
		assetPath := strings.Trim(asset.AssetPath+asset.Extension, "/\\")
		absoluteAssetPath, err := filepath.Abs(filepath.Join(absoluteProjectWorkingDir, assetPath))
		if err != nil {
			return result, err
		}
		absoluteTrackedFiles[pathKey(absoluteAssetPath)] = true
	}

	type trackedCollectionPath struct {
		CollectionPath string `db:"collection_path"`
		LocalPath      string `db:"local_path"`
	}
	trackedCollections := []trackedCollectionPath{}
	err = tx.Select(&trackedCollections, "SELECT collection_path, local_path FROM full_collection WHERE trashed = 0")
	if err != nil {
		return result, err
	}

	absoluteTrackedCollections := map[string]bool{
		pathKey(absoluteProjectWorkingDir): true,
	}
	for _, collection := range trackedCollections {
		if collection.LocalPath != "" {
			absoluteTrackedCollections[pathKey(collection.LocalPath)] = true
			continue
		}
		collectionPath := strings.Trim(utils.NormalizePath(collection.CollectionPath), "/")
		absoluteCollectionPath, err := filepath.Abs(filepath.Join(absoluteProjectWorkingDir, collectionPath))
		if err != nil {
			return result, err
		}
		absoluteTrackedCollections[pathKey(absoluteCollectionPath)] = true
	}

	directories := []string{}
	err = filepath.WalkDir(absoluteTargetPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			return nil
		}

		isRoot := pathKey(path) == pathKey(absoluteTargetPath)
		if d.Type()&fs.ModeSymlink != 0 {
			result.Skipped++
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if !isRoot && strings.HasPrefix(filepath.Base(path), ".") {
				result.Skipped++
				return filepath.SkipDir
			}
			if !isRoot {
				directories = append(directories, path)
			}
			return nil
		}

		if strings.HasPrefix(filepath.Base(path), ".") {
			result.Skipped++
			return nil
		}
		if absoluteTrackedFiles[pathKey(path)] {
			result.Skipped++
			return nil
		}

		if err := os.Remove(path); err != nil {
			result.Errors = append(result.Errors, err.Error())
			return nil
		}
		result.DeletedFiles++
		return nil
	})
	if err != nil {
		return result, err
	}

	sort.Slice(directories, func(i, j int) bool {
		return len(directories[i]) > len(directories[j])
	})

	for _, directory := range directories {
		if absoluteTrackedCollections[pathKey(directory)] {
			result.Skipped++
			continue
		}

		entries, err := os.ReadDir(directory)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			continue
		}
		if len(entries) > 0 {
			result.Skipped++
			continue
		}

		if err := os.Remove(directory); err != nil {
			result.Errors = append(result.Errors, err.Error())
			continue
		}
		result.DeletedFolders++
	}

	return result, nil
}

// GetCollectionAssets retrieves all assets belonging to a specific collection.
// Returns the list of assets or an error if the operation fails.
func (e *CollectionService) GetCollectionAssets(projectPath, collectionId string) ([]models.Asset, error) {
	if collectionId == "root" {
		collectionId = ""
	}
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
	assets, err := repository.GetCollectionAssets(tx, collectionId)
	if err != nil {
		return []models.Asset{}, err
	}
	_, assets, err = filterChildrenByViewPermission(tx, nil, assets)
	if err != nil {
		return []models.Asset{}, err
	}
	return assets, nil
}

// GetCollectionByID retrieves a collection by its ID.
// Returns the collection or an error if not found.
func (e *CollectionService) GetCollectionByID(projectPath, collectionId string) (models.Collection, error) {
	if collectionId == "root" {
		collectionId = ""
	}
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Collection{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Collection{}, err
	}
	defer tx.Rollback()
	collection, err := repository.GetCollection(tx, collectionId)
	if err != nil {
		return collection, err
	}
	if err := e.stampCollectionCanModify(tx, &collection); err != nil {
		return collection, err
	}
	return collection, nil
}

// GetCollectionByPath retrieves a collection by its filesystem path.
// Returns the collection or an error if not found.
func (e *CollectionService) GetCollectionByPath(projectPath, collectionPath string) (models.Collection, error) {
	if collectionPath == "/" {
		collectionPath = ""
	}
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Collection{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Collection{}, err
	}
	defer tx.Rollback()
	collection, err := repository.GetCollectionByPath(tx, collectionPath)
	if err != nil {
		return collection, err
	}
	if err := e.stampCollectionCanModify(tx, &collection); err != nil {
		return collection, err
	}
	return collection, nil
}

// GetCollectionStateFlags checks if a collection has any recursive children with specific states.
// Returns flags indicating presence of untracked, modified, outdated, or fetchable items.
func (e *CollectionService) GetCollectionStateFlags(projectPath, collectionId, projectWorkingDir string, ignoreList []string) (CollectionStateFlags, error) {
	flags := CollectionStateFlags{
		HasUntracked: false,
		HasModified:  false,
		HasOutdated:  false,
		HasFetchable: false,
	}

	if collectionId == "root" {
		collectionId = ""
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return flags, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return flags, err
	}
	defer tx.Rollback()

	pendingQuery := `WITH RECURSIVE subtree(id) AS (
		SELECT id FROM collection WHERE id = ?
		UNION ALL
		SELECT collection.id FROM collection JOIN subtree ON collection.parent_id = subtree.id
	) SELECT EXISTS(
		SELECT 1 FROM pending_path_update
		WHERE entity_type = 'collection' AND (? = '' OR entity_id IN (SELECT id FROM subtree))
		UNION ALL
		SELECT 1 FROM pending_path_update JOIN asset ON asset.id = pending_path_update.entity_id
		WHERE pending_path_update.entity_type = 'asset'
		AND (? = '' OR asset.collection_id IN (SELECT id FROM subtree))
	)`
	if err := tx.Get(&flags.HasRenamePending, pendingQuery, collectionId, collectionId, collectionId); err != nil {
		return flags, err
	}
	if flags.HasRenamePending {
		return flags, nil
	}

	var collectionPath string
	if collectionId == "" {
		collectionPath = ""
	} else {
		collection, err := repository.GetCollection(tx, collectionId)
		if err != nil {
			return flags, err
		}
		collectionPath = collection.CollectionPath
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return flags, err
	}

	const batchSize = 100
	offset := 0

	type assetCheckpointInfo struct {
		assetId            string
		latestChecksum     string
		latestTimeModified int64
		latestFileSize     int64
		checkpointCount    int
		allChecksums       []string
	}

	type modifiedCandidate struct {
		assetId       string
		assetFilePath string
		checkpoints   []string
	}
	var candidatesNeedingHashCheck []modifiedCandidate

	for {
		if flags.HasFetchable && flags.HasModified && flags.HasOutdated {
			break
		}

		var assets []models.Asset
		query := `
			SELECT id, asset_path, extension, collection_path, name
			FROM full_asset 
			WHERE collection_path LIKE ? AND trashed = 0 AND is_link = 0
			ORDER BY collection_path, name
			LIMIT ? OFFSET ?
		`
		err = tx.Select(&assets, query, collectionPath+"%", batchSize, offset)
		if err != nil {
			return flags, err
		}

		if len(assets) == 0 {
			break
		}

		assetIds := make([]string, len(assets))
		for i, asset := range assets {
			assetIds[i] = asset.Id
		}

		checkpointMap := make(map[string]assetCheckpointInfo)
		if len(assetIds) > 0 {
			quotedAssetIds := make([]string, len(assetIds))
			for i, id := range assetIds {
				quotedAssetIds[i] = fmt.Sprintf("'%s'", id)
			}

			checkpointQuery := fmt.Sprintf(`
				SELECT asset_id, xxhash_checksum, time_modified, file_size
				FROM asset_checkpoint 
				WHERE asset_id IN (%s) AND trashed = 0
				ORDER BY asset_id, created_at DESC
			`, strings.Join(quotedAssetIds, ","))

			var checkpoints []struct {
				AssetId        string `db:"asset_id"`
				XXHashChecksum string `db:"xxhash_checksum"`
				TimeModified   int64  `db:"time_modified"`
				FileSize       int64  `db:"file_size"`
			}
			tx.Select(&checkpoints, checkpointQuery)

			for _, cp := range checkpoints {
				if info, exists := checkpointMap[cp.AssetId]; exists {
					info.checkpointCount++
					info.allChecksums = append(info.allChecksums, cp.XXHashChecksum)
					checkpointMap[cp.AssetId] = info
				} else {
					checkpointMap[cp.AssetId] = assetCheckpointInfo{
						assetId:            cp.AssetId,
						latestChecksum:     cp.XXHashChecksum,
						latestTimeModified: cp.TimeModified,
						latestFileSize:     cp.FileSize,
						checkpointCount:    1,
						allChecksums:       []string{cp.XXHashChecksum},
					}
				}
			}
		}

		for _, asset := range assets {
			assetFilePath, err := utils.BuildAssetPath(rootFolder, asset.CollectionPath, asset.Name, asset.Extension)
			if err != nil {
				continue
			}

			fileInfo, err := os.Stat(assetFilePath)
			if os.IsNotExist(err) {
				if !flags.HasFetchable {
					flags.HasFetchable = true
				}
				continue
			}

			if err != nil {
				continue
			}

			if !flags.HasModified || !flags.HasOutdated {
				checkpointInfo, hasCheckpoint := checkpointMap[asset.Id]

				if hasCheckpoint {
					fileSize := fileInfo.Size()

					if fileSize != checkpointInfo.latestFileSize {
						candidatesNeedingHashCheck = append(candidatesNeedingHashCheck, modifiedCandidate{
							assetId:       asset.Id,
							assetFilePath: assetFilePath,
							checkpoints:   checkpointInfo.allChecksums,
						})
					} else {
						fileModTime := fileInfo.ModTime().Unix()

						if fileModTime != checkpointInfo.latestTimeModified {
							candidatesNeedingHashCheck = append(candidatesNeedingHashCheck, modifiedCandidate{
								assetId:       asset.Id,
								assetFilePath: assetFilePath,
								checkpoints:   checkpointInfo.allChecksums,
							})
						}
					}
				}
			}

			if flags.HasFetchable && flags.HasModified && flags.HasOutdated {
				break
			}
		}

		offset += batchSize
	}

	if (!flags.HasModified || !flags.HasOutdated) && len(candidatesNeedingHashCheck) > 0 {
		for _, candidate := range candidatesNeedingHashCheck {
			if flags.HasModified && flags.HasOutdated {
				break
			}

			fileHash, err := utils.GenerateXXHashChecksum(candidate.assetFilePath)
			if err != nil {
				continue
			}

			matchesLatest := (fileHash == candidate.checkpoints[0])
			matchesOlderCheckpoint := false

			if !matchesLatest && len(candidate.checkpoints) > 1 {
				for i := 1; i < len(candidate.checkpoints); i++ {
					if fileHash == candidate.checkpoints[i] {
						matchesOlderCheckpoint = true
						break
					}
				}
			}

			if matchesOlderCheckpoint {
				if !flags.HasOutdated {
					flags.HasOutdated = true
				}
			} else if !matchesLatest {
				if !flags.HasModified {
					flags.HasModified = true
				}
			}
		}
	}

	if !flags.HasUntracked && utils.DirExists(projectWorkingDir) {
		trackedFiles := make(map[string]bool)

		var allAssets []models.Asset
		query := `
			SELECT asset_path, extension, local_path
			FROM full_asset 
			WHERE collection_path LIKE ? AND trashed = 0 AND is_link = 0
		`
		err = tx.Select(&allAssets, query, collectionPath+"%")
		if err != nil {
			return flags, err
		}

		for _, asset := range allAssets {
			if asset.LocalPath != "" {
				trackedFiles[filepath.Clean(asset.LocalPath)] = true
				continue
			}
			assetFilePath, err := filepath.Abs(filepath.Join(projectWorkingDir, asset.AssetPath+asset.Extension))
			if err == nil {
				trackedFiles[assetFilePath] = true
			}
		}

		var folderToScan string
		if collectionId == "" {
			folderToScan = projectWorkingDir
		} else {
			collection, err := repository.GetCollection(tx, collectionId)
			if err != nil {
				return flags, err
			}
			folderToScan = collection.GetFilePath()
		}

		if utils.DirExists(folderToScan) {
			clusttaIgnore := ignore.CompileIgnoreLines(ignoreList...)

			err = filepath.WalkDir(folderToScan, func(path string, d fs.DirEntry, err error) error {
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
					return nil
				}

				if strings.HasPrefix(filepath.Base(path), ".") {
					return nil
				}

				absPath, err := filepath.Abs(path)
				if err != nil {
					return nil
				}

				if !trackedFiles[absPath] {
					relativePath, err := filepath.Rel(projectWorkingDir, path)
					if err != nil {
						return nil
					}
					relativePath = utils.NormalizePath(relativePath)

					if !clusttaIgnore.MatchesPath(relativePath) {
						flags.HasUntracked = true
						return filepath.SkipAll
					}
				}

				return nil
			})

			if err != nil && err != filepath.SkipAll {
				return flags, err
			}
		}
	}

	return flags, nil
}

// GetCollectionChildrenState analyzes the immediate children of a collection to determine their state.
// Returns state containing modified, outdated, fetchable assets and untracked items.
func (e *CollectionService) GetCollectionChildrenState(projectPath, collectionId, projectWorkingDir string, ignoreList []string) (CollectionChildrenState, error) {
	state := CollectionChildrenState{
		ModifiedAssets:      make([]models.Asset, 0),
		OutdatedAssets:      make([]models.Asset, 0),
		FetchableAssets:     make([]models.Asset, 0),
		NormalAssets:        make([]models.Asset, 0),
		RenamePendingAssets: make([]models.Asset, 0),
		UntrackedFiles:      make([]models.UntrackedAsset, 0),
		UntrackedFolders:    make([]models.UntrackedCollection, 0),
	}

	if collectionId == "root" {
		collectionId = ""
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return state, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return state, err
	}
	defer tx.Rollback()

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return state, err
	}

	var assets []models.Asset
	query := `
		SELECT id, asset_path, extension, collection_path, name, collection_id, local_path
		FROM full_asset 
		WHERE collection_id = ? AND trashed = 0
		ORDER BY name
	`
	err = tx.Select(&assets, query, collectionId)
	if err != nil {
		return state, err
	}

	if len(assets) == 0 {
		canModifyContext, err := e.buildUntrackedCanModifyContext(tx, false)
		if err != nil {
			return state, err
		}
		return e.detectUntrackedItems(tx, state, collectionId, projectWorkingDir, rootFolder, ignoreList, canModifyContext)
	}

	type fileMetadata struct {
		size    int64
		modTime int64
	}

	assetsMissingFiles := make([]string, 0)
	assetsWithFiles := make([]string, 0)
	assetFileMetadata := make(map[string]fileMetadata)
	assetMap := make(map[string]models.Asset)

	for _, asset := range assets {
		assetMap[asset.Id] = asset
		if asset.LocalPath != "" && utils.FileExists(asset.LocalPath) {
			asset.FilePath = asset.LocalPath
			assetMap[asset.Id] = asset
			state.RenamePendingAssets = append(state.RenamePendingAssets, asset)
			continue
		}

		assetFilePath, err := utils.BuildAssetPath(rootFolder, asset.CollectionPath, asset.Name, asset.Extension)
		if err != nil {
			continue
		}

		fileInfo, err := os.Stat(assetFilePath)
		if os.IsNotExist(err) {
			assetsMissingFiles = append(assetsMissingFiles, asset.Id)
		} else if err == nil {
			assetsWithFiles = append(assetsWithFiles, asset.Id)
			assetFileMetadata[asset.Id] = fileMetadata{
				size:    fileInfo.Size(),
				modTime: fileInfo.ModTime().Unix(),
			}
			asset.FilePath = assetFilePath
			assetMap[asset.Id] = asset
		}
	}

	if len(assetsMissingFiles) > 0 {
		quotedIds := make([]string, len(assetsMissingFiles))
		for i, id := range assetsMissingFiles {
			quotedIds[i] = fmt.Sprintf("'%s'", id)
		}

		fetchableQuery := fmt.Sprintf(`
			SELECT DISTINCT asset_id
			FROM asset_checkpoint 
			WHERE asset_id IN (%s) AND trashed = 0
		`, strings.Join(quotedIds, ","))

		var fetchableAssetIds []struct {
			AssetId string `db:"asset_id"`
		}
		err = tx.Select(&fetchableAssetIds, fetchableQuery)
		if err != nil {
			return state, err
		}

		for _, row := range fetchableAssetIds {
			if asset, exists := assetMap[row.AssetId]; exists {
				state.FetchableAssets = append(state.FetchableAssets, asset)
			}
		}
	}

	type checkpointInfo struct {
		AssetId        string `db:"asset_id"`
		XXHashChecksum string `db:"xxhash_checksum"`
		TimeModified   int64  `db:"time_modified"`
		FileSize       int64  `db:"file_size"`
	}

	assetCheckpoints := make(map[string][]checkpointInfo)

	if len(assetsWithFiles) > 0 {
		quotedIds := make([]string, len(assetsWithFiles))
		for i, id := range assetsWithFiles {
			quotedIds[i] = fmt.Sprintf("'%s'", id)
		}

		checkpointQuery := fmt.Sprintf(`
			SELECT asset_id, xxhash_checksum, time_modified, file_size
			FROM asset_checkpoint 
			WHERE asset_id IN (%s) AND trashed = 0
			ORDER BY asset_id, created_at DESC
		`, strings.Join(quotedIds, ","))

		var checkpoints []checkpointInfo
		err = tx.Select(&checkpoints, checkpointQuery)
		if err != nil {
			return state, err
		}

		for _, cp := range checkpoints {
			assetCheckpoints[cp.AssetId] = append(assetCheckpoints[cp.AssetId], cp)
		}
	}

	type hashCandidate struct {
		assetId       string
		assetFilePath string
		checkpoints   []checkpointInfo
	}
	candidatesNeedingHash := make([]hashCandidate, 0)
	assetsWithMatchingMetadata := make([]string, 0)

	for assetId, metadata := range assetFileMetadata {
		checkpoints, hasCheckpoints := assetCheckpoints[assetId]
		if !hasCheckpoints || len(checkpoints) == 0 {
			continue
		}

		latestCheckpoint := checkpoints[0]

		if metadata.size == latestCheckpoint.FileSize &&
			metadata.modTime == latestCheckpoint.TimeModified {
			assetsWithMatchingMetadata = append(assetsWithMatchingMetadata, assetId)
			continue
		}

		asset := assetMap[assetId]
		candidatesNeedingHash = append(candidatesNeedingHash, hashCandidate{
			assetId:       assetId,
			assetFilePath: asset.FilePath,
			checkpoints:   checkpoints,
		})
	}

	for _, candidate := range candidatesNeedingHash {
		fileHash, err := utils.GenerateXXHashChecksum(candidate.assetFilePath)
		if err != nil {
			continue
		}

		matchesLatest := (fileHash == candidate.checkpoints[0].XXHashChecksum)

		if matchesLatest {
			asset := assetMap[candidate.assetId]
			state.NormalAssets = append(state.NormalAssets, asset)
			continue
		}

		matchesOlderCheckpoint := false
		for i := 1; i < len(candidate.checkpoints); i++ {
			if fileHash == candidate.checkpoints[i].XXHashChecksum {
				matchesOlderCheckpoint = true
				break
			}
		}

		asset := assetMap[candidate.assetId]

		if matchesOlderCheckpoint {
			state.OutdatedAssets = append(state.OutdatedAssets, asset)
		} else {
			state.ModifiedAssets = append(state.ModifiedAssets, asset)
		}
	}

	for _, assetId := range assetsWithMatchingMetadata {
		if asset, exists := assetMap[assetId]; exists {
			state.NormalAssets = append(state.NormalAssets, asset)
		}
	}

	canModifyContext, err := e.buildUntrackedCanModifyContext(tx, false)
	if err != nil {
		return state, err
	}
	return e.detectUntrackedItems(tx, state, collectionId, projectWorkingDir, rootFolder, ignoreList, canModifyContext)
}

// detectUntrackedItems scans the filesystem for untracked files and folders.
// Builds maps of tracked names and compares against filesystem entries.
func (e *CollectionService) detectUntrackedItems(tx *sqlx.Tx, state CollectionChildrenState, collectionId, projectWorkingDir, rootFolder string, ignoreList []string, canModifyContext untrackedCanModifyContext) (CollectionChildrenState, error) {
	trackedAssetNames := make(map[string]bool)
	trackedCollectionNames := make(map[string]bool)

	var trackedAssets []models.Asset
	assetQuery := `
		SELECT name, extension, local_path
		FROM full_asset 
		WHERE collection_id = ? AND trashed = 0
	`
	err := tx.Select(&trackedAssets, assetQuery, collectionId)
	if err != nil {
		return state, err
	}

	for _, asset := range trackedAssets {
		if asset.LocalPath != "" {
			trackedAssetNames[filepath.Base(asset.LocalPath)] = true
			continue
		}
		trackedAssetNames[asset.Name+asset.Extension] = true
	}

	var trackedCollections []models.Collection
	collectionQuery := `
		SELECT name, local_path
		FROM full_collection 
		WHERE parent_id = ? AND trashed = 0
	`
	err = tx.Select(&trackedCollections, collectionQuery, collectionId)
	if err != nil {
		return state, err
	}

	for _, collection := range trackedCollections {
		if collection.LocalPath != "" {
			trackedCollectionNames[filepath.Base(filepath.Clean(collection.LocalPath))] = true
			continue
		}
		trackedCollectionNames[collection.Name] = true
	}

	var folderToScan string

	if collectionId == "" {
		folderToScan = projectWorkingDir
	} else {
		collection, err := repository.GetCollection(tx, collectionId)
		if err != nil {
			return state, err
		}
		folderToScan = collection.GetFilePath()
	}
	pendingFolders, pendingFiles, err := pendingNamesInFolder(tx, folderToScan)
	if err != nil {
		return state, err
	}
	for _, name := range pendingFolders {
		trackedCollectionNames[name] = true
	}
	for _, name := range pendingFiles {
		trackedAssetNames[name] = true
	}

	if !utils.DirExists(folderToScan) {
		return state, nil
	}

	absoluteCollectionFolderPath, err := filepath.Abs(folderToScan)
	if err != nil {
		return state, err
	}

	relativeCollectionFolderPath, err := filepath.Rel(projectWorkingDir, absoluteCollectionFolderPath)
	if err != nil {
		return state, err
	}
	relativeCollectionFolderPath = utils.NormalizePath(relativeCollectionFolderPath)

	clusttaIgnore := ignore.CompileIgnoreLines(ignoreList...)

	entries, err := os.ReadDir(absoluteCollectionFolderPath)
	if err != nil {
		return state, err
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		entryPath := filepath.Join(absoluteCollectionFolderPath, entry.Name())
		relativePath := utils.NormalizePath(filepath.Join(relativeCollectionFolderPath, entry.Name()))

		if entry.IsDir() {
			if trackedCollectionNames[entry.Name()] {
				continue
			}

			if !clusttaIgnore.MatchesPath(relativePath) {
				untrackedFolder := models.UntrackedCollection{
					Id:             utils.GetMD5Hash(entryPath),
					Name:           entry.Name(),
					FilePath:       entryPath,
					CollectionPath: "/" + relativePath + "/",
					ItemPath:       "/" + relativePath + "/",
					ParentId:       collectionId,
				}
				untrackedFolder.CanModify = canModifyContext.canModifyPath(parentLookupPath(untrackedFolder.ItemPath), untrackedFolder.ParentId)
				state.UntrackedFolders = append(state.UntrackedFolders, untrackedFolder)
			}
		} else {
			if trackedAssetNames[entry.Name()] {
				continue
			}

			if !clusttaIgnore.MatchesPath(relativePath) {
				assetName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
				untrackedFile := models.UntrackedAsset{
					Id:             utils.GetMD5Hash(entryPath),
					Name:           assetName,
					FilePath:       entryPath,
					AssetPath:      "/" + relativePath,
					CollectionId:   collectionId,
					CollectionPath: "/" + relativeCollectionFolderPath + "/",
					Extension:      filepath.Ext(entry.Name()),
					ItemPath:       "/" + relativePath + "/",
					AssetTypeIcon:  "generic",
				}
				untrackedFile.CanModify = canModifyContext.canModifyPath(untrackedFile.CollectionPath, untrackedFile.CollectionId)
				state.UntrackedFiles = append(state.UntrackedFiles, untrackedFile)
			}
		}
	}

	return state, nil
}

// GetItemsForCheckpoint efficiently collects all modified and untracked items in a collection hierarchy.
// Returns deduplicated modified assets and untracked files.
func (e *CollectionService) GetItemsForCheckpoint(projectPath, collectionId, targetPath, projectWorkingDir string, ignoreList []string) (ItemsForCheckpoint, error) {
	result := ItemsForCheckpoint{
		ModifiedAssets: make([]models.Asset, 0),
		UntrackedFiles: make([]models.UntrackedAsset, 0),
	}

	if collectionId == "root" {
		collectionId = ""
	}

	isTrackedCollection := collectionId != ""
	isUntrackedPath := targetPath != "" && !isTrackedCollection

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return result, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return result, err
	}
	defer tx.Rollback()

	modifiedAssetsMap := make(map[string]models.Asset)
	untrackedFilesMap := make(map[string]models.UntrackedAsset)
	canModifyContext, err := e.buildUntrackedCanModifyContext(tx, true)
	if err != nil {
		return result, err
	}

	if isTrackedCollection {
		err = e.processTrackedCollection(tx, collectionId, projectPath, projectWorkingDir, ignoreList, modifiedAssetsMap, untrackedFilesMap, canModifyContext)
		if err != nil {
			return result, err
		}
	} else if isUntrackedPath {
		err = e.processUntrackedPath(targetPath, projectWorkingDir, ignoreList, untrackedFilesMap, canModifyContext)
		if err != nil {
			return result, err
		}
	} else {
		err = e.processTrackedCollection(tx, "", projectPath, projectWorkingDir, ignoreList, modifiedAssetsMap, untrackedFilesMap, canModifyContext)
		if err != nil {
			return result, err
		}
	}

	for _, asset := range modifiedAssetsMap {
		result.ModifiedAssets = append(result.ModifiedAssets, asset)
	}

	for _, file := range untrackedFilesMap {
		result.UntrackedFiles = append(result.UntrackedFiles, file)
	}

	return result, nil
}

// processTrackedCollection recursively scans tracked collections for modified assets and untracked files.
// Uses flag-based optimization to avoid scanning clean collection branches.
func (e *CollectionService) processTrackedCollection(tx *sqlx.Tx, collectionId, projectPath, projectWorkingDir string, ignoreList []string, modifiedAssetsMap map[string]models.Asset, untrackedFilesMap map[string]models.UntrackedAsset, canModifyContext untrackedCanModifyContext) error {
	var processCollection func(string) error
	processCollection = func(currentCollectionId string) error {
		childrenState, err := e.GetCollectionChildrenState(projectPath, currentCollectionId, projectWorkingDir, ignoreList)
		if err != nil {
			return err
		}

		for _, asset := range childrenState.ModifiedAssets {
			modifiedAssetsMap[asset.Id] = asset
		}

		for _, file := range childrenState.UntrackedFiles {
			untrackedFilesMap[file.Id] = file
		}

		var childCollections []models.Collection
		childQuery := `
			SELECT id, name, collection_path
			FROM full_collection 
			WHERE parent_id = ? AND trashed = 0
			ORDER BY name
		`
		err = tx.Select(&childCollections, childQuery, currentCollectionId)
		if err != nil {
			return err
		}

		for _, childCollection := range childCollections {
			flags, err := e.GetCollectionStateFlags(projectPath, childCollection.Id, projectWorkingDir, ignoreList)
			if err != nil {
				continue
			}

			if flags.HasModified || flags.HasUntracked {
				err = processCollection(childCollection.Id)
				if err != nil {
					continue
				}
			}
		}

		for _, untrackedFolder := range childrenState.UntrackedFolders {
			err = e.processUntrackedPath(untrackedFolder.FilePath, projectWorkingDir, ignoreList, untrackedFilesMap, canModifyContext)
			if err != nil {
				continue
			}
		}

		return nil
	}

	return processCollection(collectionId)
}

// processUntrackedPath recursively scans an untracked filesystem location for files.
// Performs pure filesystem scanning without database queries.
func (e *CollectionService) processUntrackedPath(targetPath, projectWorkingDir string, ignoreList []string, untrackedFilesMap map[string]models.UntrackedAsset, canModifyContext untrackedCanModifyContext) error {
	absolutePath, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}

	if !utils.DirExists(absolutePath) {
		return fmt.Errorf("target path does not exist: %s", targetPath)
	}

	clusttaIgnore := ignore.CompileIgnoreLines(ignoreList...)

	var scanDirectory func(string) error
	scanDirectory = func(currentPath string) error {
		relativePath, err := filepath.Rel(projectWorkingDir, currentPath)
		if err != nil {
			return err
		}
		relativePath = utils.NormalizePath(relativePath)

		entries, err := os.ReadDir(currentPath)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			entryPath := filepath.Join(currentPath, entry.Name())
			relPath, err := filepath.Rel(projectWorkingDir, entryPath)
			if err != nil {
				continue
			}
			relPath = utils.NormalizePath(relPath)

			if clusttaIgnore.MatchesPath(relPath) {
				continue
			}

			if entry.IsDir() {
				err = scanDirectory(entryPath)
				if err != nil {
					continue
				}
			} else {
				assetName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
				untrackedFile := models.UntrackedAsset{
					Id:             utils.GetMD5Hash(entryPath),
					Name:           assetName,
					FilePath:       entryPath,
					AssetPath:      "/" + relPath,
					CollectionId:   "",
					CollectionPath: "/" + relativePath + "/",
					Extension:      filepath.Ext(entry.Name()),
					ItemPath:       "/" + relPath,
					AssetTypeIcon:  "generic",
				}
				untrackedFile.CanModify = canModifyContext.canModifyPath(untrackedFile.CollectionPath, untrackedFile.CollectionId)
				untrackedFilesMap[untrackedFile.Id] = untrackedFile
			}
		}

		return nil
	}

	return scanDirectory(absolutePath)
}

// GetOutdatedItemsInCollection efficiently collects all outdated items in a collection hierarchy.
// Returns deduplicated outdated assets.
func (e *CollectionService) GetOutdatedItemsInCollection(projectPath, collectionId, projectWorkingDir string, ignoreList []string) (ItemsForUpdate, error) {
	result := ItemsForUpdate{
		OutdatedAssets: make([]models.Asset, 0),
	}

	if collectionId == "root" {
		collectionId = ""
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return result, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return result, err
	}
	defer tx.Rollback()

	outdatedAssetsMap := make(map[string]models.Asset)

	err = e.processTrackedCollectionForOutdated(tx, collectionId, projectPath, projectWorkingDir, ignoreList, outdatedAssetsMap)
	if err != nil {
		return result, err
	}

	for _, asset := range outdatedAssetsMap {
		result.OutdatedAssets = append(result.OutdatedAssets, asset)
	}

	return result, nil
}

// processTrackedCollectionForOutdated recursively scans tracked collections for outdated assets.
// Uses flag-based optimization to avoid scanning clean collection branches.
func (e *CollectionService) processTrackedCollectionForOutdated(tx *sqlx.Tx, collectionId, projectPath, projectWorkingDir string, ignoreList []string, outdatedAssetsMap map[string]models.Asset) error {
	var processCollection func(string) error
	processCollection = func(currentCollectionId string) error {
		childrenState, err := e.GetCollectionChildrenState(projectPath, currentCollectionId, projectWorkingDir, ignoreList)
		if err != nil {
			return err
		}

		for _, asset := range childrenState.OutdatedAssets {
			outdatedAssetsMap[asset.Id] = asset
		}

		var childCollections []models.Collection
		childQuery := `
			SELECT id, name, collection_path
			FROM full_collection 
			WHERE parent_id = ? AND trashed = 0
			ORDER BY name
		`
		err = tx.Select(&childCollections, childQuery, currentCollectionId)
		if err != nil {
			return err
		}

		for _, childCollection := range childCollections {
			flags, err := e.GetCollectionStateFlags(projectPath, childCollection.Id, projectWorkingDir, ignoreList)
			if err != nil {
				return err
			}

			if flags.HasOutdated {
				err = processCollection(childCollection.Id)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	return processCollection(collectionId)
}

// Fetch restores missing working files for specified collections, downloading
// checkpoint chunks first when they are not available locally.
// Supports cancellation and sends progress updates via application events.
func (e *CollectionService) Fetch(projectPath, remoteUrl, collectionIds, userId string) (FetchResult, error) {
	result := FetchResult{RestoredAssetIds: make([]string, 0)}
	defer reset()

	ctx := getContext()
	if ctx.Err() != nil {
		return result, errors.New("operation cancelled before starting")
	}

	app := application.Get()
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return result, err
	}

	errChan := make(chan error, 1)
	progressChan := make(chan output.ProgressReport, 10)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case progress, ok := <-progressChan:
				if !ok {
					return
				}
				app.Event.Emit("progress-update", progress)
			}
		}
	}()

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return result, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return result, err
	}
	defer tx.Rollback()

	select {
	case <-ctx.Done():
		return result, errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      "Fetching",
		Message:    "Preparing to fetch files",
		Percentage: 0,
		Current:    1,
		Total:      2,
	}:
	}

	var collectionIdList []string
	if collectionIds == "" {
		collectionIdList = []string{""}
	} else {
		collectionIdList = strings.Split(collectionIds, ",")
		for i, id := range collectionIdList {
			collectionIdList[i] = strings.TrimSpace(id)
		}
	}

	collectionCollectionsQuery := `
	SELECT full_collection.*
	FROM full_collection
	WHERE full_collection.collection_path LIKE ? OR full_collection.collection_path LIKE ?;
	`

	collections := []models.Collection{}
	allAssets := []models.Asset{}

	for _, collectionId := range collectionIdList {
		if collectionId == "" {
			rootCollections, err := repository.GetCollections(tx, false)
			if err != nil {
				return result, err
			}
			collections = append(collections, rootCollections...)

			rootAssets, err := repository.GetAssets(tx, false)
			if err != nil {
				return result, err
			}
			allAssets = append(allAssets, rootAssets...)
		} else {
			parentCollection, err := repository.GetCollection(tx, collectionId)
			if err != nil {
				return result, err
			}
			err = os.MkdirAll(parentCollection.FilePath, os.ModePerm)
			if err != nil {
				return result, err
			}
			pathLike := parentCollection.CollectionPath + "%"
			var collectionChildren []models.Collection
			err = tx.Select(&collectionChildren, collectionCollectionsQuery, parentCollection.CollectionPath, pathLike)
			if err != nil {
				return result, err
			}
			collections = append(collections, collectionChildren...)

			collectionAssetsQuery := `
			SELECT full_asset.*
			FROM full_asset
			WHERE (full_asset.collection_path LIKE ? OR full_asset.collection_path LIKE ?) AND full_asset.trashed = 0;
			`

			var collectionAssets []models.Asset
			err = tx.Select(&collectionAssets, collectionAssetsQuery, parentCollection.CollectionPath, pathLike)
			if err != nil {
				return result, err
			}
			allAssets = append(allAssets, collectionAssets...)
		}
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return result, err
	}

	for _, collection := range collections {
		collectionPath, err := utils.BuildCollectionPath(rootFolder, collection.CollectionPath)
		if err != nil {
			return result, err
		}
		err = os.MkdirAll(collectionPath, os.ModePerm)
		if err != nil {
			return result, err
		}
	}

	assetIds := []string{}
	for _, asset := range allAssets {
		assetIds = append(assetIds, asset.Id)
	}

	if len(assetIds) > 0 {
		quotedAssetIds := make([]string, len(assetIds))
		for i, id := range assetIds {
			quotedAssetIds[i] = fmt.Sprintf("\"%s\"", id)
		}

		checkpoints := []models.Checkpoint{}
		err = tx.Select(&checkpoints, fmt.Sprintf("SELECT * FROM asset_checkpoint WHERE trashed = 0 AND asset_id IN (%s) ORDER BY created_at DESC", strings.Join(quotedAssetIds, ",")))
		if err != nil {
			return result, err
		}

		assetCheckpoints := map[string][]models.Checkpoint{}
		for _, assetCheckpoint := range checkpoints {
			assetCheckpoints[assetCheckpoint.AssetId] = append(assetCheckpoints[assetCheckpoint.AssetId], assetCheckpoint)
		}

		for i, asset := range allAssets {
			allAssets[i].Checkpoints = assetCheckpoints[asset.Id]
		}
	}

	assetsToFetch := []models.Asset{}
	for _, asset := range allAssets {
		assetFilePath, err := utils.BuildAssetPath(rootFolder, asset.CollectionPath, asset.Name, asset.Extension)
		if err != nil {
			return result, err
		}
		asset.FilePath = assetFilePath
		if len(asset.Checkpoints) == 0 {
			continue
		}
		if _, err := os.Stat(asset.GetFilePath()); os.IsNotExist(err) {
			assetsToFetch = append(assetsToFetch, asset)
		}
	}

	checkpointIdsToDownload := []string{}
	for _, asset := range assetsToFetch {
		latestCheckpoint := asset.Checkpoints[0]
		isMisssingChunks, err := latestCheckpoint.HasMissingChunks(tx)
		if err != nil {
			return result, err
		}
		if isMisssingChunks {
			checkpointIdsToDownload = append(checkpointIdsToDownload, latestCheckpoint.Id)
		}
	}

	err = tx.Rollback()
	if err != nil {
		return result, err
	}

	if len(checkpointIdsToDownload) != 0 {
		callBack := func(current int, total int, message string, extraMessage string) {
			if ctx.Err() != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			case progressChan <- output.ProgressReport{
				Title:        "Downloading files",
				Message:      message,
				Percentage:   (float64(current) / float64(total) * 99),
				Current:      1,
				Total:        1,
				ExtraMessage: extraMessage,
			}:
			default:
			}
		}

		go func() {
			err := sync_service.DownloadCheckpoints(ctx, projectPath, remoteUrl, checkpointIdsToDownload, user.Id, callBack)
			if ctx.Err() == nil {
				errChan <- err
			}
		}()

		select {
		case err = <-errChan:
			if err != nil {
				if errors.Is(err, syscall.ECONNREFUSED) {
					return result, errors.New("download failed, connection refused")
				}
				return result, errors.New("download failed, check your connection")
			}
		case <-ctx.Done():
			close(progressChan) // Stop progress updates
			return result, errors.New("cancelled")
		}
	}

	if ctx.Err() != nil {
		return result, ctx.Err()
	}

	tx, err = dbConn.Beginx()
	if err != nil {
		return result, err
	}
	defer tx.Rollback()

	totalItems := len(assetsToFetch)
	for i, asset := range assetsToFetch {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}

		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Restoring files",
				Message:    asset.Name,
				Percentage: float64(current) / float64(total) * 100,
				Current:    i + 1,
				Total:      totalItems,
			}
			app.Event.Emit("progress-update", progress)
		}
		err = repository.RevertToLatestCheckpoint(tx, asset.Id, asset.FilePath, callBack)
		if err != nil {
			return result, err
		}
		result.RestoredAssetIds = append(result.RestoredAssetIds, asset.Id)
	}
	if err = tx.Rollback(); err != nil {
		return result, err
	}
	if err = clearChunkCacheIfEnabled(projectPath, dbConn); err != nil {
		return result, err
	}

	close(progressChan)
	progress := output.ProgressReport{
		Title:      "Downloading Checkpoint",
		Message:    "Receiving",
		Percentage: 100,
		Current:    1,
		Total:      1,
	}
	app.Event.Emit("progress-update", progress)
	return result, nil
}

// RevealCollection opens the file explorer to show a collection's folder.
// Returns an error if the collection is not found or the operation fails.
func (e *CollectionService) RevealCollection(projectPath, collectionId string) error {
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

	collection, err := repository.GetCollection(tx, collectionId)
	if err != nil {
		return err
	}
	utils.RevealInExplorer(collection.GetFilePath())
	return nil
}

// RevertCollections reverts multiple collections to their latest checkpoints.
// Sends progress updates for each collection processed.
func (e *CollectionService) RevertCollections(projectPath string, collectionIds []string) error {
	app := application.Get()
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

	totalCollections := len(collectionIds)
	for i, collectionId := range collectionIds {
		tx, err := dbConn.Beginx()
		if err != nil {
			return err
		}
		collection, err := repository.GetCollection(tx, collectionId)
		if err != nil {
			return err
		}
		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Reverting",
				Message:    collection.CollectionTypeName,
				Percentage: float64(current) / float64(total) * 100,
				Current:    i + 1,
				Total:      totalCollections,
			}
			app.Event.Emit("progress-update", progress)
		}

		err = repository.RevertToLatestCheckpoint(tx, collectionId, collection.FilePath, callBack)
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
	}
	return nil
}

// ChangeCollectionParent moves one or more collections to a different parent collection.
// Checks for name conflicts in the target parent before moving.
// Returns an error if any collection would conflict or if the operation fails.
func (e *CollectionService) ChangeCollectionParent(projectPath string, collectionIds []string, parentId string) error {
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

	var conflicts []string
	for _, collectionId := range collectionIds {
		collection, err := repository.GetCollection(tx, collectionId)
		if err != nil {
			return err
		}
		if collection.ParentId == parentId {
			continue
		}
		_, err = repository.GetCollectionByName(tx, collection.Name, parentId)
		if err == nil {
			conflicts = append(conflicts, collection.Name)
		} else if err != error_service.ErrCollectionNotFound {
			// Some other error occurred
			return err
		}
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("collections with the same name already exist in the target location: %s", strings.Join(conflicts, ", "))
	}

	for _, collectionId := range collectionIds {
		err = repository.ChangeParent(tx, collectionId, parentId)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// ChangeType changes the type of a collection.
// Returns an error if the operation fails.
func (e *CollectionService) ChangeType(projectPath, collectionId, collectionTypeId string) (MetadataUpdateResult, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	tx, err := dbConn.Beginx()
	if err != nil {
		dbConn.Close()
		return MetadataUpdateResult{}, err
	}
	if err = authorizeUpdateCollectionTx(tx, collectionId); err != nil {
		tx.Rollback()
		dbConn.Close()
		return MetadataUpdateResult{}, err
	}
	collectionType, err := repository.GetCollectionType(tx, collectionTypeId)
	if err != nil {
		tx.Rollback()
		dbConn.Close()
		return MetadataUpdateResult{}, err
	}
	tx.Rollback()
	dbConn.Close()

	remoteURL, err := utils.ResolveProjectRemoteURL(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, fmt.Errorf("failed to resolve project remote: %w", err)
	}
	requiresSync := remoteURL != "" && !collectionType.Synced
	remoteApplied := false
	if remoteURL != "" && collectionType.Synced {
		response, patchErr := patchCollectionsRemote(remoteURL, []collectionPatch{{Id: collectionId, CollectionTypeId: &collectionTypeId}})
		if patchErr != nil {
			fallback, fallbackErr := metadataMutationAllowsLocalFallback(projectPath, metadataTableCollection, []string{collectionId}, patchErr)
			if fallbackErr != nil {
				return MetadataUpdateResult{}, fallbackErr
			}
			if !fallback {
				return MetadataUpdateResult{}, patchErr
			}
			requiresSync = true
		} else {
			for _, collection := range response.Collections {
				if collection.Id == collectionId && collection.CollectionTypeId == collectionTypeId {
					remoteApplied = true
					db, openErr := utils.OpenDb(projectPath)
					if openErr != nil {
						return MetadataUpdateResult{}, openErr
					}
					defer db.Close()
					canonicalTx, beginErr := db.Beginx()
					if beginErr != nil {
						return MetadataUpdateResult{}, beginErr
					}
					defer canonicalTx.Rollback()
					clean, applyErr := applyCanonicalCollections(canonicalTx, response.Collections)
					if applyErr != nil {
						return MetadataUpdateResult{}, applyErr
					}
					if !clean {
						canonicalTx.Rollback()
						requiresSync = true
						break
					}
					if err = applyReturnedSyncToken(canonicalTx, response.PreviousSyncToken, response.SyncToken); err != nil {
						return MetadataUpdateResult{}, err
					}
					if err = canonicalTx.Commit(); err != nil {
						return MetadataUpdateResult{}, err
					}
					return MetadataUpdateResult{RemoteApplied: true}, nil
				}
			}
			requiresSync = true
		}
	}

	dbConn, err = utils.OpenDb(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	defer dbConn.Close()
	tx, err = dbConn.Beginx()
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	defer tx.Rollback()
	if err = repository.ChangeCollectionType(tx, collectionId, collectionTypeId); err != nil {
		return MetadataUpdateResult{}, err
	}
	if err = tx.Commit(); err != nil {
		return MetadataUpdateResult{}, err
	}
	return MetadataUpdateResult{RemoteApplied: remoteApplied, RequiresSync: requiresSync}, nil
}

// ChangeIsShared toggles the shared flag on a collection.
// Returns an error if the operation fails.
func (e *CollectionService) ChangeIsShared(projectPath, collectionId string, isShared bool) (MetadataUpdateResult, error) {
	remoteURL, err := utils.ResolveProjectRemoteURL(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, fmt.Errorf("failed to resolve project remote: %w", err)
	}
	var remoteFailure error
	remoteApplied := false
	if remoteURL != "" {
		response, err := patchCollectionsRemote(remoteURL, []collectionPatch{{Id: collectionId, IsShared: &isShared}})
		if err != nil {
			fallback, fallbackErr := metadataMutationAllowsLocalFallback(projectPath, metadataTableCollection, []string{collectionId}, err)
			if fallbackErr != nil {
				return MetadataUpdateResult{}, fallbackErr
			}
			if !fallback {
				return MetadataUpdateResult{}, err
			}
			remoteFailure = err
		} else {
			remoteApplied = true
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
			clean, applyErr := applyCanonicalCollections(tx, response.Collections)
			if applyErr != nil {
				return MetadataUpdateResult{}, applyErr
			}
			if clean {
				if err = applyReturnedSyncToken(tx, response.PreviousSyncToken, response.SyncToken); err != nil {
					return MetadataUpdateResult{}, err
				}
				if err = tx.Commit(); err != nil {
					return MetadataUpdateResult{}, err
				}
				return MetadataUpdateResult{RemoteApplied: true}, nil
			}
			tx.Rollback()
			remoteFailure = errors.New("patched collection has pre-existing unsynced changes")
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

	err = repository.ChangeIsShared(tx, collectionId, isShared)
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	err = tx.Commit()
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	return MetadataUpdateResult{RemoteApplied: remoteApplied, RequiresSync: remoteFailure != nil}, nil
}

// Assign assigns a user to a collection.
// Returns an error if the operation fails.
func (e *CollectionService) Assign(projectPath, collectionId, userId string) (MetadataUpdateResult, error) {
	remoteURL, err := utils.ResolveProjectRemoteURL(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, fmt.Errorf("failed to resolve project remote: %w", err)
	}
	var remoteFailure error
	remoteApplied := false
	if remoteURL != "" {
		response, err := patchCollectionsRemote(remoteURL, []collectionPatch{{Id: collectionId, AddAssigneeIds: []string{userId}}})
		if err != nil {
			fallback, fallbackErr := metadataMutationAllowsLocalFallback(projectPath, metadataTableCollection, []string{collectionId}, err)
			if fallbackErr != nil {
				return MetadataUpdateResult{}, fallbackErr
			}
			if !fallback {
				return MetadataUpdateResult{}, err
			}
			remoteFailure = err
		} else {
			remoteApplied = true
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
			clean, applyErr := applyCanonicalCollections(tx, response.Collections)
			if applyErr != nil {
				return MetadataUpdateResult{}, applyErr
			}
			if clean {
				patchedAssignees := make([]models.CollectionAssignee, 0, 1)
				for _, assignee := range response.CollectionAssignees {
					if assignee.CollectionId == collectionId && assignee.AssigneeId == userId {
						patchedAssignees = append(patchedAssignees, assignee)
					}
				}
				if err = applyCanonicalCollectionAssignees(tx, patchedAssignees); err != nil {
					return MetadataUpdateResult{}, err
				}
				if err = applyReturnedSyncToken(tx, response.PreviousSyncToken, response.SyncToken); err != nil {
					return MetadataUpdateResult{}, err
				}
				if err = tx.Commit(); err != nil {
					return MetadataUpdateResult{}, err
				}
				return MetadataUpdateResult{RemoteApplied: true}, nil
			}
			tx.Rollback()
			remoteFailure = errors.New("patched collection has pre-existing unsynced changes")
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

	err = repository.AssignCollection(tx, collectionId, userId)
	if err != nil {
		tx.Rollback()
		return MetadataUpdateResult{}, err
	}
	err = tx.Commit()
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	return MetadataUpdateResult{RemoteApplied: remoteApplied, RequiresSync: remoteFailure != nil}, nil
}

// Unassign removes a user assignment from a collection.
// Returns an error if the operation fails.
func (e *CollectionService) Unassign(projectPath, collectionId, userId string) (MetadataUpdateResult, error) {
	remoteURL, err := utils.ResolveProjectRemoteURL(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, fmt.Errorf("failed to resolve project remote: %w", err)
	}
	var remoteFailure error
	remoteApplied := false
	if remoteURL != "" {
		response, err := patchCollectionsRemote(remoteURL, []collectionPatch{{Id: collectionId, RemoveAssigneeIds: []string{userId}}})
		if err != nil {
			fallback, fallbackErr := metadataMutationAllowsLocalFallback(projectPath, metadataTableCollection, []string{collectionId}, err)
			if fallbackErr != nil {
				return MetadataUpdateResult{}, fallbackErr
			}
			if !fallback {
				return MetadataUpdateResult{}, err
			}
			remoteFailure = err
		} else {
			remoteApplied = true
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
			clean, applyErr := applyCanonicalCollections(tx, response.Collections)
			if applyErr != nil {
				return MetadataUpdateResult{}, applyErr
			}
			if clean {
				var id string
				_ = tx.Get(&id, "SELECT id FROM collection_assignee WHERE collection_id=? AND assignee_id=?", collectionId, userId)
				if _, err = tx.Exec("DELETE FROM collection_assignee WHERE collection_id=? AND assignee_id=?", collectionId, userId); err != nil {
					return MetadataUpdateResult{}, err
				}
				if id != "" {
					if _, err = tx.Exec("UPDATE tomb SET synced=1 WHERE id=? AND table_name='collection_assignee'", id); err != nil {
						return MetadataUpdateResult{}, err
					}
				}
				if err = applyReturnedSyncToken(tx, response.PreviousSyncToken, response.SyncToken); err != nil {
					return MetadataUpdateResult{}, err
				}
				if err = tx.Commit(); err != nil {
					return MetadataUpdateResult{}, err
				}
				return MetadataUpdateResult{RemoteApplied: true}, nil
			}
			tx.Rollback()
			remoteFailure = errors.New("patched collection has pre-existing unsynced changes")
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

	err = repository.UnAssignCollection(tx, collectionId, userId)
	if err != nil {
		tx.Rollback()
		return MetadataUpdateResult{}, err
	}
	err = tx.Commit()
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	return MetadataUpdateResult{RemoteApplied: remoteApplied, RequiresSync: remoteFailure != nil}, nil
}

// UnassignCollections removes every direct assignee from the supplied collections atomically.
func (e *CollectionService) UnassignCollections(projectPath string, collectionIds []string) (MetadataUpdateResult, error) {
	if len(collectionIds) == 0 {
		return MetadataUpdateResult{}, nil
	}
	db, err := utils.OpenDb(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	tx, err := db.Beginx()
	if err != nil {
		db.Close()
		return MetadataUpdateResult{}, err
	}
	patches := make([]collectionPatch, 0, len(collectionIds))
	for _, collectionId := range collectionIds {
		var ids []string
		if err = tx.Select(&ids, "SELECT assignee_id FROM collection_assignee WHERE collection_id=?", collectionId); err != nil {
			tx.Rollback()
			db.Close()
			return MetadataUpdateResult{}, err
		}
		patches = append(patches, collectionPatch{Id: collectionId, RemoveAssigneeIds: ids})
	}
	tx.Rollback()
	db.Close()
	remoteURL, err := utils.ResolveProjectRemoteURL(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, fmt.Errorf("failed to resolve project remote: %w", err)
	}
	var remoteFailure error
	remoteApplied := false
	if remoteURL != "" {
		response, remoteErr := patchCollectionsRemote(remoteURL, patches)
		if remoteErr != nil {
			fallback, fallbackErr := metadataMutationAllowsLocalFallback(projectPath, metadataTableCollection, collectionIds, remoteErr)
			if fallbackErr != nil {
				return MetadataUpdateResult{}, fallbackErr
			}
			if !fallback {
				return MetadataUpdateResult{}, remoteErr
			}
			remoteFailure = remoteErr
		} else {
			remoteApplied = true
			db, err = utils.OpenDb(projectPath)
			if err != nil {
				return MetadataUpdateResult{}, err
			}
			defer db.Close()
			tx, err = db.Beginx()
			if err != nil {
				return MetadataUpdateResult{}, err
			}
			defer tx.Rollback()
			clean, applyErr := applyCanonicalCollections(tx, response.Collections)
			if applyErr != nil {
				return MetadataUpdateResult{}, applyErr
			}
			if clean {
				for _, patch := range patches {
					for _, userId := range patch.RemoveAssigneeIds {
						var id string
						_ = tx.Get(&id, "SELECT id FROM collection_assignee WHERE collection_id=? AND assignee_id=?", patch.Id, userId)
						if _, err = tx.Exec("DELETE FROM collection_assignee WHERE collection_id=? AND assignee_id=?", patch.Id, userId); err != nil {
							return MetadataUpdateResult{}, err
						}
						if id != "" {
							if _, err = tx.Exec("UPDATE tomb SET synced=1 WHERE id=? AND table_name='collection_assignee'", id); err != nil {
								return MetadataUpdateResult{}, err
							}
						}
					}
				}
				if err = applyReturnedSyncToken(tx, response.PreviousSyncToken, response.SyncToken); err != nil {
					return MetadataUpdateResult{}, err
				}
				if err = tx.Commit(); err != nil {
					return MetadataUpdateResult{}, err
				}
				return MetadataUpdateResult{RemoteApplied: true}, nil
			}
			tx.Rollback()
			remoteFailure = errors.New("patched collection has pre-existing unsynced changes")
		}
	}

	db, err = utils.OpenDb(projectPath)
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	defer db.Close()
	tx, err = db.Beginx()
	if err != nil {
		return MetadataUpdateResult{}, err
	}
	defer tx.Rollback()
	for _, patch := range patches {
		for _, userId := range patch.RemoveAssigneeIds {
			if err = repository.UnAssignCollection(tx, patch.Id, userId); err != nil {
				return MetadataUpdateResult{}, err
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return MetadataUpdateResult{}, err
	}
	return MetadataUpdateResult{RemoteApplied: remoteApplied, RequiresSync: remoteFailure != nil}, nil
}

// CreateCollectionType creates a new collection type in the project.
// Returns an error if a type with the same name already exists.
func (e *CollectionService) CreateCollectionType(projectPath, collectionTypeName, collectionTypeIcon string) (models.CollectionType, error) {
	id := uuid.New().String()
	remoteURL, err := utils.ResolveProjectRemoteURL(projectPath)
	if err != nil {
		return models.CollectionType{}, fmt.Errorf("failed to resolve project remote: %w", err)
	}
	if remoteURL != "" {
		response, remoteErr := putCollectionTypeRemote(remoteURL, id, collectionTypeName, collectionTypeIcon)
		if remoteErr == nil {
			if response.CollectionType.Id != id {
				return models.CollectionType{}, errors.New("remote collection type response did not contain the requested type")
			}
			dbConn, openErr := utils.OpenDb(projectPath)
			if openErr != nil {
				return models.CollectionType{}, openErr
			}
			defer dbConn.Close()
			tx, beginErr := dbConn.Beginx()
			if beginErr != nil {
				return models.CollectionType{}, beginErr
			}
			defer tx.Rollback()
			if err = applyCanonicalCollectionType(tx, response.CollectionType); err != nil {
				return models.CollectionType{}, err
			}
			if err = applyReturnedSyncToken(tx, response.PreviousSyncToken, response.SyncToken); err != nil {
				return models.CollectionType{}, err
			}
			if err = tx.Commit(); err != nil {
				return models.CollectionType{}, err
			}
			response.CollectionType.Synced = true
			return response.CollectionType, nil
		}
		if !IsMetadataTransportFailure(remoteErr) {
			return models.CollectionType{}, remoteErr
		}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.CollectionType{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.CollectionType{}, err
	}
	defer tx.Rollback()

	collectionType, err := repository.CreateCollectionType(tx, id, collectionTypeName, collectionTypeIcon)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: collection_type.name" {
			tx.Rollback()
			return models.CollectionType{}, error_service.ErrCollectionTypeExists
		}
		tx.Rollback()
		return models.CollectionType{}, err
	}
	tx.Commit()
	return collectionType, nil
}

// UpdateCollectionType updates an existing collection type.
// Returns an error if a type with the new name already exists.
func (e *CollectionService) UpdateCollectionType(projectPath, id, collectionTypeName, collectionTypeIcon string) (models.CollectionType, error) {
	remoteURL, err := utils.ResolveProjectRemoteURL(projectPath)
	if err != nil {
		return models.CollectionType{}, fmt.Errorf("failed to resolve project remote: %w", err)
	}
	if remoteURL != "" {
		response, remoteErr := putCollectionTypeRemote(remoteURL, id, collectionTypeName, collectionTypeIcon)
		if remoteErr == nil {
			if response.CollectionType.Id != id {
				return models.CollectionType{}, errors.New("remote collection type response did not contain the requested type")
			}
			dbConn, openErr := utils.OpenDb(projectPath)
			if openErr != nil {
				return models.CollectionType{}, openErr
			}
			defer dbConn.Close()
			tx, beginErr := dbConn.Beginx()
			if beginErr != nil {
				return models.CollectionType{}, beginErr
			}
			defer tx.Rollback()
			if err = applyCanonicalCollectionType(tx, response.CollectionType); err != nil {
				return models.CollectionType{}, err
			}
			if err = applyReturnedSyncToken(tx, response.PreviousSyncToken, response.SyncToken); err != nil {
				return models.CollectionType{}, err
			}
			if err = tx.Commit(); err != nil {
				return models.CollectionType{}, err
			}
			response.CollectionType.Synced = true
			return response.CollectionType, nil
		}
		if !IsMetadataTransportFailure(remoteErr) {
			return models.CollectionType{}, remoteErr
		}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.CollectionType{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.CollectionType{}, err
	}
	defer tx.Rollback()

	collectionType, err := repository.UpdateCollectionType(tx, id, collectionTypeName, collectionTypeIcon)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: collection_type.name" {
			tx.Rollback()
			return models.CollectionType{}, error_service.ErrCollectionTypeExists
		}
		tx.Rollback()
		return models.CollectionType{}, err
	}
	err = tx.Commit()
	if err != nil {
		return models.CollectionType{}, err
	}
	return collectionType, nil
}

// DeleteCollectionType removes a collection type from the project.
// Returns an error if the operation fails.
func (e *CollectionService) DeleteCollectionType(projectPath, id string) error {
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

	err = repository.DeleteCollectionType(tx, id)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// GetCollectionTypes retrieves all collection types in the project.
// Returns the list of collection types or an error if the operation fails.
func (e *CollectionService) GetCollectionTypes(projectPath string) ([]models.CollectionType, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.CollectionType{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.CollectionType{}, err
	}
	defer tx.Rollback()

	collectionTypes, err := repository.GetCollectionTypes(tx)
	if err != nil {
		return collectionTypes, err
	}
	return collectionTypes, nil
}

// IsUserAssignedToCollectionOrAncestor checks if a user is assigned to a collection
// or any of its parent collections recursively.
func (e *CollectionService) IsUserAssignedToCollectionOrAncestor(projectPath, collectionId, userId string) (bool, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return false, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	return repository.IsUserAssignedToCollectionOrAncestor(tx, collectionId, userId)
}

// UpdatePreview updates the preview image for a collection.
// Returns an error if the project is not found or the operation fails.
func (p *CollectionService) UpdatePreview(projectPath, collectionId, previewPath string) error {
	if !utils.FileExists(projectPath) {
		return error_service.ErrProjectNotFound
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

	err = repository.UpdateCollectionPreview(tx, collectionId, previewPath)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
