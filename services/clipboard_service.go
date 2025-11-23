package services

import (
	"encoding/base64"
	"fmt"

	"golang.design/x/clipboard"
)

type ClipboardService struct{}

//ReadImageBase64 reads an image from the clipboard and returns it as a base64 string.
//Returns the base64 string or an error if the operation fails.
func (f *ClipboardService) ReadImageBase64() (string, error) {
	err := clipboard.Init()
	if err != nil {
		return "", fmt.Errorf("failed to initialize clipboard: %v", err)
	}
	img := clipboard.Read(clipboard.FmtImage)
	base64Str := base64.StdEncoding.EncodeToString(img)
	return base64Str, nil
}

//WriteText writes a text string to the system clipboard.
//Returns an error if the operation fails.
func (f *ClipboardService) WriteText(text string) error {
	err := clipboard.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize clipboard: %v", err)
	}
	clipboard.Write(clipboard.FmtText, []byte(text))
	return nil
}

//ReadText reads a text string from the system clipboard.
//Returns the text string or an error if the operation fails.
func (f *ClipboardService) ReadText() (string, error) {
	err := clipboard.Init()
	if err != nil {
		return "", fmt.Errorf("failed to initialize clipboard: %v", err)
	}
	text := clipboard.Read(clipboard.FmtText)
	return string(text), nil
}
