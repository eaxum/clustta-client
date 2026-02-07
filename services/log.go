package services

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

type LogService struct{}

// LogError logs an error message to the application logger.
func (l *LogService) LogError(message string) {
	application.Get().Logger.Error(message)
}

// LogToConsole emits a log message to the frontend debug console.
func (l *LogService) LogToConsole(level string, message string) {
	application.Get().Event.Emit("backend-log", map[string]string{
		"level":   level,
		"message": message,
	})
}

// Info logs an info message to the frontend debug console.
func (l *LogService) Info(message string) {
	l.LogToConsole("info", message)
}

// Warn logs a warning message to the frontend debug console.
func (l *LogService) Warn(message string) {
	l.LogToConsole("warn", message)
}

// Error logs an error message to the frontend debug console.
func (l *LogService) Error(message string) {
	l.LogToConsole("error", message)
}

// Log logs a general message to the frontend debug console.
func (l *LogService) Log(message string) {
	l.LogToConsole("log", message)
}
