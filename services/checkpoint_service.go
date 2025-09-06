package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/sync_service"
	"clustta/internal/utils"
	"clustta/output"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type CheckpointService struct{}

func (c *CheckpointService) DeleteCheckpoint(projectPath, checkpointId string) error {
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

	err = repository.DeleteCheckpoint(
		tx,
		checkpointId,
		true,
		true,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (c *CheckpointService) RevertToCheckpoint(projectPath, remoteUrl, taskId, checkpointId string) error {
	defer reset() // Ensure context is reset when we're done

	ctx := getContext()
	if ctx.Err() != nil {
		return errors.New("operation cancelled before starting")
	}

	app := application.Get()
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	// Create buffered channels to prevent blocking
	errChan := make(chan error, 1)
	progressChan := make(chan output.ProgressReport, 10)

	// Start progress update goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return // Exit immediately on cancellation
			case progress, ok := <-progressChan:
				if !ok {
					return
				}
				app.EmitEvent("progress-update", progress)
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

	// Initial progress
	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      "Reverting Task",
		Message:    "Preparing to Revert",
		Percentage: 0,
		Current:    1,
		Total:      2,
	}:
	}

	task, err := repository.GetTask(tx, taskId)
	if err != nil {
		return err
	}

	checkpoint, err := repository.GetCheckpoint(tx, checkpointId)
	if err != nil {
		return err
	}
	isMisssingChunks, err := checkpoint.HasMissingChunks(tx)
	if err != nil {
		return err
	}

	err = tx.Rollback()
	if err != nil {
		return err
	}

	if isMisssingChunks {
		callBack := func(current int, total int, message string, extraMessage string) {
			if ctx.Err() != nil {
				return
			}
			currentSize := utils.BytesToHumanReadable(current)
			totalSize := utils.BytesToHumanReadable(total)
			select {
			case <-ctx.Done():
				return
			case progressChan <- output.ProgressReport{
				Title:      "Downloading Checkpoint",
				Message:    fmt.Sprintf("Pulling Data %s/%s", currentSize, totalSize),
				Percentage: (float64(current) / float64(total) * 99),
				Current:    1,
				Total:      1,
			}:
			default: // Skip progress update if channel is full
			}
		}

		go func() {
			err := sync_service.DownloadCheckpoint(ctx, projectPath, remoteUrl, checkpointId, user.Id, callBack)
			if ctx.Err() == nil { // Only send error if not cancelled
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

	callBack := func(current int, total int, message string, extraMessage string) {
		progress := output.ProgressReport{
			Title:      "Reverting Task",
			Message:    task.Name,
			Percentage: float64(current) / float64(total) * 100,
			Current:    1,
			Total:      1,
		}
		app.EmitEvent("progress-update", progress)
	}
	err = repository.RevertToCheckpoint(tx, checkpointId, task.FilePath, callBack)
	if err != nil {
		return err
	}

	close(progressChan)
	progress := output.ProgressReport{
		Title:      "Reverting Task",
		Message:    task.Name,
		Percentage: 100,
		Current:    1,
		Total:      1,
	}
	app.EmitEvent("progress-update", progress)
	return nil
}

func (c *CheckpointService) AddCheckpoint(projectPath string, taskPaths []string, message, previewPath, groupId string, useAsThumbnail bool) ([]models.Checkpoint, error) {
	app := application.Get()
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Checkpoint{}, err
	}
	defer dbConn.Close()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return []models.Checkpoint{}, err
	}
	authorId := user.Id

	totalTasks := len(taskPaths)
	previewId := ""
	if previewPath != "" {
		tx, err := dbConn.Beginx()
		if err != nil {
			return []models.Checkpoint{}, err
		}

		preview, err := repository.CreatePreview(tx, previewPath)
		if err != nil {
			tx.Rollback()
			return []models.Checkpoint{}, err
		}
		previewId = preview.Hash
		err = tx.Commit()
		if err != nil {
			return []models.Checkpoint{}, err
		}
	}
	checkpoints := []models.Checkpoint{}
	for i, taskPath := range taskPaths {
		tx, err := dbConn.Beginx()
		if err != nil {
			return []models.Checkpoint{}, err
		}

		task, err := repository.GetTaskByPath(tx, taskPath)
		if err != nil {
			return []models.Checkpoint{}, err
		}
		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Creating Check Point",
				Message:    task.Name,
				Percentage: float64(current) / float64(total) * 99,
				Current:    i + 1,
				Total:      totalTasks,
			}
			app.EmitEvent("progress-update", progress)
		}

		checkpoint, err := repository.CreateCheckpoint(
			tx,
			task.Id,
			message,
			"",
			"",
			0,
			0,
			task.FilePath,
			authorId,
			previewId,
			groupId,
			callBack,
		)
		if err != nil {
			tx.Rollback()
			return []models.Checkpoint{}, err
		}
		if previewPath != "" && useAsThumbnail {
			err = repository.SetEntityPreview(tx, task.Id, "task", previewId)
			if err != nil {
				tx.Rollback()
				return []models.Checkpoint{}, err
			}
		}
		checkpoints = append(checkpoints, checkpoint)
		err = tx.Commit()
		if err != nil {
			return []models.Checkpoint{}, err
		}
	}

	progress := output.ProgressReport{
		Title:      "Creating Check Point",
		Message:    "finishing up",
		Percentage: 100,
		Current:    totalTasks,
		Total:      totalTasks,
		EntityData: checkpoints,
	}
	app.EmitEvent("progress-update", progress)
	return checkpoints, nil
}

