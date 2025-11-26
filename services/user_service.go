package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/error_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
)

type UserService struct{}

// Retrieves all users from the project database
func (u *UserService) GetUsers(projectPath string) ([]models.User, error) {
	if !utils.FileExists(projectPath) {
		return []models.User{}, error_service.ErrProjectNotFound
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.User{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.User{}, err
	}
	defer tx.Rollback()

	users, err := repository.GetUsers(tx)
	if err != nil {
		return []models.User{}, err
	}

	return users, nil
}

// Retrieves a single user by ID from the project database
func (u *UserService) GetUser(projectPath, userId string) (models.User, error) {
	if !utils.FileExists(projectPath) {
		return models.User{}, error_service.ErrProjectNotFound
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.User{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.User{}, err
	}
	defer tx.Rollback()

	user, err := repository.GetUser(tx, userId)
	if err != nil {
		return models.User{}, err
	}
	tx.Commit()

	return user, nil
}

// Retrieves all roles from the project database
func (u *UserService) GetRoles(projectPath string) ([]models.Role, error) {
	if !utils.FileExists(projectPath) {
		return []models.Role{}, error_service.ErrProjectNotFound
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Role{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Role{}, err
	}
	defer tx.Rollback()

	roles, err := repository.GetRoles(tx)
	if err != nil {
		return []models.Role{}, err
	}

	return roles, nil
}

// Creates a new role with name and permission attributes
func (u *UserService) AddRole(projectPath, name string, attributes models.RoleAttributes) (models.Role, error) {
	if !utils.FileExists(projectPath) {
		return models.Role{}, error_service.ErrProjectNotFound
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Role{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Role{}, err
	}
	defer tx.Rollback()
	role, err := repository.CreateRole(tx, "", name, attributes)
	if err != nil {
		return models.Role{}, err
	}
	err = tx.Commit()
	if err != nil {
		return models.Role{}, err
	}

	return role, err
}

// Updates an existing role's name and permission attributes
func (u *UserService) UpdateRole(projectPath, id, name string, attributes models.RoleAttributes) (models.Role, error) {
	if !utils.FileExists(projectPath) {
		return models.Role{}, error_service.ErrProjectNotFound
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Role{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Role{}, err
	}
	defer tx.Rollback()
	role, err := repository.UpdateRole(tx, id, name, attributes)
	if err != nil {
		return models.Role{}, err
	}
	err = tx.Commit()
	if err != nil {
		return models.Role{}, err
	}

	return role, err
}

// Deletes a role by ID from the project database
func (u *UserService) DeleteRole(projectPath, id string) error {
	if !utils.FileExists(projectPath) {
		return error_service.ErrProjectNotFound
	}

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

	err = repository.DeleteRole(tx, id)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return err
}

// Fetches user data from remote authentication service by ID
func (u *UserService) FetchUserById(userId string) (models.User, error) {
	user, err := auth_service.FetchUserDataById(userId)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}
