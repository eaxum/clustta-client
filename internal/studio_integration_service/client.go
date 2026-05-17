// Package studio_integration_service handles client-side HTTP calls to the
// central server for studio-scoped integration configuration (Kitsu, etc.).
// Follows the same inline request pattern used by other services in
// internal/studio_service.
package studio_integration_service

import (
	"bytes"
	"clustta/internal/auth_service"
	"clustta/internal/constants"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Config mirrors the public view returned by the central server.
// Credentials are never included in this struct.
type Config struct {
	IntegrationId   string `json:"integration_id"`
	StudioId        string `json:"studio_id"`
	ApiUrl          string `json:"api_url"`
	Email           string `json:"email"`
	Enabled         bool   `json:"enabled"`
	LastValidatedAt int64  `json:"last_validated_at"`
	LastError       string `json:"last_error"`
	Status          string `json:"status"`
	Configured      bool   `json:"configured"`
}

// CredentialsPayload is the shape accepted by save/test endpoints.
type CredentialsPayload struct {
	ApiUrl   string `json:"api_url"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Enabled  bool   `json:"enabled"`
}

// errorEnvelope unwraps the standard {"error":"..."} response body.
type errorEnvelope struct {
	Error string `json:"error"`
}

// GetConfig fetches the current studio configuration for an integration.
// Returns an unconfigured Config when none exists; only errors on transport
// or server failures.
func GetConfig(studioId, integrationId string) (Config, error) {
	var cfg Config
	url := fmt.Sprintf("%s/studio-integration/%s/%s", constants.HOST, studioId, integrationId)
	body, err := doRequest(http.MethodGet, url, nil)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(body, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode response: %w", err)
	}
	return cfg, nil
}

// SaveConfig persists the supplied credentials after the server validates
// them against the remote service and (re)starts the listener.
func SaveConfig(studioId, integrationId string, payload CredentialsPayload) (Config, error) {
	var cfg Config
	url := fmt.Sprintf("%s/studio-integration/%s/%s", constants.HOST, studioId, integrationId)
	body, err := doRequest(http.MethodPut, url, payload)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(body, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode response: %w", err)
	}
	return cfg, nil
}

// DeleteConfig stops the listener and removes stored credentials.
func DeleteConfig(studioId, integrationId string) error {
	url := fmt.Sprintf("%s/studio-integration/%s/%s", constants.HOST, studioId, integrationId)
	_, err := doRequest(http.MethodDelete, url, nil)
	return err
}

// TestConfig runs a live credential check without persisting changes.
// An empty payload falls back to the stored credentials on the server.
func TestConfig(studioId, integrationId string, payload CredentialsPayload) error {
	url := fmt.Sprintf("%s/studio-integration/%s/%s/test", constants.HOST, studioId, integrationId)
	_, err := doRequest(http.MethodPost, url, payload)
	return err
}

// SetEnabled toggles the enabled flag on an already-configured integration
// without re-supplying credentials. The server starts or stops the listener
// accordingly and returns the refreshed view.
func SetEnabled(studioId, integrationId string, enabled bool) (Config, error) {
	var cfg Config
	url := fmt.Sprintf("%s/studio-integration/%s/%s/enabled", constants.HOST, studioId, integrationId)
	body, err := doRequest(http.MethodPatch, url, map[string]bool{"enabled": enabled})
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(body, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode response: %w", err)
	}
	return cfg, nil
}

// doRequest performs an authenticated HTTP call to the central server and
// surfaces the server's error envelope for non-2xx responses.
func doRequest(method, url string, payload interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to encode payload: %w", err)
		}
		bodyReader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	auth_service.AttachBearerToken(req)
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var envelope errorEnvelope
		if json.Unmarshal(body, &envelope) == nil && envelope.Error != "" {
			return body, fmt.Errorf("%s", envelope.Error)
		}
		return body, fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return body, nil
}
