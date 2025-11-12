package checks

import (
	"fmt"
	"os/exec"
	"strings"

	"paperless-backup/internal/logger"

	"golang.org/x/sys/unix"
)

// Checker performs pre-flight validation checks
type Checker struct {
	logger    *logger.Logger
	workDir   string
	requiredMB int64
}

// New creates a new Checker instance
func New(logger *logger.Logger, workDir string, requiredMB int64) *Checker {
	return &Checker{
		logger:     logger,
		workDir:    workDir,
		requiredMB: requiredMB,
	}
}

// RequiredTools verifies required system tools are available
func (c *Checker) RequiredTools() {
	c.logger.Log("INFO", "Checking required system tools...")
	requiredTools := []string{"docker", "systemctl"}
	var missing []string

	for _, tool := range requiredTools {
		if _, err := exec.LookPath(tool); err != nil {
			missing = append(missing, tool)
		}
	}

	if len(missing) > 0 {
		c.logger.ErrorExit(fmt.Sprintf("Missing required tools: %s", strings.Join(missing, ", ")))
	}

	c.logger.Log("INFO", "All required system tools available")
}

// Docker verifies docker is running
func (c *Checker) Docker() {
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		c.logger.ErrorExit("Docker daemon is not running or not accessible")
	}
}

// DiskSpace verifies sufficient disk space is available for backup
func (c *Checker) DiskSpace() {
	var stat unix.Statfs_t
	if err := unix.Statfs(c.workDir, &stat); err != nil {
		c.logger.ErrorExit("Failed to check disk space")
	}

	// Available blocks * block size / 1024 / 1024 = Available MB
	availableMB := int64(stat.Bavail) * int64(stat.Bsize) / 1024 / 1024

	if availableMB < c.requiredMB {
		c.logger.ErrorExit(fmt.Sprintf("Insufficient disk space. Available: %dMB, Required: %dMB", availableMB, c.requiredMB))
	}

	c.logger.Logf("INFO", "Available disk space: %dMB", availableMB)
}

