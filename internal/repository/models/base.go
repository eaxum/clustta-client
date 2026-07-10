package models

import (
	"encoding/json"
	"strings"

	"github.com/jmoiron/sqlx"
)

type StudioUser struct {
	Id        string     `db:"id" json:"id"`
	MTime     int        `db:"mtime" json:"mtime"`
	AddedAt   string     `db:"added_at" json:"added_at"`
	Username  string     `db:"username" json:"username"`
	Email     string     `db:"email" json:"email"`
	FirstName string     `db:"first_name" json:"first_name"`
	LastName  string     `db:"last_name" json:"last_name"`
	Photo     []byte     `db:"photo" json:"photo"`
	RoleId    string     `db:"role_id" json:"role_id"`
	Synced    bool       `db:"synced" json:"synced"`
	Role      ServerRole `json:"role"`
}
type User struct {
	Id        string `db:"id" json:"id"`
	MTime     int    `db:"mtime" json:"mtime"`
	AddedAt   string `db:"added_at" json:"added_at"`
	Username  string `db:"username" json:"username"`
	Email     string `db:"email" json:"email"`
	FirstName string `db:"first_name" json:"first_name"`
	LastName  string `db:"last_name" json:"last_name"`
	Photo     []byte `db:"photo" json:"photo"`
	RoleId    string `db:"role_id" json:"role_id"`
	Synced    bool   `db:"synced" json:"synced"`
	Role      Role   `json:"role"`
}

type CollectionType struct {
	Id     string `db:"id" json:"id"`
	MTime  int    `db:"mtime" json:"mtime"`
	Name   string `db:"name" json:"name"`
	Icon   string `db:"icon" json:"icon"`
	Synced bool   `db:"synced" json:"synced"`
}
type AssetType struct {
	Id     string `db:"id" json:"id"`
	MTime  int    `db:"mtime" json:"mtime"`
	Name   string `db:"name" json:"name"`
	Icon   string `db:"icon" json:"icon"`
	Synced bool   `db:"synced" json:"synced"`
}

type Asset struct {
	Id              string `db:"id" json:"id"`
	MTime           int    `db:"mtime" json:"mtime"`
	CreatedAt       string `db:"created_at" json:"created_at"`
	Name            string `db:"name" json:"name"`
	Description     string `db:"description" json:"description"`
	Extension       string `db:"extension" json:"extension"`
	IsResource      bool   `db:"is_resource" json:"is_resource"`
	StatusId        string `db:"status_id" json:"status_id"`
	StatusShortName string `db:"status_short_name" json:"status_short_name"`
	AssetTypeId     string `db:"asset_type_id" json:"asset_type_id"`
	AssetTypeName   string `db:"asset_type_name" json:"asset_type_name"`
	AssetTypeIcon   string `db:"asset_type_icon" json:"asset_type_icon"`
	CollectionId    string `db:"collection_id" json:"collection_id"`
	CollectionName  string `db:"collection_name" json:"collection_name"`
	CollectionPath  string `db:"collection_path" json:"collection_path"`
	AssetPath       string `db:"asset_path" json:"asset_path"`
	AssigneeId      string `db:"assignee_id" json:"assignee_id"`
	AssigneeEmail   string `db:"assignee_email" json:"assignee_email"`
	AssigneeName    string `db:"assignee_name" json:"assignee_name"`
	AssignerId      string `db:"assigner_id" json:"assigner_id"`
	AssignerEmail   string `db:"assigner_email" json:"assigner_email"`
	AssignerName    string `db:"assigner_name" json:"assigner_name"`
	// RelationshipType string   `db:"relationship_type" json:"relationship_type"`
	IsDependency    bool     `db:"is_dependency" json:"is_dependency"`
	DependencyLevel int      `db:"dependency_level" json:"-"`
	FilePath        string   `db:"file_path" json:"file_path"`
	Tags            []string `db:"-" json:"tags"`
	TagsRaw         string   `db:"tags" json:"-"`
	// Tags             []string `db:"tags" json:"tags"`
	CollectionDependencies    []string `db:"-" json:"collection_dependencies"`
	CollectionDependenciesRaw string   `db:"collection_dependencies" json:"-"`
	Dependencies              []string `db:"-" json:"dependencies"`
	DependenciesRaw           string   `db:"dependencies" json:"-"`
	// Dependencies     []string `db:"dependencies" json:"dependencies"`
	FileStatus       string       `db:"file_status" json:"file_status"`
	Status           Status       `db:"status" json:"status"`
	IsLink           bool         `db:"is_link" json:"is_link"`
	Pointer          string       `db:"pointer" json:"pointer"`
	PreviewId        string       `db:"preview_id" json:"preview_id"`
	Preview          []byte       `db:"preview" json:"preview"`
	PreviewExtension string       `db:"preview_extension" json:"preview_extension"`
	Checkpoints      []Checkpoint `db:"-" json:"checkpoints"`
	Trashed          bool         `db:"trashed" json:"trashed"`
	Synced           bool         `db:"synced" json:"synced"`
}

