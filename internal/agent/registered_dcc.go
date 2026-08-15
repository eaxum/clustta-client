package agent

import (
	agentcommands "clustta/internal/agent/commands"
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const (
	dccOpenCommand      = "dcc_open"
	dccRunScriptCommand = "dcc_run_script"
)

func init() {
	registerPlannedDCC(dccOpenCommand, "Open tracked asset files from a structured scope in their DCC application.", map[string]interface{}{
		"app": map[string]interface{}{"type": "string", "description": "Optional DCC application override."},
	}, nil, false, execOpenInDCC)
	registerPlannedDCC("dcc_render", "Launch Blender renders for scoped .blend assets.", map[string]interface{}{
		"start_frame": map[string]interface{}{"type": "integer"},
		"end_frame":   map[string]interface{}{"type": "integer"},
		"output_path": map[string]interface{}{"type": "string"},
		"engine":      map[string]interface{}{"type": "string"},
	}, nil, false, execBlenderRender)
	registerPlannedDCC("dcc_export", "Export scoped .blend assets through Blender.", map[string]interface{}{
		"format":     map[string]interface{}{"type": "string", "enum": []string{"fbx", "obj", "gltf", "usd"}},
		"output_dir": map[string]interface{}{"type": "string"},
	}, []string{"format"}, false, execBlenderExport)
	scriptScope := agentcommands.ScopeSchema([]string{"asset"})
	scriptScope["description"] = "Exactly one tracked Python script asset. Resolve it independently from the target scope."
	registerPlannedDCC(dccRunScriptCommand, "Run one referenced Python script on tracked .blend assets anywhere in the project working directory.", map[string]interface{}{
		"script_scope": scriptScope,
		"script_path": map[string]interface{}{
			"type": "string", "description": "A project-working-directory-relative or absolute .py file path. Use script_scope for a tracked script asset.",
		},
	}, nil, true, execBlenderRunScript)
	registerPlannedDCC("dcc_run_python", "Run inline Python on scoped .blend assets.", map[string]interface{}{
		"python_code": map[string]interface{}{"type": "string"},
		"description": map[string]interface{}{"type": "string"},
	}, []string{"python_code"}, true, execBlenderRunPython)
	registerPlannedDCC("dcc_set_settings", "Change Blender settings on scoped .blend assets.", map[string]interface{}{
		"engine":        map[string]interface{}{"type": "string"},
		"resolution_x":  map[string]interface{}{"type": "integer"},
		"resolution_y":  map[string]interface{}{"type": "integer"},
		"fps":           map[string]interface{}{"type": "integer"},
		"output_format": map[string]interface{}{"type": "string"},
		"samples":       map[string]interface{}{"type": "integer"},
	}, nil, true, execBlenderSetSettings)
	registerDCCLink()
}

func registerDCCLink() {
	agentcommands.Register(agentcommands.Definition{
		Name: "dcc_link_dependencies", Description: "Link or append dependency data from scoped source assets into one scoped target .blend asset.",
		Permission: "update_asset", Risk: "destructive",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"target_scope": agentcommands.ScopeSchema([]string{"asset"}),
				"source_scope": agentcommands.ScopeSchema([]string{"asset"}),
				"link_mode":    map[string]interface{}{"type": "string", "enum": []string{"append", "link"}},
				"object_types": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
				"data_names":   map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			},
			"required": []string{"target_scope"},
		},
		Plan: func(projectPath string, args map[string]interface{}) (planning.Plan, error) {
			target, _, err := resolveDCCArgs(projectPath, map[string]interface{}{"scope": args["target_scope"]})
			if err != nil {
				return planning.Plan{}, err
			}
			if len(target.Entities) != 1 {
				return planning.Plan{}, fmt.Errorf("target_scope must resolve to exactly one asset")
			}
			combined := target
			executorArgs := map[string]interface{}{"target_asset_id": target.Entities[0].ID}
			if raw, ok := args["source_scope"]; ok {
				source, sourceArgs, err := resolveDCCArgs(projectPath, map[string]interface{}{"scope": raw})
				if err != nil {
					return planning.Plan{}, err
				}
				combined.Entities = append(combined.Entities, source.Entities...)
				executorArgs["source_asset_ids"] = sourceArgs["asset_ids"]
			}
			for _, key := range []string{"link_mode", "object_types", "data_names"} {
				if value, ok := args[key]; ok {
					executorArgs[key] = value
				}
			}
			if err := validateBlenderEntities(combined.Entities); err != nil {
				return planning.Plan{}, err
			}
			plan := planning.Plan{
				Command: "dcc_link_dependencies", Scope: combined,
				Counts:    map[string]int{"changes": len(combined.Entities)},
				CreatedAt: time.Now().UTC(), Options: executorArgs,
				Warnings: []string{"This job modifies the target working file. Create a checkpoint before syncing it."},
			}
			for _, entity := range combined.Entities {
				plan.Changes = append(plan.Changes, planning.Change{
					Entity: entity, Action: "DCC Link Dependencies", Valid: true,
					After: map[string]interface{}{"target_asset_id": target.Entities[0].ID},
				})
			}
			return plan, nil
		},
		Execute: func(projectPath string, plan planning.Plan) (planning.Result, error) {
			result := execBlenderLink(projectPath, plan.Options)
			if !result.Success {
				return planning.Result{}, fmt.Errorf("%s", result.Error)
			}
			items := make([]map[string]interface{}, 0, len(plan.Changes))
			for _, change := range plan.Changes {
				items = append(items, map[string]interface{}{
					"type": change.Entity.Type, "id": change.Entity.ID,
					"name": change.Entity.Name, "status": "started",
				})
			}
			return planning.Result{PlanID: plan.ID, Command: plan.Command, Applied: len(items), Items: items}, nil
		},
	})
}

