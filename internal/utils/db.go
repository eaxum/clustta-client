package utils

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func SplitStatements(content string) []string {
	var statements []string
	var currentStmt strings.Builder
	var inQuote bool
	var quoteChar rune
	var inTrigger bool

	// Remove comments and handle line endings
	lines := strings.Split(content, "\n")
	var cleanContent strings.Builder
	for _, line := range lines {
		// Remove inline comments
		if idx := strings.Index(line, "--"); idx != -1 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		cleanContent.WriteString(line + "\n")
	}

	content = cleanContent.String()

	for i, char := range content {
		currentStmt.WriteRune(char)

		// Handle quotes
		if char == '\'' || char == '"' || char == '`' {
			if !inQuote {
				inQuote = true
				quoteChar = char
			} else if quoteChar == char {
				// Check if it's an escaped quote
				if i > 0 && content[i-1] != '\\' {
					inQuote = false
				}
			}
		}

		// Check for CREATE TRIGGER statement
		if !inQuote && !inTrigger &&
			strings.Contains(strings.ToUpper(currentStmt.String()), "CREATE TRIGGER") {
			inTrigger = true
		}

		// Handle statement endings
		if char == ';' && !inQuote {
			// If we're in a trigger, only end if we have END; (case insensitive)
			if inTrigger {
				stmt := strings.TrimSpace(currentStmt.String())
				upperStmt := strings.ToUpper(stmt)
				if strings.HasSuffix(upperStmt, "END;") {
					statements = append(statements, stmt)
					currentStmt.Reset()
					inTrigger = false
				}
			} else {
				stmt := strings.TrimSpace(currentStmt.String())
				if stmt != ";" {
					statements = append(statements, stmt)
				}
				currentStmt.Reset()
			}
		}
	}

	// Add any remaining statement
	if remaining := strings.TrimSpace(currentStmt.String()); remaining != "" {
		statements = append(statements, remaining)
	}

	return statements
}

// ColumnInfo holds the schema information for a column
type ColumnInfo struct {
	Cid       int            `db:"cid"`
	Name      string         `db:"name"`
	Type      string         `db:"type"`
	NotNull   int            `db:"notnull"`
	DfltValue sql.NullString `db:"dflt_value"`
	Pk        int            `db:"pk"`
}

// renameColumn renames a column in the SQLite database.
func RenameColumn(db *sqlx.DB, tableName, oldColumnName, newColumnName string) error {
	// Start a transaction
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Defer a rollback in case anything fails.
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Check if the old column exists
	exists, err := IsColumnExist(db, tableName, oldColumnName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
		// return fmt.Errorf("column %s does not exist", oldColumnName)
	}

	renameColumnSQL := `ALTER TABLE ` + tableName + ` RENAME COLUMN ` + oldColumnName + ` TO ` + newColumnName + `;`

	// Execute the SQL command
	_, err = db.Exec(renameColumnSQL)
	if err != nil {
		return err
	}
	return nil
}

func DeleteColumn(db *sqlx.DB, tableName, columnName string) error {
	deleteColumnSQL := `ALTER TABLE ` + tableName + ` DROP  COLUMN ` + columnName + `;`
	// Execute the SQL command
	_, err := db.Exec(deleteColumnSQL)
	if err != nil {
		return err
	}
	return nil
}

func DeleteColumnIfExists(db *sqlx.DB, tableName, columnName string) error {
	// Check if the column exists
	exists, err := IsColumnExist(db, tableName, columnName)
	if err != nil {
		return err
	}

	// If the column doesn't exist, return early
	if !exists {
		return nil
	}

	err = DeleteColumn(db, tableName, columnName)
	if err != nil {
		return err
	}
	return nil
}

func IsColumnExist(db *sqlx.DB, tableName, columnName string) (bool, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s);", tableName)

	var columns []ColumnInfo
	if err := db.Select(&columns, query); err != nil {
		return false, fmt.Errorf("failed to query table info: %w", err)
	}

	var exists bool
	for _, column := range columns {
		if column.Name == columnName {
			exists = true
			break
		}
	}
	return exists, nil
}

func AddColumnIfNotExist(db *sqlx.DB, tableName, columnName, columnType, defaultValue string, nullable bool) error {
	exists, err := IsColumnExist(db, tableName, columnName)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	// Construct the column definition
	var columnDef string
	if nullable {
		if defaultValue == "" {
			columnDef = fmt.Sprintf("%s %s", columnName, columnType)
		} else {
			columnDef = fmt.Sprintf("%s %s DEFAULT %s", columnName, columnType, defaultValue)
		}
	} else {
		if defaultValue == "" {
			columnDef = fmt.Sprintf("%s %s DEFAULT '' NOT NULL", columnName, columnType)
		} else {
			columnDef = fmt.Sprintf("%s %s NOT NULL DEFAULT %s", columnName, columnType, defaultValue)
		}
	}

	// Add the column if it does not exist
	alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;", tableName, columnDef)

	// Add the column if it does not exist
	// alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s DEFAULT %s;", tableName, columnName, columnType, defaultValue)
	_, err = db.Exec(alterQuery)
	if err != nil {
		return fmt.Errorf("failed to add column: %w", err)
	}
	return nil
}

func TableExists(db *sqlx.DB, tableName string) (bool, error) {
	// var name string
	// query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?;`
	// err := db.Get(&name, query, tableName)
	// if err != nil {
	// 	if errors.Is(err, sqlx.ErrNotFound) {
	// 		return false, nil
	// 	}
	// 	return false, err
	// }
	// return name == tableName, nil

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tableName)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return false, nil
	}
	return true, nil
}

func RenameTable(db *sqlx.DB, oldName, newName string) error {
	exists, err := TableExists(db, oldName)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}
	// SQL command to rename the table
	renameTableSQL := `ALTER TABLE ` + oldName + ` RENAME TO ` + newName + `;`

	// Execute the SQL command
	_, err = db.Exec(renameTableSQL)
	if err != nil {
		return err
	}
	return nil
}
