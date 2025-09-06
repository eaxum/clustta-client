package repository

import (
	"errors"
	"strings"

	"clustta/internal/base_service"
	"clustta/internal/repository/models"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func CreateDependencyType(tx *sqlx.Tx, id string, name string) (models.DependencyType, error) {
	taskDependencyType := models.DependencyType{}
	params := map[string]interface{}{
		"id":   id,
		"name": name,
	}
	base_service.Create(tx, "dependency_type", params)
	err := base_service.GetByName(tx, "dependency_type", name, &taskDependencyType)
	if err != nil {
		return taskDependencyType, err
	}
	return taskDependencyType, nil
}

func GetDependencyType(tx *sqlx.Tx, id string) (models.DependencyType, error) {
	taskDependencyType := models.DependencyType{}
	err := base_service.Get(tx, "dependency_type", id, &taskDependencyType)
	if err != nil {
		return taskDependencyType, err
	}
	return taskDependencyType, nil
}

func GetDependencyTypes(tx *sqlx.Tx) ([]models.DependencyType, error) {
	taskDependencyTypes := []models.DependencyType{}
	err := base_service.GetAll(tx, "dependency_type", &taskDependencyTypes)
	if err != nil {
		return taskDependencyTypes, err
	}
	return taskDependencyTypes, nil
}

func GetDependencyTypeByName(tx *sqlx.Tx, name string) (models.DependencyType, error) {
	taskDependencyType := models.DependencyType{}
	err := base_service.GetByName(tx, "dependency_type", name, &taskDependencyType)
	if err != nil {
		return taskDependencyType, err
	}
	return taskDependencyType, nil
}

func GetOrCreateDependencyType(tx *sqlx.Tx, name string) (models.DependencyType, error) {
	taskDependencyType, err := GetDependencyTypeByName(tx, name)
	if err == nil {
		return taskDependencyType, nil
	}
	taskDependencyType, err = CreateDependencyType(tx, "", name)
	if err != nil {
		return taskDependencyType, err
	}
	return taskDependencyType, nil
}

func DeleteDependencyType(tx *sqlx.Tx, taskDependencyTypeId string) error {
	//check if there are tasks of this type
	taskDependencies := []models.TaskDependency{}
	conditions := map[string]interface{}{
		"dependency_type_id": taskDependencyTypeId,
	}
	err := base_service.GetAllBy(tx, "task_dependency", conditions, &taskDependencies)
	if err != nil {
		return err
	}
	if len(taskDependencies) > 0 {
		return errors.New("cannot delete task type, there are tasks of this type")
	}
	err = base_service.Delete(tx, "dependency_type", taskDependencyTypeId)
	if err != nil {
		return err
	}
	return nil
}

func RenameDependencyType(tx *sqlx.Tx, taskDependencyTypeId string, newName string) error {
	newName = strings.TrimSpace(newName)
	err := base_service.Rename(tx, "dependency_type", taskDependencyTypeId, newName)
	if err != nil {
		return err
	}
	return nil
}
