package services

import (
	"archive/tar"
	"archive/zip"
	"clustta/internal/auth_service"
	"clustta/internal/custom_thumbnail"
	"clustta/internal/settings"
	"clustta/internal/system_icon"
	"clustta/internal/system_thumbnail"
	"clustta/internal/utils"
	"clustta/output"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/skratchdot/open-golang/open"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type FSService struct {
	Watcher          *fsnotify.Watcher
	debounceTimers   map[string]*time.Timer
	debounceMutex    sync.Mutex
	debounceDuration time.Duration
	app              *application.App
}

type FileInfo struct {
	Name          string `json:"name"`
	Size          int64  `json:"size"`
	FormattedSize string `json:"formattedSize"`
	IsDir         bool   `json:"isDir"`
	ModTime       int64  `json:"modTime"`
}

// AddWatcherFolder registers a directory with the file system watcher.
// Enables monitoring of file system events within the specified directory.
func (f *FSService) AddWatcherFolder(dir string) error {
	return f.Watcher.Add(dir)
}

// RemoveWatcherFolder unregisters a directory from the file system watcher.
// Stops monitoring file system events for the specified directory.
func (f *FSService) RemoveWatcherFolder(dir string) error {
	return f.Watcher.Remove(dir)
}

// SetApp sets the application instance for the FSService.
// Required for emitting events to the frontend application.
func (f *FSService) SetApp(app *application.App) {
	f.app = app
}

// StartWatching initializes the file system watcher and handles events.
// Runs a goroutine that monitors file changes and emits debounced events to the frontend.
func (f *FSService) StartWatching() {
	f.debounceTimers = make(map[string]*time.Timer)
	f.debounceDuration = 500 * time.Millisecond

	go func() {
		for {
			select {
			case event, ok := <-f.Watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create ||
					event.Op&fsnotify.Remove == fsnotify.Remove ||
					event.Op&fsnotify.Rename == fsnotify.Rename {

					f.debounceEvent(event.Name, func() {
						if f.app != nil {
							f.app.Event.Emit("fs-change", event.Name)
						}
					})
				}

			case err, ok := <-f.Watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("watcher error: %v\n", err)
			}
		}
	}()
}

// debounceEvent debounces file system events to prevent flooding.
// Delays the callback execution and cancels any pending timers for the same path.
func (f *FSService) debounceEvent(path string, callback func()) {
	f.debounceMutex.Lock()
	defer f.debounceMutex.Unlock()

	if timer, exists := f.debounceTimers[path]; exists {
		timer.Stop()
	}

	f.debounceTimers[path] = time.AfterFunc(f.debounceDuration, func() {
		callback()

		f.debounceMutex.Lock()
		delete(f.debounceTimers, path)
		f.debounceMutex.Unlock()
	})
}

// Exists checks if a file exists at the specified path.
func (f *FSService) Exists(path string) bool {
	return utils.FileExists(path)
}

// DirExists checks if a directory exists at the specified path.
func (f *FSService) DirExists(path string) bool {
	return utils.DirExists(path)
}

// IsFile checks if the specified path is a file (not a directory).
// Returns an error if the path does not exist.
func (f *FSService) IsFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, err
	}
	if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}

// GetFileIcon retrieves the system icon for a file extension.
// Returns the icon as a base64-encoded string.
func (f *FSService) GetFileIcon(ext string) (string, error) {
	img, err := system_icon.GetExtensionIcon(ext)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(img), nil
}

// GetOSThumbnail generates a thumbnail for the specified file.
// Attempts custom extraction first (Blender, Maya), falls back to OS thumbnail. Returns base64-encoded image.
func (f *FSService) GetOSThumbnail(filePath string, size int) (string, error) {
	if !utils.FileExists(filePath) {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}

	customThumbnail, err := custom_thumbnail.GetThumbnail(filePath)
	if err == nil && customThumbnail != nil && len(customThumbnail) > 0 {
		return base64.StdEncoding.EncodeToString(customThumbnail), nil
	}

	thumbnailBytes, err := system_thumbnail.GetOSThumbnail(
		filePath,
		512,
		system_thumbnail.ThumbnailUseCurrentScale,
	)
	if err != nil {
		return "", nil
	}

	return base64.StdEncoding.EncodeToString(thumbnailBytes), nil
}

