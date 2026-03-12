package font

// Render converts a string into a pixel buffer for scrolling display.
// Returns a [GlyphHeight][]byte buffer and the total width in pixels.
// Unknown characters are skipped. The brightness parameter scales lit
// glyph pixels (0xFF) to produce the final PWM value.
func Render(s string, brightness byte) ([GlyphHeight][]byte, int) {
	var buf [GlyphHeight][]byte
	// Collect known glyphs
	var gs []Glyph
	for _, r := range s {
		if g, ok := glyphs[r]; ok {
			gs = append(gs, g)
		}
	}

	if len(gs) == 0 {
		return buf, 0
	}

	width := len(gs)*(GlyphWidth+GlyphSpacing) - GlyphSpacing
	for y := range GlyphHeight {
		buf[y] = make([]byte, width)
	}

	for i, g := range gs {
		xOff := i * (GlyphWidth + GlyphSpacing)
		for y := range GlyphHeight {
			for x := range GlyphWidth {
				if g[y][x] > 0 {
					buf[y][xOff+x] = byte(uint16(g[y][x]) * uint16(brightness) / 255) //nolint:gosec // result always fits in byte
				}
			}
		}
	}

	return buf, width
}
