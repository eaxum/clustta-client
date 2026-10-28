package projectindex

import (
	"path/filepath"
	"testing"
)

func TestFindEntryPrefersMostSpecificWorkingDirectory(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "Project")
	nested := filepath.Join(parent, "Nested")
	filePath := filepath.Join(nested, "asset.blend")
	entries := map[string]Entry{
		"parent": {ProjectID: "parent", WorkingDirectory: parent},
		"nested": {ProjectID: "nested", WorkingDirectory: nested},
	}

	entry, found := findEntry(entries, filePath)
	if !found {
		t.Fatal("expected a matching project")
	}
	if entry.ProjectID != "nested" {
		t.Fatalf("expected nested project, got %q", entry.ProjectID)
	}
}

func TestFindEntryRejectsSiblingDirectory(t *testing.T) {
	root := t.TempDir()
	entries := map[string]Entry{
		"project": {
			ProjectID:        "project",
			WorkingDirectory: filepath.Join(root, "Project"),
		},
	}

	_, found := findEntry(entries, filepath.Join(root, "Project Copy", "asset.blend"))
	if found {
		t.Fatal("expected sibling directory not to match")
	}
}

func TestWriteIndexFileReplacesExistingSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "account.json")
	first := indexFile{
		Version: indexVersion,
		Entries: map[string]Entry{"old": {ProjectID: "old"}},
	}
	second := indexFile{
		Version: indexVersion,
		Entries: map[string]Entry{"new": {ProjectID: "new"}},
	}
	if err := writeIndexFile(path, first); err != nil {
		t.Fatalf("write first snapshot: %v", err)
	}
	if err := writeIndexFile(path, second); err != nil {
		t.Fatalf("replace snapshot: %v", err)
	}

	stored, err := readIndexFile(path)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if _, exists := stored.Entries["old"]; exists {
		t.Fatal("old snapshot entry was preserved")
	}
	if _, exists := stored.Entries["new"]; !exists {
		t.Fatal("new snapshot entry is missing")
	}
}

func TestReadIndexFileIgnoresUnsupportedVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "account.json")
	unsupported := indexFile{
		Version: indexVersion + 1,
		Entries: map[string]Entry{"project": {ProjectID: "project"}},
	}
	if err := writeIndexFile(path, unsupported); err != nil {
		t.Fatal(err)
	}

	stored, err := readIndexFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.Entries) != 0 {
		t.Fatal("unsupported index entries should not be loaded")
	}
}

func TestWriteIndexFilePreservesBackfillState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "account.json")
	stored := indexFile{
		Version:          indexVersion,
		BackfillComplete: true,
		Entries:          map[string]Entry{},
	}
	if err := writeIndexFile(path, stored); err != nil {
		t.Fatal(err)
	}

	loaded, err := readIndexFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.BackfillComplete {
		t.Fatal("backfill state was not preserved")
	}
}

func TestPrepareEntryCanonicalizesPaths(t *testing.T) {
	root := t.TempDir()
	entry, err := prepareEntry(Entry{
		ProjectID:        "project",
		DatabasePath:     filepath.Join(root, "metadata", "..", "project.clst"),
		WorkingDirectory: filepath.Join(root, "work", "."),
	})
	if err != nil {
		t.Fatal(err)
	}
	if entry.DatabasePath != filepath.Join(root, "project.clst") {
		t.Fatalf("unexpected database path: %q", entry.DatabasePath)
	}
	if entry.WorkingDirectory != filepath.Join(root, "work") {
		t.Fatalf("unexpected working directory: %q", entry.WorkingDirectory)
	}
	if entry.UpdatedAt == "" {
		t.Fatal("expected an update timestamp")
	}
}
