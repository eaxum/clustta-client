package commands

import (
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"clustta/internal/repository"
	"clustta/internal/utils"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	dependencyTargetScope       = "target_scope"
	dependencySourceScope       = "dependency_scope"
	dependencyPairing           = "pairing"
	dependencyPairs             = "dependency_pairs"
	dependencyStrategyAllToEach = "all_to_each"
	dependencyStrategySibling   = "same_name_sibling"
	dependencyModePaired        = "paired"
	linkedDependencyType        = "linked"
)

type dependencyPair struct {
	Target     scope.Entity
	Dependency scope.Entity
	Warning    string
}

type requestedDependencyPair struct {
	TargetID             string `json:"target_id"`
	DependencyID         string `json:"dependency_id"`
	DependencyEntityType string `json:"dependency_entity_type"`
}

type dependencyState struct {
	assets      map[string]map[string]bool
	collections map[string]map[string]bool
}

func init() {
	Register(dependencyCommand("batch_add_dependency", true))
	Register(dependencyCommand("batch_remove_dependency", false))
}

func dependencyCommand(name string, adding bool) Definition {
	action := "Add Dependency"
	description := "Add linked asset or collection dependencies to one or more tracked target assets. Supports shared dependencies, all-to-each pairing, and same-name sibling collection pairing. Local-only; manual sync required."
	if !adding {
		action = "Remove Dependency"
		description = "Remove asset or collection dependencies from one or more tracked target assets. Supports shared dependencies, all-to-each pairing, and same-name sibling collection pairing. Local-only; manual sync required."
	}
	targetScopeSchema := ScopeSchema([]string{"asset"})
	targetScopeSchema["description"] = "One target asset for shared dependency mode, or one or more target assets when pairing is provided."
	dependencyScopeSchema := ScopeSchema([]string{"asset", "collection"})
	dependencyScopeSchema["description"] = "Assets or collections to add or remove. Required for shared and all-to-each modes; optional for same-name sibling mode."
	properties := map[string]interface{}{
		dependencyTargetScope: targetScopeSchema,
		dependencySourceScope: dependencyScopeSchema,
		dependencyPairing: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"strategy": map[string]interface{}{
					"type": "string", "enum": []string{dependencyStrategyAllToEach, dependencyStrategySibling},
				},
			},
			"required": []string{"strategy"},
		},
		dependencyPairs: map[string]interface{}{
			"type":        "array",
			"description": "Explicit target and dependency pairs. Normally generated internally when approval rows are selected.",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"target_id":              map[string]interface{}{"type": "string"},
					"dependency_id":          map[string]interface{}{"type": "string"},
					"dependency_entity_type": map[string]interface{}{"type": "string", "enum": []string{"asset", "collection"}},
				},
				"required": []string{"target_id", "dependency_id", "dependency_entity_type"},
			},
		},
	}

	return Definition{
		Name: name, Description: description, Permission: "manage_dependencies", Risk: "destructive",
		Parameters: map[string]interface{}{
			"type": "object", "properties": properties, "required": []string{dependencyTargetScope},
		},
		Plan: func(projectPath string, args map[string]interface{}) (planning.Plan, error) {
			return planDependencies(projectPath, name, action, adding, args)
		},
		Execute: func(projectPath string, plan planning.Plan) (planning.Result, error) {
			return executeDependencies(projectPath, plan, adding)
		},
		Select: selectDependencies,
	}
}

func planDependencies(projectPath, name, action string, adding bool, args map[string]interface{}) (planning.Plan, error) {
	requestedPairs, err := parseRequestedDependencyPairs(args[dependencyPairs])
	if err != nil {
		return planning.Plan{}, err
	}
	if len(requestedPairs) > 0 {
		pairs, err := resolveRequestedDependencyPairs(projectPath, requestedPairs)
		if err != nil {
			return planning.Plan{}, err
		}
		return planPairedDependencies(projectPath, name, action, adding, pairs)
	}

	strategy, err := dependencyPairingStrategy(args)
	if err != nil {
		return planning.Plan{}, err
	}
	if strategy == "" {
		return planSharedDependencies(projectPath, name, action, adding, args)
	}
	return planMatchedDependencies(projectPath, name, action, adding, strategy, args)
}

