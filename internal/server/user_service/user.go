package user_service

import (
	"clustta/internal/base_service"
	"clustta/internal/error_service"
	"clustta/internal/server/models"
	"clustta/internal/utils"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(tx *sqlx.Tx, id string, firstName string, lastName string, username string, email string,
	password string) (models.User, error) {
	if id == "" {
		id = uuid.New().String()
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return models.User{}, err
	}
	lastPresence := utils.GetCurrentTime()
	params := map[string]interface{}{
		"id":            id,
		"first_name":    firstName,
		"last_name":     lastName,
		"username":      username,
		"email":         email,
		"password":      hashedPassword,
		"last_presence": lastPresence,
	}

	err = base_service.Create(tx, "user", params)
	if err != nil {
		return models.User{}, err
	}

	var createdUser models.User
	err = base_service.Get(tx, "user", id, &createdUser)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, fmt.Errorf("user not found after creation")
		}
		return models.User{}, err
	}
	return createdUser, nil
}

func GetUser(tx *sqlx.Tx, id string) (models.User, error) {
	var user models.User
	query := `SELECT id,first_name,last_name ,username, email, active FROM user WHERE id = ?`
	err := tx.Get(&user, query, id)
	if err != nil {
		return user, err
	}
	return user, nil
}

func GetUsers(tx *sqlx.Tx) ([]models.User, error) {
	var users []models.User
	err := base_service.GetAll(tx, "user", &users)
	if err != nil {
		return users, err
	}
	return users, nil
}
func GetUsersByUsername(tx *sqlx.Tx, usernames []string) ([]models.User, error) {
	var users []models.User

	// Convert []string to []interface{} for the IN clause
	usernamesInterface := make([]interface{}, len(usernames))
	for i, v := range usernames {
		usernamesInterface[i] = v
	}

	// Pass the condition with the IN clause
	conditions := map[string]interface{}{
		"username": usernamesInterface,
	}

	err := base_service.GetAllBy(tx, "user", conditions, &users)
	if err != nil {
		return users, err
	}
	// if len(users) != len(usernames) {
	// 	return nil, fmt.Errorf("failed to retrieve users , some username(s) were not found or username of the same name in list")
	// }
	return users, nil
}

func GetUserByUsername(tx *sqlx.Tx, username string) (models.User, error) {
	var user models.User
	condition := map[string]interface{}{
		"username": username,
	}
	err := base_service.GetBy(tx, "user", condition, &user)
	if err != nil {
		print(err.Error())
		return user, err
	}
	return user, nil
}
func GetUserByEmail(tx *sqlx.Tx, email string) (models.User, error) {
	var user models.User
	condition := map[string]interface{}{
		"email": email,
	}
	err := base_service.GetBy(tx, "user", condition, &user)
	if err != nil {
		// print(err.Error())
		return user, err
	}
	return user, nil
}
func IsUserDeleted(tx *sqlx.Tx, emailOrUsername string) (bool, error) {
	var isDeleted bool
	query := `SELECT is_deleted FROM user WHERE email = ? OR username = ?`
	err := tx.Get(&isDeleted, query, emailOrUsername, emailOrUsername)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, error_service.ErrUserNotFound
		}
		return false, err
	}
	return isDeleted, nil
}

func GetUserByUsernameOrEmail(tx *sqlx.Tx, emailOrUsername string) (models.User, error) {
	var user models.User
	query := `
		SELECT id, first_name, last_name, username, email, active, password 
		FROM user 
		WHERE email = ? OR username = ?
	`
	err := tx.Get(&user, query, emailOrUsername, emailOrUsername)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, error_service.ErrUserNotFound
		}
		return user, err
	}
	return user, nil
}
func DeleteUser(tx *sqlx.Tx, id string) error {
	var user models.User

	err := base_service.Get(tx, "user", id, &user)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user with id %s not found", id)
		}
		return fmt.Errorf("error fetching user: %v", err)
	}

	err = base_service.Delete(tx, "user", id)
	if err != nil {
		return fmt.Errorf("error deleting user: %v", err)
	}

	fmt.Println("User deleted successfully")
	return nil
}

func AuthenticateUser(tx *sqlx.Tx, emailOrUsername, password string) (models.User, bool, error) {
	var storedPassword string
	var user models.User
	user, err := GetUserByUsernameOrEmail(tx, emailOrUsername)
	if err != nil {
		if err == error_service.ErrUserNotFound {
			return user, false, nil
		}
		return user, false, err
	}

	storedPassword = user.Password
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return user, false, nil
		}
		return user, false, err
	}
	return user, true, nil
}

func UpdateUser(tx *sqlx.Tx, id string, params map[string]interface{}) error {
	err := base_service.Update(tx, "user", id, params)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func UpdatePassword(tx *sqlx.Tx, id string, new_password string) error {
	hashedPassword, err := utils.HashPassword(new_password)
	if err != nil {
		return err
	}
	params := map[string]interface{}{
		"password": hashedPassword,
	}
	err = base_service.Update(tx, "user", id, params)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func ValidateUser(user models.User) (bool, string) {
	if strings.TrimSpace(user.FirstName) == "" || len(user.FirstName) < 3 {
		return false, "First Name must be at least 3 characters long and cannot be empty"
	}
	if strings.TrimSpace(user.LastName) == "" || len(user.LastName) < 3 {
		return false, "Last Name must be at least 3 characters long and cannot be empty"
	}

	// Check if email is valid
	if !utils.ValidateEmail(user.Email) {
		return false, "Invalid email format"
	}
	// if isValid, errMsg := ValidatePassword(user.Password); !isValid {
	// 	return false, errMsg
	// }

	return true, ""
}
