package services

import (
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
)

type TagService struct{}

// Retrieves all tags from the project database.
func (t *TagService) GetTags(projectPath string) ([]models.Tag, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Tag{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Tag{}, err
	}
	defer tx.Rollback()

	tags, err := repository.GetTags(tx)
	if err != nil {
		return []models.Tag{}, err
	}
	return tags, nil
}

// Retrieves all tags associated with a specific asset.
func (t *TagService) GetAssetTags(projectPath string, assetId string) ([]models.Tag, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Tag{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Tag{}, err
	}
	defer tx.Rollback()

	tags, err := repository.GetAssetTags(tx, assetId)
	if err != nil {
		return []models.Tag{}, err
	}
	return tags, nil
}

// Adds a tag to an asset by name, creating the tag if it doesn't exist.
// Returns the updated list of tags for the asset.
func (t *TagService) AddTagToAsset(projectPath string, assetId string, tagName string) ([]models.Tag, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Tag{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Tag{}, err
	}
	defer tx.Rollback()

	err = repository.AddTagToAsset(tx, assetId, tagName)
	if err != nil {
		return []models.Tag{}, err
	}

	tags, err := repository.GetAssetTags(tx, assetId)
	if err != nil {
		return []models.Tag{}, err
	}

	err = tx.Commit()
	if err != nil {
		return []models.Tag{}, err
	}
	return tags, nil
}

// Removes a tag from an asset by tag ID.
// Returns the updated list of tags for the asset.
func (t *TagService) RemoveTagFromAsset(projectPath string, assetId string, tagId string) ([]models.Tag, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Tag{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Tag{}, err
	}
	defer tx.Rollback()

	err = repository.RemoveTagFromAsset(tx, assetId, tagId)
	if err != nil {
		return []models.Tag{}, err
	}

	tags, err := repository.GetAssetTags(tx, assetId)
	if err != nil {
		return []models.Tag{}, err
	}

	err = tx.Commit()
	if err != nil {
		return []models.Tag{}, err
	}
	return tags, nil
}

// Adds a tag (by name) to multiple assets in a single transaction.
// Skips assets that already have the tag. Creates the tag if it doesn't exist.
func (t *TagService) AddTagToAssets(projectPath string, assetIds []string, tagName string) error {
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

	for _, assetId := range assetIds {
		existing, err := repository.GetAssetTags(tx, assetId)
		if err != nil {
			return err
		}
		alreadyTagged := false
		for _, tag := range existing {
			if tag.Name == tagName {
				alreadyTagged = true
				break
			}
		}
		if alreadyTagged {
			continue
		}
		if err := repository.AddTagToAsset(tx, assetId, tagName); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Removes a tag (by tag ID) from multiple assets in a single transaction.
// Skips assets that don't have the tag.
func (t *TagService) RemoveTagFromAssets(projectPath string, assetIds []string, tagId string) error {
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

	for _, assetId := range assetIds {
		if err := repository.RemoveTagFromAsset(tx, assetId, tagId); err != nil {
			return err
		}
	}

	return tx.Commit()
}
