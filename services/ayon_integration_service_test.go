package services

import (
	"testing"

	"clustta/internal/integrations"
)

func TestAYONWorkfileExternalIDNormalizesApplication(t *testing.T) {
	actual := ayonWorkfileExternalID(" task-id ", " Blender ")
	if actual != "task-id::blender" {
		t.Fatalf("expected normalized workfile ID, got %q", actual)
	}
}

func TestAYONTemplateIDPrefersTaskTypeID(t *testing.T) {
	options := integrations.SyncOptions{
		AssetTypeTemplates: map[string]string{
			"modeling-id": "id-template",
			"Modeling":    "name-template",
		},
	}

	actual := ayonTemplateID(options, "modeling-id", "Modeling")
	if actual != "id-template" {
		t.Fatalf("expected ID template, got %q", actual)
	}
}

func TestValidateAYONWorkfileContextRequiresTaskType(t *testing.T) {
	context := AYONWorkfileContext{
		ProjectID:   "project",
		FolderID:    "folder",
		TaskID:      "task",
		TaskName:    "modeling",
		Application: "blender",
	}

	if err := validateAYONWorkfileContext(context); err == nil {
		t.Fatal("expected missing task type error")
	}
}
