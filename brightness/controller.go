package brightness

import (
	"log"
	"net/http"
	"time"

	"github.com/misham/botpi/display"
)

// debugMode is set to "1" via ldflags in debug builds.
var debugMode string //nolint:gochecknoglobals // set via ldflags

func debugf(format string, args ...any) {
	if debugMode == "1" {
		log.Printf(format, args...)
	}
}

// DefaultWeatherURL is the Open-Meteo API base URL.
const DefaultWeatherURL = "https://api.open-meteo.com/v1/forecast"

// Location holds geographic coordinates.
type Location struct {
	Lat float64
	Lon float64
}

// Controller dynamically adjusts display brightness based on solar elevation
// and weather cloud cover.
type Controller struct {
	display    Setter
	client     *http.Client
	loc        Location
	weatherURL string
	interval   time.Duration
	now        func() time.Time // injectable clock for testing
}

// NewController creates a brightness controller with the given location.
func NewController(d Setter, loc Location, weatherURL string) *Controller {
	debugf("brightness: using location lat=%.2f lon=%.2f", loc.Lat, loc.Lon)

	return &Controller{
		display:    d,
		client:     &http.Client{Timeout: 10 * time.Second},
		loc:        loc,
		weatherURL: weatherURL,
		interval:   30 * time.Minute,
		now:        time.Now,
	}
}

// Run starts the brightness control loop. Blocks until stop is closed.
func (c *Controller) Run(stop <-chan struct{}) {
	// Set brightness immediately on start.
	c.update()

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			c.update()
		}
	}
}

// update computes and applies the current brightness.
func (c *Controller) update() {
	now := c.now()
	base := SolarBrightness(c.loc.Lat, c.loc.Lon, now)
	debugf("brightness: solar base=%d at %s", base, now.Format(time.RFC3339))

	// Only fetch weather if it's daytime (solar says normal or bright).
	if base == display.BrightnessDark {
		debugf("brightness: night — setting dark")
		c.display.SetBrightness(base)
		return
	}

	cloudCover, err := FetchCloudCover(c.client, c.weatherURL, c.loc.Lat, c.loc.Lon)
	if err != nil {
		// Weather unavailable — use solar-only brightness.
		log.Printf("brightness: weather fetch failed: %v", err)
		debugf("brightness: weather failed, using solar fallback=%d", base)
		c.display.SetBrightness(base)
		return
	}

	debugf("brightness: cloud_cover=%d%%", cloudCover)

	// Heavy cloud cover during daytime dims one level.
	if cloudCover >= 80 && base == display.BrightnessBright {
		debugf("brightness: overcast — dimming bright to normal")
		c.display.SetBrightness(display.BrightnessNormal)
		return
	}

	debugf("brightness: setting %d", base)
	c.display.SetBrightness(base)
}
