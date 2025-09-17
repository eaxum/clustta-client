package repository

import (
	"clustta/internal/chunk_service"
	"clustta/internal/repository/models"
	"clustta/internal/repository/repositorypb"
)

// --- Conversion helpers ---

func ToPbUsers(users []models.User) []*repositorypb.User {
	pb := make([]*repositorypb.User, len(users))
	for i, u := range users {
		pb[i] = &repositorypb.User{
			Id:        u.Id,
			Mtime:     int64(u.MTime),
			AddedAt:   u.AddedAt,
			Username:  u.Username,
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Photo:     u.Photo,
			RoleId:    u.RoleId,
			Synced:    u.Synced,
			Role:      u.Role.Name, // or other field if needed
		}
	}
	return pb
}

func ToPbEntityTypes(entityTypes []models.EntityType) []*repositorypb.EntityType {
	pb := make([]*repositorypb.EntityType, len(entityTypes))
	for i, et := range entityTypes {
		pb[i] = &repositorypb.EntityType{
			Id:     et.Id,
			Mtime:  int64(et.MTime),
			Name:   et.Name,
			Icon:   et.Icon,
			Synced: et.Synced,
		}
	}
	return pb
}

func ToPbTaskTypes(types []models.TaskType) []*repositorypb.TaskType {
	pb := make([]*repositorypb.TaskType, len(types))
	for i, t := range types {
		pb[i] = &repositorypb.TaskType{
			Id:     t.Id,
			Mtime:  int64(t.MTime),
			Name:   t.Name,
			Icon:   t.Icon,
			Synced: t.Synced,
		}
	}
	return pb
}

func ToPbTasks(tasks []models.Task) []*repositorypb.Task {
	pb := make([]*repositorypb.Task, len(tasks))
	for i, t := range tasks {
		pb[i] = &repositorypb.Task{
			Id:          t.Id,
			Mtime:       int64(t.MTime),
			CreatedAt:   t.CreatedAt,
			Name:        t.Name,
			Description: t.Description,
			Extension:   t.Extension,
			IsResource:  t.IsResource,
			StatusId:    t.StatusId,
			TaskTypeId:  t.TaskTypeId,
			EntityId:    t.EntityId,
			AssigneeId:  t.AssigneeId,
			AssignerId:  t.AssignerId,
			IsLink:      t.IsLink,
			Pointer:     t.Pointer,
			PreviewId:   t.PreviewId,
			Trashed:     t.Trashed,
			Synced:      t.Synced,
		}
	}
	return pb
}

func ToPbEntities(entities []models.Entity) []*repositorypb.Entity {
	pb := make([]*repositorypb.Entity, len(entities))
	for i, e := range entities {
		pb[i] = &repositorypb.Entity{
			Id:           e.Id,
			Mtime:        int64(e.MTime),
			CreatedAt:    e.CreatedAt,
			Name:         e.Name,
			Description:  e.Description,
			EntityPath:   e.EntityPath,
			Trashed:      e.Trashed,
			EntityTypeId: e.EntityTypeId,
			ParentId:     e.ParentId,
			PreviewId:    e.PreviewId,
			Synced:       e.Synced,
			IsLibrary:    e.IsLibrary,
		}
	}
	return pb
}

func ToPbEntityAssignees(entityAssignees []models.EntityAssignee) []*repositorypb.EntityAssignee {
	pb := make([]*repositorypb.EntityAssignee, len(entityAssignees))
	for i, ea := range entityAssignees {
		pb[i] = &repositorypb.EntityAssignee{
			Id:         ea.Id,
			Mtime:      int64(ea.MTime),
			EntityId:   ea.EntityId,
			AssigneeId: ea.AssigneeId,
			AssignerId: ea.AssignerId,
			Synced:     ea.Synced,
		}
	}
	return pb
}

func ToPbTaskDependencies(taskDependencies []models.TaskDependency) []*repositorypb.TaskDependency {
	pb := make([]*repositorypb.TaskDependency, len(taskDependencies))
	for i, td := range taskDependencies {
		pb[i] = &repositorypb.TaskDependency{
			Id:               td.Id,
			Mtime:            int64(td.MTime),
			TaskId:           td.TaskId,
			DependencyId:     td.DependencyId,
			DependencyTypeId: td.DependencyTypeId,
			Synced:           td.Synced,
		}
	}
	return pb
}

