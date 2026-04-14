package repository

import (
	"clustta/internal/base_service"
	"clustta/internal/repository/models"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

func AddCollectionDependency(tx *sqlx.Tx, id string, assetId string, dependencyId string, dependencyTypeId string) (models.AssetDependency, error) {
	assetDependency := models.AssetDependency{}
	params := map[string]any{
		"id":                 id,
		"asset_id":            assetId,
		"dependency_id":      dependencyId,
		"dependency_type_id": dependencyTypeId,
	}
	err := base_service.Create(tx, "collection_dependency", params)
	if err != nil {
		return assetDependency, err
	}
	conditions := map[string]any{
		"asset_id":       assetId,
		"dependency_id": dependencyId,
	}
	err = base_service.GetBy(tx, "collection_dependency", conditions, &assetDependency)
	if err != nil {
		return assetDependency, err
	}
	return assetDependency, nil
}

func GetCollectionDependency(tx *sqlx.Tx, id string) (models.AssetDependency, error) {
	dependency := models.AssetDependency{}
	err := base_service.Get(tx, "collection_dependency", id, &dependency)
	if err != nil {
		return dependency, err
	}
	return dependency, nil
}

// GetCollectionDependencyByKeys looks up an collection dependency by its composite key.
func GetCollectionDependencyByKeys(tx *sqlx.Tx, assetId, dependencyId string) (models.AssetDependency, error) {
	dep := models.AssetDependency{}
	conditions := map[string]any{
		"asset_id":       assetId,
		"dependency_id": dependencyId,
	}
	err := base_service.GetBy(tx, "collection_dependency", conditions, &dep)
	return dep, err
}

func GetCollectionDependencies(tx *sqlx.Tx, assetId string) ([]models.AssetDependency, error) {
	assetDependencies := []models.AssetDependency{}
	conditions := map[string]interface{}{
		"asset_id": assetId,
	}
	err := base_service.GetAllBy(tx, "collection_dependency", conditions, &assetDependencies)
	if err != nil {
		return assetDependencies, err
	}
	return assetDependencies, nil
}

func RemoveCollectionDependency(tx *sqlx.Tx, assetId string, dependencyId string) error {
	assetDependency := models.AssetDependency{}
	conditions := map[string]interface{}{
		"asset_id":       assetId,
		"dependency_id": dependencyId,
	}
	err := base_service.DeleteBy(tx, "collection_dependency", conditions)
	if err != nil {
		return err
	}
	err = base_service.GetBy(tx, "collection_dependency", conditions, &assetDependency)
	if err == nil {
		return errors.New("collection dependency failed to remove")
	} else if err != sql.ErrNoRows {
		return err
	}
	return nil
}
