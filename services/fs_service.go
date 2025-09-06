package services

import (
	"clustta/internal/settings"
	"clustta/internal/system_icon"
	"clustta/internal/utils"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/skratchdot/open-golang/open"
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
