package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/ignore"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"clustta/output"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type ImportService struct{}

type ImportItems struct {
	Assets    []models.Asset   `json:"assets"`
	Collections []models.Collection `json:"collections"`
}

func getItemPath(root, itemFilepath string, isFile bool) string {
	// Ensure paths are using consistent separators
	root = filepath.ToSlash(root)
	itemFilepath = filepath.ToSlash(itemFilepath)

	// Make sure root ends with a slash
	if !strings.HasSuffix(root, "/") {
		root = root + "/"
	}

	// Get the relative path by removing the root prefix
	relPath := strings.TrimPrefix(itemFilepath, root)

	// If it's a file (has an extension), remove the extension
	if isFile {
		ext := filepath.Ext(relPath)
		if ext != "" {
			relPath = strings.TrimSuffix(relPath, ext)
		}
	}

	relPath = strings.TrimSuffix(relPath, "/")
	return relPath
}

func (i *ImportService) ImportFolder(projectPath, parentId string, folders, files []string, projectWorkingDir string, ignoreList []string) (ImportItems, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ImportItems{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ImportItems{}, err
	}
	defer tx.Rollback()

	ignoreObject := ignore.CompileIgnoreLines(ignoreList...)

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return ImportItems{}, err
	}

	if len(folders)+len(files) == 0 {
		return ImportItems{}, errors.New("missing or invalid parameter: folders/files")
	}

	var assets []string
	var dirs []string
	collectionsMap := map[string]models.Collection{}

	for _, file := range files {
		file, err := filepath.Abs(file)
		if err != nil {
			return ImportItems{}, err
		}
		file = filepath.ToSlash(file)

		relativePath, err := filepath.Rel(projectWorkingDir, file)
		if err != nil {
			return ImportItems{}, err
		}
		if ignoreObject.MatchesPath(relativePath) {
			continue
		}

		assets = append(assets, file)
	}

	for _, folder := range folders {
		folder = filepath.ToSlash(folder)
		relativePath, err := filepath.Rel(projectWorkingDir, folder)
		if err != nil {
			return ImportItems{}, err
		}
		if ignoreObject.MatchesPath(relativePath) {
			continue
		}

		rootAbs, err := filepath.Abs(folder)
		if err != nil {
			return ImportItems{}, err
		}
		rootAbs = filepath.ToSlash(rootAbs)
		dirs = append(dirs, rootAbs)

		err = filepath.WalkDir(rootAbs, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.Type()&fs.ModeSymlink != 0 {
				return nil
			}

			path = filepath.ToSlash(path)

			if path == rootAbs {
				return nil
			}

			// Get the parent path of the relative path
			if d.IsDir() {
				if strings.HasPrefix(filepath.Base(path), ".") {
					return filepath.SkipDir
				}

				folder := filepath.ToSlash(path)

				relativePath, err := filepath.Rel(projectWorkingDir, folder)
				if err != nil {
					return err
				}
				if ignoreObject.MatchesPath(relativePath) {
					return nil
				}

				dirs = append(dirs, path)
			} else {
				if strings.HasPrefix(filepath.Base(path), ".") {
					return nil
				}
				file := filepath.ToSlash(path)

				relativePath, err := filepath.Rel(projectWorkingDir, file)
				if err != nil {
					return err
				}
				if ignoreObject.MatchesPath(relativePath) {
					return nil
				}

				assets = append(assets, path)
			}
			return nil
		})
		if err != nil {
			return ImportItems{}, nil
		}
	}

	collectionsData := []models.Collection{}
	assetsData := []models.Asset{}
	for _, dir := range dirs {
		parentPath := filepath.ToSlash(filepath.Dir(dir))
		dirName := filepath.Base(dir)
		collectionParentId := parentId
		if utils.Contains(dirs, parentPath) {
			collectionParentId = collectionsMap[parentPath].Id
			if collectionParentId == "" {
				return ImportItems{}, errors.New("parent not found in map")
			}
		}
		ruleCollectionType, err := repository.GetCollectionTypeByName(tx, "Generic")
		if err != nil {
			return ImportItems{}, err
		}
		collectionPath := getItemPath(rootFolder, dir, false)
		collection := models.Collection{
			Id:             uuid.New().String(),
			Name:           dirName,
			ParentId:       collectionParentId,
			CollectionTypeId:   ruleCollectionType.Id,
			CollectionTypeIcon: ruleCollectionType.Icon,
			CollectionTypeName: ruleCollectionType.Name,
			FilePath:       dir,
			CollectionPath:     collectionPath,
		}
		collectionsMap[dir] = collection
		collectionsData = append(collectionsData, collection)
	}

	for _, asset := range assets {
		parentPath := filepath.ToSlash(filepath.Dir(asset))
		assetName := strings.TrimSuffix(filepath.Base(asset), filepath.Ext(asset))

		collectionParentId := parentId
		if utils.Contains(dirs, parentPath) {
			collectionParentId = collectionsMap[parentPath].Id
			if collectionParentId == "" {
				return ImportItems{}, errors.New("parent not found in map")
			}
		}

		ruleAssetType, err := repository.GetAssetTypeByName(tx, "Generic")
		if err != nil {
			return ImportItems{}, err
		}
		assetPath := getItemPath(rootFolder, asset, true)
		assetData := models.Asset{
			Id:           uuid.New().String(),
			Name:         assetName,
			AssetTypeId:   ruleAssetType.Id,
			AssetTypeName: ruleAssetType.Name,
			AssetTypeIcon: ruleAssetType.Icon,
			CollectionId:     collectionParentId,
			FilePath:     asset,
			IsResource:   true,
			AssetPath:     assetPath,
		}
		assetsData = append(assetsData, assetData)
	}

	importItems := ImportItems{
		Assets:    assetsData,
		Collections: collectionsData,
	}
	return importItems, nil
}

