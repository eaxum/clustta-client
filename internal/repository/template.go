package repository

import (
	"database/sql"
	"path/filepath"

	"os"

	"clustta/internal/base_service"
	"clustta/internal/error_service"
	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func CreateTemplate(tx *sqlx.Tx, name string, templateFile string) (models.Template, error) {
	template := models.Template{}
	extension := filepath.Ext(templateFile)
	fileStat, err := os.Stat(templateFile)
	if err != nil {
		return template, err
	}
	fileSize := int(fileStat.Size())
	xxhashChecksum, err := utils.GenerateXXHashChecksum(templateFile)
	if err != nil {
		return template, err
	}

	// chunks := "test,test,test"
	chunks, err := StoreFileChunks(tx, templateFile, func(current int, total int, message string, extraMessage string) {})
	if err != nil {
		return template, err
	}

	params := map[string]interface{}{
		"name":            name,
		"extension":       extension,
		"xxhash_checksum": xxhashChecksum,
		"file_size":       fileSize,
		"chunks":          chunks,
	}
	err = base_service.Create(tx, "template", params)
	if err != nil {
		return template, err
	}
	err = base_service.GetByName(tx, "template", name, &template)
	if err != nil {
		return template, err
	}
	return template, nil
}

func AddTemplate(tx *sqlx.Tx, id, name, extension, chunks, xxhashChecksum string, fileSize int) (models.Template, error) {
	template := models.Template{}
	params := map[string]interface{}{
		"id":              id,
		"name":            name,
		"extension":       extension,
		"xxhash_checksum": xxhashChecksum,
		"file_size":       fileSize,
		"chunks":          chunks,
	}
	err := base_service.Create(tx, "template", params)
	if err != nil {
		return template, err
	}
	err = base_service.GetByName(tx, "template", name, &template)
	if err != nil {
		return template, err
	}
	return template, nil
}

func GetTemplate(tx *sqlx.Tx, id string) (models.Template, error) {
	template := models.Template{}
	err := base_service.Get(tx, "template", id, &template)
	if err != nil && err == sql.ErrNoRows {
		return models.Template{}, error_service.ErrTemplateNotFound
	} else if err != nil {
		return models.Template{}, err
	}
	return template, nil
}

func GetTemplates(tx *sqlx.Tx, withDeleted bool) ([]models.Template, error) {
	templates := []models.Template{}

	if withDeleted {
		err := base_service.GetAll(tx, "template", &templates)
		if err != nil {
			return templates, err
		}
	} else {
		conditions := map[string]interface{}{
			"trashed": 0,
		}
		err := base_service.GetAllBy(tx, "template", conditions, &templates)
		if err != nil {
			return templates, err
		}
	}

	return templates, nil
}
func GetDeletedTemplates(tx *sqlx.Tx) ([]models.Template, error) {
	templates := []models.Template{}

	conditions := map[string]interface{}{
		"trashed": 1,
	}
	err := base_service.GetAllBy(tx, "template", conditions, &templates)
	if err != nil {
		return templates, err
	}

	return templates, err
}

func GetTemplateByName(tx *sqlx.Tx, name string) (models.Template, error) {
	template := models.Template{}
	err := base_service.GetByName(tx, "template", name, &template)
	if err != nil {
		return template, err
	}
	err = base_service.GetByName(tx, "template", name, &template)
	if err != nil {
		return template, err
	}
	return template, nil
}

func GetOrCreateTemplate(tx *sqlx.Tx, name string, templateFile string) (models.Template, error) {
	template, err := GetTemplateByName(tx, name)
	if err == nil {
		return template, nil
	}
	template, err = CreateTemplate(tx, name, templateFile)
	if err != nil {
		return template, err
	}
	return template, nil
}

func DeleteTemplate(tx *sqlx.Tx, id string, recycle bool) error {
	if recycle {
		err := base_service.MarkAsDeleted(tx, "template", id)
		if err != nil {
			return err
		}
	} else {
		err := base_service.Delete(tx, "template", id)
		if err != nil {
			return err
		}
	}
	return nil
}

func RenameTemplate(tx *sqlx.Tx, id string, name string) error {
	err := base_service.Rename(tx, "template", id, name)
	return err
}

func UpdateTemplateFile(tx *sqlx.Tx, id string, templateFile string) (models.Template, error) {
	template, err := GetTemplate(tx, id)
	if err != nil {
		return template, err
	}
	if template.Name == "" {
		return template, nil
	}

	extension := filepath.Ext(templateFile)
	fileStat, err := os.Stat(templateFile)
	if err != nil {
		return template, err
	}
	fileSize := int(fileStat.Size())
	xxhashChecksum, err := utils.GenerateXXHashChecksum(templateFile)
	if err != nil {
		return template, err
	}

	chunks, err := StoreFileChunks(tx, templateFile, func(current int, total int, message string, extraMessage string) {})
	if err != nil {
		return template, err
	}
	params := map[string]interface{}{
		"extension":       extension,
		"xxhash_checksum": xxhashChecksum,
		"file_size":       fileSize,
		"chunks":          chunks,
	}
	err = base_service.Update(tx, "template", id, params)
	if err != nil {
		return template, err
	}
	err = base_service.UpdateMtime(tx, "template", id, utils.GetEpochTime())
	if err != nil {
		return template, err
	}

	template, err = GetTemplate(tx, id)
	if err != nil {
		return template, err
	}
	return template, nil
}
