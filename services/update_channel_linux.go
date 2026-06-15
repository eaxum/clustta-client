//go:build linux

package services

import "os"

// detectPackaging returns "flathub" inside a Flatpak sandbox, else "direct".
func detectPackaging() string {
	if _, err := os.Stat("/.flatpak-info"); err == nil {
		return "flathub"
	}
	return "direct"
}
