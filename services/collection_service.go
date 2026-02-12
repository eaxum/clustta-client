package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/error_service"
	"clustta/internal/ignore"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/sync_service"
	"clustta/internal/utils"
	"clustta/output"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/jmoiron/sqlx"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type EntityItems struct {
	Tasks            []models.Task            `json:"tasks"`
	Entities         []models.Entity          `json:"entities"`
	UntrackedFiles   []models.UntrackedTask   `json:"untracked_tasks"`
	UntrackedFolders []models.UntrackedEntity `json:"untracked_entities"`
}

type CollectionStateFlags struct {
	HasUntracked   bool `json:"has_untracked"`
	HasModified    bool `json:"has_modified"`
	HasOutdated    bool `json:"has_outdated"`
	HasRebuildable bool `json:"has_rebuildable"`
}

type CollectionChildrenState struct {
	ModifiedTasks    []models.Task            `json:"modified_tasks"`
	OutdatedTasks    []models.Task            `json:"outdated_tasks"`
	RebuildableTasks []models.Task            `json:"rebuildable_tasks"`
	NormalTasks      []models.Task            `json:"normal_tasks"`
	UntrackedFiles   []models.UntrackedTask   `json:"untracked_files"`
	UntrackedFolders []models.UntrackedEntity `json:"untracked_folders"`
}

type ItemsForCheckpoint struct {
	ModifiedTasks  []models.Task          `json:"modified_tasks"`
	UntrackedFiles []models.UntrackedTask `json:"untracked_files"`
}

type ItemsForUpdate struct {
	OutdatedTasks []models.Task `json:"outdated_tasks"`
}

type CollectionService struct {
}

// GetCollectionCount returns the total number of collections in the project.
// Returns the count or an error if the operation fails.
func (t *CollectionService) GetCollectionCount(projectPath string) (int, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return 0, err
	}
	defer dbConn.Close()

	var count int
	query := "SELECT COUNT(*) FROM full_entity"

	err = dbConn.Get(&count, query)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// CreateCollection creates a new collection in the project.
// Returns the created entity or an error if the operation fails.
func (e *CollectionService) CreateCollection(projectPath, name, description, entityTypeId, parentId, previewPath string, isLibrary bool) (models.Entity, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Entity{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Entity{}, err
	}
	defer tx.Rollback()

	previewId := ""
	if previewPath != "" {
		preview, err := repository.CreatePreview(tx, previewPath)
		if err != nil {
			tx.Rollback()
			return models.Entity{}, err
		}
		previewId = preview.Hash
	}

	createdEntity, err := repository.CreateEntity(
		tx,
		"",
		name,
		description,
		entityTypeId,
		parentId,
		previewId,
		isLibrary,
	)
	if err != nil {
		tx.Rollback()
		return models.Entity{}, err
	}
	err = tx.Commit()
	if err != nil {
		return models.Entity{}, err
	}
	return createdEntity, nil
}

// RenameCollection renames an existing collection.
// Returns the updated entity or an error if the operation fails.
func (e *CollectionService) RenameCollection(projectPath, entityId, newName string) (models.Entity, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Entity{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Entity{}, err
	}
	defer tx.Rollback()

	updatedEntity, err := repository.RenameEntity(tx, entityId, newName)
	if err != nil {
		tx.Rollback()
		return models.Entity{}, err
	}

	err = tx.Commit()
	if err != nil {
		return models.Entity{}, err
	}
	return updatedEntity, nil
}

// CreateCollections creates multiple collection entities in bulk.
// Currently a stub implementation for future batch creation functionality.
func (e *CollectionService) CreateCollections(projectPath, name, description, entityTypeId, parentId string) ([]models.Entity, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Entity{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Entity{}, err
	}
	defer tx.Rollback()

	return []models.Entity{}, nil
}

// DeleteCollection removes a collection from the project.
// Optionally removes associated files if removeFiles is true.
func (e *CollectionService) DeleteCollection(projectPath, entityId string, removeFiles bool) error {
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

	err = repository.DeleteEntity(tx, entityId, removeFiles, true)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// GetCollections retrieves collections based on user permissions.
// Returns all entities or only user-accessible entities based on role.
func (e *CollectionService) GetCollections(projectPath string) ([]models.Entity, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Entity{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Entity{}, err
	}
	defer tx.Rollback()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return []models.Entity{}, err
	}
	userData, err := repository.GetUser(tx, user.Id)
	if err != nil {
		return []models.Entity{}, err
	}
	userRole, err := repository.GetRole(tx, userData.RoleId)
	if err != nil {
		return []models.Entity{}, err
	}

	if userRole.ViewTask {
		entities, err := repository.GetEntities(tx, true)
		if err != nil {
			return []models.Entity{}, err
		}
		return entities, err
	} else {
		userTaskInfo, err := repository.GetUserTasksMinimal(tx, user.Id)
		if err != nil {
			return []models.Entity{}, err
		}

		entities, err := repository.GetUserEntities(tx, userTaskInfo, user.Id)
		if err != nil {
			return []models.Entity{}, err
		}
		return entities, err
	}
}

