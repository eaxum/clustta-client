package repository

import (
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func BatchCreateEntities(tx *sqlx.Tx, entities []models.Entity) error {
	if len(entities) == 0 {
		return nil
	}

	// Build bulk insert query
	valueStrings := make([]string, 0, len(entities))
	valueArgs := make([]interface{}, 0, len(entities)*8)

	for _, entity := range entities {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			entity.Id,
			entity.Name,
			entity.Description,
			entity.EntityTypeId,
			entity.ParentId,
			entity.PreviewId,
			entity.IsLibrary,
			utils.GetEpochTime(), // mtime
		)
	}

	stmt := fmt.Sprintf(`
		INSERT INTO entity (id, name, description, entity_type_id, parent_id, preview_id, is_library, mtime)
		VALUES %s
	`, strings.Join(valueStrings, ","))

	_, err := tx.Exec(stmt, valueArgs...)
	return err
}

// BatchUpdateEntities updates multiple entities in a single transaction
func BatchUpdateEntities(tx *sqlx.Tx, entities []models.Entity) error {
	if len(entities) == 0 {
		return nil
	}

	// Use CASE statements for bulk updates
	ids := make([]string, len(entities))
	nameMap := make(map[string]string)
	parentMap := make(map[string]string)
	previewMap := make(map[string]string)
	libraryMap := make(map[string]bool)

	for i, entity := range entities {
		ids[i] = entity.Id
		nameMap[entity.Id] = entity.Name
		parentMap[entity.Id] = entity.ParentId
		previewMap[entity.Id] = entity.PreviewId
		libraryMap[entity.Id] = entity.IsLibrary
	}

	// Build CASE statements
	nameCases := make([]string, 0, len(entities))
	parentCases := make([]string, 0, len(entities))
	previewCases := make([]string, 0, len(entities))
	libraryCases := make([]string, 0, len(entities))

	for _, entity := range entities {
		nameCases = append(nameCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", entity.Id, entity.Name))
		parentCases = append(parentCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", entity.Id, entity.ParentId))
		previewCases = append(previewCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", entity.Id, entity.PreviewId))
		libraryCases = append(libraryCases, fmt.Sprintf("WHEN id = '%s' THEN %t", entity.Id, entity.IsLibrary))
	}

	query := fmt.Sprintf(`
		UPDATE entity SET 
			name = CASE %s END,
			parent_id = CASE %s END,
			preview_id = CASE %s END,
			is_library = CASE %s END,
			mtime = %d
		WHERE id IN ('%s')
	`,
		strings.Join(nameCases, " "),
		strings.Join(parentCases, " "),
		strings.Join(previewCases, " "),
		strings.Join(libraryCases, " "),
		utils.GetEpochTime(),
		strings.Join(ids, "','"),
	)

	_, err := tx.Exec(query)
	return err
}

// BatchCreateTasks creates multiple tasks in a single transaction
func BatchCreateTasks(tx *sqlx.Tx, tasks []models.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(tasks))
	valueArgs := make([]interface{}, 0, len(tasks)*12)

	for _, task := range tasks {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			task.Id,
			task.CreatedAt,
			task.Name,
			task.TaskTypeId,
			task.EntityId,
			task.StatusId,
			task.Extension,
			task.Description,
			task.Pointer,
			task.IsLink,
			task.AssigneeId,
			task.PreviewId,
		)
	}

	stmt := fmt.Sprintf(`
		INSERT INTO tasks (id, created_at, name, task_type_id, entity_id, status_id, extension, description, pointer, is_link, assignee_id, preview_id)
		VALUES %s
	`, strings.Join(valueStrings, ","))

	_, err := tx.Exec(stmt, valueArgs...)
	return err
}

// BatchUpdateTasks updates multiple tasks in a single transaction
func BatchUpdateTasks(tx *sqlx.Tx, tasks []models.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	ids := make([]string, len(tasks))
	for i, task := range tasks {
		ids[i] = task.Id
	}

	// Build CASE statements for each field that needs updating
	nameCases := make([]string, 0, len(tasks))
	entityCases := make([]string, 0, len(tasks))
	typeCases := make([]string, 0, len(tasks))
	assigneeCases := make([]string, 0, len(tasks))
	statusCases := make([]string, 0, len(tasks))
	previewCases := make([]string, 0, len(tasks))
	resourceCases := make([]string, 0, len(tasks))
	linkCases := make([]string, 0, len(tasks))
	pointerCases := make([]string, 0, len(tasks))

	for _, task := range tasks {
		nameCases = append(nameCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", task.Id, task.Name))
		entityCases = append(entityCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", task.Id, task.EntityId))
		typeCases = append(typeCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", task.Id, task.TaskTypeId))
		assigneeCases = append(assigneeCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", task.Id, task.AssigneeId))
		statusCases = append(statusCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", task.Id, task.StatusId))
		previewCases = append(previewCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", task.Id, task.PreviewId))
		resourceCases = append(resourceCases, fmt.Sprintf("WHEN id = '%s' THEN %t", task.Id, task.IsResource))
		linkCases = append(linkCases, fmt.Sprintf("WHEN id = '%s' THEN %t", task.Id, task.IsLink))
		pointerCases = append(pointerCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", task.Id, task.Pointer))
	}

	query := fmt.Sprintf(`
		UPDATE tasks SET 
			name = CASE %s END,
			entity_id = CASE %s END,
			task_type_id = CASE %s END,
			assignee_id = CASE %s END,
			status_id = CASE %s END,
			preview_id = CASE %s END,
			is_resource = CASE %s END,
			is_link = CASE %s END,
			pointer = CASE %s END,
			mtime = %d
		WHERE id IN ('%s')
	`,
		strings.Join(nameCases, " "),
		strings.Join(entityCases, " "),
		strings.Join(typeCases, " "),
		strings.Join(assigneeCases, " "),
		strings.Join(statusCases, " "),
		strings.Join(previewCases, " "),
		strings.Join(resourceCases, " "),
		strings.Join(linkCases, " "),
		strings.Join(pointerCases, " "),
		utils.GetEpochTime(),
		strings.Join(ids, "','"),
	)

	_, err := tx.Exec(query)
	return err
}

// BatchCreateCheckpoints creates multiple task checkpoints in a single transaction
func BatchCreateCheckpoints(tx *sqlx.Tx, checkpoints []models.Checkpoint) error {
	if len(checkpoints) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(checkpoints))
	valueArgs := make([]interface{}, 0, len(checkpoints)*10)

	for _, checkpoint := range checkpoints {
		epochTime, err := utils.RFC3339ToEpoch(checkpoint.CreatedAt)
		if err != nil {
			return err
		}

		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			checkpoint.Id,
			epochTime,
			checkpoint.TaskId,
			checkpoint.XXHashChecksum,
			checkpoint.TimeModified,
			checkpoint.FileSize,
			checkpoint.Comment,
			checkpoint.Chunks,
			checkpoint.AuthorUID,
			checkpoint.PreviewId,
		)
	}

	stmt := fmt.Sprintf(`
		INSERT INTO task_checkpoints (id, created_at, task_id, xxhash_checksum, time_modified, file_size, comment, chunks, author_uid, preview_id)
		VALUES %s
	`, strings.Join(valueStrings, ","))

	_, err := tx.Exec(stmt, valueArgs...)
	return err
}

// GetAllEntities retrieves all entities for batch processing
func GetAllEntities(tx *sqlx.Tx) ([]models.Entity, error) {
	var entities []models.Entity
	query := `
		SELECT id, name, description, entity_type_id, parent_id, preview_id, is_library, mtime
		FROM entity
	`
	err := tx.Select(&entities, query)
	return entities, err
}

// Alternative approach using prepared statements for very large datasets
func BatchCreateEntitiesWithPreparedStmt(tx *sqlx.Tx, entities []models.Entity) error {
	if len(entities) == 0 {
		return nil
	}

	stmt, err := tx.Preparex(`
		INSERT INTO entity (id, name, description, entity_type_id, parent_id, preview_id, is_library, mtime)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, entity := range entities {
		_, err := stmt.Exec(
			entity.Id,
			entity.Name,
			entity.Description,
			entity.EntityTypeId,
			entity.ParentId,
			entity.PreviewId,
			entity.IsLibrary,
			utils.GetEpochTime(),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