func ToPbEntityDependencies(entityDependencies []models.EntityDependency) []*repositorypb.EntityDependency {
	pb := make([]*repositorypb.EntityDependency, len(entityDependencies))
	for i, ed := range entityDependencies {
		pb[i] = &repositorypb.EntityDependency{
			Id:               ed.Id,
			Mtime:            int64(ed.MTime),
			TaskId:           ed.TaskId,
			DependencyId:     ed.DependencyId,
			DependencyTypeId: ed.DependencyTypeId,
			Synced:           ed.Synced,
		}
	}
	return pb
}

func ToPbWorkflows(workflows []models.Workflow) []*repositorypb.Workflow {
	pb := make([]*repositorypb.Workflow, len(workflows))
	for i, w := range workflows {
		pb[i] = &repositorypb.Workflow{
			Id:     w.Id,
			Mtime:  int64(w.MTime),
			Name:   w.Name,
			Synced: w.Synced,
		}
	}
	return pb
}

func ToPbWorkflowTasks(workflowTasks []models.WorkflowTask) []*repositorypb.WorkflowTask {
	pb := make([]*repositorypb.WorkflowTask, len(workflowTasks))
	for i, wt := range workflowTasks {
		pb[i] = &repositorypb.WorkflowTask{
			Id:               wt.Id,
			Mtime:            int64(wt.MTime),
			Name:             wt.Name,
			TemplateId:       wt.TemplateId,
			IsResource:       wt.IsResource,
			WorkflowId:       wt.WorkflowId,
			TaskTypeId:       wt.TaskTypeId,
			WorkflowEntityId: wt.WorkflowEntityId,
			IsLink:           wt.IsLink,
			Pointer:          wt.Pointer,
			Synced:           wt.Synced,
		}
	}
	return pb
}

func ToPbWorkflowEntities(workflowEntities []models.WorkflowEntity) []*repositorypb.WorkflowEntity {
	pb := make([]*repositorypb.WorkflowEntity, len(workflowEntities))
	for i, we := range workflowEntities {
		pb[i] = &repositorypb.WorkflowEntity{
			Id:           we.Id,
			Mtime:        int64(we.MTime),
			Name:         we.Name,
			WorkflowId:   we.WorkflowId,
			EntityTypeId: we.EntityTypeId,
			ParentId:     we.ParentId,
			Synced:       we.Synced,
		}
	}
	return pb
}

func ToPbWorkflowLinks(workflowLinks []models.WorkflowLink) []*repositorypb.WorkflowLink {
	pb := make([]*repositorypb.WorkflowLink, len(workflowLinks))
	for i, wl := range workflowLinks {
		pb[i] = &repositorypb.WorkflowLink{
			Id:                 wl.Id,
			Mtime:              int64(wl.MTime),
			Name:               wl.Name,
			EntityTypeId:       wl.EntityTypeId,
			WorkflowId:         wl.WorkflowId,
			LinkedWorkflowId:   wl.LinkedWorkflowId,
			LinkedWorkflowName: wl.LinkedWorkflowName,
			Synced:             wl.Synced,
		}
	}
	return pb
}

func ToPbDependencyTypes(dependencyTypes []models.DependencyType) []*repositorypb.DependencyType {
	pb := make([]*repositorypb.DependencyType, len(dependencyTypes))
	for i, dt := range dependencyTypes {
		pb[i] = &repositorypb.DependencyType{
			Id:     dt.Id,
			Mtime:  int64(dt.MTime),
			Name:   dt.Name,
			Synced: dt.Synced,
		}
	}
	return pb
}

func ToPbStatuses(statuses []models.Status) []*repositorypb.Status {
	pb := make([]*repositorypb.Status, len(statuses))
	for i, s := range statuses {
		pb[i] = &repositorypb.Status{
			Id:        s.Id,
			Mtime:     int64(s.MTime),
			Name:      s.Name,
			ShortName: s.ShortName,
			Color:     s.Color,
			Synced:    s.Synced,
		}
	}
	return pb
}

func ToPbTags(tags []models.Tag) []*repositorypb.Tag {
	pb := make([]*repositorypb.Tag, len(tags))
	for i, t := range tags {
		pb[i] = &repositorypb.Tag{
			Id:     t.Id,
			Mtime:  int64(t.MTime),
			Name:   t.Name,
			Synced: t.Synced,
		}
	}
	return pb
}

func ToPbTaskTags(taskTags []models.TaskTag) []*repositorypb.TaskTag {
	pb := make([]*repositorypb.TaskTag, len(taskTags))
	for i, tt := range taskTags {
		pb[i] = &repositorypb.TaskTag{
			Id:     tt.Id,
			Mtime:  int64(tt.MTime),
			TaskId: tt.TaskId,
			TagId:  tt.TagId,
			Synced: tt.Synced,
		}
	}
	return pb
}

