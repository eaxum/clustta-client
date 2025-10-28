package custom_thumbnail

import (
	"path/filepath"
	"strings"
)

// ThumbnailExtractor defines the interface for file-specific thumbnail extractors
type ThumbnailExtractor interface {
	CanHandle(extension string) bool
	ExtractThumbnail(filePath string) ([]byte, error)
	GetName() string
}

type ExtractorRegistry struct {
	extractors []ThumbnailExtractor
}

var registry = &ExtractorRegistry{
	extractors: []ThumbnailExtractor{
		&BlenderExtractor{},
		&MayaExtractor{},
	},
}

// GetThumbnail attempts to extract a thumbnail from the given file
// Returns PNG bytes if found, nil if no thumbnail available
func GetThumbnail(filePath string) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	for _, extractor := range registry.extractors {
		if extractor.CanHandle(ext) {
			thumbnailBytes, err := extractor.ExtractThumbnail(filePath)
			if err != nil {
				continue
			}

			if len(thumbnailBytes) > 0 {
				return thumbnailBytes, nil
			}
		}
	}

	return nil, nil
}

// RegisterExtractor adds a custom extractor to the registry
func RegisterExtractor(extractor ThumbnailExtractor) {
	registry.extractors = append(registry.extractors, extractor)
}

// GetSupportedExtensions returns all file extensions supported by custom extractors
func GetSupportedExtensions() []string {
	extensionSet := make(map[string]bool)

	testExtensions := []string{
		".blend", ".ma", ".mb", ".max", ".c4d", ".hip", ".hiplc",
		".sbs", ".sbsar", ".nk", ".comp", ".zpr", ".ztl",
	}

	for _, ext := range testExtensions {
		for _, extractor := range registry.extractors {
			if extractor.CanHandle(ext) {
				extensionSet[ext] = true
				break
			}
		}
	}

	extensions := make([]string, 0, len(extensionSet))
	for ext := range extensionSet {
		extensions = append(extensions, ext)
	}

	return extensions
}

// GetExtractorForFile returns the name of the extractor that handles the given file
func GetExtractorForFile(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	for _, extractor := range registry.extractors {
		if extractor.CanHandle(ext) {
			return extractor.GetName()
		}
	}

	return ""
}
