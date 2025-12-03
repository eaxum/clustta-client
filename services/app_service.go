package services

import (
	"context"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
)

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
// Gets the main window instance and minimizes it to the taskbar.
func (s *AppService) Minimize() {
	app := application.Get()
	if app != nil {
		window, _ := app.Window.GetByName("main")
		if window != nil {
			window.Minimise()
		}
	}
}
