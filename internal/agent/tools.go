package agent

import (
	"clustta/internal/repository"
	"clustta/internal/utils"
	"encoding/json"
	"fmt"
	"strings"
)

// ToolDefinition describes a tool the LLM can call.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolResult holds the outcome of a tool execution.
type ToolResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// GetToolDefinitions returns all tool schemas the LLM can call.
func GetToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		// Query tools
		{
			Name:        "list_collections",
			Description: "List all collections (entities) in the project. Returns collection names, IDs, types, and hierarchy.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "list_assets_in_collection",
			Description: "List all assets (tasks) in a specific collection. Returns asset names, IDs, types, statuses, and assignees.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection_id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the collection to list assets from.",
					},
				},
				"required": []string{"collection_id"},
			},
		},
		{
			Name:        "get_asset_details",
			Description: "Get detailed information about a specific asset by ID.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the asset to retrieve.",
					},
				},
				"required": []string{"asset_id"},
			},
		},
		{
			Name:        "list_users",
			Description: "List all users/collaborators in the project.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "list_statuses",
			Description: "List all available statuses in the project (e.g., Todo, In Progress, Review, Done).",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "list_task_types",
			Description: "List all available asset/task types in the project (e.g., Model, Rig, Animation).",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "list_tags",
			Description: "List all tags in the project.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "list_templates",
			Description: "List all file templates available in the project.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "search_knowledge",
			Description: "Search the Clustta knowledge base for information about features and how-tos. Use this when the user asks about how something works, how to do something, or what a concept means. Query should be a topic keyword like 'dependencies', 'sync', 'checkpoints', 'collaborators', 'integrations', 'workflows', 'templates', 'tags', 'statuses', etc.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Topic keyword to search for (e.g., 'dependencies', 'sync', 'checkpoints').",
					},
				},
				"required": []string{"query"},
			},
		},

		// CRUD tools
		{
			Name:        "create_asset_type",
			Description: "Create a new asset type (e.g., Model, Rig, Animation, Character, Prop, Shot). Use list_task_types first to check what already exists.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the asset type to create.",
					},
					"icon": map[string]interface{}{
						"type":        "string",
						"description": "Icon for the asset type. Must be one of: bezier, bone, book, boxes, bulb, camera-flash, camera, clapboard, compass, cube, drum, film-reel, film-strip, fire, flow-chart, four-squares, home, image, lamp, link, man-running, masks, music, mystery-ball, open-book, package, palette, scissors, shapes, stall, texture, tree, video-camera, website.",
						"enum":        []string{"bezier", "bone", "book", "boxes", "bulb", "camera-flash", "camera", "clapboard", "compass", "cube", "drum", "film-reel", "film-strip", "fire", "flow-chart", "four-squares", "home", "image", "lamp", "link", "man-running", "masks", "music", "mystery-ball", "open-book", "package", "palette", "scissors", "shapes", "stall", "texture", "tree", "video-camera", "website"},
					},
				},
				"required": []string{"name", "icon"},
			},
		},
		{
			Name:        "create_collection",
			Description: "Create a new collection (folder-like container for assets) in the project.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the collection to create.",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Optional description for the collection.",
					},
					"parent_id": map[string]interface{}{
						"type":        "string",
						"description": "Optional ID of the parent collection for nesting. Empty string for root level.",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "create_asset",
			Description: "Create a new asset (file) in a collection. Requires a template — use list_templates first to find available templates.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the asset to create.",
					},
					"collection_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the collection to create the asset in.",
					},
					"task_type_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the task type (e.g., Model, Rig, Animation). Use list_task_types to find available types.",
					},
					"template_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the file template to use. Use list_templates to find available templates.",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Optional description for the asset.",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Optional list of tag names to apply.",
					},
				},
				"required": []string{"name", "collection_id", "task_type_id", "template_id"},
			},
		},
		{
			Name:        "rename_asset",
			Description: "Rename an existing asset.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the asset to rename.",
					},
					"new_name": map[string]interface{}{
						"type":        "string",
						"description": "New name for the asset.",
					},
				},
				"required": []string{"asset_id", "new_name"},
			},
		},
		{
			Name:        "rename_collection",
			Description: "Rename an existing collection.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the collection to rename.",
					},
					"new_name": map[string]interface{}{
						"type":        "string",
						"description": "New name for the collection.",
					},
				},
				"required": []string{"collection_id", "new_name"},
			},
		},
		{
			Name:        "change_asset_status",
			Description: "Change the status of an asset (e.g., from Todo to In Progress). Use list_statuses to find available status IDs.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the asset to update.",
					},
					"status_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the new status.",
					},
				},
				"required": []string{"asset_id", "status_id"},
			},
		},
		{
			Name:        "assign_asset",
			Description: "Assign a user to an asset. Use list_users to find available user IDs.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the asset to assign.",
					},
					"user_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the user to assign.",
					},
				},
				"required": []string{"asset_id", "user_id"},
			},
		},
		{
			Name:        "unassign_asset",
			Description: "Remove the current assignee from an asset.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the asset to unassign.",
					},
				},
				"required": []string{"asset_id"},
			},
		},
		{
			Name:        "move_assets",
			Description: "Move one or more assets to a different collection.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "List of asset IDs to move.",
					},
					"target_collection_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the destination collection.",
					},
				},
				"required": []string{"asset_ids", "target_collection_id"},
			},
		},
		{
			Name:        "delete_asset",
			Description: "Delete an asset from the project. This requires user confirmation.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the asset to delete.",
					},
					"remove_files": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to also remove the physical files from disk. Default false.",
					},
				},
				"required": []string{"asset_id"},
			},
		},
		{
			Name:        "delete_collection",
			Description: "Delete a collection and optionally its files. This requires user confirmation.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the collection to delete.",
					},
					"remove_files": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to also remove the physical files from disk. Default false.",
					},
				},
				"required": []string{"collection_id"},
			},
		},

		// Script generation
		{
			Name:        "generate_script",
			Description: "Generate a shell or Python script for file system operations on project assets. The script is displayed to the user for review — it is never auto-executed. Use this for batch operations like rendering, file conversion, exports, etc.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"script_type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"powershell", "bash", "python"},
						"description": "Type of script to generate.",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Description of what the script does.",
					},
					"script_content": map[string]interface{}{
						"type":        "string",
						"description": "The script content to display to the user.",
					},
				},
				"required": []string{"script_type", "description", "script_content"},
			},
		},
	}
}

