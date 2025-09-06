package repository

import (
	"clustta/internal/base_service"
	"clustta/internal/repository/models"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

func AddEntityDependency(tx *sqlx.Tx, id string, taskId string, dependencyId string, dependencyTypeId string) (models.TaskDependency, error) {
	taskDependency := models.TaskDependency{}
	params := map[string]any{
		"id":                 id,
		"task_id":            taskId,
		"dependency_id":      dependencyId,
		"dependency_type_id": dependencyTypeId,
	}
	err := base_service.Create(tx, "entity_dependency", params)
	if err != nil {
		return taskDependency, err
	}
	conditions := map[string]any{
		"task_id":       taskId,
		"dependency_id": dependencyId,
	}
	err = base_service.GetBy(tx, "entity_dependency", conditions, &taskDependency)
	if err != nil {
		return taskDependency, err
	}
	return taskDependency, nil
}

func GetEntityDependency(tx *sqlx.Tx, id string) (models.TaskDependency, error) {
	dependency := models.TaskDependency{}
	err := base_service.Get(tx, "entity_dependency", id, &dependency)
	if err != nil {
		return dependency, err
	}
	return dependency, nil
}

func GetEntityDependencies(tx *sqlx.Tx, taskId string) ([]models.TaskDependency, error) {
	taskDependencies := []models.TaskDependency{}
	conditions := map[string]interface{}{
		"task_id": taskId,
	}
	err := base_service.GetAllBy(tx, "entity_dependency", conditions, &taskDependencies)
	if err != nil {
		return taskDependencies, err
	}
	return taskDependencies, nil
}

func RemoveEntityDependency(tx *sqlx.Tx, taskId string, dependencyId string) error {
	taskDependency := models.TaskDependency{}
	conditions := map[string]interface{}{
		"task_id":       taskId,
		"dependency_id": dependencyId,
	}
	err := base_service.DeleteBy(tx, "entity_dependency", conditions)
	if err != nil {
		return err
	}
	err = base_service.GetBy(tx, "entity_dependency", conditions, &taskDependency)
	if err == nil {
		return errors.New("entity dependency failed to remove")
	} else if err != sql.ErrNoRows {
		return err
	}
	return nil
}
