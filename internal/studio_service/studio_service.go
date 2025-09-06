package studio_service

import (
	"bytes"
	"clustta/internal/auth_service"
	"clustta/internal/constants"
	"clustta/internal/server/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetUserStudios() ([]models.MinimalStudio, error) {
	url := constants.HOST + "/person/studios"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set custom headers
	token, err := auth_service.GetToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseCode := response.StatusCode
	if responseCode == 201 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			// Handle error
			return nil, fmt.Errorf("error reading response body: %s", err)
		}

		var studios []models.MinimalStudio
		err = json.Unmarshal(body, &studios)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
		}
		return studios, nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err.Error())
	}

	bodyData := string(body)

	return nil, fmt.Errorf("error loading studios: code - %d: body - %s", response.StatusCode, bodyData)
}

func GetUserPhoto(userId string) ([]byte, error) {
	url := constants.HOST + "/person/" + userId + "/photo"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set custom headers
	token, err := auth_service.GetToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseCode := response.StatusCode
	if responseCode == 200 || responseCode == 201 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %s", err)
		}

		// Check if we actually got photo data
		if len(body) == 0 {
			return nil, nil
		}
		return body, nil
	}

	return nil, nil // Return nil for non-200 responses (user has no photo)
}

func GetStudioUsers(studioId string) ([]models.StudioUserInfo, error) {
	url := constants.HOST + "/studio/" + studioId + "/persons"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set custom headers
	token, err := auth_service.GetToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseCode := response.StatusCode
	if responseCode == 201 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			// Handle error
			return nil, fmt.Errorf("error reading response body: %s", err)
		}

		var users []models.StudioUserInfo
		err = json.Unmarshal(body, &users)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
		}

		// Fetch photos for each user
		for i := range users {
			if users[i].Id != "" {
				photoData, err := GetUserPhoto(users[i].Id)
				if err == nil && photoData != nil {
					users[i].Photo = photoData
				}
			}
		}

		return users, nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err.Error())
	}

	bodyData := string(body)

	return nil, fmt.Errorf("error loading studio users: code - %d: body - %s", response.StatusCode, bodyData)
}

func AddCollaborator(email, studioId, roleName string) (interface{}, error) {
	url := constants.HOST + "/studio/person"

	requestBody := map[string]string{
		"email":     email,
		"role_name": roleName,
		"studio_id": studioId,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Set custom headers
	token, err := auth_service.GetToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseCode := response.StatusCode
	if responseCode == 201 || responseCode == 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %s", err)
		}

		var result interface{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
		}
		return result, nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err.Error())
	}

	bodyData := string(body)

	return nil, fmt.Errorf("error adding collaborator: code - %d: body - %s", response.StatusCode, bodyData)
}

func ChangeCollaboratorRole(userId, studioId, roleName string) (interface{}, error) {
	url := constants.HOST + "/studio/person"

	requestBody := map[string]string{
		"user_id":   userId,
		"role_name": roleName,
		"studio_id": studioId,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Set custom headers
	token, err := auth_service.GetToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseCode := response.StatusCode
	if responseCode == 201 || responseCode == 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %s", err)
		}

		var result interface{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
		}
		return result, nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err.Error())
	}

	bodyData := string(body)

	return nil, fmt.Errorf("error changing collaborator role: code - %d: body - %s", response.StatusCode, bodyData)
}

func RemoveCollaborator(userId, studioId string) (interface{}, error) {
	url := constants.HOST + "/studio/person/" + studioId + "/" + userId

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	// Set custom headers
	token, err := auth_service.GetToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseCode := response.StatusCode
	if responseCode == 201 || responseCode == 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %s", err)
		}

		var result interface{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
		}
		return result, nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err.Error())
	}

	bodyData := string(body)

	return nil, fmt.Errorf("error removing collaborator: code - %d: body - %s", response.StatusCode, bodyData)
}

func GetStudioStatus(studioUrl string) (string, error) {
	if studioUrl == "" {
		return "offline", nil
	}

	url := studioUrl + "/ping"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "offline", nil
	}

	// Set custom headers
	token, err := auth_service.GetToken()
	if err != nil {
		return "offline", nil
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "offline", nil
	}
	defer response.Body.Close()

	responseCode := response.StatusCode
	if responseCode == 201 || responseCode == 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return "offline", nil
		}

		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return "offline", nil
		}

		if status, ok := result["status"].(string); ok {
			return status, nil
		}
	}

	return "offline", nil
}

func CreateStudio(name string) (interface{}, error) {
	url := constants.HOST + "/studio"

	requestBody := map[string]string{
		"name": name,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Set custom headers
	token, err := auth_service.GetToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseCode := response.StatusCode
	if responseCode == 201 || responseCode == 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %s", err)
		}

		var result interface{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
		}
		return result, nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err.Error())
	}

	bodyData := string(body)

	return nil, fmt.Errorf("error creating studio: code - %d: body - %s", response.StatusCode, bodyData)
}
