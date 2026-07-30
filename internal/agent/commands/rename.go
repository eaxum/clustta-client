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
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var wordBoundary = regexp.MustCompile(`[^A-Za-z0-9]+`)

type renameMove struct {
	oldPath string
	newPath string
}

func init() {
	Register(Definition{
		Name:        "batch_rename",
		Description: "Rename assets and collections in a structured scope using a deterministic naming format or one explicit name. Supports tracked and selected untracked entities. Local-only; manual sync required.",
		Permission:  "update_asset", Risk: "destructive",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"scope":    ScopeSchema([]string{"asset", "collection", "untracked_asset", "untracked_collection"}),
				"format":   map[string]interface{}{"type": "string", "enum": []string{"camelCase", "PascalCase", "snake_case", "kebab-case", "lowercase", "UPPERCASE"}},
				"new_name": map[string]interface{}{"type": "string", "description": "Explicit new name; only valid when scope resolves to one entity."},
			},
			"required": []string{"scope"},
		},
		Plan:           planRename,
		ExecuteContext: executeRename,
	})
}

func planRename(projectPath string, args map[string]interface{}) (planning.Plan, error) {
	req, err := ParseScope(args, []scope.EntityType{scope.TypeAsset, scope.TypeCollection, scope.TypeUntrackedAsset, scope.TypeUntrackedCollection})
	if err != nil {
		return planning.Plan{}, err
	}
	resolved, err := scope.Resolve(projectPath, req)
	if err != nil {
		return planning.Plan{}, err
	}
	format := stringArg(args, "format")
	explicit := stringArg(args, "new_name")
	if format == "" && explicit == "" {
		return planning.Plan{}, fmt.Errorf("format or new_name is required")
	}
	if explicit != "" && len(resolved.Entities) != 1 {
		return planning.Plan{}, fmt.Errorf("new_name requires a scope containing exactly one entity")
	}

	plan := newPlan("batch_rename", resolved)
	plan.Options = map[string]interface{}{"format": format, "new_name": explicit}
	inventory, err := scope.Resolve(projectPath, scope.Request{
		Source: "project", Recursive: true,
		Types: []scope.EntityType{scope.TypeAsset, scope.TypeCollection, scope.TypeUntrackedAsset, scope.TypeUntrackedCollection},
	})
	if err != nil {
		return planning.Plan{}, err
	}
	pathOwners := map[string]scope.Entity{}
	for _, entity := range inventory.Entities {
		if entity.Path != "" {
			pathOwners[normalizedPath(entity.Path)] = entity
		}
	}
	plannedTargets := map[string]string{}
	plannedSources := map[string]string{}
	for _, entity := range resolved.Entities {
		newName := explicit
		if newName == "" {
			newName, err = formatEntityName(entity.Name, format)
			if err != nil {
				return planning.Plan{}, err
			}
		}
		plannedTargets[string(entity.Type)+":"+entity.ID] = normalizedPath(renameTargetPath(entity, newName))
		plannedSources[normalizedPath(entity.Path)] = string(entity.Type) + ":" + entity.ID
	}
	targets := map[string]string{}
	for _, entity := range resolved.Entities {
		newName := explicit
		if newName == "" {
			newName, err = formatEntityName(entity.Name, format)
			if err != nil {
				return planning.Plan{}, err
			}
		}
		change := planning.Change{
			Entity: entity, Action: "Rename", Valid: true,
			Before: map[string]interface{}{"name": entity.Name, "path": entity.Path},
			After:  map[string]interface{}{"name": newName},
		}
		if strings.TrimSpace(newName) == "" {
			change.Valid = false
			change.Errors = append(change.Errors, "formatted name is empty")
		}
		if newName == entity.Name {
			change.Valid = false
			change.Warnings = append(change.Warnings, "name is already in the requested format")
		}
		targetPath := renameTargetPath(entity, newName)
		change.After["path"] = targetPath
		key := collisionKey(entity, targetPath)
		if prior, exists := targets[key]; exists {
			change.Valid = false
			change.Errors = append(change.Errors, fmt.Sprintf("target conflicts with %s", prior))
		} else {
			targets[key] = entity.Name
		}
		if targetPath != "" && normalizedPath(entity.Path) != normalizedPath(targetPath) {
			if owner, exists := pathOwners[normalizedPath(targetPath)]; exists {
				ownerKey := string(owner.Type) + ":" + owner.ID
				if ownerKey == string(entity.Type)+":"+entity.ID {
					// Same entity with a case-only target.
				} else if ownerTarget, moving := plannedTargets[ownerKey]; moving && ownerTarget != normalizedPath(targetPath) {
					change.Warnings = append(change.Warnings, "rename cycle will use a temporary path")
				} else {
					change.Valid = false
					change.Errors = append(change.Errors, fmt.Sprintf("target is occupied by %s", owner.Name))
				}
			} else if _, statErr := os.Stat(targetPath); statErr == nil {
				if _, moving := plannedSources[normalizedPath(targetPath)]; !moving {
					change.Valid = false
					change.Errors = append(change.Errors, "target path already exists")
				}
			}
		}
		if newName != entity.Name && strings.EqualFold(newName, entity.Name) {
			change.Warnings = append(change.Warnings, "case-only rename will use a temporary path")
		}
		addChange(&plan, change)
	}
	if len(plan.Changes) == 0 {
		plan.Errors = append(plan.Errors, "scope resolved to no renameable entities")
	}
	return plan, nil
}

