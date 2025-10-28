package custom_thumbnail

import (
	"os"
	"path/filepath"
)

// MayaExtractor handles Maya file thumbnails (.ma and .mb files)
// Maya stores thumbnails as separate .png files adjacent to the scene file
type MayaExtractor struct{}

func (m *MayaExtractor) CanHandle(extension string) bool {
	return extension == ".ma" || extension == ".mb"
}

func (m *MayaExtractor) GetName() string {
	return "MayaExtractor"
}

func (m *MayaExtractor) ExtractThumbnail(filePath string) ([]byte, error) {
	// Maya stores thumbnails as separate .png files
	// Try multiple possible locations

	// Location 1: Same directory as file with .png extension
	// Example: scene.ma -> scene.ma.png
	thumbnailPath := filePath + ".png"
	if thumbnailBytes, err := m.readThumbnailFile(thumbnailPath); err == nil && thumbnailBytes != nil {
		return thumbnailBytes, nil
	}

	// Location 2: Maya swatches directory
	// Example: scene.ma -> .mayaSwatches/scene.ma.png
	dir := filepath.Dir(filePath)
	baseFilename := filepath.Base(filePath)
	swatchesPath := filepath.Join(dir, ".mayaSwatches", baseFilename+".png")
	if thumbnailBytes, err := m.readThumbnailFile(swatchesPath); err == nil && thumbnailBytes != nil {
		return thumbnailBytes, nil
	}

	// Location 3: Alternative naming (without double extension)
	// Example: scene.ma -> scene.png
	nameWithoutExt := baseFilename[:len(baseFilename)-len(filepath.Ext(baseFilename))]
	altPath := filepath.Join(dir, nameWithoutExt+".png")
	if thumbnailBytes, err := m.readThumbnailFile(altPath); err == nil && thumbnailBytes != nil {
		return thumbnailBytes, nil
	}

	// No thumbnail found
	return nil, nil
}

func (m *MayaExtractor) readThumbnailFile(path string) ([]byte, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	// Read the thumbnail file
	thumbnailBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Verify it's a PNG file (optional validation)
	if len(thumbnailBytes) < 8 {
		return nil, nil
	}

	// Check PNG magic bytes: 0x89 0x50 0x4E 0x47 0x0D 0x0A 0x1A 0x0A
	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i := 0; i < 8; i++ {
		if thumbnailBytes[i] != pngMagic[i] {
			return nil, nil // Not a valid PNG
		}
	}

	return thumbnailBytes, nil
}
