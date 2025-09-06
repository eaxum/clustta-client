package services

import (
	"encoding/base64"
	"fmt"

	"golang.design/x/clipboard"
)

type ClipboardService struct{}

func (f *ClipboardService) ReadImageBase64() (string, error) {
	err := clipboard.Init()
	if err != nil {
		return "", fmt.Errorf("failed to initialize clipboard: %v", err)
	}
	img := clipboard.Read(clipboard.FmtImage)
	// Convert to base64
	base64Str := base64.StdEncoding.EncodeToString(img)
	return base64Str, nil
}

func (f *ClipboardService) WriteText(text string) error {
	err := clipboard.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize clipboard: %v", err)
	}
	clipboard.Write(clipboard.FmtText, []byte(text))
	return nil
}

func (f *ClipboardService) ReadText() (string, error) {
	err := clipboard.Init()
	if err != nil {
		return "", fmt.Errorf("failed to initialize clipboard: %v", err)
	}
	text := clipboard.Read(clipboard.FmtText)
	return string(text), nil
}
