//go:build windows && cgo

package dragout

/*
#cgo LDFLAGS: -lole32 -lshell32 -luuid -luser32
#include "dragout_windows.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf16"
	"unsafe"
)

const Available = true

const (
	nativeDropped   = 0x00040100
	nativeCancelled = 0x00040101
)

// Start runs a copy-only OLE drag on the caller's UI thread until drop or cancellation.
func Start(window unsafe.Pointer, paths []string) (Result, error) {
	encoded, err := encodePaths(paths)
	if err != nil {
		return "", err
	}
	result := uint32(C.clustta_drag_files(C.uintptr_t(uintptr(window)), (*C.uint16_t)(unsafe.Pointer(&encoded[0])), C.size_t(len(paths))))
	switch result {
	case nativeDropped:
		return Dropped, nil
	case nativeCancelled:
		return Cancelled, nil
	default:
		return "", fmt.Errorf("Windows could not drag the selected files (HRESULT 0x%08X)", result)
	}
}

func encodePaths(paths []string) ([]uint16, error) {
	if len(paths) == 0 {
		return nil, errors.New("select at least one local file")
	}
	var encoded []uint16
	for _, path := range paths {
		if !filepath.IsAbs(path) || strings.ContainsRune(path, 0) {
			return nil, errors.New("drag paths must be absolute and contain no null characters")
		}
		encoded = append(encoded, utf16.Encode([]rune(path))...)
		encoded = append(encoded, 0)
	}
	return append(encoded, 0), nil
}

func probeFiles(paths []string) error {
	encoded, err := encodePaths(paths)
	if err != nil {
		return err
	}
	result := uint32(C.clustta_probe_drag_files((*C.uint16_t)(unsafe.Pointer(&encoded[0])), C.size_t(len(paths))))
	if result != 0 {
		return fmt.Errorf("shell data object failed: 0x%08X", result)
	}
	return nil
}