func planSharedDependencies(projectPath, name, action string, adding bool, args map[string]interface{}) (planning.Plan, error) {
	targets, err := resolveDependencyScope(projectPath, args, dependencyTargetScope, []scope.EntityType{scope.TypeAsset})
	if err != nil {
		return planning.Plan{}, fmt.Errorf("resolve target scope: %w", err)
	}
	if len(targets.Entities) != 1 {
		return planning.Plan{}, fmt.Errorf("target_scope must identify exactly one tracked asset when pairing is omitted; resolved %d assets", len(targets.Entities))
	}
	dependencies, err := resolveDependencyScope(projectPath, args, dependencySourceScope, []scope.EntityType{scope.TypeAsset, scope.TypeCollection})
	if err != nil {
		return planning.Plan{}, fmt.Errorf("resolve dependency scope: %w", err)
	}

	typeID, err := dependencyTypeForPlan(projectPath, adding)
	if err != nil {
		return planning.Plan{}, err
	}
	state, err := loadDependencyState(projectPath, targets.Entities)
	if err != nil {
		return planning.Plan{}, err
	}
	resolved, err := dependencyPlanScope(projectPath, append(targets.Entities, dependencies.Entities...))
	if err != nil {
		return planning.Plan{}, err
	}
	target := targets.Entities[0]
	plan := newPlan(name, resolved)
	plan.Options = map[string]interface{}{
		"target_id": target.ID, "target_name": target.Name, "dependency_type_id": typeID,
	}
	for _, dependency := range dependencies.Entities {
		change := dependencyChange(action, adding, typeID, state, target, dependency)
		change.Entity = dependency
		addChange(&plan, change)
	}
	if len(plan.Changes) == 0 {
		plan.Errors = append(plan.Errors, "dependency scope resolved to no assets or collections")
	}
	return plan, nil
}

func planMatchedDependencies(projectPath, name, action string, adding bool, strategy string, args map[string]interface{}) (planning.Plan, error) {
	targets, err := resolveDependencyScope(projectPath, args, dependencyTargetScope, []scope.EntityType{scope.TypeAsset})
	if err != nil {
		return planning.Plan{}, fmt.Errorf("resolve target scope: %w", err)
	}
	if len(targets.Entities) == 0 {
		return planning.Plan{}, fmt.Errorf("target scope resolved to no assets")
	}

	var dependencies scope.Result
	switch strategy {
	case dependencyStrategyAllToEach:
		dependencies, err = resolveDependencyScope(projectPath, args, dependencySourceScope, []scope.EntityType{scope.TypeAsset, scope.TypeCollection})
	case dependencyStrategySibling:
		dependencies, err = resolveSiblingCandidateScope(projectPath, args)
	default:
		return planning.Plan{}, fmt.Errorf("unsupported dependency pairing strategy %q", strategy)
	}
	if err != nil {
		return planning.Plan{}, fmt.Errorf("resolve dependency scope: %w", err)
	}

	var pairs []dependencyPair
	if strategy == dependencyStrategyAllToEach {
		pairs = pairAllToEach(targets.Entities, dependencies.Entities)
	} else {
		pairs = pairSameNameSiblings(targets.Entities, dependencies.Entities)
	}
	return planPairedDependencies(projectPath, name, action, adding, pairs)
}

func planPairedDependencies(projectPath, name, action string, adding bool, pairs []dependencyPair) (planning.Plan, error) {
	pairs = uniqueDependencyPairs(pairs)
	participants := make([]scope.Entity, 0, len(pairs)*2)
	targetsByID := make(map[string]scope.Entity)
	for _, pair := range pairs {
		participants = append(participants, pair.Target)
		targetsByID[pair.Target.ID] = pair.Target
		if pair.Dependency.ID != "" {
			participants = append(participants, pair.Dependency)
		}
	}
	resolved, err := dependencyPlanScope(projectPath, participants)
	if err != nil {
		return planning.Plan{}, err
	}
	targets := make([]scope.Entity, 0, len(targetsByID))
	for _, target := range targetsByID {
		targets = append(targets, target)
	}
	state, err := loadDependencyState(projectPath, targets)
	if err != nil {
		return planning.Plan{}, err
	}
	typeID, err := dependencyTypeForPlan(projectPath, adding)
	if err != nil {
		return planning.Plan{}, err
	}

	plan := newPlan(name, resolved)
	plan.Options = map[string]interface{}{"mode": dependencyModePaired, "dependency_type_id": typeID}
	dependenciesByKey := make(map[string]bool)
	for _, pair := range pairs {
		if pair.Dependency.ID != "" {
			dependenciesByKey[string(pair.Dependency.Type)+":"+pair.Dependency.ID] = true
		}
		change := dependencyChange(action, adding, typeID, state, pair.Target, pair.Dependency)
		change.Entity = pair.Target
		change.Key = dependencyPairKey(pair.Target, pair.Dependency)
		if pair.Warning != "" {
			change.Valid = false
			change.Warnings = append(change.Warnings, pair.Warning)
		}
		addChange(&plan, change)
	}
	if len(plan.Changes) == 0 {
		plan.Errors = append(plan.Errors, "dependency pairing produced no links")
	}
	plan.Counts["targets"] = len(targetsByID)
	plan.Counts["dependencies"] = len(dependenciesByKey)
	plan.Counts["links"] = plan.Counts["changes"]
	return plan, nil
}