// GetCollectionChildren retrieves all children of a collection including tracked and untracked items.
// Returns separate lists for tasks, entities, and untracked items.
func (e *CollectionService) GetCollectionChildren(projectPath, entityId, projectWorkingDir, entityFolderPath string, ignoreList []string, isUntracked bool) (EntityItems, error) {
	children := EntityItems{
		Tasks:            make([]models.Task, 0),
		Entities:         make([]models.Entity, 0),
		UntrackedFiles:   make([]models.UntrackedTask, 0),
		UntrackedFolders: make([]models.UntrackedEntity, 0),
	}
	if entityId == "root" {
		entityId = ""
	}

	entityTrackFolders := []string{}
	entityTrackFiles := []string{}
	if !isUntracked {
		dbConn, err := utils.OpenDb(projectPath)
		if err != nil {
			return children, err
		}
		defer dbConn.Close()
		tx, err := dbConn.Beginx()
		if err != nil {
			return children, err
		}
		defer tx.Rollback()

		if entityId == "root" {
			entityId = ""
		}

		entities, err := repository.GetEntityChildren(tx, entityId)
		if err != nil {
			return children, err
		}
		children.Entities = entities

		tasks, err := repository.GetEntityTasks(tx, entityId)
		if err != nil {
			return children, err
		}
		children.Tasks = tasks

		for _, child := range entities {
			entityTrackFolders = append(entityTrackFolders, child.Name)
		}

		for _, child := range tasks {
			entityTrackFiles = append(entityTrackFiles, child.Name+child.Extension)
		}
	}

	if !utils.DirExists(entityFolderPath) {
		return children, nil
	}

	absoluteEntityFolderPath, err := filepath.Abs(entityFolderPath)
	if err != nil {
		return children, err
	}

	relativeEntityFolderPath, err := filepath.Rel(projectWorkingDir, absoluteEntityFolderPath)
	if err != nil {
		return children, err
	}
	relativeEntityFolderPath = utils.NormalizePath(relativeEntityFolderPath)

	clusttaIgnore := ignore.CompileIgnoreLines(ignoreList...)

	entries, err := os.ReadDir(absoluteEntityFolderPath)
	if err != nil {
		return children, err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(absoluteEntityFolderPath, entry.Name())
		relativePath := utils.NormalizePath(filepath.Join(relativeEntityFolderPath, entry.Name()))
		parentId := entityId
		if parentId == "root" {
			parentId = ""
		}
		if entry.IsDir() {
			if slices.Contains(entityTrackFolders, entry.Name()) {
				continue
			}
			if !clusttaIgnore.MatchesPath(relativePath) {
				children.UntrackedFolders = append(children.UntrackedFolders, models.UntrackedEntity{
					Id:         utils.GetMD5Hash(entryPath),
					Name:       entry.Name(),
					FilePath:   entryPath,
					EntityPath: "/" + relativePath + "/",
					ItemPath:   "/" + relativePath + "/",
					ParentId:   parentId,
				})
			}
		} else {
			if slices.Contains(entityTrackFiles, entry.Name()) {
				continue
			}

			if !clusttaIgnore.MatchesPath(relativePath) {
				taskName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
				children.UntrackedFiles = append(children.UntrackedFiles, models.UntrackedTask{
					Id:           utils.GetMD5Hash(entryPath),
					Name:         taskName,
					FilePath:     entryPath,
					TaskPath:     "/" + relativePath,
					EntityId:     parentId,
					EntityPath:   "/" + relativeEntityFolderPath + "/",
					Extension:    filepath.Ext(entry.Name()),
					ItemPath:     "/" + relativePath + "/",
					TaskTypeIcon: "generic",
				})
			}
		}
	}

	return children, nil
}

// GetCollectionTasks retrieves all tasks belonging to a specific collection.
// Returns the list of tasks or an error if the operation fails.
func (e *CollectionService) GetCollectionTasks(projectPath, entityId string) ([]models.Task, error) {
	if entityId == "root" {
		entityId = ""
	}
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Task{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Task{}, err
	}
	defer tx.Rollback()
	return repository.GetEntityTasks(tx, entityId)
}

// GetCollectionByID retrieves a collection by its ID.
// Returns the entity or an error if not found.
func (e *CollectionService) GetCollectionByID(projectPath, entityId string) (models.Entity, error) {
	if entityId == "root" {
		entityId = ""
	}
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Entity{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Entity{}, err
	}
	defer tx.Rollback()
	return repository.GetEntity(tx, entityId)
}

// GetCollectionByPath retrieves a collection by its filesystem path.
// Returns the entity or an error if not found.
func (e *CollectionService) GetCollectionByPath(projectPath, entityPath string) (models.Entity, error) {
	if entityPath == "/" {
		entityPath = ""
	}
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Entity{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Entity{}, err
	}
	defer tx.Rollback()
	return repository.GetEntityByPath(tx, entityPath)
}

