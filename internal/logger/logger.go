package logger

import (
	"fmt"
	"os"
	"time"
)

// Logger handles logging to both stdout and file
type Logger struct {
	fileHandle *os.File
}

// New creates a new logger instance
func New(logPath string) (*Logger, error) {
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		fileHandle: logFile,
	}, nil
}

// Log writes a log message with timestamp and level
func (l *Logger) Log(level, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMsg := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, message)

	// Write to stdout
	fmt.Print(logMsg)

	// Write to file
	if l.fileHandle != nil {
		l.fileHandle.WriteString(logMsg)
	}
}

// Logf writes a formatted log message
func (l *Logger) Logf(level, format string, args ...interface{}) {
	l.Log(level, fmt.Sprintf(format, args...))
}

// ErrorExit logs an error and exits
func (l *Logger) ErrorExit(message string) {
	l.Log("ERROR", message)
	os.Exit(1)
}

// Close closes the log file handle
func (l *Logger) Close() {
	if l.fileHandle != nil {
		l.fileHandle.Close()
	}
}

