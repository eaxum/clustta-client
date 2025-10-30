//go:build windows

package services

import (
	"path/filepath"
	"syscall"
	"unsafe"
)

func (f *FSService) LaunchFileWith(path string) error {
	filePath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	h := syscall.MustLoadDLL("shell32.dll")
	c := h.MustFindProc("ShellExecuteW")

	openWithPtr, err := syscall.UTF16PtrFromString("rundll32.exe")
	if err != nil {
		return err
	}

	paramsPtr, err := syscall.UTF16PtrFromString("shell32.dll,OpenAs_RunDLL " + filePath)
	if err != nil {
		return err
	}

	ret, _, err := c.Call(
		0,                                    // hwnd
		0,                                    // verb (NULL for default)
		uintptr(unsafe.Pointer(openWithPtr)), // file
		uintptr(unsafe.Pointer(paramsPtr)),   // params
		0,                                    // directory
		1,                                    // show
	)

	if ret <= 32 {
		return err
	}
	return nil
}
