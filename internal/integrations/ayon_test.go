package integrations

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAYONAuthenticateUsesAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "secret" {
			t.Fatal("missing AYON API key header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"clustta-service","attrib":{"fullName":"Clustta Service","email":"service@example.com"}}`))
	}))
	defer server.Close()

	client := NewAYONClient()
	result, err := client.Authenticate(map[string]string{
		"api_url": server.URL,
		"api_key": "secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success || result.UserID != "clustta-service" || result.AccessToken != "secret" {
		t.Fatalf("unexpected auth result: %+v", result)
	}
}

func TestAYONUpdateAssetStatusUpdatesMappedTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/projects/demo/tasks/task-id" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("X-Api-Key") != "secret" {
			t.Fatal("missing AYON API key header")
		}
		payload := map[string]string{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload["status"] != "In progress" {
			t.Fatalf("unexpected status payload: %+v", payload)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewAYONClient()
	err := client.UpdateAssetStatus(
		"secret",
		server.URL,
		"demo",
		"task-id::blender",
		"In progress",
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAYONUploadPreviewCreatesReviewableAndUpdatesStatus(t *testing.T) {
	previewPath := filepath.Join(t.TempDir(), "preview.png")
	if err := os.WriteFile(previewPath, []byte("preview-data"), 0o600); err != nil {
		t.Fatal(err)
	}

	requests := make([]string, 0, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "secret" {
			t.Fatal("missing AYON API key header")
		}
		requests = append(requests, r.Method+" "+r.URL.Path)
		switch r.URL.Path {
		case "/api/projects/demo/files":
			if r.Header.Get("X-File-Name") != "preview.png" {
				t.Fatalf("unexpected file name: %s", r.Header.Get("X-File-Name"))
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(body) != "preview-data" {
				t.Fatalf("unexpected upload body: %q", string(body))
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"file-id"}`))
		case "/api/projects/demo/tasks/task-id/activities":
			payload := struct {
				ActivityType string   `json:"activityType"`
				Body         string   `json:"body"`
				Files        []string `json:"files"`
			}{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if payload.ActivityType != "reviewable" || payload.Body != "Ready for review" {
				t.Fatalf("unexpected activity: %+v", payload)
			}
			if len(payload.Files) != 1 || payload.Files[0] != "file-id" {
				t.Fatalf("unexpected activity files: %+v", payload.Files)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"activity-id"}`))
		case "/api/projects/demo/tasks/task-id":
			status := map[string]string{}
			if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
				t.Fatal(err)
			}
			if status["status"] != "Done" {
				t.Fatalf("unexpected status: %+v", status)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	progressCalled := false
	client := NewAYONClient()
	err := client.UploadPreview(
		"secret",
		server.URL,
		"demo",
		"task-id::blender",
		previewPath,
		"Ready for review",
		"Done",
		func(sent, total int64) {
			progressCalled = sent == total
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !progressCalled {
		t.Fatal("expected upload progress")
	}
	expected := []string{
		"POST /api/projects/demo/files",
		"POST /api/projects/demo/tasks/task-id/activities",
		"PATCH /api/projects/demo/tasks/task-id",
	}
	if strings.Join(requests, "|") != strings.Join(expected, "|") {
		t.Fatalf("unexpected requests: %+v", requests)
	}
}

func TestAYONExternalEntityURLUsesResolvedTaskURI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/projects/demo/uris" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		payload := struct {
			EntityType string   `json:"entityType"`
			IDs        []string `json:"ids"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.EntityType != "task" || len(payload.IDs) != 1 || payload.IDs[0] != "task-id" {
			t.Fatalf("unexpected URI request: %+v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uris":[{"id":"task-id","uri":"ayon://demo/assets/hero?task=modeling"}]}`))
	}))
	defer server.Close()

	client := NewAYONClient()
	externalURL, err := client.GetExternalEntityURL("secret", server.URL, "demo", "task-id::blender")
	if err != nil {
		t.Fatal(err)
	}
	expected := server.URL + "/projects/demo/browser?uri=ayon%3A%2F%2Fdemo%2Fassets%2Fhero%3Ftask%3Dmodeling"
	if externalURL != expected {
		t.Fatalf("expected %q, got %q", expected, externalURL)
	}
}

