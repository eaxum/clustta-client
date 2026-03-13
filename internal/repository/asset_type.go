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

func CreateAssetType(tx *sqlx.Tx, id string, name, icon string) (models.AssetType, error) {
	assetType := models.AssetType{}
	params := map[string]interface{}{
		"id":   id,
		"name": name,
		"icon": icon,
	}
	err := base_service.Create(tx, "asset_type", params)
	if err != nil {
		return assetType, err
	}
	err = base_service.GetByName(tx, "asset_type", name, &assetType)
	if err != nil {
		return assetType, err
	}
	return assetType, nil
}

func GetAssetType(tx *sqlx.Tx, id string) (models.AssetType, error) {
	assetType := models.AssetType{}
	err := base_service.Get(tx, "asset_type", id, &assetType)
	if err != nil {
		return assetType, err
	}
	return assetType, nil
}

func GetAssetTypes(tx *sqlx.Tx) ([]models.AssetType, error) {
	assetTypes := []models.AssetType{}
	err := base_service.GetAll(tx, "asset_type", &assetTypes)
	if err != nil {
		return assetTypes, err
	}
	sort.Slice(assetTypes, func(i, j int) bool {
		return assetTypes[i].Name < assetTypes[j].Name
	})
	return assetTypes, nil
}

func GetAssetTypeByName(tx *sqlx.Tx, name string) (models.AssetType, error) {
	assetType := models.AssetType{}
	err := base_service.GetByName(tx, "asset_type", name, &assetType)
	if err != nil {
		return assetType, err
	}
	return assetType, nil
}

func GetOrCreateAssetType(tx *sqlx.Tx, name, icon string) (models.AssetType, error) {
	assetType, err := GetAssetTypeByName(tx, name)
	if err == nil {
		return assetType, nil
	}
	assetType, err = CreateAssetType(tx, "", name, icon)
	if err != nil {
		return assetType, err
	}
	return assetType, nil
}

func RenameAssetType(tx *sqlx.Tx, AssetTypeId string, newName string) error {

	// subassetType, err := GetAssetType(tx, subassetTypeId)
	// if err != nil {
	// 	return err
	// }
	newName = strings.TrimSpace(newName)

	err := base_service.Rename(tx, "asset_type", AssetTypeId, newName)
	if err != nil {
		return err
	}
	return nil
}

func UpdateAssetType(tx *sqlx.Tx, id string, name, icon string) (models.AssetType, error) {
	assetType := models.AssetType{}
	params := map[string]interface{}{
		"name": name,
		"icon": icon,
	}
	err := base_service.Update(tx, "asset_type", id, params)
	if err != nil {
		return assetType, err
	}
	err = base_service.UpdateMtime(tx, "asset_type", id, utils.GetEpochTime())
	if err != nil {
		return models.AssetType{}, err
	}
	err = base_service.GetByName(tx, "asset_type", name, &assetType)
	if err != nil {
		return assetType, err
	}
	return assetType, nil
}

func DeleteAssetType(tx *sqlx.Tx, AssetTypeId string) error {
	//check if there are assets of this type
	assets := []models.Asset{}
	conditions := map[string]interface{}{
		"asset_type_id": AssetTypeId,
	}
	err := base_service.GetAllBy(tx, "asset", conditions, &assets)
	if err != nil {
		return err
	}
	if len(assets) > 0 {
		return errors.New("cannot delete asset_type type, there are assets of this type")
	}
	err = base_service.Delete(tx, "asset_type", AssetTypeId)
	if err != nil {
		return err
	}
	return nil
}
