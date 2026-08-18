package commands

import (
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"clustta/internal/repository"
	"clustta/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

const (
	moveDestinationMatch        = "destination_match"
	moveMappings                = "move_mappings"
	moveStrategySameNameSibling = "same_name_sibling"
	moveModeShared              = "shared"
	moveModeMapped              = "mapped"
	moveProjectRootName         = "Project root"
)

type moveDestination struct {
	CollectionID string
	Name         string
	Folder       string
	Entity       scope.Entity
}

type movePair struct {
	Source      scope.Entity
	Destination moveDestination
	Warning     string
}

type requestedMoveMapping struct {
	EntityID           string `json:"entity_id"`
	EntityType         string `json:"entity_type"`
	TargetCollectionID string `json:"target_collection_id"`
	TargetPath         string `json:"target_path"`
}

func init() {
	Register(Definition{
		Name:        "batch_move",
		Description: "Move assets and collections using one shared destination, same-name sibling matching, or explicit per-source mappings. Supports tracked and untracked entities. Local-only; manual sync required.",
		Permission:  "update_asset", Risk: "destructive",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"scope":                ScopeSchema([]string{"asset", "collection", "untracked_asset", "untracked_collection"}),
				"target_collection_id": map[string]interface{}{"type": "string", "description": "Shared tracked destination collection ID; omit for project root."},
				"target_path":          map[string]interface{}{"type": "string", "description": "Shared project-relative destination folder for untracked entities."},
				moveDestinationMatch: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"strategy": map[string]interface{}{"type": "string", "enum": []string{moveStrategySameNameSibling}},
					},
					"required": []string{"strategy"},
				},
				moveMappings: map[string]interface{}{
					"type":        "array",
					"description": "Explicit source-to-destination mappings. Normally generated internally when approval rows are selected.",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"entity_id":            map[string]interface{}{"type": "string"},
							"entity_type":          map[string]interface{}{"type": "string", "enum": []string{"asset", "collection", "untracked_asset", "untracked_collection"}},
							"target_collection_id": map[string]interface{}{"type": "string"},
							"target_path":          map[string]interface{}{"type": "string"},
						},
						"required": []string{"entity_id", "entity_type"},
					},
				},
			},
			"required": []string{"scope"},
		},
		Plan:           planMove,
		ExecuteContext: executeMove,
		Select:         selectMoves,
	})
}

func planMove(projectPath string, args map[string]interface{}) (planning.Plan, error) {
	request, err := ParseScope(args, []scope.EntityType{scope.TypeAsset, scope.TypeCollection, scope.TypeUntrackedAsset, scope.TypeUntrackedCollection})
	if err != nil {
		return planning.Plan{}, err
	}
	sources, err := scope.Resolve(projectPath, request)
	if err != nil {
		return planning.Plan{}, err
	}

	mappings, err := parseMoveMappings(args[moveMappings])
	if err != nil {
		return planning.Plan{}, err
	}
	if len(mappings) > 0 {
		pairs, workingDir, err := resolveMappedMoves(projectPath, sources.Entities, mappings)
		if err != nil {
			return planning.Plan{}, err
		}
		return buildMovePlan(projectPath, pairs, workingDir, moveModeMapped)
	}

	strategy, err := moveDestinationStrategy(args)
	if err != nil {
		return planning.Plan{}, err
	}
	if strategy == moveStrategySameNameSibling {
		pairs, workingDir, err := resolveSameNameSiblingMoves(projectPath, sources.Entities)
		if err != nil {
			return planning.Plan{}, err
		}
		return buildMovePlan(projectPath, pairs, workingDir, moveModeMapped)
	}

	destination, workingDir, err := resolveMoveDestination(projectPath, stringArg(args, "target_collection_id"), stringArg(args, "target_path"))
	if err != nil {
		return planning.Plan{}, err
	}
	pairs := make([]movePair, 0, len(sources.Entities))
	for _, source := range sources.Entities {
		pairs = append(pairs, movePair{Source: source, Destination: destination})
	}
	return buildMovePlan(projectPath, pairs, workingDir, moveModeShared)
}

