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

const (
	clickUpAPIBase = "https://api.clickup.com/api/v2"
)

// ClickUpClient implements the Integration interface for ClickUp.
type ClickUpClient struct {
	httpClient *http.Client
}

// NewClickUpClient creates a new ClickUp integration client.
func NewClickUpClient() *ClickUpClient {
	return &ClickUpClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ID returns the integration identifier.
func (c *ClickUpClient) ID() string {
	return "clickup"
}

// Name returns the human-readable integration name.
func (c *ClickUpClient) Name() string {
	return "ClickUp"
}

// GetInfo returns metadata about this integration.
func (c *ClickUpClient) GetInfo() IntegrationInfo {
	return IntegrationInfo{
		ID:          "clickup",
		Name:        "ClickUp",
		Description: "Project management and productivity",
		Icon:        "clickup",
		AuthType:    "oauth",
		Configured:  false,
	}
}

// Authenticate authenticates with ClickUp using OAuth token.
// For OAuth flow, frontend handles OAuth redirect and passes the token.
// Credentials expected: "access_token" (from OAuth callback).
func (c *ClickUpClient) Authenticate(credentials map[string]string) (AuthResult, error) {
	token := credentials["access_token"]

	if token == "" {
		return AuthResult{
			Success: false,
			Error:   "access_token is required",
		}, errors.New("missing access_token")
	}

	// Verify token by getting user info
	user, err := c.getAuthorizedUser(token)
	if err != nil {
		return AuthResult{
			Success: false,
			Error:   "Invalid token: " + err.Error(),
		}, err
	}

	return AuthResult{
		Success:     true,
		UserID:      fmt.Sprintf("%d", user.ID),
		UserName:    user.Username,
		UserEmail:   user.Email,
		AccessToken: token,
	}, nil
}

// ValidateToken checks if an existing token is still valid.
func (c *ClickUpClient) ValidateToken(token, apiUrl string) (bool, error) {
	_, err := c.getAuthorizedUser(token)
	return err == nil, err
}

// GetProjects fetches all workspaces and spaces the user has access to.
// In ClickUp: Workspace → Space → Folder → List → Task
func (c *ClickUpClient) GetProjects(token, apiUrl string) ([]ExternalProject, error) {
	// Get all teams (workspaces) first
	teams, err := c.getTeams(token)
	if err != nil {
		return nil, err
	}

	projects := []ExternalProject{}
	for _, team := range teams {
		// Get spaces for each team
		spaces, err := c.getSpaces(token, team.ID)
		if err != nil {
			continue
		}

		for _, space := range spaces {
			projects = append(projects, ExternalProject{
				ID:          space.ID,
				Name:        fmt.Sprintf("%s / %s", team.Name, space.Name),
				Description: fmt.Sprintf("Team: %s", team.Name),
			})
		}
	}

	return projects, nil
}

// GetProjectEntities fetches all folders and lists within a space.
func (c *ClickUpClient) GetProjectEntities(token, apiUrl, projectID string) ([]ExternalEntity, error) {
	entities := []ExternalEntity{}

	// Get folders in space
	folders, err := c.getFolders(token, projectID)
	if err != nil {
		return nil, err
	}

	for _, folder := range folders {
		entities = append(entities, ExternalEntity{
			ID:       folder.ID,
			ParentID: projectID,
			Name:     folder.Name,
			Type:     "folder",
			Path:     folder.Name,
			HasTasks: false,
		})

		// Get lists in folder
		for _, list := range folder.Lists {
			entities = append(entities, ExternalEntity{
				ID:       list.ID,
				ParentID: folder.ID,
				Name:     list.Name,
				Type:     "list",
				Path:     folder.Name + "/" + list.Name,
				HasTasks: true,
			})
		}
	}

	// Get folderless lists directly in space
	lists, err := c.getFolderlessLists(token, projectID)
	if err != nil {
		return nil, err
	}

	for _, list := range lists {
		entities = append(entities, ExternalEntity{
			ID:       list.ID,
			ParentID: projectID,
			Name:     list.Name,
			Type:     "list",
			Path:     list.Name,
			HasTasks: true,
		})
	}

	return entities, nil
}

// GetProjectTasks fetches all tasks from a space.
func (c *ClickUpClient) GetProjectTasks(token, apiUrl, projectID string) ([]ExternalTask, error) {
	tasks := []ExternalTask{}

	// Get all entities to find lists
	entities, err := c.GetProjectEntities(token, apiUrl, projectID)
	if err != nil {
		return nil, err
	}

	// Fetch tasks from each list
	for _, entity := range entities {
		if entity.Type == "list" {
			listTasks, err := c.getListTasks(token, entity.ID)
			if err != nil {
				continue
			}

			for _, t := range listTasks {
				assignees := make([]string, 0, len(t.Assignees))
				for _, a := range t.Assignees {
					assignees = append(assignees, a.Username)
				}

				dueDate := ""
				if t.DueDate != "" {
					dueDate = t.DueDate
				}

				tasks = append(tasks, ExternalTask{
					ID:        t.ID,
					ParentID:  entity.ID,
					Name:      t.Name,
					Type:      "task",
					Status:    t.Status.Status,
					Assignees: assignees,
					DueDate:   dueDate,
					TaskType:  entity.Name, // Use list name as task type
				})
			}
		}
	}

	return tasks, nil
}

// UpdateTaskStatus updates a task's status in ClickUp.
func (c *ClickUpClient) UpdateTaskStatus(token, apiUrl, taskID, status string) error {
	payload := map[string]interface{}{
		"status": status,
	}
	_, err := c.put(token, clickUpAPIBase+"/task/"+taskID, payload)
	return err
}

// UploadPreview uploads a preview file as an attachment to a task.
func (c *ClickUpClient) UploadPreview(token, apiUrl, taskID, filePath, comment string) error {
	// First add comment if provided
	if comment != "" {
		commentPayload := map[string]interface{}{
			"comment_text": comment,
		}
		_, err := c.post(token, clickUpAPIBase+"/task/"+taskID+"/comment", commentPayload)
		if err != nil {
			return err
		}
	}

	// Upload attachment
	return c.uploadAttachment(token, taskID, filePath)
}

// GetEntityTypes returns entity types for ClickUp.
// ClickUp uses Folders and Lists as organizational hierarchy.
func (c *ClickUpClient) GetEntityTypes(token, apiUrl, projectID string) ([]ExternalTypeInfo, error) {
	// ClickUp has fixed entity types
	return []ExternalTypeInfo{
		{ID: "folder", Name: "Folder"},
		{ID: "list", Name: "List"},
	}, nil
}

// GetTaskTypes returns task types for ClickUp.
// In ClickUp, task types are custom fields or task templates - not directly applicable.
// Returns empty as ClickUp doesn't have the same task type concept as Kitsu.
func (c *ClickUpClient) GetTaskTypes(token, apiUrl, projectID string) ([]ExternalTypeInfo, error) {
	// ClickUp tasks don't have types like Kitsu (Animation, Lighting, etc.)
	// Return a generic task type
	return []ExternalTypeInfo{
		{ID: "task", Name: "Task"},
	}, nil
}

func (c *ClickUpClient) get(token, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
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

func (c *ClickUpClient) post(token, url string, payload interface{}) ([]byte, error) {
	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
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

func (c *ClickUpClient) put(token, url string, payload interface{}) ([]byte, error) {
	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
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

func (c *ClickUpClient) uploadAttachment(token, taskID, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("attachment", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}
	writer.Close()

	req, err := http.NewRequest("POST", clickUpAPIBase+"/task/"+taskID+"/attachment", body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
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

func (c *ClickUpClient) getAuthorizedUser(token string) (*clickUpUser, error) {
	data, err := c.get(token, clickUpAPIBase+"/user")
	if err != nil {
		return nil, err
	}

	var resp struct {
		User clickUpUser `json:"user"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp.User, nil
}

func (c *ClickUpClient) getTeams(token string) ([]clickUpTeam, error) {
	data, err := c.get(token, clickUpAPIBase+"/team")
	if err != nil {
		return nil, err
	}

	var resp struct {
		Teams []clickUpTeam `json:"teams"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Teams, nil
}

func (c *ClickUpClient) getSpaces(token, teamID string) ([]clickUpSpace, error) {
	data, err := c.get(token, clickUpAPIBase+"/team/"+teamID+"/space")
	if err != nil {
		return nil, err
	}

	var resp struct {
		Spaces []clickUpSpace `json:"spaces"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Spaces, nil
}

func (c *ClickUpClient) getFolders(token, spaceID string) ([]clickUpFolder, error) {
	data, err := c.get(token, clickUpAPIBase+"/space/"+spaceID+"/folder")
	if err != nil {
		return nil, err
	}

	var resp struct {
		Folders []clickUpFolder `json:"folders"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Folders, nil
}

func (c *ClickUpClient) getFolderlessLists(token, spaceID string) ([]clickUpList, error) {
	data, err := c.get(token, clickUpAPIBase+"/space/"+spaceID+"/list")
	if err != nil {
		return nil, err
	}

	var resp struct {
		Lists []clickUpList `json:"lists"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Lists, nil
}

func (c *ClickUpClient) getListTasks(token, listID string) ([]clickUpTask, error) {
	data, err := c.get(token, clickUpAPIBase+"/list/"+listID+"/task")
	if err != nil {
		return nil, err
	}

	var resp struct {
		Tasks []clickUpTask `json:"tasks"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Tasks, nil
}

type clickUpUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type clickUpTeam struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type clickUpSpace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type clickUpFolder struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Lists []clickUpList `json:"lists"`
}

type clickUpList struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type clickUpTask struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Status    clickUpStatus `json:"status"`
	Assignees []clickUpUser `json:"assignees"`
	DueDate   string        `json:"due_date"`
}

type clickUpStatus struct {
	Status string `json:"status"`
}

// Ensure strings is imported (used by Kitsu but ClickUp uses constants).
var _ = strings.TrimSpace

// init registers the ClickUp client.
func init() {
	Register(NewClickUpClient())
}