func (c *CheckpointService) AddUntrackedTask(projectPath, projectWorkingDir string, taskPaths []string, completed, totalTasks int, message, previewPath, groupId string) ([]models.Task, error) {
	app := application.Get()
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

	entities := []models.Entity{}
	err = tx.Select(&entities, "SELECT id, entity_path FROM full_entity")
	if err != nil {
		return []models.Task{}, err
	}
	entityType, err := repository.GetEntityTypeByName(tx, "generic")
	if err != nil {
		return []models.Task{}, err
	}
	taskType, err := repository.GetTaskTypeByName(tx, "generic")
	if err != nil {
		return []models.Task{}, err
	}
	status, err := repository.GetStatusByShortName(tx, "todo")
	if err != nil {
		return []models.Task{}, err
	}
	statusId := status.Id
	tx.Rollback()

	entityPathsIndex := map[string]string{}
	for _, entity := range entities {
		entityPathsIndex[entity.EntityPath] = entity.Id
	}

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return []models.Task{}, err
	}

	previewId := ""
	if previewPath != "" {
		tx, err := dbConn.Beginx()
		if err != nil {
			return []models.Task{}, err
		}

		preview, err := repository.CreatePreview(tx, previewPath)
		if err != nil {
			tx.Rollback()
			return []models.Task{}, err
		}
		previewId = preview.Hash
		err = tx.Commit()
		if err != nil {
			return []models.Task{}, err
		}
	}
	tasks := []models.Task{}
	state := completed
	for i, taskPath := range taskPaths {
		tx, err := dbConn.Beginx()
		if err != nil {
			return []models.Task{}, err
		}
		defer tx.Rollback()

		taskFilePath := filepath.Join(projectWorkingDir, taskPath)
		entityPath := utils.GetParent(taskPath)
		taskEntityId := ""

		for _, curentEntityPath := range utils.GetEntityPaths(entityPath) {
			entityId, exists := entityPathsIndex[curentEntityPath]
			if !exists {
				entityName := filepath.Base(curentEntityPath)
				parentEntityId := entityPathsIndex[utils.GetParent(curentEntityPath)]
				entityId = uuid.New().String()
				err = repository.CreateEntityFast(tx, entityId, entityName, "", entityType.Id, parentEntityId, "", false)
				if err != nil {
					return []models.Task{}, err
				}
				entityPathsIndex[curentEntityPath] = entityId
			}
			taskEntityId = entityId
		}

		taskName := strings.TrimSuffix(filepath.Base(taskPath), filepath.Ext(taskPath))

		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Creating Check Point",
				Message:    taskName,
				Percentage: float64(current) / float64(total) * 99,
				Current:    completed + (i + 1),
				Total:      totalTasks,
			}
			app.EmitEvent("progress-update", progress)
		}
		err = repository.CreateTaskFast(tx, "", taskName, taskType.Id, taskEntityId, true, "", taskFilePath, previewId, user.Id, message, groupId, taskPath, statusId, callBack)
		if err != nil {
			tx.Rollback()
			return []models.Task{}, err
		}

		err = tx.Commit()
		if err != nil {
			return []models.Task{}, err
		}
		state = completed + (i + 1)
	}

	if state == totalTasks {
		progress := output.ProgressReport{
			Title:      "Creating Check Point",
			Message:    "finishing up",
			Percentage: 100,
			Current:    totalTasks,
			Total:      totalTasks,
		}
		app.EmitEvent("progress-update", progress)
	}
	return tasks, nil
}

