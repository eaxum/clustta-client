package services

import (
	"clustta/internal/auth_service"
	"fmt"
)

type AccountService struct{}

// AccountInfo represents account data with auth context for frontend binding
type AccountInfo struct {
	UserId    string `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AuthMode  string `json:"auth_mode"`
	AuthHost  string `json:"auth_host"`
	StudioId  string `json:"studio_id"`
	SessionId string `json:"session_id"`
}

// GetAllAccounts returns all stored user accounts (basic Token for backward compatibility)
func (a *AccountService) GetAllAccounts() (map[string]auth_service.Token, error) {
	return auth_service.GetAllAccounts()
}

// GetAllAccountsWithContext returns all stored accounts with full auth context
func (a *AccountService) GetAllAccountsWithContext() (map[string]AccountInfo, error) {
	accountTokens, err := auth_service.GetAllAccountTokens()
	if err != nil {
		return nil, err
	}

	result := make(map[string]AccountInfo)
	for id, token := range accountTokens {
		result[id] = AccountInfo{
			UserId:    token.User.Id,
			Username:  token.User.Username,
			Email:     token.User.Email,
			FirstName: token.User.FirstName,
			LastName:  token.User.LastName,
			AuthMode:  string(token.AuthMode),
			AuthHost:  token.AuthHost,
			StudioId:  token.StudioId,
			SessionId: token.SessionId,
		}
	}
	return result, nil
}

// GetActiveAccount returns the currently active account (basic Token for backward compatibility)
func (a *AccountService) GetActiveAccount() (auth_service.Token, error) {
	return auth_service.GetActiveAccount()
}

// GetActiveAccountWithContext returns the currently active account with full auth context
func (a *AccountService) GetActiveAccountWithContext() (AccountInfo, error) {
	accountToken, err := auth_service.GetActiveAccountToken()
	if err != nil {
		return AccountInfo{}, err
	}

	return AccountInfo{
		UserId:    accountToken.User.Id,
		Username:  accountToken.User.Username,
		Email:     accountToken.User.Email,
		FirstName: accountToken.User.FirstName,
		LastName:  accountToken.User.LastName,
		AuthMode:  string(accountToken.AuthMode),
		AuthHost:  accountToken.AuthHost,
		StudioId:  accountToken.StudioId,
		SessionId: accountToken.SessionId,
	}, nil
}

// SwitchAccount changes the active account
func (a *AccountService) SwitchAccount(userId string) error {
	err := auth_service.SwitchToAccount(userId)
	if err != nil {
		return fmt.Errorf("failed to switch account: %w", err)
	}
	return nil
}

// RemoveAccount removes an account from storage
func (a *AccountService) RemoveAccount(userId string) error {
	err := auth_service.RemoveAccount(userId)
	if err != nil {
		return fmt.Errorf("failed to remove account: %w", err)
	}
	return nil
}

// AddAccount adds a new account (used after login)
func (a *AccountService) AddAccount(token auth_service.Token) error {
	err := auth_service.AddAccount(token)
	if err != nil {
		return fmt.Errorf("failed to add account: %w", err)
	}
	return nil
}

// GetAccountCount returns the number of stored accounts
func (a *AccountService) GetAccountCount() (int, error) {
	accounts, err := auth_service.GetAllAccounts()
	if err != nil {
		return 0, err
	}
	return len(accounts), nil
}
