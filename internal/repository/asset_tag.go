package repository

import (
	"clustta/internal/base_service"
	"clustta/internal/error_service"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"errors"

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

func GetTagUsageCount(tx *sqlx.Tx, id string) (int64, error) {
	if _, err := GetTag(tx, id); err != nil {
		return 0, err
	}
	var count int64
	err := tx.Get(&count, "SELECT COUNT(*) FROM asset_tag WHERE tag_id = ?", id)
	return count, err
}

func UpdateTag(tx *sqlx.Tx, id, name string, mergeOnCollision bool) (models.Tag, error) {
	if _, err := GetTag(tx, id); err != nil {
		return models.Tag{}, err
	}

	existingTag, err := GetTagByName(tx, name)
	if err == nil && existingTag.Id != id {
		if !mergeOnCollision {
			return models.Tag{}, error_service.ErrTagExists
		}
		return mergeTags(tx, id, existingTag)
	}
	if err != nil && !errors.Is(err, error_service.ErrTagNotFound) {
		return models.Tag{}, err
	}

	params := map[string]interface{}{
		"name":  name,
		"mtime": utils.GetEpochTime(),
	}
	if err = base_service.Update(tx, "tag", id, params); err != nil {
		return models.Tag{}, err
	}
	return GetTag(tx, id)
}

func mergeTags(tx *sqlx.Tx, sourceTagId string, targetTag models.Tag) (models.Tag, error) {
	if err := mergeCheckpointTagAssignments(tx, sourceTagId, targetTag.Id); err != nil {
		return models.Tag{}, err
	}

	deleteDuplicatesQuery := `
		DELETE FROM asset_tag
		WHERE tag_id = ?
		AND asset_id IN (SELECT asset_id FROM asset_tag WHERE tag_id = ?)`
	if _, err := tx.Exec(deleteDuplicatesQuery, sourceTagId, targetTag.Id); err != nil {
		return models.Tag{}, err
	}

	updateAssignmentsQuery := "UPDATE asset_tag SET tag_id = ?, mtime = ? WHERE tag_id = ?"
	if _, err := tx.Exec(updateAssignmentsQuery, targetTag.Id, utils.GetEpochTime(), sourceTagId); err != nil {
		return models.Tag{}, err
	}

	if err := base_service.Delete(tx, "tag", sourceTagId); err != nil {
		return models.Tag{}, err
	}
	return targetTag, nil
}

func mergeCheckpointTagAssignments(tx *sqlx.Tx, sourceTagId, targetTagId string) error {
	now := utils.GetEpochTime()
	_, err := tx.Exec(`
		UPDATE asset_dependency
		SET asset_checkpoint_tag_id = (
			SELECT target.id
			FROM asset_checkpoint_tag source
			JOIN asset_checkpoint_tag target
				ON target.asset_id = source.asset_id AND target.tag_id = ?
			WHERE source.id = asset_dependency.asset_checkpoint_tag_id
		), mtime = ?, synced = 0
		WHERE asset_checkpoint_tag_id IN (
			SELECT source.id
			FROM asset_checkpoint_tag source
			JOIN asset_checkpoint_tag target
				ON target.asset_id = source.asset_id AND target.tag_id = ?
			WHERE source.tag_id = ?
		)
	`, targetTagId, now, targetTagId, sourceTagId)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		DELETE FROM asset_checkpoint_tag
		WHERE tag_id = ?
		AND asset_id IN (SELECT asset_id FROM asset_checkpoint_tag WHERE tag_id = ?)
	`, sourceTagId, targetTagId)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		UPDATE asset_checkpoint_tag SET tag_id = ?, mtime = ?, synced = 0 WHERE tag_id = ?
	`, targetTagId, now, sourceTagId)
	return err
}

func DeleteTag(tx *sqlx.Tx, id string) error {
	if _, err := GetTag(tx, id); err != nil {
		return err
	}
	if err := base_service.DeleteBy(tx, "asset_checkpoint_tag", map[string]interface{}{"tag_id": id}); err != nil {
		return err
	}
	if err := base_service.DeleteBy(tx, "asset_tag", map[string]interface{}{"tag_id": id}); err != nil {
		return err
	}
	return base_service.Delete(tx, "tag", id)
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
		"tag_id":   tagObj.Id,
	}
	err = base_service.Create(tx, "asset_tag", params)
	if err != nil {
		return err
	}
	return nil
}

func AddTagToAssetById(tx *sqlx.Tx, id, assetId string, tagId string) error {
	params := map[string]interface{}{
		"id":       id,
		"asset_id": assetId,
		"tag_id":   tagId,
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
		"tag_id":   tagId,
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