func ToPbCheckpoints(checkpoints []models.Checkpoint) []*repositorypb.Checkpoint {
	pb := make([]*repositorypb.Checkpoint, len(checkpoints))
	for i, c := range checkpoints {
		pb[i] = &repositorypb.Checkpoint{
			Id:             c.Id,
			Mtime:          int64(c.MTime),
			CreatedAt:      c.CreatedAt,
			TaskId:         c.TaskId,
			XxhashChecksum: c.XXHashChecksum,
			TimeModified:   int64(c.TimeModified),
			FileSize:       int64(c.FileSize),
			Comment:        c.Comment,
			Chunks:         c.Chunks,
			AuthorUid:      c.AuthorUID,
			PreviewId:      c.PreviewId,
			Trashed:        c.Trashed,
			Synced:         c.Synced,
			GroupId:        c.GroupId,
		}
	}
	return pb
}

func ToPbRoles(roles []models.Role) []*repositorypb.Role {
	pb := make([]*repositorypb.Role, len(roles))
	for i, r := range roles {
		pb[i] = &repositorypb.Role{
			Id:     r.Id,
			Mtime:  int64(r.MTime),
			Name:   r.Name,
			Synced: r.Synced,

			ViewEntity:   r.ViewEntity,
			CreateEntity: r.CreateEntity,
			UpdateEntity: r.UpdateEntity,
			DeleteEntity: r.DeleteEntity,

			ViewTask:   r.ViewTask,
			CreateTask: r.CreateTask,
			UpdateTask: r.UpdateTask,
			DeleteTask: r.DeleteTask,

			ViewTemplate:   r.ViewTemplate,
			CreateTemplate: r.CreateTemplate,
			UpdateTemplate: r.UpdateTemplate,
			DeleteTemplate: r.DeleteTemplate,

			ViewCheckpoint:   r.ViewCheckpoint,
			CreateCheckpoint: r.CreateCheckpoint,
			DeleteCheckpoint: r.DeleteCheckpoint,

			PullChunk: r.PullChunk,

			AssignTask:   r.AssignTask,
			UnassignTask: r.UnassignTask,

			AddUser:    r.AddUser,
			RemoveUser: r.RemoveUser,
			ChangeRole: r.ChangeRole,

			ChangeStatus:  r.ChangeStatus,
			SetDoneTask:   r.SetDoneTask,
			SetRetakeTask: r.SetRetakeTask,

			ViewDoneTask:       r.ViewDoneTask,
			ManageDependencies: r.ManageDependencies,
		}
	}
	return pb
}

func ToPbUserRoles(userRoles []models.UserRole) []*repositorypb.UserRole {
	pb := make([]*repositorypb.UserRole, len(userRoles))
	for i, ur := range userRoles {
		pb[i] = &repositorypb.UserRole{
			Mtime:  int64(ur.MTime),
			UserId: ur.UserUID,
			RoleId: ur.RoleId,
			Synced: ur.Synced,
		}
	}
	return pb
}

func ToPbTemplates(templates []models.Template) []*repositorypb.Template {
	pb := make([]*repositorypb.Template, len(templates))
	for i, t := range templates {
		pb[i] = &repositorypb.Template{
			Id:             t.Id,
			Mtime:          int64(t.MTime),
			Name:           t.Name,
			Extension:      t.Extension,
			Chunks:         t.Chunks,
			XxhashChecksum: t.XxhashChecksum,
			FileSize:       int64(t.FileSize),
			Trashed:        t.Trashed,
			Synced:         t.Synced,
		}
	}
	return pb
}

func ToPbPreviews(previews []models.Preview) []*repositorypb.Preview {
	pb := make([]*repositorypb.Preview, len(previews))
	for i, p := range previews {
		pb[i] = &repositorypb.Preview{
			Hash:      p.Hash,
			Preview:   p.Preview,
			Extension: p.Extension,
		}
	}
	return pb
}

func ToPbChunkInfos(chunkInfos []chunk_service.ChunkInfo) []*repositorypb.ChunkInfo {
	pb := make([]*repositorypb.ChunkInfo, len(chunkInfos))
	for i, p := range chunkInfos {
		pb[i] = &repositorypb.ChunkInfo{
			Hash: p.Hash,
			Size: int64(p.Size),
		}
	}
	return pb
}

