//go:build darwin && cgo

package dragout

/*
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>
#include "dragout_darwin.h"
*/
import "C"

import (
	"errors"
	"path/filepath"
	"runtime/cgo"
	"strings"
	"unicode/utf8"
	"unsafe"
)

const Available = true

// Begin starts an AppKit session on the UI thread; completion arrives when that session ends.
func Begin(window unsafe.Pointer, paths []string, complete func(Result, error)) {
	if len(paths) == 0 {
		complete("", errors.New("select at least one local file"))
		return
	}
	for _, path := range paths {
		if !filepath.IsAbs(path) || strings.ContainsRune(path, 0) || !utf8.ValidString(path) {
			complete("", errors.New("drag paths must be absolute UTF-8 and contain no null characters"))
			return
		}
	}
	encoded := C.CString(strings.Join(paths, "\x00"))
	defer C.free(unsafe.Pointer(encoded))
	callback := cgo.NewHandle(complete)
	C.clustta_begin_drag(window, encoded, C.size_t(len(paths)), C.uintptr_t(callback))
}

//export clustta_drag_completed
func clustta_drag_completed(callback C.uintptr_t, status C.int) {
	handle := cgo.Handle(callback)
	complete := handle.Value().(func(Result, error))
	handle.Delete()
	switch status {
	case C.ClusttaDragDropped:
		complete(Dropped, nil)
	case C.ClusttaDragCancelled:
		complete(Cancelled, nil)
	case C.ClusttaDragInvalidWindow:
		complete("", errors.New("dragging requires an active desktop window on the main thread"))
	case C.ClusttaDragInvalidFile:
		complete("", errors.New("a selected local file is no longer available"))
	case C.ClusttaDragBusy:
		complete("", errors.New("a file drag is already in progress"))
	default:
		complete("", errors.New("macOS could not start dragging the selected files"))
	}
}
