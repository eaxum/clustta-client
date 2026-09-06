//go:build linux && cgo

package dragout

/*
#cgo pkg-config: gtk+-3.0
#include <stdlib.h>
#include "dragout_linux.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime/cgo"
	"strings"
	"unicode/utf8"
	"unsafe"
)

const Available = true

// Begin starts a copy-only GTK session on the UI thread and completes after drag-end.
func Begin(window unsafe.Pointer, paths []string, complete func(Result, error)) {
	uris, err := fileURIs(paths)
	if err != nil {
		complete("", err)
		return
	}
	encoded := C.CString(strings.Join(uris, "\x00"))
	defer C.free(unsafe.Pointer(encoded))
	callback := cgo.NewHandle(complete)
	C.clustta_linux_begin_drag(window, encoded, C.size_t(len(uris)), C.uintptr_t(callback))
}

func fileURIs(paths []string) ([]string, error) {
	if len(paths) == 0 {
		return nil, errors.New("select at least one local file")
	}
	uris := make([]string, 0, len(paths))
	for _, path := range paths {
		if !filepath.IsAbs(path) || strings.ContainsRune(path, 0) || !utf8.ValidString(path) {
			return nil, errors.New("drag paths must be absolute UTF-8 and contain no null characters")
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("local file unavailable: %w", err)
		}
		if !info.Mode().IsRegular() {
			return nil, errors.New("only regular local files can be dragged")
		}
		uri := url.URL{Scheme: "file", Path: path}
		uris = append(uris, uri.String())
	}
	return uris, nil
}

//export clustta_linux_drag_completed
func clustta_linux_drag_completed(callback C.uintptr_t, status C.int) {
	handle := cgo.Handle(callback)
	complete := handle.Value().(func(Result, error))
	handle.Delete()
	switch status {
	case C.ClusttaLinuxDragDropped:
		complete(Dropped, nil)
	case C.ClusttaLinuxDragCancelled:
		complete(Cancelled, nil)
	case C.ClusttaLinuxDragInvalidWindow:
		complete("", errors.New("dragging requires an active desktop window on the GTK main thread"))
	case C.ClusttaLinuxDragBusy:
		complete("", errors.New("a file drag is already in progress"))
	default:
		complete("", errors.New("Linux could not start dragging the selected files"))
	}
}