// ExecuteTool runs a tool by name with the given arguments against the project.
func ExecuteTool(projectPath string, toolName string, args map[string]interface{}) ToolResult {
	switch toolName {
	case "list_collections":
		return execListCollections(projectPath)
	case "list_assets_in_collection":
		return execListAssetsInCollection(projectPath, args)
	case "get_asset_details":
		return execGetAssetDetails(projectPath, args)
	case "list_users":
		return execListUsers(projectPath)
	case "list_statuses":
		return execListStatuses(projectPath)
	case "list_task_types":
		return execListTaskTypes(projectPath)
	case "list_tags":
		return execListTags(projectPath)
	case "list_templates":
		return execListTemplates(projectPath)
	case "search_knowledge":
		return execSearchKnowledge(args)
	case "create_asset_type":
		return execCreateAssetType(projectPath, args)
	case "create_collection":
		return execCreateCollection(projectPath, args)
	case "create_asset":
		return execCreateAsset(projectPath, args)
	case "rename_asset":
		return execRenameAsset(projectPath, args)
	case "rename_collection":
		return execRenameCollection(projectPath, args)
	case "change_asset_status":
		return execChangeAssetStatus(projectPath, args)
	case "assign_asset":
		return execAssignAsset(projectPath, args)
	case "unassign_asset":
		return execUnassignAsset(projectPath, args)
	case "move_assets":
		return execMoveAssets(projectPath, args)
	case "delete_asset":
		return execDeleteAsset(projectPath, args)
	case "delete_collection":
		return execDeleteCollection(projectPath, args)
	case "generate_script":
		return execGenerateScript(args)
	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("unknown tool: %s", toolName)}
	}
}

