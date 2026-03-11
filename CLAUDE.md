# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development

```bash
make build          # Cross-compile for Pi Zero W v1 (ARMv6)
make debug          # Debug build with verbose brightness logging (ldflags sets debugMode=1)
make test           # Run tests with race detector
make test-cover     # Tests with coverage report
make lint           # golangci-lint (requires: make install-tools)
make fmt            # Format with gofumpt (requires: make install-tools)
make check          # Run all checks (fmt-check + vet + lint)
make install-tools  # Install golangci-lint and gofumpt
```

Run a single test:
```bash
go test -race -count=1 -run TestPixelAddr ./display/
```

## Deployment

```bash
make deploy          # Build, stop service, scp binary, start service
make deploy-debug    # Deploy debug build (verbose brightness logging)
make run             # Deploy and run interactively
make install-service # Install systemd service on Pi
make restart         # Restart service
make status          # Check service status
make logs            # Tail service logs
```

Target: Raspberry Pi Zero W v1, ARMv6, Raspbian trixie. Deploy targets require `PI_USER` and `PI_HOST` via `.env` file or command line. See `make deploy` for help.

## Architecture

Animated robot face on a Pimoroni Scroll pHAT HD (17x7 white LED matrix, IS31FL3731 over I2C).

### Four-layer structure

- **`driver/`** — IS31FL3731 I2C driver. Register communication, initialization, frame selection, double-buffered PWM writes. Reusable for any IS31FL3731 board.
- **`display/`** — 17x7 display buffer with Scroll pHAT HD pixel mapping, gamma correction (γ=2.2), three brightness modes, 180° rotation (board is mounted upside down).
- **`face/`** — Expression definitions (neutral, happy, surprised, sleepy, blink) and animation state machine with randomized timing.
- **`brightness/`** — Dynamic brightness controller. Solar elevation via `go-sunrise`, cloud cover via Open-Meteo. Requires `-lat`/`-lon` flags for location. Runs a 30-minute tick goroutine. Debug logging controlled by `debugMode` ldflags variable.
- **`main.go`** — CLI entry point with `-brightness` (default `auto`), `-lat`, `-lon` flags and SIGTERM/SIGINT handling.

### Key patterns

- **Pure-Go I2C** via `periph.io/x/conn/v3` and `periph.io/x/host/v3` — no CGO, cross-compiles cleanly.
- **Double-buffering** — alternates between IS31FL3731 frames 0/1 for flicker-free updates.
- **Pixel mapping** — splits 17 columns across two internal matrices (A: x=0-8, B: x=9-16). Formula in `display/display.go:pixelAddr`.
- **180° rotation** — applied in `display/display.go:Show()` because the board is mounted upside down.
- **Expression type** — `[Height][Width]byte` arrays built via helper functions in `face/expressions.go`.
- **Thread-safe brightness** — `Display.brightness` uses `atomic.Int32` for safe concurrent access between animator and brightness controller goroutines.
- **Debug builds** — `make debug` injects `debugMode=1` via ldflags into `brightness` package, enabling verbose logging of solar elevation, cloud cover, and brightness decisions.

### Testing

Unit tests cover pixel mapping (bounds, uniqueness, known values), expression validation, concurrent brightness access (race detector), and brightness controller logic (httptest + injected clock). No I2C mocking — the driver is thin and tested on real hardware.
