package integrations

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ayonRequestTimeout      = 15 * time.Second
	ayonWorkfileIDSeparator = "::"
	ayonTasksPageSize       = 300
	ayonTasksQuery          = `query TasksQuery($projectName: String!, $pageSize: Int!, $after: String) {
  project(name: $projectName) {
    tasks(first: $pageSize, after: $after) {
      edges {
        node {
          id
          name
          label
          taskType
          folderId
          status
          assignees
          active
        }
      }
      pageInfo {
        endCursor
        hasNextPage
      }
    }
  }
}`
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

type ayonTask struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Label     string   `json:"label"`
	TaskType  string   `json:"taskType"`
	FolderID  string   `json:"folderId"`
	Status    string   `json:"status"`
	Assignees []string `json:"assignees"`
	Active    bool     `json:"active"`
}

type ayonTasksResponse struct {
	Data struct {
		Project *struct {
			Tasks struct {
				Edges []struct {
					Node ayonTask `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					EndCursor   string `json:"endCursor"`
					HasNextPage bool   `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"tasks"`
		} `json:"project"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type ayonTasksVariables struct {
	ProjectName string  `json:"projectName"`
	PageSize    int     `json:"pageSize"`
	After       *string `json:"after"`
}

type ayonGraphQLRequest struct {
	Query     string             `json:"query"`
	Variables ayonTasksVariables `json:"variables"`
}

type ayonFileResponse struct {
	ID string `json:"id"`
}

type ayonEntityURIResponse struct {
	URIs []struct {
		ID  string `json:"id"`
		URI string `json:"uri"`
	} `json:"uris"`
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

func (a *AYONClient) GetProjectAssets(token, apiURL, projectID string) ([]ExternalAsset, error) {
	tasks, err := a.getTasks(token, apiURL, projectID)
	if err != nil {
		return nil, err
	}
	assets := make([]ExternalAsset, 0, len(tasks))
	for _, task := range tasks {
		if !task.Active {
			continue
		}
		name := task.TaskType
		if name == "" {
			name = task.Name
		}
		assets = append(assets, ExternalAsset{
			ID:          task.ID,
			ParentID:    task.FolderID,
			Name:        name,
			Type:        "asset",
			Status:      task.Status,
			Assignees:   task.Assignees,
			AssetType:   task.TaskType,
			AssetTypeID: task.TaskType,
			Description: task.Label,
		})
	}
	return assets, nil
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

func (a *AYONClient) UploadPreview(token, apiURL, projectID, assetID, filePath, comment, taskStatusID string, onProgress UploadProgressFunc) error {
	taskID := strings.SplitN(assetID, ayonWorkfileIDSeparator, 2)[0]
	if projectID == "" || taskID == "" {
		return errors.New("AYON project and task IDs are required")
	}

	fileID, err := a.uploadProjectFile(token, apiURL, projectID, filePath, onProgress)
	if err != nil {
		return fmt.Errorf("failed to upload AYON preview: %w", err)
	}

	activityPath := fmt.Sprintf(
		"/api/projects/%s/tasks/%s/activities",
		url.PathEscape(projectID),
		url.PathEscape(taskID),
	)
	activity := map[string]any{
		"activityType": "reviewable",
		"body":         comment,
		"files":        []string{fileID},
	}
	activityResponse := struct {
		ID string `json:"id"`
	}{}
	if err := a.post(token, apiURL, activityPath, activity, &activityResponse); err != nil {
		return fmt.Errorf("failed to create AYON reviewable: %w", err)
	}
	if strings.HasPrefix(mime.TypeByExtension(filepath.Ext(filePath)), "image/") {
		if err := a.uploadTaskThumbnail(token, apiURL, projectID, taskID, filePath); err != nil {
			return fmt.Errorf("failed to update AYON task thumbnail: %w", err)
		}
	}
	if taskStatusID == "" {
		return nil
	}
	return a.UpdateAssetStatus(token, apiURL, projectID, taskID, taskStatusID)
}

// GetExternalEntityURL returns the AYON browser URL for a task.
func (a *AYONClient) GetExternalEntityURL(token, apiURL, projectID, entityID string) (string, error) {
	path := fmt.Sprintf("/api/projects/%s/uris", url.PathEscape(projectID))
	response := ayonEntityURIResponse{}
	payload := map[string]any{
		"entityType": "task",
		"ids":        []string{strings.SplitN(entityID, ayonWorkfileIDSeparator, 2)[0]},
	}
	if err := a.post(token, apiURL, path, payload, &response); err != nil {
		return "", err
	}
	if len(response.URIs) != 1 || response.URIs[0].URI == "" {
		return "", errors.New("AYON task did not resolve to one entity URI")
	}

	browserURL, err := url.Parse(strings.TrimSpace(apiURL))
	if err != nil || (browserURL.Scheme != "http" && browserURL.Scheme != "https") || browserURL.Host == "" {
		return "", errors.New("invalid AYON URL")
	}
	browserURL.RawQuery = ""
	browserURL.Fragment = ""
	browserURL.Path = strings.TrimSuffix(browserURL.Path, "/") + fmt.Sprintf(
		"/projects/%s/browser",
		url.PathEscape(projectID),
	)
	query := browserURL.Query()
	query.Set("uri", response.URIs[0].URI)
	browserURL.RawQuery = query.Encode()
	return browserURL.String(), nil
}

func (a *AYONClient) uploadProjectFile(token, apiURL, projectID, filePath string, onProgress UploadProgressFunc) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	var body io.Reader = file
	if onProgress != nil {
		body = &progressReader{
			reader:     file,
			total:      fileInfo.Size(),
			onProgress: onProgress,
		}
	}
	path := fmt.Sprintf("/api/projects/%s/files", url.PathEscape(projectID))
	request, err := http.NewRequest(
		http.MethodPost,
		strings.TrimRight(apiURL, "/")+path,
		body,
	)
	if err != nil {
		return "", err
	}
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	request.Header.Set("X-Api-Key", token)
	request.Header.Set("X-File-Name", filepath.Base(filePath))
	request.Header.Set("Content-Type", contentType)
	request.Header.Set("Accept", "application/json")
	request.ContentLength = fileInfo.Size()

	response, err := a.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to connect to AYON: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(response.Body)
		return "", fmt.Errorf("AYON upload failed with HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	payload := ayonFileResponse{}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("invalid AYON upload response: %w", err)
	}
	if payload.ID == "" {
		return "", errors.New("AYON upload response did not contain a file ID")
	}
	return payload.ID, nil
}

func (a *AYONClient) uploadTaskThumbnail(token, apiURL, projectID, taskID, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	path := fmt.Sprintf(
		"/api/projects/%s/tasks/%s/thumbnail",
		url.PathEscape(projectID),
		url.PathEscape(taskID),
	)
	request, err := http.NewRequest(
		http.MethodPost,
		strings.TrimRight(apiURL, "/")+path,
		file,
	)
	if err != nil {
		return err
	}
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	request.Header.Set("X-Api-Key", token)
	request.Header.Set("Content-Type", contentType)
	request.Header.Set("Accept", "application/json")
	request.ContentLength = fileInfo.Size()

	response, err := a.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to connect to AYON: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(response.Body)
		return fmt.Errorf(
			"AYON thumbnail upload failed with HTTP %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(responseBody)),
		)
	}
	return nil
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

func (a *AYONClient) getTasks(token, apiURL, projectID string) ([]ayonTask, error) {
	tasks := make([]ayonTask, 0)
	var cursor *string
	for {
		payload := ayonGraphQLRequest{
			Query: ayonTasksQuery,
			Variables: ayonTasksVariables{
				ProjectName: projectID,
				PageSize:    ayonTasksPageSize,
				After:       cursor,
			},
		}
		response := ayonTasksResponse{}
		if err := a.post(token, apiURL, "/graphql", payload, &response); err != nil {
			return nil, err
		}
		if len(response.Errors) > 0 {
			return nil, fmt.Errorf("AYON GraphQL request failed: %s", response.Errors[0].Message)
		}
		if response.Data.Project == nil {
			return nil, fmt.Errorf("AYON project %q was not found", projectID)
		}

		page := response.Data.Project.Tasks
		for _, edge := range page.Edges {
			tasks = append(tasks, edge.Node)
		}
		if !page.PageInfo.HasNextPage {
			return tasks, nil
		}
		if page.PageInfo.EndCursor == "" {
			return nil, errors.New("AYON task pagination returned an empty cursor")
		}
		cursor = &page.PageInfo.EndCursor
	}
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

func (a *AYONClient) post(token, apiURL, path string, payload, target any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(
		http.MethodPost,
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
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("invalid AYON response: %w", err)
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
