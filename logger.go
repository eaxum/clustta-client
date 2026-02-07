package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type FileLogger struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
}

// ConsoleWriter captures log output and sends to frontend debug console.
type ConsoleWriter struct {
	file          *os.File
	buffer        []LogEntry
	mu            sync.Mutex
	frontendReady bool
}

// LogEntry represents a buffered log entry.
type LogEntry struct {
	Level   string
	Message string
	Time    time.Time
}

var consoleWriter *ConsoleWriter

// NewConsoleWriter creates a writer that logs to both file and frontend.
func NewConsoleWriter(filePath string) (*ConsoleWriter, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	cw := &ConsoleWriter{
		file:          file,
		buffer:        make([]LogEntry, 0),
		frontendReady: false,
	}
	consoleWriter = cw
	return cw, nil
}

// Write implements io.Writer interface for use with Go's log package.
func (cw *ConsoleWriter) Write(p []byte) (n int, err error) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	message := strings.TrimSpace(string(p))
	if message == "" {
		return len(p), nil
	}

	// Determine log level from message content
	level := "log"
	lowerMsg := strings.ToLower(message)
	if strings.Contains(lowerMsg, "error") || strings.Contains(lowerMsg, "fatal") {
		level = "error"
	} else if strings.Contains(lowerMsg, "warn") {
		level = "warn"
	} else if strings.Contains(lowerMsg, "info") {
		level = "info"
	}

	// Write to file
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s\n", timestamp, message)
	if cw.file != nil {
		cw.file.WriteString(logEntry)
	}

	// Send to frontend or buffer
	entry := LogEntry{
		Level:   level,
		Message: message,
		Time:    time.Now(),
	}

	if cw.frontendReady {
		cw.emitLog(entry)
	} else {
		cw.buffer = append(cw.buffer, entry)
	}

	return len(p), nil
}

// emitLog sends a log entry to the frontend.
func (cw *ConsoleWriter) emitLog(entry LogEntry) {
	app := application.Get()
	if app != nil {
		app.Event.Emit("backend-log", map[string]string{
			"level":   entry.Level,
			"message": entry.Message,
		})
	}
}

// SetFrontendReady marks frontend as ready and flushes buffered logs.
func (cw *ConsoleWriter) SetFrontendReady() {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	cw.frontendReady = true

	// Flush buffered logs
	for _, entry := range cw.buffer {
		cw.emitLog(entry)
	}
	cw.buffer = nil
}

// Close closes the log file.
func (cw *ConsoleWriter) Close() error {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	if cw.file != nil {
		return cw.file.Close()
	}
	return nil
}

// CaptureStdout redirects os.Stdout to also send to the ConsoleWriter.
func (cw *ConsoleWriter) CaptureStdout() {
	r, w, err := os.Pipe()
	if err != nil {
		return
	}

	// Keep the original stdout
	originalStdout := os.Stdout
	os.Stdout = w

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				// Write to original stdout
				originalStdout.Write(buf[:n])
				// Also write to ConsoleWriter (which sends to frontend)
				cw.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()
}

// GetConsoleWriter returns the global console writer instance.
func GetConsoleWriter() *ConsoleWriter {
	return consoleWriter
}

// NewFileLogger creates a new FileLogger instance
func NewFileLogger(filePath string) (*slog.Logger, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	opts := &slog.HandlerOptions{
		Level: slog.LevelError,
	}

	logger := slog.New(slog.NewTextHandler(file, opts))

	return logger, nil
}

// NewMultiWriter creates a writer that writes to both file and console writer.
func NewMultiWriter(file *os.File, cw *ConsoleWriter) io.Writer {
	return io.MultiWriter(file, cw)
}

// log writes a message to the file with the specified level
func (l *FileLogger) log(level, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s: %s\n", timestamp, level, message)

	if _, err := l.file.WriteString(logEntry); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to log file: %v\n", err)
	}
}

func (l *FileLogger) Print(message string) {
	l.log("PRINT", message)
}

func (l *FileLogger) Trace(message string) {
	l.log("TRACE", message)
}

func (l *FileLogger) Debug(message string) {
	l.log("DEBUG", message)
}

func (l *FileLogger) Info(message string) {
	l.log("INFO", message)
}

func (l *FileLogger) Warning(message string) {
	l.log("WARNING", message)
}

func (l *FileLogger) Error(message string) {
	l.log("ERROR", message)
}

func (l *FileLogger) Fatal(message string) {
	l.log("FATAL", message)
	os.Exit(1)
}

// Close closes the log file
func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}