func (t Asset) MarshalJSON() ([]byte, error) {
	type Alias Asset
	return json.Marshal(&struct {
		Alias
		Type string `json:"type"`
	}{
		Alias: Alias(t),
		Type:  "asset",
	})
}

func (t *Asset) GetFilePath() string {
	if t.Pointer == "" {
		return t.FilePath
	}
	return t.Pointer
}

type Collection struct {
	Id                 string   `db:"id" json:"id"`
	MTime              int      `db:"mtime" json:"mtime"`
	CreatedAt          string   `db:"created_at" json:"created_at"`
	Name               string   `db:"name" json:"name"`
	Description        string   `db:"description" json:"description"`
	CollectionPath     string   `db:"collection_path" json:"collection_path"`
	FilePath           string   `db:"file_path" json:"file_path"`
	Trashed            bool     `db:"trashed" json:"trashed"`
	CollectionTypeId   string   `db:"collection_type_id" json:"collection_type_id"`
	CollectionTypeIcon string   `db:"collection_type_icon" json:"collection_type_icon"`
	ParentId           string   `db:"parent_id" json:"parent_id"`
	AssigneeIdsRaw     string   `db:"assignee_ids" json:"-"`
	AssigneeIds        []string `db:"-" json:"assignee_ids"`
	CollectionTypeName string   `db:"collection_type_name" json:"collection_type_name"`
	PreviewId          string   `db:"preview_id" json:"preview_id"`
	Preview            []byte   `db:"preview" json:"preview"`
	PreviewExtension   string   `db:"preview_extension" json:"preview_extension"`
	Synced             bool     `db:"synced" json:"synced"`
	IsDependency       bool     `db:"is_dependency" json:"is_dependency"`
	IsShared           bool     `db:"is_shared" json:"is_shared"`
	CanModify          bool     `db:"can_modify" json:"can_modify"`
	Level              int      `db:"level" json:"-"`
	HasChildren        bool     `db:"-" json:"has_children"`
}
type CollectionAssignee struct {
	Id           string `db:"id" json:"id"`
	MTime        int    `db:"mtime" json:"mtime"`
	CollectionId string `db:"collection_id" json:"collection_id"`
	AssigneeId   string `db:"assignee_id" json:"assignee_id"`
	AssignerId   string `db:"assigner_id" json:"assigner_id"`
	Synced       bool   `db:"synced" json:"synced"`
}

func (e Collection) MarshalJSON() ([]byte, error) {
	type Alias Collection
	return json.Marshal(&struct {
		Alias
		Type string `json:"type"`
	}{
		Alias: Alias(e),
		Type:  "collection",
	})
}

func (e *Collection) GetFilePath() string {
	return e.FilePath
}

type UntrackedAsset struct {
	Id             string `db:"id" json:"id"`
	Name           string `db:"name" json:"name"`
	Extension      string `db:"extension" json:"extension"`
	CollectionId   string `db:"collection_id" json:"collection_id"`
	CollectionName string `db:"collection_name" json:"collection_name"`
	CollectionPath string `db:"collection_path" json:"collection_path"`
	AssetPath      string `db:"asset_path" json:"asset_path"`
	FilePath       string `db:"file_path" json:"file_path"`
	ItemPath       string `db:"item_type" json:"item_type"`
	AssetTypeIcon  string `db:"asset_type_icon" json:"asset_type_icon"`
	CanModify      bool   `db:"can_modify" json:"can_modify"`
}

func (ut UntrackedAsset) MarshalJSON() ([]byte, error) {
	type Alias UntrackedAsset
	return json.Marshal(&struct {
		Alias
		Type string `json:"type"`
	}{
		Alias: Alias(ut),
		Type:  "untracked_asset",
	})
}

