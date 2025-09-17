package server_service

import (
	"clustta/internal/auth_service"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	_ "embed"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

//go:embed schema.sql
var ServerSchema string

func InitServerDB(serverDB string, serverName string, user auth_service.User, walMode bool) error {
	db, err := sqlx.Open("sqlite3", serverDB)
	if err != nil {
		return err
	}

	if walMode {
		_, err = db.Exec("PRAGMA journal_mode = WAL;")
		if err != nil {
			return err
		}
	}

	err = utils.CreateSchema(db, ServerSchema)
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	studio_id := uuid.New().String()
	_, err = tx.Exec("INSERT INTO config (name, value, mtime) VALUES ('studio_id', ?, ?)", studio_id, utils.GetEpochTime())
	if err != nil {
		return err
	}

	err = SetServerName(tx, serverName)
	if err != nil {
		return err
	}

	err = initData(tx)
	if err != nil {
		return err
	}

	role, err := GetRoleByName(tx, "admin")
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = AddKnownUser(tx, user.Id, user.Email, user.Username, user.FirstName, user.LastName, role.Id)
	if err != nil {
		return err
	}
	err = SetServerVersion(tx, 1.0)
	if err != nil {
		return err
	}
	tx.Commit()

	return nil
}

func UpdateServerDB(serverDB string) error {
	db, err := sqlx.Open("sqlite3", serverDB)
	if err != nil {
		return err
	}

	err = utils.CreateSchema(db, ServerSchema)
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = initData(tx)
	if err != nil {
		return err
	}

	err = SetServerVersion(tx, 1.0)
	if err != nil {
		return err
	}
	tx.Commit()

	return nil
}

func initData(tx *sqlx.Tx) error {

	adminRoleAttributes := models.ServerRoleAttributes{
		ViewProject:   true,
		CreateProject: true,
		UpdateProject: true,
		DeleteProject: true,
	}
	productionManagerRoleAttributes := models.ServerRoleAttributes{
		ViewProject:   true,
		CreateProject: true,
		UpdateProject: true,
		DeleteProject: false,
	}
	artistRoleAttributes := models.ServerRoleAttributes{
		ViewProject:   false,
		CreateProject: false,
		UpdateProject: false,
		DeleteProject: false,
	}
	vendorRoleAttributes := models.ServerRoleAttributes{
		ViewProject:   false,
		CreateProject: false,
		UpdateProject: false,
		DeleteProject: false,
	}
	_, err := GetOrCreateRole(tx, "admin", adminRoleAttributes)
	if err != nil {
		return err
	}
	_, err = GetOrCreateRole(tx, "production manager", productionManagerRoleAttributes)
	if err != nil {
		return err
	}
	_, err = GetOrCreateRole(tx, "artist", artistRoleAttributes)
	if err != nil {
		return err
	}
	_, err = GetOrCreateRole(tx, "vendor", vendorRoleAttributes)
	if err != nil {
		return err
	}
	return nil
}
