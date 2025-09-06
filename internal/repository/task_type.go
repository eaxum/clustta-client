package repository

import (
	"clustta/internal/base_service"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"errors"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func CreateTaskType(tx *sqlx.Tx, id string, name, icon string) (models.TaskType, error) {
	taskType := models.TaskType{}
	params := map[string]interface{}{
		"id":   id,
		"name": name,
		"icon": icon,
	}
	err := base_service.Create(tx, "task_type", params)
	if err != nil {
		return taskType, err
	}
	err = base_service.GetByName(tx, "task_type", name, &taskType)
	if err != nil {
		return taskType, err
	}
	return taskType, nil
}

func GetTaskType(tx *sqlx.Tx, id string) (models.TaskType, error) {
	taskType := models.TaskType{}
	err := base_service.Get(tx, "task_type", id, &taskType)
	if err != nil {
		return taskType, err
	}
	return taskType, nil
}

func GetTaskTypes(tx *sqlx.Tx) ([]models.TaskType, error) {
	taskTypes := []models.TaskType{}
	err := base_service.GetAll(tx, "task_type", &taskTypes)
	if err != nil {
		return taskTypes, err
	}
	sort.Slice(taskTypes, func(i, j int) bool {
		return taskTypes[i].Name < taskTypes[j].Name
	})
	return taskTypes, nil
}

func GetTaskTypeByName(tx *sqlx.Tx, name string) (models.TaskType, error) {
	taskType := models.TaskType{}
	err := base_service.GetByName(tx, "task_type", name, &taskType)
	if err != nil {
		return taskType, err
	}
	return taskType, nil
}

func GetOrCreateTaskType(tx *sqlx.Tx, name, icon string) (models.TaskType, error) {
	taskType, err := GetTaskTypeByName(tx, name)
	if err == nil {
		return taskType, nil
	}
	taskType, err = CreateTaskType(tx, "", name, icon)
	if err != nil {
		return taskType, err
	}
	return taskType, nil
}

func RenameTaskType(tx *sqlx.Tx, TaskTypeId string, newName string) error {

	// subtaskType, err := GetTaskType(tx, subtaskTypeId)
	// if err != nil {
	// 	return err
	// }
	newName = strings.TrimSpace(newName)

	err := base_service.Rename(tx, "task_type", TaskTypeId, newName)
	if err != nil {
		return err
	}
	return nil
}

func UpdateTaskType(tx *sqlx.Tx, id string, name, icon string) (models.TaskType, error) {
	taskType := models.TaskType{}
	params := map[string]interface{}{
		"name": name,
		"icon": icon,
	}
	err := base_service.Update(tx, "task_type", id, params)
	if err != nil {
		return taskType, err
	}
	err = base_service.UpdateMtime(tx, "task_type", id, utils.GetEpochTime())
	if err != nil {
		return models.TaskType{}, err
	}
	err = base_service.GetByName(tx, "task_type", name, &taskType)
	if err != nil {
		return taskType, err
	}
	return taskType, nil
}

func DeleteTaskType(tx *sqlx.Tx, TaskTypeId string) error {
	//check if there are tasks of this type
	tasks := []models.Task{}
	conditions := map[string]interface{}{
		"task_type_id": TaskTypeId,
	}
	err := base_service.GetAllBy(tx, "task", conditions, &tasks)
	if err != nil {
		return err
	}
	if len(tasks) > 0 {
		return errors.New("cannot delete task_type type, there are tasks of this type")
	}
	err = base_service.Delete(tx, "task_type", TaskTypeId)
	if err != nil {
		return err
	}
	return nil
}
