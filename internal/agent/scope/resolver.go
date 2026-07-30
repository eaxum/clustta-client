package scope

import (
	"clustta/internal/ignore"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Resolve resolves a structured scope against the current local project.
// Untracked entities are accepted from an authoritative frontend selection;
// project-wide filesystem discovery is added separately to avoid duplicating
// browser ignore and visibility rules.
func Resolve(projectPath string, req Request) (Result, error) {
	if req.Source == "" {
		req.Source = "selection"
	}
	allowed := allowedTypes(req.Types)
	entities := make([]Entity, 0)

	db, err := utils.OpenDb(projectPath)
	if err != nil {
		return Result{}, err
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback()

	collections, err := repository.GetCollections(tx, false)
	if err != nil {
		return Result{}, err
	}
	assets, err := repository.GetAssets(tx, false)
	if err != nil {
		return Result{}, err
	}

	collectionByID := make(map[string]models.Collection, len(collections))
	for _, collection := range collections {
		collectionByID[collection.Id] = collection
	}
	assetByID := make(map[string]models.Asset, len(assets))
	for _, asset := range assets {
		assetByID[asset.Id] = asset
	}

	if req.Source == "here" && req.EntityID != "" && req.Path == "" {
		if _, tracked := collectionByID[req.EntityID]; !tracked {
			if path, found := findUntrackedCollectionPath(tx, req.EntityID); found {
				req.Path = path
			}
		}
	}

	if req.Source == "selection" {
		workingDir, _ := utils.GetProjectWorkingDir(tx)
		for _, selected := range req.Selection {
			if err := validateSelectedEntity(selected); err != nil {
				return Result{}, err
			}
			if !allowed[selected.Type] {
				continue
			}
			switch selected.Type {
			case TypeAsset:
				asset, ok := assetByID[selected.ID]
				if !ok {
					return Result{}, fmt.Errorf("selected asset %q no longer exists", selected.Name)
				}
				entities = append(entities, assetEntity(asset, collectionByID))
			case TypeCollection:
				collection, ok := collectionByID[selected.ID]
				if !ok {
					return Result{}, fmt.Errorf("selected collection %q no longer exists", selected.Name)
				}
				entities = append(entities, collectionEntity(collection))
			case TypeUntrackedAsset, TypeUntrackedCollection:
				entity, err := refreshUntrackedSelection(selected, workingDir)
				if err != nil {
					return Result{}, err
				}
				entities = append(entities, entity)
			}
		}
		return finish(req, entities), nil
	}

	rootCollectionID, err := resolveRootCollectionID(req, collections)
	if err != nil {
		return Result{}, err
	}

	switch req.Source {
	case "project":
		entities = appendTracked(entities, collections, assets, collectionByID, allowed, "", true, req.Filters)
		untracked, scanErr := scanUntracked(tx, projectPath, collections, assets, allowed, "", "", true, req.Filters)
		if scanErr != nil {
			return Result{}, scanErr
		}
		entities = append(entities, untracked...)
	case "here":
		if rootCollectionID != "" || req.Path == "" {
			entities = appendTracked(entities, collections, assets, collectionByID, allowed, rootCollectionID, req.Recursive, req.Filters)
		}
		untracked, scanErr := scanUntracked(tx, projectPath, collections, assets, allowed, rootCollectionID, req.Path, req.Recursive, req.Filters)
		if scanErr != nil {
			return Result{}, scanErr
		}
		entities = append(entities, untracked...)
	case "entity":
		entities = appendEntityScope(entities, req, collections, assets, collectionByID, allowed)
	default:
		return Result{}, fmt.Errorf("unsupported scope source %q", req.Source)
	}

	return finish(req, entities), nil
}

func findUntrackedCollectionPath(tx *sqlx.Tx, entityID string) (string, bool) {
	workingDir, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return "", false
	}
	var found string
	_ = filepath.WalkDir(workingDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if !entry.IsDir() || path == workingDir {
			return nil
		}
		if strings.HasPrefix(entry.Name(), ".") {
			return filepath.SkipDir
		}
		if utils.GetMD5Hash(path) == entityID {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	return found, found != ""
}

func scanUntracked(tx *sqlx.Tx, projectPath string, collections []models.Collection, assets []models.Asset, allowed map[EntityType]bool, rootCollectionID, rootPath string, recursive bool, filters map[string]interface{}) ([]Entity, error) {
	if !allowed[TypeUntrackedAsset] && !allowed[TypeUntrackedCollection] {
		return nil, nil
	}
	workingDir, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return nil, err
	}
	scanRoot := workingDir
	if rootCollectionID != "" {
		found := false
		for _, collection := range collections {
			if collection.Id == rootCollectionID {
				scanRoot = collection.FilePath
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("scope collection not found")
		}
	} else if rootPath != "" {
		candidate := rootPath
		if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(workingDir, filepath.FromSlash(candidate))
		}
		rootAbs, absErr := filepath.Abs(workingDir)
		if absErr != nil {
			return nil, absErr
		}
		candidateAbs, absErr := filepath.Abs(candidate)
		if absErr != nil {
			return nil, absErr
		}
		relative, relErr := filepath.Rel(rootAbs, candidateAbs)
		if relErr != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("scope path is outside the project")
		}
		scanRoot = candidateAbs
	}
	if !utils.DirExists(scanRoot) {
		return nil, nil
	}

	trackedFiles := map[string]bool{}
	if path, absErr := filepath.Abs(projectPath); absErr == nil {
		trackedFiles[strings.ToLower(filepath.Clean(path))] = true
	}
	for _, asset := range assets {
		if path, absErr := filepath.Abs(asset.GetFilePath()); absErr == nil {
			trackedFiles[strings.ToLower(filepath.Clean(path))] = true
		}
	}
	trackedFolders := map[string]string{}
	for _, collection := range collections {
		if path, absErr := filepath.Abs(collection.FilePath); absErr == nil {
			trackedFolders[strings.ToLower(filepath.Clean(path))] = collection.Id
		}
	}
	ignoreLines, _ := repository.GetIgnoreList(tx)
	matcher := ignore.CompileIgnoreLines(ignoreLines...)
	out := []Entity{}

	addEntry := func(path string, entry fs.DirEntry) error {
		relative, err := filepath.Rel(workingDir, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if relative == "." || strings.HasPrefix(entry.Name(), ".") || matcher.MatchesPath(relative) {
			return nil
		}
		cleanPath := strings.ToLower(filepath.Clean(path))
		if entry.IsDir() {
			if _, tracked := trackedFolders[cleanPath]; tracked {
				return nil
			}
			if !allowed[TypeUntrackedCollection] {
				return nil
			}
			entity := Entity{
				Type: TypeUntrackedCollection, ID: utils.GetMD5Hash(path), Name: entry.Name(),
				Path: path, ParentPath: filepath.Dir(path), ParentID: nearestCollectionID(filepath.Dir(path), trackedFolders),
				Depth: pathDepth(relative),
			}
			if matches(entity, filters) {
				out = append(out, entity)
			}
			return nil
		}
		if trackedFiles[cleanPath] || !allowed[TypeUntrackedAsset] {
			return nil
		}
		extension := filepath.Ext(entry.Name())
		entity := Entity{
			Type: TypeUntrackedAsset, ID: utils.GetMD5Hash(path),
			Name: strings.TrimSuffix(entry.Name(), extension), Path: path,
			ParentPath: filepath.Dir(path), CollectionID: nearestCollectionID(filepath.Dir(path), trackedFolders),
			Extension: extension, Depth: pathDepth(relative),
		}
		entity.ParentID = entity.CollectionID
		if matches(entity, filters) {
			out = append(out, entity)
		}
		return nil
	}

	if !recursive {
		entries, err := os.ReadDir(scanRoot)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if err := addEntry(filepath.Join(scanRoot, entry.Name()), entry); err != nil {
				return nil, err
			}
		}
		return out, nil
	}

	err = filepath.WalkDir(scanRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == scanRoot {
			return nil
		}
		relative, relErr := filepath.Rel(workingDir, path)
		if relErr != nil {
			return relErr
		}
		if entry.IsDir() && (strings.HasPrefix(entry.Name(), ".") || matcher.MatchesPath(filepath.ToSlash(relative))) {
			return filepath.SkipDir
		}
		return addEntry(path, entry)
	})
	return out, err
}