// GetCachedOSThumbnail retrieves a cached thumbnail for the specified file.
// Returns base64-encoded image if cached, empty string if not in cache.
func (f *FSService) GetCachedOSThumbnail(filePath string, size int) (string, error) {
	if !utils.FileExists(filePath) {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}

	thumbnailBytes, err := system_thumbnail.GetCachedThumbnail(filePath, 512)
	if err != nil {
		return "", nil
	}

	return base64.StdEncoding.EncodeToString(thumbnailBytes), nil
}

// GetOSThumbnails generates thumbnails for multiple files in batch.
// Returns a map of file paths to base64-encoded thumbnails. Empty strings for failures.
func (f *FSService) GetOSThumbnails(filePaths []string, size int) (map[string]string, error) {
	results := make(map[string]string)

	for _, path := range filePaths {
		if !utils.FileExists(path) {
			continue
		}

		thumbnailBytes, err := system_thumbnail.GetOSThumbnail(
			path,
			512,
			system_thumbnail.ThumbnailUseCurrentScale,
		)
		if err != nil {
			results[path] = ""
			continue
		}

		results[path] = base64.StdEncoding.EncodeToString(thumbnailBytes)
	}

	return results, nil
}

// LaunchFile opens a file with its default system application.
// Validates the path exists before opening to prevent command injection.
func (f *FSService) LaunchFile(path string) error {
	cleanPath := filepath.Clean(path)
	if _, err := os.Stat(cleanPath); err != nil {
		return fmt.Errorf("file not found: %s", cleanPath)
	}
	return open.Start(cleanPath)
}

// ExtName returns the file extension of the specified path.
func (f *FSService) ExtName(path string) string {
	return filepath.Ext(path)
}

// BaseName returns the last element of the path.
func (f *FSService) BaseName(path string) string {
	return filepath.Base(path)
}

// JoinPath joins multiple path elements into a single path.
func (f *FSService) JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}

// FolderSize calculates the total size of a folder and its contents.
// Returns a formatted string (B, KB, MB, GB).
func (f *FSService) FolderSize(folderPath string) (string, error) {
	var size int64
	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if info, err := d.Info(); err == nil {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return formatSize(size), nil
}

func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// FileStat retrieves file information including name, size, and modification time.
// Returns a FileInfo struct with formatted size.
func (f *FSService) FileStat(path string) (FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileInfo{}, err
	}
	return FileInfo{
		Name:          info.Name(),
		Size:          info.Size(),
		FormattedSize: formatSize(info.Size()),
		IsDir:         info.IsDir(),
		ModTime:       info.ModTime().Unix(),
	}, nil
}

// FileCount counts the total number of files in a folder recursively.
func (f *FSService) FileCount(folderPath string) (int, error) {
	var count int
	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	return count, err
}

// FolderCount counts the total number of subdirectories in a folder recursively.
func (f *FSService) FolderCount(folderPath string) (int, error) {
	var count int
	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && path != folderPath {
			count++
		}
		return nil
	})
	return count, err
}

// FileHash generates an XXHash checksum for the specified file.
func (f *FSService) FileHash(path string) (string, error) {
	hash, err := utils.GenerateXXHashChecksum(path)
	if err != nil {
		return "", err
	}
	return hash, nil
}

// DeleteFolder removes a folder and all its contents recursively.
func (f *FSService) DeleteFolder(path string) error {
	return os.RemoveAll(path)
}

