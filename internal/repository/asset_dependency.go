package repository

import (
	"clustta/internal/base_service"
	"clustta/internal/repository/models"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

func AddDependency(tx *sqlx.Tx, id string, assetId string, dependencyId string, dependencyTypeId string) (models.AssetDependency, error) {
	assetDependency := models.AssetDependency{}
	params := map[string]any{
		"id":                 id,
		"asset_id":            assetId,
		"dependency_id":      dependencyId,
		"dependency_type_id": dependencyTypeId,
	}
	err := base_service.Create(tx, "asset_dependency", params)
	if err != nil {
		return assetDependency, err
	}
	conditions := map[string]any{
		"asset_id":       assetId,
		"dependency_id": dependencyId,
	}
	err = base_service.GetBy(tx, "asset_dependency", conditions, &assetDependency)
	if err != nil {
		return assetDependency, err
	}
	return assetDependency, nil
}

func GetDependency(tx *sqlx.Tx, id string) (models.AssetDependency, error) {
	dependency := models.AssetDependency{}
	err := base_service.Get(tx, "asset_dependency", id, &dependency)
	if err != nil {
		return dependency, err
	}
	return dependency, nil
}

func GetAssetDependencies(tx *sqlx.Tx, assetId string) ([]models.AssetDependency, error) {
	assetDependencies := []models.AssetDependency{}
	conditions := map[string]interface{}{
		"asset_id": assetId,
	}
	err := base_service.GetAllBy(tx, "asset_dependency", conditions, &assetDependencies)
	if err != nil {
		return assetDependencies, err
	}
	return assetDependencies, nil
}
func RemoveAssetDependency(tx *sqlx.Tx, assetId string, dependencyId string) error {
	assetDependency := models.AssetDependency{}
	conditions := map[string]interface{}{
		"asset_id":       assetId,
		"dependency_id": dependencyId,
	}
	err := base_service.DeleteBy(tx, "asset_dependency", conditions)
	if err != nil {
		return err
	}
	err = base_service.GetBy(tx, "asset_dependency", conditions, &assetDependency)
	if err == nil {
		return errors.New("dependency failed to remove")
	} else if err != sql.ErrNoRows {
		return err
	}
	return nil
}
