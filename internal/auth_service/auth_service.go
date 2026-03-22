package auth_service

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"clustta/internal/constants"
	"clustta/internal/error_service"
	"clustta/internal/repository/models"
)

//go:embed sso_callback.html
var ssoCallbackHTML []byte

// openBrowser opens the specified URL in the system's default browser.
func openBrowser(url string) {
	switch runtime.GOOS {
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	default:
		exec.Command("xdg-open", url).Start()
	}
}

type Token struct {
	SessionId string `json:"session_id"`
	User      User   `json:"user"`
}

type User struct {
	Id        string `db:"id" json:"id"`
	Username  string `db:"username" json:"username"`
	Email     string `db:"email" json:"email"`
	FirstName string `db:"first_name" json:"first_name"`
	LastName  string `db:"last_name" json:"last_name"`
	Photo     []byte `db:"photo" json:"photo"`
}

// AttachBearerToken adds the Authorization header with the active session token.
// Safe to call even when not authenticated (no-op if no token available).
func AttachBearerToken(req *http.Request) {
	token, err := GetToken()
	if err != nil || token.SessionId == "" {
		return
	}
	req.Header.Set("Authorization", "Bearer "+token.SessionId)
}

func GetActiveUser() (User, error) {
	// In offline mode, always return the canonical offline user
	if IsOfflineMode() {
		return OfflineUser(), nil
	}

	token, err := GetToken()
	if err != nil {
		return User{}, err
	}
	return token.User, nil
}

func IsAuthenticated() (bool, error) {
	// Check if in offline mode first
	if IsOfflineMode() {
		return true, nil
	}

	type responseMessage struct {
		Message string `json: "message" `
	}

	authHost := GetAuthHost()
	if authHost == "" {
		return false, fmt.Errorf("no auth host configured")
	}

	url := authHost + "/auth/authenticated"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	// Set custom headers
	token, err := GetToken()
	if err != nil {
		return false, err
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	responseCode := response.StatusCode
	if responseCode == 200 {
		return true, nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return false, fmt.Errorf("error reading response body: %s", err.Error())
	}
	message := responseMessage{}
	err = json.Unmarshal(body, &message)
	if err != nil {
		return false, err
	}
	bodyData := string(body)
	if message.Message == "Unauthorized" {
		return false, error_service.ErrNotUnauthorized
	}

	return false, fmt.Errorf("error loading user: code - %d: body - %s", response.StatusCode, bodyData)
}

// userLookupPrefix returns the URL path prefix for user lookup endpoints.
// Studio servers use /auth prefix to avoid route conflicts with project wildcards.
func userLookupPrefix() string {
	if GetActiveAuthMode() == AuthModeStudio {
		return "/auth"
	}
	return ""
}

func FetchUserPhoto(userId string) ([]byte, error) {
	authHost := GetAuthHost()
	if authHost == "" {
		return []byte{}, nil // No photo in offline mode
	}
	url := authHost + userLookupPrefix() + "/person/" + userId + "/photo"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err
	}

	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer response.Body.Close()

	responseCode := response.StatusCode
	if responseCode == 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			// Handle error
			return []byte{}, fmt.Errorf("error reading response body: %s", err)
		}
		return body, nil
	}

	// No photo available
	if responseCode == 204 {
		return []byte{}, nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		// Handle error
		return []byte{}, fmt.Errorf("error reading response body: %s", err.Error())
	}
	bodyData := string(body)
	if strings.Contains(bodyData, "Unauthorized") {
		return []byte{}, error_service.ErrNotAutheticated
	}

	return []byte{}, fmt.Errorf("error loading user: code - %d: body - %s", response.StatusCode, bodyData)
}

func FetchUserData(email string) (models.User, error) {
	authHost := GetAuthHost()
	if authHost == "" {
		return models.User{}, fmt.Errorf("cannot fetch user data in offline mode")
	}
	url := authHost + userLookupPrefix() + "/person/" + email

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return models.User{}, err
	}

	// Set custom headers
	// token, err := GetToken()
	// if err != nil {
	// 	return models.User{}, err
	// }
	// req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return models.User{}, err
	}
	defer response.Body.Close()

	responseCode := response.StatusCode
	if responseCode == 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			// Handle error
			return models.User{}, fmt.Errorf("error reading response body: %s", err)
		}

		var user models.User
		err = json.Unmarshal(body, &user)
		if err != nil {
			return models.User{}, fmt.Errorf("Failed to unmarshal response body: %v", err)
		}
		userPhoto, err := FetchUserPhoto(user.Id)
		if err != nil {
			return models.User{}, err
		}
		user.Photo = userPhoto
		return user, nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		// Handle error
		return models.User{}, fmt.Errorf("error reading response body: %s", err.Error())
	}
	bodyData := string(body)
	if strings.Contains(bodyData, "Unauthorized") {
		return models.User{}, error_service.ErrNotAutheticated
	}

	return models.User{}, fmt.Errorf("error loading user: code - %d: body - %s", response.StatusCode, bodyData)
}

