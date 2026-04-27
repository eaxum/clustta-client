package custom_thumbnail

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"

	kzstd "github.com/klauspost/compress/zstd"
)

type BlenderExtractor struct{}

func (b *BlenderExtractor) CanHandle(extension string) bool {
	return extension == ".blend"
}

func (b *BlenderExtractor) GetName() string {
	return "BlenderExtractor"
}

func (b *BlenderExtractor) ExtractThumbnail(filePath string) ([]byte, error) {

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Check for Zstandard compression (Blender 3.0+)
	if len(fileData) >= 4 && fileData[0] == 0x28 && fileData[1] == 0xB5 && fileData[2] == 0x2F && fileData[3] == 0xFD {
		decoder, err := kzstd.NewReader(nil)
		if err != nil {
			return nil, nil
		}
		defer decoder.Close()
		decompressed, err := decoder.DecodeAll(fileData, make([]byte, 0, 256*1024*1024))
		if err != nil {
			return nil, nil
		}
		fileData = decompressed
	} else if len(fileData) >= 2 && fileData[0] == 0x1f && fileData[1] == 0x8b {
		// Check for gzip compression (older Blender versions)
		gzReader, err := gzip.NewReader(bytes.NewReader(fileData))
		if err != nil {
			return nil, nil
		}
		defer gzReader.Close()

		fileData, err = io.ReadAll(gzReader)
		if err != nil {
			return nil, nil
		}
	}

	// Create a bytes reader for seeking
	reader := bytes.NewReader(fileData)

	// Read Blender file header (12 bytes)
	header := make([]byte, 12)
	if _, err := io.ReadFull(reader, header); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Verify it's a Blender file
	if !bytes.HasPrefix(header, []byte("BLENDER")) {
		return nil, fmt.Errorf("not a valid Blender file")
	}

	// Determine pointer size and endianness from header
	pointerSize := 4
	if header[7] == '-' {
		pointerSize = 8
	}

	var byteOrder binary.ByteOrder = binary.LittleEndian
	if header[8] == 'V' {
		byteOrder = binary.BigEndian
	}

	// Search for the thumbnail block
	thumbnail, err := b.findThumbnailBlock(reader, byteOrder, pointerSize)
	if err != nil {
		return nil, err
	}

	return thumbnail, nil
}

func (b *BlenderExtractor) findThumbnailBlock(reader io.ReadSeeker, byteOrder binary.ByteOrder, pointerSize int) ([]byte, error) {
	// Blender files are structured as a series of blocks
	// The thumbnail is in a block with code "TEST"

	_, err := reader.Seek(12, io.SeekStart)
	if err != nil {
		return nil, err
	}

	maxBlocks := 10000
	blockCount := 0

	var bestThumbnail []byte
	var bestWidth, bestHeight int

	for blockCount < maxBlocks {
		blockCount++

		blockCode := make([]byte, 4)
		n, err := reader.Read(blockCode)
		if err == io.EOF {
			break
		}
		if err != nil || n != 4 {
			return nil, fmt.Errorf("failed to read block code at block %d", blockCount)
		}

		blockCodeStr := string(blockCode)

		if blockCodeStr == "ENDB" {
			break
		}

		sizeBytes := make([]byte, 4)
		if _, err := io.ReadFull(reader, sizeBytes); err != nil {
			return nil, fmt.Errorf("failed to read block size: %w", err)
		}
		blockSize := byteOrder.Uint32(sizeBytes)

		if _, err := reader.Seek(int64(pointerSize), io.SeekCurrent); err != nil {
			return nil, err
		}

		sdnaBytes := make([]byte, 4)
		if _, err := io.ReadFull(reader, sdnaBytes); err != nil {
			return nil, err
		}

		countBytes := make([]byte, 4)
		if _, err := io.ReadFull(reader, countBytes); err != nil {
			return nil, err
		}

		headerSize := 4 + 4 + pointerSize + 4 + 4
		dataSize := int64(blockSize) - int64(headerSize)

		if blockCodeStr == "TEST" {
			if dataSize <= 0 {
				continue
			}

			thumbnailData := make([]byte, dataSize)
			if _, err := io.ReadFull(reader, thumbnailData); err != nil {
				continue
			}

			if len(thumbnailData) < 8 {
				continue
			}

			width := int(byteOrder.Uint32(thumbnailData[0:4]))
			height := int(byteOrder.Uint32(thumbnailData[4:8]))

			// Keep the largest thumbnail found
			if width*height > bestWidth*bestHeight {
				bestWidth = width
				bestHeight = height

				rawPixelData := thumbnailData[8:]

				// Check if already PNG format
				if bytes.HasPrefix(rawPixelData, []byte{0x89, 0x50, 0x4E, 0x47}) {
					bestThumbnail = rawPixelData
					continue
				}

				// Convert raw RGBA data to PNG
				expectedSize := width * height * 4
				actualSize := len(rawPixelData)

				if actualSize < expectedSize-100 {
					continue
				}

				pixelDataSize := actualSize
				if expectedSize < actualSize {
					pixelDataSize = expectedSize
				}

				pngData, err := b.convertRGBAToPNG(rawPixelData[:pixelDataSize], width, height)
				if err != nil {
					continue
				}

				bestThumbnail = pngData
			}

			continue
		}

		if dataSize > 0 {
			if _, err := reader.Seek(dataSize, io.SeekCurrent); err != nil {
				return nil, fmt.Errorf("failed to seek to next block: %w", err)
			}
		}
	}

	return bestThumbnail, nil
}

// convertRGBAToPNG converts raw RGBA pixel data to PNG format
// Blender stores pixels in bottom-to-top order (OpenGL style), so we flip vertically
func (b *BlenderExtractor) convertRGBAToPNG(rawData []byte, width, height int) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	stride := width * 4

	// Copy and flip vertically
	for y := 0; y < height; y++ {
		srcOffset := y * stride
		dstOffset := (height - 1 - y) * stride

		if srcOffset+stride <= len(rawData) {
			copy(img.Pix[dstOffset:dstOffset+stride], rawData[srcOffset:srcOffset+stride])
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("PNG encoding failed: %w", err)
	}

	return buf.Bytes(), nil
}
