package services

import (
	"clustta/internal/error_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type TagService struct{}

const tagManagementPermission = "change_role"

func authorizeTagManagementTx(tx *sqlx.Tx) error {
	_, role, err := activeAssetRole(tx)
	if err != nil {
		return err
	}
	if !role.ChangeRole {
		return fmt.Errorf("user does not have %s permission", tagManagementPermission)
	}
	return nil
}

func normalizeTagName(name string) (string, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return "", errors.New("tag name is required")
	}
	return trimmedName, nil
}

func isTagNameCollision(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: tag.name")
}

// CreateTag creates a project tag.
func (t *TagService) CreateTag(projectPath, name string) (models.Tag, error) {
	name, err := normalizeTagName(name)
	if err != nil {
		return models.Tag{}, err
	}
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Tag{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Tag{}, err
	}
	defer tx.Rollback()
	if err = authorizeTagManagementTx(tx); err != nil {
		return models.Tag{}, err
	}

	tag, err := repository.CreateTag(tx, "", name)
	if isTagNameCollision(err) {
		return models.Tag{}, error_service.ErrTagExists
	}
	if err != nil {
		return models.Tag{}, err
	}
	if err = tx.Commit(); err != nil {
		return models.Tag{}, err
	}
	return tag, nil
}

// UpdateTag renames a tag and optionally merges a colliding tag.
func (t *TagService) UpdateTag(projectPath, id, name string, mergeOnCollision bool) (models.Tag, error) {
	name, err := normalizeTagName(name)
	if err != nil {
		return models.Tag{}, err
	}
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Tag{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Tag{}, err
	}
	defer tx.Rollback()
	if err = authorizeTagManagementTx(tx); err != nil {
		return models.Tag{}, err
	}

	tag, err := repository.UpdateTag(tx, id, name, mergeOnCollision)
	if isTagNameCollision(err) {
		return models.Tag{}, error_service.ErrTagExists
	}
	if err != nil {
		return models.Tag{}, err
	}
	if err = tx.Commit(); err != nil {
		return models.Tag{}, err
	}
	return tag, nil
}

// DeleteTag removes a tag and its asset assignments.
func (t *TagService) DeleteTag(projectPath, id string) error {
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
	if err = authorizeTagManagementTx(tx); err != nil {
		return err
	}
	if err = repository.DeleteTag(tx, id); err != nil {
		return err
	}
	return tx.Commit()
}

// GetTagUsageCount returns the number of assets assigned to a tag.
func (t *TagService) GetTagUsageCount(projectPath, id string) (int64, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return 0, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	return repository.GetTagUsageCount(tx, id)
}

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
	if err := authorizeAssetActionTx(tx, assetActionUpdate, []string{assetId}); err != nil {
		return []models.Tag{}, err
	}

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
	if err := authorizeAssetActionTx(tx, assetActionUpdate, []string{assetId}); err != nil {
		return []models.Tag{}, err
	}

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
	if err := authorizeAssetActionTx(tx, assetActionUpdate, assetIds); err != nil {
		return err
	}

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
	if err := authorizeAssetActionTx(tx, assetActionUpdate, assetIds); err != nil {
		return err
	}

	for _, assetId := range assetIds {
		if err := repository.RemoveTagFromAsset(tx, assetId, tagId); err != nil {
			return err
		}
	}

	return tx.Commit()
}
