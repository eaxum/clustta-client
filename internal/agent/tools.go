package agent

import (
	agentcommands "clustta/internal/agent/commands"
	"clustta/internal/auth_service"
	"clustta/internal/constants"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"sort"
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
	tools := []ToolDefinition{
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
		{
			Name:        "get_my_permissions",
			Description: "Get the current user's role name and full list of permissions. Use this when the user asks about their role, what they can do, or before attempting an action to verify permission.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "get_user_activity",
			Description: "Get activity summary for all users: total checkpoints made and when their last checkpoint was. Use this when the user asks about activity, contributions, or who last worked on the project.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},

		{
			Name:        "list_checkpoints",
			Description: "List checkpoint (version) history for an asset. Returns checkpoint IDs, comments, timestamps, and file sizes.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the asset to list checkpoints for.",
					},
				},
				"required": []string{"asset_id"},
			},
		},

		{
			Name:        "create_tag",
			Description: "Create a new tag in the project.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the tag to create.",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "get_asset_tags",
			Description: "Get all tags currently applied to a specific asset.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the asset.",
					},
				},
				"required": []string{"asset_id"},
			},
		},

		{
			Name:        "list_dependencies",
			Description: "List all dependencies for a specific asset.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the asset to list dependencies for.",
					},
				},
				"required": []string{"asset_id"},
			},
		},
		{
			Name:        "list_dependency_types",
			Description: "List all available dependency types in the project.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},

		{
			Name:        "list_collection_types",
			Description: "List all collection types (entity types) in the project.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "delete_asset_type",
			Description: "Delete an asset type from the project. This requires user confirmation.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_type_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the asset type to delete.",
					},
				},
				"required": []string{"asset_type_id"},
			},
		},
		{
			Name:        "delete_collection_type",
			Description: "Delete a collection type from the project. This requires user confirmation.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection_type_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the collection type to delete.",
					},
				},
				"required": []string{"collection_type_id"},
			},
		},

		{
			Name:        "search_assets",
			Description: "Search for assets across all collections. Returns paginated results (default 50 per page). Use offset to page through large result sets. The response includes total_count so you know how many matched overall.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Substring to match in asset names (case-insensitive).",
					},
					"status_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by status ID.",
					},
					"task_type_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by asset type ID.",
					},
					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by assignee user ID.",
					},
					"tag_name": map[string]interface{}{
						"type":        "string",
						"description": "Filter by tag name.",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results to return (default 50, max 200).",
					},
					"offset": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results to skip for pagination (default 0).",
					},
				},
				"required": []string{},
			},
		},

		{
			Name:        "get_project_summary",
			Description: "Get a high-level summary of the project: total collections, total assets, breakdown by status, breakdown by assignee.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},

		{
			Name:        "remove_user",
			Description: "Remove a user/collaborator from the project. This requires user confirmation.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the user to remove.",
					},
				},
				"required": []string{"user_id"},
			},
		},

		{
			Name:        "list_ignore_patterns",
			Description: "List all current ignore patterns in the project.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "add_ignore_pattern",
			Description: "Add a pattern to the project's ignore list. Always use glob format: '*.ext' for extensions (e.g. '*.tmp', '*.log', '*.blend1'), folder names for directories (e.g. 'node_modules'). Bare names like 'json' or '.json' are auto-normalized to '*.json'.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "The ignore pattern in glob format, e.g. '*.json', '*.tmp', 'node_modules'.",
					},
				},
				"required": []string{"pattern"},
			},
		},
		{
			Name:        "remove_ignore_pattern",
			Description: "Remove a pattern from the project's ignore list.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "The exact ignore pattern to remove.",
					},
				},
				"required": []string{"pattern"},
			},
		},

		{
			Name:        "generate_script",
			Description: "Generate a shell or Python script for file system operations on project assets. The script is displayed to the user for review - it is never auto-executed. Use this for batch operations like rendering, file conversion, exports, etc.",
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

	tools = append(tools, []ToolDefinition{
		// --- Type maintenance ---

		// --- Reveal ---
		{
			Name:        "reveal_asset_on_disk",
			Description: "Open the OS file explorer at the asset's location on disk. Use when the user wants to see the file in their system file manager.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_id": map[string]interface{}{"type": "string", "description": "ID of the asset to reveal."},
				},
				"required": []string{"asset_id"},
			},
		},
		{
			Name:        "reveal_in_browser",
			Description: "Navigate the in-app browser to focus the given asset or collection. Use when the user wants to be taken to the item inside Clustta. Provide either asset_id or collection_id (not both).",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_id":      map[string]interface{}{"type": "string", "description": "ID of the asset to navigate to."},
					"collection_id": map[string]interface{}{"type": "string", "description": "ID of the collection to navigate to."},
				},
				"required": []string{},
			},
		},

		// --- Workflows ---
		{
			Name:        "list_workflows",
			Description: "List all workflows defined in the project. A workflow is a reusable template that creates a tree of collections and assets when applied.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},

		// --- Roles ---
		{
			Name:        "list_roles",
			Description: "List all roles defined in the project, with their permissions. Use to find role names and IDs before changing a user's role.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "change_collaborator_role",
			Description: "Change a project user's role by role name. Use list_users to find user_id and list_roles to find role names.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_id":   map[string]interface{}{"type": "string", "description": "ID of the user whose role to change."},
					"role_name": map[string]interface{}{"type": "string", "description": "Name of the new role."},
				},
				"required": []string{"user_id", "role_name"},
			},
		},
		{
			Name:        "update_role",
			Description: "Update a role's name and permission attributes. Provide a full set of permission booleans - fields not provided default to false. Use list_roles first to read the current values, then resubmit with the desired changes.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":         map[string]interface{}{"type": "string", "description": "ID of the role to update."},
					"name":       map[string]interface{}{"type": "string", "description": "New role name."},
					"attributes": map[string]interface{}{"type": "object", "description": "Permission attributes object. Keys: view_collection, create_collection, update_collection, delete_collection, view_asset, create_asset, update_asset, delete_asset, view_template, create_template, update_template, delete_template, view_checkpoint, create_checkpoint, delete_checkpoint, pull_chunk, assign_asset, unassign_asset, add_user, remove_user, change_role, change_status, set_done_asset, set_retake_asset, view_done_asset, manage_dependencies, manage_share_links. Each is a boolean."},
				},
				"required": []string{"id", "name", "attributes"},
			},
		},

		// --- Knowledge ---
		{
			Name:        "search_project_text",
			Description: "Search project content (asset names/descriptions, collection names/descriptions, checkpoint comments, tag names, role names) for the given query. Use this when the user asks about specific notes, comments, descriptions, or named items in their actual project (as opposed to general Clustta concepts - for those use search_knowledge).",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{"type": "string", "description": "Substring to look for (case-insensitive)."},
					"limit": map[string]interface{}{"type": "integer", "description": "Max matches per category. Defaults to 25."},
				},
				"required": []string{"query"},
			},
		},

		// --- Project collaborators (server-side membership) ---
		{
			Name:        "list_project_collaborators",
			Description: "List all users who have been invited to the active project's remote (server-side membership). Different from list_users, which lists users known to the local project DB. Requires the project to be synced to a remote.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "add_project_collaborator",
			Description: "Invite a user to the active project's remote by email or user_id. If email is given, the user is looked up via the global Clustta server. Optionally specify a role (server-side default applies if omitted).",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"email":   map[string]interface{}{"type": "string", "description": "Email address of an existing Clustta user. Required if user_id is not provided."},
					"user_id": map[string]interface{}{"type": "string", "description": "Clustta user ID. Required if email is not provided."},
					"role":    map[string]interface{}{"type": "string", "description": "Optional role name to assign (e.g. 'collaborator')."},
				},
				"required": []string{},
			},
		},
		{
			Name:        "remove_project_collaborator",
			Description: "Remove a user from the active project's remote (revokes server-side access).",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_id": map[string]interface{}{"type": "string", "description": "Clustta user ID of the collaborator to remove."},
				},
				"required": []string{"user_id"},
			},
		},

		// --- Studio collaborators (separate studio server) ---
		{
			Name:        "list_studios",
			Description: "List the studios configured locally with their IDs, names, URLs and hosting modes. Use this to discover the studio_id required by the studio_* tools.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "list_studio_users",
			Description: "List members of the given studio (studio-level membership, not project membership). Use list_studios to find studio_id values.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"studio_id": map[string]interface{}{"type": "string", "description": "ID of the studio."},
				},
				"required": []string{"studio_id"},
			},
		},
		{
			Name:        "add_studio_collaborator",
			Description: "Invite a user (by email) to a studio with the given role. The studio server applies authorization - the caller must have permission to add studio members.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"email":     map[string]interface{}{"type": "string", "description": "Email of an existing Clustta user."},
					"studio_id": map[string]interface{}{"type": "string", "description": "ID of the studio."},
					"role_name": map[string]interface{}{"type": "string", "description": "Role to assign (e.g. 'admin', 'manager', 'artist')."},
				},
				"required": []string{"email", "studio_id", "role_name"},
			},
		},
		{
			Name:        "change_studio_collaborator_role",
			Description: "Change an existing studio member's role.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_id":   map[string]interface{}{"type": "string", "description": "ID of the user whose studio role to change."},
					"studio_id": map[string]interface{}{"type": "string", "description": "ID of the studio."},
					"role_name": map[string]interface{}{"type": "string", "description": "New role name."},
				},
				"required": []string{"user_id", "studio_id", "role_name"},
			},
		},
		{
			Name:        "remove_studio_collaborator",
			Description: "Remove a user from the studio (revokes studio-level access). This does not affect project-level membership directly.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_id":   map[string]interface{}{"type": "string", "description": "ID of the user to remove."},
					"studio_id": map[string]interface{}{"type": "string", "description": "ID of the studio."},
				},
				"required": []string{"user_id", "studio_id"},
			},
		},

		// --- Browser filter tools ---
		{
			Name:        "list_filter_dimensions",
			Description: "Return the project's filter vocabulary (statuses, asset types, collection types, tags, users, file extensions, file states, and the current user id). Call this BEFORE apply_browser_filter when you don't already know the available values, so the filter terms you supply match real entries.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "apply_browser_filter",
			Description: "Apply filters to the asset/collection browser view. Each list contains names or ids; the backend resolves them against the project vocabulary and tells the UI to display matching items only. Use the special value '@me' inside `assignees` to refer to the current user. Use `no_assignees` for 'unassigned'. Set `deep` to true when the user wants the search to span the whole project. This does NOT modify any project data - it only changes what is visible in the browser.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"statuses":         map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Status names, short_names, or ids to include."},
					"asset_types":      map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Asset type names or ids."},
					"collection_types": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Collection type names or ids."},
					"tags":             map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Tag names or ids."},
					"assignees":        map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "User ids, usernames, emails, names, or '@me'."},
					"states":           map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "File states: normal, modified, outdated, fetchable, missing."},
					"extensions":       map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "File extensions, with or without leading dot (e.g. 'blend' or '.blend')."},
					"has_assignees":    map[string]interface{}{"type": "boolean", "description": "Show only items that have any assignee."},
					"no_assignees":     map[string]interface{}{"type": "boolean", "description": "Show only items with no assignee (unassigned)."},
					"deep":             map[string]interface{}{"type": "boolean", "description": "Search across the entire project rather than the current collection."},
					"search":           map[string]interface{}{"type": "string", "description": "Text query applied to item names."},
					"show_collections": map[string]interface{}{"type": "boolean", "description": "Toggle visibility of collections in the view (optional)."},
					"show_assets":      map[string]interface{}{"type": "boolean", "description": "Toggle visibility of assets in the view (optional)."},
					"show_resources":   map[string]interface{}{"type": "boolean", "description": "Toggle visibility of resources in the view (optional)."},
					"only_assets":      map[string]interface{}{"type": "boolean", "description": "When true, hide collections and resources entirely (optional)."},
				},
			},
		},
		{
			Name:        "clear_browser_filter",
			Description: "Reset all browser filters back to the default empty state (equivalent to clicking the Clear button in the filter bar).",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}...)

	for _, def := range agentcommands.Definitions() {
		tools = append(tools, ToolDefinition{
			Name: def.Name, Description: def.Description, Parameters: def.Parameters,
		})
	}
	sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })

	return tools
}

