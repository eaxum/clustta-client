package sync_service

import (
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func openChangeLogTestDB(t *testing.T) (*sqlx.DB, *sqlx.Tx) {
	t.Helper()
	db, err := sqlx.Open("sqlite3", filepath.Join(t.TempDir(), "project.clst"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(repository.ProjectSchema); err != nil {
		db.Close()
		t.Fatal(err)
	}
	tx, err := db.Beginx()
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		tx.Rollback()
		db.Close()
	})
	return db, tx
}

func insertChangeLogAsset(t *testing.T, tx *sqlx.Tx, id, name string) {
	t.Helper()
	_, err := tx.Exec(`
		INSERT INTO asset (
			id, created_at, mtime, name, extension, status_id, asset_type_id, synced
		) VALUES (?, 1, 1, ?, 'blend', 'status', 'type', 1)
	`, id, name)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckpointTagAndTaggedDependencyAppearInChangeSummary(t *testing.T) {
	_, tx := openChangeLogTestDB(t)
	insertChangeLogAsset(t, tx, "shot", "Shot")
	insertChangeLogAsset(t, tx, "boy", "Boy")
	tx.MustExec("INSERT INTO dependency_type (id, mtime, name, synced) VALUES ('linked', 1, 'linked', 1)")
	tx.MustExec("INSERT INTO tag (id, mtime, name, synced) VALUES ('approved', 1, 'approved', 1)")
	tx.MustExec(`INSERT INTO asset_checkpoint (
		id, created_at, mtime, asset_id, xxhash_checksum, time_modified,
		file_size, chunks, comment, author_id, group_id, synced
	) VALUES ('boy-cp', 1, 1, 'boy', 'hash', 1, 1, '', 'release', 'author', 'group', 1)`)
	tx.MustExec("INSERT INTO asset_tag (id, mtime, asset_id, tag_id, synced) VALUES ('boy-tag', 1, 'boy', 'approved', 0)")
	tx.MustExec("INSERT INTO asset_checkpoint_tag (id, mtime, asset_id, tag_id, checkpoint_id, synced) VALUES ('boy-approved', 1, 'boy', 'approved', 'boy-cp', 0)")
	tx.MustExec(`INSERT INTO asset_dependency (
		id, mtime, asset_id, dependency_id, dependency_type_id,
		resolution_mode, asset_checkpoint_tag_id, synced
	) VALUES ('shot-boy', 1, 'shot', 'boy', 'linked', 'tagged', 'boy-approved', 0)`)

	summary, err := LoadChangeSummary(tx)
	if err != nil {
		t.Fatal(err)
	}
	foundCheckpointTag := false
	foundDependency := false
	foundDuplicateAssetTag := false
	for _, asset := range summary.Assets {
		for _, child := range asset.Children {
			switch child.Source {
			case "asset_checkpoint_tag":
				foundCheckpointTag = child.Description == "approved on release"
			case "asset_dependency":
				foundDependency = child.Description == "Boy - approved" && child.ChangeType == "modified"
			case "asset_tag":
				foundDuplicateAssetTag = true
			}
		}
	}
	if !foundCheckpointTag || !foundDependency || foundDuplicateAssetTag {
		t.Fatalf("unexpected change summary: %+v", summary)
	}

	tx.MustExec("UPDATE asset_dependency SET synced = 1 WHERE id = 'shot-boy'")
	tx.MustExec("UPDATE asset_tag SET synced = 1 WHERE id = 'boy-tag'")
	unsynced, err := IsUnsynced(tx)
	if err != nil {
		t.Fatal(err)
	}
	if !unsynced {
		t.Fatal("expected checkpoint tag assignment to make the project unsynced")
	}
}

func TestDiscardAssetDependencyChangeRestoresRemoteSelector(t *testing.T) {
	_, tx := openChangeLogTestDB(t)
	insertChangeLogAsset(t, tx, "shot", "Shot")
	insertChangeLogAsset(t, tx, "boy", "Boy")
	tx.MustExec("INSERT INTO dependency_type (id, mtime, name, synced) VALUES ('linked', 1, 'linked', 1)")
	tx.MustExec(`INSERT INTO asset_checkpoint (
		id, created_at, mtime, asset_id, xxhash_checksum, time_modified,
		file_size, chunks, author_id, group_id, synced
	) VALUES ('boy-cp', 1, 1, 'boy', 'hash', 1, 1, '', 'author', 'group', 1)`)
	tx.MustExec(`INSERT INTO asset_dependency (
		id, mtime, asset_id, dependency_id, dependency_type_id,
		resolution_mode, checkpoint_id, synced
	) VALUES ('shot-boy', 2, 'shot', 'boy', 'linked', 'pinned', 'boy-cp', 0)`)

	serverData := ProjectData{AssetDependencies: []models.AssetDependency{{
		Id: "shot-boy", MTime: 1, AssetId: "shot", DependencyId: "boy",
		DependencyTypeId: "linked", Synced: true,
	}}}
	if err := DiscardAssetDependencyChange(tx, serverData, "shot-boy"); err != nil {
		t.Fatal(err)
	}
	dependency := models.AssetDependency{}
	if err := tx.Get(&dependency, "SELECT * FROM asset_dependency WHERE id = 'shot-boy'"); err != nil {
		t.Fatal(err)
	}
	if dependency.ResolutionMode != "floating" || dependency.CheckpointId != nil || !dependency.Synced {
		t.Fatalf("expected the remote floating selector, got %+v", dependency)
	}
}

func TestDiscardTagChangeRemovesLocalTagAndTombstone(t *testing.T) {
	_, tx := openChangeLogTestDB(t)
	tx.MustExec("INSERT INTO tag (id, mtime, name, synced) VALUES ('local-tag', 1, 'Local tag', 0)")

	summary, err := LoadChangeSummary(tx)
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.Other) != 1 || summary.Other[0].Source != "tag" {
		t.Fatalf("expected one project tag change, got %+v", summary.Other)
	}
	if err = DiscardTagChange(tx, ProjectData{}, "local-tag"); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = tx.Get(&count, "SELECT COUNT(*) FROM tag WHERE id = 'local-tag'"); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("expected the local tag to be removed")
	}
	if err = tx.Get(&count, "SELECT COUNT(*) FROM tomb WHERE id = 'local-tag' AND table_name = 'tag'"); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("expected the local tag tombstone to be cancelled")
	}
}

