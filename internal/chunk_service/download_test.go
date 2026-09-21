package chunk_service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"clustta/internal/compatibility"
	"clustta/internal/utils"
	"github.com/jmoiron/sqlx"
	"github.com/klauspost/compress/zstd"
)

func TestCancelledCloudDownloadReleasesSlotsBeforeReturning(t *testing.T) {
	path, _ := chunkDatabase(t)
	hash, _ := encodedChunk(t, "cancel me")
	started := make(chan struct{})
	var server *httptest.Server
	server = newCompatibleChunkServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/chunk-urls" {
			if err := json.NewEncoder(w).Encode(map[string]any{"urls": map[string]string{hash: server.URL + "/chunk"}}); err != nil {
				t.Error(err)
			}
			return
		}
		close(started)
		<-r.Context().Done()
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := make(chan error, 1)
	go func() {
		result <- PullChunksPresigned(ctx, path, server.URL+"/project", []string{hash}, []string{hash}, 0, func(int, int, string, string) {})
	}()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("download did not start")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancel result: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancelled download did not stop")
	}
	if len(downloadSlots) != 0 {
		t.Fatal("download returned before releasing network slots")
	}
}

func TestIncompleteStreamFails(t *testing.T) {
	path, _ := chunkDatabase(t)
	err := processTLVStream(context.Background(), path, bytes.NewReader(nil), 0, 1, map[string]int{"missing": 1}, func(int, int, string, string) {})
	if err == nil {
		t.Fatal("incomplete stream succeeded")
	}
}

func TestProgressIncludesCachedDuplicateAndCompressedBytes(t *testing.T) {
	t.Run("Studio", func(t *testing.T) { testByteProgress(t, false) })
	t.Run("Cloud", func(t *testing.T) { testByteProgress(t, true) })
}

func testByteProgress(t *testing.T, cloud bool) {
	path, db := chunkDatabase(t)
	cached := strings.Repeat("cached", 100)
	payload := strings.Repeat("download", 100)
	cachedHash, cachedData := encodedChunk(t, cached)
	hash, data := encodedChunk(t, payload)
	db.MustExec("INSERT INTO chunk VALUES (?, ?, ?)", cachedHash, cachedData, len(cached))
	stream := &bytes.Buffer{}
	appendStreamChunk(t, stream, hash, data)
	var server *httptest.Server
	server = newCompatibleChunkServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/chunk" {
			if _, err := w.Write(data); err != nil {
				t.Error(err)
			}
			return
		}
		var request struct {
			Chunks []string `json:"chunks"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || len(request.Chunks) != 1 || request.Chunks[0] != hash {
			t.Errorf("unexpected requested chunks: %v, %v", request.Chunks, err)
		}
		if strings.HasSuffix(r.URL.Path, "/chunk-urls") {
			if err := json.NewEncoder(w).Encode(map[string]any{"urls": map[string]string{hash: server.URL + "/chunk"}}); err != nil {
				t.Error(err)
			}
			return
		}
		if _, err := w.Write(stream.Bytes()); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	remoteURL := server.URL + "/project"
	if cloud {
		remoteURL = server.URL + "/studio/test/project"
	}
	total := len(cached) + 2*len(payload)
	current := 0
	extra := ""
	ctx := context.WithValue(context.Background(), downloadIdentityKey{}, "")
	err := PullStreamChunks(ctx, path, remoteURL, []string{hash}, []string{cachedHash, hash, hash}, total, func(received, required int, message, saved string) {
		if received < current || received > total || required != total || !strings.HasPrefix(message, "Receiving ") {
			t.Errorf("invalid byte progress: %d/%d %s", received, required, message)
		}
		current, extra = received, saved
	})
	if err != nil {
		t.Fatal(err)
	}
	wantSaved := total - len(data)
	want := fmt.Sprintf("Data saved: %s (%.2f%%)", utils.BytesToHumanReadable(wantSaved), float64(wantSaved)/float64(total)*100)
	if current != total || extra != want {
		t.Fatalf("got %d, %s; want %d, %s", current, extra, total, want)
	}
	err = PullStreamChunks(ctx, path, remoteURL, nil, []string{cachedHash, hash, hash}, total, func(received, required int, message, saved string) {
		current, extra = received, saved
	})
	if err != nil {
		t.Fatal(err)
	}
	want = fmt.Sprintf("Data saved: %s (100.00%%)", utils.BytesToHumanReadable(total))
	if current != total || extra != want {
		t.Fatalf("cached progress: %d, %s; want %d, %s", current, extra, total, want)
	}
}

func chunkDatabase(t *testing.T) (string, *sqlx.DB) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "project.clst")
	db, err := utils.OpenDb(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	db.MustExec("CREATE TABLE chunk (hash TEXT PRIMARY KEY, data BLOB, size INTEGER); CREATE TABLE config (name TEXT PRIMARY KEY, value TEXT); INSERT INTO config VALUES ('version', '2.2');")
	return path, db
}

func encodedChunk(t *testing.T, value string) (string, []byte) {
	t.Helper()
	encoder, err := zstd.NewWriter(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer encoder.Close()
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:]), encoder.EncodeAll([]byte(value), nil)
}

func appendStreamChunk(t *testing.T, stream *bytes.Buffer, hash string, data []byte) {
	t.Helper()
	tag, err := hex.DecodeString(hash)
	if err != nil {
		t.Fatal(err)
	}
	stream.Write(tag)
	if err := binary.Write(stream, binary.BigEndian, uint32(len(data))); err != nil {
		t.Fatal(err)
	}
	stream.Write(data)
}

func TestStreamReportsOnlyCommittedChunks(t *testing.T) {
	path, db := chunkDatabase(t)
	stream := &bytes.Buffer{}
	counts := map[string]int{}
	for index := 0; index < streamBatchSize+1; index++ {
		hash, data := encodedChunk(t, fmt.Sprintf("chunk-%d", index))
		counts[hash] = 1
		appendStreamChunk(t, stream, hash, data)
	}
	committed := 0
	callback := func(int, int, string, string) {
		if err := db.Get(&committed, "SELECT COUNT(*) FROM chunk"); err != nil {
			t.Error(err)
		}
		if committed == 0 {
			t.Error("progress reported before commit")
		}
	}
	if err := processTLVStream(context.Background(), path, stream, 0, 0, counts, callback); err != nil {
		t.Fatal(err)
	}
	if committed != len(counts) {
		t.Fatalf("committed %d chunks, want %d", committed, len(counts))
	}
}
func TestInvalidStreamDoesNotPublishUncommittedBatch(t *testing.T) {
	path, db := chunkDatabase(t)
	hash, data := encodedChunk(t, "valid")
	stream := &bytes.Buffer{}
	appendStreamChunk(t, stream, hash, data)
	stream.WriteByte(1)
	notifications := 0
	ctx := context.Background()
	if err := processTLVStream(ctx, path, stream, 0, 0, map[string]int{hash: 1}, func(int, int, string, string) { notifications++ }); err == nil {
		t.Fatal("truncated stream succeeded")
	}
	var count int
	if err := db.Get(&count, "SELECT COUNT(*) FROM chunk"); err != nil {
		t.Fatal(err)
	}
	if count != 0 || notifications != 0 {
		t.Fatal("published uncommitted batch")
	}
}

func TestCloudDownloadsIndividualChunksAndPublishesAfterCommit(t *testing.T) {
	path, db := chunkDatabase(t)
	hash, data := encodedChunk(t, "cloud chunk")
	var server *httptest.Server
	server = newCompatibleChunkServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/chunk-urls":
			var body struct {
				Chunks []string `json:"chunks"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Chunks) != 1 || body.Chunks[0] != hash {
				t.Errorf("unexpected URL request: %+v, %v", body, err)
			}
			if err := json.NewEncoder(w).Encode(map[string]any{"urls": map[string]string{hash: server.URL + "/chunk"}}); err != nil {
				t.Error(err)
			}
		case "/chunk":
			if _, err := w.Write(data); err != nil {
				t.Error(err)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	committed := false
	callback := func(int, int, string, string) {
		var present bool
		if err := db.Get(&present, "SELECT EXISTS(SELECT 1 FROM chunk WHERE hash = ?)", hash); err != nil || !present {
			t.Errorf("progress before commit: %v", err)
		}
		committed = present
	}
	if err := PullChunksPresigned(context.Background(), path, server.URL+"/project", []string{hash}, []string{hash}, 0, callback); err != nil {
		t.Fatal(err)
	}
	if !committed {
		t.Fatal("missing committed notification")
	}
}