func FetchUserDataById(userId string) (models.User, error) {
	authHost := GetAuthHost()
	if authHost == "" {
		return models.User{}, fmt.Errorf("cannot fetch user data in offline mode")
	}
	url := authHost + userLookupPrefix() + "/persons/" + userId

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return models.User{}, err
	}

	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return models.User{}, err
	}
	defer response.Body.Close()

	responseCode := response.StatusCode
	if responseCode == 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return models.User{}, fmt.Errorf("error reading response body: %s", err)
		}

		var user models.User
		err = json.Unmarshal(body, &user)
		if err != nil {
			return models.User{}, fmt.Errorf("Failed to unmarshal response body: %v", err)
		}
		userPhoto, err := FetchUserPhoto(user.Id)
		if err != nil {
			return models.User{}, err
		}
		user.Photo = userPhoto
		return user, nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return models.User{}, fmt.Errorf("error reading response body: %s", err.Error())
	}
	bodyData := string(body)
	if strings.Contains(bodyData, "Unauthorized") {
		return models.User{}, error_service.ErrNotAutheticated
	}

	return models.User{}, fmt.Errorf("error loading user: code - %d: body - %s", response.StatusCode, bodyData)
}

func GetToken() (Token, error) {
	// Get the active account from multi-account structure
	activeToken, err := GetActiveAccount()
	if err != nil {
		return Token{}, err
	}

	return activeToken, nil
}

func SetToken(
	token Token,
) error {
	// Store in multi-account structure
	err := AddAccount(token)
	if err != nil {
		return err
	}

	return nil
}

func DeleteToken() error {
	// Get current active account
	activeToken, err := GetActiveAccount()
	if err != nil {
		// No active account to delete
		return nil
	}

	// Remove the active account from multi-account structure
	err = RemoveAccount(activeToken.User.Id)
	if err != nil {
		return err
	}

	return nil
}

// Login authenticates a user against Clustta Cloud (api.clustta.com)
func Login(username string, password string) (Token, error) {
	return LoginWithHost(username, password, DefaultAuthHost, AuthModeGlobal, "")
}

// LoginWithHost authenticates a user against a specified authentication host
func LoginWithHost(username string, password string, authHost string, authMode AuthMode, studioId string) (Token, error) {
	if authHost == "" {
		authHost = DefaultAuthHost
	}

	url := authHost + "/auth/login"
	jsonBody := fmt.Sprintf("{\"email\": \"%s\", \"password\": \"%s\"}", username, password)
	response, err := http.Post(url, "application/json", strings.NewReader(jsonBody))
	if err != nil {
		return Token{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return Token{}, err
		}
		return Token{}, errors.New(string(body))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return Token{}, err
	}

	token := Token{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		return token, err
	}

	cookies := response.Cookies()
	for _, c := range cookies {
		if c.Name == "session" {
			token.SessionId = c.Value
		}
	}

	// Create AccountToken with full auth context and store it
	accountToken := FromToken(token, authMode, authHost, studioId)
	err = AddAccountToken(accountToken)
	if err != nil {
		return Token{}, err
	}
	return token, nil
}

// LoginWithSSO initiates Google SSO by opening the system browser and waiting for the callback.
func LoginWithSSO(authHost string) (Token, error) {
	if authHost == "" {
		authHost = DefaultAuthHost
	}

	resultCh := make(chan ssoResult, 1)

	// Start a local HTTP server on a random port to receive the callback
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return Token{}, fmt.Errorf("failed to start local callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("GET /sso/callback", func(w http.ResponseWriter, r *http.Request) {
		sessionId := r.URL.Query().Get("session_id")
		userB64 := r.URL.Query().Get("user")

		if sessionId == "" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte("<html><body><h2>Login failed</h2><p>No session received. Please try again.</p></body></html>"))
			resultCh <- ssoResult{err: fmt.Errorf("no session_id in callback")}
			return
		}

		var user User
		if userB64 != "" {
			userJSON, err := base64.URLEncoding.DecodeString(userB64)
			if err == nil {
				json.Unmarshal(userJSON, &user)
			}
		}

		token := Token{
			SessionId: sessionId,
			User:      user,
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write(ssoCallbackHTML)
		resultCh <- ssoResult{token: token}
	})

	server := &http.Server{Handler: mux}
	go server.Serve(listener)

	// Open the system browser to the SSO URL
	ssoURL := fmt.Sprintf("%s/auth/sso/google?redirect_port=%d", authHost, port)
	openBrowser(ssoURL)

	// Wait for the callback (with timeout)
	select {
	case result := <-resultCh:
		server.Close()
		if result.err != nil {
			return Token{}, result.err
		}
		// Store the token
		accountToken := FromToken(result.token, AuthModeGlobal, authHost, "")
		err = AddAccountToken(accountToken)
		if err != nil {
			return Token{}, err
		}
		return result.token, nil
	case <-time.After(5 * time.Minute):
		server.Close()
		return Token{}, fmt.Errorf("SSO login timed out")
	}
}

