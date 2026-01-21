package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/error_service"
)

type AuthService struct{}

// AuthMode type alias for frontend binding
type AuthMode = auth_service.AuthMode

// Auth mode constants
const (
	AuthModeGlobal  = auth_service.AuthModeGlobal
	AuthModeStudio  = auth_service.AuthModeStudio
	AuthModeOffline = auth_service.AuthModeOffline
)

// Login authenticates a user with username and password against Clustta Cloud.
// Returns the authentication token or an error if login fails.
func (a *AuthService) Login(username, password string) (auth_service.Token, error) {
	token, err := auth_service.Login(username, password)
	if err != nil {
		return token, err
	}
	return token, nil
}

// LoginWithHost authenticates a user against a specified authentication host.
// authMode should be "global", "studio", or "offline".
// Returns the authentication token or an error if login fails.
func (a *AuthService) LoginWithHost(username, password, authHost string, authMode string, studioId string) (auth_service.Token, error) {
	mode := auth_service.AuthMode(authMode)
	if !auth_service.IsValidAuthMode(mode) {
		mode = auth_service.AuthModeGlobal
	}
	token, err := auth_service.LoginWithHost(username, password, authHost, mode, studioId)
	if err != nil {
		return token, err
	}
	return token, nil
}

// EnableOfflineMode sets up offline mode without authentication.
// Creates a local-only pseudo-account.
func (a *AuthService) EnableOfflineMode() error {
	return auth_service.EnableOfflineMode()
}

// IsOfflineMode checks if the current session is in offline mode.
func (a *AuthService) IsOfflineMode() bool {
	return auth_service.IsOfflineMode()
}

// GetAuthMode returns the current authentication mode.
// Returns "global", "studio", or "offline".
func (a *AuthService) GetAuthMode() string {
	return string(auth_service.GetActiveAuthMode())
}

// GetAuthHost returns the current authentication host URL.
// Returns empty string if in offline mode.
func (a *AuthService) GetAuthHost() string {
	return auth_service.GetAuthHost()
}

// Register creates a new user account on Clustta Cloud.
// Returns the created user or an error if registration fails.
func (a *AuthService) Register(firstName, lastName, username, email, password, confirmPassword string) (auth_service.User, error) {
	user, err := auth_service.Register(firstName, lastName, username, email, password, confirmPassword)
	if err != nil {
		return user, err
	}
	return user, nil
}

// RegisterWithHost creates a new user account on a specified authentication host.
// Returns the created user or an error if registration fails.
func (a *AuthService) RegisterWithHost(firstName, lastName, username, email, password, confirmPassword, authHost string) (auth_service.User, error) {
	user, err := auth_service.RegisterWithHost(firstName, lastName, username, email, password, confirmPassword, authHost)
	if err != nil {
		return user, err
	}
	return user, nil
}

// UpdateUser updates the current user's profile information.
// Returns the updated user or an error if the update fails.
func (a *AuthService) UpdateUser(firstName, lastName, username, email string) (auth_service.User, error) {
	user, err := auth_service.UpdateUser(firstName, lastName, username, email)
	if err != nil {
		return user, err
	}
	return user, nil
}

// UpdateUserPhoto updates the current user's profile photo.
// Returns an error if the upload fails.
func (a *AuthService) UpdateUserPhoto(photo []byte) error {
	return auth_service.UpdateUserPhoto(photo)
}

// Logout ends the current user session.
// Returns an error if logout fails.
func (a *AuthService) Logout(username, password string) error {
	err := auth_service.Logout()
	if err != nil {
		return err
	}
	return nil
}

// AuthUser retrieves the currently authenticated user.
// Returns an error if no user is found or the token is invalid.
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

// IsAuthenticated checks if a user is currently authenticated.
// Returns authentication status, user data, and any error encountered.
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
		return true, user, nil
	}
	return false, auth_service.User{}, nil
}

// CheckUsernameExists checks if a username is already registered.
// Returns true if the username exists, false otherwise.
func (a *AuthService) CheckUsernameExists(username string) (bool, error) {
	return auth_service.CheckUsernameExists(username)
}

// CheckEmailExists checks if an email address is already registered.
// Returns true if the email exists, false otherwise.
func (a *AuthService) CheckEmailExists(email string) (bool, error) {
	return auth_service.CheckEmailExists(email)
}

// DeactivateUserAccount deactivates the current user's account.
// Returns an error if the deactivation fails.
func (a *AuthService) DeactivateUserAccount() error {
	return auth_service.DeactivateUserAccount()
}

// SendInvitationEmail sends a project invitation to an email address.
// Returns an error if the email send fails.
func (a *AuthService) SendInvitationEmail(email, studioName, projectName string) error {
	return auth_service.SendInvitationEmail(email, studioName, projectName)
}

// VerifyOTP verifies a one-time password token.
// Returns an error if verification fails.
func (a *AuthService) VerifyOTP(email, token string) error {
	return auth_service.VerifyOTP(email, token)
}

// ResendToken resends the verification token to an email address.
// Returns an error if the send fails.
func (a *AuthService) ResendToken(email string) error {
	return auth_service.ResendToken(email)
}

// ChangePassword changes the current user's password.
// Returns an error if the password change fails.
func (a *AuthService) ChangePassword(currentPassword, newPassword, confirmPassword string) error {
	return auth_service.ChangePassword(currentPassword, newPassword, confirmPassword)
}

// ResetPassword sends a password reset email to the specified email address.
// Returns an error if the reset request fails.
func (a *AuthService) ResetPassword(email string) error {
	return auth_service.ResetPassword(email)
}
