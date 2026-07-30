package commands

import (
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"clustta/internal/repository"
	"clustta/internal/utils"
	"fmt"
	"strings"
)

func init() {
	Register(Definition{
		Name:        "batch_assign",
		Description: "Assign tracked assets and/or collections in a structured scope. Asset assignment replaces the assignee; collection assignment adds a direct assignee. Local-only; manual sync required.",
		Permission:  "assign_asset", Risk: "destructive",
		Parameters: assignmentParameters(true),
		Plan:       planAssign,
		Execute:    executeAssign,
	})
	Register(Definition{
		Name:        "batch_unassign",
		Description: "Unassign tracked assets and/or collections in a structured scope. Omit user_id to clear asset assignees and all direct collection assignees. Local-only; manual sync required.",
		Permission:  "unassign_asset", Risk: "destructive",
		Parameters: assignmentParameters(false),
		Plan:       planUnassign,
		Execute:    executeUnassign,
	})
}

func assignmentParameters(requireUser bool) map[string]interface{} {
	required := []string{"scope"}
	if requireUser {
		required = append(required, "user_id")
	}
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"scope":   ScopeSchema([]string{"asset", "collection"}),
			"user_id": map[string]interface{}{"type": "string", "description": "Canonical project user ID."},
		},
		"required": required,
	}
}

func planAssign(projectPath string, args map[string]interface{}) (planning.Plan, error) {
	return planAssignment(projectPath, args, true)
}

func planUnassign(projectPath string, args map[string]interface{}) (planning.Plan, error) {
	return planAssignment(projectPath, args, false)
}

func planAssignment(projectPath string, args map[string]interface{}, assigning bool) (planning.Plan, error) {
	req, err := ParseScope(args, []scope.EntityType{scope.TypeAsset, scope.TypeCollection})
	if err != nil {
		return planning.Plan{}, err
	}
	resolved, err := scope.Resolve(projectPath, req)
	if err != nil {
		return planning.Plan{}, err
	}
	userID := stringArg(args, "user_id")
	if assigning && userID == "" {
		return planning.Plan{}, fmt.Errorf("user_id is required")
	}
	command := "batch_unassign"
	action := "Unassign"
	if assigning {
		command = "batch_assign"
		action = "Assign"
	}
	plan := newPlan(command, resolved)
	userLabel := ""

	if userID != "" {
		db, err := utils.OpenDb(projectPath)
		if err != nil {
			return planning.Plan{}, err
		}
		tx, err := db.Beginx()
		if err != nil {
			db.Close()
			return planning.Plan{}, err
		}
		user, userErr := repository.GetUser(tx, userID)
		userLabel = strings.TrimSpace(user.FirstName + " " + user.LastName)
		if userLabel == "" {
			userLabel = user.Username
		}
		tx.Rollback()
		db.Close()
		if userErr != nil {
			return planning.Plan{}, fmt.Errorf("assignee is not a project user")
		}
	}
	plan.Options = map[string]interface{}{"user_id": userID, "user_label": userLabel}

	for _, entity := range resolved.Entities {
		change := planning.Change{Entity: entity, Action: action, Valid: true}
		requireTracked(&change)
		switch entity.Type {
		case scope.TypeAsset:
			current, _ := entity.Metadata["assignee_id"].(string)
			currentLabel, _ := entity.Metadata["assignee"].(string)
			change.Before = map[string]interface{}{"assignee_id": current, "assignee": currentLabel}
			target := userID
			targetLabel := userLabel
			if !assigning {
				target = ""
				targetLabel = ""
				if userID != "" && current != userID {
					change.Valid = false
					change.Warnings = append(change.Warnings, "asset is not assigned to the requested user")
				}
			}
			change.After = map[string]interface{}{"assignee_id": target, "assignee": targetLabel}
			if current == target {
				change.Valid = false
				change.Warnings = append(change.Warnings, "assignment is already in the requested state")
			}
		case scope.TypeCollection:
			current := stringSliceMetadata(entity, "assignee_ids")
			after := append([]string(nil), current...)
			if assigning {
				if containsString(current, userID) {
					change.Valid = false
					change.Warnings = append(change.Warnings, "user is already assigned to collection")
				} else {
					after = append(after, userID)
				}
			} else if userID == "" {
				after = []string{}
				if len(current) == 0 {
					change.Valid = false
					change.Warnings = append(change.Warnings, "collection has no direct assignees")
				}
			} else if !containsString(current, userID) {
				change.Valid = false
				change.Warnings = append(change.Warnings, "user is not assigned to collection")
			} else {
				after = removeString(after, userID)
			}
			change.Before = map[string]interface{}{"assignee_ids": current, "assignment": assignmentSummary(current, "")}
			change.After = map[string]interface{}{"assignee_ids": after, "assignment": assignmentSummary(after, userLabel)}
			if userLabel != "" {
				change.After["assignee"] = userLabel
				change.After["assignee_id"] = userID
			}
		}
		addChange(&plan, change)
	}
	if len(plan.Changes) == 0 {
		plan.Errors = append(plan.Errors, "scope resolved to no assignable entities")
	}
	return plan, nil
}

func assignmentSummary(ids []string, changedUser string) string {
	if len(ids) == 0 {
		return "Unassigned"
	}
	if len(ids) == 1 && changedUser != "" {
		return changedUser
	}
	return fmt.Sprintf("%d assignees", len(ids))
}

func executeAssign(projectPath string, plan planning.Plan) (planning.Result, error) {
	return executeAssignment(projectPath, plan, true)
}

func executeUnassign(projectPath string, plan planning.Plan) (planning.Result, error) {
	return executeAssignment(projectPath, plan, false)
}

func executeAssignment(projectPath string, plan planning.Plan, assigning bool) (planning.Result, error) {
	userID, _ := plan.Options["user_id"].(string)
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
		switch change.Entity.Type {
		case scope.TypeAsset:
			if assigning {
				asset, err := repository.GetAsset(tx, change.Entity.ID)
				if err != nil {
					return planning.Result{}, err
				}
				if asset.IsResource {
					if err := repository.ToggleIsTask(tx, asset.Id, true); err != nil {
						return planning.Result{}, err
					}
				}
				err = repository.AssignAsset(tx, change.Entity.ID, userID)
				if err != nil {
					return planning.Result{}, err
				}
			} else if err := repository.UnAssignAsset(tx, change.Entity.ID); err != nil {
				return planning.Result{}, err
			}
		case scope.TypeCollection:
			if assigning {
				if err := repository.AssignCollection(tx, change.Entity.ID, userID); err != nil {
					return planning.Result{}, err
				}
			} else {
				removeIDs := stringSliceAny(change.Before["assignee_ids"])
				if userID != "" {
					removeIDs = []string{userID}
				}
				for _, removeID := range removeIDs {
					if err := repository.UnAssignCollection(tx, change.Entity.ID, removeID); err != nil {
						return planning.Result{}, err
					}
				}
			}
		}
		result.Applied++
		result.Items = append(result.Items, resultItem(change, "applied"))
	}
	if err := tx.Commit(); err != nil {
		return planning.Result{}, err
	}
	return result, nil
}

func stringSliceMetadata(entity scope.Entity, key string) []string {
	return stringSliceAny(entity.Metadata[key])
}

func stringSliceAny(value interface{}) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []interface{}:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if value, ok := item.(string); ok {
				out = append(out, value)
			}
		}
		return out
	default:
		return []string{}
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func removeString(values []string, target string) []string {
	out := values[:0]
	for _, value := range values {
		if value != target {
			out = append(out, value)
		}
	}
	return out
}
