package chunk_service

// def get_non_existing_chunks(chunks: str | list) -> list[str]:
//     """
//     Check if the chunks are valid.
//     """
//     if isinstance(chunks, str):
//         chunks: list = chunks.split(",")
//     non_existent_chunks = []
//     environment = lmdb.open(chunks_db.as_posix(), readonly=True)
//     with environment.begin() as txn:
//         for chunk in chunks:
//             if chunk in non_existent_chunks:
//                 continue
//             if not txn.get(chunk.encode()):
//                 non_existent_chunks.append(chunk)
//     environment.close()
//     return non_existent_chunks

// def write_chunks(chunks: bytes) -> list[str]:
//     """
//     Write chunks to the database. Return a list of failed chunks.
//     Chunks are encoded in TLV format.
//     The tag is a 32-byte hash, the length is a 3-byte integer, and the value is the binary data.
//     """
//     environment = lmdb.open(chunks_db.as_posix())
//     failed_chunks = []
//     with environment.begin(write=True) as txn:
//         while chunks:
//             tag = chunks[:32]
//             length = int.from_bytes(chunks[32:35], "big")
//             value = chunks[35 : 35 + length]
//             chunks = chunks[35 + length :]
//             if hashlib.sha256(value).digest() != tag:
//                 failed_chunks.append(tag.hex())
//                 continue
//             txn.put(tag, value)
//     environment.close()
//     return failed_chunks

import (
	"bytes"
	"clustta/internal/constants"
	"clustta/internal/utils"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/jmoiron/sqlx"
	kzstd "github.com/klauspost/compress/zstd"
	_ "github.com/mattn/go-sqlite3"
)

type Chunk struct {
	Hash string `db:"hash" json:"hash"`
	Data []byte `db:"data" json:"data"`
	Size int    `db:"size" json:"size"`
}

type ChunkInfo struct {
	Hash string `json:"hash"`
	Size int    `json:"size"`
}

func GetChunkInfo(tx *sqlx.Tx, chunkHash string) (ChunkInfo, error) {
	var chunkInfo ChunkInfo
	err := tx.Get(&chunkInfo, "SELECT hash, size FROM chunk WHERE hash = ?", chunkHash)
	if err != nil {
		return chunkInfo, err
	}
	return chunkInfo, nil
}

func GetChunksInfo(tx *sqlx.Tx, chunkHashes []string) ([]ChunkInfo, error) {
	var chunkInfos []ChunkInfo
	for _, chunkHash := range chunkHashes {
		chunkInfo, err := GetChunkInfo(tx, chunkHash)
		if err != nil {
			return chunkInfos, err
		}
		chunkInfos = append(chunkInfos, chunkInfo)
	}
	return chunkInfos, nil
}

func GetNonExistingChunks(tx *sqlx.Tx, chunks []string) ([]string, error) {
	var nonExistentChunks []string
	seenChunks := make(map[string]bool)
	for _, chunk := range chunks {
		if ChunkExists(chunk, tx, seenChunks) {
			continue
		}
		nonExistentChunks = append(nonExistentChunks, chunk)
	}

	return nonExistentChunks, nil
}

