package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/sync_service"
	"clustta/internal/utils"
	"log/slog"
	"sync"
	"time"
)

// writeThroughItem represents a single queued metadata change.
type writeThroughItem struct {
	projectPath string
	table       string
	id          string
	data        sync_service.ProjectData
}

// writeThroughBatcher collects items and flushes them in a single push per project.
type writeThroughBatcher struct {
	mu           sync.Mutex
	pending      map[string][]writeThroughItem // keyed by projectPath
	timer        *time.Timer
	remoteCache  sync.Map // projectPath -> cached remote URL string
	enabledCache sync.Map // projectPath -> cached write-through enabled bool
}

var batcher = &writeThroughBatcher{
	pending: make(map[string][]writeThroughItem),
}

const flushInterval = 2 * time.Second

// isWriteThroughEnabled checks the project's config table for the experimental setting.
// Returns false if the setting is missing or cannot be read.
func (b *writeThroughBatcher) isWriteThroughEnabled(projectPath string) bool {
	if cached, ok := b.enabledCache.Load(projectPath); ok {
		return cached.(bool)
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return false
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return false
	}
	defer tx.Rollback()

	enabled, err := utils.GetWriteThroughEnabled(tx)
	if err != nil {
		return false
	}

	b.enabledCache.Store(projectPath, enabled)
	return enabled
}

