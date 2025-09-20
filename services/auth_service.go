package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/error_service"
)

type AuthService struct{}

func (a *AuthService) Login(username, password string) (auth_service.Token, error) {
	token, err := auth_service.Login(username, password)
	if err != nil {
		return token, err
	}
	return token, nil
}

func (a *AuthService) Register(firstName, lastName, username, email, password, confirmPassword string) (auth_service.User, error) {
	user, err := auth_service.Register(firstName, lastName, username, email, password, confirmPassword)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (a *AuthService) UpdateUser(firstName, lastName, username, email string) (auth_service.User, error) {
	user, err := auth_service.UpdateUser(firstName, lastName, username, email)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (a *AuthService) UpdateUserPhoto(photo []byte) error {
	return auth_service.UpdateUserPhoto(photo)
}

func (a *AuthService) Logout(username, password string) error {
	err := auth_service.Logout()
	if err != nil {
		return err
	}
	return nil
}

func (a *AuthService) AuthUser() (auth_service.User, error) {
	saveToken, err := auth_service.GetToken()
	if err != nil {
		return auth_service.User{}, err
	}
	user := saveToken.User
	if user.Username == "" {
		return user, error_service.ErrUserNotFound
	}
	return user, nil
}

func (a *AuthService) IsAuthenticated() (bool, auth_service.User, error) {
	isAuthenticated, err := auth_service.IsAuthenticated()
	if err != nil {
		if err.Error() == "secret not found in keyring" {
			return false, auth_service.User{}, nil
		}
		return false, auth_service.User{}, err
	}
	if isAuthenticated {
		saveToken, err := auth_service.GetToken()
		if err != nil {
			if err.Error() == "secret not found in keyring" {
				return false, auth_service.User{}, nil
			}
			return false, auth_service.User{}, err
		}
		user := saveToken.User
		if user.Username == "" {
			return false, auth_service.User{}, nil
		}
		// data := map[string]interface{}{
		// 	"message": true,
		// 	"user":    user,
		// }
		return true, user, nil
	}
	return false, auth_service.User{}, nil
}

func (a *AuthService) CheckUsernameExists(username string) (bool, error) {
	return auth_service.CheckUsernameExists(username)
}

func (a *AuthService) CheckEmailExists(email string) (bool, error) {
	return auth_service.CheckEmailExists(email)
}

func (a *AuthService) DeactivateUserAccount() error {
	return auth_service.DeactivateUserAccount()
}

func (a *AuthService) SendInvitationEmail(email, studioName, projectName string) error {
	return auth_service.SendInvitationEmail(email, studioName, projectName)
}

func (a *AuthService) VerifyOTP(email, token string) error {
	return auth_service.VerifyOTP(email, token)
}

func (a *AuthService) ResendToken(email string) error {
	return auth_service.ResendToken(email)
}
