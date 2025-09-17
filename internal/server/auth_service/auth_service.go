package auth_service

import (
	"clustta/internal/utils"
	_ "embed"
	"fmt"

	"github.com/jmoiron/sqlx"
)

//go:embed schema.sql
var Schema string

func InitDB(authTokenDB string, walMode bool) error {
	db, err := sqlx.Open("sqlite3", authTokenDB)
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
	return nil
}

func UpdateDB(authTokenDB string) error {
	db, err := sqlx.Open("sqlite3", authTokenDB)
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
