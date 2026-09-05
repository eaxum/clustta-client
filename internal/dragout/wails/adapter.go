package wails

import (
	"context"

	"clustta/internal/dragout"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Start waits on the calling goroutine while native drag runs on the UI thread.
func Start(ctx context.Context, window *application.WebviewWindow, paths []string) (dragout.Result, error) {
	type outcome struct {
		result dragout.Result
		err    error
	}
	completed := make(chan outcome, 1)
	application.InvokeAsync(func() {
		if err := ctx.Err(); err != nil {
			completed <- outcome{err: err}
			return
		}
		result, err := dragout.Start(window.NativeWindow(), paths)
		completed <- outcome{result, err}
	})
	// Cancellation must not free an active native source or allow an overlapping session.
	result := <-completed
	return result.result, result.err
}
