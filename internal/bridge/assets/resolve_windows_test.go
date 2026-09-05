package assets

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"clustta/internal/repository/models"
)

func TestLocalAssetJunctionEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "asset.blend"), []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "junction")
	if output, err := exec.Command("cmd", "/c", "mklink", "/J", link, outside).CombinedOutput(); err != nil {
		t.Fatalf("create test junction: %s: %v", output, err)
	}
	defer func() {
		if err := os.Remove(link); err != nil {
			t.Errorf("remove test junction: %v", err)
		}
	}()
	if _, err := localAssetPath(root, models.Asset{FilePath: filepath.Join(link, "asset.blend")}); err == nil {
		t.Fatal("junction escape accepted")
	}
}