type ssoResult struct {
	token Token
	err   error
}

// Register creates a new user account on Clustta Cloud
func Register(firstName, lastName, username, email, password, confirmPassword string) (User, error) {
	return RegisterWithHost(firstName, lastName, username, email, password, confirmPassword, DefaultAuthHost)
}

// RegisterWithHost creates a new user account on a specified authentication host
func RegisterWithHost(firstName, lastName, username, email, password, confirmPassword, authHost string) (User, error) {
	if authHost == "" {
		authHost = DefaultAuthHost
	}

	url := authHost + "/auth/register"
	jsonBody := fmt.Sprintf("{\"first_name\": \"%s\", \"last_name\": \"%s\", \"username\": \"%s\", \"email\": \"%s\", \"password\": \"%s\", \"confirm_password\": \"%s\"}",
		firstName, lastName, username, email, password, confirmPassword)
	response, err := http.Post(url, "application/json", strings.NewReader(jsonBody))
	if err != nil {
		return User{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != 201 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return User{}, err
		}
		return User{}, errors.New(string(body))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return User{}, err
	}

	var responseData struct {
		Data User `json:"data"`
	}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return User{}, err
	}

	return responseData.Data, nil
}

func UpdateUser(firstName, lastName, username, email string) (User, error) {
	authHost := GetAuthHost()
	if authHost == "" {
		return User{}, fmt.Errorf("cannot update user in offline mode")
	}

	url := authHost + "/person/update"
	jsonBody := fmt.Sprintf("{\"first_name\": \"%s\", \"last_name\": \"%s\", \"username\": \"%s\", \"email\": \"%s\"}",
		firstName, lastName, username, email)

	req, err := http.NewRequest("PUT", url, strings.NewReader(jsonBody))
	if err != nil {
		return User{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	// Attach session cookie
	token, err := GetToken()
	if err != nil {
		return User{}, err
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return User{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != 201 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return User{}, err
		}
		return User{}, errors.New(string(body))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return User{}, err
	}

	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func Logout() error {
	// If in offline mode, just remove the offline account
	if IsOfflineMode() {
		return DeleteToken()
	}

	authHost := GetAuthHost()
	if authHost == "" {
		return DeleteToken()
	}

	url := authHost + "/auth/logout"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Set custom headers
	token, err := GetToken()
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseCode := response.StatusCode
	if responseCode != 200 {
		_, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		err = DeleteToken()
		if err != nil {
			return err
		}
		return nil
	}

	err = DeleteToken()
	if err != nil {
		return err
	}
	return nil
}

func CheckUsernameExists(username string) (bool, error) {
	authHost := GetAuthHost()
	if authHost == "" {
		return false, fmt.Errorf("cannot check username in offline mode")
	}

	url := authHost + "/auth/username-exists/" + username
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	var result struct {
		UsernameExist bool `json:"username_exist"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return false, err
	}
	return result.UsernameExist, nil
}

func CheckEmailExists(email string) (bool, error) {
	authHost := GetAuthHost()
	if authHost == "" {
		return false, fmt.Errorf("cannot check email in offline mode")
	}

	url := authHost + "/auth/email-exists/" + email
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	var result struct {
		EmailExist bool `json:"email_exist"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return false, err
	}
	return result.EmailExist, nil
}

func UpdateUserPhoto(photo []byte) error {
	authHost := GetAuthHost()
	if authHost == "" {
		return fmt.Errorf("cannot update photo in offline mode")
	}

	url := authHost + "/person/photo"
	fmt.Println(url)
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("photo", "profile.jpg")
	if err != nil {
		return err
	}
	_, err = fw.Write(photo)
	if err != nil {
		return err
	}
	w.Close()

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	token, err := GetToken()
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update photo: %s", string(body))
	}

	return nil
}

func DeactivateUserAccount() error {
	authHost := GetAuthHost()
	if authHost == "" {
		return fmt.Errorf("cannot deactivate account in offline mode")
	}

	url := authHost + "/person/deactivate-account"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	// Attach session cookie
	token, err := GetToken()
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete account: %s", string(body))
	}
	return nil
}

func SendInvitationEmail(email, studioName, projectName string) error {
	type requestData struct {
		Email       string `json:"email"`
		StudioName  string `json:"studio_name"`
		ProjectName string `json:"project_name"`
	}

	data := requestData{
		Email:       email,
		StudioName:  studioName,
		ProjectName: projectName,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %v", err)
	}

	authHost := GetAuthHost()
	if authHost == "" {
		return fmt.Errorf("cannot send invitation in offline mode")
	}

	url := authHost + "/auth/send-invitation"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	// Attach session cookie
	token, err := GetToken()
	if err != nil {
		return fmt.Errorf("failed to get token: %v", err)
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send invitation: %s", string(body))
	}

	return nil
}

func VerifyOTP(email, token string) error {
	authHost := GetAuthHost()
	if authHost == "" {
		return fmt.Errorf("cannot verify OTP in offline mode")
	}

	data := map[string]interface{}{
		"email": email,
		"otp":   token,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %v", err)
	}

	url := authHost + "/auth/verify-otp"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("OTP verification failed: %s", string(body))
	}

	return nil
}