func WriteChunkData(projectPath string, chunkData Chunk) error {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO chunk (hash, data, size) VALUES (?, ?, ?)",
		chunkData.Hash,
		chunkData.Data,
		chunkData.Size,
	)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func WriteChunks(projectPath string, chunks []byte) ([]string, error) {
	var failedChunks []string
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return failedChunks, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return failedChunks, err
	}
	defer tx.Rollback()

	decoder, err := kzstd.NewReader(nil)
	if err != nil {
		return nil, err
	}
	defer decoder.Close()
	seenChunks := make(map[string]bool)
	for len(chunks) > 0 {
		if len(chunks) < 36 { // Adjusted to 36 to account for 4 bytes of length
			break // Not enough data for a complete chunk
		}

		tag := chunks[:32]
		length := binary.BigEndian.Uint32(chunks[32:36]) // Use the full 4 bytes for length

		if len(chunks) < 36+int(length) {
			break // Not enough data for the value
		}

		compressedValue := chunks[36 : 36+length]
		chunks = chunks[36+length:]

		if ChunkExists(hex.EncodeToString(tag), tx, seenChunks) {
			continue
		}

		decompressedValue, err := decoder.DecodeAll(compressedValue, nil)
		if err != nil {
			failedChunks = append(failedChunks, hex.EncodeToString(tag))
			continue
		}

		hash := sha256.Sum256(decompressedValue)
		if !bytes.Equal(hash[:], tag) {
			failedChunks = append(failedChunks, hex.EncodeToString(tag))
			continue
		}

		size := len(compressedValue)
		_, err = tx.Exec("INSERT INTO chunk (hash, data, size) VALUES (?, ?, ?)",
			hex.EncodeToString(tag),
			compressedValue,
			size,
		)
		if err != nil {
			return failedChunks, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return failedChunks, nil
}

func DecodeChunk(data []byte) (hash string, compressedData []byte, bytesRead int, err error) {
	if len(data) < 36 { // Adjusted to 36 to account for 4 bytes of length
		return "", nil, 0, fmt.Errorf("not enough data for a complete chunk")
	}

	// Extract the hash
	hashBytes := data[:32]
	hash = hex.EncodeToString(hashBytes)

	// Extract the length
	length := binary.BigEndian.Uint32(data[32:36]) // Use the full 4 bytes for length

	if len(data) < 36+int(length) {
		return "", nil, 0, fmt.Errorf("not enough data for the complete chunk value")
	}

	// Extract the compressed data
	compressedData = data[36 : 36+length]

	bytesRead = 36 + int(length)
	return hash, compressedData, bytesRead, nil
}

func EncodeChunks(chunks []Chunk) ([]byte, error) {
	var buffer bytes.Buffer

	for _, chunk := range chunks {
		hashBytes, err := hex.DecodeString(chunk.Hash)
		if err != nil {
			return nil, fmt.Errorf("invalid hash string: %v", err)
		}
		if len(hashBytes) != 32 {
			return nil, fmt.Errorf("invalid hash length: expected 32 bytes, got %d", len(chunk.Hash))
		}

		// Write the hash (tag)
		buffer.Write(hashBytes)

		// Write the length (3 bytes for length up to 16MB)
		length := len(chunk.Data)
		if length > 16777215 { // 2^24 - 1, max value for 3 bytes
			return nil, fmt.Errorf("chunk size exceeds 16MB limit")
		}
		lengthBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(lengthBytes, uint32(length))
		buffer.Write(lengthBytes)

		// Write the compressed chunk data
		buffer.Write(chunk.Data)
	}

	return buffer.Bytes(), nil
}

func EncodeChunk(chunk Chunk) ([]byte, error) {
	var buffer bytes.Buffer

	// Decode the hash string into bytes
	hashBytes, err := hex.DecodeString(chunk.Hash)
	if err != nil {
		return nil, fmt.Errorf("invalid hash string: %v", err)
	}
	if len(hashBytes) != 32 {
		return nil, fmt.Errorf("invalid hash length: expected 32 bytes, got %d", len(hashBytes))
	}

	// Write the hash (tag)
	buffer.Write(hashBytes)

	// Write the length (3 bytes for length up to 16MB)
	length := len(chunk.Data)
	if length > 16777215 { // 2^24 - 1, max value for 3 bytes
		return nil, fmt.Errorf("chunk size exceeds 16MB limit")
	}
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, uint32(length))
	buffer.Write(lengthBytes) // Write only the last 3 bytes

	// Write the compressed chunk data
	buffer.Write(chunk.Data)

	return buffer.Bytes(), nil
}

