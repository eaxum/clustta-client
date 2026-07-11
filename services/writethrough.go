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
	ids         []string
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
		syncTargets[item.table] = append(syncTargets[item.table], item.ids...)
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
	dst.Assets = append(dst.Assets, src.Assets...)
	dst.AssetTypes = append(dst.AssetTypes, src.AssetTypes...)
	dst.AssetCheckpoints = append(dst.AssetCheckpoints, src.AssetCheckpoints...)
	dst.AssetDependencies = append(dst.AssetDependencies, src.AssetDependencies...)
	dst.CollectionDependencies = append(dst.CollectionDependencies, src.CollectionDependencies...)
	dst.Statuses = append(dst.Statuses, src.Statuses...)
	dst.DependencyTypes = append(dst.DependencyTypes, src.DependencyTypes...)
	dst.Users = append(dst.Users, src.Users...)
	dst.Roles = append(dst.Roles, src.Roles...)
	dst.CollectionTypes = append(dst.CollectionTypes, src.CollectionTypes...)
	dst.Collections = append(dst.Collections, src.Collections...)
	dst.CollectionAssignees = append(dst.CollectionAssignees, src.CollectionAssignees...)
	dst.Templates = append(dst.Templates, src.Templates...)
	dst.Tags = append(dst.Tags, src.Tags...)
	dst.AssetTags = append(dst.AssetTags, src.AssetTags...)
	dst.Workflows = append(dst.Workflows, src.Workflows...)
	dst.WorkflowLinks = append(dst.WorkflowLinks, src.WorkflowLinks...)
	dst.WorkflowCollections = append(dst.WorkflowCollections, src.WorkflowCollections...)
	dst.WorkflowAssets = append(dst.WorkflowAssets, src.WorkflowAssets...)
	dst.Tombs = append(dst.Tombs, src.Tombs...)
}

// enqueueWriteThrough is the single entry point for all service methods.
// It accepts pre-read data so no second DB open is needed to re-read rows.
func enqueueWriteThrough(projectPath, table, id string, data sync_service.ProjectData) {
	batcher.enqueue(writeThroughItem{
		projectPath: projectPath,
		table:       table,
		ids:         []string{id},
		data:        data,
	})
}

func enqueueWriteThroughBatch(projectPath, table string, ids []string, data sync_service.ProjectData) {
	batcher.enqueue(writeThroughItem{
		projectPath: projectPath,
		table:       table,
		ids:         ids,
		data:        data,
	})
}

// Convenience wrappers that accept pre-read model data from the service method's
// existing transaction, avoiding any additional DB reads.

// enqueueAssetWriteThrough queues a asset for batched push.
func enqueueAssetWriteThrough(projectPath string, asset models.Asset) {
	enqueueWriteThrough(projectPath, "asset", asset.Id, sync_service.ProjectData{
		Assets: []models.Asset{asset},
	})
}

func enqueueAssetsWriteThrough(projectPath string, assets []models.Asset) {
	if len(assets) == 0 {
		return
	}

	assetIds := make([]string, 0, len(assets))
	for _, asset := range assets {
		assetIds = append(assetIds, asset.Id)
	}

	enqueueWriteThroughBatch(projectPath, "asset", assetIds, sync_service.ProjectData{
		Assets: assets,
	})
}

// enqueueCollectionWriteThrough queues an collection for batched push.
func enqueueCollectionWriteThrough(projectPath string, collection models.Collection) {
	enqueueWriteThrough(projectPath, "collection", collection.Id, sync_service.ProjectData{
		Collections: []models.Collection{collection},
	})
}

// enqueueCollectionAssigneeWriteThrough queues an collection assignee for batched push.
func enqueueCollectionAssigneeWriteThrough(projectPath string, assignee models.CollectionAssignee) {
	enqueueWriteThrough(projectPath, "collection_assignee", assignee.Id, sync_service.ProjectData{
		CollectionAssignees: []models.CollectionAssignee{assignee},
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

// enqueueDependencyWriteThrough queues a asset dependency for batched push.
func enqueueDependencyWriteThrough(projectPath string, dep models.AssetDependency) {
	enqueueWriteThrough(projectPath, "asset_dependency", dep.Id, sync_service.ProjectData{
		AssetDependencies: []models.AssetDependency{dep},
	})
}

// enqueueCollectionDependencyWriteThrough queues an collection dependency for batched push.
// Accepts models.AssetDependency since repository.AddCollectionDependency returns that type,
// then converts to models.CollectionDependency for the ProjectData payload.
func enqueueCollectionDependencyWriteThrough(projectPath string, dep models.AssetDependency) {
	collectionDep := models.CollectionDependency{
		Id:               dep.Id,
		MTime:            dep.MTime,
		AssetId:          dep.AssetId,
		DependencyId:     dep.DependencyId,
		DependencyTypeId: dep.DependencyTypeId,
		Synced:           dep.Synced,
	}
	enqueueWriteThrough(projectPath, "collection_dependency", dep.Id, sync_service.ProjectData{
		CollectionDependencies: []models.CollectionDependency{collectionDep},
	})
}

// enqueueAssetTypeWriteThrough queues a asset type for batched push.
func enqueueAssetTypeWriteThrough(projectPath string, assetType models.AssetType) {
	enqueueWriteThrough(projectPath, "asset_type", assetType.Id, sync_service.ProjectData{
		AssetTypes: []models.AssetType{assetType},
	})
}

// enqueueCollectionTypeWriteThrough queues an collection type for batched push.
func enqueueCollectionTypeWriteThrough(projectPath string, collectionType models.CollectionType) {
	enqueueWriteThrough(projectPath, "collection_type", collectionType.Id, sync_service.ProjectData{
		CollectionTypes: []models.CollectionType{collectionType},
	})
}
