package handlers

import (
	"net/http"
	"sync"

	"clustta/internal/auth_service"
	"clustta/internal/repository"
	"clustta/internal/repository/sync_service"
	"clustta/internal/settings"
)

// projectResponse is the JSON shape returned for each project.
type projectResponse struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Uri              string `json:"uri"`
	WorkingDirectory string `json:"working_directory"`
	IsDownloaded     bool   `json:"is_downloaded"`
	IsTracked        bool   `json:"is_tracked"`
}

// activeProjectState holds the currently selected project in memory.
var (
	activeProjectMu sync.RWMutex
	activeProject   *repository.ProjectInfo
)

// ListProjects returns projects for the active studio.
func ListProjects(w http.ResponseWriter, r *http.Request) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "no active account: "+err.Error())
		return
	}

	studioName, err := settings.GetLastStudio()
	if err != nil || studioName == "" {
		jsonError(w, http.StatusBadRequest, "no active studio selected")
		return
	}

	// Find the studio URL and hosting info
	studioURL := ""
	hostingMode := ""
	studioId := ""
	studios, err := settings.GetStudios()
	if err == nil {
		for _, s := range studios {
			if s.Name == studioName {
				studioURL = s.Url
				hostingMode = s.HostingMode
				studioId = s.Id
				break
			}
		}
	}

	projects, err := sync_service.GetStudioProjects(user, studioURL, studioName, hostingMode, studioId)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]projectResponse, 0, len(projects))
	for _, p := range projects {
		result = append(result, projectResponse{
			ID:               p.Id,
			Name:             p.Name,
			Uri:              p.Uri,
			WorkingDirectory: p.WorkingDirectory,
			IsDownloaded:     p.IsDownloaded,
			IsTracked:        p.IsTracked,
		})
	}

	jsonResponse(w, http.StatusOK, result)
}

// GetActiveProject returns the currently selected project.
func GetActiveProject(w http.ResponseWriter, r *http.Request) {
	activeProjectMu.RLock()
	defer activeProjectMu.RUnlock()

	if activeProject == nil {
		jsonResponse(w, http.StatusOK, nil)
		return
	}

	jsonResponse(w, http.StatusOK, projectResponse{
		ID:               activeProject.Id,
		Name:             activeProject.Name,
		Uri:              activeProject.Uri,
		WorkingDirectory: activeProject.WorkingDirectory,
		IsDownloaded:     activeProject.IsDownloaded,
		IsTracked:        activeProject.IsTracked,
	})
}

// SwitchProject sets the active project by URI.
func SwitchProject(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Uri string `json:"uri"`
	}
	if err := decodeBody(r, &body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Uri == "" {
		jsonError(w, http.StatusBadRequest, "uri is required")
		return
	}

	// Resolve the project from the current studio list
	user, err := auth_service.GetActiveUser()
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "no active account: "+err.Error())
		return
	}

	studioName, err := settings.GetLastStudio()
	if err != nil || studioName == "" {
		jsonError(w, http.StatusBadRequest, "no active studio selected")
		return
	}

	studioURL := ""
	hostingMode := ""
	studioId := ""
	studios, err := settings.GetStudios()
	if err == nil {
		for _, s := range studios {
			if s.Name == studioName {
				studioURL = s.Url
				hostingMode = s.HostingMode
				studioId = s.Id
				break
			}
		}
	}

	projects, err := sync_service.GetStudioProjects(user, studioURL, studioName, hostingMode, studioId)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var found *repository.ProjectInfo
	for i := range projects {
		if projects[i].Uri == body.Uri {
			found = &projects[i]
			break
		}
	}
	if found == nil {
		jsonError(w, http.StatusNotFound, "project not found")
		return
	}

	activeProjectMu.Lock()
	activeProject = found
	activeProjectMu.Unlock()

	jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}
