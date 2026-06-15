//go:build darwin

package services

import (
	"os"
	"path/filepath"
)

// detectPackaging returns "mas" when the App Store receipt is present, else "direct".
func detectPackaging() string {
	exe, err := os.Executable()
	if err != nil {
		return "direct"
	}
	receipt := filepath.Join(filepath.Dir(exe), "..", "_MASReceipt", "receipt")
	if _, err := os.Stat(receipt); err == nil {
		return "mas"
	}
	return "direct"
}