func buildMovePlan(projectPath string, pairs []movePair, workingDir, mode string) (planning.Plan, error) {
	participants := make([]scope.Entity, 0, len(pairs)*2)
	for _, pair := range pairs {
		participants = append(participants, pair.Source)
		if pair.Destination.Entity.ID != "" {
			participants = append(participants, pair.Destination.Entity)
		}
	}
	resolved, err := resolveMovePlanScope(projectPath, participants)
	if err != nil {
		return planning.Plan{}, err
	}
	plan := newPlan("batch_move", resolved)
	plan.Options = map[string]interface{}{"mode": mode}
	targetPaths := make(map[string]string)
	movingEntityIDs := make(map[string]bool, len(pairs))
	movingCollectionPaths := make([]string, 0)
	for _, pair := range pairs {
		entity := pair.Source
		movingEntityIDs[entity.ID] = true
		if entity.Type == scope.TypeCollection || entity.Type == scope.TypeUntrackedCollection {
			movingCollectionPaths = append(movingCollectionPaths, filepath.Clean(entity.Path))
		}
	}

	for _, pair := range pairs {
		change := moveChange(pair, workingDir)
		if mode == moveModeShared {
			markSelectedDescendantMove(&change, movingCollectionPaths)
		}
		validateMoveChange(&change, pair.Destination, workingDir, movingEntityIDs, targetPaths)
		addChange(&plan, change)
	}
	if len(plan.Changes) == 0 {
		plan.Errors = append(plan.Errors, "scope resolved to no movable entities")
	}
	return plan, nil
}

func moveChange(pair movePair, workingDir string) planning.Change {
	source := pair.Source
	destination := pair.Destination
	targetPath := ""
	if destination.Folder != "" {
		targetPath = filepath.Join(destination.Folder, filepath.Base(source.Path))
	}
	change := planning.Change{
		Entity: source, Action: "Move", Valid: destination.Folder != "",
		Before: map[string]interface{}{
			"parent_id": source.ParentID, "path": source.Path,
			"location_name": sourceLocationName(source),
		},
		After: map[string]interface{}{
			"parent_id": destination.CollectionID, "path": targetPath,
			"location_name": destination.Name, "target_folder": destination.Folder,
		},
	}
	if destination.CollectionID == "" && filepath.Clean(destination.Folder) != filepath.Clean(workingDir) {
		change.After["target_path"] = destination.Folder
	}
	if pair.Warning != "" {
		change.Valid = false
		change.Warnings = append(change.Warnings, pair.Warning)
	}
	return change
}

func validateMoveChange(change *planning.Change, destination moveDestination, workingDir string, movingEntityIDs map[string]bool, targetPaths map[string]string) {
	if destination.Folder == "" {
		return
	}
	entity := change.Entity
	targetPath, _ := change.After["path"].(string)
	if entity.Type.Tracked() && destination.CollectionID == "" && filepath.Clean(destination.Folder) != filepath.Clean(workingDir) {
		change.Valid = false
		change.Errors = append(change.Errors, "tracked entities require a tracked target collection")
	}
	if destination.CollectionID != "" && movingEntityIDs[destination.CollectionID] {
		change.Valid = false
		change.Errors = append(change.Errors, "destination collection is also being moved")
	}
	if entity.Type == scope.TypeCollection || entity.Type == scope.TypeUntrackedCollection {
		if isPathInside(destination.Folder, entity.Path) {
			change.Valid = false
			change.Errors = append(change.Errors, "collection cannot be moved into itself or a descendant")
		}
	}
	if filepath.Clean(entity.Path) == filepath.Clean(targetPath) {
		change.Valid = false
		change.Warnings = append(change.Warnings, "entity is already in the target")
	}
	key := normalizedPath(targetPath)
	if prior := targetPaths[key]; prior != "" {
		change.Valid = false
		change.Errors = append(change.Errors, "target conflicts with "+prior)
	} else {
		targetPaths[key] = entity.Name
	}
	if !strings.EqualFold(entity.Path, targetPath) && (utils.FileExists(targetPath) || utils.DirExists(targetPath)) {
		change.Valid = false
		change.Errors = append(change.Errors, "target path already exists")
	}
}

