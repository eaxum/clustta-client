package base_service

import (
	"clustta/internal/error_service"
	"clustta/internal/utils"
	"database/sql"
	"fmt"
	"strings"

	// "strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Table interface {
	TableName() string
	PrimaryKeyName() string
}

// Get retrieves a record by ID
func Get(tx *sqlx.Tx, table string, id string, dest interface{}) error {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", table)
	err := tx.Get(dest, query, id)
	if err != nil && err == sql.ErrNoRows {
		switch table {
		case "status":
			return error_service.ErrStatusNotFound
		case "task_type":
			return error_service.ErrTaskTypeNotFound
		case "full_task":
			return error_service.ErrTaskNotFound
		case "task":
			return error_service.ErrTaskNotFound
		case "task_checkpoint":
			return error_service.ErrTaskCheckPointNotFound

		case "user":
			return error_service.ErrUserNotFound
		case "role":
			return error_service.ErrRoleNotFound
		case "dependency_type":
			return error_service.ErrDependencyTypeNotFound
		case "entity":
			return error_service.ErrEntityNotFound
		case "entity_assignee":
			return error_service.ErrEntityAssigneeNotFound
		case "full_entity":
			return error_service.ErrEntityNotFound
		case "entity_type":
			return error_service.ErrEntityTypeNotFound
		case "template":
			return error_service.ErrTemplateNotFound
		case "workflow":
			return error_service.ErrWorkflowNotFound
		case "workflow_entity":
			return error_service.ErrWorkflowEntityNotFound
		case "workflow_task":
			return error_service.ErrWorkflowTaskNotFound
		case "tag":
			return error_service.ErrTagNotFound
		case "task_tag":
			return error_service.ErrTaskTagNotFound
		case "task_dependency":
			return error_service.ErrTaskDependencyNotFound
		case "entity_dependency":
			return error_service.ErrEntityDependencyNotFound
		case "preview":
			return error_service.ErrPreviewNotFound
		// case "subtask_dependency":
		// 	return error_service.ErrSubtaskDe
		default:
			return fmt.Errorf("id of %s not found in %s", id, table)
		}
	} else if err != nil {
		return err
	}
	return nil
}

func GetByName(tx *sqlx.Tx, table string, name string, dest interface{}) error {
	query := fmt.Sprintf("SELECT * FROM %s WHERE name = ?", table)
	err := tx.Get(dest, query, name)
	if err != nil && err == sql.ErrNoRows {
		switch table {
		case "status":
			return error_service.ErrStatusNotFound
		case "task_type":
			return error_service.ErrTaskTypeNotFound
		case "full_task":
			return error_service.ErrTaskNotFound
		case "task":
			return error_service.ErrTaskNotFound
		case "task_checkpoint":
			return error_service.ErrTaskCheckPointNotFound

		case "user":
			return error_service.ErrUserNotFound
		case "role":
			return error_service.ErrRoleNotFound
		case "dependency_type":
			return error_service.ErrDependencyTypeNotFound
		case "entity":
			return error_service.ErrEntityNotFound
		case "entity_assignee":
			return error_service.ErrEntityAssigneeNotFound
		case "full_entity":
			return error_service.ErrEntityNotFound
		case "entity_type":
			return error_service.ErrEntityTypeNotFound
		case "template":
			return error_service.ErrTemplateNotFound
		case "workflow":
			return error_service.ErrWorkflowNotFound
		case "workflow_entity":
			return error_service.ErrWorkflowEntityNotFound
		case "workflow_task":
			return error_service.ErrWorkflowTaskNotFound
		case "tag":
			return error_service.ErrTagNotFound
		case "task_tag":
			return error_service.ErrTaskTagNotFound
		case "task_dependency":
			return error_service.ErrTaskDependencyNotFound
		case "entity_dependency":
			return error_service.ErrEntityDependencyNotFound
		case "preview":
			return error_service.ErrPreviewNotFound
		default:
			return fmt.Errorf("name of %s not found in %s", name, table)
		}
	} else if err != nil {
		return err
	}
	return nil
}

