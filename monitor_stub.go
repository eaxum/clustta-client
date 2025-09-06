//go:build !darwin

package main

// StartFullscreenMonitoring starts monitoring fullscreen events for a window with the given title
func StartFullscreenMonitoring(windowTitle string) error {
	return nil
}

// StopFullscreenMonitoring stops monitoring fullscreen events
func StopFullscreenMonitoring() {

}
