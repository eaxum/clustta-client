package services

import (
	"bytes"
	"clustta/internal/auth_service"
	"clustta/internal/constants"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type ProfileService struct{}

// NullString represents a nullable string from the database
type NullString struct {
	Valid  bool   `json:"Valid"`
	String string `json:"String"`
}

// Profile response types
type UserProfile struct {
	// Basic info
	Id        string      `json:"id"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Username  string      `json:"username"`
	Email     string      `json:"email"`
	Photo     interface{} `json:"photo"` // Can be []byte or string

	// Extended profile
	Bio               string      `json:"bio"`
	Location          string      `json:"location"`
	CountryID         *NullString `json:"country_id,omitempty"`
	GenderID          *NullString `json:"gender_id,omitempty"`
	DateOfBirth       string      `json:"date_of_birth"`
	Availability      string      `json:"availability"`
	ProfileVisibility string      `json:"profile_visibility"`

	// Professional info
	JobTitle        string  `json:"job_title"`
	Company         string  `json:"company"`
	YearsExperience int     `json:"years_experience"`
	HourlyRate      float64 `json:"hourly_rate"`

	// Individual link fields from backend
	ArtstationLink string `json:"artstation_link"`
	BehanceLink    string `json:"behance_link"`
	InstagramLink  string `json:"instagram_link"`
	XLink          string `json:"x_link"`
	LinkedInLink   string `json:"linkedin_link"`
	GithubLink     string `json:"github_link"`
	PortfolioLink  string `json:"portfolio_link"`
	ReelURL        string `json:"reel_url"`

	// Collections
	Tools   []UserTool   `json:"tools"`
	Skills  []UserSkill  `json:"skills"`
	Studios []UserStudio `json:"studios"`
}

type UserTool struct {
	ID               string `json:"id"`
	ToolID           string `json:"tool_id"`
	ToolName         string `json:"tool_name"`     // Backend returns tool_name
	ToolCategory     string `json:"tool_category"` // Backend returns tool_category
	ProficiencyLevel string `json:"proficiency_level"`
}

type UserSkill struct {
	ID               string `json:"id"`
	SkillID          string `json:"skill_id"`
	SkillName        string `json:"skill_name"`     // Backend returns skill_name
	SkillCategory    string `json:"skill_category"` // Backend returns skill_category
	ProficiencyLevel string `json:"proficiency_level"`
}

type UserStudio struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
	Logo string `json:"logo"`
}

type Tool struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Logo     string `json:"logo"`
}

type Skill struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Icon     string `json:"icon"`
}

type Country struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type Gender struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProfileUpdateData struct {
	FirstName         string  `json:"first_name,omitempty"`
	LastName          string  `json:"last_name,omitempty"`
	Username          string  `json:"username,omitempty"`
	Email             string  `json:"email,omitempty"`
	Bio               string  `json:"bio,omitempty"`
	Location          string  `json:"location,omitempty"`
	CountryID         *string `json:"country_id,omitempty"`
	GenderID          *string `json:"gender_id,omitempty"`
	ArtstationLink    string  `json:"artstation_link,omitempty"`
	BehanceLink       string  `json:"behance_link,omitempty"`
	InstagramLink     string  `json:"instagram_link,omitempty"`
	XLink             string  `json:"x_link,omitempty"`
	LinkedInLink      string  `json:"linkedin_link,omitempty"`
	GithubLink        string  `json:"github_link,omitempty"`
	PortfolioLink     string  `json:"portfolio_link,omitempty"`
	ReelURL           string  `json:"reel_url,omitempty"`
	JobTitle          string  `json:"job_title,omitempty"`
	Company           string  `json:"company,omitempty"`
	YearsExperience   int     `json:"years_experience,omitempty"`
	HourlyRate        float64 `json:"hourly_rate,omitempty"`
	Availability      string  `json:"availability,omitempty"`
	ProfileVisibility string  `json:"profile_visibility,omitempty"`
}

type ToolData struct {
	ToolID           string `json:"tool_id"`
	ProficiencyLevel string `json:"proficiency_level"`
}

type SkillData struct {
	SkillID          string `json:"skill_id"`
	ProficiencyLevel string `json:"proficiency_level"`
}

// getProfileHost returns the appropriate API host for profile operations.
// Returns the current auth host for all modes except offline.
func (p *ProfileService) getProfileHost() (string, error) {
	if auth_service.IsOfflineMode() {
		return "", fmt.Errorf("profile operations are not available in offline mode")
	}
	return auth_service.GetAuthHost(), nil
}

// isStudioAuth reports whether the active account authenticates against a
// private studio server, which only supports basic profile fields.
func (p *ProfileService) isStudioAuth() bool {
	return auth_service.GetActiveAuthMode() == auth_service.AuthModeStudio
}

// makeRequest executes authenticated HTTP requests to the profile API.
// Handles request construction, authentication headers, and response validation.
func (p *ProfileService) makeRequest(method, url string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	auth_service.AttachBearerToken(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status %d: %s", response.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

// GetUserProfile fetches the complete user profile including bio, location, and professional info.
// In studio mode only basic fields are available, built from the studio's current-user endpoint.
func (p *ProfileService) GetUserProfile(userId string) (UserProfile, error) {
	host, err := p.getProfileHost()
	if err != nil {
		return UserProfile{}, err
	}

	if p.isStudioAuth() {
		return p.getStudioUserProfile(host)
	}

	url := host + "/api/users/" + userId + "/profile"

	responseBody, err := p.makeRequest("GET", url, nil)
	if err != nil {
		return UserProfile{}, err
	}

	var profile UserProfile
	err = json.Unmarshal(responseBody, &profile)
	if err != nil {
		return UserProfile{}, fmt.Errorf("error unmarshaling profile: %w", err)
	}

	return profile, nil
}

// getStudioUserProfile builds a basic profile from the studio server's current-user endpoint.
func (p *ProfileService) getStudioUserProfile(host string) (UserProfile, error) {
	responseBody, err := p.makeRequest("GET", host+"/auth/user", nil)
	if err != nil {
		return UserProfile{}, err
	}

	var userInfo struct {
		Id        string `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
		Email     string `json:"email"`
	}
	if err := json.Unmarshal(responseBody, &userInfo); err != nil {
		return UserProfile{}, fmt.Errorf("error unmarshaling user: %w", err)
	}

	return UserProfile{
		Id:        userInfo.Id,
		FirstName: userInfo.FirstName,
		LastName:  userInfo.LastName,
		Username:  userInfo.Username,
		Email:     userInfo.Email,
		Tools:     []UserTool{},
		Skills:    []UserSkill{},
		Studios:   []UserStudio{},
	}, nil
}

