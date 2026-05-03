package services

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// pendingOpenFile stores a .clst file path received before the frontend was ready.
var (
	pendingOpenFile   string
	pendingOpenFileMu sync.Mutex
)

// pendingDeepLinkURL stores a clustta:// URL received before the frontend was ready.
var (
	pendingDeepLinkURL   string
	pendingDeepLinkURLMu sync.Mutex
)

// SystemInfo contains detailed system information.
type SystemInfo struct {
	OS        string `json:"os"`
	OSVersion string `json:"os_version"`
	Arch      string `json:"arch"`
}

type AppService struct {
}

// GetOS returns the operating system name.
// Detects the current OS and returns "windows", "darwin", "linux", or "unknown".
func (s *AppService) GetOS() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	default:
		return "unknown"
	}
}

// GetSystemInfo returns detailed system information including OS version.
func (s *AppService) GetSystemInfo() SystemInfo {
	info := SystemInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	switch runtime.GOOS {
	case "windows":
		info.OSVersion = getWindowsVersion()
	case "darwin":
		info.OSVersion = getMacOSVersion()
	case "linux":
		info.OSVersion = getLinuxVersion()
	default:
		info.OSVersion = "unknown"
	}

	return info
}

// getWindowsVersion returns the Windows version string.
func getWindowsVersion() string {
	cmd := exec.Command("cmd", "/c", "ver")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	version := strings.TrimSpace(string(output))
	// Parse "Microsoft Windows [Version 10.0.19045.3803]" format
	if strings.Contains(version, "[Version") {
		start := strings.Index(version, "[Version ")
		end := strings.Index(version, "]")
		if start != -1 && end != -1 {
			return version[start+9 : end]
		}
	}
	return version
}

// getMacOSVersion returns the macOS version string.
func getMacOSVersion() string {
	cmd := exec.Command("sw_vers", "-productVersion")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// getLinuxVersion returns the Linux distribution and version.
func getLinuxVersion() string {
	// Try to read /etc/os-release
	cmd := exec.Command("cat", "/etc/os-release")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(output), "\n")
	var prettyName string
	for _, line := range lines {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			prettyName = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
			break
		}
	}

	if prettyName != "" {
		return prettyName
	}

	// Fallback: try uname -r
	cmd = exec.Command("uname", "-r")
	output, err = cmd.Output()
	if err != nil {
		return "unknown"
	}
	return fmt.Sprintf("Linux %s", strings.TrimSpace(string(output)))
}

// ServiceStartup is called when the application starts.
// Currently a no-op placeholder for initialization logic.
func (s *AppService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	return nil
}

// Quit terminates the application.
// Gets the application instance and calls Quit to exit gracefully.
func (s *AppService) Quit() {
	app := application.Get()
	if app != nil {
		app.Quit()
	}
}

// Hide hides the main application window.
// Gets the main window instance and hides it from view.
func (s *AppService) Hide() {
	app := application.Get()
	if app != nil {
		window, _ := app.Window.GetByName("main")
		if window != nil {
			window.Hide()
		}
	}
}

// Show displays and focuses the main application window.
// Gets the main window instance, shows it, and brings it to focus.
func (s *AppService) Show() {
	app := application.Get()
	if app != nil {
		window, _ := app.Window.GetByName("main")
		if window != nil {
			window.Show()
			window.Focus()
		}
	}
}

// Minimize minimizes the main application window.
// Gets the main window instance and minimizes it to the assetbar.
func (s *AppService) Minimize() {
	app := application.Get()
	if app != nil {
		window, _ := app.Window.GetByName("main")
		if window != nil {
			window.Minimise()
		}
	}
}

// SetPendingOpenFile stores a .clst file path for the frontend to retrieve after initialization.
// Used when the app is launched via file association before the frontend is ready.
func SetPendingOpenFile(filePath string) {
	pendingOpenFileMu.Lock()
	defer pendingOpenFileMu.Unlock()
	pendingOpenFile = filePath
}

// GetPendingOpenFile returns and clears any buffered .clst file path.
// Called by the frontend after store initialization to handle cold-launch file opens.
func (s *AppService) GetPendingOpenFile() string {
	pendingOpenFileMu.Lock()
	defer pendingOpenFileMu.Unlock()
	filePath := pendingOpenFile
	pendingOpenFile = ""
	return filePath
}

// SetPendingDeepLink stores a clustta:// URL for the frontend to retrieve after initialization.
func SetPendingDeepLink(deepLink string) {
	pendingDeepLinkURLMu.Lock()
	defer pendingDeepLinkURLMu.Unlock()
	pendingDeepLinkURL = deepLink
}

// GetPendingDeepLink returns and clears any buffered deep link URL.
// Called by the frontend after store initialization to handle cold-launch deep links.
func (s *AppService) GetPendingDeepLink() string {
	pendingDeepLinkURLMu.Lock()
	defer pendingDeepLinkURLMu.Unlock()
	link := pendingDeepLinkURL
	pendingDeepLinkURL = ""
	return link
}