func nearestCollectionID(path string, trackedFolders map[string]string) string {
	for {
		if id := trackedFolders[strings.ToLower(filepath.Clean(path))]; id != "" {
			return id
		}
		parent := filepath.Dir(path)
		if parent == path {
			return ""
		}
		path = parent
	}
}

func refreshUntrackedSelection(entity Entity, workingDir string) (Entity, error) {
	path, err := filepath.Abs(entity.Path)
	if err != nil {
		return Entity{}, err
	}
	root, err := filepath.Abs(workingDir)
	if err != nil {
		return Entity{}, err
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return Entity{}, fmt.Errorf("selected untracked path is outside the project")
	}
	info, err := os.Stat(path)
	if err != nil {
		return Entity{}, fmt.Errorf("selected untracked item %q is no longer available: %w", entity.Name, err)
	}
	if entity.Type == TypeUntrackedAsset && info.IsDir() {
		return Entity{}, fmt.Errorf("selected untracked asset is a directory")
	}
	if entity.Type == TypeUntrackedCollection && !info.IsDir() {
		return Entity{}, fmt.Errorf("selected untracked collection is not a directory")
	}
	entity.Path = path
	entity.ParentPath = filepath.Dir(path)
	entity.Depth = pathDepth(rel)
	if entity.Type == TypeUntrackedAsset {
		entity.Extension = filepath.Ext(info.Name())
		entity.Name = strings.TrimSuffix(info.Name(), entity.Extension)
	} else {
		entity.Name = info.Name()
	}
	return entity, nil
}

