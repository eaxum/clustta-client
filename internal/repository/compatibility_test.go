package repository

import (
	"clustta/internal/auth_service"
	"clustta/internal/compatibility"
	"clustta/internal/projecthttp"
	"clustta/internal/repository/migrations"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

func TestCompatibilityMatchesMigrationTarget(t *testing.T) {
	if compatibility.Schema != migrations.LatestVersion {
		t.Fatal("release contract does not match migration target")
	}
}

func TestFrontendCompatibilityMatchesReleaseContract(t *testing.T) {
	contents, err := os.ReadFile(filepath.Join("..", "..", "frontend", "src", "lib", "compatibility.js"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(contents)
	for _, declaration := range []string{"PROJECT_PROTOCOL = '" + compatibility.Protocol + "'", "PROJECT_SCHEMA = '" + compatibility.Schema + "'"} {
		if !strings.Contains(contract, declaration) {
			t.Fatalf("frontend release contract is missing %q", declaration)
		}
	}
}

func TestProjectScanDoesNotUpgradeRemoteReplica(t *testing.T) {
	path := filepath.Join(t.TempDir(), "replica.clst")
	db := sqlx.MustOpen("sqlite3", path)
	defer db.Close()
	db.MustExec(`CREATE TABLE config (name TEXT PRIMARY KEY, value TEXT, mtime INTEGER, synced BOOLEAN);
		INSERT INTO config VALUES ('version', '2.1', 1, 1), ('remote', 'https://studio.test/project', 1, 1), ('sync_token', 'pending', 1, 0);`)
	if err := UpdateProject(path); err != nil {
		t.Fatal(err)
	}
	var version, token string
	if err := db.Get(&version, "SELECT value FROM config WHERE name='version'"); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&token, "SELECT value FROM config WHERE name='sync_token'"); err != nil {
		t.Fatal(err)
	}
	if version != compatibility.LegacySchema || token != "pending" {
		t.Fatalf("replica changed: %s %s", version, token)
	}
	var synced bool
	if err := db.Get(&synced, "SELECT synced FROM config WHERE name='sync_token'"); err != nil {
		t.Fatal(err)
	}
	if synced {
		t.Fatal("pending state cleared")
	}
}

func TestConfirmedHostMigratesUnsyncedReplica(t *testing.T) {
	path := filepath.Join(t.TempDir(), "replica.clst")
	db := sqlx.MustOpen("sqlite3", path)
	defer db.Close()
	db.MustExec(ProjectSchema)
	db.MustExec("INSERT OR REPLACE INTO config (name, value, mtime, synced) VALUES ('version', '2.1', 1, 1), ('sync_token', 'pending', 1, 0)")
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		json.NewEncoder(w).Encode(ProjectInfo{Compatibility: compatibility.Current(compatibility.Schema)})
	}))
	defer server.Close()
	if err := ValidateSyncCompatibility(path, server.URL+"/project"); err != nil {
		t.Fatal(err)
	}
	if requests != 1 {
		t.Fatalf("made %d discovery requests", requests)
	}
	var version, token string
	var synced bool
	if err := db.Get(&version, "SELECT value FROM config WHERE name = 'version'"); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&token, "SELECT value FROM config WHERE name = 'sync_token'"); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&synced, "SELECT synced FROM config WHERE name = 'sync_token'"); err != nil {
		t.Fatal(err)
	}
	if version != compatibility.Schema || token != "pending" || synced {
		t.Fatalf("unexpected migrated state: version=%s token=%s synced=%v", version, token, synced)
	}
}

func TestProjectInfoSeedsRequestCompatibilityCache(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path == "/project" {
			json.NewEncoder(w).Encode(ProjectInfo{Compatibility: compatibility.Current(compatibility.Schema)})
			return
		}
		w.Header().Set(compatibility.ProtocolHeader, compatibility.Protocol)
		w.Header().Set(compatibility.SchemaHeader, compatibility.Schema)
		w.Header().Set(compatibility.ProjectSchemaHeader, compatibility.Schema)
	}))
	defer server.Close()
	if _, err := GetProjectInfo(server.URL+"/project", auth_service.User{Id: "cache-test"}); err != nil {
		t.Fatal(err)
	}
	request, _ := http.NewRequest(http.MethodGet, server.URL+"/project/data", nil)
	request.Header.Set("Clustta-Agent", "test")
	request.Header.Set("UserId", "cache-test")
	auth_service.AttachBearerToken(request)
	response, err := projecthttp.New(server.Client()).Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if requests != 2 {
		t.Fatalf("info and data made %d requests", requests)
	}
}
