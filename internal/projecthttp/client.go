// Package projecthttp provides compatibility-aware HTTP transport for project APIs.
package projecthttp

import (
	"clustta/internal/compatibility"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
)

var verifiedHosts sync.Map
var replicaSchemas sync.Map

type cachedContract struct {
	contract *compatibility.Contract
	err      error
}

// Client applies the compatibility contract to Clustta project requests.
type Client struct {
	client *http.Client
}

func New(client *http.Client) *Client {
	return &Client{client: client}
}

// OnRejection is installed once at application startup before requests begin.
var OnRejection func(string, *compatibility.Rejection)

func Report(projectURL string, err error) error {
	if rejection, ok := err.(*compatibility.Rejection); ok && OnRejection != nil {
		OnRejection(projectURL, rejection)
	}
	return err
}

// Do validates the remote contract before a project data request, then validates its response.
func (c *Client) Do(request *http.Request) (*http.Response, error) {
	client := c.client
	projectURL, discovery := requestProject(request)
	if projectURL == "" {
		response, err := client.Do(request)
		if err == nil && response.StatusCode == http.StatusOK && strings.HasSuffix(request.URL.Path, "/projects") {
			Remember(request, request.URL.String(), contractFromHeaders(response.Header))
		}
		return response, err
	}
	declare(request.Header, ReplicaSchema(projectURL))
	if discovery && request.Method == http.MethodPost {
		if err := verifyCreation(client, request); err != nil {
			return nil, Report(projectURL, err)
		}
	}
	if !discovery {
		if err := ValidateRemote(client, request, projectURL); err != nil {
			return nil, Report(projectURL, err)
		}
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode == http.StatusUpgradeRequired {
		defer response.Body.Close()
		var rejection compatibility.Rejection
		if err := json.NewDecoder(response.Body).Decode(&rejection); err != nil {
			return nil, err
		}
		verifiedHosts.Store(verificationKey(request, projectURL), cachedContract{err: &rejection})
		return nil, Report(projectURL, &rejection)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return response, nil
	}
	if discovery {
		return response, nil
	}
	contract := contractFromHeaders(response.Header)
	if err := compatibility.Check(contract); err != nil {
		response.Body.Close()
		Remember(request, projectURL, contract)
		return nil, Report(projectURL, err)
	}
	return response, nil
}

// Remember reuses contracts returned by ordinary project discovery and listings.
func Remember(request *http.Request, projectURL string, contract *compatibility.Contract) {
	verifiedHosts.Store(verificationKey(request, projectURL), cachedContract{contract: contract})
}

// ValidateRemote discovers only projects absent from the current account's cache.
func ValidateRemote(client *http.Client, request *http.Request, projectURL string) error {
	if entry, ok := verifiedHosts.Load(verificationKey(request, projectURL)); ok {
		cached := entry.(cachedContract)
		if cached.err != nil {
			return cached.err
		}
		return compatibility.Check(cached.contract)
	}
	_, err := DiscoverRemote(client, request, projectURL)
	return err
}

// DiscoverRemote refreshes and validates the authoritative project contract.
func DiscoverRemote(client *http.Client, request *http.Request, projectURL string) (*compatibility.Contract, error) {
	probe, err := http.NewRequestWithContext(request.Context(), http.MethodGet, projectURL, nil)
	if err != nil {
		return nil, err
	}
	probe.Header = request.Header.Clone()
	response, err := client.Do(probe)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("project discovery failed: HTTP %d", response.StatusCode)
	}
	var info struct {
		Compatibility *compatibility.Contract `json:"compatibility"`
	}
	if err := json.NewDecoder(response.Body).Decode(&info); err != nil {
		return nil, err
	}
	Remember(request, projectURL, info.Compatibility)
	if err := compatibility.Check(info.Compatibility); err != nil {
		return info.Compatibility, err
	}
	return info.Compatibility, nil
}

func requestProject(request *http.Request) (string, bool) {
	if request.Header.Get("Clustta-Agent") == "" {
		return "", false
	}
	parts := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
	rootLength := 1
	if parts[0] == "user" || parts[0] == "studio" {
		rootLength = 3
	}
	if len(parts) < rootLength {
		return "", false
	}
	switch parts[rootLength-1] {
	case "projects", "quota", "persons", "studio-info", "ping", "version":
		return "", false
	}
	if len(parts) > rootLength {
		switch parts[rootLength] {
		case "data", "sync-token", "chunks", "stream-chunks", "chunks-info", "chunks-missing", "chunk-urls", "chunk-upload-urls", "chunk-upload-confirm", "previews", "preview", "previews-exist", "icon", "ignore-list", "toggle-close", "status", "assets", "collections", "asset-types", "collection-types", "collaborators", "leave", "storage-conversion":
		default:
			return "", false
		}
	}
	projectURL := request.URL.Scheme + "://" + request.URL.Host + "/" + strings.Join(parts[:rootLength], "/")
	discovery := len(parts) == rootLength && (request.Method == http.MethodGet || request.Method == http.MethodPost)
	return projectURL, discovery
}

func verifyCreation(client *http.Client, request *http.Request) error {
	target := *request.URL
	parts := strings.Split(strings.Trim(target.Path, "/"), "/")
	switch parts[0] {
	case "user":
		target.Path = "/user/projects"
	case "studio":
		target.Path = "/studio/" + parts[1] + "/projects"
	default:
		target.Path = "/projects"
	}
	if entry, ok := verifiedHosts.Load(verificationKey(request, target.String())); ok {
		cached := entry.(cachedContract)
		return compatibility.Check(cached.contract)
	}
	probe, err := http.NewRequestWithContext(request.Context(), http.MethodGet, target.String(), nil)
	if err != nil {
		return err
	}
	probe.Header = request.Header.Clone()
	response, err := client.Do(probe)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("host discovery failed: HTTP %d", response.StatusCode)
	}
	contract := contractFromHeaders(response.Header)
	Remember(request, target.String(), contract)
	return compatibility.Check(contract)
}

func contractFromHeaders(headers http.Header) *compatibility.Contract {
	return &compatibility.Contract{
		Protocol:      headers.Get(compatibility.ProtocolHeader),
		Schema:        headers.Get(compatibility.SchemaHeader),
		ProjectSchema: headers.Get(compatibility.ProjectSchemaHeader),
	}
}

func declare(headers http.Header, projectSchema string) {
	headers.Set(compatibility.ProtocolHeader, compatibility.Protocol)
	headers.Set(compatibility.SchemaHeader, compatibility.Schema)
	if projectSchema == "" {
		projectSchema = compatibility.Schema
	}
	headers.Set(compatibility.ProjectSchemaHeader, projectSchema)
}

func verificationKey(request *http.Request, projectURL string) string {
	identity := request.Header.Get("Authorization")
	if identity == "" {
		identity = request.Header.Get("UserId")
	}
	return projectURL + "|" + identity
}

func ValidateReplica(database sqlx.Queryer, remoteURL string) error {
	schema, err := compatibility.ReadSchema(database)
	if err != nil {
		return err
	}
	RememberReplica(remoteURL, schema)
	if schema != compatibility.Schema {
		return Report(remoteURL, compatibility.Reject(schema, "replica"))
	}
	return nil
}

func RememberReplica(projectURL, schema string) {
	if projectURL != "" && schema != "" {
		replicaSchemas.Store(strings.TrimSuffix(projectURL, "/"), schema)
	}
}

func ReplicaSchema(projectURL string) string {
	if schema, ok := replicaSchemas.Load(strings.TrimSuffix(projectURL, "/")); ok {
		return schema.(string)
	}
	return compatibility.Schema
}
