package services

import (
	"clustta/internal/server/models"
	"clustta/internal/studio_service"
	"strings"
)

type StudioService struct{}

func (s *StudioService) GetStudioUsers(studioId string) ([]models.StudioUserInfo, error) {
	users, err := studio_service.GetStudioUsers(studioId)
	if err != nil {
		return users, err
	}
	return users, nil
}

func (s *StudioService) AddCollaborator(email, studioId, roleName string) (interface{}, error) {
	// Convert role name to lowercase to match JS version
	roleName = strings.ToLower(roleName)

	result, err := studio_service.AddCollaborator(email, studioId, roleName)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (s *StudioService) ChangeCollaboratorRole(userId, studioId, roleName string) (interface{}, error) {
	// Convert role name to lowercase to match JS version
	roleName = strings.ToLower(roleName)

	result, err := studio_service.ChangeCollaboratorRole(userId, studioId, roleName)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (s *StudioService) RemoveCollaborator(userId, studioId string) (interface{}, error) {
	result, err := studio_service.RemoveCollaborator(userId, studioId)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (s *StudioService) GetStudioStatus(studioUrl string) (string, error) {
	status, err := studio_service.GetStudioStatus(studioUrl)
	if err != nil {
		return "offline", err
	}
	return status, nil
}

func (s *StudioService) CreateStudio(name string) (interface{}, error) {
	result, err := studio_service.CreateStudio(name)
	if err != nil {
		return result, err
	}
	return result, nil
}
