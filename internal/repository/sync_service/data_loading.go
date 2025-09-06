package sync_service

import (
	"clustta/internal/base_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/repositorypb"
	"clustta/internal/utils"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"
)

func OLDLoadUserData(tx *sqlx.Tx, userId string) (ProjectData, error) {
	userData := ProjectData{}

	user, err := repository.GetUser(tx, userId)
	if err != nil {
		return ProjectData{}, err
	}
	userRole, err := repository.GetRole(tx, user.RoleId)
	if err != nil {
		return ProjectData{}, err
	}

	// tasks, err := repository.GetUserTasks(tx, userId)
	// if err != nil {
	// 	return ProjectData{}, err
	// }
	tasks := []models.Task{}
	if userRole.ViewTask {
		tasks, err = repository.GetTasks(tx, true)
		if err != nil {
			return ProjectData{}, err
		}
	} else {
		tasks, err = repository.OLDGetUserTasks(tx, user.Id)
		if err != nil {
			return ProjectData{}, err
		}
	}

	var taskEntityIds []string
	var taskIds []string
	for _, task := range tasks {
		taskIds = append(taskIds, task.Id)
		if !utils.Contains(taskEntityIds, task.EntityId) {
			taskEntityIds = append(taskEntityIds, task.EntityId)
		}
	}
	quotedTaskIds := make([]string, len(taskIds))
	for i, id := range taskIds {
		quotedTaskIds[i] = fmt.Sprintf("\"%s\"", id)
	}
	quotedTaskEntityIds := make([]string, len(taskEntityIds))
	for i, id := range taskEntityIds {
		quotedTaskEntityIds[i] = fmt.Sprintf("\"%s\"", id)
	}

	// dependenciesQuery := fmt.Sprintf("SELECT * FROM task_dependencies WHERE task_id IN (%s) AND dependency_id NOT IN (%s)", strings.Join(quotedTaskIds, ","), strings.Join(quotedTaskIds, ","))
	taskDependenciesQuery := fmt.Sprintf("SELECT * FROM task_dependency WHERE task_id IN (%s)", strings.Join(quotedTaskIds, ","))
	taskDependencies := []models.TaskDependency{}
	err = tx.Select(&taskDependencies, taskDependenciesQuery)
	if err != nil {
		return ProjectData{}, err
	}

	entityDependenciesQuery := fmt.Sprintf("SELECT * FROM entity_dependency WHERE task_id IN (%s)", strings.Join(quotedTaskIds, ","))
	entityDependencies := []models.EntityDependency{}
	err = tx.Select(&entityDependencies, entityDependenciesQuery)
	if err != nil {
		return ProjectData{}, err
	}

	var uniqueDependencyIds []string
	for _, dependency := range taskDependencies {
		if !utils.Contains(uniqueDependencyIds, dependency.DependencyId) && !utils.Contains(taskIds, dependency.DependencyId) {
			uniqueDependencyIds = append(uniqueDependencyIds, dependency.DependencyId)
		}
	}
	quotedUniqueDependencyIds := make([]string, len(uniqueDependencyIds))
	for i, id := range uniqueDependencyIds {
		quotedUniqueDependencyIds[i] = fmt.Sprintf("\"%s\"", id)
	}

	uniqueDependenciesQuery := fmt.Sprintf("SELECT * FROM task WHERE trashed = 0 AND id IN (%s)", strings.Join(quotedUniqueDependencyIds, ","))
	uniqueDependencies := []models.Task{}
	err = tx.Select(&uniqueDependencies, uniqueDependenciesQuery)
	if err != nil {
		return ProjectData{}, err
	}

	tasks = append(tasks, uniqueDependencies...)
	taskIds = append(taskIds, uniqueDependencyIds...)

	quotedTaskIds = append(quotedTaskIds, quotedUniqueDependencyIds...)

	checkpointQuery := fmt.Sprintf("SELECT * FROM task_checkpoint WHERE trashed = 0 AND task_id IN (%s)", strings.Join(quotedTaskIds, ","))
	tasksCheckpoints := []models.Checkpoint{}
	err = tx.Select(&tasksCheckpoints, checkpointQuery)
	if err != nil {
		return ProjectData{}, err
	}

	statuses, err := repository.GetStatuses(tx)
	if err != nil {
		return ProjectData{}, err
	}
	taskTypes, err := repository.GetTaskTypes(tx)
	if err != nil {
		return ProjectData{}, err
	}

	users, err := repository.GetUsers(tx)
	if err != nil {
		return ProjectData{}, err
	}
	roles, err := repository.GetRoles(tx)
	if err != nil {
		return ProjectData{}, err
	}

	dependencyTypes, err := repository.GetDependencyTypes(tx)
	if err != nil {
		return ProjectData{}, err
	}

	entityTypes, err := repository.GetEntityTypes(tx)
	if err != nil {
		return ProjectData{}, err
	}
	entities := []models.Entity{}
	entityAssignees := []models.EntityAssignee{}
	if userRole.ViewTask {
		//TODO remove with trashed
		entities, err = repository.GetEntities(tx, true)
		if err != nil {
			return ProjectData{}, err
		}
		err = tx.Select(&entityAssignees, "SELECT * FROM entity_assignee")
		if err != nil {
			return ProjectData{}, err
		}
	} else {
		entities, err = repository.OLDGetUserEntities(tx, user.Id)
		if err != nil {
			return ProjectData{}, err
		}
		qoutedEntityIds := make([]string, len(entities))
		for i, entity := range entities {
			qoutedEntityIds[i] = fmt.Sprintf("\"%s\"", entity.Id)
		}
		entityAssigneesQuery := fmt.Sprintf("SELECT * FROM entity_assignee WHERE entity_id IN (%s)", strings.Join(qoutedEntityIds, ","))
		err = tx.Select(&entityAssignees, entityAssigneesQuery)
		if err != nil {
			return ProjectData{}, err
		}
	}

	templates := []models.Template{}
	if userRole.CreateTask {
		templates, err = repository.GetTemplates(tx, false)
		if err != nil {
			return ProjectData{}, err
		}
	}

	workflows := []models.Workflow{}
	if userRole.CreateTask {
		workflows, err = repository.GetWorkflows(tx)
		if err != nil {
			return ProjectData{}, err
		}
	}
	workflowLinks := []models.WorkflowLink{}
	if userRole.CreateTask {
		err = base_service.GetAll(tx, "workflow_link", &workflowLinks)
		if err != nil {
			return ProjectData{}, err
		}
	}
	workflowEntities := []models.WorkflowEntity{}
	if userRole.CreateTask {
		err = base_service.GetAll(tx, "workflow_entity", &workflowEntities)
		if err != nil {
			return ProjectData{}, err
		}
	}
	workflowTasks := []models.WorkflowTask{}
	if userRole.CreateTask {
		err = base_service.GetAll(tx, "workflow_task", &workflowTasks)
		if err != nil {
			return ProjectData{}, err
		}
	}

	tags, err := repository.GetTags(tx)
	if err != nil {
		return ProjectData{}, err
	}

	taskstagsQuery := fmt.Sprintf("SELECT * FROM task_tag WHERE task_id IN (%s)", strings.Join(quotedTaskIds, ","))
	tasksTags := []models.TaskTag{}
	err = tx.Select(&tasksTags, taskstagsQuery)
	if err != nil {
		return ProjectData{}, err
	}

	projectPreview, err := repository.GetProjectPreview(tx)
	if err != nil {
		if err.Error() != "no preview" {
			return ProjectData{}, err
		}
	}

	userData.ProjectPreview = projectPreview.Hash
	userData.EntityTypes = entityTypes
	userData.Entities = entities
	userData.EntityAssignees = entityAssignees

	userData.TaskTypes = taskTypes
	userData.Tasks = tasks
	userData.TasksCheckpoints = tasksCheckpoints
	userData.TaskDependencies = taskDependencies
	userData.EntityDependencies = entityDependencies

	userData.Statuses = statuses
	userData.DependencyTypes = dependencyTypes

	userData.Users = users
	userData.Roles = roles

	userData.Templates = templates

	userData.Workflows = workflows
	userData.WorkflowLinks = workflowLinks
	userData.WorkflowEntities = workflowEntities
	userData.WorkflowTasks = workflowTasks

	userData.Tags = tags
	userData.TasksTags = tasksTags
	return userData, nil
}