func ToPbTombs(tombs []Tomb) []*repositorypb.Tomb {
	pb := make([]*repositorypb.Tomb, len(tombs))
	for i, t := range tombs {
		pb[i] = &repositorypb.Tomb{
			Id:        t.Id,
			Mtime:     int64(t.Mtime),
			TableName: t.TableName,
			Synced:    t.Synced,
		}
	}
	return pb
}

// type FullTask struct {
// 	Id              string `db:"id" json:"id"`
// 	MTime           int    `db:"mtime" json:"mtime"`
// 	CreatedAt       string `db:"created_at" json:"created_at"`
// 	Name            string `db:"name" json:"name"`
// 	Description     string `db:"description" json:"description"`
// 	Extension       string `db:"extension" json:"extension"`
// 	IsResource      bool   `db:"is_resource" json:"is_resource"`
// 	StatusId        string `db:"status_id" json:"status_id"`
// 	StatusShortName string `db:"status_short_name" json:"status_short_name"`
// 	TaskTypeId      string `db:"task_type_id" json:"task_type_id"`
// 	TaskTypeName    string `db:"task_type_name" json:"task_type_name"`
// 	TaskTypeIcon    string `db:"task_type_icon" json:"task_type_icon"`
// 	EntityId        string `db:"entity_id" json:"entity_id"`
// 	EntityName      string `db:"entity_name" json:"entity_name"`
// 	EntityPath      string `db:"entity_path" json:"entity_path"`
// 	TaskPath        string `db:"task_path" json:"task_path"`
// 	AssigneeId      string `db:"assignee_id" json:"assignee_id"`
// 	AssigneeEmail   string `db:"assignee_email" json:"assignee_email"`
// 	AssigneeName    string `db:"assignee_name" json:"assignee_name"`
// 	AssignerId      string `db:"assigner_id" json:"assigner_id"`
// 	AssignerEmail   string `db:"assigner_email" json:"assigner_email"`
// 	AssignerName    string `db:"assigner_name" json:"assigner_name"`
// 	// RelationshipType string   `db:"relationship_type" json:"relationship_type"`
// 	IsDependency    bool     `db:"is_dependency" json:"is_dependency"`
// 	DependencyLevel int      `db:"dependency_level" json:"-"`
// 	FilePath        string   `db:"file_path" json:"file_path"`
// 	Tags            []string `db:"-" json:"tags"`
// 	TagsRaw         string   `db:"tags" json:"-"`
// 	// Tags             []string `db:"tags" json:"tags"`
// 	EntityDependencies    []string `db:"-" json:"entity_dependencies"`
// 	EntityDependenciesRaw string   `db:"entity_dependencies" json:"-"`
// 	Dependencies          []string `db:"-" json:"dependencies"`
// 	DependenciesRaw       string   `db:"dependencies" json:"-"`
// 	// Dependencies     []string `db:"dependencies" json:"dependencies"`
// 	FileStatus       string       `db:"file_status" json:"file_status"`
// 	Status           Status       `db:"status" json:"status"`
// 	IsLink           bool         `db:"is_link" json:"is_link"`
// 	Pointer          string       `db:"pointer" json:"pointer"`
// 	PreviewId        string       `db:"preview_id" json:"preview_id"`
// 	Preview          []byte       `db:"preview" json:"preview"`
// 	PreviewExtension string       `db:"preview_extension" json:"preview_extension"`
// 	Checkpoints      []Checkpoint `db:"-" json:"checkpoints"`
// 	Trashed          bool         `db:"trashed" json:"trashed"`
// 	Synced           bool         `db:"synced" json:"synced"`
// 	Type             string       `db:"type" json:"type"`
// }

