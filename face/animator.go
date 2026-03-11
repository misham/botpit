package face

import (
	"math/rand/v2"
	"time"

	"github.com/misham/botpi/display"
)

// Animator manages face animation state and timing.
type Animator struct {
	display     Displayer
	expressions []Expression
	current     int
}

// NewAnimator creates a new face animator.
func NewAnimator(d Displayer) *Animator {
	return &Animator{
		display: d,
		expressions: []Expression{
			Neutral,
			Happy,
			Surprised,
			Sleepy,
		},
	}
}

// Run starts the animation loop. Blocks until stop is closed.
func (a *Animator) Run(stop <-chan struct{}) error {
	// Start with neutral face
	a.current = 0
	if err := a.drawExpression(a.expressions[a.current]); err != nil {
		return err
	}

	for {
		// Wait 3-6 seconds between actions
		wait := time.Duration(3000+rand.IntN(3000)) * time.Millisecond //nolint:gosec // animation timing does not need crypto rand
		select {
		case <-stop:
			return nil
		case <-time.After(wait):
		}

		// 40% chance to blink, 60% chance to change expression
		if rand.Float64() < 0.4 { //nolint:gosec // animation timing does not need crypto rand
			if err := a.doBlink(stop); err != nil {
				return err
			}
		} else {
			// Pick a different expression
			next := rand.IntN(len(a.expressions) - 1) //nolint:gosec // animation timing does not need crypto rand
			if next >= a.current {
				next++
			}
			a.current = next
			if err := a.drawExpression(a.expressions[a.current]); err != nil {
				return err
			}
		}
	}
}

// doBlink performs a quick blink animation (close eyes, pause, reopen).
func (a *Animator) doBlink(stop <-chan struct{}) error {
	// Create blink frame: closed eyes + current mouth
	blinkFrame := Blink
	// Copy mouth from current expression (rows 4-6)
	for y := 4; y < display.Height; y++ {
		for x := range display.Width {
			if blinkFrame[y][x] == 0 {
				blinkFrame[y][x] = a.expressions[a.current][y][x]
			}
		}
	}

	if err := a.drawExpression(blinkFrame); err != nil {
		return err
	}

	// Eyes closed for 100-200ms
	blinkDur := time.Duration(100+rand.IntN(100)) * time.Millisecond //nolint:gosec // animation timing does not need crypto rand
	select {
	case <-stop:
		return nil
	case <-time.After(blinkDur):
	}

	return a.drawExpression(a.expressions[a.current])
}

// drawExpression renders an expression to the display.
func (a *Animator) drawExpression(expr Expression) error {
	a.display.Clear()
	for y := range display.Height {
		for x := range display.Width {
			if expr[y][x] > 0 {
				a.display.SetPixel(x, y, expr[y][x])
			}
		}
	}
	return a.display.Show()
}
