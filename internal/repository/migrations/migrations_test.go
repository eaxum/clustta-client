package migrations

import "testing"

func TestRunMigrationsRejectsNewerSchema(t *testing.T) {
	if err := RunMigrations(nil, "2.10", ""); err == nil {
		t.Fatal("expected newer schema to be rejected")
	}
}
