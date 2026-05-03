package main

import (
	"clustta/internal/bridge"
	"clustta/internal/repository"
	"clustta/internal/settings"
	"clustta/services"
	"crypto/sha256"
	"embed"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var trayIcon []byte

var app *application.App

// handleDeepLink processes a clustta:// URL and emits the appropriate event.
func handleDeepLink(rawURL string) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		log.Printf("Failed to parse deep link URL: %v", err)
		return
	}
	if parsed.Scheme != "clustta" {
		return
	}
	log.Printf("Deep link received: %s", rawURL)

	if app != nil {
		window, _ := application.Get().Window.GetByName("main")
		if window != nil {
			window.Show()
			window.Focus()
		}
		app.Event.Emit("deep-link", rawURL)
	} else {
		services.SetPendingDeepLink(rawURL)
	}
}

// InitializeFullscreenMonitoring starts fullscreen state monitoring for the application window.
// Logs warnings if monitoring cannot be started.
func InitializeFullscreenMonitoring() {
	err := StartFullscreenMonitoring("Clustta")
	if err != nil {
		fmt.Printf("Warning: Could not start fullscreen monitoring: %v\n", err)
	} else {
		fmt.Println("Manually started fullscreen monitoring for window 'main'")
	}
}

// generateEncryptionKey generates a machine-specific encryption key for single-instance IPC.
func generateEncryptionKey() [32]byte {
	hostname, _ := os.Hostname()
	u, _ := user.Current()
	username := ""
	if u != nil {
		username = u.Username
	}
	seed := "clustta-single-instance:" + hostname + ":" + username
	return sha256.Sum256([]byte(seed))
}

var fsServiceInstance *services.FSService

// createFSService initializes FSService with a file system watcher.
// Stores reference globally and starts watching for file system events.
func createFSService() *services.FSService {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create file watcher: %v", err)
		return &services.FSService{}
	}

	fsService := &services.FSService{
		Watcher: watcher,
	}

	fsServiceInstance = fsService
	fsService.StartWatching()

	return fsService
}

