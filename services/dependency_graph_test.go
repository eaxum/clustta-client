package services

import (
	"path/filepath"
	"slices"
	"testing"

	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"github.com/jmoiron/sqlx"
)

func TestRecursiveDependenciesDirectAndFullGraph(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "graph.db")
	db, err := sqlx.Open("sqlite3", projectPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.MustExec(repository.ProjectSchema)
	db.MustExec(`
        INSERT INTO status (id, mtime, name, short_name) VALUES ('status', 1, 'Todo', 'todo');
        INSERT INTO asset_type (id, mtime, name, icon) VALUES ('type', 1, 'Asset', 'asset');
        INSERT INTO collection_type (id, mtime, name, icon) VALUES ('collection-type', 1, 'Folder', 'folder');
        INSERT INTO dependency_type (id, mtime, name) VALUES ('reference', 1, 'Reference');
        INSERT INTO collection (id, created_at, mtime, name, collection_type_id, parent_id)
        VALUES ('group', 1, 1, 'Group', 'collection-type', ''), ('nested', 1, 1, 'Nested', 'collection-type', 'group');
    `)
	for _, id := range []string{"root", "a", "b", "c", "d", "e", "member", "external", "trashed", "hidden"} {
		db.MustExec(`INSERT INTO asset (id, created_at, mtime, name, extension, status_id, asset_type_id)
            VALUES (?, 1, 1, ?, '.blend', 'status', 'type')`, id, id)
	}
	db.MustExec("UPDATE collection SET description = ''")
	db.MustExec("UPDATE asset SET collection_id = 'nested' WHERE id = 'member'")
	db.MustExec("UPDATE asset SET trashed = 1 WHERE id = 'trashed'")
	for _, edge := range [][2]string{{"root", "a"}, {"a", "b"}, {"b", "c"}, {"c", "d"}, {"d", "e"}, {"e", "root"}, {"member", "external"}, {"root", "trashed"}, {"trashed", "hidden"}} {
		db.MustExec(`INSERT INTO asset_dependency (id, mtime, asset_id, dependency_id, dependency_type_id)
            VALUES (?, 1, ?, ?, 'reference')`, edge[0]+"-"+edge[1], edge[0], edge[1])
	}
	db.MustExec(`INSERT INTO collection_dependency (id, mtime, asset_id, dependency_id, dependency_type_id)
        VALUES ('root-group', 1, 'root', 'group', 'reference'), ('a-group', 1, 'a', 'group', 'reference')`)

	service := AssetService{}
	for _, scenario := range []struct {
		name     string
		depth    int
		expected []string
	}{
		{name: "direct", depth: 1, expected: []string{"a", "group"}},
		{name: "full", depth: 0, expected: []string{"a", "b", "c", "d", "e", "group", "nested", "member", "external"}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			items, err := service.GetRecursiveDependencies(projectPath, "root", scenario.depth)
			if err != nil {
				t.Fatal(err)
			}
			byID := make(map[string]map[string]interface{})
			for _, value := range items {
				item := value.(map[string]interface{})
				var id string
				if asset, ok := item["asset"].(models.Asset); ok {
					id = asset.Id
				} else {
					id = item["collection"].(models.Collection).Id
				}
				if _, exists := byID[id]; exists {
					t.Fatalf("duplicate dependency %s", id)
				}
				byID[id] = item
			}
			if len(byID) != len(scenario.expected) {
				t.Fatalf("expected %v, got %v", scenario.expected, byID)
			}
			for _, id := range scenario.expected {
				if _, exists := byID[id]; !exists {
					t.Fatalf("missing dependency %s", id)
				}
			}
			if scenario.depth == 1 {
				for _, item := range byID {
					if item["depth"] != 1 || item["parentId"] != "root" {
						t.Fatalf("unexpected direct relationship: %+v", item)
					}
				}
			} else {
				if byID["e"]["depth"] != 5 {
					t.Fatalf("expected dependency beyond four levels: %+v", byID["e"])
				}
				parents := byID["group"]["parentIds"].([]string)
				if !slices.Contains(parents, "root") || !slices.Contains(parents, "a") {
					t.Fatalf("missing shared collection relationships: %v", parents)
				}
			}
		})
	}
}
