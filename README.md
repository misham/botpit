# botpi

Animated robot face on a Pimoroni Scroll pHAT HD, driven by a Raspberry Pi Zero W.

## Features

- **Robot face** — cycles through neutral, happy, surprised, and sleepy expressions
- **Scrolling words** — displays words scrolling across the LED matrix between expression changes
- **Blinking** — randomized eye blinks between expression changes
- **Dynamic brightness** — auto-adjusts based on solar elevation and cloud cover (Open-Meteo API)
- **Brightness modes** — bright, normal, and dark with gamma correction; static override available
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
make install-service  # install systemd service (creates /etc/botpi.conf)
make status           # verify it's running
```

Configure location on the Pi (`/etc/botpi.conf`):
```
LAT=37.77
LON=-122.43
```

Create a `.env` file (gitignored) for your Pi deployment config:
```bash
echo 'PI_USER=pi' >> .env
echo 'PI_HOST=pi@mypi.local' >> .env
echo 'LAT=37.77' >> .env
echo 'LON=-122.43' >> .env
```

### Run Interactively

```bash
make run        # uses .env values
make run LAT=0  # command-line overrides .env
```

## Usage

```
botpi [-brightness auto|bright|normal|dark] [-lat <latitude> -lon <longitude>]
```

| Flag | Values | Default | Description |
|------|--------|---------|-------------|
| `-brightness` | `auto`, `bright`, `normal`, `dark` | `auto` | LED brightness level |
| `-lat` | float | — | Latitude (required for `auto` mode) |
| `-lon` | float | — | Longitude (required for `auto` mode) |

In `auto` mode, brightness adjusts every 30 minutes based on:
- **Solar elevation** — dark at night, normal at twilight, bright during the day
- **Cloud cover** — heavy overcast (≥80%) dims bright to normal

If weather data is unavailable, falls back to solar-only brightness.

## Development

```bash
make install-tools  # install golangci-lint and gofumpt
make test           # run tests with race detector
make lint           # run golangci-lint
make fmt            # format with gofumpt
make check          # run all checks (fmt + vet + lint)
make debug          # debug build with verbose brightness logging
make deploy-debug   # deploy debug build to Pi
```

## Customizing Words

The robot displays scrolling words between face expressions. To enable this,
create a `words.json` file next to the binary:

```json
{
  "words": ["compassion", "connection", "creativity", "curiosity"]
}
```

If `words.json` is absent or empty, word display is disabled (faces only).
For deployment, `scp` the file to the same directory as the binary on the Pi.

## Architecture

```
main.go          CLI entry point, signal handling
font/            5x7 pixel font and text rendering for LED scrolling
face/            Expression definitions, animation state machine, word display
display/         17x7 pixel buffer, brightness, gamma correction
brightness/      Dynamic brightness controller (solar + weather)
driver/          IS31FL3731 I2C driver (reusable)
```

## License

MIT
