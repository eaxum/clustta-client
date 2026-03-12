package integrations

import (
	"errors"
	"sync"
)

// Integration defines the contract that all external integrations must implement.
// All operations are stateless - tokens are passed per call for flexibility.
type Integration interface {
	// Identity and metadata
	ID() string               // "kitsu", "clickup"
	Name() string             // "Kitsu", "ClickUp"
	GetInfo() IntegrationInfo // Full metadata for UI

	// Authentication (returns token to store locally)
	Authenticate(credentials map[string]string) (AuthResult, error)
	ValidateToken(token, apiUrl string) (bool, error)

	// Project operations (token passed per call)
	GetProjects(token, apiUrl string) ([]ExternalProject, error)
	GetProjectEntities(token, apiUrl, projectID string) ([]ExternalEntity, error)
	GetProjectTasks(token, apiUrl, projectID string) ([]ExternalTask, error)

	// Type discovery (for mapping configuration)
	GetEntityTypes(token, apiUrl, projectID string) ([]ExternalTypeInfo, error)
	GetTaskTypes(token, apiUrl, projectID string) ([]ExternalTypeInfo, error)

	// Push operations
	UpdateTaskStatus(token, apiUrl, taskID, status string) error
	UploadPreview(token, apiUrl, taskID, filePath, comment string) error
}