func uniqueDependencyPairs(pairs []dependencyPair) []dependencyPair {
	seen := make(map[string]bool, len(pairs))
	unique := make([]dependencyPair, 0, len(pairs))
	for _, pair := range pairs {
		key := dependencyPairKey(pair.Target, pair.Dependency)
		if seen[key] {
			continue
		}
		seen[key] = true
		unique = append(unique, pair)
	}
	return unique
}

func dependencyChange(action string, adding bool, typeID string, state dependencyState, target, dependency scope.Entity) planning.Change {
	change := planning.Change{
		Entity: target, Action: action, Valid: dependency.ID != "",
		After: map[string]interface{}{
			"target_id": target.ID, "target_name": target.Name,
			"dependency_id": dependency.ID, "dependency_name": dependency.Name,
			"dependency_entity_type": string(dependency.Type), "dependency_type_id": typeID,
		},
	}
	if dependency.ID == "" {
		return change
	}
	if dependency.Type == scope.TypeAsset && dependency.ID == target.ID {
		change.Valid = false
		change.Warnings = append(change.Warnings, "target asset excluded from its own dependencies")
		return change
	}
	exists := state.assets[target.ID][dependency.ID]
	if dependency.Type == scope.TypeCollection {
		exists = state.collections[target.ID][dependency.ID]
	}
	if adding && exists {
		change.Valid = false
		change.Warnings = append(change.Warnings, "dependency already exists")
	}
	if !adding && !exists {
		change.Valid = false
		change.Warnings = append(change.Warnings, "dependency does not exist")
	}
	return change
}

func dependencyPairingStrategy(args map[string]interface{}) (string, error) {
	raw, exists := args[dependencyPairing]
	if !exists || raw == nil {
		return "", nil
	}
	values, ok := raw.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("pairing must be an object")
	}
	strategy := strings.TrimSpace(stringArg(values, "strategy"))
	if strategy != dependencyStrategyAllToEach && strategy != dependencyStrategySibling {
		return "", fmt.Errorf("pairing strategy must be %q or %q", dependencyStrategyAllToEach, dependencyStrategySibling)
	}
	return strategy, nil
}

func resolveDependencyScope(projectPath string, args map[string]interface{}, key string, types []scope.EntityType) (scope.Result, error) {
	if _, exists := args[key]; !exists {
		return scope.Result{}, fmt.Errorf("%s is required", key)
	}
	request, err := ParseNamedScope(args, key, types)
	if err != nil {
		return scope.Result{}, err
	}
	return scope.Resolve(projectPath, request)
}

func resolveSiblingCandidateScope(projectPath string, args map[string]interface{}) (scope.Result, error) {
	if _, exists := args[dependencySourceScope]; exists {
		return resolveDependencyScope(projectPath, args, dependencySourceScope, []scope.EntityType{scope.TypeCollection})
	}
	return scope.Resolve(projectPath, scope.Request{
		Source: "project", Recursive: true, Types: []scope.EntityType{scope.TypeCollection},
	})
}

func pairAllToEach(targets, dependencies []scope.Entity) []dependencyPair {
	pairs := make([]dependencyPair, 0, len(targets)*len(dependencies))
	for _, target := range targets {
		for _, dependency := range dependencies {
			pairs = append(pairs, dependencyPair{Target: target, Dependency: dependency})
		}
	}
	return pairs
}

func pairSameNameSiblings(targets, dependencies []scope.Entity) []dependencyPair {
	candidates := make(map[string][]scope.Entity)
	for _, dependency := range dependencies {
		if dependency.Type != scope.TypeCollection {
			continue
		}
		key := siblingNameKey(dependency.ParentID, dependency.Name)
		candidates[key] = append(candidates[key], dependency)
	}
	pairs := make([]dependencyPair, 0, len(targets))
	for _, target := range targets {
		matches := candidates[siblingNameKey(target.ParentID, target.Name)]
		switch len(matches) {
		case 0:
			pairs = append(pairs, dependencyPair{Target: target, Warning: "no same-name sibling collection found"})
		case 1:
			pairs = append(pairs, dependencyPair{Target: target, Dependency: matches[0]})
		default:
			pairs = append(pairs, dependencyPair{Target: target, Warning: "multiple same-name sibling collections found"})
		}
	}
	return pairs
}

