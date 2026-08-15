package scope

// EntityType is Clustta's existing browser entity discriminator.
type EntityType string

const (
	TypeAsset               EntityType = "asset"
	TypeCollection          EntityType = "collection"
	TypeUntrackedAsset      EntityType = "untracked_asset"
	TypeUntrackedCollection EntityType = "untracked_collection"
)

// Entity is the common envelope used by scoped agent commands.
type Entity struct {
	Type         EntityType             `json:"type"`
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Path         string                 `json:"path,omitempty"`
	ParentID     string                 `json:"parent_id,omitempty"`
	ParentPath   string                 `json:"parent_path,omitempty"`
	CollectionID string                 `json:"collection_id,omitempty"`
	Extension    string                 `json:"extension,omitempty"`
	Depth        int                    `json:"depth"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Request identifies the entities a command should operate on.
type Request struct {
	Source    string                 `json:"source"`
	EntityID  string                 `json:"entity_id,omitempty"`
	EntityIDs []string               `json:"entity_ids,omitempty"`
	Path      string                 `json:"path,omitempty"`
	Recursive bool                   `json:"recursive"`
	Types     []EntityType           `json:"types,omitempty"`
	Selection []Entity               `json:"selection,omitempty"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	Limit     int                    `json:"limit,omitempty"`
}

// Result is an ordered, deduplicated scope snapshot.
type Result struct {
	Request  Request  `json:"request"`
	Entities []Entity `json:"entities"`
}

func (t EntityType) Tracked() bool {
	return t == TypeAsset || t == TypeCollection
}

func (t EntityType) Valid() bool {
	switch t {
	case TypeAsset, TypeCollection, TypeUntrackedAsset, TypeUntrackedCollection:
		return true
	default:
		return false
	}
}