// UpdateUserProfile updates user profile fields with the provided data.
// In studio mode only the basic fields (name, username, email) are sent.
func (p *ProfileService) UpdateUserProfile(userId string, updateData ProfileUpdateData) error {
	host, err := p.getProfileHost()
	if err != nil {
		return err
	}

	if p.isStudioAuth() {
		body := map[string]string{
			"first_name": updateData.FirstName,
			"last_name":  updateData.LastName,
			"username":   updateData.Username,
			"email":      updateData.Email,
		}
		_, err = p.makeRequest("PUT", host+"/auth/user/profile", body)
		return err
	}

	url := host + "/api/users/" + userId + "/profile"

	_, err = p.makeRequest("PUT", url, updateData)
	if err != nil {
		return err
	}

	return nil
}

// GetUserTools fetches all tools associated with the user's profile.
func (p *ProfileService) GetUserTools(userId string) ([]UserTool, error) {
	host, err := p.getProfileHost()
	if err != nil {
		return nil, err
	}
	url := host + "/api/users/" + userId + "/tools"

	responseBody, err := p.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var tools []UserTool
	err = json.Unmarshal(responseBody, &tools)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling tools: %w", err)
	}

	return tools, nil
}

// AddUserTool adds a new tool with proficiency level to the user's profile.
func (p *ProfileService) AddUserTool(userId string, toolData ToolData) error {
	host, err := p.getProfileHost()
	if err != nil {
		return err
	}
	url := host + "/api/users/" + userId + "/tools"

	_, err = p.makeRequest("POST", url, toolData)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUserTool updates the proficiency level for an existing tool.
func (p *ProfileService) UpdateUserTool(userId, toolId, proficiencyLevel string) error {
	host, err := p.getProfileHost()
	if err != nil {
		return err
	}
	url := host + "/api/users/" + userId + "/tools/" + toolId

	body := map[string]string{
		"proficiency_level": proficiencyLevel,
	}

	_, err = p.makeRequest("PUT", url, body)
	if err != nil {
		return err
	}

	return nil
}

// RemoveUserTool removes a tool from the user's profile.
func (p *ProfileService) RemoveUserTool(userId, toolId string) error {
	host, err := p.getProfileHost()
	if err != nil {
		return err
	}
	url := host + "/api/users/" + userId + "/tools/" + toolId

	_, err = p.makeRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetUserSkills fetches all skills associated with the user's profile.
func (p *ProfileService) GetUserSkills(userId string) ([]UserSkill, error) {
	host, err := p.getProfileHost()
	if err != nil {
		return nil, err
	}
	url := host + "/api/users/" + userId + "/skills"

	responseBody, err := p.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var skills []UserSkill
	err = json.Unmarshal(responseBody, &skills)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling skills: %w", err)
	}

	return skills, nil
}

// AddUserSkill adds a new skill with proficiency level to the user's profile.
func (p *ProfileService) AddUserSkill(userId string, skillData SkillData) error {
	host, err := p.getProfileHost()
	if err != nil {
		return err
	}
	url := host + "/api/users/" + userId + "/skills"

	_, err = p.makeRequest("POST", url, skillData)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUserSkill updates the proficiency level for an existing skill.
func (p *ProfileService) UpdateUserSkill(userId, skillId, proficiencyLevel string) error {
	host, err := p.getProfileHost()
	if err != nil {
		return err
	}
	url := host + "/api/users/" + userId + "/skills/" + skillId

	body := map[string]string{
		"proficiency_level": proficiencyLevel,
	}

	_, err = p.makeRequest("PUT", url, body)
	if err != nil {
		return err
	}

	return nil
}

// RemoveUserSkill removes a skill from the user's profile.
func (p *ProfileService) RemoveUserSkill(userId, skillId string) error {
	host, err := p.getProfileHost()
	if err != nil {
		return err
	}
	url := host + "/api/users/" + userId + "/skills/" + skillId

	_, err = p.makeRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetAllTools fetches all available tools from the system.
// Returns an empty list in studio mode, which has no tools catalog.
func (p *ProfileService) GetAllTools() ([]Tool, error) {
	host, err := p.getProfileHost()
	if err != nil {
		return nil, err
	}
	if p.isStudioAuth() {
		return []Tool{}, nil
	}
	url := host + "/api/tools"

	responseBody, err := p.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var tools []Tool
	err = json.Unmarshal(responseBody, &tools)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling tools: %w", err)
	}

	return tools, nil
}

// GetToolsByCategory fetches tools filtered by the specified category.
func (p *ProfileService) GetToolsByCategory(category string) ([]Tool, error) {
	host, err := p.getProfileHost()
	if err != nil {
		return nil, err
	}
	url := host + "/api/tools/category/" + category

	responseBody, err := p.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var tools []Tool
	err = json.Unmarshal(responseBody, &tools)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling tools: %w", err)
	}

	return tools, nil
}

// GetAllSkills fetches all available skills from the system.
// Returns an empty list in studio mode, which has no skills catalog.
func (p *ProfileService) GetAllSkills() ([]Skill, error) {
	host, err := p.getProfileHost()
	if err != nil {
		return nil, err
	}
	if p.isStudioAuth() {
		return []Skill{}, nil
	}
	url := host + "/api/skills"

	responseBody, err := p.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var skills []Skill
	err = json.Unmarshal(responseBody, &skills)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling skills: %w", err)
	}

	return skills, nil
}