func PullChunks(ctx context.Context, projectPath, remoteUrl string, chunkInfos []ChunkInfo, callback func(int, int, string, string)) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	dataUrl := remoteUrl + "/chunks"
	client := &http.Client{}

	totalChunksSize := 0
	for _, chunkInfo := range chunkInfos {
		totalChunksSize += chunkInfo.Size
	}
	processedChunks := 0

	if utils.IsValidURL(remoteUrl) {
		for _, chunkInfo := range chunkInfos {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			data := map[string]any{
				"chunks": []string{chunkInfo.Hash},
			}
			jsonData, err := json.Marshal(data)
			if err != nil {
				return err
			}

			req, err := http.NewRequest("GET", dataUrl, bytes.NewBuffer(jsonData))
			if err != nil {
				return err
			}
			req.Header.Set("Clustta-Agent", constants.USER_AGENT)
			response, err := client.Do(req)
			if err != nil {
				return err
			}
			defer response.Body.Close()

			responseCode := response.StatusCode
			if responseCode == 200 {
				body, err := io.ReadAll(response.Body)
				if err != nil {
					return fmt.Errorf("error reading response body: %s", err.Error())
				}
				_, err = WriteChunks(projectPath, body)
				if err != nil {
					return fmt.Errorf("error writing chunks: %s", err.Error())
				}
				processedChunks += chunkInfo.Size
				message := fmt.Sprintf("Pulling data %s/%s", utils.BytesToHumanReadable(processedChunks), utils.BytesToHumanReadable(totalChunksSize))
				callback(processedChunks, totalChunksSize, message, "")
			} else if responseCode == 400 {
				body, err := io.ReadAll(response.Body)
				if err != nil {
					return err
				}
				return errors.New(string(body))
			} else {

				return errors.New("unknown error while fetching data")
			}
		}
	} else if utils.FileExists(remoteUrl) {
		dbConn, err := utils.OpenDb(remoteUrl)
		if err != nil {
			return err
		}
		defer dbConn.Close()
		remoteTx, err := dbConn.Beginx()
		if err != nil {
			return err
		}
		defer remoteTx.Rollback()
		for _, chunkInfo := range chunkInfos {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			var chunkData Chunk
			err = remoteTx.Get(&chunkData, "SELECT * FROM chunk WHERE hash = ?", chunkInfo.Hash)
			if err != nil {
				return err
			}
			err = WriteChunkData(projectPath, chunkData)
			if err != nil {
				return err
			}
			processedChunks += chunkInfo.Size
			message := fmt.Sprintf("Pulling data %s/%s", utils.BytesToHumanReadable(processedChunks), utils.BytesToHumanReadable(totalChunksSize))
			callback(processedChunks, totalChunksSize, message, "")
		}
	}
	return nil
}

