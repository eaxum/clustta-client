//go:build darwin

package services

import (
	"os/exec"
)

//LaunchFileWith opens the macOS Finder at the specified path.
//Returns an error if the operation fails.
func (f *FSService) LaunchFileWith(path string) error {
	cmd := exec.Command("open", "-a", "Finder", path)
	return cmd.Run()
}
