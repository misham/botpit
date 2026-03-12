package font

import "testing"

func TestAllGlyphsDefined(t *testing.T) {
	// All lowercase letters a-z and space should be in the font map
	for r := 'a'; r <= 'z'; r++ {
		if _, ok := glyphs[r]; !ok {
			t.Errorf("missing glyph for %q", r)
		}
	}
	if _, ok := glyphs[' ']; !ok {
		t.Error("missing glyph for space")
	}
}

func TestGlyphDimensions(t *testing.T) {
	for r, g := range glyphs {
		if len(g) != GlyphHeight {
			t.Errorf("glyph %q: height = %d, want %d", r, len(g), GlyphHeight)
		}
		for row := range g {
			if len(g[row]) != GlyphWidth {
				t.Errorf("glyph %q row %d: width = %d, want %d", r, row, len(g[row]), GlyphWidth)
			}
		}
	}
}

func TestGlyphsHaveLitPixels(t *testing.T) {
	for r, g := range glyphs {
		if r == ' ' {
			continue // space is blank
		}
		hasLit := false
		for _, row := range g {
			for _, v := range row {
				if v > 0 {
					hasLit = true
				}
			}
		}
		if !hasLit {
			t.Errorf("glyph %q has no lit pixels", r)
		}
	}
}

func TestSpaceGlyphIsBlank(t *testing.T) {
	g := glyphs[' ']
	for y, row := range g {
		for x, v := range row {
			if v != 0 {
				t.Errorf("space glyph pixel (%d,%d) = %d, want 0", x, y, v)
			}
		}
	}
}
