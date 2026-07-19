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

// StudioInfo represents metadata returned by a studio server's /studio-info endpoint
type StudioInfo struct {
	Id           string             `json:"id"`
	Name         string             `json:"name"`
	Url          string             `json:"url"`
	AltUrl       string             `json:"alt_url"`
	HostingMode  string             `json:"hosting_mode"`
	Capabilities StudioCapabilities `json:"capabilities"`
}

type StudioCapabilities struct {
	ProjectStorage ProjectStorageCapabilities `json:"project_storage"`
}

type ProjectStorageCapabilities struct {
	SupportedModes []string `json:"supported_modes"`
	AvailableModes []string `json:"available_modes"`
}

// StudioUsage represents VM-local usage metrics returned by a private studio server.
type StudioUsage struct {
	ProjectCount          int   `json:"project_count"`
	StorageBytes          int64 `json:"storage_bytes"`
	StorageAvailableBytes int64 `json:"storage_available_bytes"`
	StorageTotalBytes     int64 `json:"storage_total_bytes"`
}

// GetStudioInfo fetches studio metadata from a private studio server.
// Used when authenticated against a private server to discover its details.
func GetStudioInfo(studioUrl string) (StudioInfo, error) {
	if studioUrl == "" {
		return StudioInfo{}, fmt.Errorf("no studio URL provided")
	}

	url := studioUrl + "/studio-info"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return StudioInfo{}, err
	}

	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return StudioInfo{}, fmt.Errorf("failed to connect to studio server: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return StudioInfo{}, fmt.Errorf("error reading response body: %v", err)
		}

		var info StudioInfo
		err = json.Unmarshal(body, &info)
		if err != nil {
			return StudioInfo{}, fmt.Errorf("failed to parse studio info: %v", err)
		}

		// Use the request URL as fallback if server doesn't return URL
		if info.Url == "" {
			info.Url = studioUrl
		}

		return info, nil
	}

	return StudioInfo{}, fmt.Errorf("failed to get studio info: status code %d", response.StatusCode)
}

// UpdateStudioInfo writes new metadata (name, URL, alt URL, port) directly to a
// private studio server via PUT /studio-info. Used for renaming a self-hosted
// studio — the global Clustta server is not contacted.
// Pass "" for any field that should be left unchanged.
func UpdateStudioInfo(studioUrl, name, url, altUrl, port string) (StudioInfo, error) {
	if studioUrl == "" {
		return StudioInfo{}, fmt.Errorf("no studio URL provided")
	}

	payload := map[string]string{}
	if name != "" {
		payload["name"] = name
	}
	if url != "" {
		payload["url"] = url
	}
	payload["alt_url"] = altUrl
	if port != "" {
		payload["port"] = port
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return StudioInfo{}, err
	}

	req, err := http.NewRequest("PUT", studioUrl+"/studio-info", bytes.NewBuffer(body))
	if err != nil {
		return StudioInfo{}, err
	}

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return StudioInfo{}, fmt.Errorf("no active user: %v", err)
	}
	userJson, err := json.Marshal(user)
	if err != nil {
		return StudioInfo{}, err
	}

	auth_service.AttachBearerToken(req)
	req.Header.Set("UserData", string(userJson))
	req.Header.Set("UserId", user.Id)
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return StudioInfo{}, fmt.Errorf("failed to reach studio server: %v", err)
	}
	defer response.Body.Close()

	respBody, _ := io.ReadAll(response.Body)
	if response.StatusCode != http.StatusOK {
		return StudioInfo{}, fmt.Errorf("studio info update failed: code %d: %s", response.StatusCode, string(respBody))
	}

	var info StudioInfo
	if err := json.Unmarshal(respBody, &info); err != nil {
		return StudioInfo{}, fmt.Errorf("failed to parse response: %v", err)
	}
	return info, nil
}

// getEffectiveHost returns the appropriate API host based on auth mode.
// Returns empty string for offline mode (caller should handle this).
// Returns the auth host (private server URL) for studio mode.
// Returns constants.HOST (api.clustta.com) for global mode.
func getEffectiveHost() string {
	if auth_service.IsOfflineMode() {
		return ""
	}
	return auth_service.GetAuthHost()
}

// isGlobalMode returns true if in global authentication mode.
// Some operations (like registering studios) only make sense in global mode.
func isGlobalMode() bool {
	return auth_service.GetActiveAuthMode() == auth_service.AuthModeGlobal
}