func TestAYONProjectCollectionsFlattenHierarchyForSync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"hierarchy":[{"id":"characters","name":"characters","folderType":"AssetBuild","children":[{"id":"hero","name":"hero","label":"Hero","folderType":"Character","hasTasks":true,"children":[]}]}]}`))
	}))
	defer server.Close()

	client := NewAYONClient()
	collections, err := client.GetProjectCollections("secret", server.URL, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(collections) != 2 {
		t.Fatalf("unexpected hierarchy: %+v", collections)
	}
	characters := collections[0]
	if characters.ID != "characters" || characters.ParentID != "" || characters.Path != "/characters/" {
		t.Fatalf("unexpected root collection: %+v", characters)
	}
	hero := collections[1]
	if hero.ID != "hero" || hero.ParentID != "characters" || hero.Path != "/characters/hero/" || !hero.HasAssets {
		t.Fatalf("unexpected child collection: %+v", hero)
	}
	if len(characters.Children) != 0 || len(hero.Children) != 0 {
		t.Fatalf("expected a flat collection list: %+v", collections)
	}
}

func TestAYONProjectAssetsMapTasksToProviderNeutralAssets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphql" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-Api-Key") != "secret" {
			t.Fatal("missing AYON API key header")
		}
		payload := ayonGraphQLRequest{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.Variables.ProjectName != "demo" {
			t.Fatalf("unexpected GraphQL variables: %+v", payload.Variables)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data":{"project":{"tasks":{
				"edges":[
					{"node":{"id":"model-task","name":"modeling","label":"Model the character","taskType":"Modeling","folderId":"hero","status":"In progress","assignees":["artist"],"active":true}},
					{"node":{"id":"old-task","name":"old","taskType":"Modeling","folderId":"hero","status":"Done","assignees":[],"active":false}}
				],
				"pageInfo":{"endCursor":null,"hasNextPage":false}
			}}}
		}`))
	}))
	defer server.Close()

	client := NewAYONClient()
	assets, err := client.GetProjectAssets("secret", server.URL, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 1 {
		t.Fatalf("unexpected assets: %+v", assets)
	}
	asset := assets[0]
	if asset.ID != "model-task" || asset.ParentID != "hero" || asset.Name != "Modeling" {
		t.Fatalf("unexpected task asset identity: %+v", asset)
	}
	if asset.AssetType != "Modeling" || asset.AssetTypeID != "Modeling" {
		t.Fatalf("unexpected task type mapping: %+v", asset)
	}
	if asset.Status != "In progress" || len(asset.Assignees) != 1 || asset.Assignees[0] != "artist" {
		t.Fatalf("unexpected task metadata: %+v", asset)
	}
}

func TestAYONProjectMetadataProvidesTaskTypesAndStatuses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/projects/demo" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name":"demo",
			"taskTypes":[{"name":"Modeling"},{"name":"Animation"}],
			"statuses":[
				{"name":"In progress","shortName":"WIP","color":"#3498db","scope":["task","workfile"]},
				{"name":"Published","shortName":"PUB","color":"#00ff00","scope":["version"]}
			]
		}`))
	}))
	defer server.Close()

	client := NewAYONClient()
	taskTypes, err := client.GetAssetTypes("secret", server.URL, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(taskTypes) != 2 || taskTypes[0].ID != "Modeling" {
		t.Fatalf("unexpected task types: %+v", taskTypes)
	}

	statuses, err := client.GetTaskStatuses("secret", server.URL, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 1 || statuses[0].ID != "In progress" || statuses[0].ShortName != "WIP" {
		t.Fatalf("unexpected statuses: %+v", statuses)
	}
}
