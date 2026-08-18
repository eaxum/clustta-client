package commands

import (
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"clustta/internal/auth_service"
	"clustta/internal/repository"
	"clustta/internal/utils"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type Definition struct {
	Name           string                                                                `json:"name"`
	Description    string                                                                `json:"description"`
	Parameters     map[string]interface{}                                                `json:"parameters"`
	Permission     string                                                                `json:"-"`
	Risk           string                                                                `json:"-"`
	Plan           func(string, map[string]interface{}) (planning.Plan, error)           `json:"-"`
	Execute        func(string, planning.Plan) (planning.Result, error)                  `json:"-"`
	ExecuteContext func(context.Context, string, planning.Plan) (planning.Result, error) `json:"-"`
	Direct         func(string, map[string]interface{}) (interface{}, error)             `json:"-"`
	Select         func(map[string]interface{}, planning.Plan, []string) error           `json:"-"`
}

var (
	registryMu sync.RWMutex
	registry   = map[string]Definition{}
	planStore  = planning.NewStore()
)

func Register(def Definition) {
	if def.Name == "" || (def.Direct == nil && (def.Plan == nil || (def.Execute == nil && def.ExecuteContext == nil))) {
		panic("agent command requires name and either a direct handler or planner/executor")
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := registry[def.Name]; exists {
		panic("duplicate agent command: " + def.Name)
	}
	registry[def.Name] = def
}

func ExecuteDirect(projectPath, name string, args map[string]interface{}) (interface{}, error) {
	def, ok := DefinitionFor(name)
	if !ok || def.Direct == nil {
		return nil, fmt.Errorf("command %q is not directly executable", name)
	}
	return def.Direct(projectPath, args)
}

func Definitions() []Definition {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]Definition, 0, len(registry))
	for _, def := range registry {
		out = append(out, def)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func DefinitionFor(name string) (Definition, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	def, ok := registry[name]
	return def, ok
}

// Prepare resolves and stores an immutable plan. It mutates args only to bind
// the approved tool call to the stored plan ID.
func Prepare(projectPath, name string, args map[string]interface{}) (planning.Plan, error) {
	def, ok := DefinitionFor(name)
	if !ok {
		return planning.Plan{}, fmt.Errorf("unknown command %q", name)
	}
	plan, err := def.Plan(projectPath, args)
	if err != nil {
		return planning.Plan{}, err
	}
	plan.ID = uuid.NewString()
	plan.Command = name
	plan.Fingerprint = fingerprint(plan.Scope)
	if err := authorize(projectPath, name, plan); err != nil {
		return planning.Plan{}, err
	}
	planStore.Put(plan)
	args["_plan_id"] = plan.ID
	return plan, nil
}

// PrepareSelection replaces a prepared plan with a newly validated plan scoped
// to the entities the user kept checked in the approval modal.
func PrepareSelection(projectPath, name string, args map[string]interface{}, selectedKeys []string) error {
	id, _ := args["_plan_id"].(string)
	approved, ok := planStore.Take(id)
	if !ok || approved.Command != name {
		return fmt.Errorf("approved plan expired; please retry")
	}
	def, _ := DefinitionFor(name)
	if def.Select != nil {
		if err := def.Select(args, approved, selectedKeys); err != nil {
			return err
		}
		delete(args, "_plan_id")
		_, err := Prepare(projectPath, name, args)
		return err
	}

	selected := make(map[string]struct{}, len(selectedKeys))
	for _, key := range selectedKeys {
		selected[key] = struct{}{}
	}
	entities := make([]scope.Entity, 0, len(selected))
	selectedChanges := make([]planning.Change, 0, len(selected))
	for _, change := range approved.Changes {
		key := changeSelectionKey(change)
		if _, ok := selected[key]; ok && change.Valid {
			entities = append(entities, change.Entity)
			selectedChanges = append(selectedChanges, change)
		}
	}
	if len(entities) == 0 {
		return fmt.Errorf("no items selected")
	}

	args["scope"] = scope.Request{
		Source: "selection", Selection: entities,
		Types: approved.Scope.Request.Types,
	}
	if name == "batch_rename" {
		preserveSelectedRenameTargets(args, selectedChanges)
	}
	delete(args, "_plan_id")
	_, err := Prepare(projectPath, name, args)
	return err
}

func changeSelectionKey(change planning.Change) string {
	if change.Key != "" {
		return change.Key
	}
	return string(change.Entity.Type) + ":" + change.Entity.ID
}

func authorize(projectPath, name string, plan planning.Plan) error {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return fmt.Errorf("could not determine current user: %w", err)
	}
	db, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	projectUser, err := repository.GetUser(tx, user.Id)
	if err != nil {
		return err
	}
	role, err := repository.GetRole(tx, projectUser.RoleId)
	if err != nil {
		return err
	}
	for _, change := range plan.Changes {
		switch name {
		case "batch_create_collections", "apply_workflow", "setup_animation_production":
			if !role.CreateCollection {
				return fmt.Errorf("permission denied: role %q cannot create collections", role.Name)
			}
		case "batch_create_assets", "batch_create_asset_types", "batch_create_collection_types",
			"batch_update_asset_types", "batch_update_collection_types", "setup_project_types":
			if !role.CreateAsset {
				return fmt.Errorf("permission denied: role %q cannot create assets or maintain types", role.Name)
			}
		case "batch_change_status":
			if !role.ChangeStatus {
				return fmt.Errorf("permission denied: role %q cannot change status", role.Name)
			}
		case "batch_assign", "batch_distribute":
			if change.Entity.Type == scope.TypeAsset && !role.AssignAsset {
				return fmt.Errorf("permission denied: role %q cannot assign assets", role.Name)
			}
			if change.Entity.Type == scope.TypeCollection && !role.UpdateCollection {
				return fmt.Errorf("permission denied: role %q cannot update collections", role.Name)
			}
		case "batch_unassign":
			if change.Entity.Type == scope.TypeAsset && !role.UnassignAsset {
				return fmt.Errorf("permission denied: role %q cannot unassign assets", role.Name)
			}
			if change.Entity.Type == scope.TypeCollection && !role.UpdateCollection {
				return fmt.Errorf("permission denied: role %q cannot update collections", role.Name)
			}
		case "batch_add_dependency", "batch_remove_dependency":
			if !role.ManageDependencies {
				return fmt.Errorf("permission denied: role %q cannot manage dependencies", role.Name)
			}
		case "batch_delete":
			if (change.Entity.Type == scope.TypeAsset || change.Entity.Type == scope.TypeUntrackedAsset) && !role.DeleteAsset {
				return fmt.Errorf("permission denied: role %q cannot delete assets", role.Name)
			}
			if (change.Entity.Type == scope.TypeCollection || change.Entity.Type == scope.TypeUntrackedCollection) && !role.DeleteCollection {
				return fmt.Errorf("permission denied: role %q cannot delete collections", role.Name)
			}
		default:
			if (change.Entity.Type == scope.TypeAsset || change.Entity.Type == scope.TypeUntrackedAsset) && !role.UpdateAsset {
				return fmt.Errorf("permission denied: role %q cannot update assets", role.Name)
			}
			if (change.Entity.Type == scope.TypeCollection || change.Entity.Type == scope.TypeUntrackedCollection) && !role.UpdateCollection {
				return fmt.Errorf("permission denied: role %q cannot update collections", role.Name)
			}
		}
	}
	return nil
}

func Verify(projectPath, name string, args map[string]interface{}) error {
	id, _ := args["_plan_id"].(string)
	approved, ok := planStore.Get(id)
	if !ok || approved.Command != name {
		return fmt.Errorf("approved plan expired; please retry")
	}
	current, err := scope.Resolve(projectPath, approved.Scope.Request)
	if err != nil {
		return err
	}
	if fingerprint(current) != approved.Fingerprint {
		return fmt.Errorf("scope changed since approval; please retry")
	}
	return nil
}

func ExecutePrepared(ctx context.Context, projectPath, name string, args map[string]interface{}) (planning.Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return planning.Result{}, err
	}
	id, _ := args["_plan_id"].(string)
	plan, ok := planStore.Take(id)
	if !ok || plan.Command != name {
		return planning.Result{}, fmt.Errorf("approved plan expired; please retry")
	}
	if !plan.Executable() {
		return planning.Result{}, fmt.Errorf("plan contains validation errors")
	}
	def, _ := DefinitionFor(name)
	if def.ExecuteContext != nil {
		return def.ExecuteContext(ctx, projectPath, plan)
	}
	return def.Execute(projectPath, plan)
}