func ResendToken(email string) error {
	authHost := GetAuthHost()
	if authHost == "" {
		return fmt.Errorf("cannot resend token in offline mode")
	}

	data := map[string]interface{}{
		"email": email,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %v", err)
	}

	url := authHost + "/auth/token/resend"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to resend token: %s", string(body))
	}

	return nil
}

func ChangePassword(currentPassword, newPassword, confirmPassword string) error {
	authHost := GetAuthHost()
	if authHost == "" {
		return fmt.Errorf("cannot change password in offline mode")
	}

	data := map[string]interface{}{
		"password":         currentPassword,
		"new_password":     newPassword,
		"confirm_password": confirmPassword,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %v", err)
	}

	url := authHost + "/auth/change-password"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	// Attach session cookie
	token, err := GetToken()
	if err != nil {
		return fmt.Errorf("failed to get token: %v", err)
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to change password: %s", string(body))
	}

	return nil
}

func ResetPassword(email string) error {
	authHost := GetAuthHost()
	if authHost == "" {
		return fmt.Errorf("cannot reset password in offline mode")
	}

	url := authHost + "/auth/reset-password"
	jsonBody := fmt.Sprintf("{\"email\": \"%s\"}", email)
	response, err := http.Post(url, "application/json", strings.NewReader(jsonBody))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		return errors.New(string(body))
	}

	return nil
}

// ContactSales sends a sales inquiry to the Clustta sales team.
func ContactSales(name, email, company, teamSize, source, website, message string) error {
	type requestData struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Company  string `json:"company"`
		TeamSize string `json:"team_size"`
		Source   string `json:"source"`
		Website  string `json:"website"`
		Message  string `json:"message"`
	}

	data := requestData{
		Name:     name,
		Email:    email,
		Company:  company,
		TeamSize: teamSize,
		Source:   source,
		Website:  website,
		Message:  message,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %v", err)
	}

	authHost := GetAuthHost()
	if authHost == "" {
		return fmt.Errorf("cannot contact sales in offline mode")
	}

	url := authHost + "/contact-sales"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to submit sales inquiry: %s", string(body))
	}

	return nil
}

func SubmitDiagnostics(email, description, os, arch, clusttaVersion, logContents string) error {
	type requestData struct {
		Email          string `json:"email"`
		Description    string `json:"description"`
		OS             string `json:"os"`
		Arch           string `json:"arch"`
		ClusttaVersion string `json:"clustta_version"`
		LogContents    string `json:"log_contents"`
	}

	data := requestData{
		Email:          email,
		Description:    description,
		OS:             os,
		Arch:           arch,
		ClusttaVersion: clusttaVersion,
		LogContents:    logContents,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %v", err)
	}

	authHost := GetAuthHost()
	if authHost == "" {
		return fmt.Errorf("cannot submit diagnostics in offline mode")
	}

	url := authHost + "/auth/submit-diagnostics"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to submit diagnostics: %s", string(body))
	}

	return nil
}