func ToPbFullTasks(tasks []models.Task) []*repositorypb.FullTask {
	pb := make([]*repositorypb.FullTask, len(tasks))
	for i, t := range tasks {
		pb[i] = &repositorypb.FullTask{
			Id:                    t.Id,
			Mtime:                 int64(t.MTime),
			CreatedAt:             t.CreatedAt,
			Name:                  t.Name,
			Description:           t.Description,
			Extension:             t.Extension,
			IsResource:            t.IsResource,
			StatusId:              t.StatusId,
			StatusShortName:       t.StatusShortName,
			TaskTypeId:            t.TaskTypeId,
			TaskTypeName:          t.TaskTypeName,
			TaskTypeIcon:          t.TaskTypeIcon,
			EntityId:              t.EntityId,
			EntityName:            t.EntityName,
			EntityPath:            t.EntityPath,
			TaskPath:              t.TaskPath,
			AssigneeId:            t.AssigneeId,
			AssigneeEmail:         t.AssigneeEmail,
			AssigneeName:          t.AssigneeName,
			AssignerId:            t.AssignerId,
			AssignerEmail:         t.AssignerEmail,
			AssignerName:          t.AssignerName,
			IsDependency:          t.IsDependency,
			DependencyLevel:       int32(t.DependencyLevel),
			FilePath:              t.FilePath,
			Tags:                  t.Tags,
			TagsRaw:               t.TagsRaw,
			EntityDependencies:    t.EntityDependencies,
			EntityDependenciesRaw: t.EntityDependenciesRaw,
			Dependencies:          t.Dependencies,
			DependenciesRaw:       t.DependenciesRaw,
			FileStatus:            t.FileStatus,
			Status:                ToPbStatus(t.Status),
			IsLink:                t.IsLink,
			Pointer:               t.Pointer,
			PreviewId:             t.PreviewId,
			Preview:               t.Preview,
			PreviewExtension:      t.PreviewExtension,
			Checkpoints:           ToPbCheckpoints(t.Checkpoints),
			Trashed:               t.Trashed,
			Synced:                t.Synced,
			Type:                  "task",
		}
	}
	return pb
}

// Helper for single Status
func ToPbStatus(s models.Status) *repositorypb.Status {
	return &repositorypb.Status{
		Id:        s.Id,
		Mtime:     int64(s.MTime),
		Name:      s.Name,
		ShortName: s.ShortName,
		Color:     s.Color,
		Synced:    s.Synced,
	}
}

// --- From Conversion helpers ---

func FromPbUser(pb *repositorypb.User) models.User {
	return models.User{
		Id:        pb.Id,
		MTime:     int(pb.Mtime),
		AddedAt:   pb.AddedAt,
		Username:  pb.Username,
		Email:     pb.Email,
		FirstName: pb.FirstName,
		LastName:  pb.LastName,
		Photo:     pb.Photo,
		RoleId:    pb.RoleId,
		Synced:    pb.Synced,
		Role:      models.Role{Name: pb.Role},
	}
}

func FromPbUsers(pbs []*repositorypb.User) []models.User {
	users := make([]models.User, len(pbs))
	for i, pb := range pbs {
		users[i] = FromPbUser(pb)
	}
	return users
}

func FromPbEntityType(pb *repositorypb.EntityType) models.EntityType {
	return models.EntityType{
		Id:     pb.Id,
		MTime:  int(pb.Mtime),
		Name:   pb.Name,
		Icon:   pb.Icon,
		Synced: pb.Synced,
	}
}

func FromPbEntityTypes(pbs []*repositorypb.EntityType) []models.EntityType {
	entityTypes := make([]models.EntityType, len(pbs))
	for i, pb := range pbs {
		entityTypes[i] = FromPbEntityType(pb)
	}
	return entityTypes
}

func FromPbTaskType(pb *repositorypb.TaskType) models.TaskType {
	return models.TaskType{
		Id:     pb.Id,
		MTime:  int(pb.Mtime),
		Name:   pb.Name,
		Icon:   pb.Icon,
		Synced: pb.Synced,
	}
}

func FromPbTaskTypes(pbs []*repositorypb.TaskType) []models.TaskType {
	taskTypes := make([]models.TaskType, len(pbs))
	for i, pb := range pbs {
		taskTypes[i] = FromPbTaskType(pb)
	}
	return taskTypes
}

func FromPbTask(pb *repositorypb.Task) models.Task {
	return models.Task{
		Id:          pb.Id,
		MTime:       int(pb.Mtime),
		CreatedAt:   pb.CreatedAt,
		Name:        pb.Name,
		Description: pb.Description,
		Extension:   pb.Extension,
		IsResource:  pb.IsResource,
		StatusId:    pb.StatusId,
		TaskTypeId:  pb.TaskTypeId,
		EntityId:    pb.EntityId,
		AssigneeId:  pb.AssigneeId,
		AssignerId:  pb.AssignerId,
		IsLink:      pb.IsLink,
		Pointer:     pb.Pointer,
		PreviewId:   pb.PreviewId,
		Trashed:     pb.Trashed,
		Synced:      pb.Synced,
	}
}

func FromPbTasks(pbs []*repositorypb.Task) []models.Task {
	tasks := make([]models.Task, len(pbs))
	for i, pb := range pbs {
		tasks[i] = FromPbTask(pb)
	}
	return tasks
}

