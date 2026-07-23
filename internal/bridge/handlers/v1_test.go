package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"clustta/internal/repository"
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
