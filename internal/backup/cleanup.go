package backup

import (
	"os"
	"path/filepath"
	"sort"
	"time"
)

// FileInfo holds backup file metadata
type FileInfo struct {
	Path    string
	ModTime time.Time
}

// cleanupOldBackups removes backups older than retention period (always keeps at least one)
func (b *Backup) cleanupOldBackups() {
	b.logger.Logf("INFO", "Cleaning up backups older than %d days...", b.config.MaxBackupAgeDays)

	entries, err := os.ReadDir(b.config.BackupDir)
	if err != nil {
		b.logger.Log("INFO", "Failed to read backup directory")
		return
	}

	var allBackups []FileInfo
	var oldBackups []FileInfo
	cutoffTime := time.Now().AddDate(0, 0, -b.config.MaxBackupAgeDays)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if it's a tar.gz file
		name := entry.Name()
		matched, _ := filepath.Match("*.tar.gz", name)
		if !matched {
			continue
		}

		fullPath := filepath.Join(b.config.BackupDir, name)
		info, err := entry.Info()
		if err != nil {
			continue
		}

		backupInfo := FileInfo{
			Path:    fullPath,
			ModTime: info.ModTime(),
		}

		allBackups = append(allBackups, backupInfo)

		// Check if backup is older than cutoff
		if info.ModTime().Before(cutoffTime) {
			oldBackups = append(oldBackups, backupInfo)
		}
	}

	totalBackups := len(allBackups)

	if len(oldBackups) == 0 {
		b.logger.Log("INFO", "No old backups to delete")
		b.logger.Logf("INFO", "Total backups: %d", totalBackups)
		return
	}

	// Sort old backups by modification time (newest first)
	sort.Slice(oldBackups, func(i, j int) bool {
		return oldBackups[i].ModTime.After(oldBackups[j].ModTime)
	})

	// Check if we would delete all backups
	backupsToDelete := len(oldBackups)
	if backupsToDelete >= totalBackups {
		b.logger.Logf("INFO", "All backups are older than %d days - keeping the most recent one", b.config.MaxBackupAgeDays)
		// Remove first (most recent) from deletion list
		if len(oldBackups) > 0 {
			oldBackups = oldBackups[1:]
		}
	}

	// Delete old backups
	deletedCount := 0
	for _, backup := range oldBackups {
		b.logger.Logf("INFO", "Deleting old backup: %s", filepath.Base(backup.Path))
		if err := os.Remove(backup.Path); err != nil {
			b.logger.Logf("WARN", "Failed to delete %s: %v", backup.Path, err)
		} else {
			deletedCount++
		}
	}

	if deletedCount > 0 {
		b.logger.Logf("INFO", "Deleted %d old backup(s)", deletedCount)
	} else {
		b.logger.Log("INFO", "No old backups to delete (keeping at least one backup)")
	}

	// Count remaining backups
	remainingBackups := totalBackups - deletedCount
	b.logger.Logf("INFO", "Total backups: %d", remainingBackups)
}

