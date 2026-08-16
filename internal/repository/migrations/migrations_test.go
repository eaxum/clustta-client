package migrations

import "testing"

func TestRunMigrationsRejectsNewerSchema(t *testing.T) {
	if err := RunMigrations(nil, LatestVersion+0.1, ""); err == nil {
		t.Fatal("expected newer schema to be rejected")
	}
}
