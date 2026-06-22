package services

import (
	"clustta/internal/server/models"
	"clustta/internal/studio_service"
	"strings"
)

type StudioService struct{}

// Fetches all users associated with a studio by studio ID
func (s *StudioService) GetStudioUsers(studioId string) ([]models.StudioUserInfo, error) {
	users, err := studio_service.GetStudioUsers(studioId)
	if err != nil {
		return users, err
	}
	return users, nil
}

// Adds a new collaborator to a studio with specified role
func (s *StudioService) AddCollaborator(email, studioId, roleName string) (interface{}, error) {
	// Convert role name to lowercase to match JS version
	roleName = strings.ToLower(roleName)

	result, err := studio_service.AddCollaborator(email, studioId, roleName)
	if err != nil {
		return result, err
	}
	return result, nil
}

// Changes the role of an existing collaborator in a studio
func (s *StudioService) ChangeCollaboratorRole(userId, studioId, roleName string) (interface{}, error) {
	// Convert role name to lowercase to match JS version
	roleName = strings.ToLower(roleName)

	result, err := studio_service.ChangeCollaboratorRole(userId, studioId, roleName)
	if err != nil {
		return result, err
	}
	return result, nil
}

// Removes a collaborator from a studio by user ID
func (s *StudioService) RemoveCollaborator(userId, studioId string) (interface{}, error) {
	result, err := studio_service.RemoveCollaborator(userId, studioId)
	if err != nil {
		return result, err
	}
	return result, nil
}

// Checks if a studio server is online or offline by URL
func (s *StudioService) GetStudioStatus(studioUrl string) (string, error) {
	status, err := studio_service.GetStudioStatus(studioUrl)
	if err != nil {
		return "offline", err
	}
	return status, nil
}

// Gets the version of a studio server
func (s *StudioService) GetServerVersion(studioUrl string) (string, error) {
	version, err := studio_service.GetServerVersion(studioUrl)
	if err != nil {
		return "", err
	}
	return version, nil
}

// Gets VM-local project and storage usage from a private studio server.
func (s *StudioService) GetStudioUsage(studioUrl string) (studio_service.StudioUsage, error) {
	usage, err := studio_service.GetStudioUsage(studioUrl)
	if err != nil {
		return studio_service.StudioUsage{}, err
	}
	return usage, nil
}

// Registers a new studio with name, URL, and hosting mode.
func (s *StudioService) RegisterStudio(name, studioUrl, hostingMode string) (interface{}, error) {
	result, err := studio_service.RegisterStudio(name, studioUrl, hostingMode)
	if err != nil {
		return result, err
	}
	return result, nil
}

// Updates studio configuration including URLs, port, and key
func (s *StudioService) UpdateStudio(studioName, url, altUrl, port, key string) (interface{}, error) {
	result, err := studio_service.UpdateStudio(studioName, url, altUrl, port, key)
	if err != nil {
		return result, err
	}
	return result, nil
}

// Updates a private (self-hosted) studio's local config via PUT /studio-info on the
// studio server itself. Pass "" for any field that should be left unchanged.
func (s *StudioService) UpdateStudioInfo(studioUrl, name, url, altUrl, port string) (studio_service.StudioInfo, error) {
	return studio_service.UpdateStudioInfo(studioUrl, name, url, altUrl, port)
}

// Verifies a deployment code for studio access
func (s *StudioService) VerifyDeploymentCode(code string) (bool, string, error) {
	valid, message, err := studio_service.VerifyDeploymentCode(code)
	if err != nil {
		return false, "", err
	}
	return valid, message, nil
}

// Races the primary and alternative studio URLs, returning whichever responds first.
// Falls back to the primary URL if no alternative is set.
func (s *StudioService) ResolveStudioUrl(url, altUrl string) (string, error) {
	resolvedUrl, err := studio_service.ResolveStudioUrl(url, altUrl)
	if err != nil {
		return resolvedUrl, err
	}
	return resolvedUrl, nil
}

// Checks if a studio name is already registered
func (s *StudioService) CheckStudioNameExists(studioName string) (bool, error) {
	exists, err := studio_service.CheckStudioNameExists(studioName)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// StudioInfo type alias for frontend bindings
type StudioInfo = studio_service.StudioInfo

// GetStudioInfo fetches studio metadata from a private studio server.
// Used when authenticated against a private server to discover its details.
func (s *StudioService) GetStudioInfo(studioUrl string) (studio_service.StudioInfo, error) {
	info, err := studio_service.GetStudioInfo(studioUrl)
	if err != nil {
		return studio_service.StudioInfo{}, err
	}
	return info, nil
}