// DeleteFile removes a single file from the file system.
func (f *FSService) DeleteFile(path string) error {
	return os.Remove(path)
}

// TempDir returns the system's temporary directory path.
func (f *FSService) TempDir() string {
	return os.TempDir()
}

// UserProjectTemplatesPath retrieves the user's project templates directory path.
func (f *FSService) UserProjectTemplatesPath() (string, error) {
	return settings.GetUserProjectTemplatesPath()
}

// WriteFile writes base64-encoded data to a file.
// Decodes the data before writing to disk.
func (f *FSService) WriteFile(path string, data string) error {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, decoded, 0644)
}

// ReadFile reads a file and returns its contents as base64-encoded string.
func (f *FSService) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// RevealInExplorer opens the system file explorer and highlights the specified path.
func (f *FSService) RevealInExplorer(path string) {
	utils.RevealInExplorer(path)
}

// MakeDirs creates a directory and all necessary parent directories.
func (f *FSService) MakeDirs(path string) {
	os.MkdirAll(path, os.ModePerm)
}

// Rename moves or renames a file or directory from oldPath to newPath.
func (f *FSService) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// RenameOperation represents a single rename operation with old and new paths.
type RenameOperation struct {
	OldPath string `json:"oldPath"`
	NewPath string `json:"newPath"`
}

// RenameBatch moves or renames multiple files or directories.
// Accepts a JSON string containing an array of RenameOperation objects.
// Returns an error if any operation fails.
func (f *FSService) RenameBatch(operationsJSON string) error {
	var operations []RenameOperation
	if err := json.Unmarshal([]byte(operationsJSON), &operations); err != nil {
		return fmt.Errorf("failed to parse rename operations: %w", err)
	}

	for _, op := range operations {
		err := os.Rename(op.OldPath, op.NewPath)
		if err != nil {
			return err
		}
	}
	return nil
}

// BackupFile creates a backup copy of a file with progress reporting.
// Sends progress updates to the frontend during the copy operation.
func (f *FSService) BackupFile(sourcePath, destinationPath string) (string, error) {
	app := application.Get()

	progress := output.ProgressReport{
		Title:         "Backing Up Project",
		Message:       "Preparing backup...",
		Percentage:    0,
		Current:       0,
		Total:         100,
		OperationType: "read",
	}
	app.Event.Emit("progress-update", progress)

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get source file info: %w", err)
	}

	destInfo, err := os.Stat(destinationPath)
	if err == nil && destInfo.IsDir() {
		destinationPath = filepath.Join(destinationPath, filepath.Base(sourcePath))
	}

	destFile, err := os.Create(destinationPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	totalSize := sourceInfo.Size()
	buffer := make([]byte, 1024*1024)
	var copiedBytes int64
	lastProgressUpdate := int64(0)
	updateInterval := totalSize / 20

	for {
		nr, err := sourceFile.Read(buffer)
		if nr > 0 {
			nw, err := destFile.Write(buffer[0:nr])
			if err != nil {
				return "", fmt.Errorf("failed to write to destination: %w", err)
			}
			if nr != nw {
				return "", fmt.Errorf("short write")
			}
			copiedBytes += int64(nw)

			if copiedBytes-lastProgressUpdate >= updateInterval || copiedBytes == totalSize {
				percentage := (float64(copiedBytes) / float64(totalSize)) * 100
				progress = output.ProgressReport{
					Title:         "Backing Up Project",
					Message:       filepath.Base(sourcePath),
					Percentage:    percentage,
					Current:       1,
					Total:         1,
					OperationType: "read",
				}
				app.Event.Emit("progress-update", progress)
				lastProgressUpdate = copiedBytes
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to copy file contents: %w", err)
		}
	}

	err = os.Chmod(destinationPath, sourceInfo.Mode())
	if err != nil {
		return "", fmt.Errorf("failed to set file permissions: %w", err)
	}

	progress = output.ProgressReport{
		Title:         "Backing Up Project",
		Message:       "Backup complete",
		Percentage:    100,
		Current:       1,
		Total:         1,
		OperationType: "read",
	}
	app.Event.Emit("progress-update", progress)

	return destinationPath, nil
}

// DuplicateFile creates a copy of a file to the specified destination.
// Preserves file permissions and automatically handles directory destinations.
func (f *FSService) DuplicateFile(sourcePath, destinationPath string) (string, error) {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get source file info: %w", err)
	}

	destInfo, err := os.Stat(destinationPath)
	if err == nil && destInfo.IsDir() {
		destinationPath = filepath.Join(destinationPath, filepath.Base(sourcePath))
	}

	destFile, err := os.Create(destinationPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return "", fmt.Errorf("failed to copy file contents: %w", err)
	}

	err = os.Chmod(destinationPath, sourceInfo.Mode())
	if err != nil {
		return "", fmt.Errorf("failed to set file permissions: %w", err)
	}
	return destinationPath, nil
}

