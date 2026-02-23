package integrations

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// KitsuClient implements the Integration interface for Kitsu/CGWire.
type KitsuClient struct {
	httpClient *http.Client
}

// NewKitsuClient creates a new Kitsu integration client.
func NewKitsuClient() *KitsuClient {
	return &KitsuClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ID returns the integration identifier.
func (k *KitsuClient) ID() string {
	return "kitsu"
}

// Name returns the human-readable integration name.
func (k *KitsuClient) Name() string {
	return "Kitsu"
}

// GetInfo returns metadata about this integration.
func (k *KitsuClient) GetInfo() IntegrationInfo {
	return IntegrationInfo{
		ID:          "kitsu",
		Name:        "Kitsu",
		Description: "Production tracking for animation and VFX",
		Icon:        "kitsu",
		AuthType:    "password",
		Configured:  false,
	}
}

// Authenticate authenticates with Kitsu using email and password.
// Credentials expected: "email", "password", "api_url".
func (k *KitsuClient) Authenticate(credentials map[string]string) (AuthResult, error) {
	email := credentials["email"]
	password := credentials["password"]
	apiUrl := strings.TrimSuffix(credentials["api_url"], "/")

	if email == "" || password == "" || apiUrl == "" {
		return AuthResult{
			Success: false,
			Error:   "email, password, and api_url are required",
		}, errors.New("missing required credentials")
	}

	// Kitsu auth endpoint
	authUrl := apiUrl + "/api/auth/login"
	payload := map[string]string{
		"email":    email,
		"password": password,
	}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", authUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return AuthResult{Success: false, Error: err.Error()}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return AuthResult{Success: false, Error: "Failed to connect to Kitsu server"}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return AuthResult{
			Success: false,
			Error:   fmt.Sprintf("Authentication failed: %s", string(body)),
		}, errors.New("authentication failed")
	}

	var authResp kitsuAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return AuthResult{Success: false, Error: "Failed to parse response"}, err
	}

	// Extract access token from response
	return AuthResult{
		Success:      true,
		UserID:       authResp.User.ID,
		UserName:     authResp.User.FullName,
		UserEmail:    authResp.User.Email,
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    0, // Kitsu tokens don't typically expire
	}, nil
}