type UntrackedCollection struct {
	Id             string `db:"id" json:"id"`
	Name           string `db:"name" json:"name"`
	CollectionPath string `db:"collection_path" json:"collection_path"`
	ItemPath       string `db:"item_path" json:"item_path"`
	FilePath       string `db:"file_path" json:"file_path"`
	ParentId       string `db:"parent_id" json:"parent_id"`
	CanModify      bool   `db:"can_modify" json:"can_modify"`
}

func (ue UntrackedCollection) MarshalJSON() ([]byte, error) {
	type Alias UntrackedCollection
	return json.Marshal(&struct {
		Alias
		Type string `json:"type"`
	}{
		Alias: Alias(ue),
		Type:  "untracked_collection",
	})
}

type AssetDependency struct {
	Id               string `db:"id" json:"id"`
	MTime            int    `db:"mtime" json:"mtime"`
	AssetId          string `db:"asset_id" json:"asset_id"`
	DependencyId     string `db:"dependency_id" json:"dependency_id"`
	DependencyTypeId string `db:"dependency_type_id" json:"dependency_type_id"`
	Synced           bool   `db:"synced" json:"synced"`
}
type CollectionDependency struct {
	Id               string `db:"id" json:"id"`
	MTime            int    `db:"mtime" json:"mtime"`
	AssetId          string `db:"asset_id" json:"asset_id"`
	DependencyId     string `db:"dependency_id" json:"dependency_id"`
	DependencyTypeId string `db:"dependency_type_id" json:"dependency_type_id"`
	Synced           bool   `db:"synced" json:"synced"`
}
type Workflow struct {
	Id          string               `db:"id" json:"id"`
	MTime       int                  `db:"mtime" json:"mtime"`
	Name        string               `db:"name" json:"name"`
	Synced      bool                 `db:"synced" json:"synced"`
	Assets      []WorkflowAsset      `db:"-" json:"assets"`
	Collections []WorkflowCollection `db:"-" json:"collections"`
	Links       []WorkflowLink       `db:"-" json:"links"`
}
type WorkflowAsset struct {
	Id                   string `db:"id" json:"id"`
	MTime                int    `db:"mtime" json:"mtime"`
	Name                 string `db:"name" json:"name"`
	TemplateId           string `db:"template_id" json:"template_id"`
	IsResource           bool   `db:"is_resource" json:"is_resource"`
	WorkflowId           string `db:"workflow_id" json:"workflow_id"`
	AssetTypeId          string `db:"asset_type_id" json:"asset_type_id"`
	WorkflowCollectionId string `db:"workflow_collection_id" json:"workflow_collection_id"`
	IsLink               bool   `db:"is_link" json:"is_link"`
	Pointer              string `db:"pointer" json:"pointer"`
	Synced               bool   `db:"synced" json:"synced"`
}
type WorkflowCollection struct {
	Id               string `db:"id" json:"id"`
	MTime            int    `db:"mtime" json:"mtime"`
	Name             string `db:"name" json:"name"`
	WorkflowId       string `db:"workflow_id" json:"workflow_id"`
	CollectionTypeId string `db:"collection_type_id" json:"collection_type_id"`
	ParentId         string `db:"parent_id" json:"parent_id"`
	Synced           bool   `db:"synced" json:"synced"`
}

type WorkflowLink struct {
	Id                 string `db:"id" json:"id"`
	MTime              int    `db:"mtime" json:"mtime"`
	Name               string `db:"name" json:"name"`
	CollectionTypeId   string `db:"collection_type_id" json:"collection_type_id"`
	WorkflowId         string `db:"workflow_id" json:"workflow_id"`
	LinkedWorkflowId   string `db:"linked_workflow_id" json:"linked_workflow_id"`
	LinkedWorkflowName string `db:"-" json:"linked_workflow_name"`
	Synced             bool   `db:"synced" json:"synced"`
}

type DependencyType struct {
	Id     string `db:"id" json:"id"`
	MTime  int    `db:"mtime" json:"mtime"`
	Name   string `db:"name" json:"name"`
	Synced bool   `db:"synced" json:"synced"`
}

type Status struct {
	Id        string `db:"id" json:"id"`
	MTime     int    `db:"mtime" json:"mtime"`
	Name      string `db:"name" json:"name"`
	ShortName string `db:"short_name" json:"short_name"`
	Color     string `db:"color" json:"color"`
	Synced    bool   `db:"synced" json:"synced"`
}

