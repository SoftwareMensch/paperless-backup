.PHONY: build install clean test uninstall

BINARY_NAME=paperless-backup
INSTALL_PATH=/usr/local/bin
SERVICE_PATH=/etc/systemd/system

build:
	go build -o $(BINARY_NAME) ./cmd/paperless-backup && \
	strip $(BINARY_NAME)

install: build
	@echo "Installing binary..."
	sudo install -m 0755 $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installing systemd service files..."
	sudo install -m 0644 systemd/paperless-backup.service $(SERVICE_PATH)/paperless-backup.service
	sudo install -m 0644 systemd/paperless-backup.timer $(SERVICE_PATH)/paperless-backup.timer
	@echo "Reloading systemd daemon..."
	sudo systemctl daemon-reload
	@echo ""
	@echo "Installation complete!"
	@echo "To enable and start the timer:"
	@echo "  sudo systemctl enable --now paperless-backup.timer"
	@echo ""
	@echo "To check timer status:"
	@echo "  sudo systemctl status paperless-backup.timer"
	@echo "  systemctl list-timers paperless-backup.timer"

clean:
	-unlink $(BINARY_NAME)
	go clean

test:
	go test -count=1 -v ./...

run: build
	sudo PAPERLESS_BACKUP_ALLOW_DIRECT=1 ./$(BINARY_NAME)

uninstall:
	@echo "Stopping and disabling timer..."
	-sudo systemctl stop paperless-backup.timer
	-sudo systemctl disable paperless-backup.timer
	@echo "Removing systemd service files..."
	-sudo unlink $(SERVICE_PATH)/paperless-backup.service
	-sudo unlink $(SERVICE_PATH)/paperless-backup.timer
	@echo "Reloading systemd daemon..."
	sudo systemctl daemon-reload
	@echo "Removing binary..."
	-sudo unlink $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstallation complete!"

