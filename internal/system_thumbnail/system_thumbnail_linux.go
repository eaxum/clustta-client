//go:build linux

package system_thumbnail

import (
	"fmt"
)

// ThumbnailOptions configures thumbnail generation behavior
type ThumbnailOptions uint32

const (
	ThumbnailDefault         ThumbnailOptions = 0
	ThumbnailOnlyIfCached    ThumbnailOptions = 1 << 0 // Only return if already cached
	ThumbnailIconFallback    ThumbnailOptions = 1 << 1 // Fall back to icon if thumbnail fails
	ThumbnailUseCurrentScale ThumbnailOptions = 1 << 2 // Apply DPI scaling
	ThumbnailHighQuality     ThumbnailOptions = 1 << 3 // Prefer quality over speed
)

// GetOSThumbnail generates a thumbnail for the specified file using Linux APIs
// Note: This is a stub implementation. Linux thumbnail generation would require
// using freedesktop.org thumbnail spec or other native integration.
func GetOSThumbnail(filePath string, size int, options ThumbnailOptions) ([]byte, error) {
	return nil, fmt.Errorf("OS thumbnail generation not implemented for Linux")
}

// GetCachedThumbnail attempts to get a cached thumbnail without generating a new one
func GetCachedThumbnail(filePath string, size int) ([]byte, error) {
	return nil, fmt.Errorf("cached thumbnail retrieval not implemented for Linux")
}
