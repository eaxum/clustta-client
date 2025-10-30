//go:build linux

package services

import "fmt"

func (f *FSService) LaunchFileWith(path string) error {
	return fmt.Errorf("LaunchFileWith not supported on Linux")
}