func processTLVStream(ctx context.Context, projectPath string, r io.Reader, downloadedSize, totalSize int, chunksCountMap map[string]int, callback func(int, int, string, string)) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	decoder, err := kzstd.NewReader(nil)
	if err != nil {
		return err
	}
	defer decoder.Close()
	seenChunks := make(map[string]bool)

	savedSize := downloadedSize

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// Read the TLV tag (hash, 32 bytes)
		tag := make([]byte, 32)
		_, err = io.ReadFull(r, tag)
		if err == io.EOF {
			break // End of stream
		} else if err != nil {
			return fmt.Errorf("error reading tag: %w", err)
		}

		// Read the TLV length (4 bytes, uint32)
		lengthBuf := make([]byte, 4)
		_, err = io.ReadFull(r, lengthBuf)
		if err != nil {
			return fmt.Errorf("error reading length: %w", err)
		}
		length := binary.BigEndian.Uint32(lengthBuf) // Use the full 4 bytes for length

		// Validate the length
		if length == 0 || length > 16777215 { // 3-byte max value
			return fmt.Errorf("invalid length: %d", length)
		}

		// Read the TLV value (chunk data)
		compressedValue := make([]byte, length)
		_, err = io.ReadFull(r, compressedValue)
		if err != nil {
			return fmt.Errorf("error reading value: %w", err)
		}

		tx, err := dbConn.Beginx()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		if ChunkExists(hex.EncodeToString(tag), tx, seenChunks) {
			tx.Rollback()
			continue
		}

		// Validate Zstandard magic number
		// if len(compressedValue) < 4 || !bytes.Equal(compressedValue[:4], []byte{0x28, 0xB5, 0x2F, 0xFD}) {
		// 	return fmt.Errorf("invalid input: magic number mismatch")
		// }

		decompressedValue, err := decoder.DecodeAll(compressedValue, nil)
		if err != nil {
			return fmt.Errorf("error decoding chunk: %w", err)
		}

		hash := sha256.Sum256(decompressedValue)
		if !bytes.Equal(hash[:], tag) {
			return errors.New("invalid chunk data")
		}
		compressedSize := len(compressedValue)
		size := len(decompressedValue)
		// Store chunk in SQLite
		_, err = tx.Exec("INSERT INTO chunk (hash, data, size) VALUES (?, ?, ?)",
			hex.EncodeToString(tag),
			compressedValue,
			size,
		)
		if err != nil {
			return fmt.Errorf("error inserting into DB: %w", err)
		}
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("error writing data: %w", err)
		}

		downloadedSize += size * chunksCountMap[hex.EncodeToString(tag)]
		savedSize += size - compressedSize
		if chunksCountMap[hex.EncodeToString(tag)] > 1 {
			savedSize += size * (chunksCountMap[hex.EncodeToString(tag)] - 1)
		}
		message := fmt.Sprintf("Pulling data %s/%s", utils.BytesToHumanReadable(downloadedSize), utils.BytesToHumanReadable(totalSize))
		extraMessage := ""

		dataSavedPercentage := 0.0
		if totalSize > 0 {
			dataSavedPercentage = (float64(savedSize) / float64(downloadedSize)) * 100
		}
		if savedSize > 0 {
			extraMessage = fmt.Sprintf("Data saved: %s (%.2f%%)", utils.BytesToHumanReadable(savedSize), dataSavedPercentage)
		}

		callback(downloadedSize, totalSize, message, extraMessage)
	}

	return nil
}

func ProcessDownloadedChunksProgress(ctx context.Context, projectPath, remoteUrl string, missingChunkHashes []string, allChunkHashes []string, totalSize int, callback func(int, int, string, string)) (int, int, map[string]int, error) {
	if ctx.Err() != nil {
		return 0, 0, map[string]int{}, ctx.Err()
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return 0, 0, map[string]int{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return 0, 0, map[string]int{}, err
	}
	defer tx.Rollback()

	downloadedSize := 0

	missingChunksMap := map[string]bool{}
	for _, hash := range missingChunkHashes {
		missingChunksMap[hash] = true
	}

	chunksCountMap := map[string]int{}
	for _, hash := range allChunkHashes {
		chunksCountMap[hash] += 1
	}

	for hash, count := range chunksCountMap {
		if missingChunksMap[hash] {
			continue
		}
		var size int
		err := tx.Get(&size, "SELECT size FROM chunk WHERE hash = ?", hash)
		if err != nil {
			return downloadedSize, totalSize, chunksCountMap, err
		}
		downloadedSize += size * count

		message := fmt.Sprintf("Pulling data %s/%s", utils.BytesToHumanReadable(downloadedSize), utils.BytesToHumanReadable(totalSize))
		extraMessage := ""

		// dataSavedPercentage := 0.0
		// if totalSize > 0 {
		// 	dataSavedPercentage = (float64(downloadedSize) / float64(totalSize)) * 100
		// }
		if downloadedSize > 0 {
			extraMessage = fmt.Sprintf("Data saved: %s (%.2f%%)", utils.BytesToHumanReadable(downloadedSize), 100.00)
		}
		callback(downloadedSize, totalSize, message, extraMessage)
	}

	return downloadedSize, totalSize, chunksCountMap, nil
}

func PullStreamChunks(ctx context.Context, projectPath, remoteUrl string, missingChunkHashes []string, allChunkHashes []string, totalSize int, callback func(int, int, string, string)) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	downloadedSize, _, chunksCountMap, err := ProcessDownloadedChunksProgress(ctx, projectPath, remoteUrl, missingChunkHashes, allChunkHashes, totalSize, callback)
	if err != nil {
		return err
	}

	dataUrl := remoteUrl + "/stream-chunks"
	client := &http.Client{}

	if utils.IsValidURL(remoteUrl) {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		data := map[string]any{
			"chunks": missingChunkHashes,
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}

		req, err := http.NewRequest("GET", dataUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		req.Header.Set("Clustta-Agent", constants.USER_AGENT)
		response, err := client.Do(req)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		responseCode := response.StatusCode
		if responseCode == 200 {
			// Process the TLV stream
			err = processTLVStream(ctx, projectPath, response.Body, downloadedSize, totalSize, chunksCountMap, callback)
			if err != nil {
				return fmt.Errorf("error processing stream: %s", err.Error())
			}
		} else if responseCode == 400 {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return err
			}
			return errors.New(string(body))
		} else {
			return errors.New("unknown error while fetching data")
		}

	} else {
		return errors.New("invalid url")
	}
	return nil
}

