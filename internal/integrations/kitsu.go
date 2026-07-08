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

// GetProjectCollections fetches all hierarchy collections (episodes, sequences, shots).
func (k *KitsuClient) GetProjectCollections(token, apiUrl, projectID string) ([]ExternalCollection, error) {
	apiUrl = strings.TrimSuffix(apiUrl, "/")
	collections := []ExternalCollection{}

	// Fetch asset types first to build lookup map for resolving asset type names
	assetTypeMap := make(map[string]string)
	data, err := k.get(token, apiUrl+"/api/data/asset-types")
	if err == nil {
		var assetTypes []kitsuAssetType
		if json.Unmarshal(data, &assetTypes) == nil {
			for _, at := range assetTypes {
				assetTypeMap[at.ID] = at.Name
			}
		}
	}

	// Create virtual folder collections for asset types (Characters, Props, etc.)
	// These serve as parent containers for assets
	for typeID, typeName := range assetTypeMap {
		collections = append(collections, ExternalCollection{
			ID:        "asset-type-" + typeID,
			ParentID:  "",
			Name:      typeName,
			Type:      "folder",
			Path:      typeName,
			HasAssets: false,
		})
	}

	// Fetch episodes
	episodes, err := k.getEpisodes(token, apiUrl, projectID)
	if err != nil {
		return nil, err
	}
	collections = append(collections, episodes...)

	// Fetch sequences
	sequences, err := k.getSequences(token, apiUrl, projectID)
	if err != nil {
		return nil, err
	}
	collections = append(collections, sequences...)

	// Fetch shots
	shots, err := k.getShots(token, apiUrl, projectID)
	if err != nil {
		return nil, err
	}
	collections = append(collections, shots...)

	// Fetch assets (3D models, rigs, etc.)
	assets, err := k.getAssets(token, apiUrl, projectID, assetTypeMap)
	if err != nil {
		return nil, err
	}
	collections = append(collections, assets...)

	return collections, nil
}

// GetProjectAssets fetches all assets from the project.
func (k *KitsuClient) GetProjectAssets(token, apiUrl, projectID string) ([]ExternalAsset, error) {
	apiUrl = strings.TrimSuffix(apiUrl, "/")

	// Fetch task types first to build lookup map
	taskTypeMap := make(map[string]string)
	typeData, err := k.get(token, apiUrl+"/api/data/task-types")
	if err == nil {
		var taskTypes []kitsuTaskType
		if json.Unmarshal(typeData, &taskTypes) == nil {
			for _, tt := range taskTypes {
				taskTypeMap[tt.ID] = tt.Name
			}
		}
	}

	data, err := k.get(token, apiUrl+"/api/data/projects/"+projectID+"/tasks")
	if err != nil {
		return nil, err
	}

	var kitsuTasks []kitsuTask
	if err := json.Unmarshal(data, &kitsuTasks); err != nil {
		return nil, err
	}

	assets := make([]ExternalAsset, 0, len(kitsuTasks))
	for _, t := range kitsuTasks {
		// Resolve task type name from ID if not provided
		taskTypeName := t.TaskTypeName
		if taskTypeName == "" {
			if name, ok := taskTypeMap[t.TaskTypeID]; ok {
				taskTypeName = name
			}
		}
		// Use task type as display name (e.g., "Animation" instead of "main")
		displayName := taskTypeName
		if displayName == "" {
			displayName = t.Name
		}
		assets = append(assets, ExternalAsset{
			ID:          t.ID,
			ParentID:    t.EntityID,
			Name:        displayName,
			Type:        "asset",
			Status:      t.TaskStatusID,
			Assignees:   t.Assignees,
			AssetType:   taskTypeName,
			AssetTypeID: t.TaskTypeID,
			Description: t.Description,
		})
	}
	return assets, nil
}

// UpdateAssetStatus updates an asset's status in Kitsu.
func (k *KitsuClient) UpdateAssetStatus(token, apiUrl, assetID, status string) error {
	apiUrl = strings.TrimSuffix(apiUrl, "/")

	// Kitsu status changes are done via the comment action endpoint
	payload := map[string]string{
		"task_status_id": status,
		"comment":        "",
	}
	_, err := k.post(token, apiUrl+"/api/actions/tasks/"+assetID+"/comment", payload)
	return err
}

