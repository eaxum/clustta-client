package auth_service

import (
	"encoding/json"
	"fmt"

	"github.com/zalando/go-keyring"
)

// MultiAccountToken represents a collection of user accounts
type MultiAccountToken struct {
	ActiveAccountId string           `json:"active_account_id"`
	Accounts        map[string]Token `json:"accounts"` // key: user.id, value: Token
}

// GetMultiAccountToken retrieves the multi-account token structure from keyring
func GetMultiAccountToken() (MultiAccountToken, error) {
	service := "clustta"
	key := "clustta-accounts"

	tokenData, err := keyring.Get(service, key)
	if err != nil {
		// If multi-account structure doesn't exist, return empty structure
		return MultiAccountToken{
			ActiveAccountId: "",
			Accounts:        make(map[string]Token),
		}, nil
	}

	var multiToken MultiAccountToken
	err = json.Unmarshal([]byte(tokenData), &multiToken)
	if err != nil {
		return MultiAccountToken{}, err
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
func AddAccount(token Token) error {
	multiToken, err := GetMultiAccountToken()
	if err != nil {
		// If no multi-account structure exists, create a new one
		multiToken = MultiAccountToken{
			ActiveAccountId: token.User.Id,
			Accounts:        make(map[string]Token),
		}
	}

	// Add the new account
	multiToken.Accounts[token.User.Id] = token

	// If this is the first account or no active account is set, make it active
	if multiToken.ActiveAccountId == "" || len(multiToken.Accounts) == 1 {
		multiToken.ActiveAccountId = token.User.Id
	}

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

// GetActiveAccount returns the currently active account token
func GetActiveAccount() (Token, error) {
	multiToken, err := GetMultiAccountToken()
	if err != nil {
		return Token{}, err
	}

	if multiToken.ActiveAccountId == "" {
		return Token{}, fmt.Errorf("no active account set")
	}

	token, exists := multiToken.Accounts[multiToken.ActiveAccountId]
	if !exists {
		return Token{}, fmt.Errorf("active account not found in accounts list")
	}

	return token, nil
}

// GetAllAccounts returns all stored accounts
func GetAllAccounts() (map[string]Token, error) {
	multiToken, err := GetMultiAccountToken()
	if err != nil {
		return make(map[string]Token), err
	}

	return multiToken.Accounts, nil
}

// migrateFromSingleToken migrates from the old single token structure to multi-account
func migrateFromSingleToken() (MultiAccountToken, error) {
	// Try to get the old single token
	oldToken, err := getOldSingleToken()
	if err != nil {
		// No old token exists, return empty multi-account structure
		return MultiAccountToken{
			ActiveAccountId: "",
			Accounts:        make(map[string]Token),
		}, nil
	}

	// Create new multi-account structure with the old token
	multiToken := MultiAccountToken{
		ActiveAccountId: oldToken.User.Id,
		Accounts: map[string]Token{
			oldToken.User.Id: oldToken,
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
