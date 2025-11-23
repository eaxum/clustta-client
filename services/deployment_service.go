package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"clustta/internal/auth_service"
	"clustta/internal/constants"
)

type DeploymentService struct{}

type DeploymentRequest struct {
	StudioName      string `json:"studio_name"`
	StudioURL       string `json:"studio_url"`
	StudioSecretKey string `json:"studio_secret_key"`
	AzureRegion     string `json:"azure_region"`
	VMSize          string `json:"vm_size"`
	DiskSizeGB      int    `json:"disk_size_gb"`
}

type DeploymentResponse struct {
	DeploymentID string    `json:"deployment_id"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	WebSocketURL string    `json:"websocket_url"`
}

type DeploymentStatus struct {
	DeploymentID    string     `json:"deployment_id"`
	StudioName      string     `json:"studio_name"`
	StudioURL       string     `json:"studio_url"`
	StudioSecretKey string     `json:"studio_secret_key"`
	AzureRegion     string     `json:"azure_region"`
	VMSize          string     `json:"vm_size"`
	DiskSizeGB      int        `json:"disk_size_gb"`
	Status          string     `json:"status"`
	Progress        int        `json:"progress"`
	CurrentStep     string     `json:"current_step"`
	VMName          string     `json:"vm_name"`
	ResourceGroup   string     `json:"resource_group"`
	PublicIP        string     `json:"public_ip"`
	CreatedAt       time.Time  `json:"created_at"`
	StartedAt       *time.Time `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	Error           string     `json:"error"`
}

// Deployment API endpoints
const (
	DeployEndpoint       = "/api/deploy"
	DeployStatusEndpoint = "/api/deploy"
	DeployDeleteEndpoint = "/api/deploy"
)

//DeployStudio initiates a new studio deployment on Azure.
//Returns the deployment response with ID and WebSocket URL, or an error if deployment fails.
func (d *DeploymentService) DeployStudio(request DeploymentRequest) (*DeploymentResponse, error) {
	token, err := auth_service.GetToken()
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", constants.HOST+DeployEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deployment failed with status %d: %s", resp.StatusCode, string(body))
	}

	var deploymentResponse DeploymentResponse
	if err := json.Unmarshal(body, &deploymentResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &deploymentResponse, nil
}

//GetDeploymentStatus retrieves the current status of a deployment.
//Returns the deployment status details or an error if the request fails.
func (d *DeploymentService) GetDeploymentStatus(deploymentID string) (*DeploymentStatus, error) {
	token, err := auth_service.GetToken()
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	req, err := http.NewRequest("GET", constants.HOST+DeployStatusEndpoint+"/"+deploymentID+"/status", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var status DeploymentStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &status, nil
}

//DestroyDeployment tears down a studio deployment and releases Azure resources.
//Returns an error if the destruction fails.
func (d *DeploymentService) DestroyDeployment(deploymentID string) error {
	token, err := auth_service.GetToken()
	if err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}

	req, err := http.NewRequest("DELETE", constants.HOST+DeployDeleteEndpoint+"/"+deploymentID, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("destroy request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
