package agent

import (
	agentcommands "clustta/internal/agent/commands"
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"clustta/internal/repository"
	"clustta/internal/utils"
	"fmt"
	"strings"
	"time"
)

const pendingEntityPrefix = "pending:"

type plannedBatchDefinition struct {
	name        string
	description string
	permission  string
	parameters  map[string]interface{}
	itemKey     string
	entityType  scope.EntityType
	nameKey     string
	required    []string
	execute     func(string, map[string]interface{}) ToolResult
}

func init() {
	registerPlannedBatchDefinitions()
	registerBatchDistribute()
}

func registerPlannedBatchDefinitions() {
	iconEnum := []string{"bezier", "bone", "book", "boxes", "bulb", "camera-flash", "camera", "clapboard", "compass", "cube", "drum", "film-reel", "film-strip", "fire", "flow-chart", "four-squares", "home", "image", "lamp", "link", "man-running", "masks", "music", "mystery-ball", "open-book", "package", "palette", "scissors", "shapes", "stall", "texture", "tree", "video-camera", "website"}
	registerPlannedBatch(plannedBatchDefinition{
		name: "batch_create_collections", description: "Create multiple collections, including nested collections, through one reviewed local plan.",
		permission: "create_collection", itemKey: "collections", entityType: scope.TypeCollection, nameKey: "name", required: []string{"name"},
		parameters: collectionBatchSchema(), execute: execBatchCreateCollections,
	})
	registerPlannedBatch(plannedBatchDefinition{
		name: "batch_create_assets", description: "Create multiple assets through one reviewed local plan and transaction.",
		permission: "create_asset", itemKey: "assets", entityType: scope.TypeAsset, nameKey: "name", required: []string{"name", "task_type_id", "template_id"},
		parameters: assetBatchSchema(), execute: execBatchCreateAssets,
	})
	registerPlannedBatch(plannedBatchDefinition{
		name: "batch_create_asset_types", description: "Create multiple asset types through one reviewed local plan and transaction.",
		permission: "create_asset", itemKey: "items", entityType: scope.TypeAsset, nameKey: "name", required: []string{"name", "icon"},
		parameters: typeBatchSchema(iconEnum, false), execute: execBatchCreateAssetTypes,
	})
	registerPlannedBatch(plannedBatchDefinition{
		name: "batch_create_collection_types", description: "Create multiple collection types through one reviewed local plan and transaction.",
		permission: "create_asset", itemKey: "items", entityType: scope.TypeCollection, nameKey: "name", required: []string{"name", "icon"},
		parameters: typeBatchSchema(iconEnum, false), execute: execBatchCreateCollectionTypes,
	})
	registerPlannedBatch(plannedBatchDefinition{
		name: "batch_update_asset_types", description: "Update multiple asset type names and icons through one reviewed local plan and transaction.",
		permission: "create_asset", itemKey: "items", entityType: scope.TypeAsset, nameKey: "name", required: []string{"id", "name", "icon"},
		parameters: typeBatchSchema(iconEnum, true), execute: execBatchUpdateAssetTypes,
	})
	registerPlannedBatch(plannedBatchDefinition{
		name: "batch_update_collection_types", description: "Update multiple collection type names and icons through one reviewed local plan and transaction.",
		permission: "create_asset", itemKey: "items", entityType: scope.TypeCollection, nameKey: "name", required: []string{"id", "name", "icon"},
		parameters: typeBatchSchema(iconEnum, true), execute: execBatchUpdateCollectionTypes,
	})

	registerAggregateBatch("setup_project_types", "Create a standard set of project types through one reviewed local plan.", "create_asset", setupProjectTypesSchema(), scope.TypeCollection, "Set Up Project Types", execSetupProjectTypes)
	animationDefinition := animationSetupToolDef()
	registerAggregateBatch(animationDefinition.Name, animationDefinition.Description, "create_collection", animationDefinition.Parameters, scope.TypeCollection, "Set Up Animation Production", execSetupAnimationProduction)
	registerAggregateBatch("apply_workflow", "Apply a workflow tree through one reviewed local plan.", "create_collection", applyWorkflowSchema(), scope.TypeCollection, "Apply Workflow", execApplyWorkflow)
}

func registerPlannedBatch(config plannedBatchDefinition) {
	agentcommands.Register(agentcommands.Definition{
		Name: config.name, Description: config.description, Permission: config.permission, Risk: RiskDestructive,
		Parameters: config.parameters,
		Plan: func(projectPath string, args map[string]interface{}) (planning.Plan, error) {
			return planExplicitBatch(projectPath, config, args)
		},
		Select: selectBatchItems(config.itemKey),
		Execute: func(projectPath string, plan planning.Plan) (planning.Result, error) {
			return executePlannedBatch(projectPath, plan, config.execute)
		},
	})
}