func main() {

	logFile, err := settings.GetLogPath()
	if err != nil {
		log.Fatal(err)
	}

	// Create console writer to capture logs for frontend debug console
	consoleWriter, err := NewConsoleWriter(logFile)
	if err != nil {
		log.Fatal(err)
	}

	// Redirect Go's standard log package to console writer
	log.SetOutput(consoleWriter)
	log.SetFlags(0) // Remove default timestamp since we add our own

	// Capture stdout (fmt.Printf etc.) and send to debug console
	consoleWriter.CaptureStdout()

	logger, err := NewFileLogger(logFile)
	if err != nil {
		log.Fatal(err)
	}

	// Start the bridge HTTP server if the user has enabled it
	bridgeEnabled, err := settings.GetBridgeEnabled()
	log.Printf("Bridge enabled: %v (err: %v)", bridgeEnabled, err)
	if err == nil && bridgeEnabled {
		bridge.Start()
	}
	defer bridge.Stop()

	var singleInstanceOpt *application.SingleInstanceOptions = nil

	// macOS handles single-instance natively via app bundle.
	// Flatpak sandboxes the app per D-Bus session, making this redundant.
	isFlatpak := false
	if _, err := os.Stat("/.flatpak-info"); err == nil {
		isFlatpak = true
	}

	if runtime.GOOS != "darwin" && !isFlatpak {
		singleInstanceOpt = &application.SingleInstanceOptions{
			UniqueID:      "com.clustta.clustta.single-instance",
			EncryptionKey: generateEncryptionKey(),
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				window, _ := application.Get().Window.GetByName("main")
				if window != nil {
					window.Show()
					window.Focus()
				}
				log.Printf("Second instance launched with args: %v\n", data.Args)
				log.Printf("Working directory: %s\n", data.WorkingDir)
				if data.AdditionalData != nil {
					log.Printf("Additional data: %v\n", data.AdditionalData)
				}

				// Forward .clst file path from second instance to frontend
				for _, arg := range data.Args {
					if strings.HasSuffix(strings.ToLower(arg), ".clst") {
						log.Printf("Second instance opened with project file: %s", arg)
						app.Event.Emit("open-project-file", arg)
						break
					}
					if strings.HasPrefix(strings.ToLower(arg), "clustta://") {
						handleDeepLink(arg)
						break
					}
				}
			},
			AdditionalData: map[string]string{
				"launchtime": time.Now().Local().String(),
			},
		}
	}

	app = application.New(application.Options{
		Name:             "clustta",
		Description:      "Version control, Asset management and Collaboration",
		LogLevel:         slog.LevelError,
		Logger:           logger,
		SingleInstance:   singleInstanceOpt,
		FileAssociations: []string{".clst"},
		Services: []application.Service{
			application.NewService(&services.AccountService{}),
			application.NewService(&services.AgentService{}),
			application.NewService(&services.AppService{}),
			application.NewService(&services.AssetService{}),
			application.NewService(&services.AuthService{}),
			application.NewService(&services.CheckpointService{}),
			application.NewService(&services.ClipboardService{}),
			application.NewService(&services.CollaboratorService{}),
			application.NewService(&services.CollectionService{}),
			application.NewService(&services.DependencyTypeService{}),
			application.NewService(&services.DeploymentService{}),
			application.NewService(&services.DialogService{}),
			application.NewService(&services.EntitlementService{}),
			application.NewService(createFSService()),
			application.NewService(&services.ImportService{}),
			application.NewService(&services.IntegrationService{}),
			application.NewService(&services.LogService{}),
			application.NewService(&services.ProfileService{}),
			application.NewService(&services.ProjectService{}),
			application.NewService(&services.SettingsService{}),
			application.NewService(&services.ShareService{}),
			application.NewService(&services.StatusService{}),
			application.NewService(&services.StudioService{}),
			application.NewService(&services.SyncService{}),
			application.NewService(&services.TagService{}),
			application.NewService(&services.TemplateService{}),
			application.NewService(&services.TrashService{}),
			application.NewService(&services.UserService{}),
			application.NewService(&services.WorkflowService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Listen for file open events (when user double-clicks a .clst file)
	// Buffer the path for the frontend to retrieve after initialization.
	app.Event.OnApplicationEvent(events.Common.ApplicationOpenedWithFile, func(event *application.ApplicationEvent) {
		filePath := event.Context().Filename()
		log.Printf("Application opened with file: %s", filePath)
		services.SetPendingOpenFile(filePath)
		app.Event.Emit("open-project-file", filePath)
	})

	// Listen for URL open events (deep links like clustta://invite?studio=...)
	// Handles cold launch on all platforms and macOS warm launch via Apple Events.
	app.Event.OnApplicationEvent(events.Common.ApplicationLaunchedWithUrl, func(event *application.ApplicationEvent) {
		rawURL := event.Context().URL()
		log.Printf("Application launched with URL: %s", rawURL)
		handleDeepLink(rawURL)
	})

	if fsServiceInstance != nil {
		fsServiceInstance.SetApp(app)
	}

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

	app.Menu.SetApplicationMenu(menu)

	frameless := runtime.GOOS != "darwin"

	var modifier string
	switch runtime.GOOS {
	case "darwin":
		modifier = "cmd"
	default:
		modifier = "ctrl"
	}

	keyBindings := map[string]func(window application.Window){
		"F2": func(window application.Window) {
			app.Event.Emit("rename-item")
		},
		"F3": func(window application.Window) {
			app.Event.Emit("search")
		},
		"F5": func(window application.Window) {
			app.Event.Emit("reload-view")
		},
		"return": func(window application.Window) {
			app.Event.Emit("enter-item")
		},
		"shift+delete": func(window application.Window) {
			app.Event.Emit("delete-item")
		},
		"delete": func(window application.Window) {
			app.Event.Emit("free-item-space")
		},
	}

	keyBindings[modifier+"+F2"] = func(window application.Window) {
		app.Event.Emit("edit-item")
	}
	keyBindings[modifier+"+n"] = func(window application.Window) {
		app.Event.Emit("new-project")
	}
	keyBindings[modifier+"+s"] = func(window application.Window) {
		app.Event.Emit("sync-project")
	}
	keyBindings[modifier+"+shift+c"] = func(window application.Window) {
		app.Event.Emit("add-checkpoint")
	}
	keyBindings[modifier+"+k"] = func(window application.Window) {
		app.Event.Emit("new-collection")
	}
	keyBindings[modifier+"+t"] = func(window application.Window) {
		app.Event.Emit("new-asset")
	}
	keyBindings[modifier+"+l"] = func(window application.Window) {
		app.Event.Emit("new-web-link")
	}
	keyBindings[modifier+"+g"] = func(window application.Window) {
		app.Event.Emit("group-items")
	}
	keyBindings[modifier+"+c"] = func(window application.Window) {
		app.Event.Emit("copy-items")
	}
	keyBindings[modifier+"+x"] = func(window application.Window) {
		app.Event.Emit("cut-items")
	}
	keyBindings[modifier+"+v"] = func(window application.Window) {
		app.Event.Emit("paste-items")
	}
	keyBindings[modifier+"+d"] = func(window application.Window) {
		app.Event.Emit("duplicate-asset")
	}

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            "Clustta",
		Frameless:        frameless,
		Height:           734,
		Width:            1020,
		MinHeight:        734,
		MinWidth:         1020,
		EnableFileDrop:   true,
		BackgroundColour: application.NewRGB(0, 0, 0),

		Mac: application.MacWindow{
			Backdrop: application.MacBackdropTranslucent,
			TitleBar: application.MacTitleBar{
				AppearsTransparent:   true,
				Hide:                 false,
				HideTitle:            true,
				FullSizeContent:      true,
				UseToolbar:           true,
				HideToolbarSeparator: false,
				ToolbarStyle:         application.MacToolbarStyleUnifiedCompact,
			},
		},
		Windows: application.WindowsWindow{
			DisableFramelessWindowDecorations: false,
		},
		KeyBindings:    keyBindings,
		BackgroundType: application.BackgroundTypeSolid,
		URL:            "/",
	})

	window.OnWindowEvent(events.Common.WindowFilesDropped, func(event *application.WindowEvent) {

		files := event.Context().DroppedFiles()
		details := event.Context().DropTargetDetails()

		log.Printf("Files dropped: %v", event)
		if details != nil {
			log.Printf("Drop target: id=%s, classes=%v, x=%d, y=%d",
				details.ElementID, details.ClassList, details.X, details.Y)
		}

		app.Event.Emit("files-dropped", map[string]any{
			"files":   files,
			"details": details,
		})
	})

	window.OnWindowEvent(events.Common.WindowFocus, func(event *application.WindowEvent) {
		app.Event.Emit("window-focused", nil)
	})

	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		minimizeOnClose, err := settings.GetMinimizeOnClose()
		if err != nil {
			log.Printf("Failed to get minimize on close setting: %v", err)
			minimizeOnClose = true
		}
		if minimizeOnClose {
			window.Hide()
			e.Cancel()
		}
	})

	// System tray icon with context menu
	systemTray := app.SystemTray.New()
	systemTray.SetIcon(trayIcon)

	trayMenu := app.NewMenu()
	trayMenu.Add("Show Clustta").OnClick(func(ctx *application.Context) {
		window.Show()
		window.Focus()
	})
	trayMenu.AddSeparator()
	trayMenu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})
	systemTray.SetMenu(trayMenu)

	systemTray.OnClick(func() {
		window.Show()
		window.Focus()
	})

	if runtime.GOOS == "darwin" {

		app.Event.On("frontend-ready", func(event *application.CustomEvent) {
			InitializeFullscreenMonitoring()
		})

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

	err = repository.InitializeDefaultTemplates(nil)
	if err != nil {
		log.Printf("Warning: Failed to initialize default templates: %v", err)
	} else {
		log.Println("Default project templates initialized successfully")
	}

	// Listen for frontend ready signal to flush buffered logs
	app.Event.On("debug-console-ready", func(event *application.CustomEvent) {
		if cw := GetConsoleWriter(); cw != nil {
			cw.SetFrontendReady()
		}
	})

	err = app.Run()

	if err != nil {
		log.Fatal(err)
	}
}
