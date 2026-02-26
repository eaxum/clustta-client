package handlers

import (
	"net/http"

	"clustta/internal/settings"
)

// studioResponse is the JSON shape returned for each studio.
type studioResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// ListStudios returns all studios for the active account.
func ListStudios(w http.ResponseWriter, r *http.Request) {
	studios, err := settings.GetStudios()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]studioResponse, 0, len(studios))
	for _, s := range studios {
		result = append(result, studioResponse{
			ID:   s.Id,
			Name: s.Name,
			URL:  s.Url,
		})
	}

	jsonResponse(w, http.StatusOK, result)
}

// GetActiveStudio returns the last selected studio name.
func GetActiveStudio(w http.ResponseWriter, r *http.Request) {
	name, err := settings.GetLastStudio()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"name": name})
}

// SwitchStudio sets the active studio by name.
func SwitchStudio(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := decodeBody(r, &body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Name == "" {
		jsonError(w, http.StatusBadRequest, "name is required")
		return
	}

	if err := settings.SetLastStudio(body.Name); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}
