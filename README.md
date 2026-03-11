# botpi

Animated robot face on a Pimoroni Scroll pHAT HD, driven by a Raspberry Pi Zero W.

## Features

- **Robot face** — cycles through neutral, happy, surprised, and sleepy expressions
- **Blinking** — randomized eye blinks between expression changes
- **Brightness modes** — bright, normal, and dark with gamma correction
- **Graceful shutdown** — clears the display on SIGTERM/SIGINT
- **systemd service** — auto-starts on boot, restarts on failure
- **No CGO** — pure-Go I2C driver, single static binary

## Hardware

- Raspberry Pi Zero W v1 (ARMv6)
- [Pimoroni Scroll pHAT HD](https://shop.pimoroni.com/products/scroll-phat-hd) (17x7 white LED matrix, IS31FL3731)
- I2C bus 1, address 0x74

## Quick Start

### Pi Setup

```bash
# Enable I2C
sudo raspi-config nonint do_i2c 0
sudo apt-get install -y i2c-tools
sudo usermod -a -G i2c $USER
sudo reboot

# Verify device
sudo i2cdetect -y 1   # should show 0x74
```

### Build & Deploy (from macOS)

```bash
make deploy           # cross-compile and scp to Pi
make install-service  # install systemd service
make status           # verify it's running
```

### Run Interactively

```bash
make run                                                    # default: normal brightness
ssh misham@botpi.local "/home/misham/botpi -brightness dark"  # dark mode
```

## Usage

```
botpi [-brightness bright|normal|dark]
```

| Flag | Values | Default | Description |
|------|--------|---------|-------------|
| `-brightness` | `bright`, `normal`, `dark` | `normal` | LED brightness level |

## Development

```bash
make install-tools  # install golangci-lint and gofumpt
make test           # run tests with race detector
make lint           # run golangci-lint
make fmt            # format with gofumpt
make check          # run all checks (fmt + vet + lint)
```

## Architecture

```
main.go          CLI entry point, signal handling
face/            Expression definitions and animation state machine
display/         17x7 pixel buffer, brightness, gamma correction
driver/          IS31FL3731 I2C driver (reusable)
```

## License

MIT
