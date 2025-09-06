package main

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

type FileLogger struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
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