func registerAggregateBatch(name, description, permission string, parameters map[string]interface{}, entityType scope.EntityType, action string, execute func(string, map[string]interface{}) ToolResult) {
	agentcommands.Register(agentcommands.Definition{
		Name: name, Description: description, Permission: permission, Risk: RiskDestructive, Parameters: parameters,
		Plan: func(projectPath string, args map[string]interface{}) (planning.Plan, error) {
			projectScope, err := resolveProjectScope(projectPath)
			if err != nil {
				return planning.Plan{}, err
			}
			if err := validateBatchParameters(parameters, args); err != nil {
				return planning.Plan{}, err
			}
			entity := scope.Entity{Type: entityType, ID: pendingEntityPrefix + name, Name: action}
			return planning.Plan{
				Command: name, Scope: projectScope, CreatedAt: projectScopeTime(),
				Counts: map[string]int{"changes": 1}, LocalOnly: true, RequiresSync: true,
				Options: copyBatchArgs(args),
				Changes: []planning.Change{{Entity: entity, Action: action, Valid: true, After: copyBatchArgs(args)}},
			}, nil
		},
		Select: selectWholeBatch,
		Execute: func(projectPath string, plan planning.Plan) (planning.Result, error) {
			return executePlannedBatch(projectPath, plan, execute)
		},
	})
}

func planExplicitBatch(projectPath string, config plannedBatchDefinition, args map[string]interface{}) (planning.Plan, error) {
	items, err := batchObjectItems(args[config.itemKey])
	if err != nil {
		return planning.Plan{}, fmt.Errorf("%s %w", config.itemKey, err)
	}
	projectScope, err := resolveProjectScope(projectPath)
	if err != nil {
		return planning.Plan{}, err
	}
	plan := planning.Plan{
		Command: config.name, Scope: projectScope, CreatedAt: projectScopeTime(),
		Counts: map[string]int{}, LocalOnly: true, RequiresSync: true, Options: copyBatchArgs(args),
	}
	for index, item := range items {
		name := batchString(item, config.nameKey)
		if name == "" {
			name = fmt.Sprintf("Item %d", index+1)
		}
		entityID := batchString(item, "id")
		if entityID == "" {
			entityID = fmt.Sprintf("%s%s:%d", pendingEntityPrefix, config.name, index)
		}
		change := planning.Change{
			Entity: scope.Entity{Type: config.entityType, ID: entityID, Name: name, Metadata: map[string]interface{}{"input_index": index}},
			Action: config.description, Valid: true, After: item,
		}
		for _, key := range config.required {
			if batchString(item, key) == "" {
				change.Valid = false
				change.Errors = append(change.Errors, fmt.Sprintf("%s is required", key))
			}
		}
		plan.Changes = append(plan.Changes, change)
		if change.Valid {
			plan.Counts["changes"]++
		} else {
			plan.Counts["invalid"]++
		}
	}
	if len(plan.Changes) == 0 {
		plan.Errors = append(plan.Errors, fmt.Sprintf("%s must not be empty", config.itemKey))
	}
	return plan, nil
}

func executePlannedBatch(projectPath string, plan planning.Plan, execute func(string, map[string]interface{}) ToolResult) (planning.Result, error) {
	result := execute(projectPath, plan.Options)
	if !result.Success {
		return planning.Result{}, fmt.Errorf("%s", result.Error)
	}
	items := make([]map[string]interface{}, 0, len(plan.Changes))
	for _, change := range plan.Changes {
		items = append(items, map[string]interface{}{
			"name": change.Entity.Name, "action": change.Action, "status": "applied",
		})
	}
	if len(items) > 0 {
		items[0]["result"] = result.Data
	}
	return planning.Result{
		PlanID: plan.ID, Command: plan.Command, Applied: len(plan.Changes), Items: items,
		LocalOnly: true, RequiresSync: true,
	}, nil
}

func selectBatchItems(itemKey string) func(map[string]interface{}, planning.Plan, []string) error {
	return func(args map[string]interface{}, approved planning.Plan, selectedKeys []string) error {
		selected := make(map[string]bool, len(selectedKeys))
		for _, key := range selectedKeys {
			selected[key] = true
		}
		items, err := batchObjectItems(args[itemKey])
		if err != nil {
			return err
		}
		kept := make([]interface{}, 0, len(selected))
		for _, change := range approved.Changes {
			key := string(change.Entity.Type) + ":" + change.Entity.ID
			if !selected[key] || !change.Valid {
				continue
			}
			index, ok := change.Entity.Metadata["input_index"].(int)
			if !ok || index < 0 || index >= len(items) {
				return fmt.Errorf("approved batch item is no longer available")
			}
			kept = append(kept, items[index])
		}
		if len(kept) == 0 {
			return fmt.Errorf("no items selected")
		}
		args[itemKey] = kept
		return nil
	}
}