// GetCollectionStateFlags checks if a collection has any recursive children with specific states.
// Returns flags indicating presence of untracked, modified, outdated, or rebuildable items.
func (e *CollectionService) GetCollectionStateFlags(projectPath, entityId, projectWorkingDir string, ignoreList []string) (CollectionStateFlags, error) {
	flags := CollectionStateFlags{
		HasUntracked:   false,
		HasModified:    false,
		HasOutdated:    false,
		HasRebuildable: false,
	}

	if entityId == "root" {
		entityId = ""
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return flags, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return flags, err
	}
	defer tx.Rollback()

	var entityPath string
	if entityId == "" {
		entityPath = ""
	} else {
		entity, err := repository.GetEntity(tx, entityId)
		if err != nil {
			return flags, err
		}
		entityPath = entity.EntityPath
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return flags, err
	}

	const batchSize = 100
	offset := 0

	type taskCheckpointInfo struct {
		taskId             string
		latestChecksum     string
		latestTimeModified int64
		latestFileSize     int64
		checkpointCount    int
		allChecksums       []string
	}

	type modifiedCandidate struct {
		taskId       string
		taskFilePath string
		checkpoints  []string
	}
	var candidatesNeedingHashCheck []modifiedCandidate

	for {
		if flags.HasRebuildable && flags.HasModified && flags.HasOutdated {
			break
		}

		var tasks []models.Task
		query := `
			SELECT id, task_path, extension, entity_path, name
			FROM full_task 
			WHERE entity_path LIKE ? AND trashed = 0 AND is_link = 0
			ORDER BY entity_path, name
			LIMIT ? OFFSET ?
		`
		err = tx.Select(&tasks, query, entityPath+"%", batchSize, offset)
		if err != nil {
			return flags, err
		}

		if len(tasks) == 0 {
			break
		}

		taskIds := make([]string, len(tasks))
		for i, task := range tasks {
			taskIds[i] = task.Id
		}

		checkpointMap := make(map[string]taskCheckpointInfo)
		if len(taskIds) > 0 {
			quotedTaskIds := make([]string, len(taskIds))
			for i, id := range taskIds {
				quotedTaskIds[i] = fmt.Sprintf("'%s'", id)
			}

			checkpointQuery := fmt.Sprintf(`
				SELECT task_id, xxhash_checksum, time_modified, file_size
				FROM task_checkpoint 
				WHERE task_id IN (%s) AND trashed = 0
				ORDER BY task_id, created_at DESC
			`, strings.Join(quotedTaskIds, ","))

			var checkpoints []struct {
				TaskId         string `db:"task_id"`
				XXHashChecksum string `db:"xxhash_checksum"`
				TimeModified   int64  `db:"time_modified"`
				FileSize       int64  `db:"file_size"`
			}
			tx.Select(&checkpoints, checkpointQuery)

			for _, cp := range checkpoints {
				if info, exists := checkpointMap[cp.TaskId]; exists {
					info.checkpointCount++
					info.allChecksums = append(info.allChecksums, cp.XXHashChecksum)
					checkpointMap[cp.TaskId] = info
				} else {
					checkpointMap[cp.TaskId] = taskCheckpointInfo{
						taskId:             cp.TaskId,
						latestChecksum:     cp.XXHashChecksum,
						latestTimeModified: cp.TimeModified,
						latestFileSize:     cp.FileSize,
						checkpointCount:    1,
						allChecksums:       []string{cp.XXHashChecksum},
					}
				}
			}
		}

		for _, task := range tasks {
			taskFilePath, err := utils.BuildTaskPath(rootFolder, task.EntityPath, task.Name, task.Extension)
			if err != nil {
				continue
			}

			fileInfo, err := os.Stat(taskFilePath)
			if os.IsNotExist(err) {
				if !flags.HasRebuildable {
					flags.HasRebuildable = true
				}
				continue
			}

			if err != nil {
				continue
			}

			if !flags.HasModified || !flags.HasOutdated {
				checkpointInfo, hasCheckpoint := checkpointMap[task.Id]

				if hasCheckpoint {
					fileSize := fileInfo.Size()

					if fileSize != checkpointInfo.latestFileSize {
						candidatesNeedingHashCheck = append(candidatesNeedingHashCheck, modifiedCandidate{
							taskId:       task.Id,
							taskFilePath: taskFilePath,
							checkpoints:  checkpointInfo.allChecksums,
						})
					} else {
						fileModTime := fileInfo.ModTime().Unix()

						if fileModTime != checkpointInfo.latestTimeModified {
							candidatesNeedingHashCheck = append(candidatesNeedingHashCheck, modifiedCandidate{
								taskId:       task.Id,
								taskFilePath: taskFilePath,
								checkpoints:  checkpointInfo.allChecksums,
							})
						}
					}
				}
			}

			if flags.HasRebuildable && flags.HasModified && flags.HasOutdated {
				break
			}
		}

		offset += batchSize
	}

	if (!flags.HasModified || !flags.HasOutdated) && len(candidatesNeedingHashCheck) > 0 {
		for _, candidate := range candidatesNeedingHashCheck {
			if flags.HasModified && flags.HasOutdated {
				break
			}

			fileHash, err := utils.GenerateXXHashChecksum(candidate.taskFilePath)
			if err != nil {
				continue
			}

			matchesLatest := (fileHash == candidate.checkpoints[0])
			matchesOlderCheckpoint := false

			if !matchesLatest && len(candidate.checkpoints) > 1 {
				for i := 1; i < len(candidate.checkpoints); i++ {
					if fileHash == candidate.checkpoints[i] {
						matchesOlderCheckpoint = true
						break
					}
				}
			}

			if matchesOlderCheckpoint {
				if !flags.HasOutdated {
					flags.HasOutdated = true
				}
			} else if !matchesLatest {
				if !flags.HasModified {
					flags.HasModified = true
				}
			}
		}
	}

	if !flags.HasUntracked && utils.DirExists(projectWorkingDir) {
		trackedFiles := make(map[string]bool)

		var allTasks []models.Task
		query := `
			SELECT task_path, extension
			FROM full_task 
			WHERE entity_path LIKE ? AND trashed = 0 AND is_link = 0
		`
		err = tx.Select(&allTasks, query, entityPath+"%")
		if err != nil {
			return flags, err
		}

		for _, task := range allTasks {
			taskFilePath, err := filepath.Abs(filepath.Join(projectWorkingDir, task.TaskPath+task.Extension))
			if err == nil {
				trackedFiles[taskFilePath] = true
			}
		}

		var folderToScan string
		if entityId == "" {
			folderToScan = projectWorkingDir
		} else {
			entity, err := repository.GetEntity(tx, entityId)
			if err != nil {
				return flags, err
			}
			folderToScan, err = utils.BuildEntityPath(rootFolder, entity.EntityPath)
			if err != nil {
				return flags, err
			}
		}

		if utils.DirExists(folderToScan) {
			clusttaIgnore := ignore.CompileIgnoreLines(ignoreList...)

			err = filepath.WalkDir(folderToScan, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.IsDir() {
					if strings.HasPrefix(filepath.Base(path), ".") {
						return filepath.SkipDir
					}
					return nil
				}

				if strings.HasPrefix(filepath.Base(path), ".") {
					return nil
				}

				absPath, err := filepath.Abs(path)
				if err != nil {
					return nil
				}

				if !trackedFiles[absPath] {
					relativePath, err := filepath.Rel(projectWorkingDir, path)
					if err != nil {
						return nil
					}
					relativePath = utils.NormalizePath(relativePath)

					if !clusttaIgnore.MatchesPath(relativePath) {
						flags.HasUntracked = true
						return filepath.SkipAll
					}
				}

				return nil
			})

			if err != nil && err != filepath.SkipAll {
				return flags, err
			}
		}
	}

	return flags, nil
}

