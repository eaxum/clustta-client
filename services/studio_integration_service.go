package services

import (
	"clustta/internal/studio_integration_service"
)

// StudioIntegrationService exposes studio integration configuration to
// the Vue frontend via Wails bindings. Pure delegation to internal/.
type StudioIntegrationService struct{}

// GetStudioIntegration fetches the current configuration for the studio.
// Returns an unconfigured Config when none exists.
func (s *StudioIntegrationService) GetStudioIntegration(studioId, integrationId string) (studio_integration_service.Config, error) {
	return studio_integration_service.GetConfig(studioId, integrationId)
}

// SaveStudioIntegration stores or updates the studio's credentials.
// The server validates against the remote service before persisting.
func (s *StudioIntegrationService) SaveStudioIntegration(studioId, integrationId string, payload studio_integration_service.CredentialsPayload) (studio_integration_service.Config, error) {
	return studio_integration_service.SaveConfig(studioId, integrationId, payload)
}

// DeleteStudioIntegration stops the listener and removes stored credentials.
func (s *StudioIntegrationService) DeleteStudioIntegration(studioId, integrationId string) error {
	return studio_integration_service.DeleteConfig(studioId, integrationId)
}

// TestStudioIntegration runs a live credential check without persisting.
// Empty payload falls back to stored credentials on the server.
func (s *StudioIntegrationService) TestStudioIntegration(studioId, integrationId string, payload studio_integration_service.CredentialsPayload) error {
	return studio_integration_service.TestConfig(studioId, integrationId, payload)
}

// SetStudioIntegrationEnabled toggles the enabled flag on an already-configured
// studio integration without requiring credentials. The server starts or stops
// the listener accordingly and returns the refreshed config view.
func (s *StudioIntegrationService) SetStudioIntegrationEnabled(studioId, integrationId string, enabled bool) (studio_integration_service.Config, error) {
	return studio_integration_service.SetEnabled(studioId, integrationId, enabled)
}
