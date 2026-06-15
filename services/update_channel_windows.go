//go:build windows

package services

import (
	"syscall"
	"unsafe"
)

// appModelErrorNoPackage is returned when the process is not in an MSIX package.
const appModelErrorNoPackage = 15700

// detectPackaging returns "msstore" for MSIX builds, otherwise "direct".
func detectPackaging() string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetCurrentPackageFullName")
	if err := proc.Find(); err != nil {
		return "direct"
	}

	var length uint32
	r1, _, _ := proc.Call(uintptr(unsafe.Pointer(&length)), 0)
	if r1 == appModelErrorNoPackage {
		return "direct"
	}
	return "msstore"
}
