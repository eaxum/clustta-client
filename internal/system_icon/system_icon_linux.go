package system_icon

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetExtensionIcon returns a PNG image of the system icon for the given file extension.
// It resolves the MIME type via the shared MIME database, then searches the active icon theme.
func GetExtensionIcon(extension string) ([]byte, error) {
	if len(extension) == 0 {
		return nil, fmt.Errorf("empty extension")
	}
	if extension[0] == '.' {
		extension = extension[1:]
	}

	mimeType, err := mimeTypeForExtension(extension)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve MIME type for .%s: %w", extension, err)
	}

	iconNames := iconNamesForMime(mimeType)
	themes := themeSearchOrder()
	baseDirs := iconBaseDirs()

	// Preferred sizes in descending order for crisp icons
	pngSizes := []string{"256x256", "128x128", "64x64", "48x48", "32x32"}

	for _, name := range iconNames {
		for _, theme := range themes {
			for _, baseDir := range baseDirs {
				// Try PNG at fixed sizes first
				for _, size := range pngSizes {
					p := filepath.Join(baseDir, theme, size, "mimetypes", name+".png")
					if data, err := os.ReadFile(p); err == nil {
						return data, nil
					}
				}
				// Try SVG and convert to PNG
				svgPath := filepath.Join(baseDir, theme, "scalable", "mimetypes", name+".svg")
				if data, err := svgToPNG(svgPath); err == nil {
					return data, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no icon found for .%s (MIME: %s)", extension, mimeType)
}

// mimeTypeForExtension resolves a file extension to its MIME type using the shared MIME database.
func mimeTypeForExtension(ext string) (string, error) {
	ext = strings.ToLower(ext)
	globPattern := "*." + ext

	// Try globs2 first (weighted, more reliable), then globs
	for _, name := range []string{"globs2", "globs"} {
		for _, baseDir := range mimeBaseDirs() {
			path := filepath.Join(baseDir, "mime", name)
			f, err := os.Open(path)
			if err != nil {
				continue
			}
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				if len(line) == 0 || line[0] == '#' {
					continue
				}
				// globs2 format: "weight:mime:glob"  globs format: "mime:glob"
				parts := strings.Split(line, ":")
				var mime, glob string
				if len(parts) == 3 {
					mime, glob = parts[1], parts[2]
				} else if len(parts) == 2 {
					mime, glob = parts[0], parts[1]
				} else {
					continue
				}
				if strings.EqualFold(glob, globPattern) {
					f.Close()
					return mime, nil
				}
			}
			f.Close()
		}
	}
	return "", fmt.Errorf("no MIME type found for extension .%s", ext)
}

// iconNamesForMime returns icon name candidates for a MIME type in priority order.
func iconNamesForMime(mimeType string) []string {
	// Replace "/" with "-" per freedesktop Icon Naming Spec
	primary := strings.ReplaceAll(mimeType, "/", "-")
	parts := strings.SplitN(mimeType, "/", 2)
	generic := parts[0] + "-x-generic"

	names := []string{primary}
	if generic != primary {
		names = append(names, generic)
	}
	return names
}

// themeSearchOrder returns the icon theme chain (active theme + inherited themes).
func themeSearchOrder() []string {
	theme := detectIconTheme()
	visited := map[string]bool{}
	var chain []string
	buildThemeChain(theme, iconBaseDirs(), visited, &chain)

	// Always include hicolor as the final fallback
	if !visited["hicolor"] {
		chain = append(chain, "hicolor")
	}
	return chain
}

// buildThemeChain recursively resolves the Inherits chain from index.theme files.
func buildThemeChain(theme string, baseDirs []string, visited map[string]bool, chain *[]string) {
	if visited[theme] || theme == "" {
		return
	}
	visited[theme] = true
	*chain = append(*chain, theme)

	for _, baseDir := range baseDirs {
		indexPath := filepath.Join(baseDir, theme, "index.theme")
		f, err := os.Open(indexPath)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "Inherits=") {
				parents := strings.Split(strings.TrimPrefix(line, "Inherits="), ",")
				f.Close()
				for _, parent := range parents {
					buildThemeChain(strings.TrimSpace(parent), baseDirs, visited, chain)
				}
				return
			}
		}
		f.Close()
		return
	}
}

// detectIconTheme determines the active icon theme.
func detectIconTheme() string {
	// Try gsettings (GNOME/Ubuntu)
	out, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "icon-theme").Output()
	if err == nil {
		theme := strings.Trim(strings.TrimSpace(string(out)), "'\"")
		if theme != "" {
			return theme
		}
	}

	// Try GTK3/4 config files
	home, _ := os.UserHomeDir()
	for _, rel := range []string{".config/gtk-4.0/settings.ini", ".config/gtk-3.0/settings.ini"} {
		path := filepath.Join(home, rel)
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "gtk-icon-theme-name=") {
				f.Close()
				return strings.TrimSpace(strings.TrimPrefix(line, "gtk-icon-theme-name="))
			}
		}
		f.Close()
	}

	return "hicolor"
}

// iconBaseDirs returns directories to search for icon themes, per the freedesktop spec.
func iconBaseDirs() []string {
	var dirs []string

	// User-specific icon dirs
	home, _ := os.UserHomeDir()
	if home != "" {
		dirs = append(dirs, filepath.Join(home, ".icons"))
		dirs = append(dirs, filepath.Join(home, ".local", "share", "icons"))
	}

	// System dirs from XDG_DATA_DIRS
	xdgDataDirs := os.Getenv("XDG_DATA_DIRS")
	if xdgDataDirs == "" {
		xdgDataDirs = "/usr/local/share:/usr/share"
	}
	for _, d := range strings.Split(xdgDataDirs, ":") {
		d = strings.TrimSpace(d)
		if d != "" {
			dirs = append(dirs, filepath.Join(d, "icons"))
		}
	}

	return dirs
}

// mimeBaseDirs returns directories to search for the shared MIME database.
func mimeBaseDirs() []string {
	var dirs []string
	home, _ := os.UserHomeDir()
	if home != "" {
		dirs = append(dirs, filepath.Join(home, ".local", "share"))
	}
	xdgDataDirs := os.Getenv("XDG_DATA_DIRS")
	if xdgDataDirs == "" {
		xdgDataDirs = "/usr/local/share:/usr/share"
	}
	for _, d := range strings.Split(xdgDataDirs, ":") {
		d = strings.TrimSpace(d)
		if d != "" {
			dirs = append(dirs, d)
		}
	}
	return dirs
}

// svgToPNG attempts to convert an SVG file to 256x256 PNG using rsvg-convert.
func svgToPNG(svgPath string) ([]byte, error) {
	if _, err := os.Stat(svgPath); err != nil {
		return nil, err
	}

	rsvg, err := exec.LookPath("rsvg-convert")
	if err != nil {
		return nil, fmt.Errorf("rsvg-convert not available")
	}

	var buf bytes.Buffer
	cmd := exec.Command(rsvg, "-w", "256", "-h", "256", svgPath)
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("rsvg-convert failed: %w", err)
	}
	if buf.Len() == 0 {
		return nil, fmt.Errorf("rsvg-convert produced empty output")
	}
	return buf.Bytes(), nil
}
