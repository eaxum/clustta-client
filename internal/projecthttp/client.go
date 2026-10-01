// Package projecthttp applies the negotiated API version to project requests.
package projecthttp

import (
	"clustta/internal/compatibility"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var negotiatedAPIs sync.Map

// Client sends project requests using a versioned API contract.
type Client struct {
	client *http.Client
}

func New(client *http.Client) *Client {
	if client == nil {
		client = http.DefaultClient
	}
	return &Client{client: client}
}

// Do sends a request through the current API unless the caller selected another version.
func (c *Client) Do(request *http.Request) (*http.Response, error) {
	if request == nil {
		return nil, fmt.Errorf("request is required")
	}
	if request.Header.Get(compatibility.APIVersionHeader) == "" {
		request.Header.Set(compatibility.APIVersionHeader, apiVersionForURL(request.URL))
	}
	return c.client.Do(request)
}

// RegisterAPI stores the negotiated API for a Studio origin.
func RegisterAPI(studioURL string, info compatibility.APIInfo) {
	parsed, err := url.Parse(studioURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return
	}
	api := compatibility.SelectHighestMutual(info)
	negotiatedAPIs.Store(parsed.Scheme+"://"+parsed.Host, api.Version)
}

func apiVersionForURL(requestURL *url.URL) string {
	if requestURL != nil {
		origin := requestURL.Scheme + "://" + requestURL.Host
		if version, ok := negotiatedAPIs.Load(strings.TrimSuffix(origin, "/")); ok {
			return version.(string)
		}
	}
	return compatibility.CurrentAPIVersion
}
