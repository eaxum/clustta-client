package repository

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"clustta/internal/repository/migrations"
	"clustta/internal/utils"
)

func TestProjectMigrationFixtures(t *testing.T) {
	fixtureRoot := os.Getenv("CLUSTTA_MIGRATION_FIXTURES")
	if fixtureRoot == "" {
		t.Skip("CLUSTTA_MIGRATION_FIXTURES is not set")
	}

	for _, version := range []string{"v2_0", "v2_1"} {
		t.Run(version, func(t *testing.T) {
			sourcePath := filepath.Join(fixtureRoot, version, "Tired King.clst")
			projectPath := filepath.Join(t.TempDir(), "Tired King.clst")
			copyMigrationFixture(t, sourcePath, projectPath)

			beforeDependencies := migrationRowCount(t, projectPath, "asset_dependency")
			beforeCheckpoints := migrationRowCount(t, projectPath, "asset_checkpoint")

			if err := UpdateProject(projectPath); err != nil {
				t.Fatal(err)
			}
			if err := UpdateProject(projectPath); err != nil {
				t.Fatalf("second migration failed: %v", err)
			}

			db, err := utils.OpenDb(projectPath)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			var version string
			if err = db.Get(&version, "SELECT value FROM config WHERE name = 'version'"); err != nil {
				t.Fatal(err)
			}
			if version != migrations.LatestVersion {
				t.Fatalf("expected schema %s, got %s", migrations.LatestVersion, version)
			}

			var integrity string
			if err = db.Get(&integrity, "PRAGMA quick_check"); err != nil {
				t.Fatal(err)
			}
			if integrity != "ok" {
				t.Fatalf("integrity check failed: %s", integrity)
			}

			if after := migrationRowCount(t, projectPath, "asset_dependency"); after != beforeDependencies {
				t.Fatalf("dependency count changed from %d to %d", beforeDependencies, after)
			}
			if after := migrationRowCount(t, projectPath, "asset_checkpoint"); after != beforeCheckpoints {
				t.Fatalf("checkpoint count changed from %d to %d", beforeCheckpoints, after)
			}
		})
	}
}

func copyMigrationFixture(t *testing.T, sourcePath, destinationPath string) {
	t.Helper()
	source, err := os.Open(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()

	destination, err := os.Create(destinationPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.Copy(destination, source); err != nil {
		destination.Close()
		t.Fatal(err)
	}
	if err = destination.Close(); err != nil {
		t.Fatal(err)
	}
}

func migrationRowCount(t *testing.T, projectPath, table string) int {
	t.Helper()
	db, err := utils.OpenDb(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var count int
	if err = db.Get(&count, "SELECT count(*) FROM "+table); err != nil {
		t.Fatal(err)
	}
	return count
}
