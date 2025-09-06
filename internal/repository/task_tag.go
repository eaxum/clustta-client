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

func AddTagToTask(tx *sqlx.Tx, taskId string, tag string) error {
	tagObj, err := GetOrCreateTag(tx, tag)
	if err != nil {
		return err
	}
	params := map[string]interface{}{
		"task_id": taskId,
		"tag_id":  tagObj.Id,
	}
	err = base_service.Create(tx, "task_tag", params)
	if err != nil {
		return err
	}
	return nil
}

func AddTagToTaskById(tx *sqlx.Tx, id, taskId string, tagId string) error {
	params := map[string]interface{}{
		"id":      id,
		"task_id": taskId,
		"tag_id":  tagId,
	}
	err := base_service.Create(tx, "task_tag", params)
	if err != nil {
		return err
	}
	return nil
}

func GetTaskTag(tx *sqlx.Tx, Id string) (models.TaskTag, error) {
	taskTag := models.TaskTag{}
	err := base_service.Get(tx, "task_tag", Id, &taskTag)
	if err != nil {
		return taskTag, err
	}
	return taskTag, nil
}

func GetTaskTags(tx *sqlx.Tx, taskId string) ([]models.Tag, error) {
	tags := []models.Tag{}
	err := tx.Select(&tags, "SELECT * FROM tag WHERE id IN (SELECT tag_id FROM task_tag WHERE task_id = ?)", taskId)
	if err != nil {
		return tags, err
	}
	return tags, nil
}

func RemoveTagFromTask(tx *sqlx.Tx, taskId string, tagId string) error {
	conditions := map[string]interface{}{
		"task_id": taskId,
		"tag_id":  tagId,
	}
	err := base_service.DeleteBy(tx, "task_tag", conditions)
	return err
}

func RemoveAllTagsFromTask(tx *sqlx.Tx, taskId string) error {
	conditions := map[string]interface{}{
		"task_id": taskId,
	}
	err := base_service.DeleteBy(tx, "task_tag", conditions)
	return err
}

// func GetTaskTagsByTagId(tx *sqlx.Tx, tagId string) []Task {
// 	dbConn, err := utils.OpenDb( projectPath)
// 	if err != nil {
// 		panic(err)
// 	}
// 	tasks := []Task{}
// 	tx.Select(&tasks, "SELECT * FROM task WHERE id IN (SELECT task_id FROM task_tag WHERE tag_id = ?)", tagId)
// 	return tasks
// }

// func GetTaskTagsByTagName(tx *sqlx.Tx, tagName string) []Task {
// 	dbConn, err := utils.OpenDb( projectPath)
// 	if err != nil {
// 		panic(err)
// 	}
// 	tasks := []Task{}
// 	tx.Select(&tasks, "SELECT * FROM task WHERE id IN (SELECT task_id FROM task_tag WHERE tag_id IN (SELECT id FROM tag WHERE name = ?))", tagName)
// 	return tasks
// }

// func GetTaskTagsByTaskId(tx *sqlx.Tx, taskId string) []models.Tag {
// 	dbConn, err := utils.OpenDb( projectPath)
// 	if err != nil {
// 		panic(err)
// 	}
// 	tags := []models.Tag{}
// 	tx.Select(&tags, "SELECT * FROM tag WHERE id IN (SELECT tag_id FROM task_tag WHERE task_id = ?)", taskId)
// 	return tags
// }
