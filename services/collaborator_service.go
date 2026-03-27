package services

import (
	"bytes"
	"clustta/internal/auth_service"
	"clustta/internal/constants"
	"clustta/internal/repository/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CollaboratorService struct{}

// Collaborator represents a project collaborator returned by the server.
type Collaborator struct {
	Id        string `json:"id"`
	UserId    string `json:"user_id"`
	AddedAt   int64  `json:"added_at"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
}

// GetCollaborators fetches all collaborators for a personal remote project.
// The remoteUrl is the project's remote URL (e.g. http://host/user/{owner_id}/{project}).
func (c *CollaboratorService) GetCollaborators(remoteUrl string) ([]Collaborator, error) {
	url := remoteUrl + "/collaborators"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	auth_service.AttachBearerToken(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get collaborators: %s", string(body))
	}

	var collaborators []Collaborator
	err = json.NewDecoder(resp.Body).Decode(&collaborators)
	if err != nil {
		return nil, err
	}

	return collaborators, nil
}

// AddCollaborators adds one or more collaborators to a personal remote project by user IDs.
// Returns the results array from the server indicating which were added, skipped, or errored.
func (c *CollaboratorService) AddCollaborators(remoteUrl string, userIds []string) ([]map[string]string, error) {
	url := remoteUrl + "/collaborators"

	payload := struct {
		UserIds []string `json:"user_ids"`
	}{UserIds: userIds}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	auth_service.AttachBearerToken(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to add collaborators: %s", string(body))
	}

	var results []map[string]string
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// AddCollaboratorsWithRole adds collaborators to a project with a specific role.
// Used for studio projects where a per-project role is specified.
func (c *CollaboratorService) AddCollaboratorsWithRole(remoteUrl string, userIds []string, role string) ([]map[string]string, error) {
	url := remoteUrl + "/collaborators"

	payload := struct {
		UserIds []string `json:"user_ids"`
		Role    string   `json:"role"`
	}{UserIds: userIds, Role: role}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	auth_service.AttachBearerToken(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to add collaborators: %s", string(body))
	}

	var results []map[string]string
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// RemoveCollaborator removes a collaborator from a personal remote project by user ID.
func (c *CollaboratorService) RemoveCollaborator(remoteUrl, userId string) error {
	url := remoteUrl + "/collaborators/" + userId

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	auth_service.AttachBearerToken(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to remove collaborator: %s", string(body))
	}

	return nil
}

// FetchUserByEmail looks up a Clustta user by email address.
// Returns the user data if found, or an error if the user does not exist.
func (c *CollaboratorService) FetchUserByEmail(email string) (models.User, error) {
	user, err := auth_service.FetchUserData(email)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}