// AuthResult contains the result of an authentication attempt.
type AuthResult struct {
	Success      bool   `json:"success"`
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
	UserEmail    string `json:"user_email"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	Error        string `json:"error,omitempty"`
}

// ExternalUser represents a user in the external system.
type ExternalUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ExternalProject represents a project in the external system.
type ExternalProject struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Stats       string `json:"stats,omitempty"` // e.g., "12 episodes, 89 shots"
}

// ExternalEntity represents a hierarchy item in the external system.
// This can be an episode, sequence, shot, folder, list, task type, etc.
type ExternalEntity struct {
	ID           string                 `json:"id"`
	ParentID     string                 `json:"parent_id,omitempty"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"` // "episode", "sequence", "shot", "folder", "list", "task_type"
	Path         string                 `json:"path"` // Full path for display
	Children     []ExternalEntity       `json:"children,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	HasTasks     bool                   `json:"has_tasks"`
	TaskTypeName string                 `json:"task_type_name,omitempty"` // For task type entities
}

// ExternalTask represents a task/work item in the external system.
type ExternalTask struct {
	ID          string   `json:"id"`
	ParentID    string   `json:"parent_id"` // Parent entity ID
	Name        string   `json:"name"`
	Type        string   `json:"type"` // "task", "subtask"
	Status      string   `json:"status"`
	Assignees   []string `json:"assignees,omitempty"`
	DueDate     string   `json:"due_date,omitempty"`
	TaskType    string   `json:"task_type,omitempty"`    // e.g., "Animation", "Lighting"
	TaskTypeID  string   `json:"task_type_id,omitempty"` // External task type ID
	Description string   `json:"description,omitempty"`
}

// ProjectHierarchy contains the full hierarchy of a project.
type ProjectHierarchy struct {
	Project  ExternalProject  `json:"project"`
	Entities []ExternalEntity `json:"entities"`
	Tasks    []ExternalTask   `json:"tasks"`
}

// SyncPreview contains preview data for what will be synced.
type SyncPreview struct {
	IntegrationID string             `json:"integration_id"`
	PreviewItems  []PreviewItem      `json:"preview_items"` // Unified list of all items (collections, assets, virtual folders)
	Collections   []SyncCollection   `json:"collections"`   // Deprecated: use PreviewItems
	Assets        []SyncAsset        `json:"assets"`        // Deprecated: use PreviewItems
	MissingTypes  []MissingType      `json:"missing_types"`
	Summary       SyncPreviewSummary `json:"summary"`
}

// PreviewItem represents a unified item in the sync preview (collection, asset, or virtual folder).
type PreviewItem struct {
	ID                string `json:"id"`                 // Unique ID (external_id or generated for virtual)
	Name              string `json:"name"`               // Display name
	ItemType          string `json:"item_type"`          // "entity", "task", "virtual"
	CollectionPath    string `json:"collection_path"`    // Full path e.g. "/episodes/ep01/"
	ParentPath        string `json:"parent_path"`        // Parent's collection_path e.g. "/episodes/"
	ExternalID        string `json:"external_id"`        // ID in external system (empty for virtual)
	ExternalType      string `json:"external_type"`      // Type in external system
	ExternalTypeID    string `json:"external_type_id"`   // External type ID (for tasks)
	ExternalName      string `json:"external_name"`      // Name in external system
	TypeName          string `json:"type_name"`          // Clustta type name (entity_type or task_type)
	TypeIcon          string `json:"type_icon"`          // Icon for the type
	Action            string `json:"action"`             // "create", "link", "skip", "virtual"
	Selected          bool   `json:"selected"`           // User selected for sync
	IsVirtual         bool   `json:"is_virtual"`         // True for path segment folders
	HasChildren       bool   `json:"has_children"`       // True if has child items
	TemplateID        string `json:"template_id"`        // Clustta template ID for this task type
	TemplateExtension string `json:"template_extension"` // File extension from template (e.g., ".blend")
}

// SyncCollection represents a collection to be created or linked.
type SyncCollection struct {
	TempID           string `json:"temp_id"`            // Temporary ID for UI
	ExternalID       string `json:"external_id"`        // ID in external system
	ExternalType     string `json:"external_type"`      // "episode", "sequence", "shot", etc.
	ExternalName     string `json:"external_name"`      // Name in external system
	ExternalParentID string `json:"external_parent_id"` // Parent in external system
	ExternalPath     string `json:"external_path"`      // Full path in external system
	CollectionPath   string `json:"collection_path"`    // Proposed Clustta collection path
	Action           string `json:"action"`             // "create", "link", "skip"
	CollectionID     string `json:"collection_id"`      // Existing Clustta collection ID (if linking)
	EntityTypeName   string `json:"entity_type_name"`   // Clustta entity type to use
	EntityTypeIcon   string `json:"entity_type_icon"`   // Icon for the entity type
	Selected         bool   `json:"selected"`           // User selected for sync
}

// SyncAsset represents an asset to be created or linked.
type SyncAsset struct {
	TempID            string   `json:"temp_id"`
	ExternalID        string   `json:"external_id"`
	ExternalName      string   `json:"external_name"`
	ExternalParentID  string   `json:"external_parent_id"`
	ExternalType      string   `json:"external_type"`    // Task type name (e.g., "Animation")
	ExternalTypeID    string   `json:"external_type_id"` // External task type ID
	ExternalStatus    string   `json:"external_status"`
	ExternalAssignees []string `json:"external_assignees"`
	CollectionPath    string   `json:"collection_path"` // Parent collection path
	Action            string   `json:"action"`          // "create", "link", "skip"
	AssetID           string   `json:"asset_id"`        // Existing Clustta asset ID (if linking)
	TaskTypeName      string   `json:"task_type_name"`  // Clustta task type to use
	TaskTypeIcon      string   `json:"task_type_icon"`
	Selected          bool     `json:"selected"`
	TemplateID        string   `json:"template_id"`        // Clustta template ID for this task type
	TemplateExtension string   `json:"template_extension"` // File extension from template (e.g., ".blend")
}

// SyncPreviewSummary contains counts for the sync preview.
type SyncPreviewSummary struct {
	TotalCollections    int `json:"total_collections"`
	TotalAssets         int `json:"total_assets"`
	CollectionsToCreate int `json:"collections_to_create"`
	CollectionsToLink   int `json:"collections_to_link"`
	AssetsToCreate      int `json:"assets_to_create"`
	AssetsToLink        int `json:"assets_to_link"`
}

// IntegrationInfo contains metadata about an integration for UI display.
type IntegrationInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	AuthType    string `json:"auth_type"`  // "password", "oauth", "token"
	Configured  bool   `json:"configured"` // User has authenticated
}

// TypeMapping maps an external type name to a Clustta type.
type TypeMapping struct {
	ExternalName   string `json:"external_name"`   // Name in external system (e.g., "Animation")
	ExternalID     string `json:"external_id"`     // ID in external system
	ClustttaTypeID string `json:"clustta_type_id"` // Clustta entity_type or task_type ID
	ClustttaName   string `json:"clustta_name"`    // Clustta type name (e.g., "animation")
	ClustttaIcon   string `json:"clustta_icon"`    // Icon name
}

// SyncOptions contains configuration stored in integration_project.sync_options.
type SyncOptions struct {
	EntityTypeMappings map[string]TypeMapping `json:"entity_type_mappings"` // External entity type → Clustta collection type
	TaskTypeMappings   map[string]TypeMapping `json:"task_type_mappings"`   // External task type → Clustta asset type
	TaskTypeTemplates  map[string]string      `json:"task_type_templates"`  // External task type ID → Clustta template ID
	DirectoryStructure DirectoryStructure     `json:"directory_structure"`  // Folder path templates
	LastSyncAt         string                 `json:"last_sync_at"`
}

// DirectoryStructure defines path templates for synced items.
type DirectoryStructure struct {
	Preset string                 `json:"preset"` // "3d-animation", "custom"
	Style  string                 `json:"style"`  // "lowercase", "uppercase", "capitalize", "kebab-case"
	Paths  map[string]interface{} `json:"paths"`  // Template ID -> { name, icon, template }
}

// MissingType represents an external type that doesn't exist in Clustta.
type MissingType struct {
	ExternalName  string `json:"external_name"`  // Name in external system
	ExternalID    string `json:"external_id"`    // ID in external system
	TypeCategory  string `json:"type_category"`  // "entity" or "task"
	SuggestedName string `json:"suggested_name"` // Suggested Clustta name (lowercase, sanitized)
	SuggestedIcon string `json:"suggested_icon"` // Random icon suggestion
}

// ExternalTypeInfo represents a type definition from an external system.
type ExternalTypeInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var (
	registry     = make(map[string]Integration)
	registryLock sync.RWMutex
)

// Register adds an integration to the registry.
func Register(integration Integration) {
	registryLock.Lock()
	defer registryLock.Unlock()
	registry[integration.ID()] = integration
}

// Get retrieves an integration by ID.
func Get(id string) (Integration, error) {
	registryLock.RLock()
	defer registryLock.RUnlock()
	if integration, ok := registry[id]; ok {
		return integration, nil
	}
	return nil, errors.New("integration not found: " + id)
}

// GetAll returns all registered integrations.
func GetAll() []Integration {
	registryLock.RLock()
	defer registryLock.RUnlock()
	result := make([]Integration, 0, len(registry))
	for _, integration := range registry {
		result = append(result, integration)
	}
	return result
}

// GetAllInfo returns info about all registered integrations.
func GetAllInfo() []IntegrationInfo {
	integrations := GetAll()
	infos := make([]IntegrationInfo, 0, len(integrations))
	for _, integration := range integrations {
		infos = append(infos, integration.GetInfo())
	}
	return infos
}
