package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := New(logPath)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	if logger.fileHandle == nil {
		t.Error("Expected file handle to be initialized")
	}
}

func TestLoggerLog(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := New(logPath)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	// Write a test log message
	logger.Log("INFO", "test message")

	// Read the log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "[INFO]") {
		t.Error("Log should contain [INFO] level")
	}
	if !strings.Contains(logContent, "test message") {
		t.Error("Log should contain test message")
	}
}

func TestLoggerLogf(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := New(logPath)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	// Write a formatted log message
	logger.Logf("WARN", "test %s %d", "formatted", 123)

	// Read the log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "[WARN]") {
		t.Error("Log should contain [WARN] level")
	}
	if !strings.Contains(logContent, "test formatted 123") {
		t.Error("Log should contain formatted message")
	}
}

func TestLoggerClose(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := New(logPath)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	logger.Log("INFO", "before close")
	logger.Close()

	// Verify file was written before close
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file should exist after close")
	}
}