func markSelectedDescendantMove(change *planning.Change, movingCollectionPaths []string) {
	for _, ancestorPath := range movingCollectionPaths {
		if filepath.Clean(change.Entity.Path) == ancestorPath || !isPathInside(change.Entity.Path, ancestorPath) {
			continue
		}
		change.Valid = false
		change.Warnings = append(change.Warnings, "entity moves with its selected ancestor collection")
		return
	}
}

func moveDestinationStrategy(args map[string]interface{}) (string, error) {
	raw, exists := args[moveDestinationMatch]
	if !exists || raw == nil {
		return "", nil
	}
	values, ok := raw.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("destination_match must be an object")
	}
	strategy := stringArg(values, "strategy")
	if strategy != moveStrategySameNameSibling {
		return "", fmt.Errorf("destination_match strategy must be %q", moveStrategySameNameSibling)
	}
	return strategy, nil
}

func resolveSameNameSiblingMoves(projectPath string, sources []scope.Entity) ([]movePair, string, error) {
	collections, err := scope.Resolve(projectPath, scope.Request{
		Source: "project", Recursive: true, Types: []scope.EntityType{scope.TypeCollection},
	})
	if err != nil {
		return nil, "", err
	}
	workingDir, err := projectWorkingDir(projectPath)
	if err != nil {
		return nil, "", err
	}
	candidates := make(map[string][]scope.Entity)
	for _, collection := range collections.Entities {
		key := moveSiblingKey(collection.ParentID, collection.Name)
		candidates[key] = append(candidates[key], collection)
	}
	return pairSameNameMoveSources(sources, candidates), workingDir, nil
}

func pairSameNameMoveSources(sources []scope.Entity, candidates map[string][]scope.Entity) []movePair {
	pairs := make([]movePair, 0, len(sources))
	for _, source := range sources {
		matches := candidates[moveSiblingKey(source.ParentID, source.Name)]
		switch len(matches) {
		case 0:
			pairs = append(pairs, movePair{Source: source, Warning: "no same-name sibling collection found"})
		case 1:
			pairs = append(pairs, movePair{Source: source, Destination: collectionMoveDestination(matches[0])})
		default:
			pairs = append(pairs, movePair{Source: source, Warning: "multiple same-name sibling collections found"})
		}
	}
	return pairs
}

func moveSiblingKey(parentID, name string) string {
	return strings.ToLower(strings.TrimSpace(parentID)) + "\x00" + strings.ToLower(strings.TrimSpace(name))
}

func parseMoveMappings(raw interface{}) ([]requestedMoveMapping, error) {
	if raw == nil {
		return nil, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid move_mappings: %w", err)
	}
	var mappings []requestedMoveMapping
	if err := json.Unmarshal(data, &mappings); err != nil {
		return nil, fmt.Errorf("invalid move_mappings: %w", err)
	}
	seen := make(map[string]bool, len(mappings))
	for index, mapping := range mappings {
		entityType := scope.EntityType(mapping.EntityType)
		if mapping.EntityID == "" || !entityType.Valid() {
			return nil, fmt.Errorf("move_mappings[%d] requires a valid entity_id and entity_type", index)
		}
		if mapping.TargetCollectionID != "" && mapping.TargetPath != "" {
			return nil, fmt.Errorf("move_mappings[%d] cannot use target_collection_id and target_path together", index)
		}
		key := mapping.EntityType + ":" + mapping.EntityID
		if seen[key] {
			return nil, fmt.Errorf("move_mappings contains duplicate source %q", mapping.EntityID)
		}
		seen[key] = true
	}
	return mappings, nil
}

func resolveMappedMoves(projectPath string, sources []scope.Entity, mappings []requestedMoveMapping) ([]movePair, string, error) {
	sourcesByKey := make(map[string]scope.Entity, len(sources))
	for _, source := range sources {
		sourcesByKey[string(source.Type)+":"+source.ID] = source
	}
	destinations, workingDir, err := resolveMappingDestinations(projectPath, mappings)
	if err != nil {
		return nil, "", err
	}
	pairs := make([]movePair, 0, len(mappings))
	for index, mapping := range mappings {
		source, exists := sourcesByKey[mapping.EntityType+":"+mapping.EntityID]
		if !exists {
			return nil, "", fmt.Errorf("mapped source %q is outside the resolved move scope", mapping.EntityID)
		}
		pairs = append(pairs, movePair{Source: source, Destination: destinations[index]})
	}
	return pairs, workingDir, nil
}

