package services

import (
	"clustta/internal/utils"
	"encoding/base64"
	"os"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type DialogService struct{}

func (f *DialogService) SelectIconDialog() (string, error) {
	dialog := application.OpenFileDialog().
		CanChooseFiles(true).
		AttachToWindow(application.Get().CurrentWindow()).
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

func (f *DialogService) SelectFileDialog(title, filters string) (string, error) {
	dialog := application.OpenFileDialog().
		CanChooseFiles(true).
		AttachToWindow(application.Get().CurrentWindow()).
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

func (f *DialogService) SelectFilesDialog() ([]string, error) {
	result, _ := application.OpenFileDialog().
		CanChooseFiles(true).
		AttachToWindow(application.Get().CurrentWindow()).
		CanCreateDirectories(true).
		ShowHiddenFiles(true).
		PromptForMultipleSelection()
	return result, nil
}

func (f *DialogService) SelectItemsDialog() ([]string, error) {
	dialog := application.OpenFileDialog().
		CanChooseDirectories(true).
		CanChooseFiles(true).
		AttachToWindow(application.Get().CurrentWindow()).
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

func (f *DialogService) SelectFolderDialog(title string) (string, error) {
	dialog := application.OpenFileDialog().
		CanChooseDirectories(true).
		AttachToWindow(application.Get().CurrentWindow()).
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

func (f *DialogService) SelectSpecificFolderDialog(title string, defaultPath string) (string, error) {
	dialog := application.OpenFileDialog().
		CanChooseDirectories(true).
		AttachToWindow(application.Get().CurrentWindow()).
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