func dccParameters(extra map[string]interface{}, required []string) map[string]interface{} {
	targetScope := agentcommands.ScopeSchema([]string{"asset"})
	targetScope["description"] = "Target tracked assets. Use source entity for one named collection, or source entities with resolved asset or collection IDs for targets across locations. Explicit collection references default to recursive targeting."
	properties := map[string]interface{}{
		"scope": targetScope,
	}
	for key, value := range extra {
		properties[key] = value
	}
	return map[string]interface{}{
		"type": "object", "properties": properties,
		"required": append([]string{"scope"}, required...),
	}
}

func registerPlannedDCC(name, description string, extra map[string]interface{}, required []string, modifiesFiles bool, execute func(string, map[string]interface{}) ToolResult) {
	agentcommands.Register(agentcommands.Definition{
		Name: name, Description: description, Permission: "update_asset", Risk: "destructive",
		Parameters: dccParameters(extra, required),
		Plan: func(projectPath string, args map[string]interface{}) (planning.Plan, error) {
			resolved, executorArgs, err := resolveDCCArgs(projectPath, args)
			if err != nil {
				return planning.Plan{}, err
			}
			if name != dccOpenCommand {
				if err := validateBlenderEntities(resolved.Entities); err != nil {
					return planning.Plan{}, err
				}
			}
			if name == dccRunScriptCommand {
				scriptPath, scriptErr := resolveDCCScriptPath(projectPath, args)
				if scriptErr != nil {
					return planning.Plan{}, scriptErr
				}
				executorArgs["script_path"] = scriptPath
				delete(executorArgs, "script_scope")
			}
			plan := planning.Plan{
				Command: name, Scope: resolved, Counts: map[string]int{"changes": len(resolved.Entities)},
				CreatedAt: time.Now().UTC(), Options: executorArgs,
			}
			if modifiesFiles {
				plan.Warnings = append(plan.Warnings, "This job can modify working files. Create checkpoints before syncing the file changes.")
			}
			if name == dccRunScriptCommand {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("Script: %s", filepath.Base(executorArgs["script_path"].(string))))
			}
			for _, entity := range resolved.Entities {
				plan.Changes = append(plan.Changes, planning.Change{
					Entity: entity, Action: name, Valid: true,
					After: map[string]interface{}{"job": name},
				})
			}
			if len(plan.Changes) == 0 {
				plan.Errors = append(plan.Errors, "scope resolved to no assets")
			}
			return plan, nil
		},
		Execute: func(projectPath string, plan planning.Plan) (planning.Result, error) {
			result := execute(projectPath, plan.Options)
			if !result.Success {
				return planning.Result{}, fmt.Errorf("%s", result.Error)
			}
			items := make([]map[string]interface{}, 0, len(plan.Changes))
			for _, change := range plan.Changes {
				items = append(items, map[string]interface{}{
					"type": change.Entity.Type, "id": change.Entity.ID, "name": change.Entity.Name,
					"status": "started", "action": name,
				})
			}
			return planning.Result{PlanID: plan.ID, Command: name, Applied: len(items), Items: items}, nil
		},
	})
}

