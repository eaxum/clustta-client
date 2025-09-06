package services

import (
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
)

type TagService struct{}

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