func LoadUserData(tx *sqlx.Tx, userId string) (ProjectData, error) {
	userData := ProjectData{}

	user, err := repository.GetUser(tx, userId)
	if err != nil {
		return ProjectData{}, err
	}
	userRole, err := repository.GetRole(tx, user.RoleId)
	if err != nil {
		return ProjectData{}, err
	}

	// tasks, err := repository.GetUserTasks(tx, userId)
	// if err != nil {
	// 	return ProjectData{}, err
	// }
	tasks := []models.Task{}
	if userRole.ViewTask {
		tasks, err = repository.GetTasks(tx, true)
		if err != nil {
			return ProjectData{}, err
		}
	} else {
		tasks, err = repository.GetUserTasks(tx, user.Id)
		if err != nil {
			return ProjectData{}, err
		}
	}

	var taskEntityIds []string
	var taskIds []string
	for _, task := range tasks {
		taskIds = append(taskIds, task.Id)
		if !utils.Contains(taskEntityIds, task.EntityId) {
			taskEntityIds = append(taskEntityIds, task.EntityId)
		}
	}
	quotedTaskIds := make([]string, len(taskIds))
	for i, id := range taskIds {
		quotedTaskIds[i] = fmt.Sprintf("\"%s\"", id)
	}
	quotedTaskEntityIds := make([]string, len(taskEntityIds))
	for i, id := range taskEntityIds {
		quotedTaskEntityIds[i] = fmt.Sprintf("\"%s\"", id)
	}

	// dependenciesQuery := fmt.Sprintf("SELECT * FROM task_dependencies WHERE task_id IN (%s) AND dependency_id NOT IN (%s)", strings.Join(quotedTaskIds, ","), strings.Join(quotedTaskIds, ","))
	taskDependenciesQuery := fmt.Sprintf("SELECT * FROM task_dependency WHERE task_id IN (%s)", strings.Join(quotedTaskIds, ","))
	taskDependencies := []models.TaskDependency{}
	err = tx.Select(&taskDependencies, taskDependenciesQuery)
	if err != nil {
		return ProjectData{}, err
	}

	entityDependenciesQuery := fmt.Sprintf("SELECT * FROM entity_dependency WHERE task_id IN (%s)", strings.Join(quotedTaskIds, ","))
	entityDependencies := []models.EntityDependency{}
	err = tx.Select(&entityDependencies, entityDependenciesQuery)
	if err != nil {
		return ProjectData{}, err
	}

	var uniqueDependencyIds []string
	for _, dependency := range taskDependencies {
		if !utils.Contains(uniqueDependencyIds, dependency.DependencyId) && !utils.Contains(taskIds, dependency.DependencyId) {
			uniqueDependencyIds = append(uniqueDependencyIds, dependency.DependencyId)
		}
	}
	quotedUniqueDependencyIds := make([]string, len(uniqueDependencyIds))
	for i, id := range uniqueDependencyIds {
		quotedUniqueDependencyIds[i] = fmt.Sprintf("\"%s\"", id)
	}

	uniqueDependenciesQuery := fmt.Sprintf("SELECT * FROM task WHERE trashed = 0 AND id IN (%s)", strings.Join(quotedUniqueDependencyIds, ","))
	uniqueDependencies := []models.Task{}
	err = tx.Select(&uniqueDependencies, uniqueDependenciesQuery)
	if err != nil {
		return ProjectData{}, err
	}

	tasks = append(tasks, uniqueDependencies...)
	taskIds = append(taskIds, uniqueDependencyIds...)

	quotedTaskIds = append(quotedTaskIds, quotedUniqueDependencyIds...)

	checkpointQuery := fmt.Sprintf("SELECT * FROM task_checkpoint WHERE trashed = 0 AND task_id IN (%s)", strings.Join(quotedTaskIds, ","))
	tasksCheckpoints := []models.Checkpoint{}
	err = tx.Select(&tasksCheckpoints, checkpointQuery)
	if err != nil {
		return ProjectData{}, err
	}

	statuses, err := repository.GetStatuses(tx)
	if err != nil {
		return ProjectData{}, err
	}
	taskTypes, err := repository.GetTaskTypes(tx)
	if err != nil {
		return ProjectData{}, err
	}

	users, err := repository.GetUsers(tx)
	if err != nil {
		return ProjectData{}, err
	}
	roles, err := repository.GetRoles(tx)
	if err != nil {
		return ProjectData{}, err
	}

	dependencyTypes, err := repository.GetDependencyTypes(tx)
	if err != nil {
		return ProjectData{}, err
	}

	entityTypes, err := repository.GetEntityTypes(tx)
	if err != nil {
		return ProjectData{}, err
	}

	entities := []models.Entity{}
	entityAssignees := []models.EntityAssignee{}
	if userRole.ViewTask {
		//TODO remove with trashed
		entities, err = repository.GetEntities(tx, true)
		if err != nil {
			return ProjectData{}, err
		}
		err = tx.Select(&entityAssignees, "SELECT * FROM entity_assignee")
		if err != nil {
			return ProjectData{}, err
		}
	} else {
		entities, err = repository.GetUserEntities(tx, tasks, user.Id)
		if err != nil {
			return ProjectData{}, err
		}

		qoutedEntityIds := make([]string, len(entities))
		for i, entity := range entities {
			qoutedEntityIds[i] = fmt.Sprintf("\"%s\"", entity.Id)
		}
		entityAssigneesQuery := fmt.Sprintf("SELECT * FROM entity_assignee WHERE entity_id IN (%s)", strings.Join(qoutedEntityIds, ","))
		err = tx.Select(&entityAssignees, entityAssigneesQuery)
		if err != nil {
			return ProjectData{}, err
		}
	}

	templates := []models.Template{}
	if userRole.CreateTask {
		templates, err = repository.GetTemplates(tx, false)
		if err != nil {
			return ProjectData{}, err
		}
	}

	workflows := []models.Workflow{}
	if userRole.CreateTask {
		workflows, err = repository.GetWorkflows(tx)
		if err != nil {
			return ProjectData{}, err
		}
	}
	workflowLinks := []models.WorkflowLink{}
	if userRole.CreateTask {
		err = base_service.GetAll(tx, "workflow_link", &workflowLinks)
		if err != nil {
			return ProjectData{}, err
		}
	}
	workflowEntities := []models.WorkflowEntity{}
	if userRole.CreateTask {
		err = base_service.GetAll(tx, "workflow_entity", &workflowEntities)
		if err != nil {
			return ProjectData{}, err
		}
	}
	workflowTasks := []models.WorkflowTask{}
	if userRole.CreateTask {
		err = base_service.GetAll(tx, "workflow_task", &workflowTasks)
		if err != nil {
			return ProjectData{}, err
		}
	}

	tags, err := repository.GetTags(tx)
	if err != nil {
		return ProjectData{}, err
	}

	taskstagsQuery := fmt.Sprintf("SELECT * FROM task_tag WHERE task_id IN (%s)", strings.Join(quotedTaskIds, ","))
	tasksTags := []models.TaskTag{}
	err = tx.Select(&tasksTags, taskstagsQuery)
	if err != nil {
		return ProjectData{}, err
	}

	projectPreview, err := repository.GetProjectPreview(tx)
	if err != nil {
		if err.Error() != "no preview" {
			return ProjectData{}, err
		}
	}

	userData.ProjectPreview = projectPreview.Hash
	userData.EntityTypes = entityTypes
	userData.Entities = entities
	userData.EntityAssignees = entityAssignees

	userData.TaskTypes = taskTypes
	userData.Tasks = tasks
	userData.TasksCheckpoints = tasksCheckpoints
	userData.TaskDependencies = taskDependencies
	userData.EntityDependencies = entityDependencies

	userData.Statuses = statuses
	userData.DependencyTypes = dependencyTypes

	userData.Users = users
	userData.Roles = roles

	userData.Templates = templates

	userData.Workflows = workflows
	userData.WorkflowLinks = workflowLinks
	userData.WorkflowEntities = workflowEntities
	userData.WorkflowTasks = workflowTasks

	userData.Tags = tags
	userData.TasksTags = tasksTags
	return userData, nil
}

