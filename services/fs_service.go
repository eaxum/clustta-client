package services

import (
	"archive/tar"
	"archive/zip"
	"clustta/internal/custom_thumbnail"
	"clustta/internal/settings"
	"clustta/internal/system_icon"
	"clustta/internal/system_thumbnail"
	"clustta/internal/utils"
	"clustta/output"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/skratchdot/open-golang/open"
	"github.com/wailsapp/wails/v3/pkg/application"
	// "syscall"
	// "unsafe"
)

type FSService struct {
	Watcher *fsnotify.Watcher
}

type FileInfo struct {
	Name          string `json:"name"`
	Size          int64  `json:"size"`
	FormattedSize string `json:"formattedSize"`
	IsDir         bool   `json:"isDir"`
	ModTime       int64  `json:"modTime"`
}

func (f *FSService) AddWatcherFolder(dir string) error {
	return f.Watcher.Add(dir)
}

func (f *FSService) RemoveWatcherFolder(dir string) error {
	return f.Watcher.Remove(dir)
}

func (f *FSService) Exists(path string) bool {
	return utils.FileExists(path)
}
func (f *FSService) DirExists(path string) bool {
	return utils.DirExists(path)
}
func (f *FSService) IsFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, err
	}
	if err != nil {
		// panic(err)
		return false, err
	}
	return !info.IsDir(), nil
}
func (f *FSService) GetFileIcon(ext string) (string, error) {
	img, err := system_icon.GetExtensionIcon(ext)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(img), nil
}

// GetOSThumbnail generates a thumbnail for the specified file using OS APIs
// Always fetches full-resolution (512px) thumbnails for maximum quality
// Now with custom thumbnail extraction for Blender, Maya, and other 3D files
// Returns base64-encoded PNG thumbnail or empty string on error
func (f *FSService) GetOSThumbnail(filePath string, size int) (string, error) {
	// Check if file exists
	if !utils.FileExists(filePath) {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}

	// Try custom thumbnail extraction first (Blender, Maya, etc.)
	customThumbnail, err := custom_thumbnail.GetThumbnail(filePath)
	if err == nil && customThumbnail != nil && len(customThumbnail) > 0 {
		// Successfully extracted custom thumbnail
		return base64.StdEncoding.EncodeToString(customThumbnail), nil
	}

	// Fall back to OS thumbnail generation
	// Always use 512px for maximum quality - CSS handles sizing
	// The size parameter is kept for API compatibility but not used
	thumbnailBytes, err := system_thumbnail.GetOSThumbnail(
		filePath,
		512, // Fixed size for full-resolution thumbnails
		system_thumbnail.ThumbnailUseCurrentScale,
	)
	if err != nil {
		// Return empty string instead of error to allow graceful fallback
		return "", nil
	}

	// Encode to base64
	return base64.StdEncoding.EncodeToString(thumbnailBytes), nil
}

// GetCachedOSThumbnail attempts to get a cached thumbnail without generating a new one
// Always fetches full-resolution (512px) thumbnails for maximum quality
// Returns base64-encoded PNG thumbnail or empty string if not cached
func (f *FSService) GetCachedOSThumbnail(filePath string, size int) (string, error) {
	// Check if file exists
	if !utils.FileExists(filePath) {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}

	// Always use 512px for maximum quality - CSS handles sizing
	// The size parameter is kept for API compatibility but not used
	thumbnailBytes, err := system_thumbnail.GetCachedThumbnail(filePath, 512)
	if err != nil {
		// Return empty string if not in cache
		return "", nil
	}

	// Encode to base64
	return base64.StdEncoding.EncodeToString(thumbnailBytes), nil
}

