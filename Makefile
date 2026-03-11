.PHONY: build test lint fmt vet vuln check clean install-tools deploy run install-service restart status logs

APP      := botpi
PI_HOST  := misham@botpi.local
PI_PATH  := /home/misham/$(APP)

GOOS     := linux
GOARCH   := arm
GOARM    := 6

# Build for Raspberry Pi Zero W v1 (ARMv6)
build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) \
		go build -ldflags="-s -w" -o $(APP) .

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

# Deploy binary to Pi (stops/starts service automatically)
deploy: build
	-ssh $(PI_HOST) "sudo systemctl stop $(APP).service 2>/dev/null"
	scp $(APP) $(PI_HOST):$(PI_PATH)
	ssh $(PI_HOST) "chmod +x $(PI_PATH)"
	-ssh $(PI_HOST) "sudo systemctl start $(APP).service 2>/dev/null"

# Deploy and run interactively
run: deploy
	ssh $(PI_HOST) "$(PI_PATH) -brightness normal"

# Install systemd service on Pi
install-service:
	ssh $(PI_HOST) 'sudo tee /etc/systemd/system/$(APP).service > /dev/null' < botpi.service
	ssh $(PI_HOST) "sudo systemctl daemon-reload && sudo systemctl enable $(APP).service && sudo systemctl restart $(APP).service"

# Restart service on Pi
restart:
	ssh $(PI_HOST) "sudo systemctl restart $(APP).service"

# Check service status
status:
	ssh $(PI_HOST) "sudo systemctl status $(APP).service"

# Tail service logs
logs:
	ssh $(PI_HOST) "journalctl -u $(APP).service -f"
