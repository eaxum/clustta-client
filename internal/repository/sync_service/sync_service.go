package sync_service

import (
	"bytes"
	"clustta/internal/auth_service"
	"clustta/internal/chunk_service"
	"clustta/internal/constants"
	"clustta/internal/error_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/repositorypb"
	"clustta/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/DataDog/zstd"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"
)

type ProjectData struct {
	ProjectPreview     string                    `json:"project_preview"`
	Tasks              []models.Task             `json:"tasks"`
	TaskTypes          []models.TaskType         `json:"task_types"`
	TasksCheckpoints   []models.Checkpoint       `json:"tasks_checkpoints"`
	TaskDependencies   []models.TaskDependency   `json:"task_dependencies"`
	EntityDependencies []models.EntityDependency `json:"entity_dependencies"`

	Statuses        []models.Status         `json:"statuses"`
	DependencyTypes []models.DependencyType `json:"dependency_types"`

	Users []models.User `json:"users"`
	Roles []models.Role `json:"roles"`

	EntityTypes     []models.EntityType     `json:"entity_types"`
	Entities        []models.Entity         `json:"entities"`
	EntityAssignees []models.EntityAssignee `json:"entity_assignees"`

	Templates []models.Template `json:"templates"`
	Tags      []models.Tag      `json:"tags"`
	TasksTags []models.TaskTag  `json:"tasks_tags"`

	Workflows        []models.Workflow       `json:"workflows"`
	WorkflowLinks    []models.WorkflowLink   `json:"workflow_links"`
	WorkflowEntities []models.WorkflowEntity `json:"workflow_entities"`
	WorkflowTasks    []models.WorkflowTask   `json:"workflow_tasks"`

	Tombs []repository.Tomb `json:"tomb"`
}

func (d *ProjectData) IsEmpty() bool {
	return len(d.Tasks) == 0 &&
		len(d.TaskTypes) == 0 &&
		len(d.TasksCheckpoints) == 0 &&
		len(d.TaskDependencies) == 0 &&
		len(d.EntityDependencies) == 0 &&
		len(d.EntityTypes) == 0 &&
		len(d.Entities) == 0 &&
		len(d.EntityAssignees) == 0 &&
		len(d.Templates) == 0 &&
		len(d.Tags) == 0 &&
		len(d.TasksTags) == 0 &&
		len(d.Statuses) == 0 &&
		len(d.DependencyTypes) == 0 &&
		len(d.Users) == 0 &&
		len(d.Roles) == 0 &&
		len(d.Workflows) == 0 &&
		len(d.WorkflowLinks) == 0 &&
		len(d.WorkflowEntities) == 0 &&
		len(d.WorkflowTasks) == 0 &&
		len(d.Tombs) == 0 &&
		d.ProjectPreview == ""
}

