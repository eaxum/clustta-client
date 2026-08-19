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