// toolPermission maps a tool to the Role field getter that must return true.
type toolPermission struct {
	Check func(role models.Role) bool
	Label string
}

// toolPermissions maps mutating tools to the permission(s) they require.
var toolPermissions = map[string]toolPermission{
	"create_tag":             {func(r models.Role) bool { return r.CreateAsset }, "Create Asset"},
	"delete_asset_type":      {func(r models.Role) bool { return r.DeleteAsset }, "Delete Asset"},
	"delete_collection_type": {func(r models.Role) bool { return r.DeleteAsset }, "Delete Asset"},

	"remove_user":              {func(r models.Role) bool { return r.RemoveUser }, "Remove User"},
	"change_collaborator_role": {func(r models.Role) bool { return r.ChangeRole }, "Change Role"},
	"update_role":              {func(r models.Role) bool { return r.ChangeRole }, "Change Role"},

	"add_project_collaborator":    {func(r models.Role) bool { return r.AddUser }, "Add User"},
	"remove_project_collaborator": {func(r models.Role) bool { return r.RemoveUser }, "Remove User"},

	"add_ignore_pattern":    {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"remove_ignore_pattern": {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
}

// checkPermission verifies the current user has the required role permission for a tool.
func checkPermission(projectPath string, toolName string) error {
	perm, ok := toolPermissions[toolName]
	if !ok {
		return nil
	}

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return fmt.Errorf("could not determine current user: %w", err)
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return fmt.Errorf("could not open project: %w", err)
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	userData, err := repository.GetUser(tx, user.Id)
	if err != nil {
		return fmt.Errorf("user not found in project: %w", err)
	}

	role, err := repository.GetRole(tx, userData.RoleId)
	if err != nil {
		return fmt.Errorf("could not load user role: %w", err)
	}

	if !perm.Check(role) {
		return fmt.Errorf("permission denied: your role '%s' does not have '%s' permission", role.Name, perm.Label)
	}

	return nil
}

// ExecuteTool runs a tool by name with the given arguments against the project.
func ExecuteTool(projectPath string, toolName string, args map[string]interface{}) ToolResult {
	return ExecuteToolContext(context.Background(), projectPath, toolName, args)
}

// ExecuteToolContext runs a tool with cancellation propagated into registry commands.
func ExecuteToolContext(ctx context.Context, projectPath string, toolName string, args map[string]interface{}) ToolResult {
	if _, ok := agentcommands.DefinitionFor(toolName); ok {
		if def, _ := agentcommands.DefinitionFor(toolName); def.Direct != nil {
			result, err := agentcommands.ExecuteDirect(projectPath, toolName, args)
			if err != nil {
				return ToolResult{Success: false, Error: err.Error()}
			}
			return ToolResult{Success: true, Data: result}
		}
		result, err := agentcommands.ExecutePrepared(ctx, projectPath, toolName, args)
		if err != nil {
			return ToolResult{Success: false, Error: err.Error()}
		}
		return ToolResult{Success: true, Data: result}
	}
	if err := checkPermission(projectPath, toolName); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	switch toolName {
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
	case "get_my_permissions":
		return execGetMyPermissions(projectPath)
	case "get_user_activity":
		return execGetUserActivity(projectPath)
	case "list_checkpoints":
		return execListCheckpoints(projectPath, args)
	case "create_tag":
		return execCreateTag(projectPath, args)
	case "get_asset_tags":
		return execGetAssetTags(projectPath, args)
	case "list_dependencies":
		return execListDependencies(projectPath, args)
	case "list_dependency_types":
		return execListDependencyTypes(projectPath)
	case "list_collection_types":
		return execListCollectionTypes(projectPath)
	case "delete_asset_type":
		return execDeleteAssetType(projectPath, args)
	case "delete_collection_type":
		return execDeleteCollectionType(projectPath, args)
	case "search_assets":
		return execSearchAssets(projectPath, args)
	case "get_project_summary":
		return execGetProjectSummary(projectPath)
	case "remove_user":
		return execRemoveUser(projectPath, args)
	case "list_ignore_patterns":
		return execListIgnorePatterns(projectPath)
	case "add_ignore_pattern":
		return execAddIgnorePattern(projectPath, args)
	case "remove_ignore_pattern":
		return execRemoveIgnorePattern(projectPath, args)
	case "generate_script":
		return execGenerateScript(args)

	case "reveal_asset_on_disk":
		return execRevealAssetOnDisk(projectPath, args)
	case "reveal_in_browser":
		return execRevealInBrowser(projectPath, args)
	case "list_workflows":
		return execListWorkflows(projectPath)
	case "list_roles":
		return execListRoles(projectPath)
	case "change_collaborator_role":
		return execChangeCollaboratorRole(projectPath, args)
	case "update_role":
		return execUpdateRole(projectPath, args)
	case "search_project_text":
		return execSearchProjectText(projectPath, args)

	case "list_project_collaborators":
		return execListProjectCollaborators(projectPath)
	case "add_project_collaborator":
		return execAddProjectCollaborator(projectPath, args)
	case "remove_project_collaborator":
		return execRemoveProjectCollaborator(projectPath, args)
	case "list_studios":
		return execListStudios()
	case "list_studio_users":
		return execListStudioUsers(args)
	case "add_studio_collaborator":
		return execAddStudioCollaborator(args)
	case "change_studio_collaborator_role":
		return execChangeStudioCollaboratorRole(args)
	case "remove_studio_collaborator":
		return execRemoveStudioCollaborator(args)

	case "list_filter_dimensions":
		return execListFilterDimensions(projectPath)
	case "apply_browser_filter":
		return execApplyBrowserFilter(projectPath, args)
	case "clear_browser_filter":
		return execClearBrowserFilter(projectPath, args)

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

// activeUserID returns the signed-in user's ID, or empty string if unavailable.
func activeUserID(projectPath string) string {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return ""
	}
	return user.Id
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

// getObjSliceArg extracts a slice of map[string]interface{} from arguments.
func getObjSliceArg(args map[string]interface{}, key string) []map[string]interface{} {
	if v, ok := args[key]; ok {
		if arr, ok := v.([]interface{}); ok {
			result := make([]map[string]interface{}, 0, len(arr))
			for _, item := range arr {
				if obj, ok := item.(map[string]interface{}); ok {
					result = append(result, obj)
				}
			}
			return result
		}
	}
	return nil
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

	asset, err := repository.GetAsset(tx, assetID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: asset}
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

	assetTypes, err := repository.GetAssetTypes(tx)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: assetTypes}
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

func execGetMyPermissions(projectPath string) ToolResult {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return ToolResult{Success: false, Error: "could not determine current user: " + err.Error()}
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

	userData, err := repository.GetUser(tx, user.Id)
	if err != nil {
		return ToolResult{Success: false, Error: "user not found in project: " + err.Error()}
	}

	role, err := repository.GetRole(tx, userData.RoleId)
	if err != nil {
		return ToolResult{Success: false, Error: "could not load role: " + err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{
		"user": user.Username,
		"role": role.Name,
		"permissions": map[string]bool{
			"view_collection":     role.ViewCollection,
			"create_collection":   role.CreateCollection,
			"update_collection":   role.UpdateCollection,
			"delete_collection":   role.DeleteCollection,
			"view_asset":          role.ViewAsset,
			"create_asset":        role.CreateAsset,
			"update_asset":        role.UpdateAsset,
			"delete_asset":        role.DeleteAsset,
			"view_template":       role.ViewTemplate,
			"create_template":     role.CreateTemplate,
			"update_template":     role.UpdateTemplate,
			"delete_template":     role.DeleteTemplate,
			"view_checkpoint":     role.ViewCheckpoint,
			"create_checkpoint":   role.CreateCheckpoint,
			"delete_checkpoint":   role.DeleteCheckpoint,
			"assign_asset":        role.AssignAsset,
			"unassign_asset":      role.UnassignAsset,
			"add_user":            role.AddUser,
			"remove_user":         role.RemoveUser,
			"change_role":         role.ChangeRole,
			"change_status":       role.ChangeStatus,
			"manage_dependencies": role.ManageDependencies,
		},
	}}
}

func execGetUserActivity(projectPath string) ToolResult {
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

	checkpoints, err := repository.GetSimpleCheckpoints(tx)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	userNames := map[string]string{}
	for _, u := range users {
		userNames[u.Id] = strings.TrimSpace(u.FirstName + " " + u.LastName)
	}

	type userActivity struct {
		UserID           string `json:"user_id"`
		Name             string `json:"name"`
		TotalCheckpoints int    `json:"total_checkpoints"`
		LastCheckpoint   string `json:"last_checkpoint,omitempty"`
		LastComment      string `json:"last_comment,omitempty"`
	}

	activityMap := map[string]*userActivity{}
	for _, u := range users {
		activityMap[u.Id] = &userActivity{
			UserID: u.Id,
			Name:   userNames[u.Id],
		}
	}

	for _, cp := range checkpoints {
		ua, ok := activityMap[cp.AuthorUID]
		if !ok {
			continue
		}
		ua.TotalCheckpoints++
		if cp.CreatedAt > ua.LastCheckpoint {
			ua.LastCheckpoint = cp.CreatedAt
			ua.LastComment = cp.Comment
		}
	}

	results := make([]userActivity, 0, len(activityMap))
	for _, ua := range activityMap {
		results = append(results, *ua)
	}

	return ToolResult{Success: true, Data: results}
}

// --- CRUD tool implementations ---

// validAssetTypeIcons lists the allowed icon names for asset types.
var validAssetTypeIcons = constants.ValidTypeIcons

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

// --- Checkpoint tool ---

func execListCheckpoints(projectPath string, args map[string]interface{}) ToolResult {
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

	checkpoints, err := repository.GetCheckpoints(tx, assetID, false)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	type checkpointSummary struct {
		ID        string `json:"id"`
		Comment   string `json:"comment"`
		CreatedAt string `json:"created_at"`
		FileSize  int    `json:"file_size"`
		AuthorID  string `json:"author_id"`
		GroupID   string `json:"group_id"`
	}
	summaries := make([]checkpointSummary, 0, len(checkpoints))
	for _, cp := range checkpoints {
		summaries = append(summaries, checkpointSummary{
			ID:        cp.Id,
			Comment:   cp.Comment,
			CreatedAt: cp.CreatedAt,
			FileSize:  cp.FileSize,
			AuthorID:  cp.AuthorUID,
			GroupID:   cp.GroupId,
		})
	}
	return ToolResult{Success: true, Data: summaries}
}

// --- Tag management tools ---

func execCreateTag(projectPath string, args map[string]interface{}) ToolResult {
	name := getStringArg(args, "name", "")
	if name == "" {
		return ToolResult{Success: false, Error: "name is required"}
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

	tag, err := repository.CreateTag(tx, "", name)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]string{"id": tag.Id, "name": tag.Name}}
}

func execGetAssetTags(projectPath string, args map[string]interface{}) ToolResult {
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

	tags, err := repository.GetAssetTags(tx, assetID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: tags}
}

// --- Dependency tools ---

func execListDependencies(projectPath string, args map[string]interface{}) ToolResult {
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

	deps, err := repository.GetAssetDependencies(tx, assetID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: deps}
}

func execListDependencyTypes(projectPath string) ToolResult {
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

	depTypes, err := repository.GetDependencyTypes(tx)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: depTypes}
}

// --- Collection type tools ---

func execListCollectionTypes(projectPath string) ToolResult {
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

	collectionTypes, err := repository.GetCollectionTypes(tx)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: collectionTypes}
}

func execDeleteAssetType(projectPath string, args map[string]interface{}) ToolResult {
	id := getStringArg(args, "asset_type_id", "")
	if id == "" {
		return ToolResult{Success: false, Error: "asset_type_id is required"}
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

	err = repository.DeleteAssetType(tx, id)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]string{"deleted": id}}
}

func execDeleteCollectionType(projectPath string, args map[string]interface{}) ToolResult {
	id := getStringArg(args, "collection_type_id", "")
	if id == "" {
		return ToolResult{Success: false, Error: "collection_type_id is required"}
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

	err = repository.DeleteCollectionType(tx, id)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]string{"deleted": id}}
}