// GetUserStudios fetches studios for the current user from the global Clustta server.
// Note: This only works in global mode. For studio mode, use GetStudioInfo instead.
func GetUserStudios() ([]models.MinimalStudio, error) {
	if !isGlobalMode() {
		return nil, fmt.Errorf("GetUserStudios is only available in global auth mode")
	}
	url := constants.HOST + "/person/studios"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set custom headers
	auth_service.AttachBearerToken(req)
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

// GetUserPhoto fetches a user's photo from the current auth server.
func GetUserPhoto(userId string) ([]byte, error) {
	host := getEffectiveHost()
	if host == "" {
		return nil, nil // No photos in offline mode
	}

	url := host + "/person/" + userId + "/photo"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set custom headers
	auth_service.AttachBearerToken(req)
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

// GetStudioUsers fetches all users for a studio from the current auth server.
// In studio mode, calls /studio/persons on the private server.
// In global mode, calls /studio/{studioId}/persons on the global server.
func GetStudioUsers(studioId string) ([]models.StudioUserInfo, error) {
	host := getEffectiveHost()
	if host == "" {
		return nil, fmt.Errorf("cannot fetch studio users in offline mode")
	}

	var url string
	if isGlobalMode() {
		url = host + "/studio/" + studioId + "/persons"
	} else {
		url = host + "/studio/persons"
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set custom headers
	auth_service.AttachBearerToken(req)
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

// AddCollaborator adds a collaborator to a studio on the current auth server.
func AddCollaborator(email, studioId, roleName string) (interface{}, error) {
	host := getEffectiveHost()
	if host == "" {
		return nil, fmt.Errorf("cannot add collaborator in offline mode")
	}

	url := host + "/studio/person"

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
	auth_service.AttachBearerToken(req)
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

// ChangeCollaboratorRole changes a collaborator's role on the current auth server.
// In studio mode, calls /studio/person/role on the private server.
// In global mode, calls /studio/person on the global server.
func ChangeCollaboratorRole(userId, studioId, roleName string) (interface{}, error) {
	host := getEffectiveHost()
	if host == "" {
		return nil, fmt.Errorf("cannot change collaborator role in offline mode")
	}

	var url string
	var requestBody map[string]string
	if isGlobalMode() {
		url = host + "/studio/person"
		requestBody = map[string]string{
			"user_id":   userId,
			"role_name": roleName,
			"studio_id": studioId,
		}
	} else {
		url = host + "/studio/person/role"
		requestBody = map[string]string{
			"user_id":   userId,
			"role_name": roleName,
		}
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
	auth_service.AttachBearerToken(req)
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

// RemoveCollaborator removes a collaborator from a studio on the current auth server.
// In studio mode, calls /studio/person/{userId} on the private server.
// In global mode, calls /studio/person/{studioId}/{userId} on the global server.
func RemoveCollaborator(userId, studioId string) (interface{}, error) {
	host := getEffectiveHost()
	if host == "" {
		return nil, fmt.Errorf("cannot remove collaborator in offline mode")
	}

	var url string
	if isGlobalMode() {
		url = host + "/studio/person/" + studioId + "/" + userId
	} else {
		url = host + "/studio/person/" + userId
	}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	// Set custom headers
	auth_service.AttachBearerToken(req)
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

func GetStudioUsage(studioUrl string) (StudioUsage, error) {
	if studioUrl == "" {
		return StudioUsage{}, fmt.Errorf("no studio URL provided")
	}

	url := studioUrl + "/usage"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return StudioUsage{}, err
	}

	auth_service.AttachBearerToken(req)
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return StudioUsage{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return StudioUsage{}, fmt.Errorf("failed to get studio usage: status code %d: %s", response.StatusCode, string(body))
	}

	var usage StudioUsage
	if err := json.NewDecoder(response.Body).Decode(&usage); err != nil {
		return StudioUsage{}, fmt.Errorf("failed to decode studio usage: %w", err)
	}

	return usage, nil
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
	auth_service.AttachBearerToken(req)
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

// RegisterStudio registers a new studio on the global Clustta server.
// This operation is only available in global auth mode.
func RegisterStudio(studioName, studioUrl, hostingMode string) (interface{}, error) {
	if !isGlobalMode() {
		return nil, fmt.Errorf("studio registration is only available in global auth mode")
	}

	url := constants.HOST + "/studio"

	requestBody := map[string]string{
		"name":         studioName,
		"url":          studioUrl,
		"hosting_mode": hostingMode,
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
	auth_service.AttachBearerToken(req)
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

// UpdateStudio updates a studio's configuration on the global Clustta server.
// This operation is only available in global auth mode.
func UpdateStudio(studioName, url, altUrl, port, key string) (interface{}, error) {
	if !isGlobalMode() {
		return nil, fmt.Errorf("studio update is only available in global auth mode")
	}

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
	auth_service.AttachBearerToken(req)
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

// VerifyDeploymentCode verifies a deployment code on the global Clustta server.
// This operation is only available in global auth mode.
func VerifyDeploymentCode(code string) (bool, string, error) {
	if !isGlobalMode() {
		return false, "", fmt.Errorf("deployment code verification is only available in global auth mode")
	}

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
	auth_service.AttachBearerToken(req)
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

// CheckStudioNameExists checks if a studio name is available on the global Clustta server.
// This operation is only available in global auth mode.
func CheckStudioNameExists(studioName string) (bool, error) {
	if !isGlobalMode() {
		return false, fmt.Errorf("studio name check is only available in global auth mode")
	}

	url := constants.HOST + "/check-studio-availability/" + studioName

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	// Set custom headers
	auth_service.AttachBearerToken(req)
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

	auth_service.AttachBearerToken(req)
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
