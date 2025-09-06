//go:build !darwin

package settings

import (
	"errors"
)

func CreateBookmarkFromPath(path string) ([]byte, error) {
	return []byte{}, nil
}

func ResolveBookmark(bookmarkData []byte) (string, error) {
	return "", errors.New("bookmarks are not supported on this platform")
}

func IsBookmarkStale(bookmarkData []byte) bool {
	return len(bookmarkData) == 0
}

func StopAccessingResource(bookmarkData []byte) {
}
