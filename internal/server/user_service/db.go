package user_service

import (
	"clustta/internal/utils"
	_ "embed"
	"fmt"

	"github.com/jmoiron/sqlx"
)

//go:embed schema.sql
var Schema string

func InitDB(userDB string, walMode bool) error {
	db, err := sqlx.Open("sqlite3", userDB)
	if err != nil {
		return err
	}
	defer db.Close()

	if walMode {
		_, err = db.Exec("PRAGMA journal_mode = WAL;")
		if err != nil {
			return err
		}
	}

	err = utils.CreateSchema(db, Schema)
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}

	err = utils.AddColumnIfNotExist(db, "user", "active", "BOOLEAN", "0", false)
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}
	return nil
}

func UpdateDB(userDB string) error {
	db, err := sqlx.Open("sqlite3", userDB)
	if err != nil {
		return err
	}
	defer db.Close()

	err = utils.CreateSchema(db, Schema)
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}
	return nil
}
