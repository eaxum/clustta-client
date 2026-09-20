package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"clustta/internal/repository"
	"clustta/internal/repository/models"
)

func TestRefreshRequested(t *testing.T) {
	request := httptest.NewRequest("GET", "/v1/bootstrap?refresh=true", nil)
	if !refreshRequested(request) {
		t.Fatal("expected refresh query to be recognized")
	}
}

func TestProjectCacheReturnsSnapshot(t *testing.T) {
	invalidateProjectCache()
	t.Cleanup(invalidateProjectCache)

	projectCacheMu.Lock()
	projectCacheItems["user\x00studio"] = projectCacheEntry{
		expiresAt: time.Now().Add(time.Minute),
		projects:  []repository.ProjectInfo{{Id: "project-id", Name: "Project"}},
	}
	projectCacheMu.Unlock()

	projects, ok := readProjectCache("user\x00studio")
	if !ok || len(projects) != 1 {
		t.Fatal("expected a cached project")
	}
	projects[0].Name = "Changed"
	projects, ok = readProjectCache("user\x00studio")
	if !ok || projects[0].Name != "Project" {
		t.Fatal("expected callers to receive a cache snapshot")
	}
}

func TestPathWithin(t *testing.T) {
	root := filepath.Join("C:", "projects", "example")
	inside := filepath.Join(root, "shots", "scene.blend")
	outside := filepath.Join("C:", "projects", "other", "scene.blend")

	if !pathWithin(inside, root) {
		t.Fatalf("expected %q to be within %q", inside, root)
	}
	if pathWithin(outside, root) {
		t.Fatalf("expected %q to be outside %q", outside, root)
	}
}

func TestSamePathCleansSegments(t *testing.T) {
	first := filepath.Join("project", "assets", "..", "scene.blend")
	second := filepath.Join("project", "scene.blend")
	if !samePath(first, second) {
		t.Fatalf("expected %q and %q to resolve to the same path", first, second)
	}
}

func TestAssetDeepLink(t *testing.T) {
	deepLink, err := url.Parse(assetDeepLink("My Studio", "project-id", "asset-id"))
	if err != nil {
		t.Fatal(err)
	}
	if deepLink.Scheme != "clustta" || deepLink.Host != "open" {
		t.Fatalf("unexpected deep link target: %s", deepLink)
	}
	query := deepLink.Query()
	if query.Get("studio") != "My Studio" ||
		query.Get("project") != "project-id" ||
		query.Get("asset") != "asset-id" {
		t.Fatalf("unexpected deep link query: %s", query.Encode())
	}
}

func TestStudioForRequestPrefersDCCContext(t *testing.T) {
	request := httptest.NewRequest("GET", "/v1/projects/project-id", nil)
	request.Header.Set("X-Clustta-Studio", "Addon Studio")
	studio, err := studioForRequest(request)
	if err != nil {
		t.Fatal(err)
	}
	if studio != "Addon Studio" {
		t.Fatalf("expected addon studio, got %q", studio)
	}
}

func TestDCCProjectResponseDoesNotExposeDatabaseURI(t *testing.T) {
	project := repository.ProjectInfo{
		Id:               "project-id",
		Uri:              "C:/private/project.clst",
		Name:             "Project",
		WorkingDirectory: "C:/project",
		HasRemote:        true,
	}

	payload, err := json.Marshal(projectToDCCResponse(project))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(string(payload)), "uri") {
		t.Fatalf("DCC response exposes a URI field: %s", payload)
	}
	if strings.Contains(string(payload), project.Uri) {
		t.Fatalf("DCC response exposes the database path: %s", payload)
	}
}

func TestCheckpointResponseIncludesSource(t *testing.T) {
	sourceID := "source-checkpoint"
	response := checkpointToResponse(models.Checkpoint{Id: "output", SourceCheckpointId: &sourceID})
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"source_checkpoint_id":"source-checkpoint"`) {
		t.Fatalf("source missing from checkpoint response: %s", encoded)
	}
	var request checkpointRequest
	if err = json.Unmarshal([]byte(`{"sourceAssetId":"source","sourceCheckpointId":"source-checkpoint"}`), &request); err != nil {
		t.Fatal(err)
	}
	if request.SourceAssetID != "source" || request.SourceCheckpointID != sourceID {
		t.Fatal("source selection not decoded")
	}
}
