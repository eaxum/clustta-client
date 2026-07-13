package services_test

import (
	"clustta/internal/repository"
	repositorysync "clustta/internal/repository/sync_service"
	"clustta/services"
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestPatchMetadataRemoteUsesTypedPatchEndpoint(t *testing.T) {
	type assetPatch struct {
		Id       string `json:"id"`
		StatusId string `json:"status_id"`
	}
	payload := map[string]any{"assets": []assetPatch{{Id: "asset-1", StatusId: "review"}}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/assets" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body struct {
			Assets []assetPatch `json:"assets"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if len(body.Assets) != 1 || body.Assets[0].Id != "asset-1" || body.Assets[0].StatusId != "review" {
			t.Errorf("unexpected payload: %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"assets":[{"id":"asset-1","status_id":"review"}],"sync_token":"token"}`))
	}))
	defer server.Close()
	var response struct {
		SyncToken string `json:"sync_token"`
	}
	if err := services.PatchMetadataRemote(server.URL, "/assets", payload, &response); err != nil {
		t.Fatal(err)
	}
	if response.SyncToken != "token" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestPatchMetadataRemoteReturnsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()
	var result any
	err := services.PatchMetadataRemote(server.URL, "/assets", map[string]any{"assets": []any{}}, &result)
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("expected typed HTTP failure, got %v", err)
	}
	if services.IsMetadataTransportFailure(err) {
		t.Fatal("HTTP responses must not be eligible for local fallback")
	}
}

func TestPatchMetadataRemoteClassifiesTransportFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	remoteURL := server.URL
	server.Close()

	var result any
	err := services.PatchMetadataRemote(remoteURL, "/assets", map[string]any{"assets": []any{}}, &result)
	if err == nil || !services.IsMetadataTransportFailure(err) {
		t.Fatalf("expected fallback-eligible transport failure, got %v", err)
	}
}

func TestSyncedTombDoesNotDirtyProject(t *testing.T) {
	db, err := sqlx.Open("sqlite3", filepath.Join(t.TempDir(), "project.clst"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(repository.ProjectSchema); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO tomb(id, mtime, table_name, synced) VALUES('relation-1', 1, 'collection_assignee', 1)"); err != nil {
		t.Fatal(err)
	}
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	dirty, err := repositorysync.IsUnsynced(tx)
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Fatal("synced tomb incorrectly reported as a pending change")
	}
}