// UploadPreview uploads a preview file to a task's comment.
// If taskStatusId is provided, the comment will also set that status. Otherwise the current status is preserved.
func (k *KitsuClient) UploadPreview(token, apiUrl, assetID, filePath, comment, taskStatusId string, onProgress UploadProgressFunc) error {
	apiUrl = strings.TrimSuffix(apiUrl, "/")

	// If no status provided, fetch the current task status to preserve it
	if taskStatusId == "" {
		taskData, err := k.get(token, apiUrl+"/api/data/tasks/"+assetID)
		if err != nil {
			return fmt.Errorf("failed to fetch task: %w", err)
		}
		var task struct {
			TaskStatusID string `json:"task_status_id"`
		}
		if err := json.Unmarshal(taskData, &task); err != nil {
			return fmt.Errorf("failed to parse task: %w", err)
		}
		taskStatusId = task.TaskStatusID
	}

	// Step 1: Create comment via the action endpoint
	commentPayload := map[string]string{
		"task_status_id": taskStatusId,
		"comment":        comment,
	}
	commentData, err := k.post(token, apiUrl+"/api/actions/tasks/"+assetID+"/comment", commentPayload)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	var commentResp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(commentData, &commentResp); err != nil {
		return fmt.Errorf("failed to parse comment response: %w", err)
	}

	// Step 2: Create a preview slot on the comment
	previewData, err := k.post(token, apiUrl+"/api/actions/tasks/"+assetID+"/comments/"+commentResp.ID+"/add-preview", map[string]string{})
	if err != nil {
		return fmt.Errorf("failed to create preview slot: %w", err)
	}

	var previewResp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(previewData, &previewResp); err != nil {
		return fmt.Errorf("failed to parse preview response: %w", err)
	}

	// Step 3: Upload the actual file to the preview slot
	return k.uploadFile(token, apiUrl+"/api/pictures/preview-files/"+previewResp.ID, filePath, onProgress)
}

// GetCollectionTypes fetches all collection types (asset types) from Kitsu.
// Kitsu has built-in collection types: Episode, Sequence, Shot plus user-defined asset types.
func (k *KitsuClient) GetCollectionTypes(token, apiUrl, projectID string) ([]ExternalTypeInfo, error) {
	apiUrl = strings.TrimSuffix(apiUrl, "/")
	types := []ExternalTypeInfo{}

	// Built-in collection types that Kitsu always has
	builtinTypes := []ExternalTypeInfo{
		{ID: "episode", Name: "Episode"},
		{ID: "sequence", Name: "Sequence"},
		{ID: "shot", Name: "Shot"},
	}
	types = append(types, builtinTypes...)

	// Fetch asset types (user-defined collection types like Character, Prop, Environment)
	data, err := k.get(token, apiUrl+"/api/data/asset-types")
	if err != nil {
		// If we can't fetch asset types, still return built-in types
		return types, nil
	}

	var kitsuAssetTypes []kitsuAssetType
	if err := json.Unmarshal(data, &kitsuAssetTypes); err != nil {
		return types, nil
	}

	for _, at := range kitsuAssetTypes {
		types = append(types, ExternalTypeInfo{
			ID:   at.ID,
			Name: at.Name,
		})
	}

	return types, nil
}

// GetAssetTypes fetches task types used in a specific Kitsu project.
// Only returns task types that have actual tasks in the project.
func (k *KitsuClient) GetAssetTypes(token, apiUrl, projectID string) ([]ExternalTypeInfo, error) {
	apiUrl = strings.TrimSuffix(apiUrl, "/")

	// Fetch all studio task types
	data, err := k.get(token, apiUrl+"/api/data/task-types")
	if err != nil {
		return nil, err
	}

	var kitsuTaskTypes []kitsuTaskType
	if err := json.Unmarshal(data, &kitsuTaskTypes); err != nil {
		return nil, err
	}

	// Build map of all task types
	allTaskTypes := make(map[string]kitsuTaskType)
	for _, tt := range kitsuTaskTypes {
		allTaskTypes[tt.ID] = tt
	}

	// Fetch tasks for this project to find which task types are actually used
	tasksData, err := k.get(token, apiUrl+"/api/data/projects/"+projectID+"/tasks")
	if err != nil {
		// If we can't get project tasks, fall back to all task types
		types := make([]ExternalTypeInfo, 0, len(kitsuTaskTypes))
		for _, tt := range kitsuTaskTypes {
			types = append(types, ExternalTypeInfo{
				ID:        tt.ID,
				Name:      tt.Name,
				ForEntity: tt.ForEntity,
			})
		}
		return types, nil
	}

	var projectTasks []kitsuTask
	if err := json.Unmarshal(tasksData, &projectTasks); err != nil {
		return nil, err
	}

	// Extract unique task type IDs from project tasks
	usedTaskTypeIDs := make(map[string]bool)
	for _, task := range projectTasks {
		usedTaskTypeIDs[task.TaskTypeID] = true
	}

	// Return only task types that are used in this project
	types := make([]ExternalTypeInfo, 0)
	for typeID := range usedTaskTypeIDs {
		if tt, exists := allTaskTypes[typeID]; exists {
			types = append(types, ExternalTypeInfo{
				ID:        tt.ID,
				Name:      tt.Name,
				ForEntity: tt.ForEntity,
			})
		}
	}

	return types, nil
}

