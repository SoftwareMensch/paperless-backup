# Paperless-NGX Backup (Go Version)

A robust backup tool for Paperless-NGX written in Go with minimal dependencies and comprehensive error handling.
It's focussed on my personal use case. If you find it useful for your own, feel free to use or modify.

## Features

- ✅ **Automated daily backups** via systemd timer
- 🗜️ **Compressed archives** (gzip) to save disk space
- 🧹 **Automatic cleanup** - Removes old backups (keeps at least one)
- 🔒 **Secure** - Restrictive file permissions (0600)
- 🛡️ **Systemd-only execution** - Binary only runs when invoked by systemd (security hardening)
- 📊 **Comprehensive logging** - Both to file and systemd journal
- 🔐 **Safe operations** - Stops service during backup, restores state after
- 🚫 **Concurrent run prevention** - Lock file mechanism
- 📦 **Single binary** - Easy deployment and updates

## Building

### Prerequisites
- Go 1.21 or higher
- Root access (for docker volume access)

### Build the binary
```bash
make build
```

### Install system-wide
```bash
make install
```

This will:
- Install the binary to `/usr/local/bin/paperless-backup`
- Install systemd service files to `/etc/systemd/system/`
- Reload the systemd daemon

After installation, enable and start the timer:
```bash
sudo systemctl enable --now paperless-backup.timer
```

Check the timer status:
```bash
systemctl list-timers paperless-backup.timer
```

### Uninstall
```bash
make uninstall
```

This will stop and disable the timer, remove all installed files.

## Usage

### Run manually (requires root)

**For security, the binary should only be executed by systemd:**
```bash
sudo systemctl start paperless-backup.service
```

**For testing/manual runs, use the override:**
```bash
sudo PAPERLESS_BACKUP_ALLOW_DIRECT=1 paperless-backup
```

Or use the Makefile:
```bash
make run
```

### Using with systemd

The service files are automatically installed with `make install`. To manually manage the service:

**Enable and start the timer:**
```bash
sudo systemctl enable --now paperless-backup.timer
```

**Check status:**
```bash
systemctl status paperless-backup.timer
systemctl list-timers paperless-backup.timer
```

**Run backup manually:**
```bash
sudo systemctl start paperless-backup.service
```

**View logs:**
```bash
journalctl -u paperless-backup.service
```

## Project Structure

The project follows idiomatic Go package structure:

```
paperless-backup/
├── cmd/
│   └── paperless-backup/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   ├── config.go           # Configuration management
│   │   └── config_test.go
│   ├── logger/
│   │   ├── logger.go           # Logging functionality
│   │   └── logger_test.go
│   ├── checks/
│   │   ├── checks.go           # Pre-flight validation checks
│   │   └── checks_test.go
│   ├── service/
│   │   ├── service.go          # Systemd service management
│   │   └── service_test.go
│   ├── archive/
│   │   ├── tar.go              # Tar.gz archive operations
│   │   └── tar_test.go
│   └── backup/
│       ├── backup.go           # Core backup orchestration
│       ├── backup_test.go
│       ├── cleanup.go          # Backup retention management
│       └── cleanup_test.go
├── systemd/
│   ├── paperless-backup.service # Systemd service unit
│   └── paperless-backup.timer   # Systemd timer unit
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Configuration

Configuration is managed in `internal/config/config.go` via the `Config` struct:

```go
cfg := config.Default()
// Default values:
// BackupDir:        "/var/local/paperless-ngx/backups"
// MaxBackupAgeDays: 30
// RequiredSpaceMB:  10000
// PaperlessService: "paperless-ngx.service"
// DataVolume:       "paperless-ngx_data"
// MediaVolume:      "paperless-ngx_media"
// RedisVolume:      "paperless-ngx_redisdata"
```

Modify the `Default()` function in `internal/config/config.go` and rebuild to change settings.

## Development

### Build
```bash
make build
```

### Clean
```bash
make clean
```

### Run (for testing)
```bash
make run
```

## Requirements

**System Dependencies:**
- `docker` - For volume inspection
- `systemctl` - For service management

**That's it!** All other functionality (compression, checksumming, file operations) is built-in.

## Logging

Logs to both:
- stdout (for systemd journal)
- `/var/local/paperless-ngx/backups/backup.log`