func WriteProjectData(tx *sqlx.Tx, data ProjectData, strict bool) error {

	// Sort
	sortedEntities, err := repository.TopologicalSort(data.Entities)
	if err != nil {
		return err
	}
	data.Entities = sortedEntities
	// sort.Slice(data.Entities, func(i, j int) bool {
	// 	iDepth := strings.Count(data.Entities[i].EntityPath, "/")
	// 	jDepth := strings.Count(data.Entities[j].EntityPath, "/")
	// 	if iDepth != jDepth {
	// 		return iDepth < jDepth
	// 	}
	// 	return data.Entities[i].EntityPath < data.Entities[j].EntityPath
	// })
	// sort.Slice(data.Tasks, func(i, j int) bool {
	// 	iDepth := strings.Count(data.Tasks[i].TaskPath, "/")
	// 	jDepth := strings.Count(data.Tasks[j].TaskPath, "/")
	// 	if iDepth != jDepth {
	// 		return iDepth < jDepth
	// 	}
	// 	return data.Tasks[i].TaskPath < data.Tasks[j].TaskPath
	// })

	tombItems := make(map[string]bool)
	tombedItems, err := repository.GetTombedItems(tx)
	if err != nil {
		return err
	}
	for _, tombItem := range tombedItems {
		tombItems[tombItem] = true
	}

	chunks := []string{}
	for _, TaskCheckpoint := range data.TasksCheckpoints {
		chunksString := TaskCheckpoint.Chunks
		chunkHashes := strings.Split(chunksString, ",")
		for _, chunkHash := range chunkHashes {
			if !utils.Contains(chunks, chunkHash) {
				chunks = append(chunks, chunkHash)
			}
		}
	}
	for _, Template := range data.Templates {
		chunksString := Template.Chunks
		chunkHashes := strings.Split(chunksString, ",")
		for _, chunkHash := range chunkHashes {
			if !utils.Contains(chunks, chunkHash) {
				chunks = append(chunks, chunkHash)
			}
		}
	}
	if strict {
		missingChunks, err := chunk_service.GetNonExistingChunks(tx, chunks)
		if err != nil {
			return err
		}
		if len(missingChunks) != 0 {
			return errors.New("data have missing chunks")
		}
	}

	previewIds := []string{}
	if data.ProjectPreview != "" && !utils.Contains(previewIds, data.ProjectPreview) {
		previewIds = append(previewIds, data.ProjectPreview)
	}

	for _, task := range data.Tasks {
		if task.PreviewId != "" && !utils.Contains(previewIds, task.PreviewId) {
			previewIds = append(previewIds, task.PreviewId)
		}
	}
	for _, entity := range data.Entities {
		if entity.PreviewId != "" && !utils.Contains(previewIds, entity.PreviewId) {
			previewIds = append(previewIds, entity.PreviewId)
		}
	}
	for _, taskCheckpoint := range data.TasksCheckpoints {
		if taskCheckpoint.PreviewId != "" && !utils.Contains(previewIds, taskCheckpoint.PreviewId) {
			previewIds = append(previewIds, taskCheckpoint.PreviewId)
		}
	}

	missingPreviews, err := repository.GetNonExistingPreviews(tx, previewIds)
	if err != nil {
		return err
	}
	if len(missingPreviews) != 0 {
		return errors.New("data have missing previews")
	}

	if data.ProjectPreview != "" {
		_, err = tx.Exec(`
		INSERT INTO config (name, value, mtime, synced)
		VALUES ('project_preview', $1, $2, 1)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime, synced = 1
	`, data.ProjectPreview, utils.GetEpochTime())
	}

	for _, role := range data.Roles {
		if tombItems[role.Id] {
			continue
		}
		roleAttributes := models.RoleAttributes{
			ViewEntity:   role.ViewEntity,
			CreateEntity: role.CreateEntity,
			UpdateEntity: role.UpdateEntity,
			DeleteEntity: role.DeleteEntity,

			ViewTask:   role.ViewTask,
			CreateTask: role.CreateTask,
			UpdateTask: role.UpdateTask,
			DeleteTask: role.DeleteTask,

			ViewTemplate:   role.ViewTemplate,
			CreateTemplate: role.CreateTemplate,
			UpdateTemplate: role.UpdateTemplate,
			DeleteTemplate: role.DeleteTemplate,

			ViewCheckpoint:   role.ViewCheckpoint,
			CreateCheckpoint: role.CreateCheckpoint,
			DeleteCheckpoint: role.DeleteCheckpoint,

			PullChunk: role.PullChunk,

			AssignTask:   role.AssignTask,
			UnassignTask: role.UnassignTask,

			AddUser:    role.AddUser,
			RemoveUser: role.RemoveUser,
			ChangeRole: role.ChangeRole,

			ChangeStatus:  role.ChangeStatus,
			SetDoneTask:   role.SetDoneTask,
			SetRetakeTask: role.SetRetakeTask,

			ViewDoneTask: role.ViewDoneTask,

			ManageDependencies: role.ManageDependencies,
		}
		localRole, err := repository.GetRole(tx, role.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrRoleNotFound) {
				_, err := repository.CreateRole(tx, role.Id, role.Name, roleAttributes)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if localRole.MTime < role.MTime {
				_, err = repository.UpdateRole(tx, role.Id, role.Name, roleAttributes)
				if err != nil {
					return err
				}
			}
		}

	}
	for _, user := range data.Users {
		// if tombItems[user.Id] {
		// 	continue
		// }

		localUser, err := repository.GetUser(tx, user.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrUserNotFound) {
				_, err := repository.AddKnownUser(
					tx, user.Id, user.Email, user.Username,
					user.FirstName, user.LastName, user.RoleId, user.Photo, false)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if localUser.MTime < user.MTime {
				if localUser.RoleId != user.RoleId {
					err = repository.ChangeUserRole(tx, user.Id, user.RoleId)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	for _, entityType := range data.EntityTypes {
		if tombItems[entityType.Id] {
			continue
		}
		localEntityType, err := repository.GetEntityType(tx, entityType.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrEntityTypeNotFound) {
				_, err = repository.CreateEntityType(
					tx, entityType.Id, entityType.Name, entityType.Icon)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if localEntityType.MTime < entityType.MTime {
				_, err = repository.UpdateEntityType(tx, entityType.Id, entityType.Name, entityType.Icon)
				if err != nil {
					return err
				}
			}
		}

	}
	for _, taskType := range data.TaskTypes {
		if tombItems[taskType.Id] {
			continue
		}
		localTaskType, err := repository.GetTaskType(tx, taskType.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrTaskTypeNotFound) {
				_, err = repository.CreateTaskType(
					tx, taskType.Id, taskType.Name, taskType.Icon)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {

			if localTaskType.MTime < taskType.MTime {
				_, err = repository.UpdateTaskType(tx, taskType.Id, taskType.Name, taskType.Icon)
				if err != nil {
					return err
				}
			}
		}

	}
	for _, dependencyType := range data.DependencyTypes {
		if tombItems[dependencyType.Id] {
			continue
		}
		_, err = repository.GetDependencyType(tx, dependencyType.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrDependencyTypeNotFound) {
				_, err = repository.CreateDependencyType(
					tx, dependencyType.Id, dependencyType.Name)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	for _, status := range data.Statuses {
		if tombItems[status.Id] {
			continue
		}
		_, err = repository.GetStatus(tx, status.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrStatusNotFound) {
				_, err = repository.CreateStatus(
					tx, status.Id, status.Name, status.ShortName, status.Color)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	for _, tag := range data.Tags {
		if tombItems[tag.Id] {
			continue
		}
		_, err = repository.GetTag(tx, tag.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrTagNotFound) {
				_, err = repository.CreateTag(tx, tag.Id, tag.Name)
				if err != nil {
					return err
				}
			} else {

				return err
			}
		}
	}

	start := time.Now()
	localEntities, err := repository.GetSimpleEntities(tx)
	if err != nil {
		return err
	}
	localEntitiesIndex := make(map[string]int)
	for i, t := range localEntities {
		localEntitiesIndex[t.Id] = i
	}

	for _, entity := range data.Entities {
		if tombItems[entity.Id] {
			continue
		}
		i, exists := localEntitiesIndex[entity.Id]
		if !exists {
			fmt.Println("Creating: ", entity.Name)
			err = repository.AddEntity(
				tx, entity.Id, entity.Name, entity.Description, entity.EntityTypeId, entity.ParentId, entity.PreviewId, entity.IsLibrary)
			if err != nil {
				if err.Error() == "parent entity not found" {
					continue
				}
				return err
			}
			continue
		}

		localEntity := localEntities[i]
		if localEntity.MTime < entity.MTime {

			parentId := entity.ParentId
			previewId := entity.PreviewId
			isLibrary := entity.IsLibrary

			entity, err = repository.RenameEntity(tx, entity.Id, entity.Name)
			if err != nil {
				return err
			}

			entity.ParentId = parentId
			entity.PreviewId = previewId
			entity.IsLibrary = isLibrary

			if localEntity.ParentId != entity.ParentId {
				err = repository.ChangeParent(tx, entity.Id, entity.ParentId)
				if err != nil {
					return err
				}
			}

			if localEntity.PreviewId != entity.PreviewId {
				err = repository.SetEntityPreview(tx, entity.Id, "entity", entity.PreviewId)
				if err != nil {
					return err
				}
			}
			if localEntity.IsLibrary != entity.IsLibrary {
				err = repository.ChangeIsLibrary(tx, entity.Id, entity.IsLibrary)
				if err != nil {
					return err
				}
			}

		}
	}
	elapsed := time.Since(start)
	fmt.Printf("entity write took %s\n", elapsed)

	for _, entityAssignee := range data.EntityAssignees {
		if tombItems[entityAssignee.Id] {
			continue
		}
		_, err = repository.GetAssignee(tx, entityAssignee.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrEntityAssigneeNotFound) {
				err = repository.AddAssignee(
					tx, entityAssignee.Id, entityAssignee.EntityId, entityAssignee.AssigneeId)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	start = time.Now()
	localTasks, err := repository.GetSimpleTasks(tx)
	if err != nil {
		return err
	}
	localTasksIndex := make(map[string]int)
	for i, t := range localTasks {
		localTasksIndex[t.Id] = i
	}

	createTaskQuery := `
		INSERT INTO task 
		(id, assignee_id, mtime, created_at, name, description, extension, task_type_id, entity_id, is_resource, status_id, pointer, is_link, preview_id) 
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?);
	`
	createTaskStmt, err := tx.Prepare(createTaskQuery)
	if err != nil {
		return err
	}

	for _, task := range data.Tasks {
		if tombItems[task.Id] {
			continue
		}

		i, exists := localTasksIndex[task.Id]
		if !exists {
			_, err := createTaskStmt.Exec(task.Id, task.AssigneeId, task.MTime, task.CreatedAt, task.Name, task.Description, task.Extension, task.TaskTypeId, task.EntityId, task.IsResource, task.StatusId, task.Pointer, task.IsLink, task.PreviewId)
			if err != nil {
				return err
			}
			continue
		}

		localTask := localTasks[i]
		if localTask.MTime < task.MTime {
			err := repository.UpdateSyncTask(tx, task.Id, task.Name, task.EntityId, task.TaskTypeId, task.AssigneeId, task.AssignerId, task.StatusId, task.PreviewId, task.IsResource, task.IsLink, task.Pointer, []string{})
			if err != nil {
				return err
			}
		}
	}
	elapsed = time.Since(start)
	fmt.Printf("task write took %s\n", elapsed)

	start = time.Now()
	localTasksCheckpoints, err := repository.GetSimpleCheckpoints(tx)
	if err != nil {
		return err
	}
	localTasksCheckpointsIndex := make(map[string]int)
	for i, c := range localTasksCheckpoints {
		localTasksCheckpointsIndex[c.Id] = i
	}

	createCheckpointQuery := `
		INSERT INTO task_checkpoint 
		(id, mtime, created_at, task_id, xxhash_checksum, time_modified, file_size, comment, chunks, author_id, preview_id, group_id) 
		VALUES (?, ?,?,?,?,?,?,?,?,?,?,?);
	`
	createCheckpointStmt, err := tx.Prepare(createCheckpointQuery)
	if err != nil {
		return err
	}

	for _, taskCheckpoint := range data.TasksCheckpoints {
		if tombItems[taskCheckpoint.Id] {
			continue
		}

		_, exists := localTasksCheckpointsIndex[taskCheckpoint.Id]
		if !exists {
			EpochTime, err := utils.RFC3339ToEpoch(taskCheckpoint.CreatedAt)
			if err != nil {
				return err
			}

			_, err = createCheckpointStmt.Exec(taskCheckpoint.Id, taskCheckpoint.MTime, EpochTime, taskCheckpoint.TaskId, taskCheckpoint.XXHashChecksum, taskCheckpoint.TimeModified, taskCheckpoint.FileSize, taskCheckpoint.Comment, taskCheckpoint.Chunks, taskCheckpoint.AuthorUID, taskCheckpoint.PreviewId, taskCheckpoint.GroupId)
			if err != nil {
				return err
			}
			continue
		}
	}
	elapsed = time.Since(start)
	fmt.Printf("checkpoint write took %s\n", elapsed)

	for _, dependency := range data.TaskDependencies {
		if tombItems[dependency.Id] {
			continue
		}
		_, err = repository.GetDependency(tx, dependency.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrTaskDependencyNotFound) {
				_, err = repository.AddDependency(
					tx, dependency.Id, dependency.TaskId, dependency.DependencyId, dependency.DependencyTypeId)
				if err != nil {
					if err.Error() == "UNIQUE constraint failed: task_dependency.task_id, task_dependency.dependency_id" {
						continue
					}
					return err
				}
			} else {
				return err
			}
		}
	}

	for _, dependency := range data.EntityDependencies {
		if tombItems[dependency.Id] {
			continue
		}
		_, err = repository.GetEntityDependency(tx, dependency.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrEntityDependencyNotFound) {
				_, err = repository.AddEntityDependency(
					tx, dependency.Id, dependency.TaskId, dependency.DependencyId, dependency.DependencyTypeId)
				if err != nil {
					if err.Error() == "UNIQUE constraint failed: entity_dependency.task_id, entity_dependency.dependency_id" {
						continue
					}
					return err
				}
			} else {
				return err
			}
		}
	}

	for _, template := range data.Templates {
		if tombItems[template.Id] {
			continue
		}
		_, err = repository.GetTemplate(tx, template.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrTemplateNotFound) {
				_, err = repository.AddTemplate(
					tx, template.Id, template.Name, template.Extension, template.Chunks, template.XxhashChecksum, template.FileSize)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	for _, workflow := range data.Workflows {
		if tombItems[workflow.Id] {
			continue
		}
		localWorkflow, err := repository.GetWorkflow(tx, workflow.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrWorkflowNotFound) {
				_, err = repository.CreateWorkflow(
					tx, workflow.Id, workflow.Name, []models.WorkflowTask{}, []models.WorkflowEntity{}, []models.WorkflowLink{})
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if localWorkflow.MTime < workflow.MTime {
				err = repository.RenameWorkflow(tx, workflow.Id, workflow.Name)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, workflowLink := range data.WorkflowLinks {
		if tombItems[workflowLink.Id] {
			continue
		}
		localWorkflowLink, err := repository.GetWorkflowLink(tx, workflowLink.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrWorkflowLinkNotFound) {
				err = repository.AddLinkWorkflow(tx, workflowLink.Id, workflowLink.Name, workflowLink.EntityTypeId, workflowLink.WorkflowId, workflowLink.LinkedWorkflowId)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if localWorkflowLink.MTime < workflowLink.MTime {
				err = repository.RenameLinkedWorkflow(tx, workflowLink.Id, workflowLink.Name)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, workflowEntity := range data.WorkflowEntities {
		if tombItems[workflowEntity.Id] {
			continue
		}
		localWorkflowEntity, err := repository.GetWorkflowEntity(tx, workflowEntity.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrWorkflowEntityNotFound) {
				_, err = repository.CreateWorkflowEntity(tx, workflowEntity.Id, workflowEntity.Name, workflowEntity.WorkflowId, workflowEntity.EntityTypeId)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if localWorkflowEntity.MTime < workflowEntity.MTime {
				_, err = repository.UpdateWorkflowEntity(
					tx, workflowEntity.Id, workflowEntity.Name, workflowEntity.EntityTypeId)
				if err != nil {
					return err
				}
			}
		}

	}

	for _, workflowTask := range data.WorkflowTasks {
		if tombItems[workflowTask.Id] {
			continue
		}
		localWorkflowTask, err := repository.GetWorkflowTask(tx, workflowTask.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrWorkflowTaskNotFound) {
				_, err = repository.CreateWorkflowTask(tx, workflowTask.Id, workflowTask.Name, workflowTask.WorkflowId, workflowTask.TaskTypeId, workflowTask.IsResource, workflowTask.TemplateId, workflowTask.Pointer, workflowTask.IsLink)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if localWorkflowTask.MTime < workflowTask.MTime {
				_, err = repository.UpdateWorkflowTask(tx, workflowTask.Id, workflowTask.Name, workflowTask.TaskTypeId, workflowTask.IsResource, workflowTask.TemplateId, workflowTask.Pointer, workflowTask.IsLink)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, taskTag := range data.TasksTags {
		if tombItems[taskTag.Id] {
			continue
		}
		_, err = repository.GetTaskTag(tx, taskTag.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrTaskTagNotFound) {
				err = repository.AddTagToTaskById(tx, taskTag.Id, taskTag.TaskId, taskTag.TagId)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}
	return nil
}

func OverWriteProjectData(tx *sqlx.Tx, data ProjectData) error {
	// Sort
	start := time.Now()
	sortedEntities, err := repository.TopologicalSort(data.Entities)
	if err != nil {
		return err
	}
	data.Entities = sortedEntities
	elapsed := time.Since(start)
	fmt.Printf("sort data took %s\n", elapsed)
	// sort.Slice(data.Entities, func(i, j int) bool {
	// 	iDepth := strings.Count(data.Entities[i].EntityPath, "/")
	// 	jDepth := strings.Count(data.Entities[j].EntityPath, "/")
	// 	if iDepth != jDepth {
	// 		return iDepth < jDepth
	// 	}
	// 	return data.Entities[i].EntityPath < data.Entities[j].EntityPath
	// })
	// sort.Slice(data.Tasks, func(i, j int) bool {
	// 	iDepth := strings.Count(data.Tasks[i].TaskPath, "/")
	// 	jDepth := strings.Count(data.Tasks[j].TaskPath, "/")
	// 	if iDepth != jDepth {
	// 		return iDepth < jDepth
	// 	}
	// 	return data.Tasks[i].TaskPath < data.Tasks[j].TaskPath
	// })

	previewIds := []string{}
	if data.ProjectPreview != "" && !utils.Contains(previewIds, data.ProjectPreview) {
		previewIds = append(previewIds, data.ProjectPreview)
	}

	for _, task := range data.Tasks {
		if task.PreviewId != "" && !utils.Contains(previewIds, task.PreviewId) {
			previewIds = append(previewIds, task.PreviewId)
		}
	}
	for _, entity := range data.Entities {
		if entity.PreviewId != "" && !utils.Contains(previewIds, entity.PreviewId) {
			previewIds = append(previewIds, entity.PreviewId)
		}
	}
	for _, taskCheckpoint := range data.TasksCheckpoints {
		if taskCheckpoint.PreviewId != "" && !utils.Contains(previewIds, taskCheckpoint.PreviewId) {
			previewIds = append(previewIds, taskCheckpoint.PreviewId)
		}
	}

	missingPreviews, err := repository.GetNonExistingPreviews(tx, previewIds)
	if err != nil {
		return err
	}
	if len(missingPreviews) != 0 {
		return errors.New("data have missing previews")
	}

	if data.ProjectPreview != "" {
		_, err = tx.Exec(`
		INSERT INTO config (name, value, mtime, synced)
		VALUES ('project_preview', $1, $2, 1)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime, synced = 1
	`, data.ProjectPreview, utils.GetEpochTime())
		if err != nil {
			return err
		}
	}

	for _, role := range data.Roles {
		roleAttributes := models.RoleAttributes{
			ViewEntity:   role.ViewEntity,
			CreateEntity: role.CreateEntity,
			UpdateEntity: role.UpdateEntity,
			DeleteEntity: role.DeleteEntity,

			ViewTask:   role.ViewTask,
			CreateTask: role.CreateTask,
			UpdateTask: role.UpdateTask,
			DeleteTask: role.DeleteTask,

			ViewTemplate:   role.ViewTemplate,
			CreateTemplate: role.CreateTemplate,
			UpdateTemplate: role.UpdateTemplate,
			DeleteTemplate: role.DeleteTemplate,

			ViewCheckpoint:   role.ViewCheckpoint,
			CreateCheckpoint: role.CreateCheckpoint,
			DeleteCheckpoint: role.DeleteCheckpoint,

			PullChunk: role.PullChunk,

			AssignTask:   role.AssignTask,
			UnassignTask: role.UnassignTask,

			AddUser:    role.AddUser,
			RemoveUser: role.RemoveUser,
			ChangeRole: role.ChangeRole,

			ChangeStatus:  role.ChangeStatus,
			SetDoneTask:   role.SetDoneTask,
			SetRetakeTask: role.SetRetakeTask,

			ViewDoneTask: role.ViewDoneTask,

			ManageDependencies: role.ManageDependencies,
		}
		_, err := repository.CreateRole(tx, role.Id, role.Name, roleAttributes)
		if err != nil {
			return err
		}
	}

	for _, user := range data.Users {
		_, err := repository.AddKnownUser(
			tx, user.Id, user.Email, user.Username,
			user.FirstName, user.LastName, user.RoleId, user.Photo, false)
		if err != nil {
			return err
		}
	}

	for _, entityType := range data.EntityTypes {
		_, err = repository.CreateEntityType(
			tx, entityType.Id, entityType.Name, entityType.Icon)
		if err != nil {
			return err
		}
	}

	for _, taskType := range data.TaskTypes {
		_, err = repository.CreateTaskType(
			tx, taskType.Id, taskType.Name, taskType.Icon)
		if err != nil {
			return err
		}
	}

	for _, dependencyType := range data.DependencyTypes {
		_, err = repository.CreateDependencyType(
			tx, dependencyType.Id, dependencyType.Name)
		if err != nil {
			return err
		}
	}

	for _, status := range data.Statuses {
		_, err = repository.CreateStatus(
			tx, status.Id, status.Name, status.ShortName, status.Color)
		if err != nil {
			return err
		}
	}

	for _, tag := range data.Tags {
		_, err = repository.CreateTag(tx, tag.Id, tag.Name)
		if err != nil {
			return err
		}
	}

	start = time.Now()
	for _, entity := range data.Entities {
		err = repository.AddEntity(
			tx, entity.Id, entity.Name, entity.Description, entity.EntityTypeId, entity.ParentId, entity.PreviewId, entity.IsLibrary)
		if err != nil {
			if err.Error() == "parent entity not found" {
				continue
			}
			return err
		}
	}
	elapsed = time.Since(start)
	fmt.Printf("entity write took %s\n", elapsed)

	for _, entityAssignee := range data.EntityAssignees {
		err = repository.AddAssignee(
			tx, entityAssignee.Id, entityAssignee.EntityId, entityAssignee.AssigneeId)
		if err != nil {
			return err
		}
	}

	start = time.Now()
	localTasks, err := repository.GetSimpleTasks(tx)
	if err != nil {
		return err
	}
	localTasksIndex := make(map[string]int)
	for i, t := range localTasks {
		localTasksIndex[t.Id] = i
	}

	createTaskQuery := `
		INSERT INTO task 
		(id, assignee_id, mtime, created_at, name, description, extension, task_type_id, entity_id, is_resource, status_id, pointer, is_link, preview_id) 
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?);
	`
	createTaskStmt, err := tx.Prepare(createTaskQuery)
	if err != nil {
		return err
	}

	for _, task := range data.Tasks {
		_, err := createTaskStmt.Exec(task.Id, task.AssigneeId, task.MTime, task.CreatedAt, task.Name, task.Description, task.Extension, task.TaskTypeId, task.EntityId, task.IsResource, task.StatusId, task.Pointer, task.IsLink, task.PreviewId)
		if err != nil {
			return err
		}
	}
	elapsed = time.Since(start)
	fmt.Printf("task write took %s\n", elapsed)

	start = time.Now()
	createCheckpointQuery := `
		INSERT INTO task_checkpoint 
		(id, mtime, created_at, task_id, xxhash_checksum, time_modified, file_size, comment, chunks, author_id, preview_id, group_id) 
		VALUES (?, ?,?,?,?,?,?,?,?,?,?,?);
	`
	createCheckpointStmt, err := tx.Prepare(createCheckpointQuery)
	if err != nil {
		return err
	}

	for _, taskCheckpoint := range data.TasksCheckpoints {
		EpochTime, err := utils.RFC3339ToEpoch(taskCheckpoint.CreatedAt)
		if err != nil {
			return err
		}
		_, err = createCheckpointStmt.Exec(taskCheckpoint.Id, taskCheckpoint.MTime, EpochTime, taskCheckpoint.TaskId, taskCheckpoint.XXHashChecksum, taskCheckpoint.TimeModified, taskCheckpoint.FileSize, taskCheckpoint.Comment, taskCheckpoint.Chunks, taskCheckpoint.AuthorUID, taskCheckpoint.PreviewId, taskCheckpoint.GroupId)
		if err != nil {
			return err
		}
	}
	elapsed = time.Since(start)
	fmt.Printf("checkpoint write took %s\n", elapsed)

	for _, dependency := range data.TaskDependencies {
		_, err = repository.AddDependency(
			tx, dependency.Id, dependency.TaskId, dependency.DependencyId, dependency.DependencyTypeId)
		if err != nil {
			if err.Error() == "UNIQUE constraint failed: task_dependency.task_id, task_dependency.dependency_id" {
				continue
			}
			return err
		}
	}

	for _, dependency := range data.EntityDependencies {
		_, err = repository.AddEntityDependency(
			tx, dependency.Id, dependency.TaskId, dependency.DependencyId, dependency.DependencyTypeId)
		if err != nil {
			if err.Error() == "UNIQUE constraint failed: entity_dependency.task_id, entity_dependency.dependency_id" {
				continue
			}
			return err
		}
	}

	for _, template := range data.Templates {
		_, err = repository.AddTemplate(
			tx, template.Id, template.Name, template.Extension, template.Chunks, template.XxhashChecksum, template.FileSize)
		if err != nil {
			return err
		}
	}

	for _, workflow := range data.Workflows {
		_, err = repository.CreateWorkflow(
			tx, workflow.Id, workflow.Name, []models.WorkflowTask{}, []models.WorkflowEntity{}, []models.WorkflowLink{})
		if err != nil {
			return err
		}
	}

	for _, workflowLink := range data.WorkflowLinks {
		err = repository.AddLinkWorkflow(tx, workflowLink.Id, workflowLink.Name, workflowLink.EntityTypeId, workflowLink.WorkflowId, workflowLink.LinkedWorkflowId)
		if err != nil {
			return err
		}
	}

	for _, workflowEntity := range data.WorkflowEntities {
		_, err = repository.CreateWorkflowEntity(tx, workflowEntity.Id, workflowEntity.Name, workflowEntity.WorkflowId, workflowEntity.EntityTypeId)
		if err != nil {
			return err
		}
	}

	for _, workflowTask := range data.WorkflowTasks {
		_, err = repository.CreateWorkflowTask(tx, workflowTask.Id, workflowTask.Name, workflowTask.WorkflowId, workflowTask.TaskTypeId, workflowTask.IsResource, workflowTask.TemplateId, workflowTask.Pointer, workflowTask.IsLink)
		if err != nil {
			return err
		}
	}

	for _, taskTag := range data.TasksTags {
		err = repository.AddTagToTaskById(tx, taskTag.Id, taskTag.TaskId, taskTag.TagId)
		if err != nil {
			return err
		}
	}
	return nil
}

func FetchData(remoteUrl string, userId string) (ProjectData, error) {
	userData := ProjectData{}
	userDataPb := repositorypb.ProjectData{}
	if utils.IsValidURL(remoteUrl) {
		type userTokenStruct struct {
			UserId string `json:"user_id"`
		}
		dataUrl := remoteUrl + "/data"

		userToken := userTokenStruct{
			UserId: userId,
		}
		jsonData, err := json.Marshal(userToken)
		if err != nil {
			return userData, err
		}

		req, err := http.NewRequest("GET", dataUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			return userData, err
		}
		req.Header.Set("Clustta-Agent", constants.USER_AGENT)

		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			return userData, err
		}
		defer response.Body.Close()

		responseCode := response.StatusCode
		if responseCode == 200 {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return userData, fmt.Errorf("error reading response body: %s", err.Error())
			}

			decompressedData, err := zstd.Decompress(nil, body)
			if err != nil {
				return userData, err
			}

			err = proto.Unmarshal(decompressedData, &userDataPb)
			if err != nil {
				return userData, err
			}

			userData = ProjectData{
				ProjectPreview:  userDataPb.ProjectPreview,
				EntityTypes:     repository.FromPbEntityTypes(userDataPb.EntityTypes),
				Entities:        repository.FromPbEntities(userDataPb.Entities),
				EntityAssignees: repository.FromPbEntityAssignees(userDataPb.EntityAssignees),

				TaskTypes:          repository.FromPbTaskTypes(userDataPb.TaskTypes),
				Tasks:              repository.FromPbTasks(userDataPb.Tasks),
				TasksCheckpoints:   repository.FromPbCheckpoints(userDataPb.TasksCheckpoints),
				TaskDependencies:   repository.FromPbTaskDependencies(userDataPb.TaskDependencies),
				EntityDependencies: repository.FromPbEntityDependencies(userDataPb.EntityDependencies),

				Statuses:        repository.FromPbStatuses(userDataPb.Statuses),
				DependencyTypes: repository.FromPbDependencyTypes(userDataPb.DependencyTypes),

				Users: repository.FromPbUsers(userDataPb.Users),
				Roles: repository.FromPbRoles(userDataPb.Roles),

				Templates: repository.FromPbTemplates(userDataPb.Templates),

				Workflows:        repository.FromPbWorkflows(userDataPb.Workflows),
				WorkflowLinks:    repository.FromPbWorkflowLinks(userDataPb.WorkflowLinks),
				WorkflowEntities: repository.FromPbWorkflowEntities(userDataPb.WorkflowEntities),
				WorkflowTasks:    repository.FromPbWorkflowTasks(userDataPb.WorkflowTasks),

				Tags:      repository.FromPbTags(userDataPb.Tags),
				TasksTags: repository.FromPbTaskTags(userDataPb.TasksTags),
			}

			return userData, nil
		} else if responseCode == 400 {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return userData, err
			}
			return userData, errors.New(string(body))
		}
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return userData, err
		}
		return userData, fmt.Errorf("unknown error while fetching data. url: %s, status code: %d, message: %s", dataUrl, responseCode, string(body))
	} else if utils.FileExists(remoteUrl) {
		db, err := utils.OpenDb(remoteUrl)
		if err != nil {
			return userData, err
		}
		defer db.Close()
		remoteTx, err := db.Beginx()
		if err != nil {
			return userData, err
		}
		defer remoteTx.Rollback()
		userData, err = LoadUserData(remoteTx, userId)
		if err != nil {
			return userData, err
		}
	} else {
		return userData, fmt.Errorf("invalid url:%s", remoteUrl)
	}
	return userData, nil

}

func CalculateMissingPreviews(tx *sqlx.Tx, data ProjectData) ([]string, error) {
	previewIds := []string{}
	if data.ProjectPreview != "" && !utils.Contains(previewIds, data.ProjectPreview) {
		previewIds = append(previewIds, data.ProjectPreview)
	}

	for _, task := range data.Tasks {
		if task.PreviewId != "" && !utils.Contains(previewIds, task.PreviewId) {
			previewIds = append(previewIds, task.PreviewId)
		}
	}
	for _, entity := range data.Entities {
		if entity.PreviewId != "" && !utils.Contains(previewIds, entity.PreviewId) {
			previewIds = append(previewIds, entity.PreviewId)
		}
	}
	for _, taskCheckpoint := range data.TasksCheckpoints {
		if taskCheckpoint.PreviewId != "" && !utils.Contains(previewIds, taskCheckpoint.PreviewId) {
			previewIds = append(previewIds, taskCheckpoint.PreviewId)
		}
	}

	missingPreviews, err := repository.GetNonExistingPreviews(tx, previewIds)
	return missingPreviews, err
}

func CalculateMissingChunks(tx *sqlx.Tx, data ProjectData, userId string, syncOptions SyncOptions) ([]string, []string, int, error) {
	tasksIds := []string{}

	for _, task := range data.Tasks {
		if task.AssigneeId == userId {
			tasksIds = append(tasksIds, task.Id)
		} else if syncOptions.TaskDependencies && task.IsDependency {
			tasksIds = append(tasksIds, task.Id)
		} else if syncOptions.Tasks {
			tasksIds = append(tasksIds, task.Id)
		}
	}

	// Maps to keep track of the latest checkpoint for each entity
	latestTaskCheckpoints := make(map[string]models.Checkpoint)
	// Iterate over task checkpoints to find the latest for each entity
	for _, taskCheckpoint := range data.TasksCheckpoints {
		if utils.Contains(tasksIds, taskCheckpoint.TaskId) {
			existingCheckpoint, found := latestTaskCheckpoints[taskCheckpoint.TaskId]
			if !found || taskCheckpoint.CreatedAt > existingCheckpoint.CreatedAt {
				latestTaskCheckpoints[taskCheckpoint.TaskId] = taskCheckpoint
			}
		}
	}

	// Now gather all the chunks from the latest checkpoints
	// chunks := []string{}
	seenChunks := make(map[string]bool)
	missingChunks := []string{}
	allChunks := []string{}
	totalSize := 0
	for _, checkpoint := range latestTaskCheckpoints {
		chunkHashes := strings.Split(checkpoint.Chunks, ",")
		checkpointFullyDownloaded := true
		for _, chunkHash := range chunkHashes {
			if chunk_service.ChunkExists(chunkHash, tx, seenChunks) {
				continue
			}
			if !utils.Contains(missingChunks, chunkHash) {
				checkpointFullyDownloaded = false
				missingChunks = append(missingChunks, chunkHash)
			}
		}
		if !checkpointFullyDownloaded {
			totalSize += checkpoint.FileSize
			allChunks = append(allChunks, chunkHashes...)
		}
	}

	for _, template := range data.Templates {
		chunkHashes := strings.Split(template.Chunks, ",")
		templateFullyDownloaded := true
		for _, chunkHash := range chunkHashes {
			if chunk_service.ChunkExists(chunkHash, tx, seenChunks) {
				continue
			}
			if !utils.Contains(missingChunks, chunkHash) {
				templateFullyDownloaded = false
				missingChunks = append(missingChunks, chunkHash)
			}
		}
		if !templateFullyDownloaded {
			totalSize += template.FileSize
			allChunks = append(allChunks, chunkHashes...)
		}
	}

	return missingChunks, allChunks, totalSize, nil
}
func CalculateCheckpointsMissingChunks(tx *sqlx.Tx, checkpoints []models.Checkpoint) ([]string, []string, int, error) {
	// Now gather all the chunks from the latest checkpoints
	// chunks := []string{}
	seenChunks := make(map[string]bool)
	missingChunks := []string{}
	allChunks := []string{}
	totalSize := 0
	for _, checkpoint := range checkpoints {
		chunkHashes := strings.Split(checkpoint.Chunks, ",")
		checkpointFullyDownloaded := true
		for _, chunkHash := range chunkHashes {
			if chunk_service.ChunkExists(chunkHash, tx, seenChunks) {
				continue
			}
			if !utils.Contains(missingChunks, chunkHash) {
				checkpointFullyDownloaded = false
				missingChunks = append(missingChunks, chunkHash)
			}
		}
		if !checkpointFullyDownloaded {
			totalSize += checkpoint.FileSize
			allChunks = append(allChunks, chunkHashes...)
		}
	}

	return missingChunks, allChunks, totalSize, nil
}

func DownloadCheckpoint(ctx context.Context, projectPath, remoteUrl string, checkpointId string, userId string, callback func(int, int, string, string)) error {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	checkpoint, err := repository.GetCheckpoint(tx, checkpointId)
	if err != nil {
		return err
	}
	missingChunks, allChunks, totalSize, err := CalculateCheckpointsMissingChunks(tx, []models.Checkpoint{checkpoint})
	if err != nil {
		return err
	}
	tx.Rollback()

	if len(missingChunks) > 0 {
		err = chunk_service.PullStreamChunks(ctx, projectPath, remoteUrl, missingChunks, allChunks, totalSize, callback)
		if err != nil {
			return err
		}
	}
	return nil
}

func DownloadCheckpoints(ctx context.Context, projectPath, remoteUrl string, checkpointIds []string, userId string, callback func(int, int, string, string)) error {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	quotedcheckpointIds := make([]string, len(checkpointIds))
	for i, id := range checkpointIds {
		quotedcheckpointIds[i] = fmt.Sprintf("\"%s\"", id)
	}

	checkpoints := []models.Checkpoint{}
	err = tx.Select(&checkpoints, fmt.Sprintf("SELECT * FROM task_checkpoint WHERE id IN (%s)", strings.Join(quotedcheckpointIds, ",")))
	if err != nil {
		return err
	}
	missingChunks, allChunks, totalSize, err := CalculateCheckpointsMissingChunks(tx, checkpoints)
	if err != nil {
		return err
	}
	tx.Rollback()

	if len(missingChunks) > 0 {
		err = chunk_service.PullStreamChunks(ctx, projectPath, remoteUrl, missingChunks, allChunks, totalSize, callback)
		if err != nil {
			return err
		}
	}
	return nil
}

func SyncData(workingData, remoteUrl string, user auth_service.User) error {
	dbConn, err := utils.OpenDb(workingData)
	if err != nil {
		return err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	data, err := FetchData(remoteUrl, user.Id)
	if err != nil {
		return err
	}
	println("recieve")
	err = ClearLocalData(tx)
	if err != nil {
		return err
	}
	err = WriteProjectData(tx, data, false)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func IsUnsynced(tx *sqlx.Tx) (bool, error) {
	data, err := LoadChangedData(tx)
	if err != nil {
		return false, err
	}
	if data.IsEmpty() {
		return false, nil
	}
	return true, nil
}

func FetchChunksInfo(remoteUrl string, userId string, chunks []string) ([]chunk_service.ChunkInfo, error) {
	if utils.IsValidURL(remoteUrl) {

		dataUrl := remoteUrl + "/chunks-info"

		jsonData, err := json.Marshal(chunks)
		if err != nil {
			return []chunk_service.ChunkInfo{}, err
		}

		req, err := http.NewRequest("GET", dataUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			return []chunk_service.ChunkInfo{}, err
		}
		req.Header.Set("Clustta-Agent", constants.USER_AGENT)

		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			return []chunk_service.ChunkInfo{}, err
		}
		defer response.Body.Close()

		responseCode := response.StatusCode
		responseData := []chunk_service.ChunkInfo{}
		if responseCode == 200 {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return []chunk_service.ChunkInfo{}, fmt.Errorf("error reading response body: %s", err.Error())
			}
			err = json.Unmarshal(body, &responseData)
			if err != nil {
				return []chunk_service.ChunkInfo{}, err
			}
			return responseData, nil
		} else {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return []chunk_service.ChunkInfo{}, err
			}
			println(string(body))
			return []chunk_service.ChunkInfo{}, errors.New(string(body))
		}
	} else if utils.FileExists(remoteUrl) {
		dbConn, err := utils.OpenDb(remoteUrl)
		if err != nil {
			return []chunk_service.ChunkInfo{}, err
		}
		defer dbConn.Close()
		remoteTx, err := dbConn.Beginx()
		if err != nil {
			return []chunk_service.ChunkInfo{}, err
		}
		defer remoteTx.Rollback()

		ChunksInfo, err := chunk_service.GetChunksInfo(remoteTx, chunks)
		if err != nil {
			return []chunk_service.ChunkInfo{}, err
		}
		return ChunksInfo, nil
	} else {
		return []chunk_service.ChunkInfo{}, fmt.Errorf("invalid url:%s", remoteUrl)
	}
}

func FetchMissingChunks(remoteUrl string, userId string, chunks []string) ([]string, error) {
	if utils.IsValidURL(remoteUrl) {

		dataUrl := remoteUrl + "/chunks-missing"

		jsonData, err := json.Marshal(chunks)
		if err != nil {
			return []string{}, err
		}

		req, err := http.NewRequest("GET", dataUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			return []string{}, err
		}
		req.Header.Set("Clustta-Agent", constants.USER_AGENT)

		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			return []string{}, err
		}
		defer response.Body.Close()

		responseCode := response.StatusCode
		responseData := []string{}
		if responseCode == 200 {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return []string{}, fmt.Errorf("error reading response body: %s", err.Error())
			}
			err = json.Unmarshal(body, &responseData)
			if err != nil {
				return []string{}, err
			}
			return responseData, nil
		} else {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return []string{}, err
			}
			return []string{}, errors.New(string(body))
		}
	} else if utils.FileExists(remoteUrl) {
		dbConn, err := utils.OpenDb(remoteUrl)
		if err != nil {
			return []string{}, err
		}
		defer dbConn.Close()
		remoteTx, err := dbConn.Beginx()
		if err != nil {
			return []string{}, err
		}
		defer remoteTx.Rollback()

		missingChunks := []string{}
		seenChunks := make(map[string]bool)
		for _, chunkHash := range chunks {
			if chunk_service.ChunkExists(chunkHash, remoteTx, seenChunks) {
				continue
			}
			missingChunks = append(missingChunks, chunkHash)
		}
		return missingChunks, nil
	} else {
		return []string{}, fmt.Errorf("invalid url:%s", remoteUrl)
	}
}

func FetchMissingPreviews(remoteUrl string, userId string, previews []string) ([]string, error) {
	if utils.IsValidURL(remoteUrl) {

		dataUrl := remoteUrl + "/previews-exist"

		jsonData, err := json.Marshal(previews)
		if err != nil {
			return []string{}, err
		}

		req, err := http.NewRequest("GET", dataUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			return []string{}, err
		}
		req.Header.Set("Clustta-Agent", constants.USER_AGENT)

		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			return []string{}, err
		}
		defer response.Body.Close()

		responseCode := response.StatusCode
		responseData := []string{}
		if responseCode == 200 {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return []string{}, fmt.Errorf("error reading response body: %s", err.Error())
			}
			err = json.Unmarshal(body, &responseData)
			if err != nil {
				return []string{}, err
			}
			return responseData, nil
		} else {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return []string{}, err
			}
			return []string{}, errors.New(string(body))
		}
	} else if utils.FileExists(remoteUrl) {
		dbConn, err := utils.OpenDb(remoteUrl)
		if err != nil {
			return []string{}, err
		}
		defer dbConn.Close()
		remoteTx, err := dbConn.Beginx()
		if err != nil {
			return []string{}, err
		}
		defer remoteTx.Rollback()

		missingPreviews := []string{}
		for _, previewHash := range previews {
			if repository.PreviewExists(previewHash, remoteTx) {
				continue
			}
			missingPreviews = append(missingPreviews, previewHash)
		}
		return missingPreviews, nil
	} else {
		return []string{}, fmt.Errorf("invalid url:%s", remoteUrl)
	}
}
