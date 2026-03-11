package brightness

import (
	"time"

	"github.com/misham/botpi/display"
	sunrise "github.com/nathan-osman/go-sunrise"
)

// SolarBrightness returns a brightness mode based on solar elevation at the
// given location and time.
//
// Below 0° (night):      BrightnessDark
// 0°–30° (twilight/low): BrightnessNormal
// Above 30° (daytime):   BrightnessBright.
func SolarBrightness(lat, lon float64, t time.Time) display.Brightness {
	elevation := sunrise.Elevation(lat, lon, t)

	switch {
	case elevation < 0:
		return display.BrightnessDark
	case elevation < 30:
		return display.BrightnessNormal
	default:
		return display.BrightnessBright
	}
}