// GetOSThumbnails generates thumbnails for multiple files in batch
// Always fetches full-resolution (512px) thumbnails for maximum quality
// Returns a map of file paths to base64-encoded thumbnails
func (f *FSService) GetOSThumbnails(filePaths []string, size int) (map[string]string, error) {
	results := make(map[string]string)

	for _, path := range filePaths {
		// Skip non-existent files
		if !utils.FileExists(path) {
			continue
		}

		// Always use 512px for maximum quality - CSS handles sizing
		// The size parameter is kept for API compatibility but not used
		thumbnailBytes, err := system_thumbnail.GetOSThumbnail(
			path,
			512, // Fixed size for full-resolution thumbnails
			system_thumbnail.ThumbnailUseCurrentScale,
		)
		if err != nil {
			// Store empty string for failed thumbnails
			results[path] = ""
			continue
		}

		// Store base64-encoded thumbnail
		results[path] = base64.StdEncoding.EncodeToString(thumbnailBytes)
	}

	return results, nil
}

func (f *FSService) LaunchFile(path string) error {
	return open.Start(path)
}

func (f *FSService) LaunchFileWith(path string) error {
	// filePath, err := filepath.Abs(path)
	// if err != nil {
	// 	return err
	// }

	// h := syscall.MustLoadDLL("shell32.dll")
	// c := h.MustFindProc("ShellExecuteW")

	// openWithPtr, err := syscall.UTF16PtrFromString("rundll32.exe")
	// if err != nil {
	// 	return err
	// }

	// paramsPtr, err := syscall.UTF16PtrFromString("shell32.dll,OpenAs_RunDLL " + filePath)
	// if err != nil {
	// 	return err
	// }

	// ret, _, err := c.Call(
	// 	0,                                    // hwnd
	// 	0,                                    // verb (NULL for default)
	// 	uintptr(unsafe.Pointer(openWithPtr)), // file
	// 	uintptr(unsafe.Pointer(paramsPtr)),   // params
	// 	0,                                    // directory
	// 	1,                                    // show
	// )

	// if ret <= 32 {
	// 	return err
	// }
	return nil
}

func (f *FSService) ExtName(path string) string {
	return filepath.Ext(path)
}
func (f *FSService) BaseName(path string) string {
	return filepath.Base(path)
}
func (f *FSService) JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}
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
func (f *FSService) FileHash(path string) (string, error) {
	hash, err := utils.GenerateXXHashChecksum(path)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func (f *FSService) DeleteFolder(path string) error {
	return os.RemoveAll(path)
}

func (f *FSService) DeleteFile(path string) error {
	return os.Remove(path)
}

func (f *FSService) TempDir() string {
	return os.TempDir()
}
func (f *FSService) UserProjectTemplatesPath() (string, error) {
	return settings.GetUserProjectTemplatesPath()
}

func (f *FSService) WriteFile(path string, data string) error {
	// base64 decode

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, decoded, 0644)
}

func (f *FSService) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// base64 encode
	return base64.StdEncoding.EncodeToString(data), nil
}

func (f *FSService) RevealInExplorer(path string) {
	utils.RevealInExplorer(path)
}

func (f *FSService) MakeDirs(path string) {
	os.MkdirAll(path, os.ModePerm)
}

func (f *FSService) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (f *FSService) DuplicateFile(sourcePath, destinationPath string) (string, error) {
	// Open the source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Get file info to check if destination is a directory
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get source file info: %w", err)
	}

	// Check if destination is a directory
	destInfo, err := os.Stat(destinationPath)
	if err == nil && destInfo.IsDir() {
		// If destination is a directory, use the source filename
		destinationPath = filepath.Join(destinationPath, filepath.Base(sourcePath))
	}

	// Create the destination file
	destFile, err := os.Create(destinationPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy the contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return "", fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Preserve file mode/permissions
	err = os.Chmod(destinationPath, sourceInfo.Mode())
	if err != nil {
		return "", fmt.Errorf("failed to set file permissions: %w", err)
	}
	return destinationPath, nil
}

