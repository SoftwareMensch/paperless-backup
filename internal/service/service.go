package service

import (
	"fmt"
	"os/exec"
	"time"

	"paperless-backup/internal/logger"
)

// Manager handles systemd service state management
type Manager struct {
	logger      *logger.Logger
	serviceName string
	wasRunning  bool
}

// New creates a new service Manager
func New(logger *logger.Logger, serviceName string) *Manager {
	return &Manager{
		logger:      logger,
		serviceName: serviceName,
		wasRunning:  false,
	}
}

// Stop stops the service if it's running
func (m *Manager) Stop() {
	m.logger.Logf("INFO", "Checking %s state...", m.serviceName)

	cmd := exec.Command("systemctl", "is-active", "--quiet", m.serviceName)
	if err := cmd.Run(); err == nil {
		// Service is running
		m.logger.Logf("INFO", "%s is running - stopping for backup...", m.serviceName)
		m.wasRunning = true

		stopCmd := exec.Command("systemctl", "stop", m.serviceName)
		if err := stopCmd.Run(); err != nil {
			m.logger.ErrorExit(fmt.Sprintf("Failed to stop %s", m.serviceName))
		}

		m.logger.Logf("INFO", "%s stopped", m.serviceName)
		time.Sleep(2 * time.Second)
	} else {
		m.logger.Logf("INFO", "%s is already stopped", m.serviceName)
	}
}

// Restore restarts the service if it was running before
func (m *Manager) Restore() {
	if !m.wasRunning {
		return
	}

	m.logger.Logf("INFO", "Restoring %s to running state...", m.serviceName)
	cmd := exec.Command("systemctl", "start", m.serviceName)
	if err := cmd.Run(); err != nil {
		m.logger.Logf("WARN", "Failed to restart %s", m.serviceName)
	}
}

// WasRunning returns whether the service was running before being stopped
func (m *Manager) WasRunning() bool {
	return m.wasRunning
}

