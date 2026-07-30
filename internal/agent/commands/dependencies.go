package commands

import (
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"clustta/internal/repository"
	"clustta/internal/utils"
	"fmt"
)

func init() {
	Register(dependencyCommand("batch_add_dependency", true))
	Register(dependencyCommand("batch_remove_dependency", false))
}

func dependencyCommand(name string, adding bool) Definition {
	description := "Add one dependency asset to every tracked asset in a structured scope. Local-only; manual sync required."
	if !adding {
		description = "Remove one dependency asset from every tracked asset in a structured scope. Local-only; manual sync required."
	}
	properties := map[string]interface{}{
		"scope":         map[string]interface{}(ScopeSchema([]string{"asset"})),
		"dependency_id": map[string]interface{}{"type": "string"},
	}
	required := []string{"scope", "dependency_id"}
	if adding {
		properties["dependency_type_id"] = map[string]interface{}{"type": "string"}
		required = append(required, "dependency_type_id")
	}
	return Definition{
		Name: name, Description: description, Permission: "manage_dependencies", Risk: "destructive",
		Parameters: map[string]interface{}{"type": "object", "properties": properties, "required": required},
		Plan: func(projectPath string, args map[string]interface{}) (planning.Plan, error) {
			req, err := ParseScope(args, []scope.EntityType{scope.TypeAsset})
			if err != nil {
				return planning.Plan{}, err
			}
			resolved, err := scope.Resolve(projectPath, req)
			if err != nil {
				return planning.Plan{}, err
			}
			dependencyID := stringArg(args, "dependency_id")
			typeID := stringArg(args, "dependency_type_id")
			if dependencyID == "" || (adding && typeID == "") {
				return planning.Plan{}, fmt.Errorf("dependency_id and dependency_type_id are required")
			}
			db, err := utils.OpenDb(projectPath)
			if err != nil {
				return planning.Plan{}, err
			}
			tx, err := db.Beginx()
			if err != nil {
				db.Close()
				return planning.Plan{}, err
			}
			if _, err := repository.GetAsset(tx, dependencyID); err != nil {
				tx.Rollback()
				db.Close()
				return planning.Plan{}, fmt.Errorf("dependency asset not found")
			}
			tx.Rollback()
			db.Close()

			plan := newPlan(name, resolved)
			plan.Options = map[string]interface{}{"dependency_id": dependencyID, "dependency_type_id": typeID}
			for _, entity := range resolved.Entities {
				change := planning.Change{
					Entity: entity, Action: "Add Dependency", Valid: entity.ID != dependencyID,
					After: map[string]interface{}{"dependency_id": dependencyID, "dependency_type_id": typeID},
				}
				if !adding {
					change.Action = "Remove Dependency"
				}
				if entity.ID == dependencyID {
					change.Errors = append(change.Errors, "asset cannot depend on itself")
				}
				addChange(&plan, change)
			}
			if len(plan.Changes) == 0 {
				plan.Errors = append(plan.Errors, "scope resolved to no assets")
			}
			return plan, nil
		},
		Execute: func(projectPath string, plan planning.Plan) (planning.Result, error) {
			dependencyID, _ := plan.Options["dependency_id"].(string)
			typeID, _ := plan.Options["dependency_type_id"].(string)
			return executeAssetTransaction(projectPath, plan, func(tx txLike, entity scope.Entity) error {
				if adding {
					_, err := repository.AddDependency(tx.Tx(), "", entity.ID, dependencyID, typeID)
					return err
				}
				return repository.RemoveAssetDependency(tx.Tx(), entity.ID, dependencyID)
			})
		},
	}
}
