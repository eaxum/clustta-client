//go:build windows && cgo

package dragout

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestShellDataObject(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	root := t.TempDir()
	var paths []string
	for _, name := range []string{"first", "second"} {
		directory := filepath.Join(root, name)
		if err := os.Mkdir(directory, 0700); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(directory, "test space 日本語.blend")
		if err := os.WriteFile(path, []byte("fixture"), 0600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	for _, files := range [][]string{paths[:1], paths} {
		if err := probeFiles(files); err != nil {
			t.Fatal(err)
		}
	}
	if err := probeFiles([]string{filepath.Join(root, "missing.blend")}); err == nil {
		t.Fatal("missing file accepted")
	}
}

func TestEncodePathsRejectsInvalidInput(t *testing.T) {
	for _, paths := range [][]string{nil, {"relative.blend"}, {"C:\\bad\x00.blend"}} {
		if _, err := encodePaths(paths); err == nil {
			t.Fatalf("accepted %q", paths)
		}
	}
}
