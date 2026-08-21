package services

import (
	"testing"

	"clustta/internal/repository/models"
)

func TestBuildExternalEntityURL(t *testing.T) {
	project := models.IntegrationProject{
		IntegrationId:     "kitsu",
		ExternalProjectId: "project-id",
		ApiUrl:            "https://kitsu.example.com/studio/?source=clustta#fragment",
	}
	mapping := models.IntegrationAssetMapping{
		ExternalId:       "task-id",
		ExternalParentId: "entity-id",
	}

	tests := []struct {
		name       string
		parentType string
		expected   string
	}{
		{
			name:       "shot entity",
			parentType: "Shot",
			expected:   "https://kitsu.example.com/studio/productions/project-id/shots/entity-id",
		},
		{
			name:       "asset entity",
			parentType: "Character",
			expected:   "https://kitsu.example.com/studio/productions/project-id/assets/entity-id",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := buildExternalEntityURL(project, mapping, test.parentType)
			if actual != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestBuildExternalEntityURLRejectsUnsupportedURLs(t *testing.T) {
	mapping := models.IntegrationAssetMapping{ExternalId: "task-id", ExternalParentId: "entity-id"}
	tests := []models.IntegrationProject{
		{IntegrationId: "clickup", ExternalProjectId: "project-id", ApiUrl: "https://example.com"},
		{IntegrationId: "kitsu", ExternalProjectId: "project-id", ApiUrl: "file:///tmp/kitsu"},
		{IntegrationId: "kitsu", ExternalProjectId: "project-id", ApiUrl: "not-a-url"},
	}

	for _, project := range tests {
		if actual := buildExternalEntityURL(project, mapping, "shot"); actual != "" {
			t.Fatalf("expected an empty URL, got %q", actual)
		}
	}
}

func TestSelectAssetMappingForType(t *testing.T) {
	mappings := []models.IntegrationAssetMapping{
		{ExternalId: "modeling-task", ExternalType: "Modeling"},
		{ExternalId: "rigging-task", ExternalType: "Rigging"},
	}

	mapping, err := selectAssetMappingForType(mappings, " rigging ")
	if err != nil {
		t.Fatalf("expected matching mapping, got error: %v", err)
	}
	if mapping.ExternalId != "rigging-task" {
		t.Fatalf("expected rigging mapping, got %q", mapping.ExternalId)
	}

	if _, err := selectAssetMappingForType(mappings, "lighting"); err == nil {
		t.Fatal("expected no mapping for an unmatched asset type")
	}
}

func TestSelectAssetMappingForTypeKeepsSingleMapping(t *testing.T) {
	mappings := []models.IntegrationAssetMapping{
		{ExternalId: "animation-task", ExternalType: "Animation"},
	}

	mapping, err := selectAssetMappingForType(mappings, "lighting")
	if err != nil {
		t.Fatalf("expected the single mapping, got error: %v", err)
	}
	if mapping.ExternalId != "animation-task" {
		t.Fatalf("expected the single mapping, got %q", mapping.ExternalId)
	}
}