type Tag struct {
	Id     string `db:"id" json:"id"`
	MTime  int    `db:"mtime" json:"mtime"`
	Name   string `db:"name" json:"name"`
	Synced bool   `db:"synced" json:"synced"`
}

type AssetTag struct {
	Id      string `db:"id"`
	MTime   int    `db:"mtime" json:"mtime"`
	AssetId string `db:"asset_id"`
	TagId   string `db:"tag_id"`
	Synced  bool   `db:"synced" json:"synced"`
}

type Checkpoint struct {
	Id               string `db:"id" json:"id"`
	MTime            int    `db:"mtime" json:"mtime"`
	CreatedAt        string `db:"created_at" json:"created_at"`
	AssetId          string `db:"asset_id" json:"asset_id"`
	XXHashChecksum   string `db:"xxhash_checksum" json:"xxhash_checksum"`
	TimeModified     int    `db:"time_modified" json:"time_modified"`
	FileSize         int    `db:"file_size" json:"file_size"`
	Comment          string `db:"comment" json:"comment"`
	Chunks           string `db:"chunks" json:"chunks"`
	IsDownloaded     bool   `db:"is_downloaded" json:"is_downloaded"`
	AuthorUID        string `db:"author_id" json:"author_id"`
	GroupId          string `db:"group_id" json:"group_id"`
	PreviewId        string `db:"preview_id" json:"preview_id"`
	Preview          []byte `db:"preview" json:"preview"`
	PreviewExtension string `db:"preview_extension" json:"preview_extension"`
	Trashed          bool   `db:"trashed" json:"trashed"`
	Synced           bool   `db:"synced" json:"synced"`
}

func (cp *Checkpoint) HasMissingChunks(tx *sqlx.Tx) (bool, error) {
	chunkHashes := strings.Split(cp.Chunks, ",")
	for _, chunkHash := range chunkHashes {
		var hash string
		err := tx.Get(&hash, "SELECT hash FROM chunk WHERE hash = ?", chunkHash)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				cp.IsDownloaded = false
				return true, nil
			}
			return false, err
		}
		if hash != "" {
			continue
		}
		cp.IsDownloaded = false
		return true, nil
	}
	cp.IsDownloaded = true
	return false, nil
}

type Role struct {
	Id     string `db:"id" json:"id"`
	MTime  int    `db:"mtime" json:"mtime"`
	Name   string `db:"name" json:"name"`
	Synced bool   `db:"synced" json:"synced"`

	ViewCollection   bool `db:"view_collection" json:"view_collection"`
	CreateCollection bool `db:"create_collection" json:"create_collection"`
	UpdateCollection bool `db:"update_collection" json:"update_collection"`
	DeleteCollection bool `db:"delete_collection" json:"delete_collection"`

	ViewAsset   bool `db:"view_asset" json:"view_asset"`
	CreateAsset bool `db:"create_asset" json:"create_asset"`
	UpdateAsset bool `db:"update_asset" json:"update_asset"`
	DeleteAsset bool `db:"delete_asset" json:"delete_asset"`

	ViewTemplate   bool `db:"view_template" json:"view_template"`
	CreateTemplate bool `db:"create_template" json:"create_template"`
	UpdateTemplate bool `db:"update_template" json:"update_template"`
	DeleteTemplate bool `db:"delete_template" json:"delete_template"`

	ViewCheckpoint   bool `db:"view_checkpoint" json:"view_checkpoint"`
	CreateCheckpoint bool `db:"create_checkpoint" json:"create_checkpoint"`
	DeleteCheckpoint bool `db:"delete_checkpoint" json:"delete_checkpoint"`

	PullChunk bool `db:"pull_chunk" json:"pull_chunk"`

	AssignAsset   bool `db:"assign_asset" json:"assign_asset"`
	UnassignAsset bool `db:"unassign_asset" json:"unassign_asset"`

	AddUser    bool `db:"add_user" json:"add_user"`
	RemoveUser bool `db:"remove_user" json:"remove_user"`
	ChangeRole bool `db:"change_role" json:"change_role"`

	ChangeStatus   bool `db:"change_status" json:"change_status"`
	SetDoneAsset   bool `db:"set_done_asset" json:"set_done_asset"`
	SetRetakeAsset bool `db:"set_retake_asset" json:"set_retake_asset"`

	ViewDoneAsset bool `db:"view_done_asset" json:"view_done_asset"`

	ManageDependencies bool `db:"manage_dependencies" json:"manage_dependencies"`
	ManageShareLinks   bool `db:"manage_share_links" json:"manage_share_links"`
}

