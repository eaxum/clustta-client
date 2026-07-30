package agent

import (
	agentcommands "clustta/internal/agent/commands"
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"fmt"
	"time"
)

func init() {
	agentcommands.Register(agentcommands.Definition{
		Name:        "dcc_open",
		Description: "Open tracked asset files from a structured scope in their DCC application.",
		Risk:        "safe",
		Parameters: dccParameters(map[string]interface{}{
			"app": map[string]interface{}{"type": "string", "description": "Optional DCC application override."},
		}, nil),
		Direct: func(projectPath string, args map[string]interface{}) (interface{}, error) {
			resolved, legacyArgs, err := resolveDCCArgs(projectPath, args)
			if err != nil {
				return nil, err
			}
			result := execOpenInDCC(projectPath, legacyArgs)
			if !result.Success {
				return nil, fmt.Errorf("%s", result.Error)
			}
			return map[string]interface{}{"entities": resolved.Entities, "result": result.Data}, nil
		},
	})
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
	registerPlannedDCC("dcc_run_script", "Run a Python script on scoped .blend assets.", map[string]interface{}{
		"script_path": map[string]interface{}{"type": "string"},
	}, []string{"script_path"}, true, execBlenderRunScript)
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
			legacy := map[string]interface{}{"target_asset_id": target.Entities[0].ID}
			if raw, ok := args["source_scope"]; ok {
				source, sourceArgs, err := resolveDCCArgs(projectPath, map[string]interface{}{"scope": raw})
				if err != nil {
					return planning.Plan{}, err
				}
				combined.Entities = append(combined.Entities, source.Entities...)
				legacy["source_asset_ids"] = sourceArgs["asset_ids"]
			}
			for _, key := range []string{"link_mode", "object_types", "data_names"} {
				if value, ok := args[key]; ok {
					legacy[key] = value
				}
			}
			plan := planning.Plan{
				Command: "dcc_link_dependencies", Scope: combined,
				Counts:    map[string]int{"changes": len(combined.Entities)},
				CreatedAt: time.Now().UTC(), Options: legacy,
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
	properties := map[string]interface{}{
		"scope": agentcommands.ScopeSchema([]string{"asset"}),
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
			resolved, legacyArgs, err := resolveDCCArgs(projectPath, args)
			if err != nil {
				return planning.Plan{}, err
			}
			plan := planning.Plan{
				Command: name, Scope: resolved, Counts: map[string]int{"changes": len(resolved.Entities)},
				CreatedAt: time.Now().UTC(), Options: legacyArgs,
			}
			if modifiesFiles {
				plan.Warnings = append(plan.Warnings, "This job can modify working files. Create checkpoints before syncing the file changes.")
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
	resolved, err := scope.Resolve(projectPath, req)
	if err != nil {
		return scope.Result{}, nil, err
	}
	ids := make([]interface{}, 0, len(resolved.Entities))
	for _, entity := range resolved.Entities {
		if entity.Type == scope.TypeAsset {
			ids = append(ids, entity.ID)
		}
	}
	legacy := map[string]interface{}{}
	for key, value := range args {
		if key != "scope" && key != "_plan_id" {
			legacy[key] = value
		}
	}
	legacy["asset_ids"] = ids
	return resolved, legacy, nil
}