func selectWholeBatch(_ map[string]interface{}, approved planning.Plan, selectedKeys []string) error {
	selected := map[string]bool{}
	for _, key := range selectedKeys {
		selected[key] = true
	}
	for _, change := range approved.Changes {
		if selected[string(change.Entity.Type)+":"+change.Entity.ID] {
			return nil
		}
	}
	return fmt.Errorf("no items selected")
}

func batchObjectItems(raw interface{}) ([]map[string]interface{}, error) {
	items, ok := raw.([]interface{})
	if !ok || len(items) == 0 {
		return nil, fmt.Errorf("must be a non-empty array")
	}
	out := make([]map[string]interface{}, 0, len(items))
	for index, rawItem := range items {
		item, ok := rawItem.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("item %d must be an object", index)
		}
		out = append(out, item)
	}
	return out, nil
}

func resolveProjectScope(projectPath string) (scope.Result, error) {
	return scope.Resolve(projectPath, scope.Request{
		Source: "project", Recursive: true, Types: []scope.EntityType{scope.TypeAsset, scope.TypeCollection},
	})
}

func copyBatchArgs(args map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{}, len(args))
	for key, value := range args {
		if key != "_plan_id" && key != "scope" {
			copy[key] = value
		}
	}
	return copy
}

func batchString(values map[string]interface{}, key string) string {
	value, _ := values[key].(string)
	return strings.TrimSpace(value)
}

func collectionBatchSchema() map[string]interface{} {
	return objectWithArray("collections", map[string]interface{}{
		"name": map[string]interface{}{"type": "string"}, "description": map[string]interface{}{"type": "string"},
		"parent_id": map[string]interface{}{"type": "string"}, "parent_name": map[string]interface{}{"type": "string"},
	}, []string{"name"})
}

func assetBatchSchema() map[string]interface{} {
	return objectWithArray("assets", map[string]interface{}{
		"name": map[string]interface{}{"type": "string"}, "collection_id": map[string]interface{}{"type": "string"},
		"task_type_id": map[string]interface{}{"type": "string"}, "template_id": map[string]interface{}{"type": "string"},
		"tags": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
	}, []string{"name", "task_type_id", "template_id"})
}

func typeBatchSchema(iconEnum []string, includeID bool) map[string]interface{} {
	properties := map[string]interface{}{
		"name": map[string]interface{}{"type": "string"}, "icon": map[string]interface{}{"type": "string", "enum": iconEnum},
	}
	required := []string{"name", "icon"}
	if includeID {
		properties["id"] = map[string]interface{}{"type": "string"}
		required = append([]string{"id"}, required...)
	}
	return objectWithArray("items", properties, required)
}

func objectWithArray(key string, properties map[string]interface{}, required []string) map[string]interface{} {
	return map[string]interface{}{
		"type": "object", "properties": map[string]interface{}{
			key: map[string]interface{}{"type": "array", "items": map[string]interface{}{
				"type": "object", "properties": properties, "required": required,
			}},
		}, "required": []string{key},
	}
}

func setupProjectTypesSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object", "properties": map[string]interface{}{
			"project_type": map[string]interface{}{"type": "string", "enum": []string{"animation", "game", "vfx", "film"}},
		}, "required": []string{"project_type"},
	}
}

func applyWorkflowSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object", "properties": map[string]interface{}{
			"workflow_id": map[string]interface{}{"type": "string"}, "name": map[string]interface{}{"type": "string"},
			"collection_type_id": map[string]interface{}{"type": "string"}, "parent_id": map[string]interface{}{"type": "string"},
		}, "required": []string{"workflow_id", "name", "collection_type_id"},
	}
}

func validateBatchParameters(parameters, args map[string]interface{}) error {
	required, _ := parameters["required"].([]string)
	for _, key := range required {
		value, exists := args[key]
		if !exists {
			return fmt.Errorf("%s is required", key)
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) == "" {
				return fmt.Errorf("%s is required", key)
			}
		case []interface{}:
			if len(typed) == 0 {
				return fmt.Errorf("%s must not be empty", key)
			}
		}
	}
	return nil
}

func projectScopeTime() time.Time {
	return time.Now().UTC()
}

func registerBatchDistribute() {
	agentcommands.Register(agentcommands.Definition{
		Name: "batch_distribute", Description: "Distribute scoped tracked assets across project users in deterministic round-robin order.",
		Permission: "assign_asset", Risk: RiskDestructive,
		Parameters: map[string]interface{}{
			"type": "object", "properties": map[string]interface{}{
				"scope":    agentcommands.ScopeSchema([]string{"asset"}),
				"user_ids": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			}, "required": []string{"scope", "user_ids"},
		},
		Plan:    planBatchDistribute,
		Select:  preserveSelectedDistribution,
		Execute: executeBatchDistribute,
	})
}