// GetSkillsByCategory fetches skills filtered by the specified category.
func (p *ProfileService) GetSkillsByCategory(category string) ([]Skill, error) {
	host, err := p.getProfileHost()
	if err != nil {
		return nil, err
	}
	url := host + "/api/skills/category/" + category

	responseBody, err := p.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var skills []Skill
	err = json.Unmarshal(responseBody, &skills)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling skills: %w", err)
	}

	return skills, nil
}

// GetAllCountries fetches all available countries for profile location selection.
func (p *ProfileService) GetAllCountries() ([]Country, error) {
	host, err := p.getProfileHost()
	if err != nil {
		return nil, err
	}
	url := host + "/api/countries"

	responseBody, err := p.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var countries []Country
	err = json.Unmarshal(responseBody, &countries)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling countries: %w", err)
	}

	return countries, nil
}

// GetAllGenders fetches all available gender options for profile selection.
func (p *ProfileService) GetAllGenders() ([]Gender, error) {
	host, err := p.getProfileHost()
	if err != nil {
		return nil, err
	}
	url := host + "/api/genders"

	responseBody, err := p.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var genders []Gender
	err = json.Unmarshal(responseBody, &genders)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling genders: %w", err)
	}

	return genders, nil
}

// UpdateUserPhoto uploads a new profile photo for the user.
func (p *ProfileService) UpdateUserPhoto(photoPath string) error {
	host, err := p.getProfileHost()
	if err != nil {
		return err
	}
	url := host + "/person/photo"
	if p.isStudioAuth() {
		url = host + "/auth/user/photo"
	}

	// Read the file
	fileData, err := os.ReadFile(photoPath)
	if err != nil {
		return fmt.Errorf("error reading photo file: %w", err)
	}

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file field
	part, err := writer.CreateFormFile("photo", filepath.Base(photoPath))
	if err != nil {
		return fmt.Errorf("error creating form file: %w", err)
	}

	// Write file data
	_, err = io.Copy(part, bytes.NewReader(fileData))
	if err != nil {
		return fmt.Errorf("error writing file data: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("error closing multipart writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	token, err := auth_service.GetToken()
	if err != nil {
		return fmt.Errorf("error getting auth token: %w", err)
	}
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", token.SessionId))
	auth_service.AttachBearerToken(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	// Send request
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("request failed with status %d: %s", response.StatusCode, string(responseBody))
	}

	return nil
}
