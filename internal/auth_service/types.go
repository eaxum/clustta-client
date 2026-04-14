package auth_service

import "clustta/internal/constants"

// AuthMode represents the type of authentication being used
type AuthMode string

const (
	// AuthModeGlobal indicates authentication against api.clustta.com
	AuthModeGlobal AuthMode = "global"

	// AuthModeStudio indicates authentication against a self-hosted studio server
	AuthModeStudio AuthMode = "studio"

	// AuthModeOffline indicates no authentication (local-only mode)
	AuthModeOffline AuthMode = "offline"
)

// DefaultAuthHost is the default Clustta Cloud authentication endpoint.
// Uses constants.HOST so the build-time -ldflags override is respected.
var DefaultAuthHost = constants.HOST

// OfflineUserID is a special identifier for the offline mode pseudo-account
const OfflineUserID = "offline-user"

// OfflineUser returns a pseudo-user for offline mode
func OfflineUser() User {
	return User{
		Id:        OfflineUserID,
		Username:  "offline",
		Email:     "offline@local",
		FirstName: "Offline",
		LastName:  "User",
		Photo:     nil,
	}
}

// IsValidAuthMode checks if the given auth mode is valid
func IsValidAuthMode(mode AuthMode) bool {
	switch mode {
	case AuthModeGlobal, AuthModeStudio, AuthModeOffline:
		return true
	default:
		return false
	}
}

// AccountToken extends Token with authentication context information
// This is the enhanced token structure for multi-auth-source support
type AccountToken struct {
	SessionId string   `json:"session_id"`
	User      User     `json:"user"`
	AuthMode  AuthMode `json:"auth_mode"`
	AuthHost  string   `json:"auth_host"`
	StudioId  string   `json:"studio_id,omitempty"`
}

// ToToken converts AccountToken to the basic Token type for backward compatibility
func (at AccountToken) ToToken() Token {
	return Token{
		SessionId: at.SessionId,
		User:      at.User,
	}
}

// FromToken creates an AccountToken from a basic Token with specified auth context
func FromToken(token Token, authMode AuthMode, authHost string, studioId string) AccountToken {
	return AccountToken{
		SessionId: token.SessionId,
		User:      token.User,
		AuthMode:  authMode,
		AuthHost:  authHost,
		StudioId:  studioId,
	}
}

// NewOfflineAccountToken creates an AccountToken for offline mode
func NewOfflineAccountToken() AccountToken {
	return AccountToken{
		SessionId: "",
		User:      OfflineUser(),
		AuthMode:  AuthModeOffline,
		AuthHost:  "",
		StudioId:  "",
	}
}

// IsOffline returns true if this account is in offline mode
func (at AccountToken) IsOffline() bool {
	return at.AuthMode == AuthModeOffline
}

// IsGlobal returns true if this account uses Clustta Cloud authentication
func (at AccountToken) IsGlobal() bool {
	return at.AuthMode == AuthModeGlobal
}

// IsStudioAuth returns true if this account uses private studio authentication
func (at AccountToken) IsStudioAuth() bool {
	return at.AuthMode == AuthModeStudio
}

// GetEffectiveHost returns the auth host to use for API calls
// Returns DefaultAuthHost for global mode, the stored AuthHost for studio mode,
// and empty string for offline mode
func (at AccountToken) GetEffectiveHost() string {
	switch at.AuthMode {
	case AuthModeGlobal:
		return DefaultAuthHost
	case AuthModeStudio:
		if at.AuthHost != "" {
			return at.AuthHost
		}
		return DefaultAuthHost
	case AuthModeOffline:
		return ""
	default:
		return DefaultAuthHost
	}
}
