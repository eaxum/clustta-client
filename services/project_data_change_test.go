package services

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProjectDataChangeJSONContract(t *testing.T) {
	change := ProjectDataChange{
		ProjectID:  "project-id",
		ProjectURI: "C:/Projects/example.clst",
		AssetIDs:   []string{"asset-id"},
		Changes:    []string{ProjectChangeCheckpoints},
	}
	payload, err := json.Marshal(change)
	if err != nil {
		t.Fatal(err)
	}
	encoded := string(payload)
	for _, expected := range []string{
		`"project_id":"project-id"`,
		`"project_uri":"C:/Projects/example.clst"`,
		`"asset_ids":["asset-id"]`,
		`"changes":["checkpoints"]`,
	} {
		if !strings.Contains(encoded, expected) {
			t.Fatalf("project data change is missing %s: %s", expected, encoded)
		}
	}
}

func TestEmitProjectDataChangedWithoutApplication(t *testing.T) {
	EmitProjectDataChanged(ProjectDataChange{})
}
