package agent

import (
	"clustta/internal/auth_service"
	"clustta/internal/constants"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/jmoiron/sqlx"
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
			Description: "Create a new asset (file). It can belong to a collection or exist at the project root. Requires a template - use list_templates first to find available templates.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the asset to create.",
					},
					"collection_id": map[string]interface{}{
						"type":        "string",
						"description": "Optional. ID of the collection to create the asset in. Omit to create at the project root.",
					},
					"task_type_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the asset type (e.g., Model, Rig, Animation). Use list_task_types to find available types.",
					},
					"template_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the file template to use. Use list_templates to find available templates.",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Optional list of tag names to apply.",
					},
				},
				"required": []string{"name", "task_type_id", "template_id"},
			},
		},
		{
			Name:        "batch_create_collections",
			Description: "Create multiple collections in one operation. Supports nesting: use 'parent_name' to reference another collection being created in the same batch, or 'parent_id' for an existing collection. All collections are created in a single transaction.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collections": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"name":        map[string]interface{}{"type": "string", "description": "Name of the collection."},
								"description": map[string]interface{}{"type": "string", "description": "Optional description."},
								"parent_id":   map[string]interface{}{"type": "string", "description": "Optional ID of an existing parent collection."},
								"parent_name": map[string]interface{}{"type": "string", "description": "Optional name of a parent collection from this same batch."},
							},
							"required": []string{"name"},
						},
						"description": "Array of collections to create.",
					},
				},
				"required": []string{"collections"},
			},
		},
		{
			Name:        "batch_create_assets",
			Description: "Create multiple assets in one operation. All assets are created in a single transaction. Use list_templates and list_task_types first to find valid IDs.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"assets": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"name":          map[string]interface{}{"type": "string", "description": "Name of the asset."},
								"collection_id": map[string]interface{}{"type": "string", "description": "Optional. ID of the target collection. Omit to create at the project root."},
								"task_type_id":  map[string]interface{}{"type": "string", "description": "ID of the asset type."},
								"template_id":   map[string]interface{}{"type": "string", "description": "ID of the file template."},
								"tags":          map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Optional tag names."},
							},
							"required": []string{"name", "task_type_id", "template_id"},
						},
						"description": "Array of assets to create.",
					},
				},
				"required": []string{"assets"},
			},
		},
		{
			Name:        "batch_add_tags",
			Description: "Add tags to multiple assets in one operation. Creates tags that don't exist yet.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"items": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"asset_id": map[string]interface{}{"type": "string", "description": "ID of the asset."},
								"tag_name": map[string]interface{}{"type": "string", "description": "Name of the tag to add."},
							},
							"required": []string{"asset_id", "tag_name"},
						},
						"description": "Array of asset-tag pairs.",
					},
				},
				"required": []string{"items"},
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
		{
			Name:        "bulk_delete_assets",
			Description: "Delete multiple assets in one operation. Provide specific asset_ids, OR use filters (name, status_id, task_type_id, assignee_id, tag_name) to select which assets to delete, OR set delete_all to true to delete every asset. This requires user confirmation.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Optional list of specific asset IDs to delete.",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Optional: delete assets matching this name substring.",
					},
					"status_id": map[string]interface{}{
						"type":        "string",
						"description": "Optional: delete assets with this status.",
					},
					"task_type_id": map[string]interface{}{
						"type":        "string",
						"description": "Optional: delete assets of this type.",
					},
					"assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "Optional: delete assets assigned to this user.",
					},
					"tag_name": map[string]interface{}{
						"type":        "string",
						"description": "Optional: delete assets with this tag.",
					},
					"delete_all": map[string]interface{}{
						"type":        "boolean",
						"description": "Set to true to delete all assets in the project.",
					},
					"remove_files": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to also remove the physical files from disk. Default false.",
					},
				},
				"required": []string{},
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
			Name:        "add_tag_to_asset",
			Description: "Add a tag to an asset by tag name. Creates the tag if it doesn't exist.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the asset to tag.",
					},
					"tag_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the tag to add.",
					},
				},
				"required": []string{"asset_id", "tag_name"},
			},
		},
		{
			Name:        "remove_tag_from_asset",
			Description: "Remove a tag from an asset. Use get_asset_details or list_tags to find the tag ID.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the asset.",
					},
					"tag_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the tag to remove.",
					},
				},
				"required": []string{"asset_id", "tag_id"},
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
			Name:        "add_dependency",
			Description: "Add a dependency between two assets. Use list_dependency_types to find available dependency type IDs.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the asset that depends on another.",
					},
					"dependency_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the asset being depended on.",
					},
					"dependency_type_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the dependency type.",
					},
				},
				"required": []string{"asset_id", "dependency_id", "dependency_type_id"},
			},
		},
		{
			Name:        "remove_dependency",
			Description: "Remove a dependency between two assets.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the asset.",
					},
					"dependency_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the dependency asset to unlink.",
					},
				},
				"required": []string{"asset_id", "dependency_id"},
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
			Name:        "create_collection_type",
			Description: "Create a new collection type (entity type). Uses the same icon set as asset types.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the collection type to create.",
					},
					"icon": map[string]interface{}{
						"type":        "string",
						"description": "Icon for the collection type. Must be one of: bezier, bone, book, boxes, bulb, camera-flash, camera, clapboard, compass, cube, drum, film-reel, film-strip, fire, flow-chart, four-squares, home, image, lamp, link, man-running, masks, music, mystery-ball, open-book, package, palette, scissors, shapes, stall, texture, tree, video-camera, website.",
						"enum":        []string{"bezier", "bone", "book", "boxes", "bulb", "camera-flash", "camera", "clapboard", "compass", "cube", "drum", "film-reel", "film-strip", "fire", "flow-chart", "four-squares", "home", "image", "lamp", "link", "man-running", "masks", "music", "mystery-ball", "open-book", "package", "palette", "scissors", "shapes", "stall", "texture", "tree", "video-camera", "website"},
					},
				},
				"required": []string{"name", "icon"},
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
			Name:        "bulk_change_status",
			Description: "Change the status of multiple assets at once. Accepts explicit asset_ids OR filter criteria (same as search_assets) to match assets server-side - no need to search first.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Explicit list of asset IDs. If provided, filters are ignored.",
					},
					"status_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the new status to set.",
					},
					"filter_name": map[string]interface{}{
						"type":        "string",
						"description": "Filter: substring match on asset name (case-insensitive).",
					},
					"filter_status_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter: match assets with this current status ID.",
					},
					"filter_task_type_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter: match assets with this asset type ID.",
					},
					"filter_assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter: match assets assigned to this user ID.",
					},
					"filter_tag_name": map[string]interface{}{
						"type":        "string",
						"description": "Filter: match assets with this tag.",
					},
					"filter_extension": map[string]interface{}{
						"type":        "string",
						"description": "Filter: match assets by file extension (e.g. 'clip', '.blend').",
					},
					"filter_unassigned": map[string]interface{}{
						"type":        "boolean",
						"description": "Filter: match only assets that have no assignee.",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Cap the number of matched assets to act on (applied after filtering / shuffling).",
					},
					"limit_fraction": map[string]interface{}{
						"type":        "number",
						"description": "Take this fraction (0-1) of the matched assets - e.g. 0.5 for half.",
					},
					"random": map[string]interface{}{
						"type":        "boolean",
						"description": "Shuffle the matched assets before applying limit / limit_fraction (use this for 'random half' etc.).",
					},
				},
				"required": []string{"status_id"},
			},
		},
		{
			Name:        "bulk_assign",
			Description: "Assign a user to multiple assets at once. Accepts explicit asset_ids OR filter criteria (same as search_assets) to match assets server-side - no need to search first.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Explicit list of asset IDs. If provided, filters are ignored.",
					},
					"user_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the user to assign to all listed assets.",
					},
					"filter_name": map[string]interface{}{
						"type":        "string",
						"description": "Filter: substring match on asset name (case-insensitive).",
					},
					"filter_status_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter: match assets with this status ID.",
					},
					"filter_task_type_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter: match assets with this asset type ID.",
					},
					"filter_assignee_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter: match assets assigned to this user ID.",
					},
					"filter_tag_name": map[string]interface{}{
						"type":        "string",
						"description": "Filter: match assets with this tag.",
					},
					"filter_extension": map[string]interface{}{
						"type":        "string",
						"description": "Filter: match assets by file extension (e.g. 'clip', '.blend').",
					},
					"filter_unassigned": map[string]interface{}{
						"type":        "boolean",
						"description": "Filter: match only assets that have no assignee.",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Cap the number of matched assets to act on (applied after filtering / shuffling).",
					},
					"limit_fraction": map[string]interface{}{
						"type":        "number",
						"description": "Take this fraction (0-1) of the matched assets - e.g. 0.5 for half.",
					},
					"random": map[string]interface{}{
						"type":        "boolean",
						"description": "Shuffle the matched assets before applying limit / limit_fraction (use this for 'random half' etc.).",
					},
				},
				"required": []string{"user_id"},
			},
		},
		{
			Name:        "unassign_all_assets",
			Description: "Remove assignments from multiple assets at once. If no asset_ids provided, unassigns all assets in the project.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Optional list of asset IDs to unassign. If empty, unassigns all assets.",
					},
				},
				"required": []string{},
			},
		},
		{
			Name:        "random_assign",
			Description: "Randomly distribute assets among a list of users using round-robin. Useful for evenly splitting work.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "List of asset IDs to distribute.",
					},
					"user_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "List of user IDs to distribute assets among.",
					},
				},
				"required": []string{"asset_ids", "user_ids"},
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
			Name:        "setup_project_types",
			Description: "Set up asset types and collection types for a specific project type (e.g., 'animation', 'game', 'vfx', 'film'). Creates a standard set of types appropriate for that pipeline. Use list_task_types and list_collection_types to see what already exists before calling this.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"animation", "game", "vfx", "film"},
						"description": "The type of creative project to set up types for.",
					},
				},
				"required": []string{"project_type"},
			},
		},

		animationSetupToolDef(),

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

	tools = append(tools, GetDCCToolDefinitions()...)

	iconEnum := []string{"bezier", "bone", "book", "boxes", "bulb", "camera-flash", "camera", "clapboard", "compass", "cube", "drum", "film-reel", "film-strip", "fire", "flow-chart", "four-squares", "home", "image", "lamp", "link", "man-running", "masks", "music", "mystery-ball", "open-book", "package", "palette", "scissors", "shapes", "stall", "texture", "tree", "video-camera", "website"}

	tools = append(tools, []ToolDefinition{
		// --- Type maintenance ---
		{
			Name:        "change_asset_type",
			Description: "Change the asset type (task type) of a single asset. Use list_task_types to find available type IDs.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_id":     map[string]interface{}{"type": "string", "description": "ID of the asset."},
					"task_type_id": map[string]interface{}{"type": "string", "description": "ID of the new asset type."},
				},
				"required": []string{"asset_id", "task_type_id"},
			},
		},
		{
			Name:        "bulk_change_asset_type",
			Description: "Change the asset type for multiple assets in one transaction.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_ids":    map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Asset IDs to update."},
					"task_type_id": map[string]interface{}{"type": "string", "description": "ID of the new asset type."},
				},
				"required": []string{"asset_ids", "task_type_id"},
			},
		},
		{
			Name:        "change_collection_type",
			Description: "Change the type of a single existing collection. Use list_collection_types to find available type IDs.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection_id":      map[string]interface{}{"type": "string", "description": "ID of the collection."},
					"collection_type_id": map[string]interface{}{"type": "string", "description": "ID of the new collection type."},
				},
				"required": []string{"collection_id", "collection_type_id"},
			},
		},
		{
			Name:        "bulk_change_collection_type",
			Description: "Change the type of multiple collections in one transaction.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection_ids":     map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Collection IDs to update."},
					"collection_type_id": map[string]interface{}{"type": "string", "description": "ID of the new collection type."},
				},
				"required": []string{"collection_ids", "collection_type_id"},
			},
		},
		{
			Name:        "batch_create_asset_types",
			Description: "Create multiple asset types in a single transaction. Each entry must include name and icon. Use this when the user asks to set up several types at once.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"items": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"name": map[string]interface{}{"type": "string"},
								"icon": map[string]interface{}{"type": "string", "enum": iconEnum},
							},
							"required": []string{"name", "icon"},
						},
					},
				},
				"required": []string{"items"},
			},
		},
		{
			Name:        "batch_create_collection_types",
			Description: "Create multiple collection types in a single transaction. Each entry must include name and icon.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"items": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"name": map[string]interface{}{"type": "string"},
								"icon": map[string]interface{}{"type": "string", "enum": iconEnum},
							},
							"required": []string{"name", "icon"},
						},
					},
				},
				"required": []string{"items"},
			},
		},
		{
			Name:        "update_asset_type",
			Description: "Update an existing asset type's name and/or icon. Pass empty string for fields you don't want to change (both required by the underlying API; pass current value to keep).",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":   map[string]interface{}{"type": "string", "description": "ID of the asset type."},
					"name": map[string]interface{}{"type": "string", "description": "New name."},
					"icon": map[string]interface{}{"type": "string", "description": "New icon name.", "enum": iconEnum},
				},
				"required": []string{"id", "name", "icon"},
			},
		},
		{
			Name:        "update_collection_type",
			Description: "Update an existing collection type's name and icon.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":   map[string]interface{}{"type": "string", "description": "ID of the collection type."},
					"name": map[string]interface{}{"type": "string", "description": "New name."},
					"icon": map[string]interface{}{"type": "string", "description": "New icon name.", "enum": iconEnum},
				},
				"required": []string{"id", "name", "icon"},
			},
		},
		{
			Name:        "batch_update_asset_types",
			Description: "Update multiple asset types' names and icons in a single transaction. Each entry must include id, name and icon.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"items": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"id":   map[string]interface{}{"type": "string"},
								"name": map[string]interface{}{"type": "string"},
								"icon": map[string]interface{}{"type": "string", "enum": iconEnum},
							},
							"required": []string{"id", "name", "icon"},
						},
					},
				},
				"required": []string{"items"},
			},
		},
		{
			Name:        "batch_update_collection_types",
			Description: "Update multiple collection types' names and icons in a single transaction.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"items": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"id":   map[string]interface{}{"type": "string"},
								"name": map[string]interface{}{"type": "string"},
								"icon": map[string]interface{}{"type": "string", "enum": iconEnum},
							},
							"required": []string{"id", "name", "icon"},
						},
					},
				},
				"required": []string{"items"},
			},
		},

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
		{
			Name:        "apply_workflow",
			Description: "Apply (instantiate) a workflow into the project. This creates a new root collection (using collection_type_id) under parent_id and populates it with all entities and assets defined by the workflow. Use list_workflows to find workflow_id and list_collection_types for collection_type_id.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"workflow_id":        map[string]interface{}{"type": "string", "description": "ID of the workflow to apply."},
					"name":               map[string]interface{}{"type": "string", "description": "Name for the root collection that will hold the workflow's contents."},
					"collection_type_id": map[string]interface{}{"type": "string", "description": "Collection type ID to use for the root collection."},
					"parent_id":          map[string]interface{}{"type": "string", "description": "Optional parent collection ID. Empty string for project root."},
				},
				"required": []string{"workflow_id", "name", "collection_type_id"},
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

	return tools
}