func (e *ImportService) CreateItems(projectPath string, collections []models.Collection, assets []models.Asset, comment, groupId string) error {
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

	if len(collections)+len(assets) == 0 {
		return errors.New("missing or invalid parameter: no item passed")
	}

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	progress := output.ProgressReport{
		Title:      "Starting Import",
		Message:    "Starting Import",
		Percentage: 0,
		Current:    1,
		Total:      3,
	}
	app.Event.Emit("progress-update", progress)

	totalAssets := len(assets)
	totalCollections := len(collections)
	// totalItems := totalDirs + totalFiles

	for i, collection := range collections {
		_, err = repository.CreateCollection(
			tx, collection.Id, collection.Name, collection.Description, collection.CollectionTypeId, collection.ParentId, collection.PreviewId, collection.IsShared)
		if err != nil {
			return err
		}

		progress := output.ProgressReport{
			Title:      "Creating Collections",
			Message:    collection.Name,
			Percentage: float64(i+1) / float64(totalCollections) * 99,
			Current:    i + 1,
			Total:      totalCollections,
		}
		app.Event.Emit("progress-update", progress)
	}

	for i, asset := range assets {
		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Creating Assets",
				Message:    asset.Name,
				Percentage: float64(current) / float64(total) * 99,
				Current:    i + 1,
				Total:      totalAssets,
			}
			app.Event.Emit("progress-update", progress)
		}
		_, err = repository.CreateAsset(
			tx, asset.Id, asset.Name, asset.AssetTypeId, asset.CollectionId, asset.IsResource,
			"", asset.Description, asset.FilePath, asset.Tags,
			asset.Pointer, asset.IsLink, asset.PreviewId, user.Id, comment, groupId, callBack)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	progress = output.ProgressReport{
		Title:      "Finishing Import",
		Message:    "Finishing Import",
		Percentage: 100,
		Current:    2,
		Total:      2,
	}
	app.Event.Emit("progress-update", progress)

	return nil
}

func (e *ImportService) CreateCollections(projectPath string, collections []models.Collection, completed, totalCollections int) error {
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

	if len(collections) == 0 {
		return errors.New("missing or invalid parameter: no item passed")
	}

	progress := output.ProgressReport{
		Title:      "Starting Import",
		Message:    "Starting Import",
		Percentage: 0,
		Current:    1,
		Total:      2,
	}
	app.Event.Emit("progress-update", progress)

	state := completed
	for i, collection := range collections {
		_, err = repository.CreateCollection(
			tx, collection.Id, collection.Name, collection.Description, collection.CollectionTypeId, collection.ParentId, collection.PreviewId, collection.IsShared)
		if err != nil {
			return err
		}

		progress := output.ProgressReport{
			Title:      "Creating Collections",
			Message:    collection.Name,
			Percentage: float64(i+1) / float64(totalCollections) * 99,
			Current:    completed + (i + 1),
			Total:      totalCollections,
		}
		app.Event.Emit("progress-update", progress)
		state = completed + (i + 1)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	if state == totalCollections {
		progress = output.ProgressReport{
			Title:      "Finishing Import",
			Message:    "Finishing Import",
			Percentage: 100,
			Current:    2,
			Total:      2,
		}
		app.Event.Emit("progress-update", progress)
	}

	return nil
}

func (e *ImportService) CreateAssets(projectPath string, assets []models.Asset, completed, totalAssets int, comment, groupId string) error {
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

	if len(assets) == 0 {
		return errors.New("missing or invalid parameter: no item passed")
	}

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	progress := output.ProgressReport{
		Title:      "Starting Import",
		Message:    "Starting Import",
		Percentage: 0,
		Current:    1,
		Total:      3,
	}
	app.Event.Emit("progress-update", progress)
	state := completed
	for i, asset := range assets {
		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Creating Assets",
				Message:    asset.Name,
				Percentage: float64(current) / float64(total) * 99,
				Current:    completed + (i + 1),
				Total:      totalAssets,
			}
			app.Event.Emit("progress-update", progress)
		}
		_, err = repository.CreateAsset(
			tx, asset.Id, asset.Name, asset.AssetTypeId, asset.CollectionId, asset.IsResource,
			"", asset.Description, asset.FilePath, asset.Tags,
			asset.Pointer, asset.IsLink, asset.PreviewId, user.Id, comment, groupId, callBack)
		if err != nil {
			return err
		}
		state = completed + (i + 1)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	if state == totalAssets {
		progress = output.ProgressReport{
			Title:      "Finishing Import",
			Message:    "Finishing Import",
			Percentage: 100,
			Current:    2,
			Total:      2,
		}
		app.Event.Emit("progress-update", progress)
	}

	return nil
}
