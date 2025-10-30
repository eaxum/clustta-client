//go:build darwin

package services

import (
	"os/exec"
)

func (f *FSService) LaunchFileWith(path string) error {
	// macOS "Open With" dialog
	cmd := exec.Command("open", "-a", "Finder", path)
	return cmd.Run()
}
