package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/ignore"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"clustta/output"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type ImportService struct{}

type ImportItems struct {
	Tasks    []models.Task   `json:"tasks"`
	Entities []models.Entity `json:"entities"`
}

func getItemPath(root, itemFilepath string, isFile bool) string {
	// Ensure paths are using consistent separators
	root = filepath.ToSlash(root)
	itemFilepath = filepath.ToSlash(itemFilepath)

	// Make sure root ends with a slash
	if !strings.HasSuffix(root, "/") {
		root = root + "/"
	}

	// Get the relative path by removing the root prefix
	relPath := strings.TrimPrefix(itemFilepath, root)

	// If it's a file (has an extension), remove the extension
	if isFile {
		ext := filepath.Ext(relPath)
		if ext != "" {
			relPath = strings.TrimSuffix(relPath, ext)
		}
	}

	relPath = strings.TrimSuffix(relPath, "/")
	return relPath
}

func (i *ImportService) ImportFolder(projectPath, parentId string, folders, files []string, projectWorkingDir string, ignoreList []string) (ImportItems, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ImportItems{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ImportItems{}, err
	}
	defer tx.Rollback()

	ignoreObject := ignore.CompileIgnoreLines(ignoreList...)

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return ImportItems{}, err
	}

	if len(folders)+len(files) == 0 {
		return ImportItems{}, errors.New("missing or invalid parameter: folders/files")
	}

	var tasks []string
	var dirs []string
	entitiesMap := map[string]models.Entity{}

	for _, file := range files {
		file, err := filepath.Abs(file)
		if err != nil {
			return ImportItems{}, err
		}
		file = filepath.ToSlash(file)

		relativePath, err := filepath.Rel(projectWorkingDir, file)
		if err != nil {
			return ImportItems{}, err
		}
		if ignoreObject.MatchesPath(relativePath) {
			continue
		}

		tasks = append(tasks, file)
	}

	for _, folder := range folders {
		folder = filepath.ToSlash(folder)
		relativePath, err := filepath.Rel(projectWorkingDir, folder)
		if err != nil {
			return ImportItems{}, err
		}
		if ignoreObject.MatchesPath(relativePath) {
			continue
		}

		rootAbs, err := filepath.Abs(folder)
		if err != nil {
			return ImportItems{}, err
		}
		rootAbs = filepath.ToSlash(rootAbs)
		dirs = append(dirs, rootAbs)

		err = filepath.WalkDir(rootAbs, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			path = filepath.ToSlash(path)

			if path == rootAbs {
				return nil
			}

			// Get the parent path of the relative path
			if d.IsDir() {
				if strings.HasPrefix(filepath.Base(path), ".") {
					return filepath.SkipDir
				}

				folder := filepath.ToSlash(path)

				relativePath, err := filepath.Rel(projectWorkingDir, folder)
				if err != nil {
					return err
				}
				if ignoreObject.MatchesPath(relativePath) {
					return nil
				}

				dirs = append(dirs, path)
			} else {
				if strings.HasPrefix(filepath.Base(path), ".") {
					return nil
				}
				file := filepath.ToSlash(path)

				relativePath, err := filepath.Rel(projectWorkingDir, file)
				if err != nil {
					return err
				}
				if ignoreObject.MatchesPath(relativePath) {
					return nil
				}

				tasks = append(tasks, path)
			}
			return nil
		})
		if err != nil {
			return ImportItems{}, nil
		}
	}

	entitiesData := []models.Entity{}
	tasksData := []models.Task{}
	for _, dir := range dirs {
		parentPath := filepath.ToSlash(filepath.Dir(dir))
		dirName := filepath.Base(dir)
		entityParentId := parentId
		if utils.Contains(dirs, parentPath) {
			entityParentId = entitiesMap[parentPath].Id
			if entityParentId == "" {
				return ImportItems{}, errors.New("parent not found in map")
			}
		}
		ruleEntityType, err := repository.GetEntityTypeByName(tx, "Generic")
		if err != nil {
			return ImportItems{}, err
		}
		entityPath := getItemPath(rootFolder, dir, false)
		entity := models.Entity{
			Id:             uuid.New().String(),
			Name:           dirName,
			ParentId:       entityParentId,
			EntityTypeId:   ruleEntityType.Id,
			EntityTypeIcon: ruleEntityType.Icon,
			EntityTypeName: ruleEntityType.Name,
			FilePath:       dir,
			EntityPath:     entityPath,
		}
		entitiesMap[dir] = entity
		entitiesData = append(entitiesData, entity)
	}

	for _, task := range tasks {
		parentPath := filepath.ToSlash(filepath.Dir(task))
		taskName := strings.TrimSuffix(filepath.Base(task), filepath.Ext(task))

		entityParentId := parentId
		if utils.Contains(dirs, parentPath) {
			entityParentId = entitiesMap[parentPath].Id
			if entityParentId == "" {
				return ImportItems{}, errors.New("parent not found in map")
			}
		}

		ruleTaskType, err := repository.GetTaskTypeByName(tx, "Generic")
		if err != nil {
			return ImportItems{}, err
		}
		taskPath := getItemPath(rootFolder, task, true)
		taskData := models.Task{
			Id:           uuid.New().String(),
			Name:         taskName,
			TaskTypeId:   ruleTaskType.Id,
			TaskTypeName: ruleTaskType.Name,
			TaskTypeIcon: ruleTaskType.Icon,
			EntityId:     entityParentId,
			FilePath:     task,
			IsResource:   true,
			TaskPath:     taskPath,
		}
		tasksData = append(tasksData, taskData)
	}

	importItems := ImportItems{
		Tasks:    tasksData,
		Entities: entitiesData,
	}
	return importItems, nil
}

