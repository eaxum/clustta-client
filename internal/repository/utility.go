package repository

import (
	"bytes"
	"clustta/internal/chunk_service"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	kzstd "github.com/klauspost/compress/zstd"

	"github.com/DataDog/zstd"
	"github.com/jmoiron/sqlx"
	"github.com/jotfs/fastcdc-go"

	_ "github.com/mattn/go-sqlite3"
)

func StoreFileChunks(tx *sqlx.Tx, filePath string, callback func(int, int, string, string)) (string, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}
	fileSize := int(fileInfo.Size())
	processed_size := 0

	seenChunks := make(map[string]bool)

	kiB := 1024
	miB := 1024 * kiB

	opts := fastcdc.Options{
		MinSize:     512 * kiB,
		AverageSize: 1 * miB,
		MaxSize:     8 * miB,
	}

	chunkSequence := make([]string, 0)

	chunker, _ := fastcdc.NewChunker(file, opts)
	for {
		chunk, err := chunker.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		data := chunk.Data
		sha256Hash := sha256.New()
		sha256Hash.Write(data)
		hash := hex.EncodeToString(sha256Hash.Sum(nil))

		if chunk_service.ChunkExists(hash, tx, seenChunks) {
			chunkSequence = append(chunkSequence, hash)
			processed_size += len(data)
			callback(processed_size, fileSize, "", "")
			continue
		}

		compressedData, err := zstd.CompressLevel(nil, data, 3)
		if err != nil {
			return "", err
		}

		// encoder, err := kzstd.NewWriter(nil, kzstd.WithEncoderLevel(3))
		// if err != nil {
		// 	return "", err
		// }
		// defer encoder.Close()
		// compressedData := encoder.EncodeAll(data, nil)

		size := len(compressedData)

		_, err = tx.Exec("INSERT INTO chunk (hash, data, size) VALUES (?, ?, ?)",
			hash,
			compressedData,
			size,
		)
		if err != nil {
			return "", err
		}
		seenChunks[hash] = true
		chunkSequence = append(chunkSequence, hash)

		chunkSize := len(data)
		processed_size += chunkSize
		callback(processed_size, fileSize, "", "")
	}
	chunkSequenceStr := strings.Join(chunkSequence, ",")
	return chunkSequenceStr, nil
}

func CheckMissingChunks(tx *sqlx.Tx, chunkHashes []string) ([]string, error) {
	if len(chunkHashes) == 0 {
		return nil, nil // No hashes to check, return an empty result
	}

	quotedChunkHashes := make([]string, len(chunkHashes))
	for i, hash := range chunkHashes {
		quotedChunkHashes[i] = fmt.Sprintf("\"%s\"", hash)
	}

	query := fmt.Sprintf(`SELECT hash FROM chunk WHERE hash IN (%s)`, strings.Join(quotedChunkHashes, ","))

	rows, err := tx.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	existingHashes := make(map[string]struct{})
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		existingHashes[hash] = struct{}{}
	}

	var missingHashes []string
	for _, hash := range chunkHashes {
		if _, found := existingHashes[hash]; !found {
			missingHashes = append(missingHashes, hash)
		}
	}

	return missingHashes, nil
}

func RebuildFile(tx *sqlx.Tx, chunks string, filePath string, timeModified int64, callback func(int, int, string, string)) error {
	chunkHashes := strings.Split(chunks, ",")
	missingChunks, err := CheckMissingChunks(tx, chunkHashes)
	if err != nil {
		return err
	}
	if len(missingChunks) > 0 {
		return errors.New("missing some chunks. please sync")
	}

	buffer := bytes.Buffer{}
	bufferLimit := 100 * 1024 * 1024

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	totalChunks := len(chunkHashes)
	if timeModified != 0 {
		totalChunks++
	}
	processedChunks := 0
	for _, chunkHash := range chunkHashes {
		var data []byte
		err = tx.Get(&data, "SELECT data FROM chunk WHERE hash = ?", chunkHash)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			return nil
		}
		buffer.Write(data)
		if buffer.Len() > bufferLimit {
			// decompressor := zstd.NewReader(&buffer)
			// io.Copy(file, decompressor)
			// buffer.Reset()
			decompressor, err := kzstd.NewReader(&buffer)
			if err != nil {
				return err
			}
			decompressor.WriteTo(file)
			buffer.Reset()
		}
		processedChunks++
		callback(processedChunks, totalChunks, "", "")
	}
	if buffer.Len() > 0 {
		// decompressor := zstd.NewReader(&buffer)
		// io.Copy(file, decompressor)
		// buffer.Reset()
		decompressor, err := kzstd.NewReader(&buffer)
		if err != nil {
			return err
		}
		decompressor.WriteTo(file)
		buffer.Reset()
	}
	file.Close()
	if timeModified != 0 {
		timeModified := time.Unix(timeModified, 0)
		os.Chtimes(filePath, timeModified, timeModified)
		processedChunks++
		callback(processedChunks, totalChunks, "", "")
	}
	return nil
}

func ProgressCallback(current int, total int) {
	percentage := float64(current) / float64(total) * 100
	// fmt.Printf("%.2f\n", percentage)
	os.Stdout.WriteString(fmt.Sprintf("%.2f\n", percentage))
}