func GetByNameCaseInsensitive(tx *sqlx.Tx, table string, name string, dest interface{}) error {
	query := fmt.Sprintf("SELECT * FROM %s WHERE LOWER(name) = ?", table)
	err := tx.Select(dest, query, strings.ToLower(name))
	if err != nil {
		return err
	}
	return nil
}

// GetAll retrieves all records from the table
func GetAll(tx *sqlx.Tx, table string, dest interface{}) error {
	query := fmt.Sprintf("SELECT * FROM %s", table)
	err := tx.Select(dest, query)
	// if err != nil && err == sql.ErrNoRows {
	// 	return fmt.Errorf(" nothing found in %s", table)
	// } else if err != nil {
	// 	return err
	// }
	if err != nil {
		return err
	}
	return nil
}

func Create(tx *sqlx.Tx, table string, params map[string]interface{}) error {
	var columns []string
	var placeholders []string
	var values []any

	idProvided := false
	mtimeProvided := false
	for column, value := range params {
		if column == "id" && value != "" {
			idProvided = true
			columns = append(columns, column)
			placeholders = append(placeholders, "?")
			values = append(values, value)
		} else if column == "id" && value == "" {
			idProvided = false
		} else {
			columns = append(columns, column)
			placeholders = append(placeholders, "?")
			values = append(values, value)
		}

		if column == "mtime" && value != 0 {
			mtimeProvided = true
		} else if column == "mtime" && value == 0 {
			mtimeProvided = false
		}
	}

	if !idProvided {
		id := uuid.New().String()
		columns = append(columns, "id")
		placeholders = append(placeholders, "?")
		values = append(values, id)
	}
	if !mtimeProvided {
		mtime := utils.GetEpochTime()
		columns = append(columns, "mtime")
		placeholders = append(placeholders, "?")
		values = append(values, mtime)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := tx.Exec(query, values...)
	if err != nil {
		return err
	}

	return nil
}

func UpdateMtime(tx *sqlx.Tx, table string, id string, mtime int64) error {
	if mtime == 0 {
		mtime = utils.GetEpochTime()
	}
	query := fmt.Sprintf("UPDATE %s SET mtime = ? WHERE id = '%s'", table, id)
	_, err := tx.Exec(query, mtime)
	if err != nil {
		return err
	}
	return nil
}

func Update(tx *sqlx.Tx, table string, id string, params map[string]interface{}) error {
	//TODO add UpdateMtime to here
	var setClauses []string
	var values []interface{}

	for column, value := range params {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", column))
		values = append(values, value)
	}
	setClause := strings.Join(setClauses, ", ")
	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = '%s'", table, setClause, id)
	_, err := tx.Exec(query, values...)
	if err != nil {
		return err
	}
	return nil
}

func Rename(tx *sqlx.Tx, table string, id string, newName string) error {
	query := fmt.Sprintf("UPDATE %s SET name = ? WHERE id = '%s'", table, id)
	_, err := tx.Exec(query, newName)
	if err != nil {
		return err
	}
	return nil
}

func UpdateBy(tx *sqlx.Tx, table string, conditions map[string]interface{}, params map[string]interface{}) error {
	var setClauses []string
	var values []interface{}

	for column, value := range params {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", column))
		values = append(values, value)
	}

	var whereClauses []string
	for column, value := range conditions {
		if column == "id" {
			continue
		}
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", column))
		values = append(values, value)
	}

	setClause := strings.Join(setClauses, ", ")
	whereClause := strings.Join(whereClauses, " AND ")

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, setClause, whereClause)
	_, err := tx.Exec(query, values...)
	if err != nil {
		return err
	}
	return nil
}

func AddToTomb(tx *sqlx.Tx, table string, id string) error {
	tombQuery := "INSERT INTO tomb (id, table_name) VALUES (?, ?)"
	_, err := tx.Exec(tombQuery, id, table)
	if err != nil {
		return err
	}
	return nil
}
func Delete(tx *sqlx.Tx, table string, id string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
	_, err := tx.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}

