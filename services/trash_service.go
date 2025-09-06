package services

import (
	"clustta/internal/base_service"
	"clustta/internal/repository"
	"clustta/internal/utils"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type RecycleItem struct {
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Id         string      `json:"id"`
	ParentId   string      `json:"parent_id"`
	ParentName string      `json:"parent_name"`
	Data       interface{} `json:"data"`
}

type TrashService struct{}

func (t *TrashService) GetTrashs(projectPath string) ([]RecycleItem, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return []RecycleItem{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []RecycleItem{}, err
	}
	defer tx.Rollback()

	recycleItems := []RecycleItem{}

	entities, err := repository.GetDeletedEntities(tx)
	if err != nil {
		return []RecycleItem{}, err
	}
	for _, entity := range entities {
		recycleItem := RecycleItem{
			Name: entity.Name,
			Type: "entity",
			Id:   entity.Id,
		}
		recycleItems = append(recycleItems, recycleItem)

	}

	//template
	templates, err := repository.GetDeletedTemplates(tx)
	if err != nil {
		return []RecycleItem{}, err
	}
	for _, template := range templates {
		recycleItem := RecycleItem{
			Name: template.Name,
			Type: "template",
			Id:   template.Id,
		}
		recycleItems = append(recycleItems, recycleItem)

	}

	//task

	tasks, err := repository.GetDeletedTasks(tx)
	if err != nil {
		return []RecycleItem{}, err
	}
	for _, task := range tasks {
		recycleItem := RecycleItem{
			Name:       task.Name,
			Type:       "task",
			Id:         task.Id,
			ParentId:   task.EntityId,
			ParentName: task.EntityName,
		}
		recycleItems = append(recycleItems, recycleItem)

	}

	taskCheckpoints, err := repository.GetDeletedCheckpoints(tx)
	if err != nil {
		return []RecycleItem{}, err
	}
	for _, taskCheckpoint := range taskCheckpoints {
		checkpointTask, err := repository.GetTask(tx, taskCheckpoint.TaskId)
		if err != nil {
			return []RecycleItem{}, err
		}
		checkpointName := fmt.Sprintf("%s %s", checkpointTask.Name, taskCheckpoint.CreatedAt)
		task, err := repository.GetTask(tx, taskCheckpoint.TaskId)
		if err != nil {
			return []RecycleItem{}, err
		}
		recycleItem := RecycleItem{
			Name:       checkpointName,
			Type:       "task_checkpoint",
			Id:         taskCheckpoint.Id,
			ParentId:   taskCheckpoint.TaskId,
			ParentName: task.Name,
		}
		recycleItems = append(recycleItems, recycleItem)

	}
	return recycleItems, nil
}

func (t *TrashService) Restore(projectPath, id, itemType string) error {
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
	err = base_service.Restore(tx, itemType, id)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
