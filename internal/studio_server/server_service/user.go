package server_service

import (
	"errors"
	"reflect"

	"clustta/internal/auth_service"
	"clustta/internal/base_service"
	"clustta/internal/error_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"clustta/output"

	"github.com/jmoiron/sqlx"
)

func GetRole(tx *sqlx.Tx, id string) (models.ServerRole, error) {
	role := models.ServerRole{}
	err := base_service.Get(tx, "role", id, &role)
	if err != nil {
		return models.ServerRole{}, err
	}
	return role, nil
}

func CreateRole(
	tx *sqlx.Tx,
	id string,
	name string,
	attributes models.ServerRoleAttributes,
) (models.ServerRole, error) {
	role := models.ServerRole{}

	params := map[string]interface{}{
		"id":   id,
		"name": name,
	}

	val := reflect.ValueOf(attributes)
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i).Interface()
		fieldName := utils.ToSnakeCase(field.Name)
		params[fieldName] = value
	}

	err := base_service.Create(tx, "role", params)
	if err != nil {
		return role, err
	}
	err = base_service.GetByName(tx, "role", name, &role)
	if err != nil {
		return models.ServerRole{}, err
	}
	return role, nil
}

func GetRoles(tx *sqlx.Tx) ([]models.ServerRole, error) {
	role := []models.ServerRole{}
	err := base_service.GetAll(tx, "role", &role)
	if err != nil {
		return role, err
	}
	return role, nil
}

func GetRoleByName(tx *sqlx.Tx, name string) (models.ServerRole, error) {
	role := models.ServerRole{}
	err := base_service.GetByName(tx, "role", name, &role)
	if err != nil {
		return role, err
	}
	return role, nil
}

func GetOrCreateRole(tx *sqlx.Tx, name string,
	attributes models.ServerRoleAttributes,
) (models.ServerRole, error) {
	role, err := GetRoleByName(tx, name)
	if err == nil {
		return role, nil
	}
	createdRole, err := CreateRole(tx, "", name, attributes)
	if err != nil {
		return models.ServerRole{}, err
	}
	return createdRole, nil
}

func GetUser(tx *sqlx.Tx, id string) (models.StudioUser, error) {
	user := models.StudioUser{}
	err := base_service.Get(tx, "user", id, &user)
	if err != nil && errors.Is(err, error_service.ErrUserNotFound) {
		return user, error_service.ErrUserNotFound
	} else if err != nil {
		return user, err
	}
	userRole, err := GetRole(tx, user.RoleId)
	if err != nil {
		return user, err
	}
	user.Role = userRole
	return user, nil
}

func AddUser(
	tx *sqlx.Tx,
	email string,
	roleName string,
) (models.StudioUser, error) {
	role, err := GetRoleByName(tx, roleName)
	if err != nil {
		return models.StudioUser{}, err
	}
	userData, err := auth_service.FetchUserData(email)

	if err != nil {
		if errors.Is(err, error_service.ErrNotAutheticated) {
			output.ErrorMessage(
				"User Not Authenticated",
				"User Not Authenticated",
				"user not authenticated",
			)
			return models.StudioUser{}, nil
		} else if errors.Is(err, error_service.ErrNotUnauthorized) {
			output.ErrorMessage(
				"User Unauthorized",
				"User Unauthorized",
				"user Unauthorized",
			)
			return models.StudioUser{}, nil
		}
		return models.StudioUser{}, err
	}
	user := models.StudioUser{}
	addedAt := utils.GetCurrentTime()
	params := map[string]interface{}{
		"id":         userData.Id,
		"added_at":   addedAt,
		"username":   userData.Username,
		"email":      email,
		"first_name": userData.FirstName,
		"last_name":  userData.LastName,
		"role_id":    role.Id,
	}
	err = base_service.Create(tx, "user", params)
	if err != nil {
		return user, err
	}
	err = base_service.Get(tx, "user", userData.Id, &user)
	if err != nil {
		return models.StudioUser{}, err
	}
	return user, nil
}

func AddKnownUser(
	tx *sqlx.Tx,
	id string,
	email string,
	username string,
	firstName string,
	lastName string,
	roleId string,
) (models.StudioUser, error) {
	user := models.StudioUser{}
	role, err := GetRole(tx, roleId)
	if err != nil {
		return user, err
	}
	addedAt := utils.GetCurrentTime()
	params := map[string]interface{}{
		"id":         id,
		"added_at":   addedAt,
		"username":   username,
		"email":      email,
		"first_name": firstName,
		"last_name":  lastName,
		"role_id":    role.Id,
	}
	err = base_service.Create(tx, "user", params)
	if err != nil {
		return user, err
	}
	err = base_service.Get(tx, "user", id, &user)
	if err != nil {
		return models.StudioUser{}, err
	}
	return user, nil
}

func GetUsers(tx *sqlx.Tx) ([]models.StudioUser, error) {
	users := []models.StudioUser{}
	err := base_service.GetAll(tx, "user", &users)
	if err != nil {
		return users, err
	}
	for i, user := range users {
		userRole, err := GetRole(tx, user.RoleId)
		if err != nil {
			return users, err
		}
		users[i].Role = userRole
	}
	return users, nil
}

func ChangeUserRoleByName(tx *sqlx.Tx, userId string, role_name string) error {
	role, err := GetRoleByName(tx, role_name)
	if err != nil {
		return err
	}
	role_id := role.Id
	err = ChangeUserRole(tx, userId, role_id)
	if err != nil {
		return err
	}
	return nil
}

func getRoleUsers(tx *sqlx.Tx, roleId string) ([]models.StudioUser, error) {
	users := []models.StudioUser{}
	conditions := map[string]interface{}{
		"role_id": roleId,
	}
	err := base_service.GetAllBy(tx, "user", conditions, &users)
	if err != nil {
		return users, err
	}
	return users, err
}

func ChangeUserRole(tx *sqlx.Tx, userId string, roleId string) error {
	params := map[string]interface{}{
		"role_id": roleId,
	}
	adminRole, err := GetRoleByName(tx, "admin")
	if err != nil {
		return err
	}
	adminUsers, err := getRoleUsers(tx, adminRole.Id)
	if err != nil {
		return err
	}

	adminIds := []string{}
	for _, adminUser := range adminUsers {
		adminIds = append(adminIds, adminUser.Id)
	}
	if utils.Contains(adminIds, userId) && roleId != adminRole.Id && len(adminUsers) <= 1 {
		return error_service.ErrMustHaveAdmin
	}
	err = base_service.Update(tx, "user", userId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "user", userId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func RemoveUser(tx *sqlx.Tx, userId string) error {
	tasks, err := repository.GetUserTasks(tx, userId)
	if err != nil {
		return err
	}
	if len(tasks) != 0 {
		return error_service.ErrUserHaveTaskAssigned
	}
	activeUser, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	if activeUser.Id == userId {
		return errors.New("you cannot remove youself")
	}
	err = base_service.Delete(tx, "user", userId)
	if err != nil {
		return err
	}
	return nil
}