// getStringArg extracts a string argument with a default fallback.
func getStringArg(args map[string]interface{}, key, defaultVal string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

// getBoolArg extracts a boolean argument with a default fallback.
func getBoolArg(args map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultVal
}

// getStringSliceArg extracts a string slice argument.
func getStringSliceArg(args map[string]interface{}, key string) []string {
	if v, ok := args[key]; ok {
		if arr, ok := v.([]interface{}); ok {
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}

// --- Query tool implementations ---

func execListCollections(projectPath string) ToolResult {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	entities, err := repository.GetEntities(tx, false)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	type collectionSummary struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		TypeName string `json:"type_name"`
		ParentID string `json:"parent_id,omitempty"`
		Path     string `json:"path"`
	}
	summaries := make([]collectionSummary, 0, len(entities))
	for _, e := range entities {
		summaries = append(summaries, collectionSummary{
			ID:       e.Id,
			Name:     e.Name,
			TypeName: e.EntityTypeName,
			ParentID: e.ParentId,
			Path:     e.EntityPath,
		})
	}
	return ToolResult{Success: true, Data: summaries}
}

func execListAssetsInCollection(projectPath string, args map[string]interface{}) ToolResult {
	collectionID := getStringArg(args, "collection_id", "")
	if collectionID == "" {
		return ToolResult{Success: false, Error: "collection_id is required"}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	tasks, err := repository.GetEntityTasks(tx, collectionID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	type assetSummary struct {
		ID           string   `json:"id"`
		Name         string   `json:"name"`
		TypeName     string   `json:"type_name"`
		StatusName   string   `json:"status_name"`
		AssigneeName string   `json:"assignee_name,omitempty"`
		Tags         []string `json:"tags,omitempty"`
	}
	summaries := make([]assetSummary, 0, len(tasks))
	for _, t := range tasks {
		summaries = append(summaries, assetSummary{
			ID:           t.Id,
			Name:         t.Name,
			TypeName:     t.TaskTypeName,
			StatusName:   t.StatusShortName,
			AssigneeName: t.AssigneeName,
			Tags:         t.Tags,
		})
	}
	return ToolResult{Success: true, Data: summaries}
}

func execGetAssetDetails(projectPath string, args map[string]interface{}) ToolResult {
	assetID := getStringArg(args, "asset_id", "")
	if assetID == "" {
		return ToolResult{Success: false, Error: "asset_id is required"}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	task, err := repository.GetTask(tx, assetID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: task}
}

func execListUsers(projectPath string) ToolResult {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	users, err := repository.GetUsers(tx)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	type userSummary struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Name     string `json:"name"`
	}
	summaries := make([]userSummary, 0, len(users))
	for _, u := range users {
		summaries = append(summaries, userSummary{
			ID:       u.Id,
			Username: u.Username,
			Email:    u.Email,
			Name:     strings.TrimSpace(u.FirstName + " " + u.LastName),
		})
	}
	return ToolResult{Success: true, Data: summaries}
}

func execListStatuses(projectPath string) ToolResult {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	statuses, err := repository.GetStatuses(tx)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: statuses}
}

func execListTaskTypes(projectPath string) ToolResult {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	taskTypes, err := repository.GetTaskTypes(tx)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: taskTypes}
}

func execListTags(projectPath string) ToolResult {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	tags, err := repository.GetTags(tx)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: tags}
}

func execListTemplates(projectPath string) ToolResult {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	templates, err := repository.GetTemplates(tx, true)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	type templateSummary struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	summaries := make([]templateSummary, 0, len(templates))
	for _, t := range templates {
		summaries = append(summaries, templateSummary{
			ID:   t.Id,
			Name: t.Name,
		})
	}
	return ToolResult{Success: true, Data: summaries}
}

func execSearchKnowledge(args map[string]interface{}) ToolResult {
	query := strings.ToLower(getStringArg(args, "query", ""))
	if query == "" {
		return ToolResult{Success: false, Error: "query is required"}
	}

	var results []string
	for topic, content := range knowledgeBase {
		if strings.Contains(strings.ToLower(topic), query) || strings.Contains(strings.ToLower(content), query) {
			results = append(results, content)
		}
	}

	if len(results) == 0 {
		return ToolResult{Success: true, Data: "No specific documentation found for that topic. Try broader terms like: collections, assets, checkpoints, sync, dependencies, integrations, workflows, templates, tags, statuses, collaborators, studios, scripts."}
	}
	return ToolResult{Success: true, Data: strings.Join(results, "\n\n---\n\n")}
}

// --- CRUD tool implementations ---

// validAssetTypeIcons lists the allowed icon names for asset types.
var validAssetTypeIcons = map[string]bool{
	"bezier": true, "bone": true, "book": true, "boxes": true, "bulb": true,
	"camera-flash": true, "camera": true, "clapboard": true, "compass": true, "cube": true,
	"drum": true, "film-reel": true, "film-strip": true, "fire": true, "flow-chart": true,
	"four-squares": true, "home": true, "image": true, "lamp": true, "link": true,
	"man-running": true, "masks": true, "music": true, "mystery-ball": true, "open-book": true,
	"package": true, "palette": true, "scissors": true, "shapes": true, "stall": true,
	"texture": true, "tree": true, "video-camera": true, "website": true,
}

func execCreateAssetType(projectPath string, args map[string]interface{}) ToolResult {
	name := getStringArg(args, "name", "")
	icon := getStringArg(args, "icon", "")
	if name == "" {
		return ToolResult{Success: false, Error: "name is required"}
	}
	if icon == "" || !validAssetTypeIcons[icon] {
		return ToolResult{Success: false, Error: "icon is required and must be a valid icon name (e.g., cube, palette, camera)"}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	taskType, err := repository.CreateTaskType(tx, "", name, icon)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{
		"id":   taskType.Id,
		"name": taskType.Name,
	}}
}

func execCreateCollection(projectPath string, args map[string]interface{}) ToolResult {
	name := getStringArg(args, "name", "")
	if name == "" {
		return ToolResult{Success: false, Error: "name is required"}
	}
	description := getStringArg(args, "description", "")
	parentID := getStringArg(args, "parent_id", "")

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	// Get the first entity type as default
	entityTypes, err := repository.GetEntityTypes(tx)
	if err != nil || len(entityTypes) == 0 {
		return ToolResult{Success: false, Error: "no entity types available in project"}
	}
	entityTypeId := entityTypes[0].Id

	entity, err := repository.CreateEntity(tx, "", name, description, entityTypeId, parentID, "", false)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = tx.Commit()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{
		"id":   entity.Id,
		"name": entity.Name,
		"path": entity.EntityPath,
	}}
}