func FromPbEntity(pb *repositorypb.Entity) models.Entity {
	return models.Entity{
		Id:           pb.Id,
		MTime:        int(pb.Mtime),
		CreatedAt:    pb.CreatedAt,
		Name:         pb.Name,
		Description:  pb.Description,
		EntityPath:   pb.EntityPath,
		Trashed:      pb.Trashed,
		EntityTypeId: pb.EntityTypeId,
		ParentId:     pb.ParentId,
		PreviewId:    pb.PreviewId,
		Synced:       pb.Synced,
		IsLibrary:    pb.IsLibrary,
	}
}

func FromPbEntities(pbs []*repositorypb.Entity) []models.Entity {
	entities := make([]models.Entity, len(pbs))
	for i, pb := range pbs {
		entities[i] = FromPbEntity(pb)
	}
	return entities
}

func FromPbEntityAssignee(pb *repositorypb.EntityAssignee) models.EntityAssignee {
	return models.EntityAssignee{
		Id:         pb.Id,
		MTime:      int(pb.Mtime),
		EntityId:   pb.EntityId,
		AssigneeId: pb.AssigneeId,
		AssignerId: pb.AssignerId,
		Synced:     pb.Synced,
	}
}

func FromPbEntityAssignees(pbs []*repositorypb.EntityAssignee) []models.EntityAssignee {
	entityAssignees := make([]models.EntityAssignee, len(pbs))
	for i, pb := range pbs {
		entityAssignees[i] = FromPbEntityAssignee(pb)
	}
	return entityAssignees
}

func FromPbTaskDependency(pb *repositorypb.TaskDependency) models.TaskDependency {
	return models.TaskDependency{
		Id:               pb.Id,
		MTime:            int(pb.Mtime),
		TaskId:           pb.TaskId,
		DependencyId:     pb.DependencyId,
		DependencyTypeId: pb.DependencyTypeId,
		Synced:           pb.Synced,
	}
}

func FromPbTaskDependencies(pbs []*repositorypb.TaskDependency) []models.TaskDependency {
	taskDependencies := make([]models.TaskDependency, len(pbs))
	for i, pb := range pbs {
		taskDependencies[i] = FromPbTaskDependency(pb)
	}
	return taskDependencies
}

func FromPbEntityDependency(pb *repositorypb.EntityDependency) models.EntityDependency {
	return models.EntityDependency{
		Id:               pb.Id,
		MTime:            int(pb.Mtime),
		TaskId:           pb.TaskId,
		DependencyId:     pb.DependencyId,
		DependencyTypeId: pb.DependencyTypeId,
		Synced:           pb.Synced,
	}
}

func FromPbEntityDependencies(pbs []*repositorypb.EntityDependency) []models.EntityDependency {
	entityDependencies := make([]models.EntityDependency, len(pbs))
	for i, pb := range pbs {
		entityDependencies[i] = FromPbEntityDependency(pb)
	}
	return entityDependencies
}

func FromPbWorkflow(pb *repositorypb.Workflow) models.Workflow {
	return models.Workflow{
		Id:     pb.Id,
		MTime:  int(pb.Mtime),
		Name:   pb.Name,
		Synced: pb.Synced,
	}
}

func FromPbWorkflows(pbs []*repositorypb.Workflow) []models.Workflow {
	workflows := make([]models.Workflow, len(pbs))
	for i, pb := range pbs {
		workflows[i] = FromPbWorkflow(pb)
	}
	return workflows
}

func FromPbWorkflowTask(pb *repositorypb.WorkflowTask) models.WorkflowTask {
	return models.WorkflowTask{
		Id:               pb.Id,
		MTime:            int(pb.Mtime),
		Name:             pb.Name,
		TemplateId:       pb.TemplateId,
		IsResource:       pb.IsResource,
		WorkflowId:       pb.WorkflowId,
		TaskTypeId:       pb.TaskTypeId,
		WorkflowEntityId: pb.WorkflowEntityId,
		IsLink:           pb.IsLink,
		Pointer:          pb.Pointer,
		Synced:           pb.Synced,
	}
}

func FromPbWorkflowTasks(pbs []*repositorypb.WorkflowTask) []models.WorkflowTask {
	workflowTasks := make([]models.WorkflowTask, len(pbs))
	for i, pb := range pbs {
		workflowTasks[i] = FromPbWorkflowTask(pb)
	}
	return workflowTasks
}

func FromPbWorkflowEntity(pb *repositorypb.WorkflowEntity) models.WorkflowEntity {
	return models.WorkflowEntity{
		Id:           pb.Id,
		MTime:        int(pb.Mtime),
		Name:         pb.Name,
		WorkflowId:   pb.WorkflowId,
		EntityTypeId: pb.EntityTypeId,
		ParentId:     pb.ParentId,
		Synced:       pb.Synced,
	}
}