func siblingNameKey(parentID, name string) string {
	return strings.ToLower(strings.TrimSpace(parentID)) + "\x00" + strings.ToLower(strings.TrimSpace(name))
}

func dependencyPairKey(target, dependency scope.Entity) string {
	dependencyKey := "missing"
	if dependency.ID != "" {
		dependencyKey = string(dependency.Type) + ":" + dependency.ID
	}
	return string(target.Type) + ":" + target.ID + "->" + dependencyKey
}

func parseRequestedDependencyPairs(raw interface{}) ([]requestedDependencyPair, error) {
	if raw == nil {
		return nil, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid dependency_pairs: %w", err)
	}
	var pairs []requestedDependencyPair
	if err := json.Unmarshal(data, &pairs); err != nil {
		return nil, fmt.Errorf("invalid dependency_pairs: %w", err)
	}
	for index, pair := range pairs {
		if pair.TargetID == "" || pair.DependencyID == "" {
			return nil, fmt.Errorf("dependency_pairs[%d] requires target_id and dependency_id", index)
		}
		if pair.DependencyEntityType != string(scope.TypeAsset) && pair.DependencyEntityType != string(scope.TypeCollection) {
			return nil, fmt.Errorf("dependency_pairs[%d] has unsupported dependency_entity_type %q", index, pair.DependencyEntityType)
		}
	}
	return pairs, nil
}

func resolveRequestedDependencyPairs(projectPath string, requested []requestedDependencyPair) ([]dependencyPair, error) {
	ids := make([]string, 0, len(requested)*2)
	for _, pair := range requested {
		ids = append(ids, pair.TargetID, pair.DependencyID)
	}
	resolved, err := scope.Resolve(projectPath, scope.Request{
		Source: "entities", EntityIDs: ids, Types: []scope.EntityType{scope.TypeAsset, scope.TypeCollection},
	})
	if err != nil {
		return nil, err
	}
	entities := make(map[string]scope.Entity, len(resolved.Entities))
	for _, entity := range resolved.Entities {
		entities[string(entity.Type)+":"+entity.ID] = entity
	}
	pairs := make([]dependencyPair, 0, len(requested))
	for _, requestedPair := range requested {
		target, ok := entities[string(scope.TypeAsset)+":"+requestedPair.TargetID]
		if !ok {
			return nil, fmt.Errorf("target asset %q no longer exists", requestedPair.TargetID)
		}
		dependencyType := scope.EntityType(requestedPair.DependencyEntityType)
		dependency, ok := entities[string(dependencyType)+":"+requestedPair.DependencyID]
		if !ok {
			return nil, fmt.Errorf("dependency %q no longer exists", requestedPair.DependencyID)
		}
		pairs = append(pairs, dependencyPair{Target: target, Dependency: dependency})
	}
	return pairs, nil
}

func dependencyPlanScope(projectPath string, entities []scope.Entity) (scope.Result, error) {
	seen := make(map[string]bool)
	ids := make([]string, 0, len(entities))
	for _, entity := range entities {
		if entity.ID == "" || seen[entity.ID] {
			continue
		}
		seen[entity.ID] = true
		ids = append(ids, entity.ID)
	}
	if len(ids) == 0 {
		return scope.Result{Request: scope.Request{Source: "selection"}}, nil
	}
	return scope.Resolve(projectPath, scope.Request{
		Source: "entities", EntityIDs: ids, Types: []scope.EntityType{scope.TypeAsset, scope.TypeCollection},
	})
}

func dependencyTypeForPlan(projectPath string, adding bool) (string, error) {
	if !adding {
		return "", nil
	}
	return dependencyTypeID(projectPath, linkedDependencyType)
}

func dependencyTypeID(projectPath, name string) (string, error) {
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
	dependencyType, err := repository.GetDependencyTypeByName(tx, name)
	if err != nil {
		return "", fmt.Errorf("required dependency type %q not found: %w", name, err)
	}
	return dependencyType.Id, nil
}