// DuplicateFolder recursively copies a folder and all its contents to the destination.
// Preserves directory structure and file permissions.
func (f *FSService) DuplicateFolder(sourcePath, destinationPath string) error {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to stat source folder: %w", err)
	}

	if !sourceInfo.IsDir() {
		return fmt.Errorf("source path is not a directory")
	}

	destInfo, err := os.Stat(destinationPath)
	if err == nil {
		if destInfo.IsDir() {
			destinationPath = filepath.Join(destinationPath, filepath.Base(sourcePath))
		} else {
			return fmt.Errorf("destination exists and is not a directory")
		}
	}

	err = os.MkdirAll(destinationPath, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	err = filepath.WalkDir(sourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destinationPath, relPath)

		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			return os.MkdirAll(destPath, info.Mode())
		} else {
			_, err := f.DuplicateFile(path, destPath)
			return err
		}
	})

	if err != nil {
		return fmt.Errorf("failed to copy folder contents: %w", err)
	}

	return nil
}

// ExtractAll extracts archive contents to a folder in the current location.
// Supports .zip, .tar, .tar.gz, .gz formats with progress reporting to the frontend.
func (f *FSService) ExtractAll(archivePath string) error {
	app := application.Get()

	ext := strings.ToLower(filepath.Ext(archivePath))
	baseNameWithoutExt := strings.TrimSuffix(filepath.Base(archivePath), ext)

	if strings.HasSuffix(strings.ToLower(archivePath), ".tar.gz") || strings.HasSuffix(strings.ToLower(archivePath), ".tar.bz2") {
		ext = ".tar.gz"
		baseNameWithoutExt = strings.TrimSuffix(baseNameWithoutExt, ".tar")
	}

	archiveDir := filepath.Dir(archivePath)
	extractionDir := filepath.Join(archiveDir, baseNameWithoutExt)

	if _, err := os.Stat(extractionDir); err == nil {
		counter := 1
		for {
			testDir := fmt.Sprintf("%s_%d", extractionDir, counter)
			if _, err := os.Stat(testDir); os.IsNotExist(err) {
				extractionDir = testDir
				break
			}
			counter++
		}
	}

	err := os.MkdirAll(extractionDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create extraction directory: %w", err)
	}

	progress := output.ProgressReport{
		Title:         "Extracting Archive",
		Message:       filepath.Base(archivePath),
		Percentage:    0,
		Current:       0,
		Total:         100,
		OperationType: "read",
	}
	app.Event.Emit("progress-update", progress)

	switch ext {
	case ".zip":
		err = f.extractZip(archivePath, extractionDir, app)
	case ".tar":
		err = f.extractTar(archivePath, extractionDir, app)
	case ".tar.gz", ".gz":
		err = f.extractTarGz(archivePath, extractionDir, app)
	default:
		return fmt.Errorf("unsupported archive format: %s", ext)
	}

	if err != nil {
		os.RemoveAll(extractionDir)
		return fmt.Errorf("extraction failed: %w", err)
	}

	progress = output.ProgressReport{
		Title:         "Extracting Archive",
		Message:       "Complete",
		Percentage:    100,
		Current:       100,
		Total:         100,
		OperationType: "read",
	}
	app.Event.Emit("progress-update", progress)

	return nil
}

