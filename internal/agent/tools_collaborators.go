package agent

import (
	"bytes"
	"clustta/internal/auth_service"
	"clustta/internal/constants"
	"clustta/internal/settings"
	"clustta/internal/studio_service"
	"clustta/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// --- Project (server) collaborator HTTP helpers ---
//
// These tools talk to the project's own remote URL (read from the project DB
// via utils.GetRemoteUrl). They mirror the CollaboratorService HTTP plumbing
// in services/collaborator_service.go so internal/agent stays free of the
// services import (which would create a cycle).

// projectCollaborator matches the JSON returned by GET {remoteUrl}/collaborators.
type projectCollaborator struct {
	Id        string `json:"id"`
	UserId    string `json:"user_id"`
	AddedAt   int64  `json:"added_at"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
}

// resolveRemoteUrl opens the project DB and reads its remote URL.
func resolveRemoteUrl(projectPath string) (string, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return "", err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	remoteUrl, err := utils.GetRemoteUrl(tx)
	if err != nil {
		return "", fmt.Errorf("project has no remote URL configured: %w", err)
	}
	if remoteUrl == "" {
		return "", fmt.Errorf("project has no remote URL configured (this project is not synced to a server)")
	}
	if err := validateOutboundURL(remoteUrl); err != nil {
		return "", err
	}
	return remoteUrl, nil
}

// validateOutboundURL rejects URLs that look like SSRF targets.
//   - Scheme must be http or https
//   - Plain http only allowed when host is loopback (self-hosted dev)
//   - Reject link-local 169.254/16 (cloud instance metadata service / IMDS)
//   - Reject multicast/unspecified addresses
// Private RFC1918 ranges are intentionally allowed because Clustta servers
// are commonly self-hosted on a LAN.
func validateOutboundURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https, got %q", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return errors.New("URL has no host")
	}

	if u.Scheme == "http" {
		isLoopback := host == "localhost" || host == "127.0.0.1" || host == "::1"
		if !isLoopback {
			ip := net.ParseIP(host)
			if ip == nil || !ip.IsLoopback() {
				return errors.New("plain http is only permitted for loopback hosts; use https")
			}
		}
	}

	// IP-literal hosts: block link-local + multicast + unspecified up front.
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
			return fmt.Errorf("URL host %s is in a blocked address range", host)
		}
		return nil
	}

	// DNS-resolvable hosts: resolve and screen each address.
	addrs, err := net.LookupIP(host)
	if err != nil {
		// DNS failure surfaces later as a transport error; don't block here.
		return nil
	}
	for _, ip := range addrs {
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
			return fmt.Errorf("URL host %s resolves to a blocked address (%s)", host, ip)
		}
	}
	return nil
}

// agentHTTPClient is the shared client for outbound calls from agent tools.
// Bounded timeouts at every layer to prevent slowloris-style hangs.
// CheckRedirect strips Authorization on cross-origin redirects so a
// compromised remote cannot leak the user's bearer token.
var agentHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 20 * time.Second,
		ExpectContinueTimeout: 2 * time.Second,
		IdleConnTimeout:       90 * time.Second,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("stopped after 5 redirects")
		}
		if len(via) > 0 {
			origURL := via[0].URL
			if req.URL.Scheme != origURL.Scheme || req.URL.Host != origURL.Host {
				req.Header.Del("Authorization")
			}
			if err := validateOutboundURL(req.URL.String()); err != nil {
				return fmt.Errorf("redirect blocked: %w", err)
			}
		}
		return nil
	},
}

// httpDoAuth performs an authenticated request to the Clustta server.
// On non-2xx, returns an error containing the response body.
func httpDoAuth(method, url string, body []byte) ([]byte, int, error) {
	if err := validateOutboundURL(url); err != nil {
		return nil, 0, err
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	auth_service.AttachBearerToken(req)

	resp, err := agentHTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return respBody, resp.StatusCode, nil
}

// --- Project collaborator exec functions ---

// execListProjectCollaborators fetches all collaborators on the active project's remote.
func execListProjectCollaborators(projectPath string) ToolResult {
	remoteUrl, err := resolveRemoteUrl(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	body, status, err := httpDoAuth("GET", remoteUrl+"/collaborators", nil)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	if status != http.StatusOK {
		return ToolResult{Success: false, Error: fmt.Sprintf("server returned %d: %s", status, string(body))}
	}

	var collaborators []projectCollaborator
	if err := json.Unmarshal(body, &collaborators); err != nil {
		return ToolResult{Success: false, Error: "failed to parse server response: " + err.Error()}
	}
	return ToolResult{Success: true, Data: collaborators}
}

// execAddProjectCollaborator invites a user (by email or user_id) to the active project.
// If email is given, the user is looked up via the global server first.
func execAddProjectCollaborator(projectPath string, args map[string]interface{}) ToolResult {
	email := strings.TrimSpace(getStringArg(args, "email", ""))
	userID := strings.TrimSpace(getStringArg(args, "user_id", ""))
	role := strings.TrimSpace(getStringArg(args, "role", ""))
	if email == "" && userID == "" {
		return ToolResult{Success: false, Error: "either email or user_id is required"}
	}

	remoteUrl, err := resolveRemoteUrl(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	if userID == "" {
		user, err := auth_service.FetchUserData(email)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("could not look up user with email '%s': %s", email, err.Error())}
		}
		if user.Id == "" {
			return ToolResult{Success: false, Error: fmt.Sprintf("no Clustta user found with email '%s'", email)}
		}
		userID = user.Id
	}

	payload := struct {
		UserIds []string `json:"user_ids"`
		Role    string   `json:"role"`
	}{UserIds: []string{userID}, Role: role}
	jsonData, _ := json.Marshal(payload)

	body, status, err := httpDoAuth("POST", remoteUrl+"/collaborators", jsonData)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	if status != http.StatusCreated {
		return ToolResult{Success: false, Error: fmt.Sprintf("server returned %d: %s", status, string(body))}
	}

	var results []map[string]string
	if err := json.Unmarshal(body, &results); err != nil {
		return ToolResult{Success: true, Data: map[string]interface{}{"user_id": userID, "role": role, "raw_response": string(body)}}
	}
	return ToolResult{Success: true, Data: map[string]interface{}{"user_id": userID, "role": role, "results": results}}
}

// execRemoveProjectCollaborator removes a user from the active project's remote.
func execRemoveProjectCollaborator(projectPath string, args map[string]interface{}) ToolResult {
	userID := strings.TrimSpace(getStringArg(args, "user_id", ""))
	if userID == "" {
		return ToolResult{Success: false, Error: "user_id is required"}
	}

	remoteUrl, err := resolveRemoteUrl(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	body, status, err := httpDoAuth("DELETE", remoteUrl+"/collaborators/"+userID, nil)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	if status != http.StatusOK {
		return ToolResult{Success: false, Error: fmt.Sprintf("server returned %d: %s", status, string(body))}
	}
	return ToolResult{Success: true, Data: map[string]string{"removed_user_id": userID}}
}

// --- Studio collaborator exec functions ---

// execListStudios returns the locally configured studios (id, name, url, hosting_mode).
// Use this to discover the studio_id required by the studio collaborator tools.
func execListStudios() ToolResult {
	studios, err := settings.GetStudios()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	type studioSummary struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Url         string `json:"url"`
		HostingMode string `json:"hosting_mode,omitempty"`
		Active      bool   `json:"active"`
	}
	results := make([]studioSummary, 0, len(studios))
	for _, s := range studios {
		results = append(results, studioSummary{
			Id:          s.Id,
			Name:        s.Name,
			Url:         s.Url,
			HostingMode: s.HostingMode,
			Active:      bool(s.Active),
		})
	}
	return ToolResult{Success: true, Data: results}
}

// execListStudioUsers lists members of the given studio.
func execListStudioUsers(args map[string]interface{}) ToolResult {
	studioID := strings.TrimSpace(getStringArg(args, "studio_id", ""))
	if studioID == "" {
		return ToolResult{Success: false, Error: "studio_id is required (use list_studios to find IDs)"}
	}
	users, err := studio_service.GetStudioUsers(studioID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	type userSummary struct {
		Id        string `json:"id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		RoleName  string `json:"role_name"`
		Active    bool   `json:"active"`
	}
	results := make([]userSummary, 0, len(users))
	for _, u := range users {
		results = append(results, userSummary{
			Id:        u.Id,
			Username:  u.UserName,
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			RoleName:  u.RoleName,
			Active:    u.Active,
		})
	}
	return ToolResult{Success: true, Data: results}
}

