package brightness

import "github.com/misham/botpi/display"

// Setter is the interface for setting display brightness.
type Setter interface {
	SetBrightness(b display.Brightness)
}
