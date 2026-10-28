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

func TestAssetLookupPathUsesProjectRelativePath(t *testing.T) {
	workingDirectory := filepath.Join("C:", "Project")
	filePath := filepath.Join(workingDirectory, "Assets", "Hero.blend")
	assetPath, extension, err := assetLookupPath(workingDirectory, filePath)
	if err != nil {
		t.Fatal(err)
	}
	if assetPath != "/Assets/Hero" || extension != ".blend" {
		t.Fatalf("unexpected lookup path %q and extension %q", assetPath, extension)
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

func TestContextResponseAllowsUntrackedProjectFile(t *testing.T) {
	response := contextResponse{Studio: "Studio", Project: dccProjectResponse{ID: "project-id"}}
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"asset":null`) {
		t.Fatalf("untracked context must contain a null asset: %s", encoded)
	}
}

func TestDependencyResponseIncludesReferencedAsset(t *testing.T) {
	response := dependencyResponse{
		AssetDependencyEdge: models.AssetDependencyEdge{
			AssetDependency: models.AssetDependency{DependencyId: "dependency-id"},
		},
		Dependency: assetResponse{ID: "dependency-id", FilePath: "C:/Project/Rig.blend"},
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	encodedText := string(encoded)
	if !strings.Contains(encodedText, `"dependency":{"id":"dependency-id"`) ||
		!strings.Contains(encodedText, `"file_path":"C:/Project/Rig.blend"`) {
		t.Fatalf("dependency asset is missing: %s", encodedText)
	}
}

func TestDependencyRequestsDoNotExposeDependencyType(t *testing.T) {
	requests := []any{
		dependencyRequest{
			DependencyID:   "dependency-id",
			ResolutionMode: repository.DependencyResolutionFloating,
		},
		importDependenciesRequest{
			References: []dependencyReference{{FilePath: "C:/Project/Rig.blend"}},
		},
	}
	for _, request := range requests {
		encoded, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(encoded), "dependency_type_id") {
			t.Fatalf("dependency type is exposed in request: %s", encoded)
		}
	}
}

func TestCapabilitiesIncludeAssetTracking(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/v1/capabilities", nil)
	V1Capabilities(recorder, request)

	var response capabilitiesResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	required := map[string]bool{
		"assets.track":               false,
		"assets.sync":                false,
		"assets.get_latest":          false,
		"assets.dependencies.import": false,
	}
	for _, operation := range response.Operations {
		if _, ok := required[operation]; ok {
			required[operation] = true
		}
	}
	for operation, found := range required {
		if !found {
			t.Fatalf("%s capability is missing", operation)
		}
	}
}

func TestTrackAssetContractIncludesCreatedResources(t *testing.T) {
	requestBody := trackAssetRequest{}
	if err := json.Unmarshal([]byte(`{
		"filePath":"C:/project/hero.blend",
		"sourceAssetId":"source-asset",
		"sourceCheckpointId":"source-checkpoint",
		"assignToCurrentUser":true,
		"sync":true
	}`), &requestBody); err != nil {
		t.Fatal(err)
	}
	if requestBody.FilePath != "C:/project/hero.blend" ||
		requestBody.SourceAssetID != "source-asset" ||
		requestBody.SourceCheckpointID != "source-checkpoint" ||
		!requestBody.AssignToCurrentUser ||
		requestBody.Sync == nil || !*requestBody.Sync {
		t.Fatalf("unexpected tracking request: %#v", requestBody)
	}

	response := trackAssetResponse{
		Asset:      assetResponse{ID: "asset-id"},
		Checkpoint: checkpointResponse{ID: "checkpoint-id"},
		Created:    true,
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	encodedText := string(encoded)
	if !strings.Contains(encodedText, `"asset":{"id":"asset-id"`) ||
		!strings.Contains(encodedText, `"checkpoint":{"source_checkpoint_id":null,"id":"checkpoint-id"`) ||
		!strings.Contains(encodedText, `"created":true`) {
		t.Fatalf("unexpected tracking response: %s", encodedText)
	}
}

func TestAYONPrepareContractIncludesFolderHierarchy(t *testing.T) {
	var request ayonWorkfileRequest
	if err := json.Unmarshal([]byte(`{
		"ayon_project_id":"ayon-project",
		"folder_id":"folder-id",
		"folder_path":"/assets/hero/",
		"workfile_path":"C:/project/assets/hero/Hero_Modeling.blend",
		"folders":[{"id":"folder-id","name":"hero","path":"/assets/hero/","type":"asset"}],
		"task_id":"task-id",
		"task_name":"Modeling",
		"task_type":"Modeling",
		"application":"blender",
		"extension":".blend"
	}`), &request); err != nil {
		t.Fatal(err)
	}
	context := ayonContextFromRequest(request)
	if context.FolderPath != "/assets/hero/" ||
		context.WorkfilePath != "C:/project/assets/hero/Hero_Modeling.blend" ||
		len(context.Folders) != 1 {
		t.Fatalf("unexpected AYON context: %#v", context)
	}
}