func LoadUserDataPb(tx *sqlx.Tx, userId string) ([]byte, error) {
	user, err := repository.GetUser(tx, userId)
	if err != nil {
		return []byte{}, err
	}
	userRole, err := repository.GetRole(tx, user.RoleId)
	if err != nil {
		return []byte{}, err
	}

	// tasks, err := repository.GetUserTasks(tx, userId)
	// if err != nil {
	// 	return ProjectData{}, err
	// }
	tasks := []models.Task{}
	if userRole.ViewTask {
		tasks, err = repository.GetTasks(tx, true)
		if err != nil {
			return []byte{}, err
		}
	} else {
		tasks, err = repository.GetUserTasks(tx, user.Id)
		if err != nil {
			return []byte{}, err
		}
	}

	var taskEntityIds []string
	var taskIds []string
	for _, task := range tasks {
		taskIds = append(taskIds, task.Id)
		if !utils.Contains(taskEntityIds, task.EntityId) {
			taskEntityIds = append(taskEntityIds, task.EntityId)
		}
	}
	quotedTaskIds := make([]string, len(taskIds))
	for i, id := range taskIds {
		quotedTaskIds[i] = fmt.Sprintf("\"%s\"", id)
	}
	quotedTaskEntityIds := make([]string, len(taskEntityIds))
	for i, id := range taskEntityIds {
		quotedTaskEntityIds[i] = fmt.Sprintf("\"%s\"", id)
	}

	// dependenciesQuery := fmt.Sprintf("SELECT * FROM task_dependencies WHERE task_id IN (%s) AND dependency_id NOT IN (%s)", strings.Join(quotedTaskIds, ","), strings.Join(quotedTaskIds, ","))
	taskDependenciesQuery := fmt.Sprintf("SELECT * FROM task_dependency WHERE task_id IN (%s)", strings.Join(quotedTaskIds, ","))
	taskDependencies := []models.TaskDependency{}
	err = tx.Select(&taskDependencies, taskDependenciesQuery)
	if err != nil {
		return []byte{}, err
	}

	entityDependenciesQuery := fmt.Sprintf("SELECT * FROM entity_dependency WHERE task_id IN (%s)", strings.Join(quotedTaskIds, ","))
	entityDependencies := []models.EntityDependency{}
	err = tx.Select(&entityDependencies, entityDependenciesQuery)
	if err != nil {
		return []byte{}, err
	}

	var uniqueDependencyIds []string
	for _, dependency := range taskDependencies {
		if !utils.Contains(uniqueDependencyIds, dependency.DependencyId) && !utils.Contains(taskIds, dependency.DependencyId) {
			uniqueDependencyIds = append(uniqueDependencyIds, dependency.DependencyId)
		}
	}
	quotedUniqueDependencyIds := make([]string, len(uniqueDependencyIds))
	for i, id := range uniqueDependencyIds {
		quotedUniqueDependencyIds[i] = fmt.Sprintf("\"%s\"", id)
	}

	uniqueDependenciesQuery := fmt.Sprintf("SELECT * FROM task WHERE trashed = 0 AND id IN (%s)", strings.Join(quotedUniqueDependencyIds, ","))
	uniqueDependencies := []models.Task{}
	err = tx.Select(&uniqueDependencies, uniqueDependenciesQuery)
	if err != nil {
		return []byte{}, err
	}

	tasks = append(tasks, uniqueDependencies...)
	taskIds = append(taskIds, uniqueDependencyIds...)

	quotedTaskIds = append(quotedTaskIds, quotedUniqueDependencyIds...)

	checkpointQuery := fmt.Sprintf("SELECT * FROM task_checkpoint WHERE trashed = 0 AND task_id IN (%s)", strings.Join(quotedTaskIds, ","))
	tasksCheckpoints := []models.Checkpoint{}
	err = tx.Select(&tasksCheckpoints, checkpointQuery)
	if err != nil {
		return []byte{}, err
	}

	statuses, err := repository.GetStatuses(tx)
	if err != nil {
		return []byte{}, err
	}
	taskTypes, err := repository.GetTaskTypes(tx)
	if err != nil {
		return []byte{}, err
	}

	users, err := repository.GetUsers(tx)
	if err != nil {
		return []byte{}, err
	}
	roles, err := repository.GetRoles(tx)
	if err != nil {
		return []byte{}, err
	}

	dependencyTypes, err := repository.GetDependencyTypes(tx)
	if err != nil {
		return []byte{}, err
	}

	entityTypes, err := repository.GetEntityTypes(tx)
	if err != nil {
		return []byte{}, err
	}

	entities := []models.Entity{}
	entityAssignees := []models.EntityAssignee{}
	if userRole.ViewTask {
		//TODO remove with trashed
		entities, err = repository.GetEntities(tx, true)
		if err != nil {
			return []byte{}, err
		}
		err = tx.Select(&entityAssignees, "SELECT * FROM entity_assignee")
		if err != nil {
			return []byte{}, err
		}
	} else {
		entities, err = repository.GetUserEntities(tx, tasks, user.Id)
		if err != nil {
			return []byte{}, err
		}

		qoutedEntityIds := make([]string, len(entities))
		for i, entity := range entities {
			qoutedEntityIds[i] = fmt.Sprintf("\"%s\"", entity.Id)
		}
		entityAssigneesQuery := fmt.Sprintf("SELECT * FROM entity_assignee WHERE entity_id IN (%s)", strings.Join(qoutedEntityIds, ","))
		err = tx.Select(&entityAssignees, entityAssigneesQuery)
		if err != nil {
			return []byte{}, err
		}
	}

	templates := []models.Template{}
	if userRole.CreateTask {
		templates, err = repository.GetTemplates(tx, false)
		if err != nil {
			return []byte{}, err
		}
	}

	workflows := []models.Workflow{}
	if userRole.CreateTask {
		workflows, err = repository.GetWorkflows(tx)
		if err != nil {
			return []byte{}, err
		}
	}
	workflowLinks := []models.WorkflowLink{}
	if userRole.CreateTask {
		err = base_service.GetAll(tx, "workflow_link", &workflowLinks)
		if err != nil {
			return []byte{}, err
		}
	}
	workflowEntities := []models.WorkflowEntity{}
	if userRole.CreateTask {
		err = base_service.GetAll(tx, "workflow_entity", &workflowEntities)
		if err != nil {
			return []byte{}, err
		}
	}
	workflowTasks := []models.WorkflowTask{}
	if userRole.CreateTask {
		err = base_service.GetAll(tx, "workflow_task", &workflowTasks)
		if err != nil {
			return []byte{}, err
		}
	}

	tags, err := repository.GetTags(tx)
	if err != nil {
		return []byte{}, err
	}

	taskstagsQuery := fmt.Sprintf("SELECT * FROM task_tag WHERE task_id IN (%s)", strings.Join(quotedTaskIds, ","))
	tasksTags := []models.TaskTag{}
	err = tx.Select(&tasksTags, taskstagsQuery)
	if err != nil {
		return []byte{}, err
	}

	projectPreview, err := repository.GetProjectPreview(tx)
	if err != nil {
		if err.Error() != "no preview" {
			return []byte{}, err
		}
	}

	userData := &repositorypb.ProjectData{
		ProjectPreview:  projectPreview.Hash,
		EntityTypes:     repository.ToPbEntityTypes(entityTypes),
		Entities:        repository.ToPbEntities(entities),
		EntityAssignees: repository.ToPbEntityAssignees(entityAssignees),

		TaskTypes:          repository.ToPbTaskTypes(taskTypes),
		Tasks:              repository.ToPbTasks(tasks),
		TasksCheckpoints:   repository.ToPbCheckpoints(tasksCheckpoints),
		TaskDependencies:   repository.ToPbTaskDependencies(taskDependencies),
		EntityDependencies: repository.ToPbEntityDependencies(entityDependencies),

		Statuses:        repository.ToPbStatuses(statuses),
		DependencyTypes: repository.ToPbDependencyTypes(dependencyTypes),

		Users: repository.ToPbUsers(users),
		Roles: repository.ToPbRoles(roles),

		Templates: repository.ToPbTemplates(templates),

		Workflows:        repository.ToPbWorkflows(workflows),
		WorkflowLinks:    repository.ToPbWorkflowLinks(workflowLinks),
		WorkflowEntities: repository.ToPbWorkflowEntities(workflowEntities),
		WorkflowTasks:    repository.ToPbWorkflowTasks(workflowTasks),

		Tags:      repository.ToPbTags(tags),
		TasksTags: repository.ToPbTaskTags(tasksTags),
	}
	userDataBytes, err := proto.Marshal(userData)
	if err != nil {
		return []byte{}, err
	}

	return userDataBytes, nil

	// userData.ProjectPreview = projectPreview.Hash
	// userData.EntityTypes = repository.ToPbEntityTypes(entityTypes)
	// userData.Entities = repository.ToPbEntities(entities)
	// userData.EntityAssignees = repository.ToPbEntityAssignees(entityAssignees)

	// userData.TaskTypes = repository.ToPbTaskTypes(taskTypes)
	// userData.Tasks = repository.ToPbTasks(tasks)
	// userData.TasksCheckpoints = repository.ToPbCheckpoints(tasksCheckpoints)
	// userData.TaskDependencies = repository.ToPbTaskDependencies(taskDependencies)
	// userData.EntityDependencies = repository.ToPbEntityDependencies(entityDependencies)

	// userData.Statuses = repository.ToPbStatuses(statuses)
	// userData.DependencyTypes = repository.ToPbDependencyTypes(dependencyTypes)

	// userData.Users = repository.ToPbUsers(users)
	// userData.Roles = repository.ToPbRoles(roles)

	// userData.Templates = repository.ToPbTemplates(templates)

	// userData.Workflows = repository.ToPbWorkflows(workflows)
	// userData.WorkflowLinks = repository.ToPbWorkflowLinks(workflowLinks)
	// userData.WorkflowEntities = repository.ToPbWorkflowEntities(workflowEntities)
	// userData.WorkflowTasks = repository.ToPbWorkflowTasks(workflowTasks)

	// userData.Tags = repository.ToPbTags(tags)
	// userData.TasksTags = repository.ToPbTaskTags(tasksTags)
	// return userData, nil
}

