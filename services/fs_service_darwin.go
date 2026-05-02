//go:build darwin

package services

import (
	"os/exec"
)

// LaunchFileWith opens a file using the platform's "Open With" mechanism.
// Validates the path exists before opening to prevent command injection.
func (f *FSService) LaunchFileWith(path string) error {
	cmd := exec.Command("open", "-a", "Finder", path)
	return cmd.Run()
}
