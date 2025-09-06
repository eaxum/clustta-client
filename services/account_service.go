package services

import (
	"clustta/internal/auth_service"
	"fmt"
)

type AccountService struct{}

// GetAllAccounts returns all stored user accounts
func (a *AccountService) GetAllAccounts() (map[string]auth_service.Token, error) {
	return auth_service.GetAllAccounts()
}

// GetActiveAccount returns the currently active account
func (a *AccountService) GetActiveAccount() (auth_service.Token, error) {
	return auth_service.GetActiveAccount()
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
