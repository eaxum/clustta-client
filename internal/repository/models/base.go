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

type EntityType struct {
	Id     string `db:"id" json:"id"`
	MTime  int    `db:"mtime" json:"mtime"`
	Name   string `db:"name" json:"name"`
	Icon   string `db:"icon" json:"icon"`
	Synced bool   `db:"synced" json:"synced"`
}
type TaskType struct {
	Id     string `db:"id" json:"id"`
	MTime  int    `db:"mtime" json:"mtime"`
	Name   string `db:"name" json:"name"`
	Icon   string `db:"icon" json:"icon"`
	Synced bool   `db:"synced" json:"synced"`
}

type Task struct {
	Id              string `db:"id" json:"id"`
	MTime           int    `db:"mtime" json:"mtime"`
	CreatedAt       string `db:"created_at" json:"created_at"`
	Name            string `db:"name" json:"name"`
	Description     string `db:"description" json:"description"`
	Extension       string `db:"extension" json:"extension"`
	IsResource      bool   `db:"is_resource" json:"is_resource"`
	StatusId        string `db:"status_id" json:"status_id"`
	StatusShortName string `db:"status_short_name" json:"status_short_name"`
	TaskTypeId      string `db:"task_type_id" json:"task_type_id"`
	TaskTypeName    string `db:"task_type_name" json:"task_type_name"`
	TaskTypeIcon    string `db:"task_type_icon" json:"task_type_icon"`
	EntityId        string `db:"entity_id" json:"entity_id"`
	EntityName      string `db:"entity_name" json:"entity_name"`
	EntityPath      string `db:"entity_path" json:"entity_path"`
	TaskPath        string `db:"task_path" json:"task_path"`
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
	EntityDependencies    []string `db:"-" json:"entity_dependencies"`
	EntityDependenciesRaw string   `db:"entity_dependencies" json:"-"`
	Dependencies          []string `db:"-" json:"dependencies"`
	DependenciesRaw       string   `db:"dependencies" json:"-"`
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

func (t Task) MarshalJSON() ([]byte, error) {
	type Alias Task
	return json.Marshal(&struct {
		Alias
		Type string `json:"type"`
	}{
		Alias: Alias(t),
		Type:  "task",
	})
}

func (t *Task) GetFilePath() string {
	if t.Pointer == "" {
		return t.FilePath
	}
	return t.Pointer
}

type Entity struct {
	Id               string   `db:"id" json:"id"`
	MTime            int      `db:"mtime" json:"mtime"`
	CreatedAt        string   `db:"created_at" json:"created_at"`
	Name             string   `db:"name" json:"name"`
	Description      string   `db:"description" json:"description"`
	EntityPath       string   `db:"entity_path" json:"entity_path"`
	FilePath         string   `db:"file_path" json:"file_path"`
	Trashed          bool     `db:"trashed" json:"trashed"`
	EntityTypeId     string   `db:"entity_type_id" json:"entity_type_id"`
	EntityTypeIcon   string   `db:"entity_type_icon" json:"entity_type_icon"`
	ParentId         string   `db:"parent_id" json:"parent_id"`
	AssigneeIdsRaw   string   `db:"assignee_ids" json:"-"`
	AssigneeIds      []string `db:"-" json:"assignee_ids"`
	EntityTypeName   string   `db:"entity_type_name" json:"entity_type_name"`
	PreviewId        string   `db:"preview_id" json:"preview_id"`
	Preview          []byte   `db:"preview" json:"preview"`
	PreviewExtension string   `db:"preview_extension" json:"preview_extension"`
	Synced           bool     `db:"synced" json:"synced"`
	IsDependency     bool     `db:"is_dependency" json:"is_dependency"`
	IsLibrary        bool     `db:"is_library" json:"is_library"`
	CanModify        bool     `db:"can_modify" json:"can_modify"`
	Level            int      `db:"level" json:"-"`
	HasChildren      bool     `db:"-" json:"has_children"`
}
type EntityAssignee struct {
	Id         string `db:"id" json:"id"`
	MTime      int    `db:"mtime" json:"mtime"`
	EntityId   string `db:"entity_id" json:"entity_id"`
	AssigneeId string `db:"assignee_id" json:"assignee_id"`
	AssignerId string `db:"assigner_id" json:"assigner_id"`
	Synced     bool   `db:"synced" json:"synced"`
}

func (e Entity) MarshalJSON() ([]byte, error) {
	type Alias Entity
	return json.Marshal(&struct {
		Alias
		Type string `json:"type"`
	}{
		Alias: Alias(e),
		Type:  "entity",
	})
}

func (e *Entity) GetFilePath() string {
	return e.FilePath
}

type UntrackedTask struct {
	Id           string `db:"id" json:"id"`
	Name         string `db:"name" json:"name"`
	Extension    string `db:"extension" json:"extension"`
	EntityId     string `db:"entity_id" json:"entity_id"`
	EntityName   string `db:"entity_name" json:"entity_name"`
	EntityPath   string `db:"entity_path" json:"entity_path"`
	TaskPath     string `db:"task_path" json:"task_path"`
	FilePath     string `db:"file_path" json:"file_path"`
	ItemPath     string `db:"item_type" json:"item_type"`
	TaskTypeIcon string `db:"task_type_icon" json:"task_type_icon"`
}

func (ut UntrackedTask) MarshalJSON() ([]byte, error) {
	type Alias UntrackedTask
	return json.Marshal(&struct {
		Alias
		Type string `json:"type"`
	}{
		Alias: Alias(ut),
		Type:  "untracked_task",
	})
}

type UntrackedEntity struct {
	Id         string `db:"id" json:"id"`
	Name       string `db:"name" json:"name"`
	EntityPath string `db:"entity_path" json:"entity_path"`
	ItemPath   string `db:"item_path" json:"item_path"`
	FilePath   string `db:"file_path" json:"file_path"`
	ParentId   string `db:"parent_id" json:"parent_id"`
}

func (ue UntrackedEntity) MarshalJSON() ([]byte, error) {
	type Alias UntrackedEntity
	return json.Marshal(&struct {
		Alias
		Type string `json:"type"`
	}{
		Alias: Alias(ue),
		Type:  "untracked_entity",
	})
}

type TaskDependency struct {
	Id               string `db:"id" json:"id"`
	MTime            int    `db:"mtime" json:"mtime"`
	TaskId           string `db:"task_id" json:"task_id"`
	DependencyId     string `db:"dependency_id" json:"dependency_id"`
	DependencyTypeId string `db:"dependency_type_id" json:"dependency_type_id"`
	Synced           bool   `db:"synced" json:"synced"`
}
type EntityDependency struct {
	Id               string `db:"id" json:"id"`
	MTime            int    `db:"mtime" json:"mtime"`
	TaskId           string `db:"task_id" json:"task_id"`
	DependencyId     string `db:"dependency_id" json:"dependency_id"`
	DependencyTypeId string `db:"dependency_type_id" json:"dependency_type_id"`
	Synced           bool   `db:"synced" json:"synced"`
}
type Workflow struct {
	Id       string           `db:"id" json:"id"`
	MTime    int              `db:"mtime" json:"mtime"`
	Name     string           `db:"name" json:"name"`
	Synced   bool             `db:"synced" json:"synced"`
	Tasks    []WorkflowTask   `db:"-" json:"tasks"`
	Entities []WorkflowEntity `db:"-" json:"entities"`
	Links    []WorkflowLink   `db:"-" json:"links"`
}
type WorkflowTask struct {
	Id               string `db:"id" json:"id"`
	MTime            int    `db:"mtime" json:"mtime"`
	Name             string `db:"name" json:"name"`
	TemplateId       string `db:"template_id" json:"template_id"`
	IsResource       bool   `db:"is_resource" json:"is_resource"`
	WorkflowId       string `db:"workflow_id" json:"workflow_id"`
	TaskTypeId       string `db:"task_type_id" json:"task_type_id"`
	WorkflowEntityId string `db:"workflow_entity_id" json:"workflow_entity_id"`
	IsLink           bool   `db:"is_link" json:"is_link"`
	Pointer          string `db:"pointer" json:"pointer"`
	Synced           bool   `db:"synced" json:"synced"`
}
type WorkflowEntity struct {
	Id           string `db:"id" json:"id"`
	MTime        int    `db:"mtime" json:"mtime"`
	Name         string `db:"name" json:"name"`
	WorkflowId   string `db:"workflow_id" json:"workflow_id"`
	EntityTypeId string `db:"entity_type_id" json:"entity_type_id"`
	ParentId     string `db:"parent_id" json:"parent_id"`
	Synced       bool   `db:"synced" json:"synced"`
}

type WorkflowLink struct {
	Id                 string `db:"id" json:"id"`
	MTime              int    `db:"mtime" json:"mtime"`
	Name               string `db:"name" json:"name"`
	EntityTypeId       string `db:"entity_type_id" json:"entity_type_id"`
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

type TaskTag struct {
	Id     string `db:"id"`
	MTime  int    `db:"mtime" json:"mtime"`
	TaskId string `db:"task_id"`
	TagId  string `db:"tag_id"`
	Synced bool   `db:"synced" json:"synced"`
}

type Checkpoint struct {
	Id               string `db:"id" json:"id"`
	MTime            int    `db:"mtime" json:"mtime"`
	CreatedAt        string `db:"created_at" json:"created_at"`
	TaskId           string `db:"task_id" json:"task_id"`
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

	ViewEntity   bool `db:"view_entity" json:"view_entity"`
	CreateEntity bool `db:"create_entity" json:"create_entity"`
	UpdateEntity bool `db:"update_entity" json:"update_entity"`
	DeleteEntity bool `db:"delete_entity" json:"delete_entity"`

	ViewTask   bool `db:"view_task" json:"view_task"`
	CreateTask bool `db:"create_task" json:"create_task"`
	UpdateTask bool `db:"update_task" json:"update_task"`
	DeleteTask bool `db:"delete_task" json:"delete_task"`

	ViewTemplate   bool `db:"view_template" json:"view_template"`
	CreateTemplate bool `db:"create_template" json:"create_template"`
	UpdateTemplate bool `db:"update_template" json:"update_template"`
	DeleteTemplate bool `db:"delete_template" json:"delete_template"`

	ViewCheckpoint   bool `db:"view_checkpoint" json:"view_checkpoint"`
	CreateCheckpoint bool `db:"create_checkpoint" json:"create_checkpoint"`
	DeleteCheckpoint bool `db:"delete_checkpoint" json:"delete_checkpoint"`

	PullChunk bool `db:"pull_chunk" json:"pull_chunk"`

	AssignTask   bool `db:"assign_task" json:"assign_task"`
	UnassignTask bool `db:"unassign_task" json:"unassign_task"`

	AddUser    bool `db:"add_user" json:"add_user"`
	RemoveUser bool `db:"remove_user" json:"remove_user"`
	ChangeRole bool `db:"change_role" json:"change_role"`

	ChangeStatus  bool `db:"change_status" json:"change_status"`
	SetDoneTask   bool `db:"set_done_task" json:"set_done_task"`
	SetRetakeTask bool `db:"set_retake_task" json:"set_retake_task"`

	ViewDoneTask bool `db:"view_done_task" json:"view_done_task"`

	ManageDependencies bool `db:"manage_dependencies" json:"manage_dependencies"`
}

type RoleAttributes struct {
	ViewEntity   bool `db:"view_entity" json:"view_entity"`
	CreateEntity bool `db:"create_entity" json:"create_entity"`
	UpdateEntity bool `db:"update_entity" json:"update_entity"`
	DeleteEntity bool `db:"delete_entity" json:"delete_entity"`

	ViewTask   bool `db:"view_task" json:"view_task"`
	CreateTask bool `db:"create_task" json:"create_task"`
	UpdateTask bool `db:"update_task" json:"update_task"`
	DeleteTask bool `db:"delete_task" json:"delete_task"`

	ViewTemplate   bool `db:"view_template" json:"view_template"`
	CreateTemplate bool `db:"create_template" json:"create_template"`
	UpdateTemplate bool `db:"update_template" json:"update_template"`
	DeleteTemplate bool `db:"delete_template" json:"delete_template"`

	ViewCheckpoint   bool `db:"view_checkpoint" json:"view_checkpoint"`
	CreateCheckpoint bool `db:"create_checkpoint" json:"create_checkpoint"`
	DeleteCheckpoint bool `db:"delete_checkpoint" json:"delete_checkpoint"`

	PullChunk bool `db:"pull_chunk" json:"pull_chunk"`

	AssignTask   bool `db:"assign_task" json:"assign_task"`
	UnassignTask bool `db:"unassign_task" json:"unassign_task"`

	AddUser    bool `db:"add_user" json:"add_user"`
	RemoveUser bool `db:"remove_user" json:"remove_user"`
	ChangeRole bool `db:"change_role" json:"change_role"`

	ChangeStatus  bool `db:"change_status" json:"change_status"`
	SetDoneTask   bool `db:"set_done_task" json:"set_done_task"`
	SetRetakeTask bool `db:"set_retake_task" json:"set_retake_task"`

	ViewDoneTask bool `db:"view_done_task" json:"view_done_task"`

	ManageDependencies bool `db:"manage_dependencies" json:"manage_dependencies"`
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
