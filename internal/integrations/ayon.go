package integrations

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	ayonRequestTimeout      = 15 * time.Second
	ayonWorkfileIDSeparator = "::"
)

// AYONClient implements project and folder discovery for AYON.
type AYONClient struct {
	httpClient *http.Client
}

type ayonUser struct {
	Name   string `json:"name"`
	Attrib struct {
		FullName string `json:"fullName"`
		Email    string `json:"email"`
	} `json:"attrib"`
}

type ayonProject struct {
	Name      string         `json:"name"`
	Code      string         `json:"code"`
	Active    bool           `json:"active"`
	TaskTypes []ayonTaskType `json:"taskTypes"`
	Statuses  []ayonStatus   `json:"statuses"`
}

type ayonHierarchyResponse struct {
	Hierarchy []ayonFolder `json:"hierarchy"`
}

type ayonFolder struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Label      string       `json:"label"`
	FolderType string       `json:"folderType"`
	ParentID   string       `json:"parentId"`
	HasTasks   bool         `json:"hasTasks"`
	Children   []ayonFolder `json:"children"`
}

type ayonTaskType struct {
	Name string `json:"name"`
}

type ayonStatus struct {
	Name      string   `json:"name"`
	ShortName string   `json:"shortName"`
	Color     string   `json:"color"`
	Scope     []string `json:"scope"`
}

// NewAYONClient creates an AYON integration client.
func NewAYONClient() *AYONClient {
	return &AYONClient{httpClient: &http.Client{Timeout: ayonRequestTimeout}}
}

func (a *AYONClient) ID() string {
	return "ayon"
}

func (a *AYONClient) Name() string {
	return "AYON"
}

func (a *AYONClient) GetInfo() IntegrationInfo {
	return IntegrationInfo{
		ID:          a.ID(),
		Name:        a.Name(),
		Description: "Production workflow with Clustta-backed workfiles",
		Icon:        "plug",
		AuthType:    "token",
	}
}

func (a *AYONClient) Authenticate(credentials map[string]string) (AuthResult, error) {
	apiURL := strings.TrimRight(strings.TrimSpace(credentials["api_url"]), "/")
	apiKey := strings.TrimSpace(credentials["api_key"])
	if apiURL == "" || apiKey == "" {
		return AuthResult{Success: false, Error: "AYON URL and API key are required"}, errors.New("AYON URL and API key are required")
	}
	user := ayonUser{}
	if err := a.get(apiKey, apiURL, "/api/profile", &user); err != nil {
		return AuthResult{Success: false, Error: err.Error()}, err
	}
	userName := user.Attrib.FullName
	if userName == "" {
		userName = user.Name
	}
	return AuthResult{
		Success:     true,
		UserID:      user.Name,
		UserName:    userName,
		UserEmail:   user.Attrib.Email,
		AccessToken: apiKey,
	}, nil
}

func (a *AYONClient) ValidateToken(token, apiURL string) (bool, error) {
	user := ayonUser{}
	if err := a.get(token, apiURL, "/api/profile", &user); err != nil {
		return false, err
	}
	return user.Name != "", nil
}

func (a *AYONClient) GetProjects(token, apiURL string) ([]ExternalProject, error) {
	var payload json.RawMessage
	if err := a.get(token, apiURL, "/api/projects?active=true", &payload); err != nil {
		return nil, err
	}
	projects := []ayonProject{}
	if err := decodeAYONList(payload, "projects", &projects); err != nil {
		return nil, err
	}
	result := make([]ExternalProject, 0, len(projects))
	for _, project := range projects {
		result = append(result, ExternalProject{
			ID:    project.Name,
			Name:  project.Name,
			Stats: project.Code,
		})
	}
	return result, nil
}

func (a *AYONClient) GetProjectCollections(token, apiURL, projectID string) ([]ExternalCollection, error) {
	hierarchy, err := a.getHierarchy(token, apiURL, projectID)
	if err != nil {
		return nil, err
	}
	result := make([]ExternalCollection, 0)
	for _, folder := range hierarchy {
		result = appendAYONFolderCollections(result, folder, "", "/")
	}
	return result, nil
}

// AYON tasks are created lazily by the Bridge, not during folder import.
func (a *AYONClient) GetProjectAssets(_, _, _ string) ([]ExternalAsset, error) {
	return []ExternalAsset{}, nil
}

func (a *AYONClient) GetCollectionTypes(token, apiURL, projectID string) ([]ExternalTypeInfo, error) {
	hierarchy, err := a.getHierarchy(token, apiURL, projectID)
	if err != nil {
		return nil, err
	}
	types := map[string]ExternalTypeInfo{}
	collectAYONFolderTypes(hierarchy, types)
	result := make([]ExternalTypeInfo, 0, len(types))
	for _, item := range types {
		result = append(result, item)
	}
	return result, nil
}