func TestCloudConcurrencyLimitIsSharedAcrossActions(t *testing.T) {
	firstPath, _ := chunkDatabase(t)
	secondPath, _ := chunkDatabase(t)
	hashes := []string{}
	dataByHash := map[string][]byte{}
	for index := 0; index < maxDownloadRequests+1; index++ {
		hash, data := encodedChunk(t, fmt.Sprintf("parallel-%d", index))
		hashes = append(hashes, hash)
		dataByHash[hash] = data
	}
	var active, maximum atomic.Int32
	started := make(chan struct{}, len(hashes)*2)
	release := make(chan struct{})
	var server *httptest.Server
	server = newCompatibleChunkServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/chunk-urls" {
			urls := map[string]string{}
			for _, hash := range hashes {
				urls[hash] = server.URL + "/" + hash
			}
			if err := json.NewEncoder(w).Encode(map[string]any{"urls": urls}); err != nil {
				t.Error(err)
			}
			return
		}
		count := active.Add(1)
		defer active.Add(-1)
		for old := maximum.Load(); count > old && !maximum.CompareAndSwap(old, count); old = maximum.Load() {
		}
		started <- struct{}{}
		select {
		case <-release:
		case <-r.Context().Done():
			return
		}
		if _, err := w.Write(dataByHash[r.URL.Path[1:]]); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	results := make(chan error, 2)
	for _, path := range []string{firstPath, secondPath} {
		go func() {
			results <- PullChunksPresigned(ctx, path, server.URL+"/project", hashes, hashes, 0, func(int, int, string, string) {})
		}()
	}
	for index := 0; index < maxDownloadRequests; index++ {
		select {
		case <-started:
		case <-ctx.Done():
			t.Fatal("downloads did not start")
		}
	}
	close(release)
	for index := 0; index < 2; index++ {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	if maximum.Load() > maxDownloadRequests {
		t.Fatalf("parallel actions exceeded shared limit: %d", maximum.Load())
	}
}

func newCompatibleChunkServer(next http.Handler) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, prefix := range []string{"/studio/test/project", "/project"} {
			if r.URL.Path == prefix {
				json.NewEncoder(w).Encode(map[string]*compatibility.Contract{"compatibility": compatibility.Current(compatibility.Schema)})
				return
			}
			if strings.HasPrefix(r.URL.Path, prefix+"/") {
				w.Header().Set(compatibility.ProtocolHeader, compatibility.Protocol)
				w.Header().Set(compatibility.SchemaHeader, compatibility.Schema)
				w.Header().Set(compatibility.ProjectSchemaHeader, compatibility.Schema)
				r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
				break
			}
		}
		next.ServeHTTP(w, r)
	}))
}
