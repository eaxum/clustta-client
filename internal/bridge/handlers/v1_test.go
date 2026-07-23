package handlers

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"clustta/internal/repository"
)

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
