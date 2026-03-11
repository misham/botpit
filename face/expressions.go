package face

import "github.com/misham/botpi/display"

// Expression represents a face state as a pixel grid.
type Expression [display.Height][display.Width]byte

// Face expressions initialized in init().
var (
	Neutral   Expression
	Happy     Expression
	Surprised Expression
	Sleepy    Expression
	Blink     Expression
)

func init() {
	// Neutral: standard eyes + straight mouth
	setEyes(&Neutral, 200)
	setMouthStraight(&Neutral, 150)

	// Happy: standard eyes + curved smile
	setEyes(&Happy, 200)
	setMouthSmile(&Happy, 150)

	// Surprised: wide eyes + small mouth
	setEyesSurprised(&Surprised, 220)
	setMouthSurprised(&Surprised, 150)

	// Sleepy: half-closed eyes + straight mouth
	setEyesSleepy(&Sleepy, 120)
	setMouthStraight(&Sleepy, 100)

	// Blink: closed eyes (thin lines), no mouth (filled per current state)
	setEyesClosed(&Blink, 150)
}

func setEyes(e *Expression, brightness byte) {
	// Left eye at (2,2)-(3,3), Right eye at (13,2)-(14,3)
	for _, ox := range []int{2, 13} {
		for dx := range 2 {
			for dy := range 2 {
				e[2+dy][ox+dx] = brightness
			}
		}
	}
}

func setEyesSurprised(e *Expression, brightness byte) {
	// Taller eyes: 2x3 at (2,1)-(3,3) and (13,1)-(14,3)
	for _, ox := range []int{2, 13} {
		for dx := range 2 {
			for dy := range 3 {
				e[1+dy][ox+dx] = brightness
			}
		}
	}
}

func setEyesSleepy(e *Expression, brightness byte) {
	// Thin line eyes at y=3: (2,3)-(4,3) and (12,3)-(14,3)
	for _, ox := range []int{2, 12} {
		for dx := range 3 {
			e[3][ox+dx] = brightness
		}
	}
}

func setEyesClosed(e *Expression, brightness byte) {
	setEyesSleepy(e, brightness)
}

func setMouthStraight(e *Expression, brightness byte) {
	// Straight line at y=5, x=4 to x=12
	for x := 4; x <= 12; x++ {
		e[5][x] = brightness
	}
}

func setMouthSmile(e *Expression, brightness byte) {
	// Curved smile: straight line at y=5, corners down at y=6
	for x := 6; x <= 10; x++ {
		e[5][x] = brightness
	}
	e[6][5] = brightness
	e[6][11] = brightness
}

func setMouthSurprised(e *Expression, brightness byte) {
	// Small O mouth at center
	e[5][7] = brightness
	e[5][9] = brightness
	e[4][8] = brightness
	e[6][8] = brightness
}
