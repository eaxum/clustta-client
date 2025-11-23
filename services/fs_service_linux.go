//go:build linux

package services

import "fmt"

//LaunchFileWith is not supported on Linux.
//Returns an error indicating unsupported platform.
func (f *FSService) LaunchFileWith(path string) error {
	return fmt.Errorf("LaunchFileWith not supported on Linux")
}
