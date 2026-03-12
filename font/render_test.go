package font

import "testing"

func TestRenderSingleChar(t *testing.T) {
	buf, width := Render("a", 255)
	if width != GlyphWidth {
		t.Fatalf("width = %d, want %d", width, GlyphWidth)
	}
	// Verify buffer matches glyph 'a' at full brightness
	g := glyphs['a']
	for y := range GlyphHeight {
		if len(buf[y]) != width {
			t.Fatalf("row %d length = %d, want %d", y, len(buf[y]), width)
		}
		for x := range GlyphWidth {
			want := g[y][x] // 0xFF scaled by 255/255 = 0xFF
			if buf[y][x] != want {
				t.Errorf("pixel (%d,%d) = %d, want %d", x, y, buf[y][x], want)
			}
		}
	}
}

func TestRenderMultiChar(t *testing.T) {
	buf, width := Render("ab", 255)
	wantWidth := 2*(GlyphWidth+GlyphSpacing) - GlyphSpacing
	if width != wantWidth {
		t.Fatalf("width = %d, want %d", width, wantWidth)
	}
	for y := range GlyphHeight {
		if len(buf[y]) != wantWidth {
			t.Errorf("row %d length = %d, want %d", y, len(buf[y]), wantWidth)
		}
	}
}

func TestRenderBrightnessScaling(t *testing.T) {
	buf, _ := Render("a", 128)
	g := glyphs['a']
	for y := range GlyphHeight {
		for x := range GlyphWidth {
			if g[y][x] == 0 {
				if buf[y][x] != 0 {
					t.Errorf("unlit pixel (%d,%d) = %d, want 0", x, y, buf[y][x])
				}
			} else {
				want := byte(uint16(g[y][x]) * 128 / 255)
				if buf[y][x] != want {
					t.Errorf("pixel (%d,%d) = %d, want %d", x, y, buf[y][x], want)
				}
			}
		}
	}
}

func TestRenderSpace(t *testing.T) {
	buf, width := Render(" ", 255)
	if width != GlyphWidth {
		t.Fatalf("width = %d, want %d", width, GlyphWidth)
	}
	for y := range GlyphHeight {
		for x := range width {
			if buf[y][x] != 0 {
				t.Errorf("space pixel (%d,%d) = %d, want 0", x, y, buf[y][x])
			}
		}
	}
}

func TestRenderUnknownCharSkipped(t *testing.T) {
	buf, width := Render("a1b", 255)
	// '1' is unknown, skipped — result should be same as "ab"
	wantWidth := 2*(GlyphWidth+GlyphSpacing) - GlyphSpacing
	if width != wantWidth {
		t.Fatalf("width = %d, want %d", width, wantWidth)
	}
	for y := range GlyphHeight {
		if len(buf[y]) != wantWidth {
			t.Errorf("row %d length = %d, want %d", y, len(buf[y]), wantWidth)
		}
	}
}

func TestRenderEmptyString(t *testing.T) {
	buf, width := Render("", 255)
	if width != 0 {
		t.Fatalf("width = %d, want 0", width)
	}
	for y := range GlyphHeight {
		if len(buf[y]) != 0 {
			t.Errorf("row %d length = %d, want 0", y, len(buf[y]))
		}
	}
}
