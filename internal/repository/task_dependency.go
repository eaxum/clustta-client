package repository

import (
	"clustta/internal/base_service"
	"clustta/internal/repository/models"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

func AddDependency(tx *sqlx.Tx, id string, taskId string, dependencyId string, dependencyTypeId string) (models.TaskDependency, error) {
	taskDependency := models.TaskDependency{}
	params := map[string]any{
		"id":                 id,
		"task_id":            taskId,
		"dependency_id":      dependencyId,
		"dependency_type_id": dependencyTypeId,
	}
	err := base_service.Create(tx, "task_dependency", params)
	if err != nil {
		return taskDependency, err
	}
	conditions := map[string]any{
		"task_id":       taskId,
		"dependency_id": dependencyId,
	}
	err = base_service.GetBy(tx, "task_dependency", conditions, &taskDependency)
	if err != nil {
		return taskDependency, err
	}
	return taskDependency, nil
}

func GetDependency(tx *sqlx.Tx, id string) (models.TaskDependency, error) {
	dependency := models.TaskDependency{}
	err := base_service.Get(tx, "task_dependency", id, &dependency)
	if err != nil {
		return dependency, err
	}
	return dependency, nil
}

func GetTaskDependencies(tx *sqlx.Tx, taskId string) ([]models.TaskDependency, error) {
	taskDependencies := []models.TaskDependency{}
	conditions := map[string]interface{}{
		"task_id": taskId,
	}
	err := base_service.GetAllBy(tx, "task_dependency", conditions, &taskDependencies)
	if err != nil {
		return taskDependencies, err
	}
	return taskDependencies, nil
}
func RemoveTaskDependency(tx *sqlx.Tx, taskId string, dependencyId string) error {
	taskDependency := models.TaskDependency{}
	conditions := map[string]interface{}{
		"task_id":       taskId,
		"dependency_id": dependencyId,
	}
	err := base_service.DeleteBy(tx, "task_dependency", conditions)
	if err != nil {
		return err
	}
	err = base_service.GetBy(tx, "task_dependency", conditions, &taskDependency)
	if err == nil {
		return errors.New("dependency failed to remove")
	} else if err != sql.ErrNoRows {
		return err
	}
	return nil
}