func XDeleteBy(tx *sqlx.Tx, table string, conditions map[string]interface{}) error {
	var whereClauses []string
	var values []interface{}
	for column, value := range conditions {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", column))
		values = append(values, value)
	}
	whereClause := strings.Join(whereClauses, " AND ")

	query := fmt.Sprintf("SELECT id FROM %s WHERE %s", table, whereClause)
	rows, err := tx.Queryx(query, values...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return err
		}

		err = Delete(tx, table, id)
		if err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func DeleteBy(tx *sqlx.Tx, table string, conditions map[string]interface{}) error {
	var whereClauses []string
	var values []interface{}
	for column, value := range conditions {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", column))
		values = append(values, value)
	}
	whereClause := strings.Join(whereClauses, " AND ")

	query := fmt.Sprintf("DELETE FROM %s WHERE %s", table, whereClause)
	_, err := tx.Exec(query, values...)
	if err != nil {
		return err
	}
	return nil
}

func MarkAsDeleted(tx *sqlx.Tx, table string, id string) error {
	params := map[string]interface{}{
		"trashed": true,
	}
	err := Update(tx, table, id, params)
	if err != nil {
		return nil
	}
	return nil
}

func Restore(tx *sqlx.Tx, table string, id string) error {
	params := map[string]interface{}{
		"trashed": false,
	}
	err := Update(tx, table, id, params)
	if err != nil {
		return nil
	}
	return nil
}

// GetAllBy retrieves all records by multiple column and value conditions
func GetAllBy(tx *sqlx.Tx, table string, conditions map[string]interface{}, dest interface{}) error {

	// Build the WHERE clause
	var whereClauses []string
	var values []interface{}
	for column, value := range conditions {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", column))
		values = append(values, value)
	}
	whereClause := strings.Join(whereClauses, " AND ")

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", table, whereClause)
	err := tx.Select(dest, query, values...)
	if err != nil {
		return err
	}

	return nil
}

// GetBy retrieves a record by multiple column and value conditions
func GetBy(tx *sqlx.Tx, table string, conditions map[string]interface{}, dest interface{}) error {

	// Build the WHERE clause
	var whereClauses []string
	var values []interface{}
	for column, value := range conditions {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", column))
		values = append(values, value)
	}
	whereClause := strings.Join(whereClauses, " AND ")

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", table, whereClause)
	err := tx.Get(dest, query, values...)
	if err != nil {
		return err
	}
	return nil
}

// params := map[string]interface{}{
// 	"column1": value1,
// 	"column2": value2,
// }

// GetByCaseInsensitive retrieves a record by multiple column and value conditions case insensitively
func GetByCaseInsensitive(tx *sqlx.Tx, table string, conditions map[string]interface{}, dest interface{}) error {

	// Build the WHERE clause
	var whereClauses []string
	var values []interface{}
	for column, value := range conditions {
		whereClauses = append(whereClauses, fmt.Sprintf("LOWER(%s) = ?", column))
		values = append(values, strings.ToLower(fmt.Sprintf("%v", value)))
	}
	whereClause := strings.Join(whereClauses, " AND ")

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", table, whereClause)
	err := tx.Select(dest, query, values...)
	if err != nil {
		return err
	}
	return nil
}

// GetAllByCaseInsensitive retrieves all records by multiple column and value conditions case insensitively
func GetAllByCaseInsensitive(tx *sqlx.Tx, table string, conditions map[string]interface{}, dest interface{}) error {

	// Build the WHERE clause
	var whereClauses []string
	var values []interface{}
	for column, value := range conditions {
		whereClauses = append(whereClauses, fmt.Sprintf("LOWER(%s) = ?", column))
		values = append(values, strings.ToLower(fmt.Sprintf("%v", value)))
	}
	whereClause := strings.Join(whereClauses, " AND ")

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", table, whereClause)
	err := tx.Select(dest, query, values...)
	if err != nil {
		return err
	}

	return nil
}
