package repository

import (
	"clustta/internal/error_service"
	"errors"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func openTagTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Open("sqlite3", filepath.Join(t.TempDir(), "project.clst"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ProjectSchema); err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func createTestTag(t *testing.T, db *sqlx.DB, id, name string) {
	t.Helper()
	if _, err := db.Exec("INSERT INTO tag(id, mtime, name) VALUES(?, 1, ?)", id, name); err != nil {
		t.Fatal(err)
	}
}

func createTestAssetTag(t *testing.T, db *sqlx.DB, id, assetId, tagId string) {
	t.Helper()
	if _, err := db.Exec("INSERT INTO asset_tag(id, mtime, asset_id, tag_id) VALUES(?, 1, ?, ?)", id, assetId, tagId); err != nil {
		t.Fatal(err)
	}
}

func TestCreateTagRejectsCaseInsensitiveCollision(t *testing.T) {
	db := openTagTestDB(t)
	createTestTag(t, db, "existing", "Review")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	if _, err = CreateTag(tx, "", "review"); err == nil {
		t.Fatal("expected case-insensitive collision")
	}
}

func TestUpdateTagRejectsCaseInsensitiveCollision(t *testing.T) {
	db := openTagTestDB(t)
	createTestTag(t, db, "source", "Review")
	createTestTag(t, db, "target", "Approved")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	_, err = UpdateTag(tx, "source", "approved", false)
	if !errors.Is(err, error_service.ErrTagExists) {
		t.Fatalf("expected tag collision, got %v", err)
	}
}

func TestUpdateTagAllowsCaseOnlyRename(t *testing.T) {
	db := openTagTestDB(t)
	createTestTag(t, db, "tag", "review")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	tag, err := UpdateTag(tx, "tag", "Review", false)
	if err != nil {
		t.Fatal(err)
	}
	if tag.Name != "Review" {
		t.Fatalf("expected case-only rename, got %q", tag.Name)
	}
}

func TestUpdateTagMergesAssignmentsWithoutDuplicates(t *testing.T) {
	db := openTagTestDB(t)
	createTestTag(t, db, "source", "Review")
	createTestTag(t, db, "target", "Approved")
	createTestAssetTag(t, db, "source-only", "asset-1", "source")
	createTestAssetTag(t, db, "source-duplicate", "asset-2", "source")
	createTestAssetTag(t, db, "target-existing", "asset-2", "target")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	tag, err := UpdateTag(tx, "source", "Approved", true)
	if err != nil {
		t.Fatal(err)
	}
	if tag.Id != "target" {
		t.Fatalf("expected target tag, got %q", tag.Id)
	}
	var assignmentCount int
	if err = tx.Get(&assignmentCount, "SELECT COUNT(*) FROM asset_tag WHERE tag_id = 'target'"); err != nil {
		t.Fatal(err)
	}
	if assignmentCount != 2 {
		t.Fatalf("expected two distinct assignments, got %d", assignmentCount)
	}
	var movedTagId string
	if err = tx.Get(&movedTagId, "SELECT tag_id FROM asset_tag WHERE id = 'source-only'"); err != nil {
		t.Fatal(err)
	}
	if movedTagId != "target" {
		t.Fatalf("expected relationship ID to be preserved, got tag %q", movedTagId)
	}
	var tombCount int
	if err = tx.Get(&tombCount, "SELECT COUNT(*) FROM tomb WHERE id IN ('source', 'source-duplicate')"); err != nil {
		t.Fatal(err)
	}
	if tombCount != 2 {
		t.Fatalf("expected tag and duplicate relationship tombstones, got %d", tombCount)
	}
}

func TestDeleteTagRemovesAssignmentsAndCreatesTombstones(t *testing.T) {
	db := openTagTestDB(t)
	createTestTag(t, db, "tag", "Review")
	createTestAssetTag(t, db, "assignment-1", "asset-1", "tag")
	createTestAssetTag(t, db, "assignment-2", "asset-2", "tag")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	usageCount, err := GetTagUsageCount(tx, "tag")
	if err != nil {
		t.Fatal(err)
	}
	if usageCount != 2 {
		t.Fatalf("expected two assignments, got %d", usageCount)
	}
	if err = DeleteTag(tx, "tag"); err != nil {
		t.Fatal(err)
	}
	var tombCount int
	if err = tx.Get(&tombCount, "SELECT COUNT(*) FROM tomb WHERE id IN ('tag', 'assignment-1', 'assignment-2')"); err != nil {
		t.Fatal(err)
	}
	if tombCount != 3 {
		t.Fatalf("expected three tombstones, got %d", tombCount)
	}
}