// GetTaskStatuses fetches all task statuses from Kitsu.
func (k *KitsuClient) GetTaskStatuses(token, apiUrl string) ([]ExternalStatusInfo, error) {
	apiUrl = strings.TrimSuffix(apiUrl, "/")
	data, err := k.get(token, apiUrl+"/api/data/task-status")
	if err != nil {
		return nil, err
	}

	var kitsuStatuses []kitsuTaskStatus
	if err := json.Unmarshal(data, &kitsuStatuses); err != nil {
		return nil, err
	}

	statuses := make([]ExternalStatusInfo, 0, len(kitsuStatuses))
	for _, s := range kitsuStatuses {
		statuses = append(statuses, ExternalStatusInfo{
			ID:        s.ID,
			Name:      s.Name,
			ShortName: s.ShortName,
			Color:     s.Color,
		})
	}
	return statuses, nil
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

func (k *KitsuClient) uploadFile(token, url, filePath string, onProgress UploadProgressFunc) error {
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

	totalSize := int64(body.Len())
	var reqBody io.Reader = body
	if onProgress != nil {
		reqBody = &progressReader{reader: body, total: totalSize, onProgress: onProgress}
	}

	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.ContentLength = totalSize

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

func (k *KitsuClient) getEpisodes(token, apiUrl, projectID string) ([]ExternalCollection, error) {
	data, err := k.get(token, apiUrl+"/api/data/projects/"+projectID+"/episodes")
	if err != nil {
		return nil, err
	}

	var kitsuEpisodes []kitsuEntity
	if err := json.Unmarshal(data, &kitsuEpisodes); err != nil {
		return nil, err
	}

	collections := make([]ExternalCollection, 0, len(kitsuEpisodes))
	for _, e := range kitsuEpisodes {
		collections = append(collections, ExternalCollection{
			ID:        e.ID,
			ParentID:  projectID,
			Name:      e.Name,
			Type:      "episode",
			Path:      e.Name,
			HasAssets: false,
		})
	}
	return collections, nil
}

func (k *KitsuClient) getSequences(token, apiUrl, projectID string) ([]ExternalCollection, error) {
	data, err := k.get(token, apiUrl+"/api/data/projects/"+projectID+"/sequences")
	if err != nil {
		return nil, err
	}

	var kitsuSequences []kitsuEntity
	if err := json.Unmarshal(data, &kitsuSequences); err != nil {
		return nil, err
	}

	collections := make([]ExternalCollection, 0, len(kitsuSequences))
	for _, s := range kitsuSequences {
		collections = append(collections, ExternalCollection{
			ID:        s.ID,
			ParentID:  s.ParentID,
			Name:      s.Name,
			Type:      "sequence",
			Path:      s.Name, // Full path built later during sync
			HasAssets: false,
		})
	}
	return collections, nil
}

func (k *KitsuClient) getShots(token, apiUrl, projectID string) ([]ExternalCollection, error) {
	data, err := k.get(token, apiUrl+"/api/data/projects/"+projectID+"/shots")
	if err != nil {
		return nil, err
	}

	var kitsuShots []kitsuShot
	if err := json.Unmarshal(data, &kitsuShots); err != nil {
		return nil, err
	}

	collections := make([]ExternalCollection, 0, len(kitsuShots))
	for _, s := range kitsuShots {
		// Use sequence_id if available, otherwise fall back to parent_id
		parentID := s.SequenceID
		if parentID == "" {
			parentID = s.ParentID
		}
		collections = append(collections, ExternalCollection{
			ID:        s.ID,
			ParentID:  parentID,
			Name:      s.Name,
			Type:      "shot",
			Path:      s.Name,
			HasAssets: true,
		})
	}
	return collections, nil
}

func (k *KitsuClient) getAssets(token, apiUrl, projectID string, assetTypeMap map[string]string) ([]ExternalCollection, error) {
	data, err := k.get(token, apiUrl+"/api/data/projects/"+projectID+"/assets")
	if err != nil {
		return nil, err
	}

	var kitsuAssets []kitsuAsset
	if err := json.Unmarshal(data, &kitsuAssets); err != nil {
		return nil, err
	}

	collections := make([]ExternalCollection, 0, len(kitsuAssets))
	for _, a := range kitsuAssets {
		// Use asset type name directly from response, fallback to lookup, then default
		typeName := a.AssetTypeName
		if typeName == "" {
			if name, ok := assetTypeMap[a.AssetTypeID]; ok {
				typeName = name
			} else {
				typeName = "Asset"
			}
		}
		// Parent to virtual asset type folder (e.g., "asset-type-xxx" for Characters)
		parentID := ""
		if a.AssetTypeID != "" {
			parentID = "asset-type-" + a.AssetTypeID
		}
		collections = append(collections, ExternalCollection{
			ID:        a.ID,
			ParentID:  parentID,
			Name:      a.Name,
			Type:      typeName,
			Path:      a.Name,
			HasAssets: true,
		})
	}
	return collections, nil
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
	ParentID   string `json:"parent_id"`
}

type kitsuAsset struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	AssetTypeID   string `json:"entity_type_id"`
	AssetTypeName string `json:"asset_type_name"`
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

type kitsuAssetType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type kitsuTaskType struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ForEntity string `json:"for_entity"`
}

type kitsuTaskStatus struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
	Color     string `json:"color"`
}

// init registers the Kitsu client.
func init() {
	Register(NewKitsuClient())
}
