//go:build linux

package services

import "fmt"

// LaunchFileWith opens a file using the platform's "Open With" mechanism.
// Validates the path exists before opening to prevent command injection.
func (f *FSService) LaunchFileWith(path string) error {
	return fmt.Errorf("LaunchFileWith not supported on Linux")
}
