package backup

import (
	"os"
	"path/filepath"
	"testing"

	"paperless-backup/internal/config"
	"paperless-backup/internal/logger"
)

func TestNew(t *testing.T) {
	cfg := config.Default()
	backup, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	expectedLockPath := filepath.Join(cfg.BackupDir, cfg.LockFile)
	if backup.lockPath != expectedLockPath {
		t.Errorf("Expected lockPath %s, got %s", expectedLockPath, backup.lockPath)
	}

	expectedLogPath := filepath.Join(cfg.BackupDir, cfg.LogFile)
	if backup.logPath != expectedLogPath {
		t.Errorf("Expected logPath %s, got %s", expectedLogPath, backup.logPath)
	}
}

func TestCheckLock(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	lockPath := filepath.Join(tmpDir, "test.lock")

	log, _ := logger.New(logPath)
	defer log.Close()

	cfg := config.Default()
	backup := &Backup{
		config:   cfg,
		lockPath: lockPath,
		logger:   log,
	}

	// First call should succeed (no lock exists)
	backup.checkLock()

	// Verify lock was created
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Error("Lock file should be created")
	}

	// Cleanup for next test
	os.Remove(lockPath)
}

func TestGetVolumePath(t *testing.T) {
	// Note: This test requires Docker to be available
	t.Skip("Skipping GetVolumePath test - requires Docker")

	// This is how you would test it with Docker available:
	// tmpDir := t.TempDir()
	// logPath := filepath.Join(tmpDir, "test.log")
	// logger, _ := NewLogger(logPath)
	// defer logger.Close()
	// backup := &Backup{workDir: tmpDir, logger: logger}
	// path := backup.GetVolumePath("some_volume")
	// if path == "" {
	//     t.Error("Expected non-empty path")
	// }
}

func TestBackupSetup(t *testing.T) {
	tmpDir := t.TempDir()
	
	cfg := &config.Config{
		BackupDir:        tmpDir,
		LogFile:          "test.log",
		LockFile:         "test.lock",
		MaxBackupAgeDays: 30,
		RequiredSpaceMB:  1000,
		PaperlessService: "paperless-ngx.service",
		DataVolume:       "paperless-ngx_data",
		MediaVolume:      "paperless-ngx_media",
		RedisVolume:      "paperless-ngx_redisdata",
	}
	
	backup, _ := New(cfg)

	err := backup.Setup()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer backup.Cleanup()

	// Verify backup directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Backup directory should be created")
	}

	// Verify logger was initialized
	if backup.logger == nil {
		t.Error("Logger should be initialized")
	}
}

func TestCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")
	logPath := filepath.Join(tmpDir, "test.log")
	
	log, _ := logger.New(logPath)

	cfg := config.Default()
	backup := &Backup{
		config:   cfg,
		lockPath: lockPath,
		logger:   log,
	}

	// Create lock file
	os.WriteFile(lockPath, []byte{}, 0644)

	// Run cleanup
	backup.Cleanup()

	// Verify lock file was removed
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Error("Lock file should be removed")
	}
}

func TestCleanupWithServiceRestore(t *testing.T) {
	// Note: This would attempt to restart the service
	// We skip actual execution in tests
	t.Skip("Skipping service restore test - requires systemd")
}

