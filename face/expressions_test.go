package face

import (
	"testing"

	"github.com/misham/botpi/display"
)

func TestExpressionsNonEmpty(t *testing.T) {
	exprs := map[string]Expression{
		"Neutral":   Neutral,
		"Happy":     Happy,
		"Surprised": Surprised,
		"Sleepy":    Sleepy,
		"Blink":     Blink,
	}
	for name, expr := range exprs {
		pixels := 0
		for y := range display.Height {
			for x := range display.Width {
				if expr[y][x] > 0 {
					pixels++
				}
			}
		}
		if pixels == 0 {
			t.Errorf("expression %s has no lit pixels", name)
		}
	}
}

func TestBlinkComposition(t *testing.T) {
	// Simulate what doBlink does: overlay current expression's mouth onto Blink
	blinkFrame := Blink
	source := Neutral
	for y := 4; y < display.Height; y++ {
		for x := range display.Width {
			if blinkFrame[y][x] == 0 {
				blinkFrame[y][x] = source[y][x]
			}
		}
	}
	// Should have closed eyes (from Blink, rows 0-3)
	eyePixels := 0
	for y := range 4 {
		for x := range display.Width {
			if blinkFrame[y][x] > 0 {
				eyePixels++
			}
		}
	}
	if eyePixels == 0 {
		t.Error("blink composition should have closed eyes")
	}
	// Should have mouth (from Neutral, rows 4-6)
	mouthPixels := 0
	for y := 4; y < display.Height; y++ {
		for x := range display.Width {
			if blinkFrame[y][x] > 0 {
				mouthPixels++
			}
		}
	}
	if mouthPixels == 0 {
		t.Error("blink composition should have mouth from source expression")
	}
}

func TestExpressionsHaveEyesAndMouth(t *testing.T) {
	exprs := map[string]Expression{
		"Neutral":   Neutral,
		"Happy":     Happy,
		"Surprised": Surprised,
		"Sleepy":    Sleepy,
	}
	for name, expr := range exprs {
		eyePixels, mouthPixels := 0, 0
		for y := range display.Height {
			for x := range display.Width {
				if expr[y][x] > 0 {
					if y <= 3 {
						eyePixels++
					} else {
						mouthPixels++
					}
				}
			}
		}
		if eyePixels == 0 {
			t.Errorf("expression %s has no eye pixels (rows 0-3)", name)
		}
		if mouthPixels == 0 {
			t.Errorf("expression %s has no mouth pixels (rows 4-6)", name)
		}
	}
}
