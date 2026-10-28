package projectindex

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"clustta/internal/auth_service"
	"clustta/internal/settings"
)

const indexVersion = 1

type Entry struct {
	ProjectID        string `json:"project_id"`
	ProjectName      string `json:"project_name"`
	StudioName       string `json:"studio_name"`
	DatabasePath     string `json:"database_path"`
	WorkingDirectory string `json:"working_directory"`
	UpdatedAt        string `json:"updated_at"`
}

type indexFile struct {
	Version          int              `json:"version"`
	BackfillComplete bool             `json:"backfill_complete"`
	Entries          map[string]Entry `json:"entries"`
}

var indexState = struct {
	sync.RWMutex
	accountID        string
	backfillComplete bool
	entries          map[string]Entry
}{entries: map[string]Entry{}}

func Resolve(filePath string) (Entry, bool, error) {
	entry, _, found, err := ResolveCanonical(filePath)
	return entry, found, err
}

func ResolveCanonical(filePath string) (Entry, string, bool, error) {
	if err := ensureLoaded(); err != nil {
		return Entry{}, "", false, err
	}
	canonicalFile, err := canonicalPath(filePath)
	if err != nil {
		return Entry{}, "", false, err
	}

	indexState.RLock()
	defer indexState.RUnlock()
	entry, found := findEntry(indexState.entries, canonicalFile)
	return entry, canonicalFile, found, nil
}

func Upsert(entry Entry) error {
	return UpsertMany([]Entry{entry})
}

func UpsertMany(entries []Entry) error {
	if len(entries) == 0 {
		return nil
	}
	if err := ensureLoaded(); err != nil {
		return err
	}

	prepared := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		canonicalEntry, err := prepareEntry(entry)
		if err != nil {
			return err
		}
		prepared = append(prepared, canonicalEntry)
	}
	if len(prepared) == 0 {
		return nil
	}

	indexState.Lock()
	defer indexState.Unlock()
	for _, entry := range prepared {
		indexState.entries[entry.ProjectID] = entry
	}
	return saveLocked()
}

func RemoveProject(projectID string) error {
	if strings.TrimSpace(projectID) == "" {
		return nil
	}
	if err := ensureLoaded(); err != nil {
		return err
	}
	indexState.Lock()
	defer indexState.Unlock()
	if _, exists := indexState.entries[projectID]; !exists {
		return nil
	}
	delete(indexState.entries, projectID)
	return saveLocked()
}

func RemoveDatabasePath(databasePath string) error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	canonicalDatabase, err := canonicalPath(databasePath)
	if err != nil {
		return err
	}

	indexState.Lock()
	defer indexState.Unlock()
	changed := false
	for projectID, entry := range indexState.entries {
		if samePath(entry.DatabasePath, canonicalDatabase) {
			delete(indexState.entries, projectID)
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return saveLocked()
}

func Reset() {
	resetDiscovery()
	indexState.Lock()
	indexState.accountID = ""
	indexState.backfillComplete = false
	indexState.entries = map[string]Entry{}
	indexState.Unlock()
}

func isBackfillComplete() (bool, error) {
	if err := ensureLoaded(); err != nil {
		return false, err
	}
	indexState.RLock()
	defer indexState.RUnlock()
	return indexState.backfillComplete, nil
}

func markBackfillComplete() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	indexState.Lock()
	defer indexState.Unlock()
	indexState.backfillComplete = true
	return saveLocked()
}

func ensureLoaded() error {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	indexState.RLock()
	loaded := indexState.accountID == user.Id
	indexState.RUnlock()
	if loaded {
		return nil
	}

	indexState.Lock()
	defer indexState.Unlock()
	if indexState.accountID == user.Id {
		return nil
	}
	path, err := indexPath(user.Id)
	if err != nil {
		return err
	}
	stored, err := readIndexFile(path)
	if err != nil {
		return err
	}
	indexState.accountID = user.Id
	indexState.backfillComplete = stored.BackfillComplete
	indexState.entries = stored.Entries
	return nil
}

func prepareEntry(entry Entry) (Entry, error) {
	if strings.TrimSpace(entry.ProjectID) == "" ||
		strings.TrimSpace(entry.DatabasePath) == "" ||
		strings.TrimSpace(entry.WorkingDirectory) == "" {
		return Entry{}, errors.New("project index entry is incomplete")
	}
	databasePath, err := canonicalPath(entry.DatabasePath)
	if err != nil {
		return Entry{}, err
	}
	workingDirectory, err := canonicalPath(entry.WorkingDirectory)
	if err != nil {
		return Entry{}, err
	}
	entry.DatabasePath = databasePath
	entry.WorkingDirectory = workingDirectory
	entry.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return entry, nil
}

func sortedEntries(entries map[string]Entry) []Entry {
	result := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry)
	}
	sort.Slice(result, func(first, second int) bool {
		return len(result[first].WorkingDirectory) > len(result[second].WorkingDirectory)
	})
	return result
}

func findEntry(entries map[string]Entry, canonicalFile string) (Entry, bool) {
	for _, entry := range sortedEntries(entries) {
		if pathWithin(canonicalFile, entry.WorkingDirectory) {
			return entry, true
		}
	}
	return Entry{}, false
}

func saveLocked() error {
	path, err := indexPath(indexState.accountID)
	if err != nil {
		return err
	}
	return writeIndexFile(path, indexFile{
		Version:          indexVersion,
		BackfillComplete: indexState.backfillComplete,
		Entries:          indexState.entries,
	})
}

func indexPath(accountID string) (string, error) {
	roamingPath, err := settings.GetRoamingPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(
		roamingPath,
		"Clustta",
		"project-indexes",
		accountID+".json",
	), nil
}

func readIndexFile(path string) (indexFile, error) {
	stored := indexFile{Version: indexVersion, Entries: map[string]Entry{}}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return stored, nil
	}
	if err != nil {
		return stored, err
	}
	if err = json.Unmarshal(data, &stored); err != nil {
		log.Printf("Ignoring invalid project index %s: %v", path, err)
		return indexFile{Version: indexVersion, Entries: map[string]Entry{}}, nil
	}
	if stored.Version != indexVersion {
		log.Printf("Ignoring unsupported project index version %d", stored.Version)
		return indexFile{Version: indexVersion, Entries: map[string]Entry{}}, nil
	}
	if stored.Entries == nil {
		stored.Entries = map[string]Entry{}
	}
	return stored, nil
}

func writeIndexFile(path string, index indexFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".project-index-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err = temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err = temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err = temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err = temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func canonicalPath(path string) (string, error) {
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err == nil {
		return resolved, nil
	}
	return absolute, nil
}

func pathWithin(path string, root string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func samePath(first string, second string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(first, second)
	}
	return first == second
}