func allowedTypes(types []EntityType) map[EntityType]bool {
	if len(types) == 0 {
		return map[EntityType]bool{
			TypeAsset: true, TypeCollection: true,
			TypeUntrackedAsset: true, TypeUntrackedCollection: true,
		}
	}
	out := make(map[EntityType]bool, len(types))
	for _, entityType := range types {
		if entityType.Valid() {
			out[entityType] = true
		}
	}
	return out
}

func validateSelectedEntity(entity Entity) error {
	if !entity.Type.Valid() {
		return fmt.Errorf("invalid selected entity type %q", entity.Type)
	}
	if entity.ID == "" {
		return fmt.Errorf("selected %s is missing id", entity.Type)
	}
	if !entity.Type.Tracked() && entity.Path == "" {
		return fmt.Errorf("selected %s %q is missing path", entity.Type, entity.Name)
	}
	return nil
}

func resolveRootCollectionID(req Request, collections []models.Collection) (string, error) {
	if req.EntityID != "" {
		for _, collection := range collections {
			if collection.Id == req.EntityID {
				return collection.Id, nil
			}
		}
		if req.Source == "here" && req.Path != "" {
			return "", nil
		}
		return "", fmt.Errorf("scope collection not found")
	}
	path := cleanRelativePath(req.Path)
	if path == "" {
		return "", nil
	}
	for _, collection := range collections {
		if strings.EqualFold(cleanRelativePath(collection.CollectionPath), path) ||
			strings.EqualFold(filepath.Clean(collection.FilePath), filepath.Clean(req.Path)) {
			return collection.Id, nil
		}
	}
	if req.Source == "here" {
		// A navigated untracked collection has a filesystem path but no
		// collection database row. Let scanUntracked validate and resolve it.
		return "", nil
	}
	return "", fmt.Errorf("scope collection path not found")
}

func appendTracked(out []Entity, collections []models.Collection, assets []models.Asset, collectionByID map[string]models.Collection, allowed map[EntityType]bool, rootID string, recursive bool, filters map[string]interface{}) []Entity {
	descendants := map[string]bool{}
	if rootID != "" {
		descendants[rootID] = true
		if recursive {
			changed := true
			for changed {
				changed = false
				for _, collection := range collections {
					if descendants[collection.ParentId] && !descendants[collection.Id] {
						descendants[collection.Id] = true
						changed = true
					}
				}
			}
		}
	}

	if allowed[TypeCollection] {
		for _, collection := range collections {
			include := rootID == "" && recursive
			if rootID != "" {
				include = recursive && collection.Id != rootID && descendants[collection.Id]
				if !recursive {
					include = collection.ParentId == rootID
				}
			} else if !recursive {
				include = collection.ParentId == ""
			}
			if include {
				entity := collectionEntity(collection)
				if matches(entity, filters) {
					out = append(out, entity)
				}
			}
		}
	}

	if allowed[TypeAsset] {
		for _, asset := range assets {
			include := false
			if rootID == "" {
				include = recursive || asset.CollectionId == ""
			} else if recursive {
				include = descendants[asset.CollectionId]
			} else {
				include = asset.CollectionId == rootID
			}
			if include {
				entity := assetEntity(asset, collectionByID)
				if matches(entity, filters) {
					out = append(out, entity)
				}
			}
		}
	}
	return out
}

func appendEntityScope(out []Entity, req Request, collections []models.Collection, assets []models.Asset, collectionByID map[string]models.Collection, allowed map[EntityType]bool) []Entity {
	for _, asset := range assets {
		if asset.Id == req.EntityID && allowed[TypeAsset] {
			entity := assetEntity(asset, collectionByID)
			if matches(entity, req.Filters) {
				return append(out, entity)
			}
		}
	}
	for _, collection := range collections {
		if collection.Id != req.EntityID {
			continue
		}
		if allowed[TypeCollection] {
			entity := collectionEntity(collection)
			if matches(entity, req.Filters) {
				out = append(out, entity)
			}
		}
		if req.Recursive {
			out = appendTracked(out, collections, assets, collectionByID, allowed, collection.Id, true, req.Filters)
		}
		return out
	}
	return out
}

