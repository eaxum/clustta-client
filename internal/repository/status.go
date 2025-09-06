package repository

import (
	"clustta/internal/base_service"
	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
)

func CreateStatus(tx *sqlx.Tx, id string, name string, shortName string, color string) (models.Status, error) {
	params := map[string]interface{}{
		"id":         id,
		"name":       name,
		"short_name": shortName,
		"color":      color,
	}
	base_service.Create(tx, "status", params)
	status := models.Status{}
	err := base_service.GetByName(tx, "status", name, &status)
	if err != nil {
		return status, err
	}
	return status, nil
}

func GetStatus(tx *sqlx.Tx, id string) (models.Status, error) {
	status := models.Status{}
	err := base_service.Get(tx, "status", id, &status)
	if err != nil {
		return status, err
	}
	return status, nil
}

func GetStatuses(tx *sqlx.Tx) ([]models.Status, error) {
	statuses := []models.Status{}
	err := base_service.GetAll(tx, "status", &statuses)
	if err != nil {
		return statuses, err
	}
	return statuses, nil
}

func GetStatusByShortName(tx *sqlx.Tx, shortName string) (models.Status, error) {
	status := models.Status{}
	conditions := map[string]interface{}{
		"short_name": shortName,
	}
	err := base_service.GetBy(tx, "status", conditions, &status)
	if err != nil {
		return status, err
	}
	return status, nil
}

func GetOrCreateStatus(tx *sqlx.Tx, name string, shortName string, color string) (models.Status, error) {
	//TODO investigate all GetOrCreate
	status, err := GetStatusByShortName(tx, shortName)
	if err == nil {
		return status, nil
	}
	status, err = CreateStatus(tx, "", name, shortName, color)
	if err != nil {
		return status, err
	}
	return status, nil
}

func Updatestatus(tx *sqlx.Tx, taskId string, statusId string) error {
	params := map[string]interface{}{
		"status_id": statusId,
	}
	err := base_service.Update(tx, "task", taskId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "task", taskId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}
