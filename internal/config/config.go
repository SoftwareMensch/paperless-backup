package config

// Config holds all configuration for the paperless backup tool
type Config struct {
	BackupDir        string
	LogFile          string
	LockFile         string
	MaxBackupAgeDays int
	RequiredSpaceMB  int64
	PaperlessService string
	DataVolume       string
	MediaVolume      string
	RedisVolume      string
}

// Default returns a Config with default values
func Default() *Config {
	return &Config{
		BackupDir:        "/var/local/paperless-ngx/backups",
		LogFile:          "backup.log",
		LockFile:         "backup.lock",
		MaxBackupAgeDays: 3,
		RequiredSpaceMB:  10000,
		PaperlessService: "paperless-ngx.service",
		DataVolume:       "paperless-ngx_data",
		MediaVolume:      "paperless-ngx_media",
		RedisVolume:      "paperless-ngx_redisdata",
	}
}