// enqueue adds an item to the batch and starts or resets the flush timer.
func (b *writeThroughBatcher) enqueue(item writeThroughItem) {
	if !b.isWriteThroughEnabled(item.projectPath) {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.pending[item.projectPath] = append(b.pending[item.projectPath], item)

	if b.timer == nil {
		b.timer = time.AfterFunc(flushInterval, b.flush)
	} else {
		b.timer.Reset(flushInterval)
	}
}

// flush drains all pending items, merges them by project, and pushes each batch.
func (b *writeThroughBatcher) flush() {
	b.mu.Lock()
	batches := b.pending
	b.pending = make(map[string][]writeThroughItem)
	b.timer = nil
	b.mu.Unlock()

	for projectPath, items := range batches {
		go b.pushBatch(projectPath, items)
	}
}

// pushBatch merges items into a single ProjectData and pushes to the server.
func (b *writeThroughBatcher) pushBatch(projectPath string, items []writeThroughItem) {
	remoteURL, err := b.getRemoteURL(projectPath)
	if err != nil || remoteURL == "" {
		return
	}

	activeUser, err := auth_service.GetActiveUser()
	if err != nil {
		slog.Warn("write-through: auth error", "error", err)
		return
	}

	// Merge all items into one ProjectData and one syncTargets map
	merged := sync_service.ProjectData{}
	syncTargets := make(map[string][]string)

	for _, item := range items {
		syncTargets[item.table] = append(syncTargets[item.table], item.id)
		mergeProjectData(&merged, &item.data)
	}

	err = sync_service.PushPartialData(projectPath, remoteURL, activeUser.Id, merged, syncTargets)
	if err != nil {
		slog.Warn("write-through: push error", "error", err, "project", projectPath, "items", len(items))
	}
}

// getRemoteURL returns the cached remote URL for a project, reading from DB on first access.
func (b *writeThroughBatcher) getRemoteURL(projectPath string) (string, error) {
	if cached, ok := b.remoteCache.Load(projectPath); ok {
		return cached.(string), nil
	}

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

	remote, err := utils.GetRemoteUrl(tx)
	if err != nil || remote == "" {
		return "", err
	}

	b.remoteCache.Store(projectPath, remote)
	return remote, nil
}

// InvalidateRemoteCache clears the cached remote URL for a project.
// Call this after a full sync or project configuration change.
func InvalidateRemoteCache(projectPath string) {
	batcher.remoteCache.Delete(projectPath)
}

// InvalidateEnabledCache clears the cached write-through enabled flag for a project.
// Call this after the setting is toggled.
func InvalidateEnabledCache(projectPath string) {
	batcher.enabledCache.Delete(projectPath)
}

// mergeProjectData appends all slices from src into dst.
func mergeProjectData(dst, src *sync_service.ProjectData) {
	dst.Tasks = append(dst.Tasks, src.Tasks...)
	dst.TaskTypes = append(dst.TaskTypes, src.TaskTypes...)
	dst.TasksCheckpoints = append(dst.TasksCheckpoints, src.TasksCheckpoints...)
	dst.TaskDependencies = append(dst.TaskDependencies, src.TaskDependencies...)
	dst.EntityDependencies = append(dst.EntityDependencies, src.EntityDependencies...)
	dst.Statuses = append(dst.Statuses, src.Statuses...)
	dst.DependencyTypes = append(dst.DependencyTypes, src.DependencyTypes...)
	dst.Users = append(dst.Users, src.Users...)
	dst.Roles = append(dst.Roles, src.Roles...)
	dst.EntityTypes = append(dst.EntityTypes, src.EntityTypes...)
	dst.Entities = append(dst.Entities, src.Entities...)
	dst.EntityAssignees = append(dst.EntityAssignees, src.EntityAssignees...)
	dst.Templates = append(dst.Templates, src.Templates...)
	dst.Tags = append(dst.Tags, src.Tags...)
	dst.TasksTags = append(dst.TasksTags, src.TasksTags...)
	dst.Workflows = append(dst.Workflows, src.Workflows...)
	dst.WorkflowLinks = append(dst.WorkflowLinks, src.WorkflowLinks...)
	dst.WorkflowEntities = append(dst.WorkflowEntities, src.WorkflowEntities...)
	dst.WorkflowTasks = append(dst.WorkflowTasks, src.WorkflowTasks...)
	dst.Tombs = append(dst.Tombs, src.Tombs...)
}

// enqueueWriteThrough is the single entry point for all service methods.
// It accepts pre-read data so no second DB open is needed to re-read rows.
func enqueueWriteThrough(projectPath, table, id string, data sync_service.ProjectData) {
	batcher.enqueue(writeThroughItem{
		projectPath: projectPath,
		table:       table,
		id:          id,
		data:        data,
	})
}

// Convenience wrappers that accept pre-read model data from the service method's
// existing transaction, avoiding any additional DB reads.

// enqueueTaskWriteThrough queues a task for batched push.
func enqueueTaskWriteThrough(projectPath string, task models.Task) {
	enqueueWriteThrough(projectPath, "task", task.Id, sync_service.ProjectData{
		Tasks: []models.Task{task},
	})
}

// enqueueEntityWriteThrough queues an entity for batched push.
func enqueueEntityWriteThrough(projectPath string, entity models.Entity) {
	enqueueWriteThrough(projectPath, "entity", entity.Id, sync_service.ProjectData{
		Entities: []models.Entity{entity},
	})
}

// enqueueEntityAssigneeWriteThrough queues an entity assignee for batched push.
func enqueueEntityAssigneeWriteThrough(projectPath string, assignee models.EntityAssignee) {
	enqueueWriteThrough(projectPath, "entity_assignee", assignee.Id, sync_service.ProjectData{
		EntityAssignees: []models.EntityAssignee{assignee},
	})
}

// enqueueUserWriteThrough queues a user for batched push.
func enqueueUserWriteThrough(projectPath string, user models.User) {
	enqueueWriteThrough(projectPath, "user", user.Id, sync_service.ProjectData{
		Users: []models.User{user},
	})
}

// enqueueTombWriteThrough queues a tomb entry for batched push.
func enqueueTombWriteThrough(projectPath string, tomb repository.Tomb) {
	enqueueWriteThrough(projectPath, "tomb", tomb.Id, sync_service.ProjectData{
		Tombs: []repository.Tomb{tomb},
	})
}

// enqueueDependencyWriteThrough queues a task dependency for batched push.
func enqueueDependencyWriteThrough(projectPath string, dep models.TaskDependency) {
	enqueueWriteThrough(projectPath, "task_dependency", dep.Id, sync_service.ProjectData{
		TaskDependencies: []models.TaskDependency{dep},
	})
}

// enqueueEntityDependencyWriteThrough queues an entity dependency for batched push.
// Accepts models.TaskDependency since repository.AddEntityDependency returns that type,
// then converts to models.EntityDependency for the ProjectData payload.
func enqueueEntityDependencyWriteThrough(projectPath string, dep models.TaskDependency) {
	entityDep := models.EntityDependency{
		Id:               dep.Id,
		MTime:            dep.MTime,
		TaskId:           dep.TaskId,
		DependencyId:     dep.DependencyId,
		DependencyTypeId: dep.DependencyTypeId,
		Synced:           dep.Synced,
	}
	enqueueWriteThrough(projectPath, "entity_dependency", dep.Id, sync_service.ProjectData{
		EntityDependencies: []models.EntityDependency{entityDep},
	})
}

// enqueueTaskTypeWriteThrough queues a task type for batched push.
func enqueueTaskTypeWriteThrough(projectPath string, taskType models.TaskType) {
	enqueueWriteThrough(projectPath, "task_type", taskType.Id, sync_service.ProjectData{
		TaskTypes: []models.TaskType{taskType},
	})
}

// enqueueEntityTypeWriteThrough queues an entity type for batched push.
func enqueueEntityTypeWriteThrough(projectPath string, entityType models.EntityType) {
	enqueueWriteThrough(projectPath, "entity_type", entityType.Id, sync_service.ProjectData{
		EntityTypes: []models.EntityType{entityType},
	})
}