func execCreateAsset(projectPath string, args map[string]interface{}) ToolResult {
	name := getStringArg(args, "name", "")
	collectionID := getStringArg(args, "collection_id", "")
	taskTypeID := getStringArg(args, "task_type_id", "")
	templateID := getStringArg(args, "template_id", "")
	description := getStringArg(args, "description", "")
	tags := getStringSliceArg(args, "tags")

	if name == "" || collectionID == "" || taskTypeID == "" || templateID == "" {
		return ToolResult{Success: false, Error: "name, collection_id, task_type_id, and template_id are required"}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	callBack := func(current int, total int, message string, extraMessage string) {}

	task, err := repository.CreateTask(
		tx, "", name, taskTypeID, collectionID,
		false,      // isResource
		templateID, // templateId
		description,
		"", // templateFilePath
		tags,
		"",    // pointer
		false, // isLink
		"",    // previewId
		"",    // userId (agent-created)
		"Created by Clustta Agent",
		"", // groupId
		callBack,
	)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = tx.Commit()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{
		"id":   task.Id,
		"name": task.Name,
	}}
}

func execRenameAsset(projectPath string, args map[string]interface{}) ToolResult {
	assetID := getStringArg(args, "asset_id", "")
	newName := getStringArg(args, "new_name", "")
	if assetID == "" || newName == "" {
		return ToolResult{Success: false, Error: "asset_id and new_name are required"}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	_, err = repository.GetTask(tx, assetID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	_, err = repository.RenameTask(tx, assetID, newName)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = tx.Commit()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]string{"id": assetID, "new_name": newName}}
}

