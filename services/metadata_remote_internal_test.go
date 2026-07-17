package services

import (
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
)

func openMetadataTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Open("sqlite3", filepath.Join(t.TempDir(), "project.clst"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(repository.ProjectSchema); err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func setTestSyncToken(t *testing.T, db *sqlx.DB, token string) {
	t.Helper()
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if err = utils.SetProjectSyncToken(tx, token); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func getTestSyncToken(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	token, err := utils.GetProjectSyncToken(tx)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestReturnedSyncTokenAdvancesForContinuousCleanState(t *testing.T) {
	db := openMetadataTestDB(t)
	setTestSyncToken(t, db, "before")

	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if err = applyCanonicalAssetType(tx, models.AssetType{Id: "type-1", MTime: 2, Name: "Animation", Icon: "animation"}); err != nil {
		t.Fatal(err)
	}
	previous := "before"
	if err = applyReturnedSyncToken(tx, &previous, "after"); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if token := getTestSyncToken(t, db); token != "after" {
		t.Fatalf("expected token to advance, got %q", token)
	}
}

func TestReturnedSyncTokenStaysWhenOtherRowsAreUnsynced(t *testing.T) {
	db := openMetadataTestDB(t)
	setTestSyncToken(t, db, "before")
	if _, err := db.Exec("INSERT INTO asset_type(id,mtime,name,icon,synced) VALUES('dirty',1,'Dirty','dirty',0)"); err != nil {
		t.Fatal(err)
	}

	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if err = applyCanonicalAssetType(tx, models.AssetType{Id: "type-1", MTime: 2, Name: "Animation", Icon: "animation"}); err != nil {
		t.Fatal(err)
	}
	previous := "before"
	if err = applyReturnedSyncToken(tx, &previous, "after"); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if token := getTestSyncToken(t, db); token != "before" {
		t.Fatalf("dirty data must retain the old token, got %q", token)
	}
}

func TestReturnedSyncTokenStaysWhenPredecessorDoesNotMatch(t *testing.T) {
	db := openMetadataTestDB(t)
	setTestSyncToken(t, db, "local-token")

	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	previous := "different-server-token"
	if err = applyReturnedSyncToken(tx, &previous, "after"); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if token := getTestSyncToken(t, db); token != "local-token" {
		t.Fatalf("token discontinuity must retain the local token, got %q", token)
	}
}

func TestCanonicalAssetDoesNotOverwritePreexistingDirtyRow(t *testing.T) {
	db := openMetadataTestDB(t)
	_, err := db.Exec(`INSERT INTO asset(id,created_at,mtime,name,extension,status_id,asset_type_id,synced)
		VALUES('asset-1','now',1,'Local edit','.txt','status','type',0)`)
	if err != nil {
		t.Fatal(err)
	}

	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	clean, err := applyCanonicalAssets(tx, []models.Asset{{Id: "asset-1", CreatedAt: "now", MTime: 2, Name: "Server value", Extension: ".txt", StatusId: "status", AssetTypeId: "type"}})
	if err != nil {
		t.Fatal(err)
	}
	if clean {
		t.Fatal("expected the dirty target to reject canonical replacement")
	}
	var name string
	if err = tx.Get(&name, "SELECT name FROM asset WHERE id='asset-1'"); err != nil {
		t.Fatal(err)
	}
	if name != "Local edit" {
		t.Fatalf("dirty row was overwritten: %q", name)
	}
}
