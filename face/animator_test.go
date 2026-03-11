package face

import (
	"errors"
	"testing"
	"time"

	"github.com/misham/botpi/display"
)

var errShow = errors.New("show error")

// mockDisplay records calls to SetPixel, Clear, and Show.
type mockDisplay struct {
	pixels     map[[2]int]byte
	showErr    error
	showCalls  int
	clearCalls int
}

func newMockDisplay() *mockDisplay {
	return &mockDisplay{pixels: make(map[[2]int]byte)}
}

func (m *mockDisplay) SetPixel(x, y int, value byte) {
	m.pixels[[2]int{x, y}] = value
}

func (m *mockDisplay) Clear() {
	m.pixels = make(map[[2]int]byte)
	m.clearCalls++
}

func (m *mockDisplay) Show() error {
	m.showCalls++
	return m.showErr
}

func TestNewAnimator(t *testing.T) {
	mock := newMockDisplay()
	a := NewAnimator(mock)
	if a == nil {
		t.Fatal("NewAnimator returned nil")
	}
	if len(a.expressions) != 4 {
		t.Errorf("expected 4 expressions, got %d", len(a.expressions))
	}
}

func TestDrawExpression(t *testing.T) {
	mock := newMockDisplay()
	a := NewAnimator(mock)

	if err := a.drawExpression(Neutral); err != nil {
		t.Fatalf("drawExpression failed: %v", err)
	}

	if mock.clearCalls != 1 {
		t.Errorf("expected 1 Clear call, got %d", mock.clearCalls)
	}
	if mock.showCalls != 1 {
		t.Errorf("expected 1 Show call, got %d", mock.showCalls)
	}

	// Verify pixels were set for the expression.
	litPixels := 0
	for y := range display.Height {
		for x := range display.Width {
			if Neutral[y][x] > 0 {
				litPixels++
			}
		}
	}
	if len(mock.pixels) != litPixels {
		t.Errorf("expected %d lit pixels, got %d", litPixels, len(mock.pixels))
	}
}

func TestDrawExpressionError(t *testing.T) {
	mock := newMockDisplay()
	mock.showErr = errShow
	a := NewAnimator(mock)

	if err := a.drawExpression(Neutral); !errors.Is(err, errShow) {
		t.Errorf("drawExpression error = %v, want %v", err, errShow)
	}
}

func TestDoBlink(t *testing.T) {
	mock := newMockDisplay()
	a := NewAnimator(mock)
	a.current = 0

	stop := make(chan struct{})
	if err := a.doBlink(stop); err != nil {
		t.Fatalf("doBlink failed: %v", err)
	}

	// doBlink draws the blink frame, waits, then redraws the current expression.
	if mock.showCalls != 2 {
		t.Errorf("expected 2 Show calls (blink + restore), got %d", mock.showCalls)
	}
}

func TestDoBlinkStopDuringWait(t *testing.T) {
	mock := newMockDisplay()
	a := NewAnimator(mock)
	a.current = 0

	stop := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- a.doBlink(stop)
	}()

	// Let blink frame draw, then stop during the sleep.
	time.Sleep(50 * time.Millisecond)
	close(stop)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("doBlink returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("doBlink did not stop within timeout")
	}

	// Should have drawn the blink frame but not restored.
	if mock.showCalls != 1 {
		t.Errorf("expected 1 Show call (blink only), got %d", mock.showCalls)
	}
}

func TestDoBlinkShowError(t *testing.T) {
	mock := newMockDisplay()
	mock.showErr = errShow
	a := NewAnimator(mock)
	a.current = 0

	stop := make(chan struct{})
	if err := a.doBlink(stop); !errors.Is(err, errShow) {
		t.Errorf("doBlink error = %v, want %v", err, errShow)
	}
}

func TestRunStopsOnSignal(t *testing.T) {
	mock := newMockDisplay()
	a := NewAnimator(mock)

	stop := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- a.Run(stop)
	}()

	// Let it draw the initial expression.
	time.Sleep(50 * time.Millisecond)
	close(stop)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not stop within timeout")
	}

	// Should have drawn at least the initial expression.
	if mock.showCalls < 1 {
		t.Error("expected at least 1 Show call for initial expression")
	}
}

func TestRunInitialDrawError(t *testing.T) {
	mock := newMockDisplay()
	mock.showErr = errShow
	a := NewAnimator(mock)

	stop := make(chan struct{})
	err := a.Run(stop)
	if !errors.Is(err, errShow) {
		t.Errorf("Run error = %v, want %v", err, errShow)
	}
}

func TestRunDrawsAllExpressions(t *testing.T) {
	mock := newMockDisplay()
	a := NewAnimator(mock)

	// Draw each expression directly to verify they all work.
	for i, expr := range a.expressions {
		mock.showCalls = 0
		mock.clearCalls = 0
		if err := a.drawExpression(expr); err != nil {
			t.Fatalf("drawExpression[%d] failed: %v", i, err)
		}
		if mock.showCalls != 1 {
			t.Errorf("expression %d: Show calls = %d, want 1", i, mock.showCalls)
		}
		if len(mock.pixels) == 0 {
			t.Errorf("expression %d: no pixels drawn", i)
		}
	}
}
