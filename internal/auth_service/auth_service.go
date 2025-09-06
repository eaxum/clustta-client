package auth_service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"clustta/internal/constants"
	"clustta/internal/error_service"
	"clustta/internal/repository/models"

	"github.com/zalando/go-keyring"
)

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

func GetActiveUser() (User, error) {
	token, err := GetToken()
	if err != nil {
		return User{}, err
	}
	return token.User, nil
}

func IsAuthenticated() (bool, error) {
	type responseMessage struct {
		Message string `json: "message" `
	}
	url := constants.HOST + "/auth/authenticated"

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

	client := &http.Client{}
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

func FetchUserPhoto(userId string) ([]byte, error) {
	url := constants.HOST + "/person/" + userId + "/photo"

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
	url := constants.HOST + "/person/" + email

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

func GetToken() (Token, error) {
	// Try to get the active account from multi-account structure first
	activeToken, err := GetActiveAccount()
	if err == nil {
		return activeToken, nil
	}

	// Fallback to old single token format for backwards compatibility
	service := "clustta"
	key := "token"

	token, err := keyring.Get(service, key)
	if err != nil {
		return Token{}, err
	}
	var tokenStruct Token
	err = json.Unmarshal([]byte(token), &tokenStruct)
	if err != nil {
		return Token{}, err
	}
	return tokenStruct, nil
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
	service := "clustta"
	key := "token"
	err := keyring.Delete(service, key)
	if err != nil {
		return err
	}
	return nil
}

func Login(username string, password string) (Token, error) {
	url := constants.HOST + "/auth/login"
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
	err = SetToken(token)
	if err != nil {
		return Token{}, err
	}
	return token, nil
}

func Register(firstName, lastName, username, email, password, confirmPassword string) (User, error) {
	url := constants.HOST + "/auth/register"
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
	url := constants.HOST + "/person/update"
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
	url := constants.HOST + "/auth/logout"

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
	url := constants.HOST + "/auth/username-exists/" + username
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
	url := constants.HOST + "/auth/email-exists/" + email
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
	url := constants.HOST + "/person/photo"
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
	url := constants.HOST + "/person/deactivate-account"
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

	url := constants.HOST + "/auth/send-invitation"
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
