package projecthttp

import (
	. "clustta/internal/compatibility"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLegacyHostNeverReceivesMutation(t *testing.T) {
	mutations := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			mutations++
			t.Error("sent data to legacy host")
		}
		w.Write([]byte(`{"version":2.1}`))
	}))
	defer server.Close()
	for _, path := range []string{"/project/data", "/project/chunks", "/project/assets", "/project"} {
		request, err := http.NewRequest(http.MethodPost, server.URL+path, strings.NewReader("pending changes"))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Clustta-Agent", "test")
		_, err = New(server.Client()).Do(request)
		var rejection *Rejection
		if !errors.As(err, &rejection) || rejection.RequiredUpdate != "server" {
			t.Fatalf("expected host update: %v", err)
		}
	}
	if mutations != 0 {
		t.Fatalf("received %d writes", mutations)
	}
}

func TestHostDriftRejectsActualOperation(t *testing.T) {
	schema := Schema
	mutations := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/project" {
			json.NewEncoder(w).Encode(map[string]*Contract{"compatibility": Current(schema)})
			return
		}
		if r.Header.Get(SchemaHeader) != schema {
			writeError(w, Reject(schema, "client"))
			return
		}
		respond(w, schema)
		mutations++
	}))
	defer server.Close()
	write := func() error {
		request, err := http.NewRequest(http.MethodPost, server.URL+"/project/data", strings.NewReader("pending changes"))
		if err != nil {
			return err
		}
		request.Header.Set("Clustta-Agent", "test")
		response, err := New(server.Client()).Do(request)
		if response != nil {
			response.Body.Close()
		}
		return err
	}
	if err := write(); err != nil {
		t.Fatal(err)
	}
	schema = "2.3"
	var rejection *Rejection
	if !errors.As(write(), &rejection) {
		t.Fatal("schema drift not rejected")
	}
	if mutations != 1 {
		t.Fatalf("committed %d mutations", mutations)
	}
}

func TestDiscoveryDoesNotProbeEachProject(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Write([]byte(`[]`))
	}))
	defer server.Close()
	request, _ := http.NewRequest(http.MethodGet, server.URL+"/studio/id/projects", nil)
	request.Header.Set("Clustta-Agent", "test")
	response, err := New(server.Client()).Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if requests != 1 {
		t.Fatalf("discovery made %d requests", requests)
	}
}

func TestCachedContractAvoidsDiscovery(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/project/data" {
			t.Errorf("unnecessary discovery: %s", r.URL.Path)
		}
		respond(w, Schema)
	}))
	defer server.Close()
	listing, _ := http.NewRequest(http.MethodGet, server.URL+"/projects", nil)
	listing.Header.Set("Authorization", "Bearer account-one")
	Remember(listing, server.URL+"/project", Current(Schema))
	request, _ := http.NewRequest(http.MethodPost, server.URL+"/project/data", strings.NewReader("changes"))
	request.Header.Set("Clustta-Agent", "test")
	request.Header.Set("Authorization", "Bearer account-one")
	request.Header.Set("UserId", "optional-legacy-id")
	response, err := New(server.Client()).Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if requests != 1 {
		t.Fatalf("made %d requests", requests)
	}
}

func TestProjectClientDeclaresActualReplicaSchema(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(ProjectSchemaHeader); got != LegacySchema {
			t.Errorf("declared project schema %q", got)
		}
		respond(w, Schema)
	}))
	defer server.Close()
	request, _ := http.NewRequest(http.MethodGet, server.URL+"/project/data", nil)
	request.Header.Set("Clustta-Agent", "test")
	Remember(request, server.URL+"/project", Current(Schema))
	RememberReplica(server.URL+"/project", LegacySchema)
	response, err := New(server.Client()).Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
}

func TestCachedRejectionBlocksUntilDiscoveryRefresh(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		writeError(w, Reject("2.3", "client"))
	}))
	defer server.Close()
	request, _ := http.NewRequest(http.MethodPost, server.URL+"/project/data", nil)
	request.Header.Set("Clustta-Agent", "test")
	Remember(request, server.URL+"/project", Current(Schema))
	for range 2 {
		if _, err := New(server.Client()).Do(request); err == nil {
			t.Fatal("expected rejection")
		}
	}
	if requests != 1 {
		t.Fatalf("made %d requests after rejection", requests)
	}
	Remember(request, server.URL+"/project", Current(Schema))
	if err := ValidateRemote(server.Client(), request, server.URL+"/project"); err != nil {
		t.Fatal(err)
	}
}

func TestContractCacheDoesNotCrossAccounts(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()
	request, _ := http.NewRequest(http.MethodGet, server.URL+"/project", nil)
	request.Header.Set("Authorization", "Bearer one")
	Remember(request, server.URL+"/project", Current(Schema))
	request.Header.Set("Authorization", "Bearer two")
	if err := ValidateRemote(server.Client(), request, server.URL+"/project"); err == nil {
		t.Fatal("reused another account's contract")
	}
	if requests != 1 {
		t.Fatalf("made %d requests", requests)
	}
}

func TestCreationReusesHostListing(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		respond(w, Schema)
		if r.URL.Path == "/projects" {
			w.Write([]byte("[]"))
			return
		}
		if r.Method != http.MethodPost || r.URL.Path != "/new-project" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	for _, operation := range []struct{ method, path string }{{http.MethodGet, "/projects"}, {http.MethodPost, "/new-project"}} {
		request, _ := http.NewRequest(operation.method, server.URL+operation.path, nil)
		request.Header.Set("Clustta-Agent", "test")
		response, err := New(server.Client()).Do(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
	}
	if requests != 2 {
		t.Fatalf("listing and creation made %d requests", requests)
	}
}

func respond(w http.ResponseWriter, schema string) {
	w.Header().Set(ProtocolHeader, Protocol)
	w.Header().Set(SchemaHeader, Schema)
	w.Header().Set(ProjectSchemaHeader, schema)
}

func writeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUpgradeRequired)
	json.NewEncoder(w).Encode(err)
}
