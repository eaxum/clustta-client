package services

import (
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"errors"
	"os"

	"github.com/jmoiron/sqlx"
)

type TemplateService struct{}

func (t *TemplateService) GetTemplates(projectPath string) ([]models.Template, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return []models.Template{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Template{}, err
	}
	defer tx.Rollback()

	templates, err := repository.GetTemplates(tx, true)
	if err != nil {
		return []models.Template{}, err
	}
	return templates, nil
}

func (t *TemplateService) ChangeTemplateFile(projectPath, templateName, filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.New("file does not exist")
	}

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

	template, err := repository.GetTemplateByName(tx, templateName)
	if err != nil {
		return err
	}

	_, err = repository.UpdateTemplateFile(tx, template.Id, filePath)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (t *TemplateService) CreateTemplate(projectPath, templateName, filePath string) (models.Template, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return models.Template{}, errors.New("file does not exist")
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Template{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Template{}, err
	}
	defer tx.Rollback()

	template, err := repository.CreateTemplate(tx, templateName, filePath)
	if err != nil {
		tx.Rollback()
		return models.Template{}, err
	}
	tx.Commit()

	return template, nil
}

func (t *TemplateService) DeleteTemplate(projectPath, templateName string) error {
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

	template, err := repository.GetTemplateByName(tx, templateName)
	if err != nil {
		return err
	}
	err = repository.DeleteTemplate(tx, template.Id, true)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	return nil
}
func (t *TemplateService) RenameTemplate(projectPath, templateName, newName string) error {
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

	template, err := repository.GetTemplateByName(tx, templateName)
	if err != nil {
		return err
	}

	err = repository.RenameTemplate(tx, template.Id, newName)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (t *TemplateService) GetTemplate(projectPath, templateId string) (models.Template, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Template{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Template{}, err
	}
	defer tx.Rollback()

	template, err := repository.GetTemplate(tx, templateId)
	if err != nil {
		return models.Template{}, err
	}

	return template, nil
}
