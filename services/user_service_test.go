package services

import (
	"errors"
	"testing"
)

func TestNormalizeRoleName(t *testing.T) {
	name, err := normalizeRoleName("  Lead Artist  ")
	if err != nil {
		t.Fatalf("normalizeRoleName returned an error: %v", err)
	}
	if name != "Lead Artist" {
		t.Fatalf("normalizeRoleName returned %q", name)
	}
}

func TestNormalizeRoleNameRejectsWhitespace(t *testing.T) {
	if _, err := normalizeRoleName("   "); err == nil {
		t.Fatal("normalizeRoleName accepted a whitespace-only name")
	}
}

func TestIsRoleNameCollision(t *testing.T) {
	err := errors.New("UNIQUE constraint failed: role.name")
	if !isRoleNameCollision(err) {
		t.Fatal("isRoleNameCollision did not recognize a role name collision")
	}
}