func resolveMappingDestinations(projectPath string, mappings []requestedMoveMapping) ([]moveDestination, string, error) {
	db, err := utils.OpenDb(projectPath)
	if err != nil {
		return nil, "", err
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		return nil, "", err
	}
	defer tx.Rollback()
	workingDir, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return nil, "", err
	}
	collectionCache := make(map[string]moveDestination)
	destinations := make([]moveDestination, 0, len(mappings))
	for _, mapping := range mappings {
		if mapping.TargetCollectionID != "" {
			destination, exists := collectionCache[mapping.TargetCollectionID]
			if !exists {
				collection, err := repository.GetCollection(tx, mapping.TargetCollectionID)
				if err != nil {
					return nil, "", fmt.Errorf("target collection not found")
				}
				destination = moveDestination{
					CollectionID: collection.Id, Name: collection.Name, Folder: collection.FilePath,
					Entity: scope.Entity{
						Type: scope.TypeCollection, ID: collection.Id, Name: collection.Name,
						Path: collection.FilePath, ParentID: collection.ParentId,
					},
				}
				collectionCache[mapping.TargetCollectionID] = destination
			}
			destinations = append(destinations, destination)
			continue
		}
		targetFolder := workingDir
		if mapping.TargetPath != "" {
			targetFolder, err = safeProjectTarget(workingDir, mapping.TargetPath)
			if err != nil {
				return nil, "", err
			}
			if !utils.DirExists(targetFolder) {
				return nil, "", fmt.Errorf("target folder not found")
			}
		}
		destination := moveDestination{Name: moveProjectRootName, Folder: targetFolder}
		if filepath.Clean(targetFolder) != filepath.Clean(workingDir) {
			destination.Name = filepath.Base(targetFolder)
			destination.Entity = untrackedMoveDestination(targetFolder)
		}
		destinations = append(destinations, destination)
	}
	return destinations, workingDir, nil
}

func resolveMoveDestination(projectPath, collectionID, targetPath string) (moveDestination, string, error) {
	if collectionID != "" && targetPath != "" {
		return moveDestination{}, "", fmt.Errorf("target_collection_id and target_path cannot be used together")
	}
	mappings := []requestedMoveMapping{{TargetCollectionID: collectionID, TargetPath: targetPath}}
	destinations, workingDir, err := resolveMappingDestinations(projectPath, mappings)
	if err != nil {
		return moveDestination{}, "", err
	}
	return destinations[0], workingDir, nil
}

func collectionMoveDestination(collection scope.Entity) moveDestination {
	return moveDestination{
		CollectionID: collection.ID, Name: collection.Name, Folder: collection.Path, Entity: collection,
	}
}

func untrackedMoveDestination(path string) scope.Entity {
	return scope.Entity{
		Type: scope.TypeUntrackedCollection, ID: utils.GetMD5Hash(path), Name: filepath.Base(path), Path: path,
	}
}

