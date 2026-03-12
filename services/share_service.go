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
// The studioUrl should be the resolved studio server URL (e.g. "http://studio.example.com").
// The projectName should be the project identifier (without .clst extension).
func (s *ShareService) CreateShareLink(studioUrl, projectName string, checkpointIds []string, label string, expiresInHours int) (ShareLinkResponse, error) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return ShareLinkResponse{}, fmt.Errorf("not authenticated: %v", err)
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		return ShareLinkResponse{}, fmt.Errorf("failed to marshal user: %v", err)
	}

	payload := map[string]interface{}{
		"checkpoint_ids":   checkpointIds,
		"label":            label,
		"expires_in_hours": expiresInHours,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return ShareLinkResponse{}, fmt.Errorf("failed to marshal request: %v", err)
	}

	url := studioUrl + "/" + projectName + "/share"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return ShareLinkResponse{}, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("UserData", string(userJson))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ShareLinkResponse{}, fmt.Errorf("failed to connect to studio server: %v", err)
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
