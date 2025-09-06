package main

import (
	"clustta/internal/settings"
	"clustta/services"
	"embed"
	"fmt"
	"log"
	"log/slog"
	"os/exec"
	"runtime"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

var app *application.App

// InitializeFullscreenMonitoring
func InitializeFullscreenMonitoring() {
	err := StartFullscreenMonitoring("Clustta")
	if err != nil {
		fmt.Printf("Warning: Could not start fullscreen monitoring: %v\n", err)
	} else {
		fmt.Println("Manually started fullscreen monitoring for window 'main'")
	}
}

var encryptionKey = [32]byte{
	0x1e, 0x1f, 0x1c, 0x1d, 0x1a, 0x1b, 0x18, 0x19,
	0x16, 0x17, 0x14, 0x15, 0x12, 0x13, 0x10, 0x11,
	0x0e, 0x0f, 0x0c, 0x0d, 0x0a, 0x0b, 0x08, 0x09,
	0x06, 0x07, 0x04, 0x05, 0x02, 0x03, 0x00, 0x01,
}

func main() {

	logFile, err := settings.GetLogPath()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := NewFileLogger(logFile)
	if err != nil {
		log.Fatal(err)
	}

	var singleInstanceOpt *application.SingleInstanceOptions = nil

	if runtime.GOOS != "darwin" {
		singleInstanceOpt = &application.SingleInstanceOptions{
			UniqueID:      "com.clustta.clustta.single-instance",
			EncryptionKey: encryptionKey,
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				window := application.Get().GetWindowByName("main")
				if window != nil {
					window.Show()
					window.Focus()
				}
				log.Printf("Second instance launched with args: %v\n", data.Args)
				log.Printf("Working directory: %s\n", data.WorkingDir)
				if data.AdditionalData != nil {
					log.Printf("Additional data: %v\n", data.AdditionalData)
				}
			},
			AdditionalData: map[string]string{
				"launchtime": time.Now().Local().String(),
			},
		}
	}

	app = application.New(application.Options{
		Name:           "clustta",
		Description:    "File management and Collaboration tool",
		LogLevel:       slog.LevelError,
		Logger:         logger,
		SingleInstance: singleInstanceOpt,
		Services: []application.Service{
			application.NewService(&services.FSService{}),
			application.NewService(&services.LogService{}),
			application.NewService(&services.ClipboardService{}),
			application.NewService(&services.DialogService{}),
			application.NewService(&services.AuthService{}),
			application.NewService(&services.AccountService{}),
			application.NewService(&services.CheckpointService{}),
			application.NewService(&services.DependencyTypeService{}),
			application.NewService(&services.EntityService{}),
			application.NewService(&services.ImportService{}),
			application.NewService(&services.ProjectService{}),
			application.NewService(&services.SettingsService{}),
			application.NewService(&services.StatusService{}),
			application.NewService(&services.SyncService{}),
			application.NewService(&services.UserService{}),
			application.NewService(&services.TagService{}),
			application.NewService(&services.TaskService{}),
			application.NewService(&services.TemplateService{}),
			application.NewService(&services.TrashService{}),
			application.NewService(&services.WorkflowService{}),
			application.NewService(&services.AppService{}),
			application.NewService(&services.StudioService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	menu := application.NewMenu()

	openURLDefault := func(url string) {
		var cmd string
		var args []string

		switch runtime.GOOS {
		case "windows":
			cmd = "cmd"
			args = []string{"/c", "start", url}
		case "darwin":
			cmd = "open"
			args = []string{url}
		default:
			cmd = "xdg-open"
			args = []string{url}
		}

		exec.Command(cmd, args...).Start()
	}

	fileMenu := menu.AddSubmenu("File")
	fileMenu.Add("Quit").SetAccelerator("Cmd+Q").OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	helpMenu := menu.AddSubmenu("Help")
	helpMenu.Add("Learn More").OnClick(func(ctx *application.Context) {
		openURLDefault("https://docs.clustta.com")
	})

	app.SetMenu(menu)

	frameless := runtime.GOOS != "darwin"

	window := app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Name:              "main",
		Title:             "Clustta",
		Frameless:         frameless,
		Height:            734,
		Width:             1020,
		MinHeight:         734,
		MinWidth:          1020,
		EnableDragAndDrop: true,
		BackgroundColour:  application.NewRGB(0, 0, 0),

		Mac: application.MacWindow{
			Backdrop: application.MacBackdropTranslucent,
			TitleBar: application.MacTitleBar{
				AppearsTransparent:   true,
				Hide:                 false,
				HideTitle:            true,
				FullSizeContent:      true,
				UseToolbar:           true,
				HideToolbarSeparator: false,
				ToolbarStyle:         application.MacToolbarStyleAutomatic,
			},
		},
		Windows: application.WindowsWindow{
			DisableFramelessWindowDecorations: false,
		},
		KeyBindings: map[string]func(window *application.WebviewWindow){
			"F2": func(window *application.WebviewWindow) {
				app.EmitEvent("rename-item")
			},
			"ctrl+F2": func(window *application.WebviewWindow) {
				app.EmitEvent("edit-item")
			},
			"F3": func(window *application.WebviewWindow) {
				app.EmitEvent("search")
			},
			"F5": func(window *application.WebviewWindow) {
				app.EmitEvent("reload-view")
			},
			"return": func(window *application.WebviewWindow) {
				app.EmitEvent("enter-item")
			},
			"shift+delete": func(window *application.WebviewWindow) {
				app.EmitEvent("delete-item")
			},
			"delete": func(window *application.WebviewWindow) {
				app.EmitEvent("free-item-space")
			},
			"ctrl+n": func(window *application.WebviewWindow) {
				app.EmitEvent("new-project")
			},
			"ctrl+s": func(window *application.WebviewWindow) {
				app.EmitEvent("sync-project")
			},
			"ctrl+shift+c": func(window *application.WebviewWindow) {
				app.EmitEvent("add-checkpoint")
			},
			"ctrl+k": func(window *application.WebviewWindow) {
				app.EmitEvent("new-collection")
			},
			"ctrl+t": func(window *application.WebviewWindow) {
				app.EmitEvent("new-task")
			},
			"ctrl+l": func(window *application.WebviewWindow) {
				app.EmitEvent("new-web-link")
			},
			"ctrl+g": func(window *application.WebviewWindow) {
				app.EmitEvent("group-items")
			},
			"ctrl+c": func(window *application.WebviewWindow) {
				app.EmitEvent("copy-items")
			},
			"ctrl+x": func(window *application.WebviewWindow) {
				app.EmitEvent("cut-items")
			},
			"ctrl+v": func(window *application.WebviewWindow) {
				app.EmitEvent("paste-items")
			},
			"ctrl+d": func(window *application.WebviewWindow) {
				app.EmitEvent("duplicate-task")
			},
		},
		BackgroundType: application.BackgroundTypeTransparent,
		URL:            "/",
	})

	window.OnWindowEvent(events.Common.WindowFilesDropped, func(event *application.WindowEvent) {
		//TODO
	})

	window.OnWindowEvent(events.Common.WindowFocus, func(event *application.WindowEvent) {
		app.EmitEvent("window-focused", nil)
	})

	window.OnWindowEvent(events.Windows.WindowDragOver, func(event *application.WindowEvent) {
		//TODO
	})

	// Register a hook to hide the window when the window is closing
	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		window.Hide()
		e.Cancel()
	})

	if runtime.GOOS == "darwin" {

		//InitializeFullscreenMonitoring for detecting fullscreen events
		app.OnEvent("frontend-ready", func(event *application.CustomEvent) {
			InitializeFullscreenMonitoring()
		})

		//Check if the user settings file exists and initialize bookmarks
		// This is done to ensure that bookmarks are available when the app starts
		// This is especially useful for macOS where the app may not have been run before
		// and the user settings file may not have been created yet.
		settingsFile, err := settings.GetUserSettingsPath()
		if err != nil {
			log.Printf("Failed to get user settings path: %v", err)
		}

		if settingsFile != "" {
			err = settings.InitializeBookmarks()
			if err != nil {
				log.Printf("Failed to initialize bookmarks: %v", err)
			}
		}
	}

	// Run the application. This blocks until the application has been exited.
	err = app.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