func ParseScope(args map[string]interface{}, defaultTypes []scope.EntityType) (scope.Request, error) {
	return ParseNamedScope(args, "scope", defaultTypes)
}

func ParseNamedScope(args map[string]interface{}, key string, defaultTypes []scope.EntityType) (scope.Request, error) {
	raw, ok := args[key]
	if !ok {
		return scope.Request{}, fmt.Errorf("%s is required", key)
	}
	// Accept the frontend entity-envelope key as a compatibility alias. The
	// canonical command schema remains entity_id.
	if values, ok := raw.(map[string]interface{}); ok {
		if _, exists := values["types"]; !exists {
			if entityTypes, exists := values["type"]; exists {
				if text, ok := entityTypes.(string); ok {
					values["types"] = []string{text}
				} else {
					values["types"] = entityTypes
				}
				delete(values, "type")
			}
		}
		if _, exists := values["entity_id"]; !exists {
			if id, exists := values["id"]; exists {
				values["entity_id"] = id
			}
		}
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return scope.Request{}, err
	}
	var req scope.Request
	if err := json.Unmarshal(data, &req); err != nil {
		return scope.Request{}, fmt.Errorf("invalid %s: %w", key, err)
	}
	if len(req.Types) == 0 {
		req.Types = append([]scope.EntityType(nil), defaultTypes...)
	}
	if err := validateScopeFilters(req); err != nil {
		return scope.Request{}, err
	}
	return req, nil
}

