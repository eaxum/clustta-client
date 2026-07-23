package handlers

import (
	"errors"
	"net/http"
	"sync"

	"clustta/internal/auth_service"
	"clustta/internal/repository"
	"clustta/internal/repository/sync_service"
	"clustta/internal/settings"
)

var errProjectNotFound = errors.New("project not found")

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
	projects, err := listStudioProjects()
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

	projects, err := listStudioProjects()
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

func listStudioProjects() ([]repository.ProjectInfo, error) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return nil, err
	}

	studioName, err := settings.GetLastStudio()
	if err != nil {
		return nil, err
	}
	if studioName == "" {
		return nil, errors.New("no active studio selected")
	}

	studioURL := ""
	hostingMode := ""
	studioID := ""
	studios, err := settings.GetStudios()
	if err != nil {
		return nil, err
	}
	for _, studio := range studios {
		if studio.Name != studioName {
			continue
		}
		studioURL = studio.Url
		hostingMode = studio.HostingMode
		studioID = studio.Id
		break
	}

	return sync_service.GetStudioProjects(user, studioURL, studioName, hostingMode, studioID)
}

func resolveProject(projectID string) (repository.ProjectInfo, error) {
	projects, err := listStudioProjects()
	if err != nil {
		return repository.ProjectInfo{}, err
	}
	for _, project := range projects {
		if project.Id == projectID {
			return project, nil
		}
	}
	return repository.ProjectInfo{}, errProjectNotFound
}
