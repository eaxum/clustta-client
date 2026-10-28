package bridge

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRemoveBridgeTokenOnlyRemovesMatchingToken(t *testing.T) {
	tokenPath := configureTokenTestDirectory(t)

	if err := writeBridgeToken("active-token"); err != nil {
		t.Fatalf("write token: %v", err)
	}
	if err := os.WriteFile(tokenPath, []byte("new-token"), 0600); err != nil {
		t.Fatalf("replace token: %v", err)
	}
	if err := removeBridgeToken("active-token"); err != nil {
		t.Fatalf("remove stale token: %v", err)
	}

	storedToken, err := os.ReadFile(tokenPath)
	if err != nil {
		t.Fatalf("read preserved token: %v", err)
	}
	if string(storedToken) != "new-token" {
		t.Fatalf("expected new token to remain, got %q", storedToken)
	}

	if err := removeBridgeToken("new-token"); err != nil {
		t.Fatalf("remove current token: %v", err)
	}
	if _, err := os.Stat(tokenPath); !os.IsNotExist(err) {
		t.Fatalf("expected token file to be removed, got %v", err)
	}
}

func configureTokenTestDirectory(t *testing.T) string {
	t.Helper()
	tempDirectory := t.TempDir()
	t.Setenv("APPDATA", tempDirectory)
	t.Setenv("HOME", tempDirectory)

	baseDirectory := tempDirectory
	if runtime.GOOS == "darwin" {
		baseDirectory = filepath.Join(tempDirectory, "Library", "Application Support")
	} else if runtime.GOOS != "windows" {
		baseDirectory = filepath.Join(tempDirectory, ".config")
	}
	return filepath.Join(baseDirectory, "Clustta", ".bridge-token")
}