func projectWorkingDir(projectPath string) (string, error) {
	db, err := utils.OpenDb(projectPath)
	if err != nil {
		return "", err
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	return utils.GetProjectWorkingDir(tx)
}

func resolveMovePlanScope(projectPath string, entities []scope.Entity) (scope.Result, error) {
	selection := make([]scope.Entity, 0, len(entities))
	seen := make(map[string]bool, len(entities))
	for _, entity := range entities {
		if entity.ID == "" {
			continue
		}
		key := string(entity.Type) + ":" + entity.ID
		if seen[key] {
			continue
		}
		seen[key] = true
		selection = append(selection, entity)
	}
	request := scope.Request{
		Source: "selection", Selection: selection,
		Types: []scope.EntityType{scope.TypeAsset, scope.TypeCollection, scope.TypeUntrackedAsset, scope.TypeUntrackedCollection},
	}
	return scope.Resolve(projectPath, request)
}

func sourceLocationName(entity scope.Entity) string {
	name := filepath.Base(filepath.Clean(entity.ParentPath))
	if name == "." || name == string(filepath.Separator) || entity.ParentPath == "" {
		return moveProjectRootName
	}
	return name
}

func selectMoves(args map[string]interface{}, approved planning.Plan, selectedKeys []string) error {
	selected := make(map[string]bool, len(selectedKeys))
	for _, key := range selectedKeys {
		selected[key] = true
	}
	entities := make([]scope.Entity, 0, len(selectedKeys))
	mappings := make([]map[string]interface{}, 0, len(selectedKeys))
	for _, change := range approved.Changes {
		if !selected[changeSelectionKey(change)] || !change.Valid {
			continue
		}
		entities = append(entities, change.Entity)
		mapping := map[string]interface{}{
			"entity_id": change.Entity.ID, "entity_type": string(change.Entity.Type),
		}
		if targetID, _ := change.After["parent_id"].(string); targetID != "" {
			mapping["target_collection_id"] = targetID
		} else if targetPath, _ := change.After["target_path"].(string); targetPath != "" {
			mapping["target_path"] = targetPath
		}
		mappings = append(mappings, mapping)
	}
	if len(entities) == 0 {
		return fmt.Errorf("no move items selected")
	}
	args["scope"] = scope.Request{
		Source: "selection", Selection: entities,
		Types: []scope.EntityType{scope.TypeAsset, scope.TypeCollection, scope.TypeUntrackedAsset, scope.TypeUntrackedCollection},
	}
	args[moveMappings] = mappings
	return nil
}

func executeMove(ctx context.Context, projectPath string, plan planning.Plan) (planning.Result, error) {
	changes := append([]planning.Change(nil), plan.Changes...)
	sort.SliceStable(changes, func(i, j int) bool { return changes[i].Entity.Depth > changes[j].Entity.Depth })
	db, err := utils.OpenDb(projectPath)
	if err != nil {
		return planning.Result{}, err
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		return planning.Result{}, err
	}
	defer tx.Rollback()
	journal := []renameMove{}
	rollback := func() {
		for index := len(journal) - 1; index >= 0; index-- {
			_ = utils.RenamePathCaseSafe(journal[index].newPath, journal[index].oldPath)
		}
	}
	result := planning.Result{PlanID: plan.ID, Command: plan.Command, LocalOnly: true, RequiresSync: true}
	for _, change := range changes {
		if err := ctx.Err(); err != nil {
			rollback()
			return result, err
		}
		if !change.Valid {
			result.Skipped++
			result.Items = append(result.Items, resultItem(change, "skipped"))
			continue
		}
		oldPath, _ := change.Before["path"].(string)
		newPath, _ := change.After["path"].(string)
		targetID, _ := change.After["parent_id"].(string)
		switch change.Entity.Type {
		case scope.TypeAsset:
			if err := repository.ChangeCollection(tx, change.Entity.ID, targetID); err != nil {
				rollback()
				return planning.Result{}, err
			}
		case scope.TypeCollection:
			if err := repository.ChangeParent(tx, change.Entity.ID, targetID); err != nil {
				rollback()
				return planning.Result{}, err
			}
		case scope.TypeUntrackedAsset, scope.TypeUntrackedCollection:
			if err := utils.RenamePathCaseSafe(oldPath, newPath); err != nil {
				rollback()
				return planning.Result{}, err
			}
		}
		if oldPath != "" && newPath != "" && oldPath != newPath {
			journal = append(journal, renameMove{oldPath: oldPath, newPath: newPath})
		}
		result.Applied++
		result.Items = append(result.Items, resultItem(change, "applied"))
	}
	if err := tx.Commit(); err != nil {
		rollback()
		return planning.Result{}, err
	}
	return result, nil
}

func safeProjectTarget(root, relative string) (string, error) {
	target := relative
	if !filepath.IsAbs(target) {
		target = filepath.Join(root, relative)
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	relativeTarget, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil || relativeTarget == ".." || strings.HasPrefix(relativeTarget, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("target path is outside the project")
	}
	return targetAbs, nil
}

func isPathInside(target, source string) bool {
	relative, err := filepath.Rel(filepath.Clean(source), filepath.Clean(target))
	return err == nil && (relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))))
}
