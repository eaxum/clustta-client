package services

import (
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
)

type StatusService struct{}

func (s *StatusService) GetStatuses(projectPath string) ([]models.Status, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Status{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Status{}, err
	}
	defer tx.Rollback()
	statuses, err := repository.GetStatuses(tx)
	if err != nil {
		return []models.Status{}, err
	}
	return statuses, nil
}