func TestDiscardAssetChangesClearsRelationTombstones(t *testing.T) {
	_, tx := openChangeLogTestDB(t)
	insertChangeLogAsset(t, tx, "shot", "Shot")
	insertChangeLogAsset(t, tx, "boy", "Boy")
	tx.MustExec("INSERT INTO dependency_type (id, mtime, name, synced) VALUES ('linked', 1, 'linked', 1)")
	tx.MustExec("INSERT INTO tag (id, mtime, name, synced) VALUES ('approved', 1, 'approved', 1)")
	tx.MustExec(`INSERT INTO asset_checkpoint (
		id, created_at, mtime, asset_id, xxhash_checksum, time_modified,
		file_size, chunks, author_id, group_id, synced
	) VALUES ('boy-cp', 1, 1, 'boy', 'hash', 1, 1, '', 'author', 'group', 1)`)
	tx.MustExec("INSERT INTO asset_tag (id, mtime, asset_id, tag_id, synced) VALUES ('boy-tag', 1, 'boy', 'approved', 1)")
	tx.MustExec("INSERT INTO asset_checkpoint_tag (id, mtime, asset_id, tag_id, checkpoint_id, synced) VALUES ('boy-approved', 1, 'boy', 'approved', 'boy-cp', 1)")
	tx.MustExec(`INSERT INTO asset_dependency (
		id, mtime, asset_id, dependency_id, dependency_type_id,
		resolution_mode, asset_checkpoint_tag_id, synced
	) VALUES ('shot-boy', 1, 'shot', 'boy', 'linked', 'tagged', 'boy-approved', 1)`)

	if err := DiscardAssetChanges(tx, ProjectData{}, "boy"); err != nil {
		t.Fatal(err)
	}
	var relationTombCount int
	if err := tx.Get(&relationTombCount, `
		SELECT COUNT(*) FROM tomb WHERE table_name IN (
			'asset_dependency', 'asset_checkpoint_tag', 'asset_tag', 'asset_checkpoint'
		)
	`); err != nil {
		t.Fatal(err)
	}
	if relationTombCount != 0 {
		t.Fatalf("expected discarded relation tombstones to be cleared, got %d", relationTombCount)
	}
}
