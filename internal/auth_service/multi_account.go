package auth_service

import (
	"encoding/json"
	"fmt"

	"github.com/zalando/go-keyring"
)

// MultiAccountToken represents a collection of user accounts with enhanced auth context
type MultiAccountToken struct {
	ActiveAccountId string                  `json:"active_account_id"`
	Accounts        map[string]AccountToken `json:"accounts"` // key: user.id, value: AccountToken
}

// LegacyMultiAccountToken is the old structure for migration purposes
type LegacyMultiAccountToken struct {
	ActiveAccountId string           `json:"active_account_id"`
	Accounts        map[string]Token `json:"accounts"`
}

// GetMultiAccountToken retrieves the multi-account token structure from keyring
// It handles migration from legacy Token format to AccountToken format
func GetMultiAccountToken() (MultiAccountToken, error) {
	service := "clustta"
	key := "clustta-accounts"

	tokenData, err := keyring.Get(service, key)
	if err != nil {
		// If multi-account structure doesn't exist, return empty structure
		return MultiAccountToken{
			ActiveAccountId: "",
			Accounts:        make(map[string]AccountToken),
		}, nil
	}

	// First try to unmarshal as new AccountToken format
	var multiToken MultiAccountToken
	err = json.Unmarshal([]byte(tokenData), &multiToken)
	if err != nil {
		return MultiAccountToken{}, err
	}

	// Check if migration is needed (accounts exist but have no AuthMode set)
	needsMigration := false
	for _, account := range multiToken.Accounts {
		if account.AuthMode == "" {
			needsMigration = true
			break
		}
	}

	if needsMigration {
		// Migrate legacy accounts to new format with default global auth mode
		migratedAccounts := make(map[string]AccountToken)
		for id, account := range multiToken.Accounts {
			if account.AuthMode == "" {
				account.AuthMode = AuthModeGlobal
				account.AuthHost = DefaultAuthHost
			}
			migratedAccounts[id] = account
		}
		multiToken.Accounts = migratedAccounts
		// Save migrated data
		SetMultiAccountToken(multiToken)
	}

	return multiToken, nil
}

// SetMultiAccountToken stores the multi-account token structure in keyring
func SetMultiAccountToken(multiToken MultiAccountToken) error {
	service := "clustta"
	key := "clustta-accounts"

	jsonToken, err := json.Marshal(multiToken)
	if err != nil {
		return err
	}

	err = keyring.Set(service, key, string(jsonToken))
	if err != nil {
		return err
	}

	return nil
}

// AddAccount adds a new account to the multi-account structure
// This is a convenience function that creates an AccountToken with global auth mode
func AddAccount(token Token) error {
	accountToken := FromToken(token, AuthModeGlobal, DefaultAuthHost, "")
	return AddAccountToken(accountToken)
}

// AddAccountToken adds a new account with full auth context to the multi-account structure.
// The newly added account is always set as the active account.
func AddAccountToken(accountToken AccountToken) error {
	multiToken, err := GetMultiAccountToken()
	if err != nil {
		// If no multi-account structure exists, create a new one
		multiToken = MultiAccountToken{
			ActiveAccountId: accountToken.User.Id,
			Accounts:        make(map[string]AccountToken),
		}
	}

	// Add the new account
	multiToken.Accounts[accountToken.User.Id] = accountToken

	// Always set the newly logged-in account as active
	multiToken.ActiveAccountId = accountToken.User.Id

	return SetMultiAccountToken(multiToken)
}

// SwitchToAccount changes the active account
func SwitchToAccount(userId string) error {
	multiToken, err := GetMultiAccountToken()
	if err != nil {
		return err
	}

	// Check if the account exists
	if _, exists := multiToken.Accounts[userId]; !exists {
		return fmt.Errorf("account with id %s not found", userId)
	}

	multiToken.ActiveAccountId = userId
	return SetMultiAccountToken(multiToken)
}

// RemoveAccount removes an account from the multi-account structure
func RemoveAccount(userId string) error {
	multiToken, err := GetMultiAccountToken()
	if err != nil {
		return err
	}

	// Remove the account
	delete(multiToken.Accounts, userId)

	// If we removed the active account, set a new active account
	if multiToken.ActiveAccountId == userId {
		if len(multiToken.Accounts) > 0 {
			// Set the first available account as active
			for id := range multiToken.Accounts {
				multiToken.ActiveAccountId = id
				break
			}
		} else {
			// No accounts left
			multiToken.ActiveAccountId = ""
		}
	}

	return SetMultiAccountToken(multiToken)
}