func execRenameCollection(projectPath string, args map[string]interface{}) ToolResult {
	collectionID := getStringArg(args, "collection_id", "")
	newName := getStringArg(args, "new_name", "")
	if collectionID == "" || newName == "" {
		return ToolResult{Success: false, Error: "collection_id and new_name are required"}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	_, err = repository.RenameEntity(tx, collectionID, newName)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = tx.Commit()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]string{"id": collectionID, "new_name": newName}}
}

func execChangeAssetStatus(projectPath string, args map[string]interface{}) ToolResult {
	assetID := getStringArg(args, "asset_id", "")
	statusID := getStringArg(args, "status_id", "")
	if assetID == "" || statusID == "" {
		return ToolResult{Success: false, Error: "asset_id and status_id are required"}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	err = repository.Updatestatus(tx, assetID, statusID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = tx.Commit()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]string{"asset_id": assetID, "status_id": statusID}}
}

func execAssignAsset(projectPath string, args map[string]interface{}) ToolResult {
	assetID := getStringArg(args, "asset_id", "")
	userID := getStringArg(args, "user_id", "")
	if assetID == "" || userID == "" {
		return ToolResult{Success: false, Error: "asset_id and user_id are required"}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	err = repository.AssignTask(tx, assetID, userID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = tx.Commit()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]string{"asset_id": assetID, "user_id": userID}}
}

func execUnassignAsset(projectPath string, args map[string]interface{}) ToolResult {
	assetID := getStringArg(args, "asset_id", "")
	if assetID == "" {
		return ToolResult{Success: false, Error: "asset_id is required"}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	err = repository.UnAssignTask(tx, assetID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = tx.Commit()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]string{"asset_id": assetID}}
}

func execMoveAssets(projectPath string, args map[string]interface{}) ToolResult {
	assetIDs := getStringSliceArg(args, "asset_ids")
	targetCollectionID := getStringArg(args, "target_collection_id", "")
	if len(assetIDs) == 0 || targetCollectionID == "" {
		return ToolResult{Success: false, Error: "asset_ids and target_collection_id are required"}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	for _, assetID := range assetIDs {
		err = repository.ChangeEntity(tx, assetID, targetCollectionID)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to move asset %s: %s", assetID, err.Error())}
		}
	}

	err = tx.Commit()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{"moved": len(assetIDs), "target": targetCollectionID}}
}

func execDeleteAsset(projectPath string, args map[string]interface{}) ToolResult {
	assetID := getStringArg(args, "asset_id", "")
	removeFiles := getBoolArg(args, "remove_files", false)
	if assetID == "" {
		return ToolResult{Success: false, Error: "asset_id is required"}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	task, err := repository.GetTask(tx, assetID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = repository.DeleteTask(tx, assetID, removeFiles, true)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = tx.Commit()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]string{"deleted": assetID, "name": task.Name}}
}

func execDeleteCollection(projectPath string, args map[string]interface{}) ToolResult {
	collectionID := getStringArg(args, "collection_id", "")
	removeFiles := getBoolArg(args, "remove_files", false)
	if collectionID == "" {
		return ToolResult{Success: false, Error: "collection_id is required"}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	entity, err := repository.GetEntity(tx, collectionID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = repository.DeleteEntity(tx, collectionID, removeFiles, true)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = tx.Commit()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]string{"deleted": collectionID, "name": entity.Name}}
}

