package integrations

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAssetsPreservesKitsuMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`[{"id":"chair-id","name":"Chair","entity_type_id":"props-id","asset_type_name":"Props","data":{"Category":"Furniture"}}]`))
	}))
	defer server.Close()

	client := NewKitsuClient()
	assets, err := client.getAssets("token", server.URL, "project-id", nil)
	if err != nil {
		t.Fatalf("get assets: %v", err)
	}
	if len(assets) != 1 {
		t.Fatalf("expected one asset, got %d", len(assets))
	}
	if assets[0].Metadata["Category"] != "Furniture" {
		t.Fatalf("expected Category metadata, got %#v", assets[0].Metadata)
	}
}

func TestGetEditsPreservesHierarchyAndMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`[{
			"id":"edit-id",
			"name":"Director Cut",
			"episode_id":"episode-id",
			"data":{"Delivery":"Festival"}
		}]`))
	}))
	defer server.Close()

	client := NewKitsuClient()
	edits, err := client.getEdits("token", server.URL, "project-id")
	if err != nil {
		t.Fatalf("get edits: %v", err)
	}
	if len(edits) != 1 {
		t.Fatalf("expected one edit, got %d", len(edits))
	}
	edit := edits[0]
	if edit.ParentID != "episode-id" || edit.Type != "edit" || !edit.HasAssets {
		t.Fatalf("unexpected edit hierarchy: %#v", edit)
	}
	if edit.Metadata["Delivery"] != "Festival" {
		t.Fatalf("expected Delivery metadata, got %#v", edit.Metadata)
	}
}

func TestGetCollectionTypesIncludesEdit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := NewKitsuClient()
	types, err := client.GetCollectionTypes("token", server.URL, "project-id")
	if err != nil {
		t.Fatalf("get collection types: %v", err)
	}
	for _, collectionType := range types {
		if collectionType.ID == "edit" && collectionType.Name == "Edit" {
			return
		}
	}
	t.Fatal("expected Edit collection type")
}