// ValidateToken checks if an existing token is still valid.
func (k *KitsuClient) ValidateToken(token, apiUrl string) (bool, error) {
	apiUrl = strings.TrimSuffix(apiUrl, "/")
	req, err := http.NewRequest("GET", apiUrl+"/api/auth/authenticated", nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// GetProjects fetches all projects the user has access to.
func (k *KitsuClient) GetProjects(token, apiUrl string) ([]ExternalProject, error) {
	apiUrl = strings.TrimSuffix(apiUrl, "/")
	data, err := k.get(token, apiUrl+"/api/data/projects")
	if err != nil {
		return nil, err
	}

	var kitsuProjects []kitsuProject
	if err := json.Unmarshal(data, &kitsuProjects); err != nil {
		return nil, err
	}

	projects := make([]ExternalProject, 0, len(kitsuProjects))
	for _, p := range kitsuProjects {
		projects = append(projects, ExternalProject{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		})
	}
	return projects, nil
}

// GetProjectEntities fetches all hierarchy entities (episodes, sequences, shots).
func (k *KitsuClient) GetProjectEntities(token, apiUrl, projectID string) ([]ExternalEntity, error) {
	apiUrl = strings.TrimSuffix(apiUrl, "/")
	entities := []ExternalEntity{}

	// Fetch episodes
	episodes, err := k.getEpisodes(token, apiUrl, projectID)
	if err != nil {
		return nil, err
	}
	entities = append(entities, episodes...)

	// Fetch sequences
	sequences, err := k.getSequences(token, apiUrl, projectID)
	if err != nil {
		return nil, err
	}
	entities = append(entities, sequences...)

	// Fetch shots
	shots, err := k.getShots(token, apiUrl, projectID)
	if err != nil {
		return nil, err
	}
	entities = append(entities, shots...)

	// Fetch assets (3D models, rigs, etc.)
	assets, err := k.getAssets(token, apiUrl, projectID)
	if err != nil {
		return nil, err
	}
	entities = append(entities, assets...)

	return entities, nil
}

// GetProjectTasks fetches all tasks from the project.
func (k *KitsuClient) GetProjectTasks(token, apiUrl, projectID string) ([]ExternalTask, error) {
	apiUrl = strings.TrimSuffix(apiUrl, "/")
	data, err := k.get(token, apiUrl+"/api/data/projects/"+projectID+"/tasks")
	if err != nil {
		return nil, err
	}

	var kitsuTasks []kitsuTask
	if err := json.Unmarshal(data, &kitsuTasks); err != nil {
		return nil, err
	}

	tasks := make([]ExternalTask, 0, len(kitsuTasks))
	for _, t := range kitsuTasks {
		tasks = append(tasks, ExternalTask{
			ID:          t.ID,
			ParentID:    t.EntityID,
			Name:        t.Name,
			Type:        "task",
			Status:      t.TaskStatusID,
			Assignees:   t.Assignees,
			TaskType:    t.TaskTypeName,
			Description: t.Description,
		})
	}
	return tasks, nil
}

// UpdateTaskStatus updates a task's status in Kitsu.
func (k *KitsuClient) UpdateTaskStatus(token, apiUrl, taskID, status string) error {
	apiUrl = strings.TrimSuffix(apiUrl, "/")
	payload := map[string]string{
		"task_status_id": status,
	}
	_, err := k.put(token, apiUrl+"/api/data/tasks/"+taskID, payload)
	return err
}

// UploadPreview uploads a preview file to a task.
func (k *KitsuClient) UploadPreview(token, apiUrl, taskID, filePath, comment string) error {
	apiUrl = strings.TrimSuffix(apiUrl, "/")

	// First, create a comment with the preview
	commentPayload := map[string]string{
		"task_id": taskID,
		"text":    comment,
	}
	commentData, err := k.post(token, apiUrl+"/api/data/comments", commentPayload)
	if err != nil {
		return err
	}

	var commentResp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(commentData, &commentResp); err != nil {
		return err
	}

	// Upload preview to the comment
	return k.uploadFile(token, apiUrl+"/api/data/comments/"+commentResp.ID+"/preview", filePath)
}

func (k *KitsuClient) get(token, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed (%d): %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (k *KitsuClient) post(token, url string, payload interface{}) ([]byte, error) {
	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed (%d): %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (k *KitsuClient) put(token, url string, payload interface{}) ([]byte, error) {
	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed (%d): %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (k *KitsuClient) uploadFile(token, url, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}
	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (k *KitsuClient) getEpisodes(token, apiUrl, projectID string) ([]ExternalEntity, error) {
	data, err := k.get(token, apiUrl+"/api/data/projects/"+projectID+"/episodes")
	if err != nil {
		return nil, err
	}

	var kitsuEpisodes []kitsuEntity
	if err := json.Unmarshal(data, &kitsuEpisodes); err != nil {
		return nil, err
	}

	entities := make([]ExternalEntity, 0, len(kitsuEpisodes))
	for _, e := range kitsuEpisodes {
		entities = append(entities, ExternalEntity{
			ID:       e.ID,
			ParentID: projectID,
			Name:     e.Name,
			Type:     "episode",
			Path:     e.Name,
			HasTasks: false,
		})
	}
	return entities, nil
}

func (k *KitsuClient) getSequences(token, apiUrl, projectID string) ([]ExternalEntity, error) {
	data, err := k.get(token, apiUrl+"/api/data/projects/"+projectID+"/sequences")
	if err != nil {
		return nil, err
	}

	var kitsuSequences []kitsuEntity
	if err := json.Unmarshal(data, &kitsuSequences); err != nil {
		return nil, err
	}

	entities := make([]ExternalEntity, 0, len(kitsuSequences))
	for _, s := range kitsuSequences {
		entities = append(entities, ExternalEntity{
			ID:       s.ID,
			ParentID: s.ParentID,
			Name:     s.Name,
			Type:     "sequence",
			Path:     s.Name, // Full path built later during sync
			HasTasks: false,
		})
	}
	return entities, nil
}

func (k *KitsuClient) getShots(token, apiUrl, projectID string) ([]ExternalEntity, error) {
	data, err := k.get(token, apiUrl+"/api/data/projects/"+projectID+"/shots")
	if err != nil {
		return nil, err
	}

	var kitsuShots []kitsuShot
	if err := json.Unmarshal(data, &kitsuShots); err != nil {
		return nil, err
	}

	entities := make([]ExternalEntity, 0, len(kitsuShots))
	for _, s := range kitsuShots {
		entities = append(entities, ExternalEntity{
			ID:       s.ID,
			ParentID: s.SequenceID,
			Name:     s.Name,
			Type:     "shot",
			Path:     s.Name,
			HasTasks: true,
		})
	}
	return entities, nil
}

func (k *KitsuClient) getAssets(token, apiUrl, projectID string) ([]ExternalEntity, error) {
	data, err := k.get(token, apiUrl+"/api/data/projects/"+projectID+"/assets")
	if err != nil {
		return nil, err
	}

	var kitsuAssets []kitsuAsset
	if err := json.Unmarshal(data, &kitsuAssets); err != nil {
		return nil, err
	}

	entities := make([]ExternalEntity, 0, len(kitsuAssets))
	for _, a := range kitsuAssets {
		entities = append(entities, ExternalEntity{
			ID:       a.ID,
			ParentID: a.AssetTypeID,
			Name:     a.Name,
			Type:     "asset",
			Path:     a.Name,
			HasTasks: true,
		})
	}
	return entities, nil
}

type kitsuAuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	User         kitsuUser `json:"user"`
}

type kitsuUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type kitsuProject struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type kitsuEntity struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parent_id"`
}

type kitsuShot struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	SequenceID string `json:"sequence_id"`
}

type kitsuAsset struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AssetTypeID string `json:"asset_type_id"`
}

type kitsuTask struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	EntityID     string   `json:"entity_id"`
	TaskTypeID   string   `json:"task_type_id"`
	TaskTypeName string   `json:"task_type_name"`
	TaskStatusID string   `json:"task_status_id"`
	Assignees    []string `json:"assignees"`
}

// init registers the Kitsu client.
func init() {
	Register(NewKitsuClient())
}