// GetCollectionChildrenState analyzes the immediate children of a collection to determine their state.
// Returns state containing modified, outdated, rebuildable tasks and untracked items.
func (e *CollectionService) GetCollectionChildrenState(projectPath, entityId, projectWorkingDir string, ignoreList []string) (CollectionChildrenState, error) {
	state := CollectionChildrenState{
		ModifiedTasks:    make([]models.Task, 0),
		OutdatedTasks:    make([]models.Task, 0),
		RebuildableTasks: make([]models.Task, 0),
		NormalTasks:      make([]models.Task, 0),
		UntrackedFiles:   make([]models.UntrackedTask, 0),
		UntrackedFolders: make([]models.UntrackedEntity, 0),
	}

	if entityId == "root" {
		entityId = ""
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return state, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return state, err
	}
	defer tx.Rollback()

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return state, err
	}

	var tasks []models.Task
	query := `
		SELECT id, task_path, extension, entity_path, name, entity_id
		FROM full_task 
		WHERE entity_id = ? AND trashed = 0
		ORDER BY name
	`
	err = tx.Select(&tasks, query, entityId)
	if err != nil {
		return state, err
	}

	if len(tasks) == 0 {
		return e.detectUntrackedItems(tx, state, entityId, projectWorkingDir, rootFolder, ignoreList)
	}

	type fileMetadata struct {
		size    int64
		modTime int64
	}

	tasksMissingFiles := make([]string, 0)
	tasksWithFiles := make([]string, 0)
	taskFileMetadata := make(map[string]fileMetadata)
	taskMap := make(map[string]models.Task)

	for _, task := range tasks {
		taskMap[task.Id] = task

		taskFilePath, err := utils.BuildTaskPath(rootFolder, task.EntityPath, task.Name, task.Extension)
		if err != nil {
			continue
		}

		fileInfo, err := os.Stat(taskFilePath)
		if os.IsNotExist(err) {
			tasksMissingFiles = append(tasksMissingFiles, task.Id)
		} else if err == nil {
			tasksWithFiles = append(tasksWithFiles, task.Id)
			taskFileMetadata[task.Id] = fileMetadata{
				size:    fileInfo.Size(),
				modTime: fileInfo.ModTime().Unix(),
			}
			task.FilePath = taskFilePath
			taskMap[task.Id] = task
		}
	}

	if len(tasksMissingFiles) > 0 {
		quotedIds := make([]string, len(tasksMissingFiles))
		for i, id := range tasksMissingFiles {
			quotedIds[i] = fmt.Sprintf("'%s'", id)
		}

		rebuildableQuery := fmt.Sprintf(`
			SELECT DISTINCT task_id
			FROM task_checkpoint 
			WHERE task_id IN (%s) AND trashed = 0
		`, strings.Join(quotedIds, ","))

		var rebuildableTaskIds []struct {
			TaskId string `db:"task_id"`
		}
		err = tx.Select(&rebuildableTaskIds, rebuildableQuery)
		if err != nil {
			return state, err
		}

		for _, row := range rebuildableTaskIds {
			if task, exists := taskMap[row.TaskId]; exists {
				state.RebuildableTasks = append(state.RebuildableTasks, task)
			}
		}
	}

	type checkpointInfo struct {
		TaskId         string `db:"task_id"`
		XXHashChecksum string `db:"xxhash_checksum"`
		TimeModified   int64  `db:"time_modified"`
		FileSize       int64  `db:"file_size"`
	}

	taskCheckpoints := make(map[string][]checkpointInfo)

	if len(tasksWithFiles) > 0 {
		quotedIds := make([]string, len(tasksWithFiles))
		for i, id := range tasksWithFiles {
			quotedIds[i] = fmt.Sprintf("'%s'", id)
		}

		checkpointQuery := fmt.Sprintf(`
			SELECT task_id, xxhash_checksum, time_modified, file_size
			FROM task_checkpoint 
			WHERE task_id IN (%s) AND trashed = 0
			ORDER BY task_id, created_at DESC
		`, strings.Join(quotedIds, ","))

		var checkpoints []checkpointInfo
		err = tx.Select(&checkpoints, checkpointQuery)
		if err != nil {
			return state, err
		}

		for _, cp := range checkpoints {
			taskCheckpoints[cp.TaskId] = append(taskCheckpoints[cp.TaskId], cp)
		}
	}

	type hashCandidate struct {
		taskId       string
		taskFilePath string
		checkpoints  []checkpointInfo
	}
	candidatesNeedingHash := make([]hashCandidate, 0)
	tasksWithMatchingMetadata := make([]string, 0)

	for taskId, metadata := range taskFileMetadata {
		checkpoints, hasCheckpoints := taskCheckpoints[taskId]
		if !hasCheckpoints || len(checkpoints) == 0 {
			continue
		}

		latestCheckpoint := checkpoints[0]

		if metadata.size == latestCheckpoint.FileSize &&
			metadata.modTime == latestCheckpoint.TimeModified {
			tasksWithMatchingMetadata = append(tasksWithMatchingMetadata, taskId)
			continue
		}

		task := taskMap[taskId]
		candidatesNeedingHash = append(candidatesNeedingHash, hashCandidate{
			taskId:       taskId,
			taskFilePath: task.FilePath,
			checkpoints:  checkpoints,
		})
	}

	for _, candidate := range candidatesNeedingHash {
		fileHash, err := utils.GenerateXXHashChecksum(candidate.taskFilePath)
		if err != nil {
			continue
		}

		matchesLatest := (fileHash == candidate.checkpoints[0].XXHashChecksum)

		if matchesLatest {
			task := taskMap[candidate.taskId]
			state.NormalTasks = append(state.NormalTasks, task)
			continue
		}

		matchesOlderCheckpoint := false
		for i := 1; i < len(candidate.checkpoints); i++ {
			if fileHash == candidate.checkpoints[i].XXHashChecksum {
				matchesOlderCheckpoint = true
				break
			}
		}

		task := taskMap[candidate.taskId]

		if matchesOlderCheckpoint {
			state.OutdatedTasks = append(state.OutdatedTasks, task)
		} else {
			state.ModifiedTasks = append(state.ModifiedTasks, task)
		}
	}

	for _, taskId := range tasksWithMatchingMetadata {
		if task, exists := taskMap[taskId]; exists {
			state.NormalTasks = append(state.NormalTasks, task)
		}
	}

	return e.detectUntrackedItems(tx, state, entityId, projectWorkingDir, rootFolder, ignoreList)
}

