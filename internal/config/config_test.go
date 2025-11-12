package config

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{"BackupDir", cfg.BackupDir, "/var/local/paperless-ngx/backups"},
		{"LogFile", cfg.LogFile, "backup.log"},
		{"LockFile", cfg.LockFile, "backup.lock"},
		{"MaxBackupAgeDays", cfg.MaxBackupAgeDays, 3},
		{"RequiredSpaceMB", cfg.RequiredSpaceMB, int64(10000)},
		{"PaperlessService", cfg.PaperlessService, "paperless-ngx.service"},
		{"DataVolume", cfg.DataVolume, "paperless-ngx_data"},
		{"MediaVolume", cfg.MediaVolume, "paperless-ngx_media"},
		{"RedisVolume", cfg.RedisVolume, "paperless-ngx_redisdata"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.value, tt.expected)
			}
		})
	}
}

func TestConfigValuesReasonable(t *testing.T) {
	cfg := Default()

	if cfg.MaxBackupAgeDays <= 0 {
		t.Error("MaxBackupAgeDays should be positive")
	}

	if cfg.RequiredSpaceMB <= 0 {
		t.Error("RequiredSpaceMB should be positive")
	}

	if cfg.BackupDir == "" {
		t.Error("BackupDir should not be empty")
	}
}
