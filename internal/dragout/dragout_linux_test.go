//go:build linux && cgo

package dragout

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinuxFileURIs(t *testing.T) {
	paths := []string{
		filepath.Join(t.TempDir(), "space #100%.blend"),
		filepath.Join(t.TempDir(), "\u753b\u50cf\nfile.blend"),
	}
	for _, path := range paths {
		if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	uris, err := fileURIs(paths)
	if err != nil || len(uris) != len(paths) {
		t.Fatalf("uris=%v err=%v", uris, err)
	}
	for index, uri := range uris {
		parsed, err := url.Parse(uri)
		if err != nil || parsed.Scheme != "file" || parsed.Host != "" || parsed.Path != paths[index] {
			t.Fatalf("invalid file URI %q: %v", uri, err)
		}
		if strings.ContainsAny(uri, " \r\n#") {
			t.Fatalf("URI contains unescaped characters: %q", uri)
		}
	}
}

func TestLinuxFileURIsRejectInvalidSelection(t *testing.T) {
	root := t.TempDir()
	valid := filepath.Join(root, "asset.blend")
	if err := os.WriteFile(valid, nil, 0600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"relative.blend", "/tmp/null\x00.blend", "/tmp/invalid\xff.blend", root, filepath.Join(root, "missing")} {
		if uris, err := fileURIs([]string{valid, path}); err == nil || uris != nil {
			t.Fatalf("invalid selection returned uris=%v err=%v", uris, err)
		}
	}
}

func TestBeginRejectsMissingLinuxWindow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "asset.blend")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	completions := 0
	Begin(nil, []string{path}, func(result Result, err error) {
		completions++
		if result != "" || err == nil {
			t.Fatalf("result=%q err=%v", result, err)
		}
	})
	if completions != 1 {
		t.Fatalf("expected one native completion, got %d", completions)
	}
}
