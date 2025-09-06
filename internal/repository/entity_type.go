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

func CreateEntityType(tx *sqlx.Tx, id string, name, icon string) (models.EntityType, error) {
	entityType := models.EntityType{}
	params := map[string]interface{}{
		"id":   id,
		"name": name,
		"icon": icon,
	}
	err := base_service.Create(tx, "entity_type", params)
	if err != nil {
		return entityType, err
	}
	err = base_service.GetByName(tx, "entity_type", name, &entityType)
	if err != nil {
		return entityType, err
	}
	return entityType, nil
}

func GetEntityType(tx *sqlx.Tx, id string) (models.EntityType, error) {
	entityType := models.EntityType{}
	err := base_service.Get(tx, "entity_type", id, &entityType)
	if err != nil {
		return entityType, err
	}
	return entityType, nil
}

func GetEntityTypes(tx *sqlx.Tx) ([]models.EntityType, error) {
	entityTypes := []models.EntityType{}
	err := base_service.GetAll(tx, "entity_type", &entityTypes)
	if err != nil {
		return entityTypes, err
	}
	sort.Slice(entityTypes, func(i, j int) bool {
		return entityTypes[i].Name < entityTypes[j].Name
	})
	return entityTypes, nil
}

func GetEntityTypeByName(tx *sqlx.Tx, name string) (models.EntityType, error) {
	entityType := models.EntityType{}
	err := base_service.GetByName(tx, "entity_type", name, &entityType)
	if err != nil {
		return entityType, err
	}
	return entityType, nil
}

func GetOrCreateEntityType(tx *sqlx.Tx, name, icon string) (models.EntityType, error) {
	entityType, err := GetEntityTypeByName(tx, name)
	if err == nil {
		return entityType, nil
	}
	entityType, err = CreateEntityType(tx, "", name, icon)
	if err != nil {
		return entityType, err
	}
	return entityType, nil
}

func RenameEntityType(tx *sqlx.Tx, entityTypeId string, newName string) error {

	// entityType, err := GetEntityType(tx, entityTypeId)
	// if err != nil {
	// 	return err
	// }
	newName = strings.TrimSpace(newName)

	err := base_service.Rename(tx, "entity_type", entityTypeId, newName)
	if err != nil {
		return err
	}
	return nil
}

func UpdateEntityType(tx *sqlx.Tx, id string, name, icon string) (models.EntityType, error) {
	entityType := models.EntityType{}
	params := map[string]any{
		"name": name,
		"icon": icon,
	}
	err := base_service.Update(tx, "entity_type", id, params)
	if err != nil {
		return entityType, err
	}
	err = base_service.UpdateMtime(tx, "entity_type", id, utils.GetEpochTime())
	if err != nil {
		return models.EntityType{}, err
	}
	err = base_service.GetByName(tx, "entity_type", name, &entityType)
	if err != nil {
		return entityType, err
	}
	return entityType, nil
}

func DeleteEntityType(tx *sqlx.Tx, entityTypeId string) error {
	//check if there are tasks of this type
	entities := []models.Entity{}
	conditions := map[string]interface{}{
		"entity_type_id": entityTypeId,
	}
	err := base_service.GetAllBy(tx, "entity", conditions, &entities)
	if err != nil {
		return err
	}
	if len(entities) > 0 {
		return errors.New("cannot delete entity type, there are entities of this type")
	}
	err = base_service.Delete(tx, "entity_type", entityTypeId)
	if err != nil {
		return err
	}
	return nil
}
