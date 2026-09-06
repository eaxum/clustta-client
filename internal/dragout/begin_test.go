package dragout

import "testing"

func TestBeginRejectsEmptySelectionOnce(t *testing.T) {
	completions := 0
	Begin(nil, nil, func(result Result, err error) {
		completions++
		if result != "" || err == nil {
			t.Fatalf("result=%q err=%v", result, err)
		}
	})
	if completions != 1 {
		t.Fatalf("expected one completion, got %d", completions)
	}
}