// detectUntrackedItems scans the filesystem for untracked files and folders.
// Builds maps of tracked names and compares against filesystem entries.
func (e *CollectionService) detectUntrackedItems(tx *sqlx.Tx, state CollectionChildrenState, entityId, projectWorkingDir, rootFolder string, ignoreList []string) (CollectionChildrenState, error) {
	trackedTaskNames := make(map[string]bool)
	trackedEntityNames := make(map[string]bool)

	var trackedTasks []models.Task
	taskQuery := `
		SELECT name, extension
		FROM full_task 
		WHERE entity_id = ? AND trashed = 0
	`
	err := tx.Select(&trackedTasks, taskQuery, entityId)
	if err != nil {
		return state, err
	}

	for _, task := range trackedTasks {
		trackedTaskNames[task.Name+task.Extension] = true
	}

	var trackedEntities []models.Entity
	entityQuery := `
		SELECT name
		FROM full_entity 
		WHERE parent_id = ? AND trashed = 0
	`
	err = tx.Select(&trackedEntities, entityQuery, entityId)
	if err != nil {
		return state, err
	}

	for _, entity := range trackedEntities {
		trackedEntityNames[entity.Name] = true
	}

	var folderToScan string

	if entityId == "" {
		folderToScan = projectWorkingDir
	} else {
		entity, err := repository.GetEntity(tx, entityId)
		if err != nil {
			return state, err
		}
		folderToScan, err = utils.BuildEntityPath(rootFolder, entity.EntityPath)
		if err != nil {
			return state, err
		}
	}

	if !utils.DirExists(folderToScan) {
		return state, nil
	}

	absoluteEntityFolderPath, err := filepath.Abs(folderToScan)
	if err != nil {
		return state, err
	}

	relativeEntityFolderPath, err := filepath.Rel(projectWorkingDir, absoluteEntityFolderPath)
	if err != nil {
		return state, err
	}
	relativeEntityFolderPath = utils.NormalizePath(relativeEntityFolderPath)

	clusttaIgnore := ignore.CompileIgnoreLines(ignoreList...)

	entries, err := os.ReadDir(absoluteEntityFolderPath)
	if err != nil {
		return state, err
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		entryPath := filepath.Join(absoluteEntityFolderPath, entry.Name())
		relativePath := utils.NormalizePath(filepath.Join(relativeEntityFolderPath, entry.Name()))

		if entry.IsDir() {
			if trackedEntityNames[entry.Name()] {
				continue
			}

			if !clusttaIgnore.MatchesPath(relativePath) {
				state.UntrackedFolders = append(state.UntrackedFolders, models.UntrackedEntity{
					Id:         utils.GetMD5Hash(entryPath),
					Name:       entry.Name(),
					FilePath:   entryPath,
					EntityPath: "/" + relativePath + "/",
					ItemPath:   "/" + relativePath + "/",
					ParentId:   entityId,
				})
			}
		} else {
			if trackedTaskNames[entry.Name()] {
				continue
			}

			if !clusttaIgnore.MatchesPath(relativePath) {
				taskName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
				state.UntrackedFiles = append(state.UntrackedFiles, models.UntrackedTask{
					Id:           utils.GetMD5Hash(entryPath),
					Name:         taskName,
					FilePath:     entryPath,
					TaskPath:     "/" + relativePath,
					EntityId:     entityId,
					EntityPath:   "/" + relativeEntityFolderPath + "/",
					Extension:    filepath.Ext(entry.Name()),
					ItemPath:     "/" + relativePath + "/",
					TaskTypeIcon: "generic",
				})
			}
		}
	}

	return state, nil
}

// GetItemsForCheckpoint efficiently collects all modified and untracked items in a collection hierarchy.
// Returns deduplicated modified tasks and untracked files.
func (e *CollectionService) GetItemsForCheckpoint(projectPath, entityId, targetPath, projectWorkingDir string, ignoreList []string) (ItemsForCheckpoint, error) {
	result := ItemsForCheckpoint{
		ModifiedTasks:  make([]models.Task, 0),
		UntrackedFiles: make([]models.UntrackedTask, 0),
	}

	if entityId == "root" {
		entityId = ""
	}

	isTrackedCollection := entityId != ""
	isUntrackedPath := targetPath != "" && !isTrackedCollection

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return result, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return result, err
	}
	defer tx.Rollback()

	modifiedTasksMap := make(map[string]models.Task)
	untrackedFilesMap := make(map[string]models.UntrackedTask)

	if isTrackedCollection {
		err = e.processTrackedCollection(tx, entityId, projectPath, projectWorkingDir, ignoreList, modifiedTasksMap, untrackedFilesMap)
		if err != nil {
			return result, err
		}
	} else if isUntrackedPath {
		err = e.processUntrackedPath(targetPath, projectWorkingDir, ignoreList, untrackedFilesMap)
		if err != nil {
			return result, err
		}
	} else {
		err = e.processTrackedCollection(tx, "", projectPath, projectWorkingDir, ignoreList, modifiedTasksMap, untrackedFilesMap)
		if err != nil {
			return result, err
		}
	}

	for _, task := range modifiedTasksMap {
		result.ModifiedTasks = append(result.ModifiedTasks, task)
	}

	for _, file := range untrackedFilesMap {
		result.UntrackedFiles = append(result.UntrackedFiles, file)
	}

	return result, nil
}

