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

func CreateCollectionType(tx *sqlx.Tx, id string, name, icon string) (models.CollectionType, error) {
	collectionType := models.CollectionType{}
	params := map[string]interface{}{
		"id":   id,
		"name": name,
		"icon": icon,
	}
	err := base_service.Create(tx, "collection_type", params)
	if err != nil {
		return collectionType, err
	}
	err = base_service.GetByName(tx, "collection_type", name, &collectionType)
	if err != nil {
		return collectionType, err
	}
	return collectionType, nil
}

func GetCollectionType(tx *sqlx.Tx, id string) (models.CollectionType, error) {
	collectionType := models.CollectionType{}
	err := base_service.Get(tx, "collection_type", id, &collectionType)
	if err != nil {
		return collectionType, err
	}
	return collectionType, nil
}

func GetCollectionTypes(tx *sqlx.Tx) ([]models.CollectionType, error) {
	collectionTypes := []models.CollectionType{}
	err := base_service.GetAll(tx, "collection_type", &collectionTypes)
	if err != nil {
		return collectionTypes, err
	}
	sort.Slice(collectionTypes, func(i, j int) bool {
		return collectionTypes[i].Name < collectionTypes[j].Name
	})
	return collectionTypes, nil
}

func GetCollectionTypeByName(tx *sqlx.Tx, name string) (models.CollectionType, error) {
	collectionType := models.CollectionType{}
	err := base_service.GetByName(tx, "collection_type", name, &collectionType)
	if err != nil {
		return collectionType, err
	}
	return collectionType, nil
}

func GetOrCreateCollectionType(tx *sqlx.Tx, name, icon string) (models.CollectionType, error) {
	collectionType, err := GetCollectionTypeByName(tx, name)
	if err == nil {
		return collectionType, nil
	}
	collectionType, err = CreateCollectionType(tx, "", name, icon)
	if err != nil {
		return collectionType, err
	}
	return collectionType, nil
}

func RenameCollectionType(tx *sqlx.Tx, collectionTypeId string, newName string) error {

	// collectionType, err := GetCollectionType(tx, collectionTypeId)
	// if err != nil {
	// 	return err
	// }
	newName = strings.TrimSpace(newName)

	err := base_service.Rename(tx, "collection_type", collectionTypeId, newName)
	if err != nil {
		return err
	}
	return nil
}

func UpdateCollectionType(tx *sqlx.Tx, id string, name, icon string) (models.CollectionType, error) {
	collectionType := models.CollectionType{}
	params := map[string]any{
		"name": name,
		"icon": icon,
	}
	err := base_service.Update(tx, "collection_type", id, params)
	if err != nil {
		return collectionType, err
	}
	err = base_service.UpdateMtime(tx, "collection_type", id, utils.GetEpochTime())
	if err != nil {
		return models.CollectionType{}, err
	}
	err = base_service.GetByName(tx, "collection_type", name, &collectionType)
	if err != nil {
		return collectionType, err
	}
	return collectionType, nil
}

func DeleteCollectionType(tx *sqlx.Tx, collectionTypeId string) error {
	//check if there are assets of this type
	collections := []models.Collection{}
	conditions := map[string]interface{}{
		"collection_type_id": collectionTypeId,
	}
	err := base_service.GetAllBy(tx, "collection", conditions, &collections)
	if err != nil {
		return err
	}
	if len(collections) > 0 {
		return errors.New("cannot delete collection type, there are collections of this type")
	}
	err = base_service.Delete(tx, "collection_type", collectionTypeId)
	if err != nil {
		return err
	}
	return nil
}
