//go:build !windows

package services

import "os"

func replaceCodeFile(source, destination string) error {
	return os.Rename(source, destination)
}
