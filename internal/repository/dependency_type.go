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
	assetDependencyType := models.DependencyType{}
	params := map[string]interface{}{
		"id":   id,
		"name": name,
	}
	base_service.Create(tx, "dependency_type", params)
	err := base_service.GetByName(tx, "dependency_type", name, &assetDependencyType)
	if err != nil {
		return assetDependencyType, err
	}
	return assetDependencyType, nil
}

func GetDependencyType(tx *sqlx.Tx, id string) (models.DependencyType, error) {
	assetDependencyType := models.DependencyType{}
	err := base_service.Get(tx, "dependency_type", id, &assetDependencyType)
	if err != nil {
		return assetDependencyType, err
	}
	return assetDependencyType, nil
}

func GetDependencyTypes(tx *sqlx.Tx) ([]models.DependencyType, error) {
	assetDependencyTypes := []models.DependencyType{}
	err := base_service.GetAll(tx, "dependency_type", &assetDependencyTypes)
	if err != nil {
		return assetDependencyTypes, err
	}
	return assetDependencyTypes, nil
}

func GetDependencyTypeByName(tx *sqlx.Tx, name string) (models.DependencyType, error) {
	assetDependencyType := models.DependencyType{}
	err := base_service.GetByName(tx, "dependency_type", name, &assetDependencyType)
	if err != nil {
		return assetDependencyType, err
	}
	return assetDependencyType, nil
}

func GetOrCreateDependencyType(tx *sqlx.Tx, name string) (models.DependencyType, error) {
	assetDependencyType, err := GetDependencyTypeByName(tx, name)
	if err == nil {
		return assetDependencyType, nil
	}
	assetDependencyType, err = CreateDependencyType(tx, "", name)
	if err != nil {
		return assetDependencyType, err
	}
	return assetDependencyType, nil
}

func DeleteDependencyType(tx *sqlx.Tx, assetDependencyTypeId string) error {
	//check if there are assets of this type
	assetDependencies := []models.AssetDependency{}
	conditions := map[string]interface{}{
		"dependency_type_id": assetDependencyTypeId,
	}
	err := base_service.GetAllBy(tx, "asset_dependency", conditions, &assetDependencies)
	if err != nil {
		return err
	}
	if len(assetDependencies) > 0 {
		return errors.New("cannot delete asset type, there are assets of this type")
	}
	err = base_service.Delete(tx, "dependency_type", assetDependencyTypeId)
	if err != nil {
		return err
	}
	return nil
}

func RenameDependencyType(tx *sqlx.Tx, assetDependencyTypeId string, newName string) error {
	newName = strings.TrimSpace(newName)
	err := base_service.Rename(tx, "dependency_type", assetDependencyTypeId, newName)
	if err != nil {
		return err
	}
	return nil
}