type RoleAttributes struct {
	ViewCollection   bool `db:"view_collection" json:"view_collection"`
	CreateCollection bool `db:"create_collection" json:"create_collection"`
	UpdateCollection bool `db:"update_collection" json:"update_collection"`
	DeleteCollection bool `db:"delete_collection" json:"delete_collection"`

	ViewAsset   bool `db:"view_asset" json:"view_asset"`
	CreateAsset bool `db:"create_asset" json:"create_asset"`
	UpdateAsset bool `db:"update_asset" json:"update_asset"`
	DeleteAsset bool `db:"delete_asset" json:"delete_asset"`

	ViewTemplate   bool `db:"view_template" json:"view_template"`
	CreateTemplate bool `db:"create_template" json:"create_template"`
	UpdateTemplate bool `db:"update_template" json:"update_template"`
	DeleteTemplate bool `db:"delete_template" json:"delete_template"`

	ViewCheckpoint   bool `db:"view_checkpoint" json:"view_checkpoint"`
	CreateCheckpoint bool `db:"create_checkpoint" json:"create_checkpoint"`
	DeleteCheckpoint bool `db:"delete_checkpoint" json:"delete_checkpoint"`

	PullChunk bool `db:"pull_chunk" json:"pull_chunk"`

	AssignAsset   bool `db:"assign_asset" json:"assign_asset"`
	UnassignAsset bool `db:"unassign_asset" json:"unassign_asset"`

	AddUser    bool `db:"add_user" json:"add_user"`
	RemoveUser bool `db:"remove_user" json:"remove_user"`
	ChangeRole bool `db:"change_role" json:"change_role"`

	ChangeStatus   bool `db:"change_status" json:"change_status"`
	SetDoneAsset   bool `db:"set_done_asset" json:"set_done_asset"`
	SetRetakeAsset bool `db:"set_retake_asset" json:"set_retake_asset"`

	ViewDoneAsset bool `db:"view_done_asset" json:"view_done_asset"`

	ManageDependencies bool `db:"manage_dependencies" json:"manage_dependencies"`
	ManageShareLinks   bool `db:"manage_share_links" json:"manage_share_links"`
}
type ServerRole struct {
	Id    string `db:"id" json:"id"`
	MTime int    `db:"mtime" json:"mtime"`
	Name  string `db:"name" json:"name"`

	ViewProject   bool `db:"view_project" json:"view_project"`
	CreateProject bool `db:"create_project" json:"create_project"`
	UpdateProject bool `db:"update_project" json:"update_project"`
	DeleteProject bool `db:"delete_project" json:"delete_project"`
}

type ServerRoleAttributes struct {
	ViewProject   bool `db:"view_project" json:"view_project"`
	CreateProject bool `db:"create_project" json:"create_project"`
	UpdateProject bool `db:"update_project" json:"update_project"`
	DeleteProject bool `db:"delete_project" json:"delete_project"`
}

type UserRole struct {
	MTime   int `db:"mtime" json:"mtime"`
	UserUID string
	RoleId  string
	Synced  bool `db:"synced" json:"synced"`
}
type UserServerRole struct {
	MTime   int `db:"mtime" json:"mtime"`
	UserUID string
	RoleId  string
	Synced  bool `db:"synced" json:"synced"`
}

type Template struct {
	Id             string `db:"id" json:"id"`
	MTime          int    `db:"mtime" json:"mtime"`
	Name           string `db:"name" json:"name"`
	Extension      string `db:"extension" json:"extension"`
	Chunks         string `db:"chunks" json:"chunks"`
	XxhashChecksum string `db:"xxhash_checksum" json:"xxhash_checksum"`
	FileSize       int    `db:"file_size" json:"file_size"`
	Trashed        bool   `db:"trashed" json:"trashed"`
	Synced         bool   `db:"synced" json:"synced"`
}
type Preview struct {
	Hash      string `db:"hash" json:"hash"`
	Preview   []byte `db:"preview" json:"preview"`
	Extension string `db:"extension" json:"extension"`
}