func (c *CheckpointService) ViewCheckpoint(projectPath, checkpointId, entityName, extension string) error {
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

	f, err := os.CreateTemp("", "ClusttaTmpFile-")
	// For older Go checkpoints, you can use ioutil.TempFile
	if err != nil {
		output.Error(err.Error())
	}
	defer f.Close()
	defer os.Remove(f.Name())

	tempFile := f.Name() + extension
	callBack := func(current int, total int, message string, extraMessage string) {
		progress := output.ProgressReport{
			Title:      "Preparing Check Point",
			Message:    entityName,
			Percentage: float64(current) / float64(total) * 100,
			Current:    1,
			Total:      1,
		}
		app.EmitEvent("progress-update", progress)
	}
	err = repository.RevertToCheckpoint(tx, checkpointId, tempFile, callBack)
	if err != nil {
		return err
	}

	err = utils.LaunchFile(tempFile)
	if err != nil {
		return err
	}
	return nil
}

func (c *CheckpointService) GetCheckpoints(projectPath, taskId string) ([]models.Checkpoint, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Checkpoint{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Checkpoint{}, err
	}
	defer tx.Rollback()

	checkPoints, err := repository.GetCheckpoints(tx, taskId, false)
	if err != nil {
		return []models.Checkpoint{}, err
	}
	return checkPoints, nil
}
func (c *CheckpointService) GetTimeline(projectPath string) ([]repository.CompatTimeline, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []repository.CompatTimeline{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []repository.CompatTimeline{}, err
	}
	defer tx.Rollback()

	timeline, err := repository.GetTimeline(tx)
	if err != nil {
		return []repository.CompatTimeline{}, err
	}
	return timeline, nil
}

func (c *CheckpointService) Revert(projectPath, remoteUrl string, taskIds []string) error {
	defer reset() // Ensure context is reset when we're done

	ctx := getContext()
	if ctx.Err() != nil {
		return errors.New("operation cancelled before starting")
	}

	app := application.Get()
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	// Create buffered channels to prevent blocking
	errChan := make(chan error, 1)
	progressChan := make(chan output.ProgressReport, 10)

	// Start progress update goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return // Exit immediately on cancellation
			case progress, ok := <-progressChan:
				if !ok {
					return
				}
				app.EmitEvent("progress-update", progress)
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

	// Initial progress
	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      "Reverting",
		Message:    "Preparing to Revert",
		Percentage: 0,
		Current:    1,
		Total:      1,
	}:
	}

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

	checkpointIdsToDownload := []string{}
	for _, taskId := range taskIds {
		latestCheckpoint := taskCheckpoints[taskId][0]
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
			default: // Skip progress update if channel is full
			}
		}

		go func() {
			err := sync_service.DownloadCheckpoints(ctx, projectPath, remoteUrl, checkpointIdsToDownload, user.Id, callBack)
			if ctx.Err() == nil { // Only send error if not cancelled
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

	totalTasks := len(taskIds)
	for i, taskId := range taskIds {
		tx, err := dbConn.Beginx()
		if err != nil {
			return err
		}
		task, err := repository.GetTask(tx, taskId)
		if err != nil {
			return err
		}
		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Reverting",
				Message:    task.Name,
				Percentage: float64(current) / float64(total) * 100,
				Current:    i + 1,
				Total:      totalTasks,
			}
			app.EmitEvent("progress-update", progress)
		}

		err = repository.RevertToLatestCheckpoint(tx, taskId, task.FilePath, callBack)
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Rollback()
	}

	close(progressChan)
	progress := output.ProgressReport{
		Title:      "Reverting",
		Message:    "Reverting",
		Percentage: 100,
		Current:    1,
		Total:      1,
	}
	app.EmitEvent("progress-update", progress)
	return nil
}
func (c *CheckpointService) RevertTaskPaths(projectPath, remoteUrl string, taskPaths []string) error {
	defer reset() // Ensure context is reset when we're done

	ctx := getContext()
	if ctx.Err() != nil {
		return errors.New("operation cancelled before starting")
	}

	app := application.Get()
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	// Create buffered channels to prevent blocking
	errChan := make(chan error, 1)
	progressChan := make(chan output.ProgressReport, 10)

	// Start progress update goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return // Exit immediately on cancellation
			case progress, ok := <-progressChan:
				if !ok {
					return
				}
				app.EmitEvent("progress-update", progress)
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

	quotedTaskPaths := make([]string, len(taskPaths))
	for i, taskPath := range taskPaths {
		quotedTaskPaths[i] = fmt.Sprintf("\"%s\"", taskPath)
	}

	taskIds := []string{}
	err = tx.Select(&taskIds, fmt.Sprintf("SELECT id FROM full_task WHERE trashed = 0 AND task_path IN (%s) ORDER BY created_at DESC", strings.Join(quotedTaskPaths, ",")))
	if err != nil {
		return err
	}

	// Initial progress
	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      "Reverting",
		Message:    "Preparing to Revert",
		Percentage: 0,
		Current:    1,
		Total:      1,
	}:
	}

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

	checkpointIdsToDownload := []string{}
	for _, taskId := range taskIds {
		latestCheckpoint := taskCheckpoints[taskId][0]
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
			default: // Skip progress update if channel is full
			}
		}

		go func() {
			err := sync_service.DownloadCheckpoints(ctx, projectPath, remoteUrl, checkpointIdsToDownload, user.Id, callBack)
			if ctx.Err() == nil { // Only send error if not cancelled
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

	totalTasks := len(taskIds)
	for i, taskId := range taskIds {
		tx, err := dbConn.Beginx()
		if err != nil {
			return err
		}
		task, err := repository.GetTask(tx, taskId)
		if err != nil {
			return err
		}
		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Reverting",
				Message:    task.Name,
				Percentage: float64(current) / float64(total) * 100,
				Current:    i + 1,
				Total:      totalTasks,
			}
			app.EmitEvent("progress-update", progress)
		}

		err = repository.RevertToLatestCheckpoint(tx, taskId, task.FilePath, callBack)
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Rollback()
	}

	close(progressChan)
	progress := output.ProgressReport{
		Title:      "Reverting",
		Message:    "Reverting",
		Percentage: 100,
		Current:    1,
		Total:      1,
	}
	app.EmitEvent("progress-update", progress)
	return nil
}

