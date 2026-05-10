package services

import (
	"clustta/internal/repository"
	"clustta/internal/utils"
	"encoding/base64"
	"fmt"
	"os"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type DialogService struct{}

// SelectIconDialog opens a file dialog to select an icon image.
// Returns the base64-encoded resized icon or an error.
func (f *DialogService) SelectIconDialog() (string, error) {
	dialog := application.Get().Dialog.OpenFile().
		CanChooseFiles(true).
		AttachToWindow(application.Get().Window.Current()).
		CanCreateDirectories(true).
		ShowHiddenFiles(true).AddFilter("Select Icon", "*.png; *.jpg; *.jpeg;")
	if runtime.GOOS == "darwin" {
		dialog.SetMessage("Select Icon")
	} else {
		dialog.SetTitle("Select Icon")
	}

	path, err := dialog.PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}

	stat, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if stat.Size() > repository.MaxProjectIconBytes {
		return "", fmt.Errorf("icon exceeds %d KB limit", repository.MaxProjectIconBytes>>10)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	resizedImageBytes, err := utils.ResizeImage(data, 50, 50)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(resizedImageBytes), nil
}

// SelectFileDialog opens a file dialog with custom title and filters.
// Returns the selected file path or an empty string if cancelled.
func (f *DialogService) SelectFileDialog(title, filters string) (string, error) {
	dialog := application.Get().Dialog.OpenFile().
		CanChooseFiles(true).
		AttachToWindow(application.Get().Window.Current()).
		CanCreateDirectories(true).
		ShowHiddenFiles(true).AddFilter("Files", filters)
	if runtime.GOOS == "darwin" {
		dialog.SetMessage(title)
	} else {
		dialog.SetTitle(title)
	}

	result, _ := dialog.PromptForSingleSelection()
	return result, nil
}

// SelectFilesDialog opens a file dialog to select multiple files.
// Returns the selected file paths or an empty list if cancelled.
func (f *DialogService) SelectFilesDialog() ([]string, error) {
	result, _ := application.Get().Dialog.OpenFile().
		CanChooseFiles(true).
		AttachToWindow(application.Get().Window.Current()).
		CanCreateDirectories(true).
		ShowHiddenFiles(true).
		PromptForMultipleSelection()
	return result, nil
}

// SelectItemsDialog opens a dialog to select multiple files or directories.
// Returns the selected paths or an empty list if cancelled.
func (f *DialogService) SelectItemsDialog() ([]string, error) {
	dialog := application.Get().Dialog.OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(true).
		AttachToWindow(application.Get().Window.Current()).
		CanCreateDirectories(true).
		ShowHiddenFiles(true)
	if runtime.GOOS == "darwin" {
		dialog.SetMessage("select items")
	} else {
		dialog.SetTitle("select items")
	}

	results, _ := dialog.PromptForMultipleSelection()
	return results, nil
}

// SelectFolderDialog opens a folder selection dialog with a custom title.
// Returns the selected folder path or an empty string if cancelled.
func (f *DialogService) SelectFolderDialog(title string) (string, error) {
	dialog := application.Get().Dialog.OpenFile().
		CanChooseDirectories(true).
		AttachToWindow(application.Get().Window.Current()).
		CanCreateDirectories(true).
		ShowHiddenFiles(true)
	if runtime.GOOS == "darwin" {
		dialog.SetMessage(title)
	} else {
		dialog.SetTitle(title)
	}

	result, _ := dialog.PromptForSingleSelection()
	return result, nil
}

// SelectSpecificFolderDialog opens a folder dialog with a default starting path.
// Returns the selected folder path or an error.
func (f *DialogService) SelectSpecificFolderDialog(title string, defaultPath string) (string, error) {
	dialog := application.Get().Dialog.OpenFile().
		CanChooseDirectories(true).
		AttachToWindow(application.Get().Window.Current()).
		CanCreateDirectories(true).
		ShowHiddenFiles(true).
		SetDirectory(defaultPath)

	if runtime.GOOS == "darwin" {
		dialog.SetMessage(title)
	} else {
		dialog.SetTitle(title)
	}

	result, err := dialog.PromptForSingleSelection()
	return result, err
}
