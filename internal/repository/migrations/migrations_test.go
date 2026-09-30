package migrations

import "testing"

func TestRunMigrationsRejectsNewerSchema(t *testing.T) {
	if err := RunMigrations(nil, "2.10", ""); err == nil {
		t.Fatal("expected newer schema to be rejected")
	}
}

func TestNormalizeLegacyIntegerVersion(t *testing.T) {
	if version := normalizeLegacyVersion("2"); version != "2.0" {
		t.Fatalf("expected 2.0, got %s", version)
	}
	if version := normalizeLegacyVersion("02"); version != "02" {
		t.Fatalf("noncanonical version was rewritten to %s", version)
	}
}