func assetEntity(asset models.Asset, collections map[string]models.Collection) Entity {
	parentPath := ""
	if collection, ok := collections[asset.CollectionId]; ok {
		parentPath = collection.CollectionPath
	}
	return Entity{
		Type: TypeAsset, ID: asset.Id, Name: asset.Name, Path: asset.GetFilePath(),
		ParentID: asset.CollectionId, ParentPath: parentPath, CollectionID: asset.CollectionId,
		Extension: asset.Extension, Depth: pathDepth(parentPath) + 1,
		Metadata: map[string]interface{}{
			"status_id": asset.StatusId, "status": asset.StatusShortName,
			"asset_type_id": asset.AssetTypeId, "asset_type": asset.AssetTypeName,
			"asset_type_icon": asset.AssetTypeIcon,
			"assignee_id":     asset.AssigneeId, "assignee": asset.AssigneeName,
			"is_resource": asset.IsResource, "tags": asset.Tags, "state": asset.FileStatus,
		},
	}
}

func collectionEntity(collection models.Collection) Entity {
	return Entity{
		Type: TypeCollection, ID: collection.Id, Name: collection.Name,
		Path: collection.FilePath, ParentID: collection.ParentId,
		ParentPath: cleanRelativePath(filepath.Dir(collection.CollectionPath)),
		Depth:      pathDepth(collection.CollectionPath),
		Metadata: map[string]interface{}{
			"collection_type_id":   collection.CollectionTypeId,
			"collection_type":      collection.CollectionTypeName,
			"collection_type_icon": collection.CollectionTypeIcon,
			"assignee_ids":         collection.AssigneeIds,
		},
	}
}

func matches(entity Entity, filters map[string]interface{}) bool {
	if len(filters) == 0 {
		return true
	}
	if value := stringFilter(filters, "name"); value != "" && !strings.Contains(strings.ToLower(entity.Name), strings.ToLower(value)) {
		return false
	}
	if value := stringFilter(filters, "status_id"); value != "" && metadataString(entity, "status_id") != value {
		return false
	}
	if value := stringFilter(filters, "status"); value != "" && !strings.EqualFold(metadataString(entity, "status"), value) {
		return false
	}
	if value := stringFilter(filters, "asset_type_id"); value != "" && metadataString(entity, "asset_type_id") != value {
		return false
	}
	if value := stringFilter(filters, "asset_type"); value != "" && !strings.EqualFold(metadataString(entity, "asset_type"), value) {
		return false
	}
	if value := stringFilter(filters, "collection_type_id"); value != "" && metadataString(entity, "collection_type_id") != value {
		return false
	}
	if value := stringFilter(filters, "collection_type"); value != "" && !strings.EqualFold(metadataString(entity, "collection_type"), value) {
		return false
	}
	if value := stringFilter(filters, "assignee_id"); value != "" && metadataString(entity, "assignee_id") != value {
		return false
	}
	if value := stringFilter(filters, "extension"); value != "" && !strings.EqualFold(strings.TrimPrefix(entity.Extension, "."), strings.TrimPrefix(value, ".")) {
		return false
	}
	if value := stringFilter(filters, "state"); value != "" && !strings.EqualFold(metadataString(entity, "state"), value) {
		return false
	}
	if value := stringFilter(filters, "tag"); value != "" && !containsFold(metadataStrings(entity, "tags"), value) {
		return false
	}
	if unassigned, ok := filters["unassigned"].(bool); ok && unassigned && metadataString(entity, "assignee_id") != "" {
		return false
	}
	return true
}

func finish(req Request, entities []Entity) Result {
	seen := map[string]bool{}
	out := make([]Entity, 0, len(entities))
	for _, entity := range entities {
		key := string(entity.Type) + ":" + entity.ID
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, entity)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Depth != out[j].Depth {
			return out[i].Depth < out[j].Depth
		}
		if out[i].Type != out[j].Type {
			return out[i].Type < out[j].Type
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	if req.Limit > 0 && len(out) > req.Limit {
		out = out[:req.Limit]
	}
	return Result{Request: req, Entities: out}
}

func cleanRelativePath(path string) string {
	path = filepath.ToSlash(filepath.Clean(path))
	if path == "." || path == "/" {
		return ""
	}
	return strings.Trim(path, "/")
}

func pathDepth(path string) int {
	path = cleanRelativePath(path)
	if path == "" {
		return 0
	}
	return strings.Count(path, "/") + 1
}

func stringFilter(filters map[string]interface{}, key string) string {
	value, _ := filters[key].(string)
	return strings.TrimSpace(value)
}

func metadataString(entity Entity, key string) string {
	value, _ := entity.Metadata[key].(string)
	return value
}

func metadataStrings(entity Entity, key string) []string {
	switch value := entity.Metadata[key].(type) {
	case []string:
		return value
	case []interface{}:
		out := make([]string, 0, len(value))
		for _, item := range value {
			if text, ok := item.(string); ok {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}

func containsFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}
