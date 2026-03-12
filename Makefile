.PHONY: help build debug test lint fmt vet vuln check clean install-tools check-pi-vars deploy deploy-debug run install-service restart status logs

.DEFAULT_GOAL := help

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Build:"
	@echo "  build           Cross-compile for Pi Zero W v1 (ARMv6)"
	@echo "  debug           Debug build with verbose brightness logging"
	@echo "  clean           Remove build artifacts"
	@echo ""
	@echo "Development:"
	@echo "  test            Run tests with race detector"
	@echo "  test-cover      Tests with coverage report"
	@echo "  lint            Run golangci-lint"
	@echo "  fmt             Format with gofumpt"
	@echo "  check           Run all checks (fmt + vet + lint + vuln)"
	@echo "  install-tools   Install golangci-lint, gofumpt, govulncheck"
	@echo ""
	@echo "Deployment (requires PI_USER and PI_HOST via .env or command line):"
	@echo "  deploy          Build and deploy to Pi"
	@echo "  deploy-debug    Deploy debug build to Pi"
	@echo "  run             Deploy and run interactively (requires LAT, LON)"
	@echo "  install-service Install systemd service on Pi"
	@echo "  restart         Restart service on Pi"
	@echo "  status          Check service status"
	@echo "  logs            Tail service logs"
	@echo ""
	@echo "Config: create .env with PI_USER, PI_HOST, LAT, LON (see README)"

APP      := botpi

GOOS     := linux
GOARCH   := arm
GOARM    := 6

# Load .env if present (command-line args override)
-include .env

# --- Pi deployment config (required for deploy/run/service targets) ---
# Set via .env file, environment, or command line (in order of precedence):
#   echo "PI_USER=misham\nPI_HOST=misham@mypi.local" > .env
#   make deploy PI_HOST=other@host  # overrides .env
PI_USER  ?=
PI_HOST  ?=
PI_PATH  ?=

define PI_HELP
ERROR: Pi deployment requires PI_USER and PI_HOST.

Set via .env file (recommended) or command line:
  echo 'PI_USER=misham' >> .env
  echo 'PI_HOST=misham@mypi.local' >> .env
  make deploy

Or inline:
  make deploy PI_USER=<user> PI_HOST=<user@host>
  make run    PI_USER=<user> PI_HOST=<user@host> LAT=<lat> LON=<lon>

Variables:
  PI_USER  SSH user on the Pi
  PI_HOST  SSH host (user@host)
  PI_PATH  Binary path on Pi (default: /home/$$PI_USER/botpi)
  LAT      Latitude for auto brightness (required for 'run')
  LON      Longitude for auto brightness (required for 'run')
endef
export PI_HELP

check-pi-vars:
	@if [ -z "$(PI_USER)" ] || [ -z "$(PI_HOST)" ]; then \
		echo "$$PI_HELP"; exit 1; \
	fi
	$(eval PI_PATH := $(if $(PI_PATH),$(PI_PATH),/home/$(PI_USER)/$(APP)))

# Build for Raspberry Pi Zero W v1 (ARMv6)
build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) \
		go build -ldflags="-s -w" -o $(APP) .

# Debug build with verbose logging
debug:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) \
		go build -ldflags="-X github.com/misham/botpi/brightness.debugMode=1" -o $(APP) .

# Test with race detector
test:
	go test -race -count=1 ./...

# Test with coverage
test-cover:
	go test -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Lint (golangci-lint runs staticcheck, errcheck, govet, gosec, and more)
lint:
	golangci-lint run ./...

# Format (gofumpt is a strict superset of gofmt)
fmt:
	gofumpt -w -modpath github.com/misham/botpi .

# Format check (CI — fails if files need formatting)
fmt-check:
	@test -z "$$(gofumpt -l -modpath github.com/misham/botpi .)" || (echo "files need formatting:"; gofumpt -l -modpath github.com/misham/botpi .; exit 1)

# Go vet
vet:
	go vet ./...

# Check for known vulnerabilities
vuln:
	govulncheck ./...

# Run all static checks
check: fmt-check vet lint vuln

# Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

# Clean build artifacts
clean:
	rm -f $(APP) coverage.out

# --- Pi deployment ---

# Deploy debug binary to Pi
deploy-debug: debug check-pi-vars
	-ssh $(PI_HOST) "sudo systemctl stop $(APP).service 2>/dev/null"
	scp $(APP) $(PI_HOST):$(PI_PATH)
	ssh $(PI_HOST) "chmod +x $(PI_PATH)"
	@if [ -f words.json ]; then scp words.json $(PI_HOST):$$(dirname $(PI_PATH))/words.json; fi
	-ssh $(PI_HOST) "sudo systemctl start $(APP).service 2>/dev/null"

# Deploy binary to Pi (stops/starts service automatically)
deploy: build check-pi-vars
	-ssh $(PI_HOST) "sudo systemctl stop $(APP).service 2>/dev/null"
	scp $(APP) $(PI_HOST):$(PI_PATH)
	ssh $(PI_HOST) "chmod +x $(PI_PATH)"
	@if [ -f words.json ]; then scp words.json $(PI_HOST):$$(dirname $(PI_PATH))/words.json; fi
	-ssh $(PI_HOST) "sudo systemctl start $(APP).service 2>/dev/null"

# Deploy and run interactively
run: deploy
	@if [ -z "$(LAT)" ] || [ -z "$(LON)" ]; then \
		echo "ERROR: 'run' requires LAT and LON."; \
		echo "  make run PI_USER=<user> PI_HOST=<user@host> LAT=<lat> LON=<lon>"; \
		exit 1; \
	fi
	ssh $(PI_HOST) "$(PI_PATH) -lat $(LAT) -lon $(LON)"

# Install systemd service on Pi (substitutes PI_USER and PI_PATH into service file)
install-service: check-pi-vars
	@if [ -z "$(LAT)" ] || [ -z "$(LON)" ]; then \
		echo "ERROR: 'install-service' requires LAT and LON."; \
		echo "  Set them in .env or pass on command line."; \
		exit 1; \
	fi
	sed -e 's|PI_USER|$(PI_USER)|g' -e 's|PI_PATH|$(PI_PATH)|g' botpi.service > /tmp/$(APP).service
	scp /tmp/$(APP).service $(PI_HOST):/tmp/$(APP).service
	ssh $(PI_HOST) "sudo cp /tmp/$(APP).service /etc/systemd/system/$(APP).service && \
		echo 'LAT=$(LAT)' | sudo tee /etc/botpi.conf > /dev/null && \
		echo 'LON=$(LON)' | sudo tee -a /etc/botpi.conf > /dev/null && \
		sudo systemctl daemon-reload && sudo systemctl enable $(APP).service && sudo systemctl restart $(APP).service"
	rm -f /tmp/$(APP).service

# Restart service on Pi
restart: check-pi-vars
	ssh $(PI_HOST) "sudo systemctl restart $(APP).service"

# Check service status
status: check-pi-vars
	ssh $(PI_HOST) "sudo systemctl status $(APP).service"

# Tail service logs
logs: check-pi-vars
	ssh $(PI_HOST) "journalctl -u $(APP).service -f"
