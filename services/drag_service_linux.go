//go:build linux

package services

import (
	"fmt"
)

// StartNativeDrag initiates a native Linux drag-and-drop operation.
// Linux implementation is pending due to GTK/Wayland complexity.
// Returns an error indicating the feature is not yet available.
func (d *DragService) StartNativeDrag(filePaths []string) (int, error) {
	if len(filePaths) == 0 {
		return 0, nil
	}

	// Linux native drag-drop requires GTK widget integration or X11/Wayland protocols
	// This is significantly more complex than Windows/macOS due to:
	// 1. GTK requires drag to originate from widget event handlers
	// 2. Wayland has stricter security model than X11
	// 3. Need to integrate with WebKitGTK widget used by Wails

	// For now, return an error - users can use "Reveal in Files" as workaround
	return 0, fmt.Errorf("native drag-out is not yet supported on Linux; use 'Reveal in Files' instead")
}

// IsMouseButtonDown checks if the left mouse button is currently pressed.
// Returns false on Linux as native drag is not implemented.
func (d *DragService) IsMouseButtonDown() bool {
	// Would require X11/Wayland query - not implemented
	return false
}