// execAddStudioCollaborator invites a user (by email) to a studio with the given role.
func execAddStudioCollaborator(args map[string]interface{}) ToolResult {
	email := strings.TrimSpace(getStringArg(args, "email", ""))
	studioID := strings.TrimSpace(getStringArg(args, "studio_id", ""))
	role := strings.ToLower(strings.TrimSpace(getStringArg(args, "role_name", "")))
	if email == "" || studioID == "" || role == "" {
		return ToolResult{Success: false, Error: "email, studio_id, and role_name are required"}
	}
	result, err := studio_service.AddCollaborator(email, studioID, role)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: result}
}

// execChangeStudioCollaboratorRole changes an existing studio member's role.
func execChangeStudioCollaboratorRole(args map[string]interface{}) ToolResult {
	userID := strings.TrimSpace(getStringArg(args, "user_id", ""))
	studioID := strings.TrimSpace(getStringArg(args, "studio_id", ""))
	role := strings.ToLower(strings.TrimSpace(getStringArg(args, "role_name", "")))
	if userID == "" || studioID == "" || role == "" {
		return ToolResult{Success: false, Error: "user_id, studio_id, and role_name are required"}
	}
	result, err := studio_service.ChangeCollaboratorRole(userID, studioID, role)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: result}
}

// execRemoveStudioCollaborator removes a user from a studio.
func execRemoveStudioCollaborator(args map[string]interface{}) ToolResult {
	userID := strings.TrimSpace(getStringArg(args, "user_id", ""))
	studioID := strings.TrimSpace(getStringArg(args, "studio_id", ""))
	if userID == "" || studioID == "" {
		return ToolResult{Success: false, Error: "user_id and studio_id are required"}
	}
	result, err := studio_service.RemoveCollaborator(userID, studioID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: result}
}