func planBatchDistribute(projectPath string, args map[string]interface{}) (planning.Plan, error) {
	req, err := agentcommands.ParseScope(args, []scope.EntityType{scope.TypeAsset})
	if err != nil {
		return planning.Plan{}, err
	}
	resolved, err := scope.Resolve(projectPath, req)
	if err != nil {
		return planning.Plan{}, err
	}
	userIDs := stringList(args["user_ids"])
	if len(userIDs) == 0 {
		return planning.Plan{}, fmt.Errorf("user_ids must not be empty")
	}
	userNames, err := projectUserNames(projectPath, userIDs)
	if err != nil {
		return planning.Plan{}, err
	}
	plan := planning.Plan{
		Command: "batch_distribute", Scope: resolved, CreatedAt: projectScopeTime(), Counts: map[string]int{},
		Options: map[string]interface{}{"user_ids": userIDs}, LocalOnly: true, RequiresSync: true,
	}
	approvedAssignments, _ := args["_assignments"].(map[string]string)
	for index, entity := range resolved.Entities {
		userID := userIDs[index%len(userIDs)]
		if approvedUserID := approvedAssignments[entity.ID]; approvedUserID != "" {
			userID = approvedUserID
		}
		change := planning.Change{
			Entity: entity, Action: "Distribute", Valid: true,
			Before: map[string]interface{}{"assignee_id": entity.Metadata["assignee_id"], "assignee": entity.Metadata["assignee"]},
			After:  map[string]interface{}{"assignee_id": userID, "assignee": userNames[userID]},
		}
		if entity.Metadata["assignee_id"] == userID {
			change.Valid = false
			change.Warnings = append(change.Warnings, "already assigned to the selected user")
		}
		plan.Changes = append(plan.Changes, change)
		if change.Valid {
			plan.Counts["changes"]++
		} else {
			plan.Counts["invalid"]++
		}
	}
	if len(plan.Changes) == 0 {
		plan.Errors = append(plan.Errors, "scope resolved to no assets")
	}
	return plan, nil
}

func preserveSelectedDistribution(args map[string]interface{}, approved planning.Plan, selectedKeys []string) error {
	selected := make(map[string]bool, len(selectedKeys))
	for _, key := range selectedKeys {
		selected[key] = true
	}
	entities := make([]scope.Entity, 0, len(selected))
	assignments := make(map[string]string, len(selected))
	for _, change := range approved.Changes {
		key := string(change.Entity.Type) + ":" + change.Entity.ID
		if !selected[key] || !change.Valid {
			continue
		}
		userID, _ := change.After["assignee_id"].(string)
		entities = append(entities, change.Entity)
		assignments[change.Entity.ID] = userID
	}
	if len(entities) == 0 {
		return fmt.Errorf("no items selected")
	}
	args["scope"] = scope.Request{Source: "selection", Selection: entities, Types: []scope.EntityType{scope.TypeAsset}}
	args["_assignments"] = assignments
	return nil
}

func executeBatchDistribute(projectPath string, plan planning.Plan) (planning.Result, error) {
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
			continue
		}
		userID, _ := change.After["assignee_id"].(string)
		if err := repository.AssignAsset(tx, change.Entity.ID, userID); err != nil {
			return planning.Result{}, fmt.Errorf("assign %s: %w", change.Entity.Name, err)
		}
		result.Applied++
		result.Items = append(result.Items, map[string]interface{}{
			"type": change.Entity.Type, "id": change.Entity.ID, "name": change.Entity.Name,
			"entity": change.Entity, "action": change.Action, "status": "applied", "after": change.After,
		})
	}
	if err := tx.Commit(); err != nil {
		return planning.Result{}, err
	}
	return result, nil
}

func projectUserNames(projectPath string, userIDs []string) (map[string]string, error) {
	db, err := utils.OpenDb(projectPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	names := make(map[string]string, len(userIDs))
	for _, userID := range userIDs {
		user, err := repository.GetUser(tx, userID)
		if err != nil {
			return nil, fmt.Errorf("project user %q not found", userID)
		}
		name := strings.TrimSpace(user.FirstName + " " + user.LastName)
		if name == "" {
			name = user.Username
		}
		names[userID] = name
	}
	return names, nil
}

func stringList(raw interface{}) []string {
	items, _ := raw.([]interface{})
	values := make([]string, 0, len(items))
	for _, item := range items {
		if value, ok := item.(string); ok && strings.TrimSpace(value) != "" {
			values = append(values, strings.TrimSpace(value))
		}
	}
	return values
}
