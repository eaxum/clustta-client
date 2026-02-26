package handlers

import (
	"net/http"

	"clustta/internal/auth_service"
)

// accountResponse is the JSON shape returned for each account.
type accountResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// ListAccounts returns all accounts stored in the OS keyring.
func ListAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := auth_service.GetAllAccounts()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]accountResponse, 0, len(accounts))
	for id, token := range accounts {
		result = append(result, accountResponse{
			ID:        id,
			Email:     token.User.Email,
			Username:  token.User.Username,
			FirstName: token.User.FirstName,
			LastName:  token.User.LastName,
		})
	}

	jsonResponse(w, http.StatusOK, result)
}

// GetActiveAccount returns the currently active account.
func GetActiveAccount(w http.ResponseWriter, r *http.Request) {
	token, err := auth_service.GetActiveAccount()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, accountResponse{
		ID:        token.User.Id,
		Email:     token.User.Email,
		Username:  token.User.Username,
		FirstName: token.User.FirstName,
		LastName:  token.User.LastName,
	})
}

// SwitchAccount changes the active account by user ID.
func SwitchAccount(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID string `json:"id"`
	}
	if err := decodeBody(r, &body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.ID == "" {
		jsonError(w, http.StatusBadRequest, "id is required")
		return
	}

	if err := auth_service.SwitchToAccount(body.ID); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}