func PushChunks(tx *sqlx.Tx, remoteUrl string, userId string, chunkInfos []ChunkInfo, callback func(int, int, string, string)) error {
	dataUrl := remoteUrl + "/chunks"
	client := &http.Client{}

	totalChunksSize := 0
	for _, chunkInfo := range chunkInfos {
		totalChunksSize += chunkInfo.Size
	}
	processedChunks := 0

	if utils.IsValidURL(remoteUrl) {
		for _, chunkInfo := range chunkInfos {
			var chunkData []byte
			err := tx.Get(&chunkData, "SELECT data FROM chunk WHERE hash = ?", chunkInfo.Hash)
			if err != nil {
				return err
			}
			chunk := Chunk{
				Hash: chunkInfo.Hash,
				Data: chunkData,
			}
			encodedChunk, err := EncodeChunks([]Chunk{chunk})
			if err != nil {
				return err
			}

			req, err := http.NewRequest("POST", dataUrl, bytes.NewBuffer(encodedChunk))
			if err != nil {
				return err
			}
			req.Header.Set("Clustta-Agent", constants.USER_AGENT)

			response, err := client.Do(req)
			if err != nil {
				return err
			}
			defer response.Body.Close()
			responseCode := response.StatusCode
			if responseCode == 200 {
				processedChunks += chunkInfo.Size
				message := fmt.Sprintf("Pushing Data %s/%s", utils.BytesToHumanReadable(processedChunks), utils.BytesToHumanReadable(totalChunksSize))
				callback(processedChunks, totalChunksSize, message, "")
			} else if responseCode == 400 {
				body, err := io.ReadAll(response.Body)
				if err != nil {
					return err
				}
				return errors.New(string(body))
			} else {
				return errors.New("unknown error while pushing chunks")
			}
		}
	} else if utils.FileExists(remoteUrl) {
		dbConn, err := utils.OpenDb(remoteUrl)
		if err != nil {
			return err
		}
		defer dbConn.Close()
		remoteTx, err := dbConn.Beginx()
		if err != nil {
			return err
		}

		for _, chunkInfo := range chunkInfos {
			var chunkData Chunk
			err = tx.Get(&chunkData, "SELECT * FROM chunk WHERE hash = ?", chunkInfo.Hash)
			if err != nil {
				return err
			}
			_, err = remoteTx.Exec("INSERT INTO chunk (hash, data, size) VALUES (?, ?, ?)",
				chunkData.Hash,
				chunkData.Data,
				chunkData.Size,
			)
			if err != nil {
				return err
			}
			processedChunks += chunkInfo.Size
			message := fmt.Sprintf("Pushing Data %s/%s", utils.BytesToHumanReadable(processedChunks), utils.BytesToHumanReadable(totalChunksSize))
			callback(processedChunks, totalChunksSize, message, "")
		}
		err = remoteTx.Commit()
		if err != nil {
			remoteTx.Rollback()
			return err
		}
	}

	return nil
}

