package repository

import (
	"os"

	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func GetTaskFileStatus(task *models.Task, checkpoints []models.Checkpoint) (string, error) {
	filePath := task.GetFilePath()
	if task.IsLink && utils.IsValidPointer(task.Pointer) {
		return "normal", nil
	}
	if task.IsLink && !utils.IsValidPointer(task.Pointer) {
		return "missing", nil
	}

	_, err := os.Stat(filePath)
	isMissing := os.IsNotExist(err)

	// isMissing := !utils.FileExists(filePath)

	if isMissing && len(checkpoints) == 0 {
		return "missing", nil
	} else if isMissing && len(checkpoints) != 0 {
		return "rebuildable", nil
	}

	fileHash, err := utils.GenerateXXHashChecksum(filePath)
	if err != nil {
		return "", err
	}
	for i, checkpoint := range checkpoints {
		if fileHash == checkpoint.XXHashChecksum {
			if i == 0 {
				return "normal", nil
			} else {
				return "outdated", nil

			}
		}
	}
	return "modified", nil
}

func GetFilesStatus(tx *sqlx.Tx, taskIds []string) (map[string]string, error) {
	taskFilesStatus := map[string]string{}

	checkpointQuery := "SELECT * FROM task_checkpoint WHERE trashed = 0 ORDER BY created_at DESC"
	tasksCheckpoints := []models.Checkpoint{}
	tx.Select(&tasksCheckpoints, checkpointQuery)

	taskCheckpoints := map[string][]models.Checkpoint{}
	for _, taskCheckpoint := range tasksCheckpoints {
		taskCheckpoints[taskCheckpoint.TaskId] = append(taskCheckpoints[taskCheckpoint.TaskId], taskCheckpoint)
	}

	for _, taskId := range taskIds {
		query := `SELECT 
			t.id,
			t.name,
			t.description,
			t.created_at,
			t.mtime,
			t.extension,
			t.is_link,
			t.pointer,
			t.status_id,
			t.entity_path
		FROM 
			full_task t
		WHERE 
			t.id = ?;`
		// t.trashed = 0;`
		task := models.Task{}
		err := tx.Get(&task, query, taskId)
		if err != nil {
			return taskFilesStatus, err
		}

		// taskFilePath, err := utils.BuildTaskPath(tx, task.EntityPath, task.Name, task.Extension)
		// if err != nil {
		// 	return taskFilesStatus, err
		// }
		// task.FilePath = taskFilePath

		status, err := GetTaskFileStatus(&task, taskCheckpoints[task.Id])
		if err != nil {
			return taskFilesStatus, err
		}
		taskFilesStatus[taskId] = status
	}
	return taskFilesStatus, nil
}