// toolPermission maps a tool to the Role field getter that must return true.
type toolPermission struct {
	Check func(role models.Role) bool
	Label string
}

// toolPermissions maps mutating tools to the permission(s) they require.
var toolPermissions = map[string]toolPermission{
	"create_collection":        {func(r models.Role) bool { return r.CreateCollection }, "Create Collection"},
	"batch_create_collections": {func(r models.Role) bool { return r.CreateCollection }, "Create Collection"},
	"rename_collection":        {func(r models.Role) bool { return r.UpdateCollection }, "Update Collection"},
	"delete_collection":        {func(r models.Role) bool { return r.DeleteCollection }, "Delete Collection"},

	"create_asset":           {func(r models.Role) bool { return r.CreateAsset }, "Create Asset"},
	"batch_create_assets":    {func(r models.Role) bool { return r.CreateAsset }, "Create Asset"},
	"create_asset_type":      {func(r models.Role) bool { return r.CreateAsset }, "Create Asset"},
	"create_collection_type": {func(r models.Role) bool { return r.CreateAsset }, "Create Asset"},
	"create_tag":             {func(r models.Role) bool { return r.CreateAsset }, "Create Asset"},
	"rename_asset":           {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"move_assets":            {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"add_tag_to_asset":       {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"remove_tag_from_asset":  {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"batch_add_tags":         {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"delete_asset":           {func(r models.Role) bool { return r.DeleteAsset }, "Delete Asset"},
	"bulk_delete_assets":     {func(r models.Role) bool { return r.DeleteAsset }, "Delete Asset"},
	"delete_asset_type":      {func(r models.Role) bool { return r.DeleteAsset }, "Delete Asset"},
	"delete_collection_type": {func(r models.Role) bool { return r.DeleteAsset }, "Delete Asset"},

	"assign_asset":        {func(r models.Role) bool { return r.AssignAsset }, "Assign Asset"},
	"bulk_assign":         {func(r models.Role) bool { return r.AssignAsset }, "Assign Asset"},
	"random_assign":       {func(r models.Role) bool { return r.AssignAsset }, "Assign Asset"},
	"unassign_asset":      {func(r models.Role) bool { return r.UnassignAsset }, "Unassign Asset"},
	"unassign_all_assets": {func(r models.Role) bool { return r.UnassignAsset }, "Unassign Asset"},

	"change_asset_status": {func(r models.Role) bool { return r.ChangeStatus }, "Change Status"},
	"bulk_change_status":  {func(r models.Role) bool { return r.ChangeStatus }, "Change Status"},

	"add_dependency":    {func(r models.Role) bool { return r.ManageDependencies }, "Manage Dependencies"},
	"remove_dependency": {func(r models.Role) bool { return r.ManageDependencies }, "Manage Dependencies"},

	"remove_user":              {func(r models.Role) bool { return r.RemoveUser }, "Remove User"},
	"change_collaborator_role": {func(r models.Role) bool { return r.ChangeRole }, "Change Role"},
	"update_role":              {func(r models.Role) bool { return r.ChangeRole }, "Change Role"},

	"add_project_collaborator":    {func(r models.Role) bool { return r.AddUser }, "Add User"},
	"remove_project_collaborator": {func(r models.Role) bool { return r.RemoveUser }, "Remove User"},

	"change_asset_type":             {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"bulk_change_asset_type":        {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"change_collection_type":        {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"bulk_change_collection_type":   {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"update_asset_type":             {func(r models.Role) bool { return r.CreateAsset }, "Create Asset"},
	"update_collection_type":        {func(r models.Role) bool { return r.CreateAsset }, "Create Asset"},
	"batch_create_asset_types":      {func(r models.Role) bool { return r.CreateAsset }, "Create Asset"},
	"batch_create_collection_types": {func(r models.Role) bool { return r.CreateAsset }, "Create Asset"},
	"batch_update_asset_types":      {func(r models.Role) bool { return r.CreateAsset }, "Create Asset"},
	"batch_update_collection_types": {func(r models.Role) bool { return r.CreateAsset }, "Create Asset"},

	"apply_workflow": {func(r models.Role) bool { return r.CreateCollection }, "Create Collection"},

	"setup_project_types":        {func(r models.Role) bool { return r.CreateAsset }, "Create Asset"},
	"setup_animation_production": {func(r models.Role) bool { return r.CreateCollection }, "Create Collection"},
	"add_ignore_pattern":         {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"remove_ignore_pattern":      {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},

	"blender_render":       {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"blender_export":       {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"blender_run_script":   {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"blender_run_python":   {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"blender_set_settings": {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"blender_link":         {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
	"run_terminal_command": {func(r models.Role) bool { return r.UpdateAsset }, "Update Asset"},
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
	if err := checkPermission(projectPath, toolName); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

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
	case "get_my_permissions":
		return execGetMyPermissions(projectPath)
	case "get_user_activity":
		return execGetUserActivity(projectPath)
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
	case "bulk_delete_assets":
		return execBulkDeleteAssets(projectPath, args)
	case "list_checkpoints":
		return execListCheckpoints(projectPath, args)
	case "create_tag":
		return execCreateTag(projectPath, args)
	case "add_tag_to_asset":
		return execAddTagToAsset(projectPath, args)
	case "remove_tag_from_asset":
		return execRemoveTagFromAsset(projectPath, args)
	case "get_asset_tags":
		return execGetAssetTags(projectPath, args)
	case "list_dependencies":
		return execListDependencies(projectPath, args)
	case "add_dependency":
		return execAddDependency(projectPath, args)
	case "remove_dependency":
		return execRemoveDependency(projectPath, args)
	case "list_dependency_types":
		return execListDependencyTypes(projectPath)
	case "list_collection_types":
		return execListCollectionTypes(projectPath)
	case "create_collection_type":
		return execCreateCollectionType(projectPath, args)
	case "delete_asset_type":
		return execDeleteAssetType(projectPath, args)
	case "delete_collection_type":
		return execDeleteCollectionType(projectPath, args)
	case "search_assets":
		return execSearchAssets(projectPath, args)
	case "get_project_summary":
		return execGetProjectSummary(projectPath)
	case "bulk_assign":
		return execBulkAssign(projectPath, args)
	case "bulk_change_status":
		return execBulkChangeStatus(projectPath, args)
	case "unassign_all_assets":
		return execUnassignAllAssets(projectPath, args)
	case "random_assign":
		return execRandomAssign(projectPath, args)
	case "remove_user":
		return execRemoveUser(projectPath, args)
	case "list_ignore_patterns":
		return execListIgnorePatterns(projectPath)
	case "add_ignore_pattern":
		return execAddIgnorePattern(projectPath, args)
	case "remove_ignore_pattern":
		return execRemoveIgnorePattern(projectPath, args)
	case "setup_project_types":
		return execSetupProjectTypes(projectPath, args)
	case "setup_animation_production":
		return execSetupAnimationProduction(projectPath, args)
	case "batch_create_collections":
		return execBatchCreateCollections(projectPath, args)
	case "batch_create_assets":
		return execBatchCreateAssets(projectPath, args)
	case "batch_add_tags":
		return execBatchAddTags(projectPath, args)
	case "generate_script":
		return execGenerateScript(args)

	case "open_in_dcc":
		return execOpenInDCC(projectPath, args)
	case "blender_render":
		return execBlenderRender(projectPath, args)
	case "blender_export":
		return execBlenderExport(projectPath, args)
	case "blender_run_script":
		return execBlenderRunScript(projectPath, args)
	case "blender_run_python":
		return execBlenderRunPython(projectPath, args)
	case "blender_set_settings":
		return execBlenderSetSettings(projectPath, args)
	case "blender_link":
		return execBlenderLink(projectPath, args)
	case "run_terminal_command":
		return execRunTerminalCommand(projectPath, args)

	case "change_asset_type":
		return execChangeAssetType(projectPath, args)
	case "bulk_change_asset_type":
		return execBulkChangeAssetType(projectPath, args)
	case "change_collection_type":
		return execChangeCollectionType(projectPath, args)
	case "bulk_change_collection_type":
		return execBulkChangeCollectionType(projectPath, args)
	case "batch_create_asset_types":
		return execBatchCreateAssetTypes(projectPath, args)
	case "batch_create_collection_types":
		return execBatchCreateCollectionTypes(projectPath, args)
	case "update_asset_type":
		return execUpdateAssetType(projectPath, args)
	case "update_collection_type":
		return execUpdateCollectionType(projectPath, args)
	case "batch_update_asset_types":
		return execBatchUpdateAssetTypes(projectPath, args)
	case "batch_update_collection_types":
		return execBatchUpdateCollectionTypes(projectPath, args)
	case "reveal_asset_on_disk":
		return execRevealAssetOnDisk(projectPath, args)
	case "reveal_in_browser":
		return execRevealInBrowser(projectPath, args)
	case "list_workflows":
		return execListWorkflows(projectPath)
	case "apply_workflow":
		return execApplyWorkflow(projectPath, args)
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

	collections, err := repository.GetCollections(tx, false)
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
	summaries := make([]collectionSummary, 0, len(collections))
	for _, c := range collections {
		summaries = append(summaries, collectionSummary{
			ID:       c.Id,
			Name:     c.Name,
			TypeName: c.CollectionTypeName,
			ParentID: c.ParentId,
			Path:     c.CollectionPath,
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

	assets, err := repository.GetCollectionAssets(tx, collectionID)
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
	summaries := make([]assetSummary, 0, len(assets))
	for _, a := range assets {
		summaries = append(summaries, assetSummary{
			ID:           a.Id,
			Name:         a.Name,
			TypeName:     a.AssetTypeName,
			StatusName:   a.StatusShortName,
			AssigneeName: a.AssigneeName,
			Tags:         a.Tags,
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

	assetType, err := repository.CreateAssetType(tx, "", name, icon)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{
		"id":   assetType.Id,
		"name": assetType.Name,
	}}
}

func execCreateCollection(projectPath string, args map[string]interface{}) ToolResult {
	name := getStringArg(args, "name", "")
	if name == "" {
		return ToolResult{Success: false, Error: "name is required"}
	}
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

	collectionTypes, err := repository.GetCollectionTypes(tx)
	if err != nil || len(collectionTypes) == 0 {
		return ToolResult{Success: false, Error: "no collection types available in project"}
	}
	collectionTypeId := collectionTypes[0].Id

	collection, err := repository.CreateCollection(tx, "", name, "", collectionTypeId, parentID, "", false)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = tx.Commit()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{
		"id":   collection.Id,
		"name": collection.Name,
		"path": collection.CollectionPath,
	}}
}

func execCreateAsset(projectPath string, args map[string]interface{}) ToolResult {
	name := getStringArg(args, "name", "")
	collectionID := getStringArg(args, "collection_id", "")
	assetTypeID := getStringArg(args, "task_type_id", "")
	templateID := getStringArg(args, "template_id", "")
	tags := getStringSliceArg(args, "tags")

	if name == "" || assetTypeID == "" || templateID == "" {
		return ToolResult{Success: false, Error: "name, task_type_id, and template_id are required"}
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

	asset, err := repository.CreateAsset(tx, "", name, assetTypeID, collectionID, false, templateID, "", "", tags, "", false, "", activeUserID(projectPath), "Created by Clustta Agent", "", func(int, int, string, string) {})
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = repository.RevertToLatestCheckpoint(tx, asset.Id, asset.GetFilePath(), func(int, int, string, string) {})
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("asset created but failed to build file: %s", err.Error())}
	}

	err = tx.Commit()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{
		"id":   asset.Id,
		"name": asset.Name,
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

	_, err = repository.GetAsset(tx, assetID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	_, err = repository.RenameAsset(tx, assetID, newName)
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

	_, err = repository.RenameCollection(tx, collectionID, newName)
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

	err = repository.UpdateStatus(tx, assetID, statusID)
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

	err = repository.AssignAsset(tx, assetID, userID)
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

	err = repository.UnAssignAsset(tx, assetID)
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
		err = repository.ChangeCollection(tx, assetID, targetCollectionID)
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

	asset, err := repository.GetAsset(tx, assetID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = repository.DeleteAsset(tx, assetID, removeFiles, true)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = tx.Commit()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]string{"deleted": assetID, "name": asset.Name}}
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

	collection, err := repository.GetCollection(tx, collectionID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = repository.DeleteCollection(tx, collectionID, removeFiles, true)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	err = tx.Commit()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]string{"deleted": collectionID, "name": collection.Name}}
}

func execBulkDeleteAssets(projectPath string, args map[string]interface{}) ToolResult {
	removeFiles := getBoolArg(args, "remove_files", false)
	deleteAll := getBoolArg(args, "delete_all", false)
	assetIDs := getStringSliceArg(args, "asset_ids")
	nameFilter := strings.ToLower(getStringArg(args, "name", ""))
	statusFilter := getStringArg(args, "status_id", "")
	typeFilter := getStringArg(args, "task_type_id", "")
	assigneeFilter := getStringArg(args, "assignee_id", "")
	tagFilter := strings.ToLower(getStringArg(args, "tag_name", ""))

	hasFilter := nameFilter != "" || statusFilter != "" || typeFilter != "" || assigneeFilter != "" || tagFilter != ""
	if !deleteAll && len(assetIDs) == 0 && !hasFilter {
		return ToolResult{Success: false, Error: "provide asset_ids, filter criteria (name, status_id, etc.), or set delete_all to true"}
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

	if len(assetIDs) > 0 {
		deleted := 0
		for _, id := range assetIDs {
			err := repository.DeleteAsset(tx, id, removeFiles, true)
			if err != nil {
				return ToolResult{Success: false, Error: fmt.Sprintf("failed to delete asset %s: %s", id, err.Error())}
			}
			deleted++
		}
		if err := tx.Commit(); err != nil {
			return ToolResult{Success: false, Error: err.Error()}
		}
		return ToolResult{Success: true, Data: map[string]interface{}{"deleted": deleted}}
	}

	assets, err := repository.GetAssets(tx, false)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	var toDelete []string
	for _, a := range assets {
		if !deleteAll {
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
		}
		toDelete = append(toDelete, a.Id)
	}

	if len(toDelete) == 0 {
		return ToolResult{Success: true, Data: "No matching assets found to delete."}
	}

	for _, id := range toDelete {
		err := repository.DeleteAsset(tx, id, removeFiles, true)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to delete asset %s: %s", id, err.Error())}
		}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{"deleted": len(toDelete)}}
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

func execAddTagToAsset(projectPath string, args map[string]interface{}) ToolResult {
	assetID := getStringArg(args, "asset_id", "")
	tagName := getStringArg(args, "tag_name", "")
	if assetID == "" || tagName == "" {
		return ToolResult{Success: false, Error: "asset_id and tag_name are required"}
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

	err = repository.AddTagToAsset(tx, assetID, tagName)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]string{"asset_id": assetID, "tag": tagName}}
}

func execRemoveTagFromAsset(projectPath string, args map[string]interface{}) ToolResult {
	assetID := getStringArg(args, "asset_id", "")
	tagID := getStringArg(args, "tag_id", "")
	if assetID == "" || tagID == "" {
		return ToolResult{Success: false, Error: "asset_id and tag_id are required"}
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

	err = repository.RemoveTagFromAsset(tx, assetID, tagID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]string{"asset_id": assetID, "tag_id": tagID}}
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

func execAddDependency(projectPath string, args map[string]interface{}) ToolResult {
	assetID := getStringArg(args, "asset_id", "")
	depID := getStringArg(args, "dependency_id", "")
	depTypeID := getStringArg(args, "dependency_type_id", "")
	if assetID == "" || depID == "" || depTypeID == "" {
		return ToolResult{Success: false, Error: "asset_id, dependency_id, and dependency_type_id are required"}
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

	dep, err := repository.AddDependency(tx, "", assetID, depID, depTypeID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: dep}
}

func execRemoveDependency(projectPath string, args map[string]interface{}) ToolResult {
	assetID := getStringArg(args, "asset_id", "")
	depID := getStringArg(args, "dependency_id", "")
	if assetID == "" || depID == "" {
		return ToolResult{Success: false, Error: "asset_id and dependency_id are required"}
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

	err = repository.RemoveAssetDependency(tx, assetID, depID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]string{"asset_id": assetID, "dependency_id": depID}}
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

func execCreateCollectionType(projectPath string, args map[string]interface{}) ToolResult {
	name := getStringArg(args, "name", "")
	icon := getStringArg(args, "icon", "")
	if name == "" {
		return ToolResult{Success: false, Error: "name is required"}
	}
	if icon == "" || !validAssetTypeIcons[icon] {
		return ToolResult{Success: false, Error: "icon is required and must be a valid icon name"}
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

	ct, err := repository.CreateCollectionType(tx, "", name, icon)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{"id": ct.Id, "name": ct.Name}}
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

// --- Bulk assignment tools ---

// filterAssetIDs resolves asset IDs from explicit IDs or filter criteria.
func filterAssetIDs(tx *sqlx.Tx, args map[string]interface{}) ([]string, error) {
	explicitIDs := getStringSliceArg(args, "asset_ids")

	nameFilter := strings.ToLower(getStringArg(args, "filter_name", ""))
	statusFilter := getStringArg(args, "filter_status_id", "")
	typeFilter := getStringArg(args, "filter_task_type_id", "")
	assigneeFilter := getStringArg(args, "filter_assignee_id", "")
	tagFilter := strings.ToLower(getStringArg(args, "filter_tag_name", ""))
	extFilter := strings.ToLower(strings.TrimPrefix(getStringArg(args, "filter_extension", ""), "."))
	unassignedOnly := false
	if v, ok := args["filter_unassigned"].(bool); ok {
		unassignedOnly = v
	}

	var ids []string
	if len(explicitIDs) > 0 {
		ids = explicitIDs
	} else {
		hasFilter := nameFilter != "" || statusFilter != "" || typeFilter != "" ||
			assigneeFilter != "" || tagFilter != "" || extFilter != "" || unassignedOnly
		if !hasFilter {
			return nil, fmt.Errorf("either asset_ids or at least one filter (filter_name, filter_status_id, filter_task_type_id, filter_assignee_id, filter_tag_name, filter_extension, filter_unassigned) is required")
		}

		assets, err := repository.GetAssets(tx, false)
		if err != nil {
			return nil, err
		}

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
			if unassignedOnly && a.AssigneeId != "" {
				continue
			}
			if extFilter != "" {
				ae := strings.ToLower(strings.TrimPrefix(a.Extension, "."))
				if ae != extFilter {
					continue
				}
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
			ids = append(ids, a.Id)
		}
	}

	randomize := false
	if v, ok := args["random"].(bool); ok {
		randomize = v
	}
	limitFraction := 0.0
	if v, ok := args["limit_fraction"].(float64); ok {
		limitFraction = v
	}
	limit := 0
	if v, ok := args["limit"].(float64); ok {
		limit = int(v)
	}

	if randomize && len(ids) > 1 {
		rand.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })
	}
	if limitFraction > 0 && limitFraction < 1 && len(ids) > 0 {
		n := int(float64(len(ids)) * limitFraction)
		if n < 1 {
			n = 1
		}
		ids = ids[:n]
	}
	if limit > 0 && limit < len(ids) {
		ids = ids[:limit]
	}
	return ids, nil
}

// execBulkAssign assigns a user to multiple assets by explicit IDs or filter.
func execBulkAssign(projectPath string, args map[string]interface{}) ToolResult {
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

	assetIDs, err := filterAssetIDs(tx, args)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	if len(assetIDs) == 0 {
		return ToolResult{Success: true, Data: map[string]interface{}{"assigned": 0, "message": "no assets matched the criteria"}}
	}

	for _, assetID := range assetIDs {
		err = repository.AssignAsset(tx, assetID, userID)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to assign asset %s: %s", assetID, err.Error())}
		}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{"assigned": len(assetIDs), "user_id": userID}}
}

// execBulkChangeStatus changes the status of multiple assets by explicit IDs or filter.
func execBulkChangeStatus(projectPath string, args map[string]interface{}) ToolResult {
	statusID := getStringArg(args, "status_id", "")
	if statusID == "" {
		return ToolResult{Success: false, Error: "status_id is required"}
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

	assetIDs, err := filterAssetIDs(tx, args)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	if len(assetIDs) == 0 {
		return ToolResult{Success: true, Data: map[string]interface{}{"updated": 0, "message": "no assets matched the criteria"}}
	}

	for _, assetID := range assetIDs {
		err = repository.UpdateStatus(tx, assetID, statusID)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to update status for asset %s: %s", assetID, err.Error())}
		}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{"updated": len(assetIDs), "status_id": statusID}}
}

func execUnassignAllAssets(projectPath string, args map[string]interface{}) ToolResult {
	assetIDs := getStringSliceArg(args, "asset_ids")

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

	if len(assetIDs) == 0 {
		assets, err := repository.GetAssets(tx, false)
		if err != nil {
			return ToolResult{Success: false, Error: err.Error()}
		}
		for _, a := range assets {
			if a.AssigneeId != "" {
				assetIDs = append(assetIDs, a.Id)
			}
		}
	}

	if len(assetIDs) == 0 {
		return ToolResult{Success: true, Data: "No assets to unassign."}
	}

	err = repository.UnAssignAssets(tx, assetIDs)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{"unassigned": len(assetIDs)}}
}

func execRandomAssign(projectPath string, args map[string]interface{}) ToolResult {
	assetIDs := getStringSliceArg(args, "asset_ids")
	userIDs := getStringSliceArg(args, "user_ids")
	if len(assetIDs) == 0 || len(userIDs) == 0 {
		return ToolResult{Success: false, Error: "asset_ids and user_ids are required"}
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

	rand.Shuffle(len(assetIDs), func(i, j int) {
		assetIDs[i], assetIDs[j] = assetIDs[j], assetIDs[i]
	})

	assignments := map[string]int{}
	for i, assetID := range assetIDs {
		userID := userIDs[i%len(userIDs)]
		err = repository.AssignAsset(tx, assetID, userID)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to assign asset %s: %s", assetID, err.Error())}
		}
		assignments[userID]++
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{"total_assigned": len(assetIDs), "per_user": assignments}}
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

func execBatchAddTags(projectPath string, args map[string]interface{}) ToolResult {
	items := getObjSliceArg(args, "items")
	if len(items) == 0 {
		return ToolResult{Success: false, Error: "items array is required and must not be empty"}
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

	added := 0
	for i, item := range items {
		assetID := getStringArg(item, "asset_id", "")
		tagName := getStringArg(item, "tag_name", "")
		if assetID == "" || tagName == "" {
			return ToolResult{Success: false, Error: fmt.Sprintf("item at index %d is missing asset_id or tag_name", i)}
		}

		err := repository.AddTagToAsset(tx, assetID, tagName)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to add tag '%s' to asset %s: %s", tagName, assetID, err.Error())}
		}
		added++
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{"tags_added": added}}
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
