package commands

import (
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"clustta/internal/repository"
	"clustta/internal/utils"
	"fmt"
)

const (
	dependencyTargetScope = "target_scope"
	dependencySourceScope = "dependency_scope"
	linkedDependencyType  = "linked"
)

func init() {
	Register(dependencyCommand("batch_add_dependency", true))
	Register(dependencyCommand("batch_remove_dependency", false))
}

func dependencyCommand(name string, adding bool) Definition {
	action := "Add Dependency"
	description := "Add selected assets and collections as linked dependencies of one tracked target asset. Local-only; manual sync required."
	if !adding {
		action = "Remove Dependency"
		description = "Remove selected asset and collection dependencies from one tracked target asset. Local-only; manual sync required."
	}
	targetScopeSchema := ScopeSchema([]string{"asset"})
	targetScopeSchema["description"] = "Exactly one tracked target asset. Use entity for a named asset and selection only when one asset is selected."
	targetScopeProperties := targetScopeSchema["properties"].(map[string]interface{})
	targetScopeProperties["source"] = map[string]interface{}{
		"type": "string", "enum": []string{"entity", "selection"},
	}
	dependencyScopeSchema := ScopeSchema([]string{"asset", "collection"})
	dependencyScopeSchema["description"] = "One or more tracked assets and collections to add or remove as dependencies."
	properties := map[string]interface{}{
		dependencyTargetScope: targetScopeSchema,
		dependencySourceScope: dependencyScopeSchema,
	}
	required := []string{dependencyTargetScope, dependencySourceScope}

	return Definition{
		Name: name, Description: description, Permission: "manage_dependencies", Risk: "destructive",
		Parameters: map[string]interface{}{"type": "object", "properties": properties, "required": required},
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
	targetRequest, err := ParseNamedScope(args, dependencyTargetScope, []scope.EntityType{scope.TypeAsset})
	if err != nil {
		return planning.Plan{}, err
	}
	targets, err := scope.Resolve(projectPath, targetRequest)
	if err != nil {
		return planning.Plan{}, fmt.Errorf("resolve target scope: %w", err)
	}
	if len(targets.Entities) != 1 {
		return planning.Plan{}, fmt.Errorf("target_scope must identify exactly one tracked asset using entity or a single-asset selection; resolved %d assets", len(targets.Entities))
	}
	target := targets.Entities[0]

	dependencyRequest, err := ParseNamedScope(args, dependencySourceScope, []scope.EntityType{scope.TypeAsset, scope.TypeCollection})
	if err != nil {
		return planning.Plan{}, err
	}
	dependencies, err := scope.Resolve(projectPath, dependencyRequest)
	if err != nil {
		return planning.Plan{}, fmt.Errorf("resolve dependency scope: %w", err)
	}
	typeID := ""
	if adding {
		typeID, err = dependencyTypeID(projectPath, linkedDependencyType)
		if err != nil {
			return planning.Plan{}, err
		}
	}

	assetDependencies, collectionDependencies, err := existingDependencyIDs(projectPath, target.ID)
	if err != nil {
		return planning.Plan{}, err
	}
	plan := newPlan(name, dependencies)
	plan.Options = map[string]interface{}{
		"target_id": target.ID, "target_name": target.Name, "dependency_type_id": typeID,
	}
	for _, dependency := range dependencies.Entities {
		change := planning.Change{
			Entity: dependency, Action: action, Valid: true,
			After: map[string]interface{}{
				"target_id": target.ID, "target_name": target.Name,
				"dependency_type_id": typeID,
			},
		}
		if dependency.Type == scope.TypeAsset && dependency.ID == target.ID {
			change.Valid = false
			change.Warnings = append(change.Warnings, "target asset excluded from its own dependencies")
		} else {
			exists := assetDependencies[dependency.ID]
			if dependency.Type == scope.TypeCollection {
				exists = collectionDependencies[dependency.ID]
			}
			if adding && exists {
				change.Valid = false
				change.Warnings = append(change.Warnings, "dependency already exists")
			}
			if !adding && !exists {
				change.Valid = false
				change.Warnings = append(change.Warnings, "dependency does not exist")
			}
		}
		addChange(&plan, change)
	}
	if len(plan.Changes) == 0 {
		plan.Errors = append(plan.Errors, "dependency scope resolved to no assets or collections")
	}
	return plan, nil
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

func existingDependencyIDs(projectPath, targetID string) (map[string]bool, map[string]bool, error) {
	db, err := utils.OpenDb(projectPath)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()
	assetRows, err := repository.GetAssetDependencies(tx, targetID)
	if err != nil {
		return nil, nil, err
	}
	collectionRows, err := repository.GetCollectionDependencies(tx, targetID)
	if err != nil {
		return nil, nil, err
	}
	assetIDs := make(map[string]bool, len(assetRows))
	for _, row := range assetRows {
		assetIDs[row.DependencyId] = true
	}
	collectionIDs := make(map[string]bool, len(collectionRows))
	for _, row := range collectionRows {
		collectionIDs[row.DependencyId] = true
	}
	return assetIDs, collectionIDs, nil
}

func executeDependencies(projectPath string, plan planning.Plan, adding bool) (planning.Result, error) {
	targetID, _ := plan.Options["target_id"].(string)
	typeID, _ := plan.Options["dependency_type_id"].(string)
	if targetID == "" {
		return planning.Result{}, fmt.Errorf("approved dependency plan has no target asset")
	}
	return executeAssetTransaction(projectPath, plan, func(tx txLike, dependency scope.Entity) error {
		switch dependency.Type {
		case scope.TypeAsset:
			if adding {
				_, err := repository.AddDependency(tx.Tx(), "", targetID, dependency.ID, typeID)
				return err
			}
			return repository.RemoveAssetDependency(tx.Tx(), targetID, dependency.ID)
		case scope.TypeCollection:
			if adding {
				_, err := repository.AddCollectionDependency(tx.Tx(), "", targetID, dependency.ID, typeID)
				return err
			}
			return repository.RemoveCollectionDependency(tx.Tx(), targetID, dependency.ID)
		default:
			return fmt.Errorf("unsupported dependency type %q", dependency.Type)
		}
	})
}

func selectDependencies(args map[string]interface{}, approved planning.Plan, selectedKeys []string) error {
	selected := make(map[string]bool, len(selectedKeys))
	for _, key := range selectedKeys {
		selected[key] = true
	}
	entities := make([]scope.Entity, 0, len(selectedKeys))
	for _, change := range approved.Changes {
		key := string(change.Entity.Type) + ":" + change.Entity.ID
		if selected[key] && change.Valid {
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