func (c *CheckpointService) RevertProject(projectPath, remoteUrl string, checkpointTime string) error {
	defer reset() // Ensure context is reset when we're done

	ctx := getContext()
	if ctx.Err() != nil {
		return errors.New("operation cancelled before starting")
	}

	app := application.Get()
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	// Create buffered channels to prevent blocking
	errChan := make(chan error, 1)
	progressChan := make(chan output.ProgressReport, 10)

	// Start progress update goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return // Exit immediately on cancellation
			case progress, ok := <-progressChan:
				if !ok {
					return
				}
				app.EmitEvent("progress-update", progress)
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

	// Initial progress
	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      "Reverting",
		Message:    "Preparing to Revert",
		Percentage: 0,
		Current:    1,
		Total:      1,
	}:
	}

	unixTime, err := time.Parse(time.RFC3339, checkpointTime)
	if err != nil {
		return err
	}
	checkpoints, err := repository.GetLatestCheckpointsByTime(tx, unixTime.Unix())
	if err != nil {
		return err
	}

	checkpointIdsToDownload := []string{}
	for _, checkpoint := range checkpoints {
		isMisssingChunks, err := checkpoint.HasMissingChunks(tx)
		if err != nil {
			return err
		}
		if isMisssingChunks {
			checkpointIdsToDownload = append(checkpointIdsToDownload, checkpoint.Id)
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
			default: // Skip progress update if channel is full
			}
		}

		go func() {
			err := sync_service.DownloadCheckpoints(ctx, projectPath, remoteUrl, checkpointIdsToDownload, user.Id, callBack)
			if ctx.Err() == nil { // Only send error if not cancelled
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

	totalTasks := len(checkpoints)
	for i, checkpoint := range checkpoints {
		tx, err := dbConn.Beginx()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		task, err := repository.GetTask(tx, checkpoint.TaskId)
		if err != nil {
			tx.Rollback()
			return err
		}
		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Reverting",
				Message:    task.Name,
				Percentage: float64(current) / float64(total) * 100,
				Current:    i + 1,
				Total:      totalTasks,
			}
			app.EmitEvent("progress-update", progress)
		}
		if utils.FileExists(task.FilePath) {
			fileXXHash, err := utils.GenerateXXHashChecksum(task.FilePath)
			if err != nil {
				return err
			}
			if fileXXHash == checkpoint.XXHashChecksum {
				progress := output.ProgressReport{
					Title:      "Reverting",
					Message:    task.Name,
					Percentage: 100,
					Current:    i + 1,
					Total:      totalTasks,
				}
				app.EmitEvent("progress-update", progress)
				tx.Rollback()
				continue // Skip if the file is already in the correct state
			}
		}
		err = repository.RevertToCheckpoint(tx, checkpoint.Id, task.FilePath, callBack)
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Rollback()
	}

	close(progressChan)
	progress := output.ProgressReport{
		Title:      "Reverting",
		Message:    "Reverting",
		Percentage: 100,
		Current:    1,
		Total:      1,
	}
	app.EmitEvent("progress-update", progress)
	return nil
}