func execGenerateScript(args map[string]interface{}) ToolResult {
	scriptType := getStringArg(args, "script_type", "bash")
	description := getStringArg(args, "description", "")
	content := getStringArg(args, "script_content", "")

	if content == "" {
		return ToolResult{Success: false, Error: "script_content is required"}
	}

	return ToolResult{Success: true, Data: map[string]string{
		"type":        scriptType,
		"description": description,
		"content":     content,
	}}
}

// BuildProjectContext gathers project metadata for the LLM system prompt.
func BuildProjectContext(projectPath string) (string, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return "", err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var context strings.Builder

	// Collections
	entities, err := repository.GetEntities(tx, false)
	if err == nil && len(entities) > 0 {
		context.WriteString(fmt.Sprintf("Collections (%d):\n", len(entities)))
		for _, e := range entities {
			context.WriteString(fmt.Sprintf("- %s (ID: %s, type: %s", e.Name, e.Id, e.EntityTypeName))
			if e.ParentId != "" {
				context.WriteString(fmt.Sprintf(", parent: %s", e.ParentId))
			}
			context.WriteString(")\n")
		}
		context.WriteString("\n")
	} else {
		context.WriteString("Collections: None yet\n\n")
	}

	// Asset count per collection
	for _, e := range entities {
		tasks, err := repository.GetEntityTasks(tx, e.Id)
		if err == nil && len(tasks) > 0 {
			context.WriteString(fmt.Sprintf("Assets in '%s' (%d):\n", e.Name, len(tasks)))
			for _, t := range tasks {
				line := fmt.Sprintf("- %s (ID: %s, type: %s, status: %s", t.Name, t.Id, t.TaskTypeName, t.StatusShortName)
				if t.AssigneeName != "" {
					line += fmt.Sprintf(", assigned to: %s", t.AssigneeName)
				}
				line += ")"
				context.WriteString(line + "\n")
			}
			context.WriteString("\n")
		}
	}

	// Statuses
	statuses, err := repository.GetStatuses(tx)
	if err == nil && len(statuses) > 0 {
		context.WriteString("Available statuses:\n")
		for _, s := range statuses {
			context.WriteString(fmt.Sprintf("- %s (ID: %s, short: %s)\n", s.Name, s.Id, s.ShortName))
		}
		context.WriteString("\n")
	}

	// Task types
	taskTypes, err := repository.GetTaskTypes(tx)
	if err == nil && len(taskTypes) > 0 {
		context.WriteString("Available task types:\n")
		for _, tt := range taskTypes {
			context.WriteString(fmt.Sprintf("- %s (ID: %s)\n", tt.Name, tt.Id))
		}
		context.WriteString("\n")
	}

	// Users
	users, err := repository.GetUsers(tx)
	if err == nil && len(users) > 0 {
		context.WriteString("Users:\n")
		for _, u := range users {
			name := strings.TrimSpace(u.FirstName + " " + u.LastName)
			context.WriteString(fmt.Sprintf("- %s / %s (ID: %s)\n", u.Username, name, u.Id))
		}
		context.WriteString("\n")
	}

	// Tags
	tags, err := repository.GetTags(tx)
	if err == nil && len(tags) > 0 {
		context.WriteString("Tags:\n")
		for _, t := range tags {
			context.WriteString(fmt.Sprintf("- %s (ID: %s)\n", t.Name, t.Id))
		}
		context.WriteString("\n")
	}

	// Templates
	templates, err := repository.GetTemplates(tx, true)
	if err == nil && len(templates) > 0 {
		context.WriteString("Templates:\n")
		for _, t := range templates {
			context.WriteString(fmt.Sprintf("- %s (ID: %s)\n", t.Name, t.Id))
		}
	}

	return context.String(), nil
}

// serializeToolResult converts a ToolResult to a JSON string for the LLM.
func SerializeToolResult(result ToolResult) string {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Sprintf(`{"success": false, "error": "failed to serialize result: %s"}`, err.Error())
	}
	return string(data)
}
