package face

import (
	"math/rand/v2"
	"time"

	"github.com/misham/botpi/display"
	"github.com/misham/botpi/font"
)

// Animator manages face animation state and timing.
type Animator struct {
	display     Displayer
	expressions []Expression
	current     int
	words       []string
}

// NewAnimator creates a new face animator.
// Words is a list of words to randomly scroll between expression changes.
func NewAnimator(d Displayer, words []string) *Animator {
	return &Animator{
		display: d,
		expressions: []Expression{
			Neutral,
			Happy,
			Surprised,
			Sleepy,
		},
		words: words,
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
			if err := a.doExpressionChange(stop); err != nil {
				return err
			}
		}
	}
}

// doExpressionChange picks a new expression and optionally shows a scrolling word.
func (a *Animator) doExpressionChange(stop <-chan struct{}) error {
	next := rand.IntN(len(a.expressions) - 1) //nolint:gosec // animation timing does not need crypto rand
	if next >= a.current {
		next++
	}
	a.current = next
	if err := a.drawExpression(a.expressions[a.current]); err != nil {
		return err
	}

	// ~33% chance to show a word after expression change.
	if len(a.words) > 0 && rand.Float64() < 0.33 { //nolint:gosec // animation timing does not need crypto rand
		word := a.words[rand.IntN(len(a.words))] //nolint:gosec // animation timing does not need crypto rand
		if err := a.scrollWord(word, stop); err != nil {
			return err
		}
		return a.drawExpression(a.expressions[a.current])
	}

	return nil
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

// scrollWord renders a word scrolling left across the display.
// The word enters from the right edge and exits the left edge.
func (a *Animator) scrollWord(word string, stop <-chan struct{}) error {
	buf, wordWidth := font.Render(word, 180)
	if wordWidth == 0 {
		return nil
	}

	totalFrames := wordWidth + display.Width
	stepDelay := time.Duration(10000/totalFrames) * time.Millisecond

	for frame := 1; frame <= totalFrames; frame++ {
		a.display.Clear()

		for dx := range display.Width {
			srcX := frame - display.Width + dx
			if srcX < 0 || srcX >= wordWidth {
				continue
			}
			for y := range display.Height {
				if buf[y][srcX] > 0 {
					a.display.SetPixel(dx, y, buf[y][srcX])
				}
			}
		}

		if err := a.display.Show(); err != nil {
			return err
		}

		select {
		case <-stop:
			return nil
		case <-time.After(stepDelay):
		}
	}

	return nil
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