// GetActiveAccount returns the currently active account token (basic Token for backward compatibility)
func GetActiveAccount() (Token, error) {
	accountToken, err := GetActiveAccountToken()
	if err != nil {
		return Token{}, err
	}
	return accountToken.ToToken(), nil
}

// GetActiveAccountToken returns the currently active account with full auth context
func GetActiveAccountToken() (AccountToken, error) {
	multiToken, err := GetMultiAccountToken()
	if err != nil {
		return AccountToken{}, err
	}

	if multiToken.ActiveAccountId == "" {
		return AccountToken{}, fmt.Errorf("no active account set")
	}

	accountToken, exists := multiToken.Accounts[multiToken.ActiveAccountId]
	if !exists {
		return AccountToken{}, fmt.Errorf("active account not found in accounts list")
	}

	return accountToken, nil
}

// GetAllAccounts returns all stored accounts (basic Token map for backward compatibility)
func GetAllAccounts() (map[string]Token, error) {
	accountTokens, err := GetAllAccountTokens()
	if err != nil {
		return make(map[string]Token), err
	}

	tokens := make(map[string]Token)
	for id, accountToken := range accountTokens {
		tokens[id] = accountToken.ToToken()
	}
	return tokens, nil
}

// GetAllAccountTokens returns all stored accounts with full auth context
func GetAllAccountTokens() (map[string]AccountToken, error) {
	multiToken, err := GetMultiAccountToken()
	if err != nil {
		return make(map[string]AccountToken), err
	}

	return multiToken.Accounts, nil
}

// EnableOfflineMode sets up an offline mode account
func EnableOfflineMode() error {
	offlineAccount := NewOfflineAccountToken()
	return AddAccountToken(offlineAccount)
}

// IsOfflineMode checks if the current active account is in offline mode
func IsOfflineMode() bool {
	accountToken, err := GetActiveAccountToken()
	if err != nil {
		return false
	}
	return accountToken.IsOffline()
}

// GetAuthHost returns the authentication host for the active account
// Returns empty string if in offline mode
func GetAuthHost() string {
	accountToken, err := GetActiveAccountToken()
	if err != nil {
		return DefaultAuthHost
	}
	return accountToken.GetEffectiveHost()
}

// GetActiveAuthMode returns the auth mode of the active account
func GetActiveAuthMode() AuthMode {
	accountToken, err := GetActiveAccountToken()
	if err != nil {
		return AuthModeGlobal
	}
	return accountToken.AuthMode
}

// migrateFromSingleToken migrates from the old single token structure to multi-account
func migrateFromSingleToken() (MultiAccountToken, error) {
	// Try to get the old single token
	oldToken, err := getOldSingleToken()
	if err != nil {
		// No old token exists, return empty multi-account structure
		return MultiAccountToken{
			ActiveAccountId: "",
			Accounts:        make(map[string]AccountToken),
		}, nil
	}

	// Create new multi-account structure with the old token (default to global auth)
	accountToken := FromToken(oldToken, AuthModeGlobal, DefaultAuthHost, "")
	multiToken := MultiAccountToken{
		ActiveAccountId: oldToken.User.Id,
		Accounts: map[string]AccountToken{
			oldToken.User.Id: accountToken,
		},
	}

	// Store the new multi-account structure
	err = SetMultiAccountToken(multiToken)
	if err != nil {
		return MultiAccountToken{}, err
	}

	// Delete the old single token after successful migration
	deleteOldSingleToken()

	return multiToken, nil
}

// getOldSingleToken retrieves the old single token format
func getOldSingleToken() (Token, error) {
	service := "clustta"
	key := "token"

	tokenData, err := keyring.Get(service, key)
	if err != nil {
		return Token{}, err
	}

	var token Token
	err = json.Unmarshal([]byte(tokenData), &token)
	if err != nil {
		return Token{}, err
	}

	return token, nil
}

// deleteOldSingleToken removes the old single token (use with caution)
func deleteOldSingleToken() error {
	service := "clustta"
	key := "token"
	return keyring.Delete(service, key)
}
