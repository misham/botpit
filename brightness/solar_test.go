package brightness

import (
	"testing"
	"time"

	"github.com/misham/botpi/display"
)

func TestSolarBrightness(t *testing.T) {
	// San Francisco coordinates
	const lat, lon = 37.77, -122.43

	tests := []struct {
		name string
		time time.Time
		want display.Brightness
	}{
		{
			name: "midnight is dark",
			time: time.Date(2026, 6, 15, 7, 0, 0, 0, time.UTC), // midnight PDT
			want: display.BrightnessDark,
		},
		{
			name: "noon is bright",
			time: time.Date(2026, 6, 15, 19, 0, 0, 0, time.UTC), // noon PDT
			want: display.BrightnessBright,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SolarBrightness(lat, lon, tt.time)
			if got != tt.want {
				t.Errorf("SolarBrightness(%v) = %d, want %d", tt.time, got, tt.want)
			}
		})
	}
}
