//go:build darwin && cgo

package dragout

import "testing"

func TestBeginRejectsInvalidMacPaths(t *testing.T) {
	for _, path := range []string{"relative.blend", "/tmp/null\x00.blend", "/tmp/invalid\xff.blend"} {
		completions := 0
		Begin(nil, []string{path}, func(result Result, err error) {
			completions++
			if result != "" || err == nil {
				t.Fatalf("path=%q result=%q err=%v", path, result, err)
			}
		})
		if completions != 1 {
			t.Fatalf("path=%q completions=%d", path, completions)
		}
	}
}

func TestBeginRejectsMissingMacWindow(t *testing.T) {
	completions := 0
	Begin(nil, []string{"/tmp/asset.blend"}, func(result Result, err error) {
		completions++
		if result != "" || err == nil {
			t.Fatalf("result=%q err=%v", result, err)
		}
	})
	if completions != 1 {
		t.Fatalf("expected one native completion, got %d", completions)
	}
}
