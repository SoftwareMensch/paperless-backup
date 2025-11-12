package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"paperless-backup/internal/config"
	"paperless-backup/internal/logger"
)

func TestCleanupOldBackups(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	
	log, _ := logger.New(logPath)
	defer log.Close()

	cfg := config.Default()
	cfg.BackupDir = tmpDir
	backup := &Backup{
		config: cfg,
		logger: log,
	}

	// Create backups with different ages
	now := time.Now()
	
	// Recent backup (should be kept)
	recentFile := filepath.Join(tmpDir, "recent.tar.gz")
	os.WriteFile(recentFile, []byte("recent"), 0644)
	
	// Old backup (should be deleted)
	oldFile := filepath.Join(tmpDir, "old.tar.gz")
	os.WriteFile(oldFile, []byte("old"), 0644)
	oldTime := now.Add(-time.Duration(cfg.MaxBackupAgeDays+1) * 24 * time.Hour)
	os.Chtimes(oldFile, oldTime, oldTime)

	// Run cleanup
	backup.cleanupOldBackups()

	// Recent should still exist
	if _, err := os.Stat(recentFile); os.IsNotExist(err) {
		t.Error("Recent backup should not be deleted")
	}

	// Old should be deleted
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old backup should be deleted")
	}
}

func TestCleanupKeepsAtLeastOne(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	
	log, _ := logger.New(logPath)
	defer log.Close()

	cfg := config.Default()
	cfg.BackupDir = tmpDir
	backup := &Backup{
		config: cfg,
		logger: log,
	}

	// Create only old backups
	now := time.Now()
	oldTime := now.Add(-time.Duration(cfg.MaxBackupAgeDays+1) * 24 * time.Hour)
	
	oldest := filepath.Join(tmpDir, "oldest.tar.gz")
	os.WriteFile(oldest, []byte("oldest"), 0644)
	os.Chtimes(oldest, oldTime.Add(-48*time.Hour), oldTime.Add(-48*time.Hour))

	lessOld := filepath.Join(tmpDir, "less_old.tar.gz")
	os.WriteFile(lessOld, []byte("less old"), 0644)
	os.Chtimes(lessOld, oldTime, oldTime)

	// Run cleanup
	backup.cleanupOldBackups()

	// Count remaining backups
	entries, _ := os.ReadDir(tmpDir)
	backupCount := 0
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".gz" {
			backupCount++
		}
	}

	if backupCount < 1 {
		t.Error("Should keep at least one backup even if all are old")
	}

	// The most recent (less old) should be kept
	if _, err := os.Stat(lessOld); os.IsNotExist(err) {
		t.Error("Most recent backup should be kept")
	}
}

func TestCleanupNoBackups(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	
	log, _ := logger.New(logPath)
	defer log.Close()

	cfg := config.Default()
	cfg.BackupDir = tmpDir
	backup := &Backup{
		config: cfg,
		logger: log,
	}

	// Run cleanup with no backups (should not crash)
	backup.cleanupOldBackups()
}

func TestCleanupIgnoresNonBackupFiles(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	
	log, _ := logger.New(logPath)
	defer log.Close()

	cfg := config.Default()
	cfg.BackupDir = tmpDir
	backup := &Backup{
		config: cfg,
		logger: log,
	}

	// Create non-backup files
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("readme"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "backup.log"), []byte("log"), 0644)
	
	// Create old backup
	now := time.Now()
	oldTime := now.Add(-time.Duration(cfg.MaxBackupAgeDays+1) * 24 * time.Hour)
	oldBackup := filepath.Join(tmpDir, "old.tar.gz")
	os.WriteFile(oldBackup, []byte("old"), 0644)
	os.Chtimes(oldBackup, oldTime, oldTime)

	// Run cleanup
	backup.cleanupOldBackups()

	// Non-backup files should still exist
	if _, err := os.Stat(filepath.Join(tmpDir, "readme.txt")); os.IsNotExist(err) {
		t.Error("Non-backup files should not be deleted")
	}
}