func (e *ImportService) CreateItems(projectPath string, entities []models.Entity, tasks []models.Task, comment, groupId string) error {
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

	if len(entities)+len(tasks) == 0 {
		return errors.New("missing or invalid parameter: no item passed")
	}

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	progress := output.ProgressReport{
		Title:      "Starting Import",
		Message:    "Starting Import",
		Percentage: 0,
		Current:    1,
		Total:      3,
	}
	app.EmitEvent("progress-update", progress)

	totalTasks := len(tasks)
	totalEntities := len(entities)
	// totalItems := totalDirs + totalFiles

	for i, entity := range entities {
		_, err = repository.CreateEntity(
			tx, entity.Id, entity.Name, entity.Description, entity.EntityTypeId, entity.ParentId, entity.PreviewId, entity.IsLibrary)
		if err != nil {
			return err
		}

		progress := output.ProgressReport{
			Title:      "Creating Entities",
			Message:    entity.Name,
			Percentage: float64(i+1) / float64(totalEntities) * 99,
			Current:    i + 1,
			Total:      totalEntities,
		}
		app.EmitEvent("progress-update", progress)
	}

	for i, task := range tasks {
		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Creating Tasks",
				Message:    task.Name,
				Percentage: float64(current) / float64(total) * 99,
				Current:    i + 1,
				Total:      totalTasks,
			}
			app.EmitEvent("progress-update", progress)
		}
		_, err = repository.CreateTask(
			tx, task.Id, task.Name, task.TaskTypeId, task.EntityId, task.IsResource,
			"", task.Description, task.FilePath, task.Tags,
			task.Pointer, task.IsLink, task.PreviewId, user.Id, comment, groupId, callBack)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	progress = output.ProgressReport{
		Title:      "Finishing Import",
		Message:    "Finishing Import",
		Percentage: 100,
		Current:    2,
		Total:      2,
	}
	app.EmitEvent("progress-update", progress)

	return nil
}

func (e *ImportService) CreateEntities(projectPath string, entities []models.Entity, completed, totalEntities int) error {
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

	if len(entities) == 0 {
		return errors.New("missing or invalid parameter: no item passed")
	}

	progress := output.ProgressReport{
		Title:      "Starting Import",
		Message:    "Starting Import",
		Percentage: 0,
		Current:    1,
		Total:      2,
	}
	app.EmitEvent("progress-update", progress)

	state := completed
	for i, entity := range entities {
		_, err = repository.CreateEntity(
			tx, entity.Id, entity.Name, entity.Description, entity.EntityTypeId, entity.ParentId, entity.PreviewId, entity.IsLibrary)
		if err != nil {
			return err
		}

		progress := output.ProgressReport{
			Title:      "Creating Entities",
			Message:    entity.Name,
			Percentage: float64(i+1) / float64(totalEntities) * 99,
			Current:    completed + (i + 1),
			Total:      totalEntities,
		}
		app.EmitEvent("progress-update", progress)
		state = completed + (i + 1)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	if state == totalEntities {
		progress = output.ProgressReport{
			Title:      "Finishing Import",
			Message:    "Finishing Import",
			Percentage: 100,
			Current:    2,
			Total:      2,
		}
		app.EmitEvent("progress-update", progress)
	}

	return nil
}

func (e *ImportService) CreateTasks(projectPath string, tasks []models.Task, completed, totalTasks int, comment, groupId string) error {
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

	if len(tasks) == 0 {
		return errors.New("missing or invalid parameter: no item passed")
	}

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	progress := output.ProgressReport{
		Title:      "Starting Import",
		Message:    "Starting Import",
		Percentage: 0,
		Current:    1,
		Total:      3,
	}
	app.EmitEvent("progress-update", progress)
	state := completed
	for i, task := range tasks {
		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Creating Tasks",
				Message:    task.Name,
				Percentage: float64(current) / float64(total) * 99,
				Current:    completed + (i + 1),
				Total:      totalTasks,
			}
			app.EmitEvent("progress-update", progress)
		}
		_, err = repository.CreateTask(
			tx, task.Id, task.Name, task.TaskTypeId, task.EntityId, task.IsResource,
			"", task.Description, task.FilePath, task.Tags,
			task.Pointer, task.IsLink, task.PreviewId, user.Id, comment, groupId, callBack)
		if err != nil {
			return err
		}
		state = completed + (i + 1)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	if state == totalTasks {
		progress = output.ProgressReport{
			Title:      "Finishing Import",
			Message:    "Finishing Import",
			Percentage: 100,
			Current:    2,
			Total:      2,
		}
		app.EmitEvent("progress-update", progress)
	}

	return nil
}
