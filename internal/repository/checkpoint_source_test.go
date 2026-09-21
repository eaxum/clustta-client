package repository

import (
	"path/filepath"
	"strings"
	"testing"

	"clustta/internal/repository/migrations"
	"clustta/internal/repository/models"

	"github.com/jmoiron/sqlx"
)

func TestCheckpointSourceEditsAndProtection(t *testing.T) {
	_, tx := openDependencyTestDB(t)
	insertDependencyAsset(t, tx, "source")
	insertDependencyAsset(t, tx, "output")
	insertTestCheckpoint(t, tx, "source-1", "source", "source-group", 1)
	insertTestCheckpoint(t, tx, "source-2", "source", "source-group", 2)
	insertTestCheckpoint(t, tx, "output-1", "output", "output-group", 3)
	source, err := ResolveCheckpointSource(tx, "source", "")
	if err != nil || source == nil || *source != "source-2" {
		t.Fatalf("latest source: %v, %v", source, err)
	}
	if _, err = ResolveCheckpointSource(tx, "output", "source-2"); err == nil {
		t.Fatal("accepted checkpoint from wrong asset")
	}
	if err = UpdateCheckpoint(tx, "output-1", "Export", source); err != nil {
		t.Fatal(err)
	}
	checkpoint, err := GetCheckpoint(tx, "output-1")
	if err != nil {
		t.Fatal(err)
	}
	if checkpoint.Comment != "Export" || *checkpoint.SourceCheckpointId != "source-2" || checkpoint.Synced {
		t.Fatalf("incorrect metadata: %+v", checkpoint)
	}
	if checkpoint.AuthorUID != "author" || checkpoint.GroupId != "output-group" {
		t.Fatal("edit changed checkpoint identity")
	}
	mtime := checkpoint.MTime
	if err = UpdateCheckpoint(tx, checkpoint.Id, checkpoint.Comment, source); err != nil {
		t.Fatal(err)
	}
	unchanged, _ := GetCheckpoint(tx, checkpoint.Id)
	if unchanged.MTime != mtime {
		t.Fatal("no-op edit changed mtime")
	}
	outputId := "output-1"
	if err = UpdateCheckpoint(tx, "source-2", "cycle", &outputId); err == nil {
		t.Fatal("accepted cycle")
	}
	if err = UpdateCheckpoint(tx, "output-1", "self", &outputId); err == nil {
		t.Fatal("accepted self-reference")
	}
	for _, query := range []string{
		"DELETE FROM asset_checkpoint WHERE id = 'source-2'",
		"UPDATE asset_checkpoint SET trashed = 1 WHERE id = 'source-2'",
		"UPDATE asset SET trashed = 1 WHERE id = 'source'",
	} {
		if _, err = tx.Exec(query); err == nil {
			t.Fatalf("allowed referenced source deletion: %s", query)
		}
	}
	if err = UpdateCheckpoint(tx, "output-1", "Corrected", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec("DELETE FROM asset_checkpoint WHERE id = 'source-2'"); err != nil {
		t.Fatal(err)
	}
	missing := "missing"
	if err = UpdateCheckpoint(tx, "output-1", "bad source", &missing); err == nil {
		t.Fatal("accepted missing source")
	}
}

func TestCheckpointSourceSyncAndProtobuf(t *testing.T) {
	_, tx := openDependencyTestDB(t)
	insertDependencyAsset(t, tx, "output")
	sourceId := "source-not-in-partial-sync"
	checkpoint := models.Checkpoint{Id: "output-1", AssetId: "output", MTime: 2,
		CreatedAt: "2026-01-01T00:00:00Z", Comment: "Export", SourceCheckpointId: &sourceId}
	roundTrip := FromPbCheckpoints(ToPbCheckpoints([]models.Checkpoint{checkpoint}))
	if roundTrip[0].SourceCheckpointId == nil || *roundTrip[0].SourceCheckpointId != sourceId {
		t.Fatal("protobuf lost source")
	}
	if err := SaveCheckpoints(tx, roundTrip); err != nil {
		t.Fatal(err)
	}
	checkpoint.MTime = 3
	checkpoint.Comment = "Corrected"
	checkpoint.SourceCheckpointId = nil
	if err := SaveCheckpoints(tx, []models.Checkpoint{checkpoint}); err != nil {
		t.Fatal(err)
	}
	if err := SaveCheckpoints(tx, roundTrip); err != nil {
		t.Fatal(err)
	}
	saved, err := GetCheckpoint(tx, checkpoint.Id)
	if err != nil {
		t.Fatal(err)
	}
	if saved.SourceCheckpointId != nil || saved.Comment != "Corrected" || saved.MTime != 3 {
		t.Fatalf("stale sync overwrote edit: %+v", saved)
	}
}

func TestCheckpointSourceMigrationFromOlderProject(t *testing.T) {
	db, err := sqlx.Open("sqlite3", filepath.Join(t.TempDir(), "old.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	schema := strings.Replace(ProjectSchema, "    source_checkpoint_id TEXT NULL,\n", "", 1)
	start := strings.Index(schema, "CREATE INDEX IF NOT EXISTS idx_asset_checkpoint_source")
	end := strings.Index(schema, "CREATE TABLE IF NOT EXISTS asset_checkpoint_tag")
	schema = schema[:start] + schema[end:]
	if _, err = db.Exec(schema); err != nil {
		t.Fatal(err)
	}
	if err = migrations.RunMigrations(db, "2.1", ProjectSchema); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = db.Get(&count, "SELECT count(*) FROM pragma_table_info('asset_checkpoint') WHERE name = 'source_checkpoint_id'"); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("source column missing")
	}
	if err = migrations.RunMigrations(db, migrations.LatestVersion, ProjectSchema); err != nil {
		t.Fatal(err)
	}
}

func TestCheckpointSourceSyncReversesLinksInOneBatch(t *testing.T) {
	_, tx := openDependencyTestDB(t)
	insertDependencyAsset(t, tx, "asset")
	insertTestCheckpoint(t, tx, "a", "asset", "group", 1)
	insertTestCheckpoint(t, tx, "b", "asset", "group", 2)
	a, _ := GetCheckpoint(tx, "a")
	b, _ := GetCheckpoint(tx, "b")
	bId := b.Id
	if err := UpdateCheckpoint(tx, a.Id, "", &bId); err != nil {
		t.Fatal(err)
	}
	a, _ = GetCheckpoint(tx, "a")
	a.MTime++
	a.SourceCheckpointId = nil
	b.MTime = a.MTime
	b.SourceCheckpointId = &a.Id
	if err := SaveCheckpoints(tx, []models.Checkpoint{b, a}); err != nil {
		t.Fatal(err)
	}
}