func loadDependencyState(projectPath string, targets []scope.Entity) (dependencyState, error) {
	state := dependencyState{
		assets: make(map[string]map[string]bool), collections: make(map[string]map[string]bool),
	}
	db, err := utils.OpenDb(projectPath)
	if err != nil {
		return state, err
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		return state, err
	}
	defer tx.Rollback()
	for _, target := range targets {
		if _, exists := state.assets[target.ID]; exists {
			continue
		}
		assetRows, err := repository.GetAssetDependencies(tx, target.ID)
		if err != nil {
			return state, err
		}
		collectionRows, err := repository.GetCollectionDependencies(tx, target.ID)
		if err != nil {
			return state, err
		}
		state.assets[target.ID] = make(map[string]bool, len(assetRows))
		for _, row := range assetRows {
			state.assets[target.ID][row.DependencyId] = true
		}
		state.collections[target.ID] = make(map[string]bool, len(collectionRows))
		for _, row := range collectionRows {
			state.collections[target.ID][row.DependencyId] = true
		}
	}
	return state, nil
}

func executeDependencies(projectPath string, plan planning.Plan, adding bool) (planning.Result, error) {
	mode, _ := plan.Options["mode"].(string)
	if mode == dependencyModePaired {
		return executePairedDependencies(projectPath, plan, adding)
	}
	targetID, _ := plan.Options["target_id"].(string)
	typeID, _ := plan.Options["dependency_type_id"].(string)
	if targetID == "" {
		return planning.Result{}, fmt.Errorf("approved dependency plan has no target asset")
	}
	return executeAssetTransaction(projectPath, plan, func(tx txLike, dependency scope.Entity) error {
		return applyDependency(tx, targetID, dependency.ID, dependency.Type, typeID, adding)
	})
}

func executePairedDependencies(projectPath string, plan planning.Plan, adding bool) (planning.Result, error) {
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
	result := planning.Result{PlanID: plan.ID, Command: plan.Command, LocalOnly: true, RequiresSync: true}
	for _, change := range plan.Changes {
		if !change.Valid {
			result.Skipped++
			result.Items = append(result.Items, resultItem(change, "skipped"))
			continue
		}
		dependencyID, _ := change.After["dependency_id"].(string)
		dependencyType, _ := change.After["dependency_entity_type"].(string)
		typeID, _ := change.After["dependency_type_id"].(string)
		if dependencyID == "" || dependencyType == "" {
			return planning.Result{}, fmt.Errorf("approved dependency pair is incomplete")
		}
		if err := applyDependency(sqlxTx{tx}, change.Entity.ID, dependencyID, scope.EntityType(dependencyType), typeID, adding); err != nil {
			return planning.Result{}, err
		}
		result.Applied++
		result.Items = append(result.Items, resultItem(change, "applied"))
	}
	if err := tx.Commit(); err != nil {
		return planning.Result{}, err
	}
	return result, nil
}

func applyDependency(tx txLike, targetID, dependencyID string, dependencyType scope.EntityType, typeID string, adding bool) error {
	switch dependencyType {
	case scope.TypeAsset:
		if adding {
			_, err := repository.AddDependency(tx.Tx(), "", targetID, dependencyID, typeID)
			return err
		}
		return repository.RemoveAssetDependency(tx.Tx(), targetID, dependencyID)
	case scope.TypeCollection:
		if adding {
			_, err := repository.AddCollectionDependency(tx.Tx(), "", targetID, dependencyID, typeID)
			return err
		}
		return repository.RemoveCollectionDependency(tx.Tx(), targetID, dependencyID)
	default:
		return fmt.Errorf("unsupported dependency type %q", dependencyType)
	}
}

func selectDependencies(args map[string]interface{}, approved planning.Plan, selectedKeys []string) error {
	selected := make(map[string]bool, len(selectedKeys))
	for _, key := range selectedKeys {
		selected[key] = true
	}
	mode, _ := approved.Options["mode"].(string)
	if mode == dependencyModePaired {
		pairs := make([]map[string]interface{}, 0, len(selectedKeys))
		for _, change := range approved.Changes {
			if !selected[changeSelectionKey(change)] || !change.Valid {
				continue
			}
			pairs = append(pairs, map[string]interface{}{
				"target_id":              change.Entity.ID,
				"dependency_id":          change.After["dependency_id"],
				"dependency_entity_type": change.After["dependency_entity_type"],
			})
		}
		if len(pairs) == 0 {
			return fmt.Errorf("no dependency links selected")
		}
		args[dependencyPairs] = pairs
		return nil
	}

	entities := make([]scope.Entity, 0, len(selectedKeys))
	for _, change := range approved.Changes {
		if selected[changeSelectionKey(change)] && change.Valid {
			entities = append(entities, change.Entity)
		}
	}
	if len(entities) == 0 {
		return fmt.Errorf("no dependencies selected")
	}
	args[dependencySourceScope] = scope.Request{
		Source: "selection", Selection: entities,
		Types: []scope.EntityType{scope.TypeAsset, scope.TypeCollection},
	}
	return nil
}
