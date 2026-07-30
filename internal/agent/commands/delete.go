package commands

import (
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"clustta/internal/repository"
	"clustta/internal/utils"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func init() {
	Register(Definition{
		Name:        "batch_delete",
		Description: "Remove entities in a structured scope. Tracked entities are marked deleted; untracked paths are moved into a recoverable hidden project trash. Local-only; manual sync required.",
		Permission:  "delete_asset", Risk: "destructive",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"scope": ScopeSchema([]string{"asset", "collection", "untracked_asset", "untracked_collection"}),
			},
			"required": []string{"scope"},
		},
		Plan:           planDelete,
		ExecuteContext: executeDelete,
	})
}

func planDelete(projectPath string, args map[string]interface{}) (planning.Plan, error) {
	req, err := ParseScope(args, []scope.EntityType{scope.TypeAsset, scope.TypeCollection, scope.TypeUntrackedAsset, scope.TypeUntrackedCollection})
	if err != nil {
		return planning.Plan{}, err
	}
	resolved, err := scope.Resolve(projectPath, req)
	if err != nil {
		return planning.Plan{}, err
	}
	plan := newPlan("batch_delete", resolved)
	for _, entity := range resolved.Entities {
		action := "Mark Deleted"
		if !entity.Type.Tracked() {
			action = "Move to Recoverable Trash"
		}
		addChange(&plan, planning.Change{
			Entity: entity, Action: action, Valid: true,
			Before: map[string]interface{}{"path": entity.Path},
			After:  map[string]interface{}{"deleted": true},
		})
	}
	if len(plan.Changes) == 0 {
		plan.Errors = append(plan.Errors, "scope resolved to no entities")
	}
	plan.Warnings = append(plan.Warnings, "Untracked items are moved to .clustta-agent-trash for recovery rather than permanently erased.")
	return plan, nil
}

func executeDelete(ctx context.Context, projectPath string, plan planning.Plan) (planning.Result, error) {
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
	workingDir, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return planning.Result{}, err
	}
	changes := append([]planning.Change(nil), plan.Changes...)
	sort.SliceStable(changes, func(i, j int) bool { return changes[i].Entity.Depth > changes[j].Entity.Depth })
	journal := []renameMove{}
	rollback := func() {
		for i := len(journal) - 1; i >= 0; i-- {
			_ = utils.RenamePathCaseSafe(journal[i].newPath, journal[i].oldPath)
		}
	}
	result := planning.Result{PlanID: plan.ID, Command: plan.Command, LocalOnly: true, RequiresSync: true}
	for _, change := range changes {
		if err := ctx.Err(); err != nil {
			rollback()
			return result, err
		}
		entity := change.Entity
		switch entity.Type {
		case scope.TypeAsset:
			err = repository.DeleteAsset(tx, entity.ID, false, true)
		case scope.TypeCollection:
			err = repository.DeleteCollection(tx, entity.ID, false, true)
		case scope.TypeUntrackedAsset, scope.TypeUntrackedCollection:
			relative, relErr := filepath.Rel(workingDir, entity.Path)
			if relErr != nil {
				err = relErr
				break
			}
			trashPath := filepath.Join(workingDir, ".clustta-agent-trash", plan.ID, relative)
			if mkdirErr := os.MkdirAll(filepath.Dir(trashPath), os.ModePerm); mkdirErr != nil {
				err = mkdirErr
				break
			}
			err = utils.RenamePathCaseSafe(entity.Path, trashPath)
			if err == nil {
				journal = append(journal, renameMove{oldPath: entity.Path, newPath: trashPath})
				change.After["recovery_path"] = trashPath
			}
		}
		if err != nil {
			rollback()
			return planning.Result{}, fmt.Errorf("delete %s: %w", entity.Name, err)
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
