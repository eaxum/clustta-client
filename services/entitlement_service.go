package services

import (
	"bytes"
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
	Plan              string            `json:"plan"`
	PlanType          string            `json:"plan_type"`
	Status            string            `json:"status"`
	Limits            EntitlementLimits `json:"limits"`
	Usage             EntitlementUsage  `json:"usage"`
	Features          []string          `json:"features"`
	EffectiveFeatures []string          `json:"effective_features,omitempty"`
}

// privateServerEntitlements returns a default entitlement bundle for private studio servers.
// Private servers are self-hosted and do not enforce cloud-based plan limits.
func privateServerEntitlements() EntitlementBundle {
	return EntitlementBundle{
		Plan:     "studio",
		PlanType: "studio",
		Status:   "active",
		Limits: EntitlementLimits{
			StorageBytes:      -1,
			MaxRemoteProjects: -1,
			MaxCollaborators:  -1,
			AICreditsMonthly:  0,
		},
		Usage:             EntitlementUsage{},
		Features:          []string{"sync", "collaboration"},
		EffectiveFeatures: []string{"sync", "collaboration"},
	}
}

// GetEntitlements fetches the current user's entitlement bundle from the server.
// Returns default unlimited entitlements for private studio servers.
func (e *EntitlementService) GetEntitlements() (EntitlementBundle, error) {
	if auth_service.GetActiveAuthMode() != auth_service.AuthModeGlobal {
		return privateServerEntitlements(), nil
	}
	url := constants.HOST + "/entitlements"
	return fetchEntitlements(url)
}

// GetStudioEntitlements fetches entitlements for a specific studio.
// Returns default unlimited entitlements for private studio servers.
func (e *EntitlementService) GetStudioEntitlements(studioId string) (EntitlementBundle, error) {
	if auth_service.GetActiveAuthMode() != auth_service.AuthModeGlobal {
		return privateServerEntitlements(), nil
	}
	url := constants.HOST + "/entitlements?studio_id=" + studioId
	return fetchEntitlements(url)
}

func fetchEntitlements(url string) (EntitlementBundle, error) {
	bundle := EntitlementBundle{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return bundle, fmt.Errorf("error creating request: %w", err)
	}

	auth_service.AttachBearerToken(req)
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

// Plan represents a subscription plan from the server.
type Plan struct {
	Id                string   `json:"id"`
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	StorageBytes      int64    `json:"storage_bytes"`
	MaxRemoteProjects int      `json:"max_remote_projects"`
	MaxCollaborators  int      `json:"max_collaborators"`
	AICreditsMonthly  int      `json:"ai_credits_monthly"`
	HasSync           bool     `json:"has_sync"`
	HasAI             bool     `json:"has_ai"`
	HasCustomRoles    bool     `json:"has_custom_roles"`
	HasIntegrations   bool     `json:"has_integrations"`
	PriceCents        int      `json:"price_cents"`
	DisplayOrder      int      `json:"display_order"`
	IsActive          bool     `json:"is_active"`
	FeatureKeys       []string `json:"feature_keys"`
}

// GetPlans fetches all available plans from the server.
func (e *EntitlementService) GetPlans() ([]Plan, error) {
	url := constants.HOST + "/plans"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching plans: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	var plans []Plan
	err = json.NewDecoder(resp.Body).Decode(&plans)
	if err != nil {
		return nil, fmt.Errorf("error decoding plans: %w", err)
	}

	return plans, nil
}

// ChangePlan changes the subscription to a new plan and returns the updated entitlements.
func (e *EntitlementService) ChangePlan(planId string) (EntitlementBundle, error) {
	url := constants.HOST + "/subscription/change-plan"

	body := map[string]string{"plan_id": planId}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return EntitlementBundle{}, fmt.Errorf("error marshalling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return EntitlementBundle{}, fmt.Errorf("error creating request: %w", err)
	}

	auth_service.AttachBearerToken(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return EntitlementBundle{}, fmt.Errorf("error changing plan: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return EntitlementBundle{}, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var bundle EntitlementBundle
	err = json.NewDecoder(resp.Body).Decode(&bundle)
	if err != nil {
		return EntitlementBundle{}, fmt.Errorf("error decoding response: %w", err)
	}

	return bundle, nil
}

// CreateCheckout creates a Stripe Checkout Session and returns the checkout URL.
func (e *EntitlementService) CreateCheckout(planId, studioId string) (string, error) {
	url := constants.HOST + "/subscription/create-checkout"

	body := map[string]string{"plan_id": planId}
	if studioId != "" {
		body["studio_id"] = studioId
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("error marshalling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	auth_service.AttachBearerToken(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error creating checkout: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		CheckoutURL string `json:"checkout_url"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	return result.CheckoutURL, nil
}

// OpenBillingPortal creates a Stripe Billing Portal session and returns the portal URL.
func (e *EntitlementService) OpenBillingPortal() (string, error) {
	url := constants.HOST + "/subscription/portal"

	body := map[string]string{}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("error marshalling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	auth_service.AttachBearerToken(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error opening billing portal: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		PortalURL string `json:"portal_url"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	return result.PortalURL, nil
}
