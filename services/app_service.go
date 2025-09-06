package services

import (
	"context"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type AppService struct {
}

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

func (s *AppService) OnStartup(ctx context.Context, options application.ServiceOptions) error {
	return nil
}

func (s *AppService) Quit() {
	app := application.Get()
	if app != nil {
		app.Quit()
	}
}

func (s *AppService) Hide() {
	app := application.Get()
	if app != nil {
		window := app.GetWindowByName("main")
		if window != nil {
			window.Hide()
		}
	}
}

func (s *AppService) Show() {
	app := application.Get()
	if app != nil {
		window := app.GetWindowByName("main")
		if window != nil {
			window.Show()
			window.Focus()
		}
	}
}

func (s *AppService) Minimize() {
	app := application.Get()
	if app != nil {
		window := app.GetWindowByName("main")
		if window != nil {
			window.Minimise()
		}
	}
}