func (a *AYONClient) GetAssetTypes(token, apiURL, projectID string) ([]ExternalTypeInfo, error) {
	project, err := a.getProject(token, apiURL, projectID)
	if err != nil {
		return nil, err
	}
	result := make([]ExternalTypeInfo, 0, len(project.TaskTypes))
	for _, taskType := range project.TaskTypes {
		result = append(result, ExternalTypeInfo{
			ID:        taskType.Name,
			Name:      taskType.Name,
			ForEntity: "task",
		})
	}
	return result, nil
}

func (a *AYONClient) GetTaskStatuses(token, apiURL, projectID string) ([]ExternalStatusInfo, error) {
	project, err := a.getProject(token, apiURL, projectID)
	if err != nil {
		return nil, err
	}
	result := make([]ExternalStatusInfo, 0, len(project.Statuses))
	for _, status := range project.Statuses {
		if !containsAYONScope(status.Scope, "task") {
			continue
		}
		result = append(result, ExternalStatusInfo{
			ID:        status.Name,
			Name:      status.Name,
			ShortName: status.ShortName,
			Color:     status.Color,
		})
	}
	return result, nil
}

func (a *AYONClient) UpdateAssetStatus(token, apiURL, projectID, assetID, status string) error {
	taskID := strings.SplitN(assetID, ayonWorkfileIDSeparator, 2)[0]
	if taskID == "" {
		return errors.New("AYON task ID is required")
	}
	path := fmt.Sprintf(
		"/api/projects/%s/tasks/%s",
		url.PathEscape(projectID),
		url.PathEscape(taskID),
	)
	return a.patch(token, apiURL, path, map[string]string{"status": status})
}

func (a *AYONClient) UploadPreview(_, _, _, _, _, _ string, _ UploadProgressFunc) error {
	return errors.New("AYON preview uploads are not supported in the proof of concept")
}

func (a *AYONClient) getHierarchy(token, apiURL, projectID string) ([]ayonFolder, error) {
	path := fmt.Sprintf("/api/projects/%s/hierarchy", url.PathEscape(projectID))
	payload := ayonHierarchyResponse{}
	if err := a.get(token, apiURL, path, &payload); err != nil {
		return nil, err
	}
	return payload.Hierarchy, nil
}

func (a *AYONClient) getProject(token, apiURL, projectID string) (ayonProject, error) {
	path := fmt.Sprintf("/api/projects/%s", url.PathEscape(projectID))
	project := ayonProject{}
	if err := a.get(token, apiURL, path, &project); err != nil {
		return ayonProject{}, err
	}
	return project, nil
}

func containsAYONScope(scopes []string, expected string) bool {
	for _, scope := range scopes {
		if scope == expected {
			return true
		}
	}
	return false
}

func (a *AYONClient) get(token, apiURL, path string, target any) error {
	request, err := http.NewRequest(http.MethodGet, strings.TrimRight(apiURL, "/")+path, nil)
	if err != nil {
		return err
	}
	request.Header.Set("X-Api-Key", token)
	request.Header.Set("Accept", "application/json")
	response, err := a.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to connect to AYON: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("AYON request failed with HTTP %d", response.StatusCode)
	}
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("invalid AYON response: %w", err)
	}
	return nil
}

func (a *AYONClient) patch(token, apiURL, path string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(
		http.MethodPatch,
		strings.TrimRight(apiURL, "/")+path,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	request.Header.Set("X-Api-Key", token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	response, err := a.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to connect to AYON: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("AYON request failed with HTTP %d", response.StatusCode)
	}
	return nil
}

func decodeAYONList(payload json.RawMessage, key string, target any) error {
	if len(payload) == 0 {
		return errors.New("empty AYON response")
	}
	if payload[0] == '[' {
		return json.Unmarshal(payload, target)
	}
	wrapper := map[string]json.RawMessage{}
	if err := json.Unmarshal(payload, &wrapper); err != nil {
		return err
	}
	items, ok := wrapper[key]
	if !ok {
		return fmt.Errorf("AYON response does not contain %q", key)
	}
	return json.Unmarshal(items, target)
}

func appendAYONFolderCollections(
	collections []ExternalCollection,
	folder ayonFolder,
	parentID string,
	parentPath string,
) []ExternalCollection {
	name := folder.Label
	if name == "" {
		name = folder.Name
	}
	path := strings.TrimRight(parentPath, "/") + "/" + folder.Name + "/"
	if folder.ParentID != "" {
		parentID = folder.ParentID
	}
	collections = append(collections, ExternalCollection{
		ID:        folder.ID,
		ParentID:  parentID,
		Name:      name,
		Type:      folder.FolderType,
		Path:      path,
		HasAssets: folder.HasTasks,
	})
	for _, child := range folder.Children {
		collections = appendAYONFolderCollections(
			collections,
			child,
			folder.ID,
			path,
		)
	}
	return collections
}

func collectAYONFolderTypes(folders []ayonFolder, result map[string]ExternalTypeInfo) {
	for _, folder := range folders {
		if folder.FolderType != "" {
			result[folder.FolderType] = ExternalTypeInfo{
				ID:        folder.FolderType,
				Name:      folder.FolderType,
				ForEntity: "folder",
			}
		}
		collectAYONFolderTypes(folder.Children, result)
	}
}

func init() {
	Register(NewAYONClient())
}
