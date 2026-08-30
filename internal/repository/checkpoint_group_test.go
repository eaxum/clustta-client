package repository

import (
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func openCheckpointGroupTestDB(t *testing.T) (*sqlx.DB, *sqlx.Tx) {
	t.Helper()
	db, err := sqlx.Open("sqlite3", filepath.Join(t.TempDir(), "project.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ProjectSchema); err != nil {
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

func insertCheckpointGroupMember(t *testing.T, tx *sqlx.Tx, id, assetId, groupId string, createdAt int) {
	t.Helper()
	_, err := tx.Exec(`
		INSERT INTO asset_checkpoint (
			id, created_at, mtime, asset_id, xxhash_checksum, time_modified,
			file_size, chunks, author_id, group_id
		) VALUES (?, ?, 1, ?, ?, 1, 1, '', 'author', ?)
	`, id, createdAt, assetId, id, groupId)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFinalizeCheckpointGroupClassifiesSingleAndMulti(t *testing.T) {
	_, tx := openCheckpointGroupTestDB(t)

	if _, err := BeginCheckpointGroup(tx, "single-group", CheckpointGroupTypeSingle); err != nil {
		t.Fatal(err)
	}
	insertCheckpointGroupMember(t, tx, "cp-1", "asset-1", "single-group", 1)
	singleGroup, err := FinalizeCheckpointGroup(tx, "single-group")
	if err != nil {
		t.Fatal(err)
	}
	if !singleGroup.Finalized || singleGroup.GroupType != CheckpointGroupTypeSingle {
		t.Fatalf("unexpected single group: %+v", singleGroup)
	}
	if _, err = BeginCheckpointGroup(tx, "multi-group", CheckpointGroupTypeMulti); err != nil {
		t.Fatal(err)
	}
	insertCheckpointGroupMember(t, tx, "cp-2", "asset-1", "multi-group", 2)
	if _, err = FinalizeCheckpointGroup(tx, "multi-group"); err == nil {
		t.Fatal("expected incomplete multi group finalization to fail")
	}
	insertCheckpointGroupMember(t, tx, "cp-3", "asset-2", "multi-group", 3)
	multiGroup, err := FinalizeCheckpointGroup(tx, "multi-group")
	if err != nil {
		t.Fatal(err)
	}
	if !multiGroup.Finalized || multiGroup.GroupType != CheckpointGroupTypeMulti {
		t.Fatalf("unexpected multi group: %+v", multiGroup)
	}
}

func TestGetTimelinePreservesEveryGroupId(t *testing.T) {
	_, tx := openCheckpointGroupTestDB(t)
	if _, err := BeginCheckpointGroup(tx, "group-a", CheckpointGroupTypeMulti); err != nil {
		t.Fatal(err)
	}
	if _, err := BeginCheckpointGroup(tx, "group-b", CheckpointGroupTypeSingle); err != nil {
		t.Fatal(err)
	}
	insertCheckpointGroupMember(t, tx, "cp-a1", "asset-1", "group-a", 3)
	insertCheckpointGroupMember(t, tx, "cp-a2", "asset-2", "group-a", 2)
	insertCheckpointGroupMember(t, tx, "cp-b1", "asset-3", "group-b", 1)

	timeline, err := GetTimeline(tx)
	if err != nil {
		t.Fatal(err)
	}
	if len(timeline) != 2 {
		t.Fatalf("expected two timeline groups, got %d", len(timeline))
	}
	if timeline[0].GroupId != "group-a" || timeline[1].GroupId != "group-b" {
		t.Fatalf("unexpected timeline group IDs: %+v", timeline)
	}
}
