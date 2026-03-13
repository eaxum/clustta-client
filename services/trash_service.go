package services

import (
	"clustta/internal/base_service"
	"clustta/internal/repository"
	"clustta/internal/utils"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type RecycleItem struct {
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Id         string      `json:"id"`
	ParentId   string      `json:"parent_id"`
	ParentName string      `json:"parent_name"`
	Data       interface{} `json:"data"`
}

type TrashService struct{}

// Retrieves all deleted items from project database including collections, templates, assets, and checkpoints
func (t *TrashService) GetTrashs(projectPath string) ([]RecycleItem, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return []RecycleItem{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []RecycleItem{}, err
	}
	defer tx.Rollback()

	recycleItems := []RecycleItem{}

	collections, err := repository.GetDeletedCollections(tx)
	if err != nil {
		return []RecycleItem{}, err
	}
	for _, collection := range collections {
		recycleItem := RecycleItem{
			Name: collection.Name,
			Type: "collection",
			Id:   collection.Id,
		}
		recycleItems = append(recycleItems, recycleItem)

	}

	//template
	templates, err := repository.GetDeletedTemplates(tx)
	if err != nil {
		return []RecycleItem{}, err
	}
	for _, template := range templates {
		recycleItem := RecycleItem{
			Name: template.Name,
			Type: "template",
			Id:   template.Id,
		}
		recycleItems = append(recycleItems, recycleItem)

	}

	//asset

	assets, err := repository.GetDeletedAssets(tx)
	if err != nil {
		return []RecycleItem{}, err
	}
	for _, asset := range assets {
		recycleItem := RecycleItem{
			Name:       asset.Name,
			Type:       "asset",
			Id:         asset.Id,
			ParentId:   asset.CollectionId,
			ParentName: asset.CollectionName,
		}
		recycleItems = append(recycleItems, recycleItem)

	}

	assetCheckpoints, err := repository.GetDeletedCheckpoints(tx)
	if err != nil {
		return []RecycleItem{}, err
	}
	for _, assetCheckpoint := range assetCheckpoints {
		checkpointAsset, err := repository.GetAsset(tx, assetCheckpoint.AssetId)
		if err != nil {
			return []RecycleItem{}, err
		}
		checkpointName := fmt.Sprintf("%s %s", checkpointAsset.Name, assetCheckpoint.CreatedAt)
		asset, err := repository.GetAsset(tx, assetCheckpoint.AssetId)
		if err != nil {
			return []RecycleItem{}, err
		}
		recycleItem := RecycleItem{
			Name:       checkpointName,
			Type:       "asset_checkpoint",
			Id:         assetCheckpoint.Id,
			ParentId:   assetCheckpoint.AssetId,
			ParentName: asset.Name,
		}
		recycleItems = append(recycleItems, recycleItem)

	}
	return recycleItems, nil
}

// Restores a deleted item by ID and type from the recycle bin
func (t *TrashService) Restore(projectPath, id, itemType string) error {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = base_service.Restore(tx, itemType, id)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
