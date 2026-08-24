package repository

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func openCheckpointCommentTestTx(t *testing.T) *sqlx.Tx {
	t.Helper()
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err = db.Exec(`CREATE TABLE asset_checkpoint (
		id TEXT PRIMARY KEY,
		asset_id TEXT NOT NULL,
		trashed INTEGER NOT NULL DEFAULT 0
	)`); err != nil {
		t.Fatal(err)
	}
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { tx.Rollback() })
	return tx
}

func TestResolveCheckpointCommentKeepsExplicitComment(t *testing.T) {
	tx := openCheckpointCommentTestTx(t)
	comment, err := resolveCheckpointComment(tx, "asset-id", "  lighting update  ")
	if err != nil {
		t.Fatal(err)
	}
	if comment != "lighting update" {
		t.Fatalf("expected trimmed comment, got %q", comment)
	}
}

func TestResolveCheckpointCommentStartsAtVersionOne(t *testing.T) {
	tx := openCheckpointCommentTestTx(t)
	comment, err := resolveCheckpointComment(tx, "asset-id", "")
	if err != nil {
		t.Fatal(err)
	}
	if comment != "v0001" {
		t.Fatalf("expected v0001, got %q", comment)
	}
}

func TestResolveCheckpointCommentUsesHistoricalCount(t *testing.T) {
	tx := openCheckpointCommentTestTx(t)
	if _, err := tx.Exec(`INSERT INTO asset_checkpoint(id, asset_id, trashed) VALUES
		('first', 'asset-id', 0),
		('second', 'asset-id', 1),
		('other', 'other-asset', 0)`); err != nil {
		t.Fatal(err)
	}

	comment, err := resolveCheckpointComment(tx, "asset-id", "   ")
	if err != nil {
		t.Fatal(err)
	}
	if comment != "v0003" {
		t.Fatalf("expected v0003, got %q", comment)
	}
}
