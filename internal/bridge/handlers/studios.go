package handlers

import (
	"net/http"
	"sync"
	"time"

	"clustta/internal/auth_service"
	"clustta/internal/settings"
)

const studioCacheTTL = 5 * time.Minute

var (
	studioCacheMu        sync.RWMutex
	studioCacheRefreshMu sync.Mutex
	studioCacheKey       string
	studioCacheExpiresAt time.Time
	studioCacheItems     []settings.Studio
)

// studioResponse is the JSON shape returned for each studio.
type studioResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// ListStudios returns all studios for the active account.
func ListStudios(w http.ResponseWriter, r *http.Request) {
	studios, err := listStudiosCached(refreshRequested(r))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, studiosToResponse(studios))
}

func studiosToResponse(studios []settings.Studio) []studioResponse {
	result := make([]studioResponse, 0, len(studios))
	for _, s := range studios {
		result = append(result, studioResponse{
			ID:   s.Id,
			Name: s.Name,
			URL:  s.Url,
		})
	}
	return result
}

func listStudiosCached(force bool) ([]settings.Studio, error) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return nil, err
	}
	key := user.Id
	if !force {
		if studios, ok := readStudioCache(key); ok {
			return studios, nil
		}
	}

	studioCacheRefreshMu.Lock()
	defer studioCacheRefreshMu.Unlock()
	if !force {
		if studios, ok := readStudioCache(key); ok {
			return studios, nil
		}
	}

	studios, err := settings.GetStudios()
	if err != nil {
		return nil, err
	}
	studioCacheMu.Lock()
	studioCacheKey = key
	studioCacheItems = append([]settings.Studio(nil), studios...)
	studioCacheExpiresAt = time.Now().Add(studioCacheTTL)
	studioCacheMu.Unlock()
	return append([]settings.Studio(nil), studios...), nil
}

func readStudioCache(key string) ([]settings.Studio, bool) {
	studioCacheMu.RLock()
	defer studioCacheMu.RUnlock()
	if studioCacheKey != key || time.Now().After(studioCacheExpiresAt) {
		return nil, false
	}
	return append([]settings.Studio(nil), studioCacheItems...), true
}

func invalidateStudioCache() {
	studioCacheMu.Lock()
	studioCacheKey = ""
	studioCacheItems = nil
	studioCacheExpiresAt = time.Time{}
	studioCacheMu.Unlock()
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