// --- Search tool ---

func execSearchAssets(projectPath string, args map[string]interface{}) ToolResult {
	nameFilter := strings.ToLower(getStringArg(args, "name", ""))
	statusFilter := getStringArg(args, "status_id", "")
	typeFilter := getStringArg(args, "task_type_id", "")
	assigneeFilter := getStringArg(args, "assignee_id", "")
	tagFilter := strings.ToLower(getStringArg(args, "tag_name", ""))

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

	assets, err := repository.GetAssets(tx, false)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	type assetResult struct {
		ID             string   `json:"id"`
		Name           string   `json:"name"`
		TypeName       string   `json:"type_name"`
		StatusName     string   `json:"status_name"`
		AssigneeName   string   `json:"assignee_name,omitempty"`
		CollectionName string   `json:"collection_name"`
		Tags           []string `json:"tags,omitempty"`
	}

	var results []assetResult
	for _, a := range assets {
		if nameFilter != "" && !strings.Contains(strings.ToLower(a.Name), nameFilter) {
			continue
		}
		if statusFilter != "" && a.StatusId != statusFilter {
			continue
		}
		if typeFilter != "" && a.AssetTypeId != typeFilter {
			continue
		}
		if assigneeFilter != "" && a.AssigneeId != assigneeFilter {
			continue
		}
		if tagFilter != "" {
			hasTag := false
			for _, t := range a.Tags {
				if strings.ToLower(t) == tagFilter {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}
		results = append(results, assetResult{
			ID:             a.Id,
			Name:           a.Name,
			TypeName:       a.AssetTypeName,
			StatusName:     a.StatusShortName,
			AssigneeName:   a.AssigneeName,
			CollectionName: a.CollectionName,
			Tags:           a.Tags,
		})
	}

	totalCount := len(results)
	limit := getIntArg(args, "limit", 50)
	offset := getIntArg(args, "offset", 0)
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	if offset > totalCount {
		offset = totalCount
	}
	end := offset + limit
	if end > totalCount {
		end = totalCount
	}
	paged := results[offset:end]

	return ToolResult{Success: true, Data: map[string]interface{}{
		"total_count": totalCount,
		"offset":      offset,
		"limit":       limit,
		"returned":    len(paged),
		"assets":      paged,
	}}
}

// --- Project summary tool ---

func execGetProjectSummary(projectPath string) ToolResult {
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

	collections, err := repository.GetCollections(tx, false)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	assets, err := repository.GetAssets(tx, false)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	statusCounts := map[string]int{}
	assigneeCounts := map[string]int{}
	typeCounts := map[string]int{}
	unassigned := 0

	for _, a := range assets {
		statusCounts[a.StatusShortName]++
		typeCounts[a.AssetTypeName]++
		if a.AssigneeName != "" {
			assigneeCounts[a.AssigneeName]++
		} else {
			unassigned++
		}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{
		"total_collections": len(collections),
		"total_assets":      len(assets),
		"by_status":         statusCounts,
		"by_assignee":       assigneeCounts,
		"by_type":           typeCounts,
		"unassigned":        unassigned,
	}}
}

// --- User management ---

func execRemoveUser(projectPath string, args map[string]interface{}) ToolResult {
	userID := getStringArg(args, "user_id", "")
	if userID == "" {
		return ToolResult{Success: false, Error: "user_id is required"}
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

	err = repository.RemoveUser(tx, userID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]string{"removed": userID}}
}

// --- Ignore list tools ---

func execListIgnorePatterns(projectPath string) ToolResult {
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

	patterns, err := repository.GetIgnoreList(tx)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: patterns}
}

func execAddIgnorePattern(projectPath string, args map[string]interface{}) ToolResult {
	pattern := getStringArg(args, "pattern", "")
	if pattern == "" {
		return ToolResult{Success: false, Error: "pattern is required"}
	}

	if !strings.ContainsAny(pattern, "*?/\\ ") {
		if strings.HasPrefix(pattern, ".") {
			pattern = "*" + pattern
		} else if !strings.Contains(pattern, ".") && len(pattern) <= 10 {
			pattern = "*." + pattern
		}
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

	existing, err := repository.GetIgnoreList(tx)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	for _, p := range existing {
		if p == pattern {
			return ToolResult{Success: true, Data: "Pattern already exists in ignore list."}
		}
	}

	existing = append(existing, pattern)
	err = utils.SetProjectIgnoreList(tx, existing)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{"added": pattern, "total_patterns": len(existing), "ignore_list": existing}}
}

func execRemoveIgnorePattern(projectPath string, args map[string]interface{}) ToolResult {
	pattern := getStringArg(args, "pattern", "")
	if pattern == "" {
		return ToolResult{Success: false, Error: "pattern is required"}
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

	existing, err := repository.GetIgnoreList(tx)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	updated := make([]string, 0, len(existing))
	found := false
	for _, p := range existing {
		if p == pattern {
			found = true
			continue
		}
		updated = append(updated, p)
	}

	if !found {
		return ToolResult{Success: false, Error: "pattern not found in ignore list"}
	}

	err = utils.SetProjectIgnoreList(tx, updated)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{"removed": pattern, "total_patterns": len(updated), "ignore_list": updated}}
}

// --- Project type setup ---

// projectTypePresets defines standard asset types and collection types for common pipelines.
var projectTypePresets = map[string]struct {
	AssetTypes      []struct{ Name, Icon string }
	CollectionTypes []struct{ Name, Icon string }
}{
	"animation": {
		AssetTypes: []struct{ Name, Icon string }{
			{"Storyboard", "image"}, {"Animatic", "film-strip"}, {"Layout", "flow-chart"},
			{"Animation", "man-running"}, {"Cleanup", "bezier"}, {"Color", "palette"},
			{"Compositing", "film-reel"}, {"Background", "image"}, {"Character Design", "masks"},
			{"Audio", "music"}, {"Sound Effect", "drum"},
		},
		CollectionTypes: []struct{ Name, Icon string }{
			{"Sequence", "film-strip"}, {"Shot", "camera"}, {"Character", "masks"},
			{"Prop", "package"}, {"Environment", "tree"}, {"Library", "book"},
		},
	},
	"game": {
		AssetTypes: []struct{ Name, Icon string }{
			{"Model", "cube"}, {"Texture", "texture"}, {"Rig", "bone"},
			{"Animation", "man-running"}, {"VFX", "fire"}, {"UI", "website"},
			{"Audio", "music"}, {"Level Design", "compass"}, {"Concept", "palette"},
			{"Script", "flow-chart"},
		},
		CollectionTypes: []struct{ Name, Icon string }{
			{"Character", "masks"}, {"Environment", "tree"}, {"Prop", "package"},
			{"Weapon", "scissors"}, {"Vehicle", "compass"}, {"UI", "website"},
			{"Library", "book"},
		},
	},
	"vfx": {
		AssetTypes: []struct{ Name, Icon string }{
			{"Model", "cube"}, {"Texture", "texture"}, {"Rig", "bone"},
			{"Animation", "man-running"}, {"Lighting", "bulb"}, {"Compositing", "film-reel"},
			{"Simulation", "fire"}, {"Matchmove", "camera-flash"}, {"Matte Painting", "image"},
			{"Concept", "palette"},
		},
		CollectionTypes: []struct{ Name, Icon string }{
			{"Sequence", "film-strip"}, {"Shot", "camera"}, {"Asset", "cube"},
			{"Character", "masks"}, {"Environment", "tree"}, {"Library", "book"},
		},
	},
	"film": {
		AssetTypes: []struct{ Name, Icon string }{
			{"Storyboard", "image"}, {"Previz", "video-camera"}, {"Model", "cube"},
			{"Texture", "texture"}, {"Rig", "bone"}, {"Animation", "man-running"},
			{"Lighting", "bulb"}, {"Compositing", "film-reel"}, {"Edit", "clapboard"},
			{"Audio", "music"}, {"Color Grade", "palette"},
		},
		CollectionTypes: []struct{ Name, Icon string }{
			{"Sequence", "film-strip"}, {"Shot", "camera"}, {"Character", "masks"},
			{"Prop", "package"}, {"Environment", "tree"}, {"Set", "home"},
			{"Library", "book"},
		},
	},
}

func execSetupProjectTypes(projectPath string, args map[string]interface{}) ToolResult {
	projectType := getStringArg(args, "project_type", "")
	preset, ok := projectTypePresets[projectType]
	if !ok {
		return ToolResult{Success: false, Error: "project_type must be one of: animation, game, vfx, film"}
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

	createdAssetTypes := 0
	skippedAssetTypes := 0
	for _, at := range preset.AssetTypes {
		_, err := repository.GetOrCreateAssetType(tx, at.Name, at.Icon)
		if err != nil {
			skippedAssetTypes++
		} else {
			createdAssetTypes++
		}
	}

	createdCollectionTypes := 0
	skippedCollectionTypes := 0
	for _, ct := range preset.CollectionTypes {
		_, err := repository.GetOrCreateCollectionType(tx, ct.Name, ct.Icon)
		if err != nil {
			skippedCollectionTypes++
		} else {
			createdCollectionTypes++
		}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{
		"project_type":             projectType,
		"asset_types_created":      createdAssetTypes,
		"asset_types_skipped":      skippedAssetTypes,
		"collection_types_created": createdCollectionTypes,
		"collection_types_skipped": skippedCollectionTypes,
	}}
}

// --- Batch creation tools ---

func execBatchCreateCollections(projectPath string, args map[string]interface{}) ToolResult {
	items := getObjSliceArg(args, "collections")
	if len(items) == 0 {
		return ToolResult{Success: false, Error: "collections array is required and must not be empty"}
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

	collectionTypes, err := repository.GetCollectionTypes(tx)
	if err != nil || len(collectionTypes) == 0 {
		return ToolResult{Success: false, Error: "no collection types available in project"}
	}
	collectionTypeId := collectionTypes[0].Id

	nameToID := map[string]string{}

	type createdItem struct {
		Name string `json:"name"`
		ID   string `json:"id"`
		Path string `json:"path"`
	}
	var created []createdItem

	for i, item := range items {
		name := getStringArg(item, "name", "")
		if name == "" {
			return ToolResult{Success: false, Error: fmt.Sprintf("collection at index %d is missing 'name'", i)}
		}
		description := getStringArg(item, "description", "")
		parentID := getStringArg(item, "parent_id", "")

		if parentID == "" {
			parentName := getStringArg(item, "parent_name", "")
			if parentName != "" {
				resolved, ok := nameToID[parentName]
				if !ok {
					return ToolResult{Success: false, Error: fmt.Sprintf("parent_name '%s' for collection '%s' not found in previous items of this batch", parentName, name)}
				}
				parentID = resolved
			}
		}

		collection, err := repository.CreateCollection(tx, "", name, description, collectionTypeId, parentID, "", false)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to create collection '%s': %s", name, err.Error())}
		}

		nameToID[name] = collection.Id
		created = append(created, createdItem{Name: collection.Name, ID: collection.Id, Path: collection.CollectionPath})
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{"created": len(created), "collections": created}}
}

func execBatchCreateAssets(projectPath string, args map[string]interface{}) ToolResult {
	items := getObjSliceArg(args, "assets")
	if len(items) == 0 {
		return ToolResult{Success: false, Error: "assets array is required and must not be empty"}
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

	type createdItem struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	}
	var created []createdItem

	for i, item := range items {
		name := getStringArg(item, "name", "")
		collectionID := getStringArg(item, "collection_id", "")
		assetTypeID := getStringArg(item, "task_type_id", "")
		templateID := getStringArg(item, "template_id", "")
		tags := getStringSliceArg(item, "tags")

		if name == "" || assetTypeID == "" || templateID == "" {
			return ToolResult{Success: false, Error: fmt.Sprintf("asset at index %d is missing required fields (name, task_type_id, template_id)", i)}
		}

		asset, err := repository.CreateAsset(tx, "", name, assetTypeID, collectionID, false, templateID, "", "", tags, "", false, "", activeUserID(projectPath), "Created by Clustta Agent", "", func(int, int, string, string) {})
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to create asset '%s': %s", name, err.Error())}
		}

		err = repository.RevertToLatestCheckpoint(tx, asset.Id, asset.GetFilePath(), func(int, int, string, string) {})
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("asset '%s' created but failed to build file: %s", name, err.Error())}
		}

		created = append(created, createdItem{Name: asset.Name, ID: asset.Id})
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{"created": len(created), "assets": created}}
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

	collections, err := repository.GetCollections(tx, false)
	if err == nil && len(collections) > 0 {
		context.WriteString(fmt.Sprintf("Collections (%d):\n", len(collections)))
		for _, c := range collections {
			context.WriteString(fmt.Sprintf("- %s (ID: %s, type: %s", c.Name, c.Id, c.CollectionTypeName))
			if c.ParentId != "" {
				context.WriteString(fmt.Sprintf(", parent: %s", c.ParentId))
			}
			context.WriteString(")\n")
		}
		context.WriteString("\n")
	} else {
		context.WriteString("Collections: None yet\n\n")
	}

	for _, c := range collections {
		assets, err := repository.GetCollectionAssets(tx, c.Id)
		if err == nil && len(assets) > 0 {
			statusBreakdown := map[string]int{}
			for _, a := range assets {
				statusBreakdown[a.StatusShortName]++
			}
			parts := []string{}
			for status, count := range statusBreakdown {
				parts = append(parts, fmt.Sprintf("%s: %d", status, count))
			}
			context.WriteString(fmt.Sprintf("Assets in '%s': %d total (%s)\n", c.Name, len(assets), strings.Join(parts, ", ")))
		}
	}
	context.WriteString("\n")

	allAssets, err := repository.GetAssets(tx, false)
	if err == nil {
		rootCount := 0
		for _, a := range allAssets {
			if a.CollectionId == "" {
				rootCount++
			}
		}
		if rootCount > 0 {
			context.WriteString(fmt.Sprintf("Root-level assets (no collection): %d\n\n", rootCount))
		}
	}

	statuses, err := repository.GetStatuses(tx)
	if err == nil && len(statuses) > 0 {
		context.WriteString("Available statuses:\n")
		for _, s := range statuses {
			context.WriteString(fmt.Sprintf("- %s (ID: %s, short: %s)\n", s.Name, s.Id, s.ShortName))
		}
		context.WriteString("\n")
	}

	assetTypes, err := repository.GetAssetTypes(tx)
	if err == nil && len(assetTypes) > 0 {
		context.WriteString("Available asset types:\n")
		for _, at := range assetTypes {
			context.WriteString(fmt.Sprintf("- %s (ID: %s)\n", at.Name, at.Id))
		}
		context.WriteString("\n")
	}

	users, err := repository.GetUsers(tx)
	if err == nil && len(users) > 0 {
		context.WriteString("Users:\n")
		for _, u := range users {
			name := strings.TrimSpace(u.FirstName + " " + u.LastName)
			context.WriteString(fmt.Sprintf("- %s / %s (ID: %s)\n", u.Username, name, u.Id))
		}
		context.WriteString("\n")
	}

	tags, err := repository.GetTags(tx)
	if err == nil && len(tags) > 0 {
		context.WriteString("Tags:\n")
		for _, t := range tags {
			context.WriteString(fmt.Sprintf("- %s (ID: %s)\n", t.Name, t.Id))
		}
		context.WriteString("\n")
	}

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
// Large results are truncated to avoid blowing up the context window.
func SerializeToolResult(result ToolResult) string {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Sprintf(`{"success": false, "error": "failed to serialize result: %s"}`, err.Error())
	}
	content := string(data)
	if len(content) > 30000 {
		content = content[:30000] + `... [truncated - result too large. Use pagination or filters to narrow results.]"`
	}
	return content
}
