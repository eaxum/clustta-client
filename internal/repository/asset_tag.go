package repository

import (
	"clustta/internal/base_service"
	"clustta/internal/repository/models"

	"github.com/jmoiron/sqlx"
)

func CreateTag(tx *sqlx.Tx, id string, name string) (models.Tag, error) {
	tag := models.Tag{}
	params := map[string]interface{}{
		"id":   id,
		"name": name,
	}
	err := base_service.Create(tx, "tag", params)
	if err != nil {
		return tag, err
	}
	err = base_service.GetByName(tx, "tag", name, &tag)
	if err != nil {
		return tag, err
	}
	return tag, nil
}

func GetTag(tx *sqlx.Tx, id string) (models.Tag, error) {
	tag := models.Tag{}
	err := base_service.Get(tx, "tag", id, &tag)
	if err != nil {
		return tag, err
	}
	return tag, err
}

func GetTags(tx *sqlx.Tx) ([]models.Tag, error) {
	tags := []models.Tag{}
	err := base_service.GetAll(tx, "tag", &tags)
	if err != nil {
		return tags, err
	}
	return tags, nil
}

func GetTagByName(tx *sqlx.Tx, name string) (models.Tag, error) {
	tag := models.Tag{}
	err := base_service.GetByName(tx, "tag", name, &tag)
	if err != nil {
		return tag, err
	}
	return tag, err
}

func GetOrCreateTag(tx *sqlx.Tx, name string) (models.Tag, error) {
	tag, err := GetTagByName(tx, name)
	if err == nil {
		return tag, nil
	}

	tag, err = CreateTag(tx, "", name)
	if err != nil {
		return tag, err
	}

	return tag, nil
}

func AddTagToAsset(tx *sqlx.Tx, assetId string, tag string) error {
	tagObj, err := GetOrCreateTag(tx, tag)
	if err != nil {
		return err
	}
	params := map[string]interface{}{
		"asset_id": assetId,
		"tag_id":  tagObj.Id,
	}
	err = base_service.Create(tx, "asset_tag", params)
	if err != nil {
		return err
	}
	return nil
}

func AddTagToAssetById(tx *sqlx.Tx, id, assetId string, tagId string) error {
	params := map[string]interface{}{
		"id":      id,
		"asset_id": assetId,
		"tag_id":  tagId,
	}
	err := base_service.Create(tx, "asset_tag", params)
	if err != nil {
		return err
	}
	return nil
}

func GetAssetTag(tx *sqlx.Tx, Id string) (models.AssetTag, error) {
	assetTag := models.AssetTag{}
	err := base_service.Get(tx, "asset_tag", Id, &assetTag)
	if err != nil {
		return assetTag, err
	}
	return assetTag, nil
}

func GetAssetTags(tx *sqlx.Tx, assetId string) ([]models.Tag, error) {
	tags := []models.Tag{}
	err := tx.Select(&tags, "SELECT * FROM tag WHERE id IN (SELECT tag_id FROM asset_tag WHERE asset_id = ?)", assetId)
	if err != nil {
		return tags, err
	}
	return tags, nil
}

func RemoveTagFromAsset(tx *sqlx.Tx, assetId string, tagId string) error {
	conditions := map[string]interface{}{
		"asset_id": assetId,
		"tag_id":  tagId,
	}
	err := base_service.DeleteBy(tx, "asset_tag", conditions)
	return err
}

func RemoveAllTagsFromAsset(tx *sqlx.Tx, assetId string) error {
	conditions := map[string]interface{}{
		"asset_id": assetId,
	}
	err := base_service.DeleteBy(tx, "asset_tag", conditions)
	return err
}

// func GetAssetTagsByTagId(tx *sqlx.Tx, tagId string) []Asset {
// 	dbConn, err := utils.OpenDb( projectPath)
// 	if err != nil {
// 		panic(err)
// 	}
// 	assets := []Asset{}
// 	tx.Select(&assets, "SELECT * FROM asset WHERE id IN (SELECT asset_id FROM asset_tag WHERE tag_id = ?)", tagId)
// 	return assets
// }

// func GetAssetTagsByTagName(tx *sqlx.Tx, tagName string) []Asset {
// 	dbConn, err := utils.OpenDb( projectPath)
// 	if err != nil {
// 		panic(err)
// 	}
// 	assets := []Asset{}
// 	tx.Select(&assets, "SELECT * FROM asset WHERE id IN (SELECT asset_id FROM asset_tag WHERE tag_id IN (SELECT id FROM tag WHERE name = ?))", tagName)
// 	return assets
// }

// func GetAssetTagsByAssetId(tx *sqlx.Tx, assetId string) []models.Tag {
// 	dbConn, err := utils.OpenDb( projectPath)
// 	if err != nil {
// 		panic(err)
// 	}
// 	tags := []models.Tag{}
// 	tx.Select(&tags, "SELECT * FROM tag WHERE id IN (SELECT tag_id FROM asset_tag WHERE asset_id = ?)", assetId)
// 	return tags
// }
