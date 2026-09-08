package services

import (
	"path/filepath"
	"strings"
	"testing"

	"clustta/internal/repository/models"

	"github.com/jmoiron/sqlx"
)

func TestDependencyBuildPermissions(t *testing.T) {
	db, err := sqlx.Open("sqlite3", filepath.Join(t.TempDir(), "permissions.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.MustExec(`
		CREATE TABLE asset (id TEXT, name TEXT, extension TEXT, assignee_id TEXT, collection_id TEXT, trashed INTEGER);
		CREATE TABLE collection (id TEXT, parent_id TEXT, is_shared INTEGER, trashed INTEGER);
		CREATE TABLE collection_assignee (collection_id TEXT, assignee_id TEXT);
		CREATE TABLE asset_dependency (asset_id TEXT, dependency_id TEXT);
		CREATE TABLE collection_dependency (asset_id TEXT, dependency_id TEXT);
		INSERT INTO collection VALUES ('shots', '', 0, 0), ('characters', '', 0, 0), ('shared', '', 1, 0);
		INSERT INTO asset VALUES
		('shot', 'Shot', '.blend', 'artist', 'shots', 0),
		('boy', 'Boy', '.blend', 'other', 'characters', 0),
		('shared', 'Shared', '.blend', 'other', 'shared', 0),
		('private', 'Private', '.blend', 'other', 'characters', 0),
		('collection-shot', 'CollectionShot', '.blend', 'other', 'shots', 0);
		INSERT INTO asset_dependency VALUES ('shot', 'boy');
		INSERT INTO collection_assignee VALUES ('shots', 'artist');
	`)
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	user := models.User{Id: "artist"}
	for _, test := range []struct {
		name      string
		root      string
		view      bool
		viewAll   bool
		pull      bool
		missing   bool
		wantError string
	}{
		{name: "assigned shot with unassigned dependency", root: "shot", view: true},
		{name: "assigned collection", root: "collection-shot", view: true},
		{name: "shared asset", root: "shared", view: true},
		{name: "dependency is accessible", root: "boy", view: true},
		{name: "unrelated asset denied", root: "private", view: true, wantError: "Private.blend"},
		{name: "project-wide view", root: "private", view: true, viewAll: true},
		{name: "checkpoint permission required", root: "shot", wantError: "view_checkpoint"},
		{name: "download permission required", root: "shot", view: true, missing: true, wantError: "pull_chunk"},
		{name: "download permitted", root: "shot", view: true, missing: true, pull: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			role := models.Role{ViewCheckpoint: test.view, ViewAsset: test.viewAll, PullChunk: test.pull}
			plan := models.DependencyBuildPlan{
				RootAssetId: test.root,
				Entries:     []models.DependencyBuildPlanEntry{{AssetId: "boy", MissingChunks: test.missing}},
			}
			err := authorizeDependencyBuildTx(tx, user, role, plan)
			if test.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("expected %q, got %v", test.wantError, err)
			}
		})
	}
}