func validateScopeFilters(req scope.Request) error {
	if len(req.Filters) == 0 || len(req.Types) == 0 {
		return nil
	}
	hasAsset, hasCollection := false, false
	for _, entityType := range req.Types {
		hasAsset = hasAsset || entityType == scope.TypeAsset || entityType == scope.TypeUntrackedAsset
		hasCollection = hasCollection || entityType == scope.TypeCollection || entityType == scope.TypeUntrackedCollection
	}
	if hasAsset && !hasCollection {
		for _, key := range []string{"collection_type", "collection_type_id"} {
			if value, exists := req.Filters[key]; exists && scopeFilterHasValue(value) {
				return fmt.Errorf("filter %q is incompatible with asset scope", key)
			}
		}
	}
	if hasCollection && !hasAsset {
		for _, key := range []string{"asset_type", "asset_type_id", "status", "status_id", "is_resource"} {
			if value, exists := req.Filters[key]; exists && scopeFilterHasValue(value) {
				return fmt.Errorf("filter %q is incompatible with collection scope", key)
			}
		}
	}
	return nil
}

func scopeFilterHasValue(value interface{}) bool {
	if value == nil {
		return false
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text) != ""
	}
	return true
}

func fingerprint(result scope.Result) string {
	data, _ := json.Marshal(result)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func ScopeSchema(typeEnums []string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "object",
		"description": "Structured target scope. Use selection for browser-selected items, here for the current collection, entity or entities for explicit resolved references, or project for the whole project.",
		"properties": map[string]interface{}{
			"source":    map[string]interface{}{"type": "string", "enum": []string{"selection", "here", "entity", "entities", "project"}},
			"entity_id": map[string]interface{}{"type": "string"},
			"entity_ids": map[string]interface{}{
				"type": "array", "items": map[string]interface{}{"type": "string"},
				"description": "Explicit IDs returned by query_entities. Use with source entities for targets across locations.",
			},
			"path":      map[string]interface{}{"type": "string"},
			"recursive": map[string]interface{}{"type": "boolean"},
			"types":     map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string", "enum": typeEnums}},
			"selection": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"type":          map[string]interface{}{"type": "string", "enum": typeEnums},
						"id":            map[string]interface{}{"type": "string"},
						"name":          map[string]interface{}{"type": "string"},
						"path":          map[string]interface{}{"type": "string"},
						"collection_id": map[string]interface{}{"type": "string"},
						"extension":     map[string]interface{}{"type": "string"},
						"parent_id":     map[string]interface{}{"type": "string"},
						"parent_path":   map[string]interface{}{"type": "string"},
						"metadata":      map[string]interface{}{"type": "object"},
					},
					"required": []string{"type", "id", "name"},
				},
			},
			"filters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":               map[string]interface{}{"type": "string"},
					"status":             map[string]interface{}{"type": "string"},
					"status_id":          map[string]interface{}{"type": "string"},
					"asset_type":         map[string]interface{}{"type": "string"},
					"asset_type_id":      map[string]interface{}{"type": "string"},
					"collection_type":    map[string]interface{}{"type": "string"},
					"collection_type_id": map[string]interface{}{"type": "string"},
					"assignee_id":        map[string]interface{}{"type": "string"},
					"unassigned":         map[string]interface{}{"type": "boolean"},
					"is_resource":        map[string]interface{}{"type": "boolean"},
					"extension":          map[string]interface{}{"type": "string"},
					"tag":                map[string]interface{}{"type": "string"},
					"state":              map[string]interface{}{"type": "string"},
				},
			},
			"limit": map[string]interface{}{"type": "integer"},
		},
		"required": []string{"source"},
	}
}
