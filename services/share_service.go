package services

import (
	"bytes"
	"clustta/internal/auth_service"
	"clustta/internal/constants"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ShareService struct{}

// ShareLinkResponse represents the response from creating a share link.
type ShareLinkResponse struct {
	Token     string `json:"token"`
	ShareURL  string `json:"share_url"`
	ExpiresAt string `json:"expires_at"`
}

// CreateShareLink creates a shareable download link for one or more checkpoints.
// Calls the global server directly with user Bearer token auth.
func (s *ShareService) CreateShareLink(studioId, projectName string, checkpointIds []string, label string, expiresInHours int) (ShareLinkResponse, error) {
	authHost := auth_service.GetAuthHost()
	if authHost == "" {
		return ShareLinkResponse{}, fmt.Errorf("cannot create share link in offline mode")
	}

	payload := map[string]interface{}{
		"studio_id":        studioId,
		"project_name":     projectName,
		"checkpoint_ids":   checkpointIds,
		"label":            label,
		"expires_in_hours": expiresInHours,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return ShareLinkResponse{}, fmt.Errorf("failed to marshal request: %v", err)
	}

	url := authHost + "/share"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return ShareLinkResponse{}, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	auth_service.AttachBearerToken(req)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ShareLinkResponse{}, fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ShareLinkResponse{}, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return ShareLinkResponse{}, fmt.Errorf("share link creation failed: %s", string(body))
	}

	var result ShareLinkResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return ShareLinkResponse{}, fmt.Errorf("failed to parse response: %v", err)
	}

	return result, nil
}

// RevokeShareLink revokes a share link.
// Calls the global server directly with user Bearer token auth.
func (s *ShareService) RevokeShareLink(token string) error {
	authHost := auth_service.GetAuthHost()
	if authHost == "" {
		return fmt.Errorf("cannot revoke share link in offline mode")
	}

	url := authHost + "/share/" + token
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	auth_service.AttachBearerToken(req)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to revoke share link: %s", string(body))
	}

	return nil
}
