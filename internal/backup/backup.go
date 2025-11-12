package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"paperless-backup/internal/archive"
	"paperless-backup/internal/checks"
	"paperless-backup/internal/config"
	"paperless-backup/internal/logger"
	"paperless-backup/internal/service"
)

// Backup orchestrates the complete backup process
type Backup struct {
	config         *config.Config
	logger         *logger.Logger
	checker        *checks.Checker
	serviceManager *service.Manager
	archiver       *archive.Creator
	lockPath       string
	logPath        string
	backupFile     string
}

// New creates a new Backup instance with the given configuration
func New(cfg *config.Config) (*Backup, error) {
	b := &Backup{
		config:   cfg,
		lockPath: filepath.Join(cfg.BackupDir, cfg.LockFile),
		logPath:  filepath.Join(cfg.BackupDir, cfg.LogFile),
	}
	return b, nil
}

// Setup initializes the backup environment
func (b *Backup) Setup() error {
	// Create backup directory if needed
	if err := os.MkdirAll(b.config.BackupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Setup logger
	log, err := logger.New(b.logPath)
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}
	b.logger = log

	// Initialize checker
	b.checker = checks.New(b.logger, b.config.BackupDir, b.config.RequiredSpaceMB)

	// Initialize service manager
	b.serviceManager = service.New(b.logger, b.config.PaperlessService)

	// Initialize archiver
	b.archiver = archive.New(b.logger)

	return nil
}

// Cleanup removes lock file and restores service state
func (b *Backup) Cleanup() {
	if b.serviceManager != nil {
		b.serviceManager.Restore()
	}

	os.Remove(b.lockPath)

	if b.logger != nil {
		b.logger.Close()
	}
}

// checkLock checks for existing lock file and creates one
func (b *Backup) checkLock() {
	if _, err := os.Stat(b.lockPath); err == nil {
		b.logger.ErrorExit(fmt.Sprintf("Backup already running (lock file exists: %s)", b.lockPath))
	}

	// Create lock file
	if err := os.WriteFile(b.lockPath, []byte{}, 0644); err != nil {
		b.logger.ErrorExit("Failed to create lock file")
	}
}

// getVolumePath inspects docker volume and returns mount point
func (b *Backup) getVolumePath(volume string) string {
	cmd := exec.Command("docker", "volume", "inspect", volume, "--format", "{{ .Mountpoint }}")
	output, err := cmd.Output()
	if err != nil {
		b.logger.ErrorExit(fmt.Sprintf("Failed to inspect %s volume", volume))
	}

	path := strings.TrimSpace(string(output))

	// Validate path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		b.logger.ErrorExit(fmt.Sprintf("Volume path does not exist: %s", path))
	}

	return path
}

// createBackup creates a timestamped backup archive
func (b *Backup) createBackup(dataPath, mediaPath, redisPath string) error {
	timestamp := time.Now().Format("20060102_150405")
	b.backupFile = filepath.Join(b.config.BackupDir, fmt.Sprintf("%s.tar.gz", timestamp))

	sourcePaths := []string{dataPath, mediaPath, redisPath}
	return b.archiver.Create(b.backupFile, sourcePaths)
}

// Run executes the complete backup process
func (b *Backup) Run() {
	b.logger.Log("INFO", "Starting paperless-ngx backup")

	// Pre-flight checks (root check is done in main before we get here)
	b.checkLock()
	b.checker.RequiredTools()
	b.checker.Docker()

	// Stop service if running
	b.serviceManager.Stop()

	// Get volume paths
	b.logger.Log("INFO", "Inspecting docker volumes...")
	dataPath := b.getVolumePath(b.config.DataVolume)
	mediaPath := b.getVolumePath(b.config.MediaVolume)
	redisPath := b.getVolumePath(b.config.RedisVolume)

	b.logger.Log("INFO", "Volume locations:")
	b.logger.Logf("INFO", "  - Data: %s", dataPath)
	b.logger.Logf("INFO", "  - Media: %s", mediaPath)
	b.logger.Logf("INFO", "  - Redis: %s", redisPath)

	// Check available disk space
	b.checker.DiskSpace()

	// Create compressed backup archive
	if err := b.createBackup(dataPath, mediaPath, redisPath); err != nil {
		b.logger.ErrorExit(err.Error())
	}

	// Verify backup integrity
	if err := b.archiver.Verify(b.backupFile); err != nil {
		b.logger.ErrorExit(err.Error())
	}

	// Remove old backups per retention policy
	b.cleanupOldBackups()

	b.logger.Log("INFO", "Backup completed successfully")
}