func LoadChangedData(tx *sqlx.Tx) (ProjectData, error) {
	userData := ProjectData{}

	taskQuery := "SELECT * FROM task WHERE synced = 0"
	tasks := []models.Task{}
	err := tx.Select(&tasks, taskQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	checkpointQuery := "SELECT * FROM task_checkpoint WHERE synced = 0"
	tasksCheckpoints := []models.Checkpoint{}
	err = tx.Select(&tasksCheckpoints, checkpointQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	taskDependenciesQuery := "SELECT * FROM task_dependency WHERE synced = 0"
	taskDependencies := []models.TaskDependency{}
	err = tx.Select(&taskDependencies, taskDependenciesQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	entityDependenciesQuery := "SELECT * FROM entity_dependency WHERE synced = 0"
	entityDependencies := []models.EntityDependency{}
	err = tx.Select(&entityDependencies, entityDependenciesQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	statusesQuery := "SELECT * FROM status WHERE synced = 0"
	statuses := []models.Status{}
	err = tx.Select(&statuses, statusesQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	dependencyTypesQuery := "SELECT * FROM dependency_type WHERE synced = 0"
	dependencyTypes := []models.DependencyType{}
	err = tx.Select(&dependencyTypes, dependencyTypesQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	taskTypeQuery := "SELECT * FROM task_type WHERE synced = 0"
	taskTypes := []models.TaskType{}
	err = tx.Select(&taskTypes, taskTypeQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	entityTypesQuery := "SELECT * FROM entity_type WHERE synced = 0"
	entityTypes := []models.EntityType{}
	err = tx.Select(&entityTypes, entityTypesQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	entityQuery := "SELECT * FROM entity WHERE synced = 0"
	entities := []models.Entity{}
	err = tx.Select(&entities, entityQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	entityAssigneeQuery := "SELECT * FROM entity_assignee WHERE synced = 0"
	entityAssignees := []models.EntityAssignee{}
	err = tx.Select(&entityAssignees, entityAssigneeQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	userQuery := "SELECT * FROM user WHERE synced = 0"
	users := []models.User{}
	err = tx.Select(&users, userQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	roleQuery := "SELECT * FROM role WHERE synced = 0"
	roles := []models.Role{}
	err = tx.Select(&roles, roleQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	templatesQuery := "SELECT * FROM template WHERE synced = 0"
	templates := []models.Template{}
	err = tx.Select(&templates, templatesQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	workflowsQuery := "SELECT * FROM workflow WHERE synced = 0"
	workflows := []models.Workflow{}
	err = tx.Select(&workflows, workflowsQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}
	workflowLinksQuery := "SELECT * FROM workflow_link WHERE synced = 0"
	workflowLinks := []models.WorkflowLink{}
	err = tx.Select(&workflowLinks, workflowLinksQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}
	workflowEntitiesQuery := "SELECT * FROM workflow_entity WHERE synced = 0"
	workflowEntities := []models.WorkflowEntity{}
	err = tx.Select(&workflowEntities, workflowEntitiesQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}
	workflowTasksQuery := "SELECT * FROM workflow_task WHERE synced = 0"
	workflowTasks := []models.WorkflowTask{}
	err = tx.Select(&workflowTasks, workflowTasksQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	tagsQuery := "SELECT * FROM tag WHERE synced = 0"
	tags := []models.Tag{}
	err = tx.Select(&tags, tagsQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}
	tasksTagsQuery := "SELECT * FROM task_tag WHERE synced = 0"
	tasksTags := []models.TaskTag{}
	err = tx.Select(&tasksTags, tasksTagsQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	tombs, err := repository.GetTombs(tx)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}
	isProjectPreviewSynced, err := repository.IsProjectPreviewSynced(tx)
	if err != nil {
		return userData, err
	}
	if !isProjectPreviewSynced {
		projectPreview, err := repository.GetProjectPreview(tx)
		if err != nil {
			return ProjectData{}, err
		}
		userData.ProjectPreview = projectPreview.Hash
	}
	userData.TaskTypes = taskTypes
	userData.Tasks = tasks
	userData.TasksCheckpoints = tasksCheckpoints
	userData.TaskDependencies = taskDependencies
	userData.EntityDependencies = entityDependencies

	userData.Statuses = statuses
	userData.DependencyTypes = dependencyTypes

	userData.Users = users
	userData.Roles = roles

	userData.EntityTypes = entityTypes
	userData.Entities = entities
	userData.EntityAssignees = entityAssignees

	userData.Templates = templates

	userData.Workflows = workflows
	userData.WorkflowLinks = workflowLinks
	userData.WorkflowEntities = workflowEntities
	userData.WorkflowTasks = workflowTasks

	userData.Tags = tags
	userData.TasksTags = tasksTags

	userData.Tombs = tombs
	return userData, nil
}

func LoadChangedDataPb(tx *sqlx.Tx) ([]byte, error) {

	taskQuery := "SELECT * FROM task WHERE synced = 0"
	tasks := []models.Task{}
	err := tx.Select(&tasks, taskQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	checkpointQuery := "SELECT * FROM task_checkpoint WHERE synced = 0"
	tasksCheckpoints := []models.Checkpoint{}
	err = tx.Select(&tasksCheckpoints, checkpointQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	taskDependenciesQuery := "SELECT * FROM task_dependency WHERE synced = 0"
	taskDependencies := []models.TaskDependency{}
	err = tx.Select(&taskDependencies, taskDependenciesQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	entityDependenciesQuery := "SELECT * FROM entity_dependency WHERE synced = 0"
	entityDependencies := []models.EntityDependency{}
	err = tx.Select(&entityDependencies, entityDependenciesQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	statusesQuery := "SELECT * FROM status WHERE synced = 0"
	statuses := []models.Status{}
	err = tx.Select(&statuses, statusesQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	dependencyTypesQuery := "SELECT * FROM dependency_type WHERE synced = 0"
	dependencyTypes := []models.DependencyType{}
	err = tx.Select(&dependencyTypes, dependencyTypesQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	taskTypeQuery := "SELECT * FROM task_type WHERE synced = 0"
	taskTypes := []models.TaskType{}
	err = tx.Select(&taskTypes, taskTypeQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	entityTypesQuery := "SELECT * FROM entity_type WHERE synced = 0"
	entityTypes := []models.EntityType{}
	err = tx.Select(&entityTypes, entityTypesQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	entityQuery := "SELECT * FROM entity WHERE synced = 0"
	entities := []models.Entity{}
	err = tx.Select(&entities, entityQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	entityAssigneeQuery := "SELECT * FROM entity_assignee WHERE synced = 0"
	entityAssignees := []models.EntityAssignee{}
	err = tx.Select(&entityAssignees, entityAssigneeQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	userQuery := "SELECT * FROM user WHERE synced = 0"
	users := []models.User{}
	err = tx.Select(&users, userQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	roleQuery := "SELECT * FROM role WHERE synced = 0"
	roles := []models.Role{}
	err = tx.Select(&roles, roleQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	templatesQuery := "SELECT * FROM template WHERE synced = 0"
	templates := []models.Template{}
	err = tx.Select(&templates, templatesQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	workflowsQuery := "SELECT * FROM workflow WHERE synced = 0"
	workflows := []models.Workflow{}
	err = tx.Select(&workflows, workflowsQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}
	workflowLinksQuery := "SELECT * FROM workflow_link WHERE synced = 0"
	workflowLinks := []models.WorkflowLink{}
	err = tx.Select(&workflowLinks, workflowLinksQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}
	workflowEntitiesQuery := "SELECT * FROM workflow_entity WHERE synced = 0"
	workflowEntities := []models.WorkflowEntity{}
	err = tx.Select(&workflowEntities, workflowEntitiesQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}
	workflowTasksQuery := "SELECT * FROM workflow_task WHERE synced = 0"
	workflowTasks := []models.WorkflowTask{}
	err = tx.Select(&workflowTasks, workflowTasksQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	tagsQuery := "SELECT * FROM tag WHERE synced = 0"
	tags := []models.Tag{}
	err = tx.Select(&tags, tagsQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}
	tasksTagsQuery := "SELECT * FROM task_tag WHERE synced = 0"
	tasksTags := []models.TaskTag{}
	err = tx.Select(&tasksTags, tasksTagsQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	tombs, err := repository.GetTombs(tx)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}
	isProjectPreviewSynced, err := repository.IsProjectPreviewSynced(tx)
	if err != nil {
		return []byte{}, err
	}
	projectPreview := models.Preview{}
	if !isProjectPreviewSynced {
		projectPreview, err = repository.GetProjectPreview(tx)
		if err != nil {
			return []byte{}, err
		}
	}

	userData := &repositorypb.ProjectData{
		ProjectPreview:  projectPreview.Hash,
		EntityTypes:     repository.ToPbEntityTypes(entityTypes),
		Entities:        repository.ToPbEntities(entities),
		EntityAssignees: repository.ToPbEntityAssignees(entityAssignees),

		TaskTypes:          repository.ToPbTaskTypes(taskTypes),
		Tasks:              repository.ToPbTasks(tasks),
		TasksCheckpoints:   repository.ToPbCheckpoints(tasksCheckpoints),
		TaskDependencies:   repository.ToPbTaskDependencies(taskDependencies),
		EntityDependencies: repository.ToPbEntityDependencies(entityDependencies),

		Statuses:        repository.ToPbStatuses(statuses),
		DependencyTypes: repository.ToPbDependencyTypes(dependencyTypes),

		Users: repository.ToPbUsers(users),
		Roles: repository.ToPbRoles(roles),

		Templates: repository.ToPbTemplates(templates),

		Workflows:        repository.ToPbWorkflows(workflows),
		WorkflowLinks:    repository.ToPbWorkflowLinks(workflowLinks),
		WorkflowEntities: repository.ToPbWorkflowEntities(workflowEntities),
		WorkflowTasks:    repository.ToPbWorkflowTasks(workflowTasks),

		Tags:      repository.ToPbTags(tags),
		TasksTags: repository.ToPbTaskTags(tasksTags),

		Tomb: repository.ToPbTombs(tombs),
	}
	userDataBytes, err := proto.Marshal(userData)
	if err != nil {
		return []byte{}, err
	}
	return userDataBytes, nil
}

func LoadCheckpointData(tx *sqlx.Tx) (ProjectData, error) {
	userData := ProjectData{}

	taskCheckpointQuery := "SELECT * FROM task_checkpoint WHERE synced = 0"
	tasksCheckpoints := []models.Checkpoint{}
	err := tx.Select(&tasksCheckpoints, taskCheckpointQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	userData.TasksCheckpoints = tasksCheckpoints
	return userData, nil
}
