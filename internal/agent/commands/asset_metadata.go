package commands

import (
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"clustta/internal/repository"
	"clustta/internal/utils"
	"fmt"
)

func init() {
	Register(assetBooleanCommand())
	Register(assetTagCommand("batch_add_tags", true))
	Register(assetTagCommand("batch_remove_tags", false))
}

func assetBooleanCommand() Definition {
	return Definition{
		Name:        "batch_toggle_task_resource",
		Description: "Set tracked assets to task or resource state in a structured scope. Local-only; manual sync required.",
		Permission:  "update_asset", Risk: "destructive",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"scope":   ScopeSchema([]string{"asset"}),
				"is_task": map[string]interface{}{"type": "boolean"},
			},
			"required": []string{"scope", "is_task"},
		},
		Plan: func(projectPath string, args map[string]interface{}) (planning.Plan, error) {
			req, err := ParseScope(args, []scope.EntityType{scope.TypeAsset})
			if err != nil {
				return planning.Plan{}, err
			}
			resolved, err := scope.Resolve(projectPath, req)
			if err != nil {
				return planning.Plan{}, err
			}
			isTask, ok := args["is_task"].(bool)
			if !ok {
				return planning.Plan{}, fmt.Errorf("is_task is required")
			}
			plan := newPlan("batch_toggle_task_resource", resolved)
			plan.Options = map[string]interface{}{"is_task": isTask}
			for _, entity := range resolved.Entities {
				currentResource, _ := entity.Metadata["is_resource"].(bool)
				currentTask := !currentResource
				change := planning.Change{
					Entity: entity, Action: "Set Task/Resource", Valid: currentTask != isTask,
					Before: map[string]interface{}{"is_task": currentTask},
					After:  map[string]interface{}{"is_task": isTask},
				}
				if !change.Valid {
					change.Warnings = append(change.Warnings, "already in the requested state")
				}
				addChange(&plan, change)
			}
			if len(plan.Changes) == 0 {
				plan.Errors = append(plan.Errors, "scope resolved to no assets")
			}
			return plan, nil
		},
		Execute: func(projectPath string, plan planning.Plan) (planning.Result, error) {
			isTask, _ := plan.Options["is_task"].(bool)
			return executeAssetTransaction(projectPath, plan, func(tx txLike, entity scope.Entity) error {
				return repository.ToggleIsTask(tx.Tx(), entity.ID, isTask)
			})
		},
	}
}

func assetTagCommand(name string, adding bool) Definition {
	description := "Add a tag to tracked assets in a structured scope. Local-only; manual sync required."
	property := "tag_name"
	if !adding {
		description = "Remove a tag from tracked assets in a structured scope. Local-only; manual sync required."
		property = "tag_id"
	}
	return Definition{
		Name: name, Description: description, Permission: "update_asset", Risk: "destructive",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"scope":  ScopeSchema([]string{"asset"}),
				property: map[string]interface{}{"type": "string"},
			},
			"required": []string{"scope", property},
		},
		Plan: func(projectPath string, args map[string]interface{}) (planning.Plan, error) {
			req, err := ParseScope(args, []scope.EntityType{scope.TypeAsset})
			if err != nil {
				return planning.Plan{}, err
			}
			resolved, err := scope.Resolve(projectPath, req)
			if err != nil {
				return planning.Plan{}, err
			}
			value := stringArg(args, property)
			if value == "" {
				return planning.Plan{}, fmt.Errorf("%s is required", property)
			}
			plan := newPlan(name, resolved)
			plan.Options = map[string]interface{}{property: value}
			for _, entity := range resolved.Entities {
				change := planning.Change{
					Entity: entity, Action: "Add Tag", Valid: true,
					After: map[string]interface{}{property: value},
				}
				if !adding {
					change.Action = "Remove Tag"
				}
				addChange(&plan, change)
			}
			if len(plan.Changes) == 0 {
				plan.Errors = append(plan.Errors, "scope resolved to no assets")
			}
			return plan, nil
		},
		Execute: func(projectPath string, plan planning.Plan) (planning.Result, error) {
			value, _ := plan.Options[property].(string)
			return executeAssetTransaction(projectPath, plan, func(tx txLike, entity scope.Entity) error {
				if adding {
					return repository.AddTagToAsset(tx.Tx(), entity.ID, value)
				}
				return repository.RemoveTagFromAsset(tx.Tx(), entity.ID, value)
			})
		},
	}
}

func executeAssetTransaction(projectPath string, plan planning.Plan, apply func(txLike, scope.Entity) error) (planning.Result, error) {
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
		if err := apply(sqlxTx{tx}, change.Entity); err != nil {
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