func FromPbWorkflowEntities(pbs []*repositorypb.WorkflowEntity) []models.WorkflowEntity {
	workflowEntities := make([]models.WorkflowEntity, len(pbs))
	for i, pb := range pbs {
		workflowEntities[i] = FromPbWorkflowEntity(pb)
	}
	return workflowEntities
}

func FromPbWorkflowLink(pb *repositorypb.WorkflowLink) models.WorkflowLink {
	return models.WorkflowLink{
		Id:                 pb.Id,
		MTime:              int(pb.Mtime),
		Name:               pb.Name,
		EntityTypeId:       pb.EntityTypeId,
		WorkflowId:         pb.WorkflowId,
		LinkedWorkflowId:   pb.LinkedWorkflowId,
		LinkedWorkflowName: pb.LinkedWorkflowName,
		Synced:             pb.Synced,
	}
}

func FromPbWorkflowLinks(pbs []*repositorypb.WorkflowLink) []models.WorkflowLink {
	workflowLinks := make([]models.WorkflowLink, len(pbs))
	for i, pb := range pbs {
		workflowLinks[i] = FromPbWorkflowLink(pb)
	}
	return workflowLinks
}

func FromPbDependencyType(pb *repositorypb.DependencyType) models.DependencyType {
	return models.DependencyType{
		Id:     pb.Id,
		MTime:  int(pb.Mtime),
		Name:   pb.Name,
		Synced: pb.Synced,
	}
}

func FromPbDependencyTypes(pbs []*repositorypb.DependencyType) []models.DependencyType {
	dependencyTypes := make([]models.DependencyType, len(pbs))
	for i, pb := range pbs {
		dependencyTypes[i] = FromPbDependencyType(pb)
	}
	return dependencyTypes
}

func FromPbStatus(pb *repositorypb.Status) models.Status {
	return models.Status{
		Id:        pb.Id,
		MTime:     int(pb.Mtime),
		Name:      pb.Name,
		ShortName: pb.ShortName,
		Color:     pb.Color,
		Synced:    pb.Synced,
	}
}

func FromPbStatuses(pbs []*repositorypb.Status) []models.Status {
	statuses := make([]models.Status, len(pbs))
	for i, pb := range pbs {
		statuses[i] = FromPbStatus(pb)
	}
	return statuses
}

func FromPbTag(pb *repositorypb.Tag) models.Tag {
	return models.Tag{
		Id:     pb.Id,
		MTime:  int(pb.Mtime),
		Name:   pb.Name,
		Synced: pb.Synced,
	}
}

func FromPbTags(pbs []*repositorypb.Tag) []models.Tag {
	tags := make([]models.Tag, len(pbs))
	for i, pb := range pbs {
		tags[i] = FromPbTag(pb)
	}
	return tags
}

func FromPbTaskTag(pb *repositorypb.TaskTag) models.TaskTag {
	return models.TaskTag{
		Id:     pb.Id,
		MTime:  int(pb.Mtime),
		TaskId: pb.TaskId,
		TagId:  pb.TagId,
		Synced: pb.Synced,
	}
}

func FromPbTaskTags(pbs []*repositorypb.TaskTag) []models.TaskTag {
	taskTags := make([]models.TaskTag, len(pbs))
	for i, pb := range pbs {
		taskTags[i] = FromPbTaskTag(pb)
	}
	return taskTags
}

func FromPbCheckpoint(pb *repositorypb.Checkpoint) models.Checkpoint {
	return models.Checkpoint{
		Id:             pb.Id,
		MTime:          int(pb.Mtime),
		CreatedAt:      pb.CreatedAt,
		TaskId:         pb.TaskId,
		XXHashChecksum: pb.XxhashChecksum,
		TimeModified:   int(pb.TimeModified),
		FileSize:       int(pb.FileSize),
		Comment:        pb.Comment,
		Chunks:         pb.Chunks,
		AuthorUID:      pb.AuthorUid,
		PreviewId:      pb.PreviewId,
		Trashed:        pb.Trashed,
		Synced:         pb.Synced,
		GroupId:        pb.GroupId,
	}
}

func FromPbCheckpoints(pbs []*repositorypb.Checkpoint) []models.Checkpoint {
	checkpoints := make([]models.Checkpoint, len(pbs))
	for i, pb := range pbs {
		checkpoints[i] = FromPbCheckpoint(pb)
	}
	return checkpoints
}