func PushChunksBatch(tx *sqlx.Tx, remoteUrl string, userId string, chunkInfos []ChunkInfo, callback func(int, int, string, string)) error {
	// const batchSizeLimit = 1 << 20 // 1 MB
	const batchSizeLimit = 512 * 1024 // 512 KB

	dataUrl := remoteUrl + "/chunks"
	client := &http.Client{}

	totalChunksSize := 0
	for _, chunkInfo := range chunkInfos {
		totalChunksSize += chunkInfo.Size
	}
	processedChunks := 0

	if utils.IsValidURL(remoteUrl) {
		var currentBatch []Chunk
		currentBatchSize := 0

		pushBatch := func(batch []Chunk) error {
			if len(batch) == 0 {
				return nil
			}
			encodedChunk, err := EncodeChunks(batch)
			if err != nil {
				return err
			}
			req, err := http.NewRequest("POST", dataUrl, bytes.NewBuffer(encodedChunk))
			if err != nil {
				return err
			}
			req.Header.Set("Clustta-Agent", constants.USER_AGENT)
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode == 200 {
				for _, chunk := range batch {
					processedChunks += chunk.Size
				}
				message := fmt.Sprintf("Pushing Data %s/%s", utils.BytesToHumanReadable(processedChunks), utils.BytesToHumanReadable(totalChunksSize))
				callback(processedChunks, totalChunksSize, message, "")
			} else if resp.StatusCode == 400 {
				body, _ := io.ReadAll(resp.Body)
				return errors.New(string(body))
			} else {
				return fmt.Errorf("unknown error while pushing chunks, status: %d", resp.StatusCode)
			}
			return nil
		}

		for _, chunkInfo := range chunkInfos {
			var chunkData []byte
			err := tx.Get(&chunkData, "SELECT data FROM chunk WHERE hash = ?", chunkInfo.Hash)
			if err != nil {
				return err
			}
			chunk := Chunk{
				Hash: chunkInfo.Hash,
				Data: chunkData,
				Size: chunkInfo.Size,
			}
			currentBatch = append(currentBatch, chunk)
			currentBatchSize += chunkInfo.Size

			if currentBatchSize >= batchSizeLimit {
				if err := pushBatch(currentBatch); err != nil {
					return err
				}
				currentBatch = nil
				currentBatchSize = 0
			}
		}
		// push any remaining chunks
		if len(currentBatch) > 0 {
			if err := pushBatch(currentBatch); err != nil {
				return err
			}
		}
	} else if utils.FileExists(remoteUrl) {
		dbConn, err := utils.OpenDb(remoteUrl)
		if err != nil {
			return err
		}
		defer dbConn.Close()

		remoteTx, err := dbConn.Beginx()
		if err != nil {
			return err
		}
		defer func() {
			if p := recover(); p != nil {
				remoteTx.Rollback()
				panic(p)
			}
		}()

		for _, chunkInfo := range chunkInfos {
			var chunkData Chunk
			err = tx.Get(&chunkData, "SELECT * FROM chunk WHERE hash = ?", chunkInfo.Hash)
			if err != nil {
				return err
			}
			_, err = remoteTx.Exec("INSERT INTO chunk (hash, data, size) VALUES (?, ?, ?)",
				chunkData.Hash,
				chunkData.Data,
				chunkData.Size,
			)
			if err != nil {
				remoteTx.Rollback()
				return err
			}
			processedChunks += chunkInfo.Size
			message := fmt.Sprintf("Pushing Data %s/%s", utils.BytesToHumanReadable(processedChunks), utils.BytesToHumanReadable(totalChunksSize))
			callback(processedChunks, totalChunksSize, message, "")
		}
		if err = remoteTx.Commit(); err != nil {
			remoteTx.Rollback()
			return err
		}
	}

	return nil
}

func ChunkExists(chunkHash string, tx *sqlx.Tx, seenChunks map[string]bool) bool {
	if _, ok := seenChunks[chunkHash]; ok {
		return true
	}
	var hash string
	tx.Get(&hash, "SELECT hash FROM chunk WHERE hash = ?", chunkHash)
	return hash != ""
}
