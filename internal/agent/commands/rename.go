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
	"strconv"
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

type renameNumbering struct {
	start     int
	step      int
	padding   int
	position  string
	separator string
}

type renameNameMapping struct {
	entityID string
	oldName  string
	newName  string
}

type renameOptions struct {
	format       string
	newName      string
	prependText  string
	appendText   string
	findText     string
	replaceText  string
	removePrefix string
	removeSuffix string
	template     string
	nameMappings []renameNameMapping
	numbering    *renameNumbering
}

func init() {
	Register(Definition{
		Name:        "batch_rename",
		Description: "Rename assets and collections using formats, prepend/append text, find and replace, exact name mappings, sequential numbers, templates, or prefix/suffix removal. Supports tracked and selected untracked entities. Local-only; manual sync required.",
		Permission:  "update_asset", Risk: "destructive",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"scope":         ScopeSchema([]string{"asset", "collection", "untracked_asset", "untracked_collection"}),
				"format":        map[string]interface{}{"type": "string", "enum": []string{"camelCase", "PascalCase", "snake_case", "kebab-case", "lowercase", "UPPERCASE"}},
				"new_name":      map[string]interface{}{"type": "string", "description": "Explicit new name; only valid when scope resolves to one entity and cannot be combined with other rename rules."},
				"prepend_text":  map[string]interface{}{"type": "string", "description": "Text to add before every name."},
				"append_text":   map[string]interface{}{"type": "string", "description": "Text to add after every name."},
				"find_text":     map[string]interface{}{"type": "string", "description": "Exact case-sensitive text to replace in every name."},
				"replace_text":  map[string]interface{}{"type": "string", "description": "Replacement for find_text. May be empty to remove matches."},
				"remove_prefix": map[string]interface{}{"type": "string", "description": "Exact case-sensitive prefix to remove when present."},
				"remove_suffix": map[string]interface{}{"type": "string", "description": "Exact case-sensitive suffix to remove when present."},
				"template":      map[string]interface{}{"type": "string", "description": "Naming template containing {name} and optionally {number}."},
				"name_mappings": map[string]interface{}{
					"type":        "array",
					"description": "Exact old-name/new-name mappings. This rule cannot be combined with other rename rules.",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"entity_id": map[string]interface{}{"type": "string", "description": "Optional entity ID for unambiguous mappings when names repeat."},
							"old_name":  map[string]interface{}{"type": "string"},
							"new_name":  map[string]interface{}{"type": "string"},
						},
						"required": []string{"old_name", "new_name"},
					},
				},
				"numbering": map[string]interface{}{
					"type":        "object",
					"description": "Sequential numbering in deterministic scope order. Defaults to start 1, step 1, suffix position, and '-' separator.",
					"properties": map[string]interface{}{
						"start":     map[string]interface{}{"type": "integer"},
						"step":      map[string]interface{}{"type": "integer"},
						"padding":   map[string]interface{}{"type": "integer", "minimum": 0},
						"position":  map[string]interface{}{"type": "string", "enum": []string{"prefix", "suffix"}},
						"separator": map[string]interface{}{"type": "string"},
					},
				},
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
	options, err := parseRenameOptions(args)
	if err != nil {
		return planning.Plan{}, err
	}
	if options.newName != "" && len(resolved.Entities) != 1 {
		return planning.Plan{}, fmt.Errorf("new_name requires a scope containing exactly one entity")
	}

	plan := newPlan("batch_rename", resolved)
	plan.Options = args
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
	for index, entity := range resolved.Entities {
		newName, matched, renameErr := options.apply(entity, index)
		if renameErr != nil {
			return planning.Plan{}, renameErr
		}
		if !matched {
			continue
		}
		plannedTargets[string(entity.Type)+":"+entity.ID] = normalizedPath(renameTargetPath(entity, newName))
		plannedSources[normalizedPath(entity.Path)] = string(entity.Type) + ":" + entity.ID
	}
	targets := map[string]string{}
	for index, entity := range resolved.Entities {
		newName, matched, renameErr := options.apply(entity, index)
		if renameErr != nil {
			return planning.Plan{}, renameErr
		}
		change := planning.Change{
			Entity: entity, Action: "Rename", Valid: true,
			Before: map[string]interface{}{"name": entity.Name, "path": entity.Path},
			After:  map[string]interface{}{"name": newName},
		}
		if !matched {
			change.Valid = false
			change.Warnings = append(change.Warnings, "no old-name mapping matched")
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

func parseRenameOptions(args map[string]interface{}) (renameOptions, error) {
	options := renameOptions{
		format:       stringArg(args, "format"),
		newName:      stringArg(args, "new_name"),
		prependText:  stringArg(args, "prepend_text"),
		appendText:   stringArg(args, "append_text"),
		findText:     stringArg(args, "find_text"),
		replaceText:  stringValue(args, "replace_text"),
		removePrefix: stringArg(args, "remove_prefix"),
		removeSuffix: stringArg(args, "remove_suffix"),
		template:     stringArg(args, "template"),
	}

	mappings, err := parseNameMappings(args["name_mappings"])
	if err != nil {
		return renameOptions{}, err
	}
	options.nameMappings = mappings
	if rawNumbering, ok := args["numbering"]; ok {
		options.numbering, err = parseRenameNumbering(rawNumbering)
		if err != nil {
			return renameOptions{}, err
		}
	}

	hasComposableRule := options.format != "" || options.prependText != "" || options.appendText != "" ||
		options.findText != "" || options.removePrefix != "" || options.removeSuffix != "" ||
		options.template != "" || options.numbering != nil
	if options.newName == "" && len(options.nameMappings) == 0 && !hasComposableRule {
		return renameOptions{}, fmt.Errorf("at least one rename rule is required")
	}
	if options.newName != "" && (len(options.nameMappings) > 0 || hasComposableRule) {
		return renameOptions{}, fmt.Errorf("new_name cannot be combined with other rename rules")
	}
	if len(options.nameMappings) > 0 && hasComposableRule {
		return renameOptions{}, fmt.Errorf("name_mappings cannot be combined with other rename rules")
	}
	if _, hasReplace := args["replace_text"]; hasReplace && options.findText == "" {
		return renameOptions{}, fmt.Errorf("find_text is required when replace_text is provided")
	}
	if options.template != "" && !strings.Contains(options.template, "{name}") && !strings.Contains(options.template, "{number}") {
		return renameOptions{}, fmt.Errorf("template must contain {name} or {number}")
	}
	if strings.Contains(options.template, "{number}") && options.numbering == nil {
		return renameOptions{}, fmt.Errorf("numbering is required when template contains {number}")
	}
	return options, nil
}

func parseNameMappings(raw interface{}) ([]renameNameMapping, error) {
	mappings := []renameNameMapping{}
	if raw == nil {
		return mappings, nil
	}
	items, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("name_mappings must be an array")
	}
	for index, item := range items {
		mapping, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("name_mappings[%d] must be an object", index)
		}
		parsed := renameNameMapping{
			entityID: stringArg(mapping, "entity_id"),
			oldName:  stringArg(mapping, "old_name"),
			newName:  stringArg(mapping, "new_name"),
		}
		if parsed.oldName == "" || parsed.newName == "" {
			return nil, fmt.Errorf("name_mappings[%d] requires old_name and new_name", index)
		}
		for _, existing := range mappings {
			if parsed.entityID != "" && parsed.entityID == existing.entityID {
				return nil, fmt.Errorf("duplicate entity_id %q in name_mappings", parsed.entityID)
			}
			if parsed.entityID == "" && existing.entityID == "" && parsed.oldName == existing.oldName {
				return nil, fmt.Errorf("duplicate old_name %q in name_mappings", parsed.oldName)
			}
		}
		mappings = append(mappings, parsed)
	}
	return mappings, nil
}

func parseRenameNumbering(raw interface{}) (*renameNumbering, error) {
	values, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("numbering must be an object")
	}
	numbering := &renameNumbering{start: 1, step: 1, position: "suffix", separator: "-"}
	var err error
	if _, exists := values["start"]; exists {
		numbering.start, err = renameInteger(values, "start")
		if err != nil {
			return nil, err
		}
	}
	if _, exists := values["step"]; exists {
		numbering.step, err = renameInteger(values, "step")
		if err != nil {
			return nil, err
		}
	}
	if _, exists := values["padding"]; exists {
		numbering.padding, err = renameInteger(values, "padding")
		if err != nil {
			return nil, err
		}
	}
	if numbering.padding < 0 {
		return nil, fmt.Errorf("numbering.padding cannot be negative")
	}
	if numbering.step == 0 {
		return nil, fmt.Errorf("numbering.step cannot be zero")
	}
	if position := stringArg(values, "position"); position != "" {
		if position != "prefix" && position != "suffix" {
			return nil, fmt.Errorf("numbering.position must be prefix or suffix")
		}
		numbering.position = position
	}
	if separator, exists := values["separator"]; exists {
		text, ok := separator.(string)
		if !ok {
			return nil, fmt.Errorf("numbering.separator must be a string")
		}
		numbering.separator = text
	}
	return numbering, nil
}