// processTrackedCollection recursively scans tracked collections for modified tasks and untracked files.
// Uses flag-based optimization to avoid scanning clean collection branches.
func (e *CollectionService) processTrackedCollection(tx *sqlx.Tx, entityId, projectPath, projectWorkingDir string, ignoreList []string, modifiedTasksMap map[string]models.Task, untrackedFilesMap map[string]models.UntrackedTask) error {
	var processCollection func(string) error
	processCollection = func(currentEntityId string) error {
		childrenState, err := e.GetCollectionChildrenState(projectPath, currentEntityId, projectWorkingDir, ignoreList)
		if err != nil {
			return err
		}

		for _, task := range childrenState.ModifiedTasks {
			modifiedTasksMap[task.Id] = task
		}

		for _, file := range childrenState.UntrackedFiles {
			untrackedFilesMap[file.Id] = file
		}

		var childCollections []models.Entity
		childQuery := `
			SELECT id, name, entity_path
			FROM full_entity 
			WHERE parent_id = ? AND trashed = 0
			ORDER BY name
		`
		err = tx.Select(&childCollections, childQuery, currentEntityId)
		if err != nil {
			return err
		}

		for _, childCollection := range childCollections {
			flags, err := e.GetCollectionStateFlags(projectPath, childCollection.Id, projectWorkingDir, ignoreList)
			if err != nil {
				continue
			}

			if flags.HasModified || flags.HasUntracked {
				err = processCollection(childCollection.Id)
				if err != nil {
					continue
				}
			}
		}

		for _, untrackedFolder := range childrenState.UntrackedFolders {
			err = e.processUntrackedPath(untrackedFolder.FilePath, projectWorkingDir, ignoreList, untrackedFilesMap)
			if err != nil {
				continue
			}
		}

		return nil
	}

	return processCollection(entityId)
}

// processUntrackedPath recursively scans an untracked filesystem location for files.
// Performs pure filesystem scanning without database queries.
func (e *CollectionService) processUntrackedPath(targetPath, projectWorkingDir string, ignoreList []string, untrackedFilesMap map[string]models.UntrackedTask) error {
	absolutePath, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}

	if !utils.DirExists(absolutePath) {
		return fmt.Errorf("target path does not exist: %s", targetPath)
	}

	clusttaIgnore := ignore.CompileIgnoreLines(ignoreList...)

	var scanDirectory func(string) error
	scanDirectory = func(currentPath string) error {
		relativePath, err := filepath.Rel(projectWorkingDir, currentPath)
		if err != nil {
			return err
		}
		relativePath = utils.NormalizePath(relativePath)

		entries, err := os.ReadDir(currentPath)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			entryPath := filepath.Join(currentPath, entry.Name())
			relPath, err := filepath.Rel(projectWorkingDir, entryPath)
			if err != nil {
				continue
			}
			relPath = utils.NormalizePath(relPath)

			if clusttaIgnore.MatchesPath(relPath) {
				continue
			}

			if entry.IsDir() {
				err = scanDirectory(entryPath)
				if err != nil {
					continue
				}
			} else {
				taskName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
				untrackedFile := models.UntrackedTask{
					Id:           utils.GetMD5Hash(entryPath),
					Name:         taskName,
					FilePath:     entryPath,
					TaskPath:     "/" + relPath,
					EntityId:     "",
					EntityPath:   "/" + relativePath + "/",
					Extension:    filepath.Ext(entry.Name()),
					ItemPath:     "/" + relPath,
					TaskTypeIcon: "generic",
				}
				untrackedFilesMap[untrackedFile.Id] = untrackedFile
			}
		}

		return nil
	}

	return scanDirectory(absolutePath)
}

// GetOutdatedItemsInCollection efficiently collects all outdated items in a collection hierarchy.
// Returns deduplicated outdated tasks.
func (e *CollectionService) GetOutdatedItemsInCollection(projectPath, entityId, projectWorkingDir string, ignoreList []string) (ItemsForUpdate, error) {
	result := ItemsForUpdate{
		OutdatedTasks: make([]models.Task, 0),
	}

	if entityId == "root" {
		entityId = ""
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return result, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return result, err
	}
	defer tx.Rollback()

	outdatedTasksMap := make(map[string]models.Task)

	err = e.processTrackedCollectionForOutdated(tx, entityId, projectPath, projectWorkingDir, ignoreList, outdatedTasksMap)
	if err != nil {
		return result, err
	}

	for _, task := range outdatedTasksMap {
		result.OutdatedTasks = append(result.OutdatedTasks, task)
	}

	return result, nil
}