func FromPbRole(pb *repositorypb.Role) models.Role {
	return models.Role{
		Id:     pb.Id,
		MTime:  int(pb.Mtime),
		Name:   pb.Name,
		Synced: pb.Synced,

		ViewEntity:   pb.ViewEntity,
		CreateEntity: pb.CreateEntity,
		UpdateEntity: pb.UpdateEntity,
		DeleteEntity: pb.DeleteEntity,

		ViewTask:   pb.ViewTask,
		CreateTask: pb.CreateTask,
		UpdateTask: pb.UpdateTask,
		DeleteTask: pb.DeleteTask,

		ViewTemplate:   pb.ViewTemplate,
		CreateTemplate: pb.CreateTemplate,
		UpdateTemplate: pb.UpdateTemplate,
		DeleteTemplate: pb.DeleteTemplate,

		ViewCheckpoint:   pb.ViewCheckpoint,
		CreateCheckpoint: pb.CreateCheckpoint,
		DeleteCheckpoint: pb.DeleteCheckpoint,

		PullChunk: pb.PullChunk,

		AssignTask:   pb.AssignTask,
		UnassignTask: pb.UnassignTask,

		AddUser:    pb.AddUser,
		RemoveUser: pb.RemoveUser,
		ChangeRole: pb.ChangeRole,

		ChangeStatus:  pb.ChangeStatus,
		SetDoneTask:   pb.SetDoneTask,
		SetRetakeTask: pb.SetRetakeTask,

		ViewDoneTask:       pb.ViewDoneTask,
		ManageDependencies: pb.ManageDependencies,
	}
}

func FromPbRoles(pbs []*repositorypb.Role) []models.Role {
	roles := make([]models.Role, len(pbs))
	for i, pb := range pbs {
		roles[i] = FromPbRole(pb)
	}
	return roles
}

func FromPbUserRole(pb *repositorypb.UserRole) models.UserRole {
	return models.UserRole{
		MTime:   int(pb.Mtime),
		UserUID: pb.UserId,
		RoleId:  pb.RoleId,
		Synced:  pb.Synced,
	}
}

func FromPbUserRoles(pbs []*repositorypb.UserRole) []models.UserRole {
	userRoles := make([]models.UserRole, len(pbs))
	for i, pb := range pbs {
		userRoles[i] = FromPbUserRole(pb)
	}
	return userRoles
}

func FromPbTemplate(pb *repositorypb.Template) models.Template {
	return models.Template{
		Id:             pb.Id,
		MTime:          int(pb.Mtime),
		Name:           pb.Name,
		Extension:      pb.Extension,
		Chunks:         pb.Chunks,
		XxhashChecksum: pb.XxhashChecksum,
		FileSize:       int(pb.FileSize),
		Trashed:        pb.Trashed,
		Synced:         pb.Synced,
	}
}

func FromPbTemplates(pbs []*repositorypb.Template) []models.Template {
	templates := make([]models.Template, len(pbs))
	for i, pb := range pbs {
		templates[i] = FromPbTemplate(pb)
	}
	return templates
}

func FromPbPreview(pb *repositorypb.Preview) models.Preview {
	return models.Preview{
		Hash:      pb.Hash,
		Preview:   pb.Preview,
		Extension: pb.Extension,
	}
}

func FromPbPreviews(pbs []*repositorypb.Preview) []models.Preview {
	previews := make([]models.Preview, len(pbs))
	for i, pb := range pbs {
		previews[i] = FromPbPreview(pb)
	}
	return previews
}

func FromPbChunkInfo(pb *repositorypb.ChunkInfo) chunk_service.ChunkInfo {
	return chunk_service.ChunkInfo{
		Hash: pb.Hash,
		Size: int(pb.Size),
	}
}

func FromPbChunkInfos(pbs []*repositorypb.ChunkInfo) []chunk_service.ChunkInfo {
	chunkInfos := make([]chunk_service.ChunkInfo, len(pbs))
	for i, pb := range pbs {
		chunkInfos[i] = FromPbChunkInfo(pb)
	}
	return chunkInfos
}

func FromPbTomb(pb *repositorypb.Tomb) Tomb {
	return Tomb{
		Id:        pb.Id,
		Mtime:     int(pb.Mtime),
		TableName: pb.TableName,
		Synced:    pb.Synced,
	}
}

func FromPbTombs(pbs []*repositorypb.Tomb) []Tomb {
	tombs := make([]Tomb, len(pbs))
	for i, pb := range pbs {
		tombs[i] = FromPbTomb(pb)
	}
	return tombs
}
