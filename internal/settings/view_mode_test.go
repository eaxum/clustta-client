package settings

import "testing"

func TestNormalizeViewModeAcceptsColumns(t *testing.T) {
	if got := normalizeViewMode(viewModeColumns); got != viewModeColumns {
		t.Fatalf("normalizeViewMode(%q) = %q", viewModeColumns, got)
	}
}

func TestNormalizeViewModeFallsBackToList(t *testing.T) {
	if got := normalizeViewMode("unknown"); got != viewModeList {
		t.Fatalf("normalizeViewMode(%q) = %q", "unknown", got)
	}
}
