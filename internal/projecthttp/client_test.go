package projecthttp

import (
	"clustta/internal/compatibility"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoDeclaresCurrentAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(compatibility.APIVersionHeader); got != compatibility.CurrentAPIVersion {
			t.Fatalf("expected API %s, got %s", compatibility.CurrentAPIVersion, got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := New(server.Client()).Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
}

func TestDoPreservesSelectedAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(compatibility.APIVersionHeader); got != compatibility.LegacyAPIVersion {
			t.Fatalf("expected API %s, got %s", compatibility.LegacyAPIVersion, got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(compatibility.APIVersionHeader, compatibility.LegacyAPIVersion)
	response, err := New(server.Client()).Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
}

func TestDoUsesNegotiatedStudioAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(compatibility.APIVersionHeader); got != compatibility.LegacyAPIVersion {
			t.Fatalf("expected negotiated API %s, got %s", compatibility.LegacyAPIVersion, got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	RegisterAPI(server.URL, compatibility.APIInfo{
		DefaultVersion:    compatibility.LegacyAPIVersion,
		SupportedVersions: []string{compatibility.LegacyAPIVersion},
	})
	request, err := http.NewRequest(http.MethodGet, server.URL+"/project", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := New(server.Client()).Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
}
