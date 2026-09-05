package assets

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"github.com/jmoiron/sqlx"
)

func TestResolveLocalFiles(t *testing.T) {
	root := t.TempDir()
	projectPath := filepath.Join(root, "project.clst")
	filePath := filepath.Join(root, "asset.blend")
	if err := os.WriteFile(filePath, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	db, err := sqlx.Open("sqlite3", projectPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(repository.ProjectSchema)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO config (name, value, mtime) VALUES ('project_id', 'project', 1);
		INSERT INTO role (id, mtime, name, view_asset, pull_chunk) VALUES ('role', 1, 'role', 1, 1);
		INSERT INTO user (id, mtime, added_at, first_name, last_name, username, email, role_id)
		VALUES ('user', 1, 1, 'Test', 'User', 'test', 'test@example.invalid', 'role');
		INSERT INTO asset_type (id, mtime, name, icon) VALUES ('type', 1, 'Blend', 'blend');
		INSERT INTO asset (id, mtime, created_at, name, extension, asset_type_id, status_id, assignee_id)
		VALUES ('asset', 1, 1, 'asset', '.blend', 'type', 'status', 'user');`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO config (name, value, mtime) VALUES ('working_dir', ?, 1)", root); err != nil {
		t.Fatal(err)
	}
	resolve := func(userID, projectID string, ids []string) ([]string, error) {
		return ResolveLocalFiles(context.Background(), projectPath, projectID, userID, ids)
	}
	paths, err := resolve("user", "project", []string{"asset", "asset"})
	if err != nil || len(paths) != 1 {
		t.Fatalf("paths=%v err=%v", paths, err)
	}
	for _, input := range []struct {
		user, project string
		ids           []string
	}{
		{"outsider", "project", []string{"asset"}},
		{"user", "wrong", []string{"asset"}},
		{"user", "project", []string{"asset", "missing"}},
	} {
		if paths, err := resolve(input.user, input.project, input.ids); err == nil || paths != nil {
			t.Fatalf("invalid selection returned paths=%v err=%v", paths, err)
		}
	}
	if _, err := db.Exec("UPDATE role SET pull_chunk = 0"); err != nil {
		t.Fatal(err)
	}
	if _, err := resolve("user", "project", []string{"asset"}); err == nil {
		t.Fatal("export allowed without file permission")
	}
	if _, err := db.Exec("UPDATE role SET pull_chunk = 1, view_asset = 0"); err != nil {
		t.Fatal(err)
	}
	if _, err := resolve("user", "project", []string{"asset"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE asset SET assignee_id = 'someone-else'"); err != nil {
		t.Fatal(err)
	}
	if _, err := resolve("user", "project", []string{"asset"}); err == nil {
		t.Fatal("unassigned asset allowed")
	}
}

func TestLocalAssetPath(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "test.blend")
	if err := os.WriteFile(filePath, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.blend")
	if err := os.WriteFile(outside, []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, asset := range []models.Asset{
		{FilePath: outside}, {FilePath: root}, {FilePath: filePath, Pointer: outside},
		{FilePath: filePath, IsLink: true}, {FilePath: filePath, Trashed: true},
		{FilePath: filepath.Join(root, "missing.blend")},
	} {
		if _, err := localAssetPath(root, asset); err == nil {
			t.Fatalf("accepted %+v", asset)
		}
	}
	asset := models.Asset{FilePath: filepath.Join(root, "renamed.blend"), LocalPath: filePath}
	if _, err := localAssetPath(root, asset); err != nil {
		t.Fatal(err)
	}
	t.Run("symlink escape", func(t *testing.T) {
		link := filepath.Join(root, "outside-link.blend")
		if err := os.Symlink(outside, link); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		if _, err := localAssetPath(root, models.Asset{FilePath: link}); err == nil {
			t.Fatal("symlink escape accepted")
		}
	})
}