func executeRename(ctx context.Context, projectPath string, plan planning.Plan) (planning.Result, error) {
	changes := append([]planning.Change(nil), plan.Changes...)
	sort.SliceStable(changes, func(i, j int) bool {
		left, right := changes[i].Entity, changes[j].Entity
		if left.Depth != right.Depth {
			return left.Depth > right.Depth
		}
		leftRank, rightRank := renameRank(left.Type), renameRank(right.Type)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		return left.Name < right.Name
	})

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

	journal := make([]renameMove, 0, len(changes))
	staged := map[string]string{}
	rollbackFiles := func() {
		for i := len(journal) - 1; i >= 0; i-- {
			_ = utils.RenamePathCaseSafe(journal[i].newPath, journal[i].oldPath)
		}
		for oldPath, tempPath := range staged {
			if utils.FileExists(tempPath) || utils.DirExists(tempPath) {
				_ = utils.RenamePathCaseSafe(tempPath, oldPath)
			}
		}
	}

	result := planning.Result{PlanID: plan.ID, Command: plan.Command, LocalOnly: true, RequiresSync: true}
	for start := 0; start < len(changes); {
		if err := ctx.Err(); err != nil {
			rollbackFiles()
			return result, err
		}
		end := start + 1
		for end < len(changes) && changes[end].Entity.Depth == changes[start].Entity.Depth &&
			filepath.Clean(filepath.Dir(changes[end].Entity.Path)) == filepath.Clean(filepath.Dir(changes[start].Entity.Path)) {
			end++
		}
		group := changes[start:end]
		oldPaths := map[string]bool{}
		for _, change := range group {
			if change.Valid && change.Entity.Path != "" {
				oldPaths[normalizedPath(change.Entity.Path)] = true
			}
		}
		for _, change := range group {
			if !change.Valid {
				continue
			}
			targetPath, _ := change.After["path"].(string)
			oldPath := change.Entity.Path
			if oldPath == "" || !oldPaths[normalizedPath(targetPath)] || strings.EqualFold(oldPath, targetPath) {
				continue
			}
			if !utils.FileExists(oldPath) && !utils.DirExists(oldPath) {
				continue
			}
			tempPath := filepath.Join(filepath.Dir(oldPath), ".clustta-rename-"+uuid.NewString())
			if err := utils.RenamePathCaseSafe(oldPath, tempPath); err != nil {
				rollbackFiles()
				return planning.Result{}, err
			}
			staged[oldPath] = tempPath
		}
		for _, change := range group {
			if err := ctx.Err(); err != nil {
				rollbackFiles()
				return result, err
			}
			if err := applyRenameChange(tx, change, staged, &journal, &result); err != nil {
				rollbackFiles()
				return planning.Result{}, err
			}
		}
		start = end
	}
	if err := tx.Commit(); err != nil {
		rollbackFiles()
		return planning.Result{}, err
	}
	return result, nil
}

