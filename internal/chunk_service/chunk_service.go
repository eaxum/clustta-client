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
	"clustta/internal/auth_service"
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
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	kzstd "github.com/klauspost/compress/zstd"
	_ "github.com/mattn/go-sqlite3"
)

// IsPersonalRemote returns true if the remote URL points to a personal project on the Clustta server.
func IsPersonalRemote(remoteUrl string) bool {
	return strings.Contains(remoteUrl, "/user/")
}

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
			auth_service.AttachBearerToken(req)
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
				message := fmt.Sprintf("Receiving %s/%s", utils.BytesToHumanReadable(processedChunks), utils.BytesToHumanReadable(totalChunksSize))
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
			message := fmt.Sprintf("Receiving %s/%s", utils.BytesToHumanReadable(processedChunks), utils.BytesToHumanReadable(totalChunksSize))
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
		message := fmt.Sprintf("Receiving %s/%s", utils.BytesToHumanReadable(downloadedSize), utils.BytesToHumanReadable(totalSize))
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

		message := fmt.Sprintf("Receiving %s/%s", utils.BytesToHumanReadable(downloadedSize), utils.BytesToHumanReadable(totalSize))
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

	// Use presigned URLs for personal projects (server-hosted on R2).
	if IsPersonalRemote(remoteUrl) {
		return PullChunksPresigned(ctx, projectPath, remoteUrl, missingChunkHashes, allChunkHashes, totalSize, callback)
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
		auth_service.AttachBearerToken(req)
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

// PullChunksPresigned fetches presigned R2 URLs from the server and downloads
// chunks directly from R2 with bounded concurrency.
func PullChunksPresigned(ctx context.Context, projectPath, remoteUrl string, missingChunkHashes []string, allChunkHashes []string, totalSize int, callback func(int, int, string, string)) error {
	downloadedSize, _, chunksCountMap, err := ProcessDownloadedChunksProgress(ctx, projectPath, remoteUrl, missingChunkHashes, allChunkHashes, totalSize, callback)
	if err != nil {
		return err
	}

	urlsEndpoint := remoteUrl + "/chunk-urls"
	data := map[string]any{"chunks": missingChunkHashes}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", urlsEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	auth_service.AttachBearerToken(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get chunk URLs: %s", string(body))
	}

	var urlsResponse struct {
		URLs map[string]string `json:"urls"`
	}
	err = json.NewDecoder(resp.Body).Decode(&urlsResponse)
	if err != nil {
		return fmt.Errorf("failed to decode chunk URLs: %w", err)
	}

	type chunkResult struct {
		hash string
		data []byte
		err  error
	}

	const maxConcurrency = 8
	resultsCh := make(chan chunkResult, len(urlsResponse.URLs))
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for hash, url := range urlsResponse.URLs {
		wg.Add(1)
		go func(h, u string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if ctx.Err() != nil {
				resultsCh <- chunkResult{hash: h, err: ctx.Err()}
				return
			}

			dlReq, err := http.NewRequestWithContext(ctx, "GET", u, nil)
			if err != nil {
				resultsCh <- chunkResult{hash: h, err: err}
				return
			}
			dlResp, err := http.DefaultClient.Do(dlReq)
			if err != nil {
				resultsCh <- chunkResult{hash: h, err: err}
				return
			}
			defer dlResp.Body.Close()

			if dlResp.StatusCode != 200 {
				resultsCh <- chunkResult{hash: h, err: fmt.Errorf("download failed for chunk %s: status %d", h, dlResp.StatusCode)}
				return
			}

			chunkData, err := io.ReadAll(dlResp.Body)
			if err != nil {
				resultsCh <- chunkResult{hash: h, err: err}
				return
			}
			resultsCh <- chunkResult{hash: h, data: chunkData}
		}(hash, url)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

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

	savedSize := 0
	for result := range resultsCh {
		if result.err != nil {
			return result.err
		}

		compressedData := result.data
		decompressedData, err := decoder.DecodeAll(compressedData, nil)
		if err != nil {
			return fmt.Errorf("decompression failed for chunk %s: %w", result.hash, err)
		}

		hashBytes := sha256.Sum256(decompressedData)
		expectedHash, err := hex.DecodeString(result.hash)
		if err != nil {
			return fmt.Errorf("invalid hash %s: %w", result.hash, err)
		}
		if !bytes.Equal(hashBytes[:], expectedHash) {
			return fmt.Errorf("hash mismatch for chunk %s", result.hash)
		}

		tx, err := dbConn.Beginx()
		if err != nil {
			return err
		}

		_, err = tx.Exec("INSERT INTO chunk (hash, data, size) VALUES (?, ?, ?)",
			result.hash, compressedData, len(decompressedData))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error inserting chunk: %w", err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("error writing chunk: %w", err)
		}

		compressedSize := len(compressedData)
		size := len(decompressedData)
		downloadedSize += size * chunksCountMap[result.hash]
		savedSize += size - compressedSize
		if chunksCountMap[result.hash] > 1 {
			savedSize += size * (chunksCountMap[result.hash] - 1)
		}

		message := fmt.Sprintf("Receiving %s/%s", utils.BytesToHumanReadable(downloadedSize), utils.BytesToHumanReadable(totalSize))
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

// PushChunksPresigned uploads chunks directly to R2 using presigned PUT URLs.
// Phase 1: request upload URLs from server. Phase 2: upload to R2. Phase 3: confirm with server.
func PushChunksPresigned(tx *sqlx.Tx, remoteUrl string, chunkInfos []ChunkInfo, callback func(int, int, string, string)) error {
	totalChunksSize := 0
	for _, ci := range chunkInfos {
		totalChunksSize += ci.Size
	}
	processedChunks := 0

	// Phase 1: Request presigned PUT URLs from the server
	type chunkEntry struct {
		Hash string `json:"hash"`
		Size int64  `json:"size"`
	}
	entries := make([]chunkEntry, len(chunkInfos))
	for i, ci := range chunkInfos {
		entries[i] = chunkEntry{Hash: ci.Hash, Size: int64(ci.Size)}
	}

	reqBody, err := json.Marshal(map[string]any{"chunks": entries})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", remoteUrl+"/chunk-upload-urls", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Clustta-Agent", constants.USER_AGENT)
	auth_service.AttachBearerToken(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get upload URLs: %s", string(body))
	}

	var urlsResponse struct {
		URLs     map[string]string `json:"urls"`
		Existing []string          `json:"existing"`
	}
	err = json.NewDecoder(resp.Body).Decode(&urlsResponse)
	if err != nil {
		return fmt.Errorf("failed to decode upload URLs: %w", err)
	}

	// Account for already-existing chunks in progress
	existingSet := make(map[string]bool, len(urlsResponse.Existing))
	for _, hash := range urlsResponse.Existing {
		existingSet[hash] = true
	}
	for _, ci := range chunkInfos {
		if existingSet[ci.Hash] {
			processedChunks += ci.Size
		}
	}
	if processedChunks > 0 {
		message := fmt.Sprintf("Sending %s/%s", utils.BytesToHumanReadable(processedChunks), utils.BytesToHumanReadable(totalChunksSize))
		callback(processedChunks, totalChunksSize, message, "")
	}

	if len(urlsResponse.URLs) == 0 {
		return nil
	}

	// Phase 2: Upload chunks directly to R2
	type uploadResult struct {
		hash string
		size int
		err  error
	}

	const maxConcurrency = 8
	resultsCh := make(chan uploadResult, len(urlsResponse.URLs))
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	// Build a size lookup for progress
	sizeMap := make(map[string]int, len(chunkInfos))
	for _, ci := range chunkInfos {
		sizeMap[ci.Hash] = ci.Size
	}

	for hash, url := range urlsResponse.URLs {
		var chunkData []byte
		err := tx.Get(&chunkData, "SELECT data FROM chunk WHERE hash = ?", hash)
		if err != nil {
			return fmt.Errorf("error reading chunk %s: %w", hash, err)
		}

		wg.Add(1)
		go func(h, u string, data []byte) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			uploadReq, err := http.NewRequest("PUT", u, bytes.NewReader(data))
			if err != nil {
				resultsCh <- uploadResult{hash: h, err: err}
				return
			}
			uploadReq.Header.Set("Content-Type", "application/octet-stream")
			uploadReq.ContentLength = int64(len(data))

			uploadResp, err := http.DefaultClient.Do(uploadReq)
			if err != nil {
				resultsCh <- uploadResult{hash: h, err: err}
				return
			}
			defer uploadResp.Body.Close()

			if uploadResp.StatusCode != 200 {
				resultsCh <- uploadResult{hash: h, err: fmt.Errorf("R2 upload failed for chunk %s: status %d", h, uploadResp.StatusCode)}
				return
			}

			resultsCh <- uploadResult{hash: h, size: sizeMap[h]}
		}(hash, url, chunkData)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var uploadedHashes []string
	for result := range resultsCh {
		if result.err != nil {
			return result.err
		}
		uploadedHashes = append(uploadedHashes, result.hash)
		processedChunks += result.size
		message := fmt.Sprintf("Sending %s/%s", utils.BytesToHumanReadable(processedChunks), utils.BytesToHumanReadable(totalChunksSize))
		callback(processedChunks, totalChunksSize, message, "")
	}

	// Phase 3: Confirm uploads with the server
	confirmBody, err := json.Marshal(map[string]any{"chunks": uploadedHashes})
	if err != nil {
		return err
	}

	confirmReq, err := http.NewRequest("POST", remoteUrl+"/chunk-upload-confirm", bytes.NewBuffer(confirmBody))
	if err != nil {
		return err
	}
	confirmReq.Header.Set("Content-Type", "application/json")
	confirmReq.Header.Set("Clustta-Agent", constants.USER_AGENT)
	auth_service.AttachBearerToken(confirmReq)

	confirmResp, err := client.Do(confirmReq)
	if err != nil {
		return err
	}
	defer confirmResp.Body.Close()

	if confirmResp.StatusCode != 200 {
		body, _ := io.ReadAll(confirmResp.Body)
		return fmt.Errorf("failed to confirm uploads: %s", string(body))
	}

	var confirmResponse struct {
		FailedChunks []string `json:"failed_chunks"`
	}
	err = json.NewDecoder(confirmResp.Body).Decode(&confirmResponse)
	if err != nil {
		return fmt.Errorf("failed to decode confirm response: %w", err)
	}

	if len(confirmResponse.FailedChunks) > 0 {
		return fmt.Errorf("server failed to confirm %d chunks", len(confirmResponse.FailedChunks))
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
			auth_service.AttachBearerToken(req)

			response, err := client.Do(req)
			if err != nil {
				return err
			}
			defer response.Body.Close()
			responseCode := response.StatusCode
			if responseCode == 200 {
				processedChunks += chunkInfo.Size
				message := fmt.Sprintf("Sending %s/%s", utils.BytesToHumanReadable(processedChunks), utils.BytesToHumanReadable(totalChunksSize))
				callback(processedChunks, totalChunksSize, message, "")
			} else if responseCode == 400 {
				body, err := io.ReadAll(response.Body)
				if err != nil {
					return err
				}
				return errors.New(string(body))
			} else {
				return errors.New("unknown error while sending chunks")
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
			message := fmt.Sprintf("Sending %s/%s", utils.BytesToHumanReadable(processedChunks), utils.BytesToHumanReadable(totalChunksSize))
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
	// Use presigned uploads for personal projects (server-hosted on R2).
	if IsPersonalRemote(remoteUrl) {
		return PushChunksPresigned(tx, remoteUrl, chunkInfos, callback)
	}

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
			auth_service.AttachBearerToken(req)
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode == 200 {
				for _, chunk := range batch {
					processedChunks += chunk.Size
				}
				message := fmt.Sprintf("Sending %s/%s", utils.BytesToHumanReadable(processedChunks), utils.BytesToHumanReadable(totalChunksSize))
				callback(processedChunks, totalChunksSize, message, "")
			} else if resp.StatusCode == 400 {
				body, _ := io.ReadAll(resp.Body)
				return errors.New(string(body))
			} else {
				return fmt.Errorf("unknown error while sending chunks, status: %d", resp.StatusCode)
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
			message := fmt.Sprintf("Sending %s/%s", utils.BytesToHumanReadable(processedChunks), utils.BytesToHumanReadable(totalChunksSize))
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