// processTrackedCollectionForOutdated recursively scans tracked collections for outdated tasks.
// Uses flag-based optimization to avoid scanning clean collection branches.
func (e *CollectionService) processTrackedCollectionForOutdated(tx *sqlx.Tx, entityId, projectPath, projectWorkingDir string, ignoreList []string, outdatedTasksMap map[string]models.Task) error {
	var processCollection func(string) error
	processCollection = func(currentEntityId string) error {
		childrenState, err := e.GetCollectionChildrenState(projectPath, currentEntityId, projectWorkingDir, ignoreList)
		if err != nil {
			return err
		}

		for _, task := range childrenState.OutdatedTasks {
			outdatedTasksMap[task.Id] = task
		}

		var childCollections []models.Entity
		childQuery := `
			SELECT id, name, entity_path
			FROM full_entity 
			WHERE parent_id = ? AND trashed = 0
			ORDER BY name
		`
		err = tx.Select(&childCollections, childQuery, currentEntityId)
		if err != nil {
			return err
		}

		for _, childCollection := range childCollections {
			flags, err := e.GetCollectionStateFlags(projectPath, childCollection.Id, projectWorkingDir, ignoreList)
			if err != nil {
				return err
			}

			if flags.HasOutdated {
				err = processCollection(childCollection.Id)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	return processCollection(entityId)
}

// Rebuild downloads missing checkpoints and rebuilds files for specified collections.
// Supports cancellation and sends progress updates via application events.
func (e *CollectionService) Rebuild(projectPath, remoteUrl, entityIds, userId string) error {
	defer reset()

	ctx := getContext()
	if ctx.Err() != nil {
		return errors.New("operation cancelled before starting")
	}

	app := application.Get()
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	errChan := make(chan error, 1)
	progressChan := make(chan output.ProgressReport, 10)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case progress, ok := <-progressChan:
				if !ok {
					return
				}
				app.Event.Emit("progress-update", progress)
			}
		}
	}()

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

	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      "Rebuilding",
		Message:    "Preparing to Rebuild",
		Percentage: 0,
		Current:    1,
		Total:      2,
	}:
	}

	var entityIdList []string
	if entityIds == "" {
		entityIdList = []string{""}
	} else {
		entityIdList = strings.Split(entityIds, ",")
		for i, id := range entityIdList {
			entityIdList[i] = strings.TrimSpace(id)
		}
	}

	entityEntitiesQuery := `
	SELECT full_entity.*
	FROM full_entity
	WHERE full_entity.entity_path LIKE ? OR full_entity.entity_path LIKE ?;
	`

	entities := []models.Entity{}
	allTasks := []models.Task{}

	for _, entityId := range entityIdList {
		if entityId == "" {
			rootEntities, err := repository.GetEntities(tx, false)
			if err != nil {
				return err
			}
			entities = append(entities, rootEntities...)

			rootTasks, err := repository.GetTasks(tx, false)
			if err != nil {
				return err
			}
			allTasks = append(allTasks, rootTasks...)
		} else {
			parentEntity, err := repository.GetEntity(tx, entityId)
			if err != nil {
				return err
			}
			err = os.MkdirAll(parentEntity.FilePath, os.ModePerm)
			if err != nil {
				return err
			}
			pathLike := parentEntity.EntityPath + "%"
			var entityChildren []models.Entity
			err = tx.Select(&entityChildren, entityEntitiesQuery, parentEntity.EntityPath, pathLike)
			if err != nil {
				return err
			}
			entities = append(entities, entityChildren...)

			entityTasksQuery := `
			SELECT full_task.*
			FROM full_task
			WHERE (full_task.entity_path LIKE ? OR full_task.entity_path LIKE ?) AND full_task.trashed = 0;
			`

			var entityTasks []models.Task
			err = tx.Select(&entityTasks, entityTasksQuery, parentEntity.EntityPath, pathLike)
			if err != nil {
				return err
			}
			allTasks = append(allTasks, entityTasks...)
		}
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return err
	}

	for _, entity := range entities {
		entityPath, err := utils.BuildEntityPath(rootFolder, entity.EntityPath)
		if err != nil {
			return err
		}
		err = os.MkdirAll(entityPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	taskIds := []string{}
	for _, task := range allTasks {
		taskIds = append(taskIds, task.Id)
	}

	if len(taskIds) > 0 {
		quotedTaskIds := make([]string, len(taskIds))
		for i, id := range taskIds {
			quotedTaskIds[i] = fmt.Sprintf("\"%s\"", id)
		}

		checkpoints := []models.Checkpoint{}
		err = tx.Select(&checkpoints, fmt.Sprintf("SELECT * FROM task_checkpoint WHERE trashed = 0 AND task_id IN (%s) ORDER BY created_at DESC", strings.Join(quotedTaskIds, ",")))
		if err != nil {
			return err
		}

		taskCheckpoints := map[string][]models.Checkpoint{}
		for _, taskCheckpoint := range checkpoints {
			taskCheckpoints[taskCheckpoint.TaskId] = append(taskCheckpoints[taskCheckpoint.TaskId], taskCheckpoint)
		}

		for i, task := range allTasks {
			allTasks[i].Checkpoints = taskCheckpoints[task.Id]
		}
	}

	tasksToRebuild := []models.Task{}
	for _, task := range allTasks {
		taskFilePath, err := utils.BuildTaskPath(rootFolder, task.EntityPath, task.Name, task.Extension)
		if err != nil {
			return err
		}
		task.FilePath = taskFilePath
		if _, err := os.Stat(task.GetFilePath()); os.IsNotExist(err) {
			tasksToRebuild = append(tasksToRebuild, task)
		}
	}

	checkpointIdsToDownload := []string{}
	for _, task := range tasksToRebuild {
		latestCheckpoint := task.Checkpoints[0]
		isMisssingChunks, err := latestCheckpoint.HasMissingChunks(tx)
		if err != nil {
			return err
		}
		if isMisssingChunks {
			checkpointIdsToDownload = append(checkpointIdsToDownload, latestCheckpoint.Id)
		}
	}

	err = tx.Rollback()
	if err != nil {
		return err
	}

	if len(checkpointIdsToDownload) != 0 {
		callBack := func(current int, total int, message string, extraMessage string) {
			if ctx.Err() != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			case progressChan <- output.ProgressReport{
				Title:        "Downloading files",
				Message:      message,
				Percentage:   (float64(current) / float64(total) * 99),
				Current:      1,
				Total:        1,
				ExtraMessage: extraMessage,
			}:
			default:
			}
		}

		go func() {
			err := sync_service.DownloadCheckpoints(ctx, projectPath, remoteUrl, checkpointIdsToDownload, user.Id, callBack)
			if ctx.Err() == nil {
				errChan <- err
			}
		}()

		select {
		case err = <-errChan:
			if err != nil {
				if errors.Is(err, syscall.ECONNREFUSED) {
					return errors.New("download failed, connection refused")
				}
				return errors.New("download failed, check your connection")
			}
		case <-ctx.Done():
			close(progressChan) // Stop progress updates
			return errors.New("cancelled")
		}
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	tx, err = dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	totalItems := len(tasksToRebuild)
	for i, task := range tasksToRebuild {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Rebuilding Files",
				Message:    task.Name,
				Percentage: float64(current) / float64(total) * 100,
				Current:    i + 1,
				Total:      totalItems,
			}
			app.Event.Emit("progress-update", progress)
		}
		err = repository.RevertToLatestCheckpoint(tx, task.Id, task.FilePath, callBack)
		if err != nil {
			return err
		}
	}

	close(progressChan)
	progress := output.ProgressReport{
		Title:      "Downloading Checkpoint",
		Message:    "Receiving",
		Percentage: 100,
		Current:    1,
		Total:      1,
	}
	app.Event.Emit("progress-update", progress)
	return nil
}

