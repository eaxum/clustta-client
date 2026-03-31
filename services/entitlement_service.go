package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/constants"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type EntitlementService struct{}

// EntitlementLimits contains the resolved limits for an entity.
type EntitlementLimits struct {
	StorageBytes      int64 `json:"storage_bytes"`
	MaxRemoteProjects int   `json:"max_remote_projects"`
	MaxCollaborators  int   `json:"max_collaborators"`
	AICreditsMonthly  int   `json:"ai_credits_monthly"`
}

// EntitlementUsage contains current resource consumption.
type EntitlementUsage struct {
	StorageBytes  int64 `json:"storage_bytes"`
	ProjectCount  int   `json:"project_count"`
	AICreditsUsed int   `json:"ai_credits_used"`
}

// EntitlementBundle is the complete entitlement state for an entity.
type EntitlementBundle struct {
	Plan     string            `json:"plan"`
	PlanType string            `json:"plan_type"`
	Status   string            `json:"status"`
	Limits   EntitlementLimits `json:"limits"`
	Usage    EntitlementUsage  `json:"usage"`
	Features []string          `json:"features"`
}

// GetEntitlements fetches the current user's entitlement bundle from the server.
func (e *EntitlementService) GetEntitlements() (EntitlementBundle, error) {
	url := constants.HOST + "/entitlements"
	return fetchEntitlements(url)
}

// GetStudioEntitlements fetches entitlements for a specific studio.
func (e *EntitlementService) GetStudioEntitlements(studioId string) (EntitlementBundle, error) {
	url := constants.HOST + "/entitlements?studio_id=" + studioId
	return fetchEntitlements(url)
}

func fetchEntitlements(url string) (EntitlementBundle, error) {
	bundle := EntitlementBundle{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return bundle, fmt.Errorf("error creating request: %w", err)
	}

	token, err := auth_service.GetToken()
	if err != nil {
		return bundle, fmt.Errorf("error getting auth token: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.SessionId)
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return bundle, fmt.Errorf("error fetching entitlements: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return bundle, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	err = json.NewDecoder(resp.Body).Decode(&bundle)
	if err != nil {
		return bundle, fmt.Errorf("error decoding entitlements: %w", err)
	}

	return bundle, nil
}
