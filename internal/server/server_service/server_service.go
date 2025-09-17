package server_service

import (
	"clustta/internal/base_service"
	"clustta/internal/server/models"
	"clustta/internal/server/user_service"
	"clustta/internal/utils"
	"database/sql"
	_ "embed"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

//go:embed schema.sql
var Schema string

func InitData(tx *sqlx.Tx) error {
	roles := []string{"admin", "user", "guest"}
	for _, role := range roles {
		_, err := GetOrCreateRole(tx, role)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitDB(serverDB string, walMode bool) error {
	db, err := sqlx.Open("sqlite3", serverDB)
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

	tx := db.MustBegin()
	defer tx.Rollback()
	err = InitData(tx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error initializing data: %v", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}
	return nil
}

func InitWaitingListDB(waitingListDB string, walMode bool) error {
	db, err := sqlx.Open("sqlite3", waitingListDB)
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

	wlSchema := `
		CREATE TABLE IF NOT EXISTS waiting_list (
			email TEXT PRIMARY KEY UNIQUE COLLATE NOCASE,
			operating_system TEXT DEFAULT ""
		);
	`

	err = utils.CreateSchema(db, wlSchema)
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}
	return nil
}

func UpdateDB(serverDB string) error {
	db, err := sqlx.Open("sqlite3", serverDB)
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

func CreateStudio(tx *sqlx.Tx, name string, url string, user models.User) (models.Studio, error) {
	id := uuid.New().String()
	studioKey := utils.GenerateToken()
	studioKeyHash, err := utils.HashPassword(studioKey)
	if err != nil {
		return models.Studio{}, err
	}

	params := map[string]interface{}{
		"id":   id,
		"name": name,
		"url":  url,
		"key":  studioKeyHash,
	}

	err = base_service.Create(tx, "studio", params)
	if err != nil {
		return models.Studio{}, err
	}

	var studio models.Studio
	err = base_service.Get(tx, "studio", id, &studio)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Studio{}, fmt.Errorf("studio not found after creation")
		}
		return models.Studio{}, err
	}

	adminRole, err := GetRoleByName(tx, "admin")
	if err != nil {
		return models.Studio{}, err
	}

	err = AddStudioUser(tx, user.Id, studio.Id, adminRole.Id)
	if err != nil {
		return models.Studio{}, err
	}
	return studio, nil
}

func GenerateStudioKey(tx *sqlx.Tx, studioName string) (string, error) {
	studioKey := utils.GenerateToken()
	studioKeyHash, err := utils.HashPassword(studioKey)
	if err != nil {
		return "", err
	}

	studio, err := GetStudioByName(tx, studioName)
	if err != nil {
		return "", err
	}

	query := "UPDATE studio SET key = ? WHERE id = ?"
	_, err = tx.Exec(query, studioKeyHash, studio.Id)
	if err != nil {
		return "", err
	}

	return studioKey, nil
}

func AuthenticateServerKey(tx *sqlx.Tx, studioName, key string) (bool, error) {
	var storedPassword string
	studio, err := GetStudioByName(tx, studioName)
	if err != nil {
		return false, err
	}

	storedPassword = studio.Key
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(key))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func UpdateStudioURL(tx *sqlx.Tx, id, url, altUrl string) error {
	query := "UPDATE studio SET url = ?, alt_url = ? WHERE id = ?"
	_, err := tx.Exec(query, url, altUrl, id)
	if err != nil {
		return err
	}
	return nil
}

func GetStudio(tx *sqlx.Tx, id string) (models.Studio, error) {
	var studio models.Studio
	err := base_service.Get(tx, "studio", id, &studio)
	if err != nil {
		return studio, err
	}
	return studio, nil
}

func GetStudioByName(tx *sqlx.Tx, name string) (models.Studio, error) {
	var studio models.Studio
	query := "SELECT * FROM studio WHERE name = ?"
	err := tx.Get(&studio, query, name)
	if err != nil {
		return studio, err
	}
	return studio, nil
}

func GetStudioUsers(tx *sqlx.Tx, studioId string) ([]models.StudioUserInfo, error) {
	studioUserInfos := []models.StudioUserInfo{}
	query := `SELECT * FROM studio_user WHERE studio_id = ?`
	studioUsers := []models.StudioUser{}
	err := tx.Select(&studioUsers, query, studioId)
	if err != nil {
		return studioUserInfos, err
	}
	roles := map[string]models.Role{}
	allRole, err := GetRoles(tx)
	if err != nil {
		return studioUserInfos, err
	}
	for _, role := range allRole {
		roles[role.Id] = role
	}
	for _, studioUser := range studioUsers {
		user, err := user_service.GetUser(tx, studioUser.UserId)
		if err != nil {
			return studioUserInfos, err
		}
		role, ok := roles[studioUser.RoleId]
		if !ok {
			return studioUserInfos, fmt.Errorf("role not found")
		}
		studio, err := GetStudio(tx, studioUser.StudioId)
		if err != nil {
			return studioUserInfos, err
		}
		studioUserInfo := models.StudioUserInfo{
			Id:         user.Id,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			UserName:   user.UserName,
			Email:      user.Email,
			Active:     user.Active,
			RoleName:   role.Name,
			StudioName: studio.Name,
			StudioId:   studio.Id,
			RoleId:     role.Id,
		}
		studioUserInfos = append(studioUserInfos, studioUserInfo)
	}
	return studioUserInfos, nil
}

func AddStudioUser(tx *sqlx.Tx, userId, studioId, roleId string) error {
	id := uuid.New().String()
	params := map[string]interface{}{
		"id":        id,
		"user_id":   userId,
		"studio_id": studioId,
		"role_id":   roleId,
	}
	err := base_service.Create(tx, "studio_user", params)
	if err != nil {
		return err
	}
	return nil
}

func ChangeStudioUserRole(tx *sqlx.Tx, userId, studioId, roleId string) error {
	adminRole, err := GetRoleByName(tx, "admin")
	if err != nil {
		return err
	}
	var currentUserRoleId string
	err = tx.Get(&currentUserRoleId, "SELECT role_id FROM studio_user WHERE user_id = ? AND studio_id = ?", userId, studioId)
	if err != nil {
		return fmt.Errorf("failed to get current role: %w", err)
	}
	if currentUserRoleId == adminRole.Id && roleId != adminRole.Id {
		var adminCount int
		err = tx.Get(&adminCount, "SELECT COUNT(*) FROM studio_user WHERE studio_id = ? AND role_id =  ? ", studioId, adminRole.Id)
		if err != nil {
			return fmt.Errorf("failed to count admins: %w", err)
		}

		if adminCount <= 1 {
			return fmt.Errorf("cannot change role: studio must have at least one admin")
		}
	}
	query := "UPDATE studio_user SET role_id = ? WHERE user_id = ? AND studio_id = ?"
	_, err = tx.Exec(query, roleId, userId, studioId)
	if err != nil {
		return err
	}
	return nil
}

func RemoveStudioUser(tx *sqlx.Tx, userId, studioId string) error {
	adminRole, err := GetRoleByName(tx, "admin")
	if err != nil {
		return err
	}
	var currentRoleId string
	err = tx.Get(&currentRoleId, "SELECT role_id FROM studio_user WHERE user_id = ? AND studio_id = ?", userId, studioId)
	if err != nil {
		return fmt.Errorf("failed to get user role: %w", err)
	}

	if currentRoleId == adminRole.Id {
		var adminCount int
		err = tx.Get(&adminCount, "SELECT COUNT(*) FROM studio_user WHERE studio_id = ? AND role_id = ?", studioId, adminRole.Id)
		if err != nil {
			return fmt.Errorf("failed to count admins: %w", err)
		}

		if adminCount <= 1 {
			return fmt.Errorf("cannot remove user: studio must have at least one admin")
		}
	}
	query := "DELETE FROM studio_user WHERE user_id = ? AND studio_id = ?"
	_, err = tx.Exec(query, userId, studioId)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func GetUserStudios(tx *sqlx.Tx, userId string) ([]models.Studio, error) {
	query := `
		SELECT studio.*
		FROM studio
		LEFT JOIN
			studio_user ON studio.id = studio_user.studio_id
		WHERE studio_user.user_id = ?
	`
	userStudios := []models.Studio{}
	err := tx.Select(&userStudios, query, userId)
	if err != nil {
		return userStudios, err
	}
	return userStudios, nil
}

func GetAllStudios(tx *sqlx.Tx) ([]models.Studio, error) {
	query := ` SELECT * FROM studio `
	allStudios := []models.Studio{}
	err := tx.Select(&allStudios, query)
	if err != nil {
		return allStudios, err
	}
	return allStudios, nil
}

func CreateRole(tx *sqlx.Tx, name string) error {
	id := uuid.New().String()
	params := map[string]interface{}{
		"id":   id,
		"name": name,
	}
	err := base_service.Create(tx, "role", params)
	if err != nil {
		return err
	}
	return nil
}

func GetRole(tx *sqlx.Tx, id string) (models.Role, error) {
	var role models.Role
	err := base_service.Get(tx, "role", id, &role)
	if err != nil {
		return role, err
	}
	return role, nil
}

func GetRoles(tx *sqlx.Tx) ([]models.Role, error) {
	roles := []models.Role{}
	err := base_service.GetAll(tx, "role", &roles)
	if err != nil {
		return roles, err
	}
	return roles, nil
}

func GetRoleByName(tx *sqlx.Tx, name string) (models.Role, error) {
	var role models.Role
	query := "SELECT * FROM role WHERE name = ?"
	err := tx.Get(&role, query, name)
	if err != nil {
		return role, err
	}
	return role, nil
}

func GetOrCreateRole(tx *sqlx.Tx, name string) (models.Role, error) {
	var role models.Role
	query := "SELECT * FROM role WHERE name = ?"
	err := tx.Get(&role, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			err = CreateRole(tx, name)
			if err != nil {
				return role, err
			}
			err = tx.Get(&role, query, name)
			if err != nil {
				return role, err
			}
		} else {
			return role, err
		}
	}
	return role, nil
}

func IsUserInStudio(tx *sqlx.Tx, userId, studioId string) (bool, error) {
	query := "SELECT COUNT(*) FROM studio_user WHERE user_id = ? AND studio_id = ?"
	var count int
	err := tx.Get(&count, query, userId, studioId)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