func (f *FSService) DuplicateFolder(sourcePath, destinationPath string) error {
	// Get source folder info
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to stat source folder: %w", err)
	}

	if !sourceInfo.IsDir() {
		return fmt.Errorf("source path is not a directory")
	}

	// Check if destination exists and is a directory
	destInfo, err := os.Stat(destinationPath)
	if err == nil {
		if destInfo.IsDir() {
			// If destination is a directory, create subfolder with source name
			destinationPath = filepath.Join(destinationPath, filepath.Base(sourcePath))
		} else {
			return fmt.Errorf("destination exists and is not a directory")
		}
	}

	// Create the destination directory
	err = os.MkdirAll(destinationPath, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Walk through source directory and copy all contents
	err = filepath.WalkDir(sourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path from source root
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}

		// Calculate destination path
		destPath := filepath.Join(destinationPath, relPath)

		if d.IsDir() {
			// Create directory
			info, err := d.Info()
			if err != nil {
				return err
			}
			return os.MkdirAll(destPath, info.Mode())
		} else {
			// Copy file
			_, err := f.DuplicateFile(path, destPath)
			return err
		}
	})

	if err != nil {
		return fmt.Errorf("failed to copy folder contents: %w", err)
	}

	return nil
}

// ExtractAll extracts archive contents to a folder in the current location
// Supports .zip, .tar, .tar.gz, .gz formats
// Sends progress updates to the frontend using output.ProgressReport
func (f *FSService) ExtractAll(archivePath string) error {
	app := application.Get()

	// Get file extension to determine archive type
	ext := strings.ToLower(filepath.Ext(archivePath))
	baseNameWithoutExt := strings.TrimSuffix(filepath.Base(archivePath), ext)

	// Check for .tar.gz or .tar.bz2
	if strings.HasSuffix(strings.ToLower(archivePath), ".tar.gz") || strings.HasSuffix(strings.ToLower(archivePath), ".tar.bz2") {
		ext = ".tar.gz"
		baseNameWithoutExt = strings.TrimSuffix(baseNameWithoutExt, ".tar")
	}

	// Create extraction folder in the same directory as the archive
	archiveDir := filepath.Dir(archivePath)
	extractionDir := filepath.Join(archiveDir, baseNameWithoutExt)

	// Check if extraction directory already exists
	if _, err := os.Stat(extractionDir); err == nil {
		// Directory exists, append a number
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

	// Create the extraction directory
	err := os.MkdirAll(extractionDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create extraction directory: %w", err)
	}

	// Send initial progress
	progress := output.ProgressReport{
		Title:         "Extracting Archive",
		Message:       filepath.Base(archivePath),
		Percentage:    0,
		Current:       0,
		Total:         100,
		OperationType: "read", // Read operation - doesn't modify database
	}
	app.EmitEvent("progress-update", progress)

	// Extract based on file type
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
		// Clean up partial extraction on error
		os.RemoveAll(extractionDir)
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Send completion progress
	progress = output.ProgressReport{
		Title:         "Extracting Archive",
		Message:       "Complete",
		Percentage:    100,
		Current:       100,
		Total:         100,
		OperationType: "read",
	}
	app.EmitEvent("progress-update", progress)

	return nil
}

// extractZip extracts a zip archive
func (f *FSService) extractZip(archivePath, destDir string, app *application.App) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	totalFiles := len(reader.File)
	processed := 0

	for _, file := range reader.File {
		// Update progress
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
		app.EmitEvent("progress-update", progress)

		// Construct target path
		targetPath := filepath.Join(destDir, file.Name)

		// Security check: ensure the path doesn't escape the destination directory
		if !strings.HasPrefix(targetPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			// Create directory
			os.MkdirAll(targetPath, os.ModePerm)
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
			return err
		}

		// Extract file
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

// extractTar extracts a tar archive
func (f *FSService) extractTar(archivePath, destDir string, app *application.App) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	tarReader := tar.NewReader(file)
	return f.processTarReader(tarReader, destDir, app)
}

// extractTarGz extracts a tar.gz archive
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

// processTarReader processes a tar reader and extracts files
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
			Percentage:    float64(processed) * 2, // Approximate progress
			Current:       processed,
			Total:         0,
			OperationType: "read",
		}
		app.EmitEvent("progress-update", progress)

		// Construct target path
		targetPath := filepath.Join(destDir, header.Name)

		// Security check: ensure the path doesn't escape the destination directory
		if !strings.HasPrefix(targetPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}

		case tar.TypeReg:
			// Create parent directories
			if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
				return err
			}

			// Extract file
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