func applyRenameChange(tx *sqlx.Tx, change planning.Change, staged map[string]string, journal *[]renameMove, result *planning.Result) error {
	if !change.Valid {
		result.Skipped++
		result.Items = append(result.Items, resultItem(change, "skipped"))
		return nil
	}
	newName, _ := change.After["name"].(string)
	oldPath, _ := change.Before["path"].(string)
	targetPath, _ := change.After["path"].(string)
	movedOnDisk := oldPath != "" && utils.FileExists(oldPath) || utils.DirExists(oldPath)
	tempPath := staged[oldPath]
	if tempPath != "" {
		movedOnDisk = true
	}
	switch change.Entity.Type {
	case scope.TypeAsset:
		current, err := repository.GetAsset(tx, change.Entity.ID)
		if err != nil {
			return err
		}
		oldPath = current.FilePath
		movedOnDisk = tempPath != "" || utils.FileExists(oldPath)
		renamed, err := repository.RenameAsset(tx, change.Entity.ID, newName)
		if err != nil {
			return fmt.Errorf("rename %s: %w", change.Entity.Name, err)
		}
		targetPath = renamed.FilePath
	case scope.TypeCollection:
		current, err := repository.GetCollection(tx, change.Entity.ID)
		if err != nil {
			return err
		}
		oldPath = current.FilePath
		movedOnDisk = tempPath != "" || utils.DirExists(oldPath)
		renamed, err := repository.RenameCollection(tx, change.Entity.ID, newName)
		if err != nil {
			return fmt.Errorf("rename %s: %w", change.Entity.Name, err)
		}
		targetPath = renamed.FilePath
	case scope.TypeUntrackedAsset, scope.TypeUntrackedCollection:
		if tempPath == "" {
			if err := utils.RenamePathCaseSafe(oldPath, targetPath); err != nil {
				return fmt.Errorf("rename %s: %w", change.Entity.Name, err)
			}
		}
	}
	if tempPath != "" {
		if err := utils.RenamePathCaseSafe(tempPath, targetPath); err != nil {
			return fmt.Errorf("complete staged rename %s: %w", change.Entity.Name, err)
		}
		delete(staged, oldPath)
	}
	if movedOnDisk && oldPath != "" && targetPath != "" && oldPath != targetPath {
		*journal = append(*journal, renameMove{oldPath: oldPath, newPath: targetPath})
	}
	result.Applied++
	result.Items = append(result.Items, resultItem(change, "applied"))
	return nil
}

func renameRank(entityType scope.EntityType) int {
	switch entityType {
	case scope.TypeAsset, scope.TypeUntrackedAsset:
		return 0
	case scope.TypeCollection, scope.TypeUntrackedCollection:
		return 1
	default:
		return 2
	}
}

func normalizedPath(path string) string {
	if path == "" {
		return ""
	}
	return strings.ToLower(filepath.Clean(path))
}

func renameTargetPath(entity scope.Entity, newName string) string {
	if entity.Path == "" {
		return ""
	}
	suffix := ""
	if entity.Type == scope.TypeAsset || entity.Type == scope.TypeUntrackedAsset {
		suffix = entity.Extension
		if suffix == "" {
			suffix = filepath.Ext(entity.Path)
		}
	}
	return filepath.Join(filepath.Dir(entity.Path), newName+suffix)
}

func collisionKey(entity scope.Entity, targetPath string) string {
	if targetPath != "" {
		return strings.ToLower(filepath.Clean(targetPath))
	}
	return strings.ToLower(string(entity.Type) + ":" + entity.ParentID + ":" + entity.Name + ":" + entity.Extension)
}

func formatEntityName(name, format string) (string, error) {
	words := splitWords(name)
	if len(words) == 0 {
		return "", nil
	}
	switch format {
	case "camelCase":
		return strings.ToLower(words[0]) + joinTitle(words[1:]), nil
	case "PascalCase":
		return joinTitle(words), nil
	case "snake_case":
		return strings.ToLower(strings.Join(words, "_")), nil
	case "kebab-case":
		return strings.ToLower(strings.Join(words, "-")), nil
	case "lowercase":
		return strings.ToLower(strings.Join(words, "")), nil
	case "UPPERCASE":
		return strings.ToUpper(strings.Join(words, "")), nil
	default:
		return "", fmt.Errorf("unsupported naming format %q", format)
	}
}

func splitWords(value string) []string {
	var expanded strings.Builder
	var previous rune
	for i, current := range value {
		if i > 0 && unicode.IsUpper(current) && (unicode.IsLower(previous) || unicode.IsDigit(previous)) {
			expanded.WriteRune(' ')
		}
		expanded.WriteRune(current)
		previous = current
	}
	parts := wordBoundary.Split(expanded.String(), -1)
	out := parts[:0]
	for _, part := range parts {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func joinTitle(words []string) string {
	var out strings.Builder
	for _, word := range words {
		lower := strings.ToLower(word)
		runes := []rune(lower)
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		out.WriteString(string(runes))
	}
	return out.String()
}
