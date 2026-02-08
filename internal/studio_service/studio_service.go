package studio_service

import (
	"bytes"
	"clustta/internal/auth_service"
	"clustta/internal/constants"
	"clustta/internal/server/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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
		// for i := range users {
		// 	if users[i].Id != "" {
		// 		photoData, err := GetUserPhoto(users[i].Id)
		// 		if err == nil && photoData != nil {
		// 			users[i].Photo = photoData
		// 		}
		// 	}
		// }

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

func GetServerVersion(studioUrl string) (string, error) {
	if studioUrl == "" {
		return "", fmt.Errorf("no studio URL provided")
	}

	url := studioUrl + "/version"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode == 200 || response.StatusCode == 201 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return "", err
		}

		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return "", err
		}

		if version, ok := result["version"].(string); ok {
			return version, nil
		}
		return "", fmt.Errorf("version not found in response")
	}

	return "", fmt.Errorf("failed to get server version: status code %d", response.StatusCode)
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

func RegisterStudio(studioName, studioUrl string) (interface{}, error) {
	url := constants.HOST + "/studio"

	requestBody := map[string]string{
		"name": studioName,
		"url":  studioUrl,
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

func UpdateStudio(studioName, url, altUrl, port, key string) (interface{}, error) {
	apiUrl := constants.HOST + "/studio/" + studioName + "/url"

	requestBody := map[string]string{
		"url":     url,
		"alt_url": altUrl,
		"port":    port,
		"key":     key,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", apiUrl, bytes.NewBuffer(jsonData))
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

	return nil, fmt.Errorf("error updating studio: code - %d: body - %s", response.StatusCode, bodyData)
}

func VerifyDeploymentCode(code string) (bool, string, error) {
	url := constants.HOST + "/studio/verify-deployment-code"

	requestBody := map[string]string{
		"code": code,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return false, "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, "", err
	}

	// Set custom headers
	token, err := auth_service.GetToken()
	if err != nil {
		return false, "", err
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return false, "", fmt.Errorf("error reading response body: %s", err)
	}

	responseCode := response.StatusCode
	if responseCode == 200 || responseCode == 201 {
		var result struct {
			Valid   bool   `json:"valid"`
			Message string `json:"message"`
		}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return false, "", fmt.Errorf("failed to unmarshal response body: %v", err)
		}
		return result.Valid, result.Message, nil
	}

	bodyData := string(body)
	return false, "", fmt.Errorf("error verifying deployment code: code - %d: body - %s", response.StatusCode, bodyData)
}

func CheckStudioNameExists(studioName string) (bool, error) {
	url := constants.HOST + "/check-studio-availability/" + studioName

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	// Set custom headers
	token, err := auth_service.GetToken()
	if err != nil {
		return false, err
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return false, fmt.Errorf("error reading response body: %s", err)
	}

	responseCode := response.StatusCode
	if responseCode == 200 || responseCode == 201 {
		var result struct {
			StudioNameExist bool `json:"studio_name_exist"`
		}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return false, fmt.Errorf("failed to unmarshal response body: %v", err)
		}
		return result.StudioNameExist, nil
	}

	bodyData := string(body)
	return false, fmt.Errorf("error checking studio name: code - %d: body - %s", response.StatusCode, bodyData)
}

// pingStudioUrl pings a studio URL and sends the URL back on the channel if reachable.
// Uses a short timeout to fail fast for unreachable servers.
func pingStudioUrl(ctx context.Context, studioUrl string, ch chan<- string) {
	pingUrl := studioUrl + "/ping"

	req, err := http.NewRequestWithContext(ctx, "GET", pingUrl, nil)
	if err != nil {
		return
	}

	token, err := auth_service.GetToken()
	if err != nil {
		return
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK || response.StatusCode == http.StatusCreated {
		select {
		case ch <- studioUrl:
		case <-ctx.Done():
		}
	}
}

// ResolveStudioUrl races the primary and alternative studio URLs.
// Returns whichever responds successfully first. Falls back to url if altUrl is empty.
func ResolveStudioUrl(url, altUrl string) (string, error) {
	if altUrl == "" {
		return url, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	ch := make(chan string, 2)

	go pingStudioUrl(ctx, url, ch)
	go pingStudioUrl(ctx, altUrl, ch)

	select {
	case winner := <-ch:
		return winner, nil
	case <-ctx.Done():
		return url, fmt.Errorf("both studio URLs are unreachable")
	}
}
