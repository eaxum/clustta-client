package integrations

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
