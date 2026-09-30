package repository

import (
	"encoding/json"
	"testing"
)

func TestProjectInfoAcceptsLegacyNumericVersion(t *testing.T) {
	for input, expected := range map[string]string{
		`{"version":2}`:      "2.0",
		`{"version":2.2}`:    "2.2",
		`{"version":"2.10"}`: "2.10",
	} {
		var project ProjectInfo
		if err := json.Unmarshal([]byte(input), &project); err != nil {
			t.Fatalf("unmarshal %s: %v", input, err)
		}
		if project.Version != expected {
			t.Fatalf("unmarshal %s: expected %q, got %q", input, expected, project.Version)
		}
	}
}

func TestProjectInfoMarshalsVersionAsString(t *testing.T) {
	data, err := json.Marshal(ProjectInfo{Version: "2.2"})
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(data, &wire); err != nil {
		t.Fatal(err)
	}
	if string(wire["version"]) != `"2.2"` {
		t.Fatalf("expected a string version, got %s", wire["version"])
	}
}

func TestProjectInfoLegacyVersionPreservesOtherFields(t *testing.T) {
	var project ProjectInfo
	if err := json.Unmarshal([]byte(`{"id":"project-id","name":"Legacy","version":2.2}`), &project); err != nil {
		t.Fatal(err)
	}
	if project.Id != "project-id" || project.Name != "Legacy" {
		t.Fatalf("project fields were not decoded: %+v", project)
	}
}

func TestProjectInfoRejectsInvalidVersion(t *testing.T) {
	var project ProjectInfo
	if err := json.Unmarshal([]byte(`{"version":{"major":2}}`), &project); err == nil {
		t.Fatal("expected invalid version to be rejected")
	}
}