// RevealCollection opens the file explorer to show a collection's folder.
// Returns an error if the entity is not found or the operation fails.
func (e *CollectionService) RevealCollection(projectPath, entityId string) error {
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

	entity, err := repository.GetEntity(tx, entityId)
	if err != nil {
		return err
	}
	utils.RevealInExplorer(entity.GetFilePath())
	return nil
}

// RevertCollections reverts multiple collections to their latest checkpoints.
// Sends progress updates for each entity processed.
func (e *CollectionService) RevertCollections(projectPath string, entityIds []string) error {
	app := application.Get()
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

	totalEntities := len(entityIds)
	for i, entityId := range entityIds {
		tx, err := dbConn.Beginx()
		if err != nil {
			return err
		}
		entity, err := repository.GetEntity(tx, entityId)
		if err != nil {
			return err
		}
		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Reverting",
				Message:    entity.EntityTypeName,
				Percentage: float64(current) / float64(total) * 100,
				Current:    i + 1,
				Total:      totalEntities,
			}
			app.Event.Emit("progress-update", progress)
		}

		err = repository.RevertToLatestCheckpoint(tx, entityId, entity.FilePath, callBack)
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
	}
	return nil
}

// ChangeCollectionParent moves one or more collections to a different parent collection.
// Checks for name conflicts in the target parent before moving.
// Returns an error if any collection would conflict or if the operation fails.
func (e *CollectionService) ChangeCollectionParent(projectPath string, entityIds []string, parentId string) error {
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

	var conflicts []string
	for _, entityId := range entityIds {
		entity, err := repository.GetEntity(tx, entityId)
		if err != nil {
			return err
		}
		if entity.ParentId == parentId {
			continue
		}
		_, err = repository.GetEntityByName(tx, entity.Name, parentId)
		if err == nil {
			conflicts = append(conflicts, entity.Name)
		} else if err != error_service.ErrEntityNotFound {
			// Some other error occurred
			return err
		}
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("collections with the same name already exist in the target location: %s", strings.Join(conflicts, ", "))
	}

	for _, entityId := range entityIds {
		err = repository.ChangeParent(tx, entityId, parentId)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// ChangeType changes the type of a collection.
// Returns an error if the operation fails.
func (e *CollectionService) ChangeType(projectPath, entityId, entityTypeId string) error {
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

	err = repository.ChangeEntityType(tx, entityId, entityTypeId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// ChangeIsLibrary toggles the library flag on a collection.
// Returns an error if the operation fails.
func (e *CollectionService) ChangeIsLibrary(projectPath, entityId string, isLibrary bool) error {
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

	err = repository.ChangeIsLibrary(tx, entityId, isLibrary)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Assign assigns a user to a collection.
// Returns an error if the operation fails.
func (e *CollectionService) Assign(projectPath, entityId, userId string) error {
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

	err = repository.AssignEntity(tx, entityId, userId)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Unassign removes a user assignment from a collection.
// Returns an error if the operation fails.
func (e *CollectionService) Unassign(projectPath, entityId, userId string) error {
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

	err = repository.UnAssignEntity(tx, entityId, userId)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// CreateCollectionType creates a new collection type in the project.
// Returns an error if a type with the same name already exists.
func (e *CollectionService) CreateCollectionType(projectPath, entityTypeName, entityTypeIcon string) (models.EntityType, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.EntityType{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.EntityType{}, err
	}
	defer tx.Rollback()

	entityType, err := repository.CreateEntityType(tx, "", entityTypeName, entityTypeIcon)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: entity_type.name" {
			tx.Rollback()
			return models.EntityType{}, error_service.ErrEntityTypeExists
		}
		tx.Rollback()
		return models.EntityType{}, err
	}
	tx.Commit()
	return entityType, nil
}

// UpdateCollectionType updates an existing collection type.
// Returns an error if a type with the new name already exists.
func (e *CollectionService) UpdateCollectionType(projectPath, id, entityTypeName, entityTypeIcon string) (models.EntityType, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.EntityType{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.EntityType{}, err
	}
	defer tx.Rollback()

	entityType, err := repository.UpdateEntityType(tx, id, entityTypeName, entityTypeIcon)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: entity_type.name" {
			tx.Rollback()
			return models.EntityType{}, error_service.ErrEntityTypeExists
		}
		tx.Rollback()
		return models.EntityType{}, err
	}
	err = tx.Commit()
	if err != nil {
		return models.EntityType{}, err
	}
	return entityType, nil
}

// DeleteCollectionType removes a collection type from the project.
// Returns an error if the operation fails.
func (e *CollectionService) DeleteCollectionType(projectPath, id string) error {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = repository.DeleteEntityType(tx, id)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// GetCollectionTypes retrieves all collection types in the project.
// Returns the list of entity types or an error if the operation fails.
func (e *CollectionService) GetCollectionTypes(projectPath string) ([]models.EntityType, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.EntityType{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.EntityType{}, err
	}
	defer tx.Rollback()

	entityTypes, err := repository.GetEntityTypes(tx)
	if err != nil {
		return entityTypes, err
	}
	return entityTypes, nil
}

// IsUserAssignedToCollectionOrAncestor checks if a user is assigned to a collection
// or any of its parent collections recursively.
func (e *CollectionService) IsUserAssignedToCollectionOrAncestor(projectPath, entityId, userId string) (bool, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return false, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	return repository.IsUserAssignedToEntityOrAncestor(tx, entityId, userId)
}

// UpdatePreview updates the preview image for a collection.
// Returns an error if the project is not found or the operation fails.
func (p *CollectionService) UpdatePreview(projectPath, entityId, previewPath string) error {
	if !utils.FileExists(projectPath) {
		return error_service.ErrProjectNotFound
	}

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

	err = repository.UpdateEntityPreview(tx, entityId, previewPath)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
