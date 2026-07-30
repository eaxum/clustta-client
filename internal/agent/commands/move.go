package commands

import (
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"clustta/internal/repository"
	"clustta/internal/utils"
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func init() {
	Register(Definition{
		Name:        "batch_move",
		Description: "Move assets and collections to a tracked collection or project-relative folder. Supports tracked and untracked entities. Local-only; manual sync required.",
		Permission:  "update_asset", Risk: "destructive",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"scope":                ScopeSchema([]string{"asset", "collection", "untracked_asset", "untracked_collection"}),
				"target_collection_id": map[string]interface{}{"type": "string", "description": "Tracked destination collection ID; omit for project root."},
				"target_path":          map[string]interface{}{"type": "string", "description": "Optional project-relative destination folder, primarily for an untracked destination."},
			},
			"required": []string{"scope"},
		},
		Plan:           planMove,
		ExecuteContext: executeMove,
	})
}

func planMove(projectPath string, args map[string]interface{}) (planning.Plan, error) {
	req, err := ParseScope(args, []scope.EntityType{scope.TypeAsset, scope.TypeCollection, scope.TypeUntrackedAsset, scope.TypeUntrackedCollection})
	if err != nil {
		return planning.Plan{}, err
	}
	resolved, err := scope.Resolve(projectPath, req)
	if err != nil {
		return planning.Plan{}, err
	}
	targetID := stringArg(args, "target_collection_id")
	targetPathArg := stringArg(args, "target_path")

	db, err := utils.OpenDb(projectPath)
	if err != nil {
		return planning.Plan{}, err
	}
	tx, err := db.Beginx()
	if err != nil {
		db.Close()
		return planning.Plan{}, err
	}
	workingDir, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		tx.Rollback()
		db.Close()
		return planning.Plan{}, err
	}
	targetFolder := workingDir
	if targetID != "" {
		target, targetErr := repository.GetCollection(tx, targetID)
		if targetErr != nil {
			tx.Rollback()
			db.Close()
			return planning.Plan{}, fmt.Errorf("target collection not found")
		}
		targetFolder = target.FilePath
	}
	tx.Rollback()
	db.Close()
	if targetPathArg != "" {
		targetFolder, err = safeProjectTarget(workingDir, targetPathArg)
		if err != nil {
			return planning.Plan{}, err
		}
	}

	plan := newPlan("batch_move", resolved)
	plan.Options = map[string]interface{}{"target_collection_id": targetID, "target_folder": targetFolder}
	targets := map[string]string{}
	movingCollectionPaths := []string{}
	for _, entity := range resolved.Entities {
		if entity.Type == scope.TypeCollection || entity.Type == scope.TypeUntrackedCollection {
			movingCollectionPaths = append(movingCollectionPaths, filepath.Clean(entity.Path))
		}
	}
	for _, entity := range resolved.Entities {
		change := planning.Change{
			Entity: entity, Action: "Move", Valid: true,
			Before: map[string]interface{}{"parent_id": entity.ParentID, "path": entity.Path},
			After:  map[string]interface{}{"parent_id": targetID},
		}
		for _, ancestorPath := range movingCollectionPaths {
			if filepath.Clean(entity.Path) != ancestorPath && isPathInside(entity.Path, ancestorPath) {
				change.Valid = false
				change.Warnings = append(change.Warnings, "entity moves with its selected ancestor collection")
				break
			}
		}
		if entity.Type.Tracked() && targetPathArg != "" && targetID == "" && filepath.Clean(targetFolder) != filepath.Clean(workingDir) {
			change.Valid = false
			change.Errors = append(change.Errors, "tracked entities require a tracked target collection")
		}
		targetPath := filepath.Join(targetFolder, filepath.Base(entity.Path))
		change.After["path"] = targetPath
		if entity.Type == scope.TypeCollection || entity.Type == scope.TypeUntrackedCollection {
			if isPathInside(targetFolder, entity.Path) {
				change.Valid = false
				change.Errors = append(change.Errors, "collection cannot be moved into itself or a descendant")
			}
		}
		if filepath.Clean(entity.Path) == filepath.Clean(targetPath) {
			change.Valid = false
			change.Warnings = append(change.Warnings, "entity is already in the target")
		}
		key := normalizedPath(targetPath)
		if prior := targets[key]; prior != "" {
			change.Valid = false
			change.Errors = append(change.Errors, "target conflicts with "+prior)
		} else {
			targets[key] = entity.Name
		}
		if !strings.EqualFold(entity.Path, targetPath) && (utils.FileExists(targetPath) || utils.DirExists(targetPath)) {
			change.Valid = false
			change.Errors = append(change.Errors, "target path already exists")
		}
		addChange(&plan, change)
	}
	if len(plan.Changes) == 0 {
		plan.Errors = append(plan.Errors, "scope resolved to no movable entities")
	}
	return plan, nil
}

func executeMove(ctx context.Context, projectPath string, plan planning.Plan) (planning.Result, error) {
	changes := append([]planning.Change(nil), plan.Changes...)
	sort.SliceStable(changes, func(i, j int) bool { return changes[i].Entity.Depth > changes[j].Entity.Depth })
	targetID, _ := plan.Options["target_collection_id"].(string)
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
		if !change.Valid {
			result.Skipped++
			result.Items = append(result.Items, resultItem(change, "skipped"))
			continue
		}
		oldPath, _ := change.Before["path"].(string)
		newPath, _ := change.After["path"].(string)
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
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("target path is outside the project")
	}
	return targetAbs, nil
}

func isPathInside(target, source string) bool {
	rel, err := filepath.Rel(filepath.Clean(source), filepath.Clean(target))
	return err == nil && (rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))))
}
