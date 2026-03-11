package display

import "math"

// Display dimensions and hardware constants.
const (
	Width      = 17
	Height     = 7
	I2CAddr    = 0x74
	PWMBufSize = 135 // Maximum IS31FL3731 index used by Scroll pHAT HD.
)

// Brightness controls LED output level.
type Brightness int

// Brightness presets.
const (
	BrightnessDark   Brightness = iota // ~15% max.
	BrightnessNormal                   // ~40% max.
	BrightnessBright                   // 100%.
)

// brightnessScale maps Brightness to a 0-255 scaling factor.
var brightnessScale = map[Brightness]byte{
	BrightnessDark:   38,  // 15% of 255
	BrightnessNormal: 102, // 40% of 255
	BrightnessBright: 255, // 100%
}

// gammaTable computes a 2.2 gamma correction lookup table at init time.
var gamma [256]byte

func init() {
	for i := range 256 {
		gamma[i] = byte(math.Round(255.0 * math.Pow(float64(i)/255.0, 2.2)))
	}
}

// Display manages a 17x7 pixel buffer and renders to the IS31FL3731.
type Display struct {
	device     PWMShower
	buf        [Height][Width]byte // pixel buffer, 0-255 per pixel
	brightness Brightness
}

// New creates a new Display.
func New(dev PWMShower, brightness Brightness) *Display {
	return &Display{
		device:     dev,
		brightness: brightness,
	}
}

// SetPixel sets a pixel brightness (0-255) at display coordinates.
// x: 0-16 (left to right), y: 0-6 (top to bottom).
func (d *Display) SetPixel(x, y int, value byte) {
	if x < 0 || x >= Width || y < 0 || y >= Height {
		return
	}
	d.buf[y][x] = value
}

// Clear sets all pixels to 0.
func (d *Display) Clear() {
	d.buf = [Height][Width]byte{}
}

// Show flushes the buffer to the hardware display.
func (d *Display) Show() error {
	pwm := make([]byte, PWMBufSize)
	scale := brightnessScale[d.brightness]

	for y := range Height {
		for x := range Width {
			if d.buf[y][x] == 0 {
				continue
			}
			// Scale by brightness, then apply gamma
			scaled := uint16(d.buf[y][x]) * uint16(scale) / 255
			// Rotate 180° for upside-down mounted board
			idx := pixelAddr(Width-1-x, Height-1-y)
			if idx >= 0 && idx < PWMBufSize {
				pwm[idx] = gamma[scaled]
			}
		}
	}

	return d.device.ShowPWM(pwm)
}

// SetBrightness changes the brightness mode.
func (d *Display) SetBrightness(b Brightness) {
	d.brightness = b
}

// pixelAddr converts display (x, y) to IS31FL3731 PWM buffer index.
// Matrix A (x=0-8): index = (8-x)*16 + (6-y)
// Matrix B (x=9-16): index = (x-8)*16 + (y-8), where y-8 is intentionally
// negative for y in [0,6] — it maps to lower offsets within each column's
// 16-byte block in the IS31FL3731's internal layout.
func pixelAddr(x, y int) int {
	if x <= 8 {
		return (8-x)*16 + (6 - y)
	}
	return (x-8)*16 + (y - 8)
}