func resolveDCCArgs(projectPath string, args map[string]interface{}) (scope.Result, map[string]interface{}, error) {
	req, err := agentcommands.ParseScope(args, []scope.EntityType{scope.TypeAsset})
	if err != nil {
		return scope.Result{}, nil, err
	}
	if (req.Source == "entity" || req.Source == "entities") && !dccRecursiveSpecified(args) {
		req.Recursive = true
	}
	resolved, err := scope.Resolve(projectPath, req)
	if err != nil {
		return scope.Result{}, nil, err
	}
	ids := make([]interface{}, 0, len(resolved.Entities))
	assetIDs := make([]string, 0, len(resolved.Entities))
	for _, entity := range resolved.Entities {
		if entity.Type == scope.TypeAsset {
			ids = append(ids, entity.ID)
			assetIDs = append(assetIDs, entity.ID)
		}
	}
	if len(assetIDs) == 0 {
		return scope.Result{}, nil, fmt.Errorf("target scope resolved to no tracked assets")
	}
	if _, err := resolveAssetFilePaths(projectPath, assetIDs); err != nil {
		return scope.Result{}, nil, err
	}
	executorArgs := map[string]interface{}{}
	for key, value := range args {
		if key != "scope" && key != "_plan_id" {
			executorArgs[key] = value
		}
	}
	executorArgs["asset_ids"] = ids
	return resolved, executorArgs, nil
}

func dccRecursiveSpecified(args map[string]interface{}) bool {
	rawScope, ok := args["scope"].(map[string]interface{})
	if !ok {
		return false
	}
	_, specified := rawScope["recursive"]
	return specified
}

func resolveDCCScriptPath(projectPath string, args map[string]interface{}) (string, error) {
	scriptPath, hasPath := args["script_path"].(string)
	scriptPath = strings.TrimSpace(scriptPath)
	hasPath = hasPath && scriptPath != ""
	_, hasScope := args["script_scope"]
	if hasPath && hasScope {
		return "", fmt.Errorf("provide script_scope or script_path, not both")
	}

	if hasScope {
		req, err := agentcommands.ParseNamedScope(args, "script_scope", []scope.EntityType{scope.TypeAsset})
		if err != nil {
			return "", err
		}
		resolved, err := scope.Resolve(projectPath, req)
		if err != nil {
			return "", err
		}
		if len(resolved.Entities) != 1 || resolved.Entities[0].Type != scope.TypeAsset {
			return "", fmt.Errorf("script_scope must resolve to exactly one tracked asset")
		}
		paths, err := resolveAssetFilePaths(projectPath, []string{resolved.Entities[0].ID})
		if err != nil {
			return "", err
		}
		scriptPath = paths[0]
	} else if !hasPath || scriptPath == "" {
		return "", fmt.Errorf("script_scope or script_path is required")
	} else {
		resolvedPath, err := resolveProjectFilePath(projectPath, scriptPath)
		if err != nil {
			return "", fmt.Errorf("script: %w", err)
		}
		scriptPath = resolvedPath
	}

	if !strings.EqualFold(filepath.Ext(scriptPath), ".py") {
		return "", fmt.Errorf("script must be a .py file: %s", filepath.Base(scriptPath))
	}
	return scriptPath, nil
}

func validateBlenderEntities(entities []scope.Entity) error {
	for _, entity := range entities {
		if !strings.EqualFold(filepath.Ext(entity.Path), ".blend") && !strings.EqualFold(strings.TrimPrefix(entity.Extension, "."), "blend") {
			return fmt.Errorf("asset %s is not a .blend file", entity.Name)
		}
	}
	return nil
}