// extractZip extracts a zip archive with progress tracking.
func (f *FSService) extractZip(archivePath, destDir string, app *application.App) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	totalFiles := len(reader.File)
	processed := 0

	for _, file := range reader.File {
		processed++
		percentage := float64(processed) / float64(totalFiles) * 100
		progress := output.ProgressReport{
			Title:         "Extracting Archive",
			Message:       file.Name,
			Percentage:    percentage,
			Current:       processed,
			Total:         totalFiles,
			OperationType: "read",
		}
		app.Event.Emit("progress-update", progress)

		targetPath := filepath.Join(destDir, file.Name)

		if !strings.HasPrefix(targetPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(targetPath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
			return err
		}

		srcFile, err := file.Open()
		if err != nil {
			return err
		}

		destFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			srcFile.Close()
			return err
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// extractTar extracts a tar archive with progress tracking.
func (f *FSService) extractTar(archivePath, destDir string, app *application.App) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	tarReader := tar.NewReader(file)
	return f.processTarReader(tarReader, destDir, app)
}

// extractTarGz extracts a tar.gz archive with progress tracking.
func (f *FSService) extractTarGz(archivePath, destDir string, app *application.App) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	return f.processTarReader(tarReader, destDir, app)
}

// processTarReader processes a tar reader and extracts files with progress tracking.
func (f *FSService) processTarReader(tarReader *tar.Reader, destDir string, app *application.App) error {
	processed := 0

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		processed++
		progress := output.ProgressReport{
			Title:         "Extracting Archive",
			Message:       header.Name,
			Percentage:    float64(processed) * 2,
			Current:       processed,
			Total:         0,
			OperationType: "read",
		}
		app.Event.Emit("progress-update", progress)

		targetPath := filepath.Join(destDir, header.Name)

		if !strings.HasPrefix(targetPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}

		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
				return err
			}

			destFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			_, err = io.Copy(destFile, tarReader)
			destFile.Close()

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ImportClusttaFiles imports multiple .clst files to the destination directory with progress reporting.
// Validates file extensions, checks for existing files, and copies each file with progress updates.
// Returns an array of destination file paths and any error encountered.
func (f *FSService) ImportClusttaFiles(sourcePaths []string, destinationDirectory string) ([]string, error) {
	app := application.Get()

	// Validate that destination exists and is a directory
	destInfo, err := os.Stat(destinationDirectory)
	if err != nil {
		return nil, fmt.Errorf("destination directory not accessible: %w", err)
	}
	if !destInfo.IsDir() {
		return nil, fmt.Errorf("destination path is not a directory")
	}

	// Validate all source files
	validFiles := []string{}
	for _, sourcePath := range sourcePaths {
		// Check if file exists
		if _, err := os.Stat(sourcePath); err != nil {
			return nil, fmt.Errorf("source file not found: %s", sourcePath)
		}

		// Validate .clst extension
		if filepath.Ext(sourcePath) != ".clst" {
			return nil, fmt.Errorf("invalid file type: %s (only .clst files allowed)", filepath.Base(sourcePath))
		}

		validFiles = append(validFiles, sourcePath)
	}

	if len(validFiles) == 0 {
		return nil, fmt.Errorf("no valid .clst files to import")
	}

	// Check for existing files in destination
	existingFiles := []string{}
	for _, sourcePath := range validFiles {
		destPath := filepath.Join(destinationDirectory, filepath.Base(sourcePath))
		if _, err := os.Stat(destPath); err == nil {
			existingFiles = append(existingFiles, filepath.Base(sourcePath))
		}
	}

	if len(existingFiles) > 0 {
		return nil, fmt.Errorf("files already exist in destination: %v", existingFiles)
	}

	// Import files with progress tracking
	destinationPaths := []string{}
	totalFiles := len(validFiles)

	for i, sourcePath := range validFiles {
		fileName := filepath.Base(sourcePath)
		destPath := filepath.Join(destinationDirectory, fileName)

		// Send progress update
		progress := output.ProgressReport{
			Title:         "Importing Projects",
			Message:       fileName,
			Percentage:    (float64(i) / float64(totalFiles)) * 100,
			Current:       i,
			Total:         totalFiles,
			OperationType: "write",
		}
		app.Event.Emit("progress-update", progress)

		// Open source file
		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			return destinationPaths, fmt.Errorf("failed to open source file %s: %w", fileName, err)
		}

		// Get source file info
		sourceInfo, err := sourceFile.Stat()
		if err != nil {
			sourceFile.Close()
			return destinationPaths, fmt.Errorf("failed to get source file info for %s: %w", fileName, err)
		}

		// Create destination file
		destFile, err := os.Create(destPath)
		if err != nil {
			sourceFile.Close()
			return destinationPaths, fmt.Errorf("failed to create destination file %s: %w", fileName, err)
		}

		// Copy file contents
		totalSize := sourceInfo.Size()
		buffer := make([]byte, 1024*1024) // 1MB buffer
		var copiedBytes int64

		for {
			nr, err := sourceFile.Read(buffer)
			if nr > 0 {
				nw, err := destFile.Write(buffer[0:nr])
				if err != nil {
					sourceFile.Close()
					destFile.Close()
					return destinationPaths, fmt.Errorf("failed to write to destination %s: %w", fileName, err)
				}
				if nr != nw {
					sourceFile.Close()
					destFile.Close()
					return destinationPaths, fmt.Errorf("short write for file %s", fileName)
				}
				copiedBytes += int64(nw)

				// Update progress within file
				fileProgress := (float64(copiedBytes) / float64(totalSize)) * 100
				overallProgress := ((float64(i) + (fileProgress / 100)) / float64(totalFiles)) * 100
				progress = output.ProgressReport{
					Title:         "Importing Projects",
					Message:       fileName,
					Percentage:    overallProgress,
					Current:       i + 1,
					Total:         totalFiles,
					OperationType: "write",
				}
				app.Event.Emit("progress-update", progress)
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				sourceFile.Close()
				destFile.Close()
				return destinationPaths, fmt.Errorf("failed to copy file contents for %s: %w", fileName, err)
			}
		}

		// Set file permissions
		err = os.Chmod(destPath, sourceInfo.Mode())
		if err != nil {
			sourceFile.Close()
			destFile.Close()
			return destinationPaths, fmt.Errorf("failed to set file permissions for %s: %w", fileName, err)
		}

		sourceFile.Close()
		destFile.Close()

		destinationPaths = append(destinationPaths, destPath)
	}

	// Send completion progress
	progress := output.ProgressReport{
		Title:         "Importing Projects",
		Message:       "Import complete",
		Percentage:    100,
		Current:       totalFiles,
		Total:         totalFiles,
		OperationType: "write",
	}
	app.Event.Emit("progress-update", progress)

	return destinationPaths, nil
}

// GetPersonalProjectsDirectory returns the path to the user's personal projects directory.
// Creates the directory if it doesn't exist.
func (f *FSService) GetPersonalProjectsDirectory() (string, error) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return "", fmt.Errorf("failed to get active user: %w", err)
	}

	userDataFolder, err := settings.GetUserDataFolder(user)
	if err != nil {
		return "", fmt.Errorf("failed to get user data folder: %w", err)
	}

	projectsDir := filepath.Join(userDataFolder, "Personal")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(projectsDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create projects directory: %w", err)
	}

	return projectsDir, nil
}