func renameInteger(values map[string]interface{}, key string) (int, error) {
	switch value := values[key].(type) {
	case int:
		return value, nil
	case int64:
		return int(value), nil
	case float64:
		integer := int(value)
		if float64(integer) == value {
			return integer, nil
		}
	}
	return 0, fmt.Errorf("numbering.%s must be an integer", key)
}

func stringValue(values map[string]interface{}, key string) string {
	value, _ := values[key].(string)
	return value
}

func (options renameOptions) apply(entity scope.Entity, index int) (string, bool, error) {
	if options.newName != "" {
		return options.newName, true, nil
	}
	if len(options.nameMappings) > 0 {
		for _, mapping := range options.nameMappings {
			if mapping.entityID != "" && mapping.entityID == entity.ID {
				return mapping.newName, true, nil
			}
		}
		for _, mapping := range options.nameMappings {
			if mapping.entityID == "" && mapping.oldName == entity.Name {
				return mapping.newName, true, nil
			}
		}
		return entity.Name, false, nil
	}

	name := entity.Name
	if options.format != "" {
		formatted, err := formatEntityName(name, options.format)
		if err != nil {
			return "", false, err
		}
		name = formatted
	}
	if options.findText != "" {
		name = strings.ReplaceAll(name, options.findText, options.replaceText)
	}
	name = strings.TrimPrefix(name, options.removePrefix)
	name = strings.TrimSuffix(name, options.removeSuffix)
	name = options.prependText + name + options.appendText

	number := ""
	if options.numbering != nil {
		value := options.numbering.start + index*options.numbering.step
		number = strconv.Itoa(value)
		if options.numbering.padding > 0 {
			number = fmt.Sprintf("%0*d", options.numbering.padding, value)
		}
	}
	if options.template != "" {
		name = strings.ReplaceAll(options.template, "{name}", name)
		name = strings.ReplaceAll(name, "{number}", number)
	} else if options.numbering != nil {
		if options.numbering.position == "prefix" {
			name = number + options.numbering.separator + name
		} else {
			name += options.numbering.separator + number
		}
	}
	return name, true, nil
}

func preserveSelectedRenameTargets(args map[string]interface{}, changes []planning.Change) {
	for _, key := range []string{
		"format", "new_name", "prepend_text", "append_text", "find_text", "replace_text",
		"remove_prefix", "remove_suffix", "template", "name_mappings", "numbering",
	} {
		delete(args, key)
	}
	mappings := make([]interface{}, 0, len(changes))
	for _, change := range changes {
		newName, _ := change.After["name"].(string)
		mappings = append(mappings, map[string]interface{}{
			"entity_id": change.Entity.ID,
			"old_name":  change.Entity.Name,
			"new_name":  newName,
		})
	}
	args["name_mappings"] = mappings
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
