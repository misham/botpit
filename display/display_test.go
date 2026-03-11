package display

import (
	"errors"
	"testing"
)

func TestPixelAddr(t *testing.T) {
	// Verify known values from the lookup table in research doc
	tests := []struct {
		x, y int
		want int
	}{
		{0, 0, 134},
		{8, 0, 6},
		{16, 0, 120},
		{0, 6, 128},
		{8, 6, 0},
		{16, 6, 126},
		{9, 0, 8},
		{9, 6, 14},
		{4, 3, 67},
	}
	for _, tt := range tests {
		got := pixelAddr(tt.x, tt.y)
		if got != tt.want {
			t.Errorf("pixelAddr(%d, %d) = %d, want %d", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestPixelAddrBounds(t *testing.T) {
	// All valid coordinates should produce indices in [0, 134]
	for y := range Height {
		for x := range Width {
			idx := pixelAddr(x, y)
			if idx < 0 || idx >= PWMBufSize {
				t.Errorf("pixelAddr(%d, %d) = %d, out of range [0, %d)", x, y, idx, PWMBufSize)
			}
		}
	}
}

func TestPixelAddrUnique(t *testing.T) {
	// Each (x, y) should map to a unique index
	seen := make(map[int]struct{})
	for y := range Height {
		for x := range Width {
			idx := pixelAddr(x, y)
			if _, ok := seen[idx]; ok {
				t.Errorf("pixelAddr(%d, %d) = %d is a duplicate", x, y, idx)
			}
			seen[idx] = struct{}{}
		}
	}
}

// mockPWMWriter records ShowPWM calls.
type mockPWMWriter struct {
	lastPWM []byte
	calls   int
	err     error
}

func (m *mockPWMWriter) ShowPWM(pwm []byte) error {
	m.calls++
	m.lastPWM = make([]byte, len(pwm))
	copy(m.lastPWM, pwm)
	return m.err
}

func TestNew(t *testing.T) {
	d := New(&mockPWMWriter{}, BrightnessNormal)
	if d == nil {
		t.Fatal("New returned nil")
	}
}

func TestSetPixelAndClear(t *testing.T) {
	d := New(&mockPWMWriter{}, BrightnessNormal)
	d.SetPixel(5, 3, 200)
	if d.buf[3][5] != 200 {
		t.Errorf("buf[3][5] = %d, want 200", d.buf[3][5])
	}
	d.Clear()
	if d.buf[3][5] != 0 {
		t.Errorf("buf[3][5] after Clear = %d, want 0", d.buf[3][5])
	}
}

func TestSetPixelOutOfBounds(t *testing.T) {
	d := New(&mockPWMWriter{}, BrightnessNormal)
	// Should not panic.
	d.SetPixel(-1, 0, 255)
	d.SetPixel(0, -1, 255)
	d.SetPixel(Width, 0, 255)
	d.SetPixel(0, Height, 255)
	// Verify no pixels were set.
	for y := range Height {
		for x := range Width {
			if d.buf[y][x] != 0 {
				t.Errorf("out-of-bounds SetPixel modified buf[%d][%d]", y, x)
			}
		}
	}
}

func TestShow(t *testing.T) {
	mock := &mockPWMWriter{}
	d := New(mock, BrightnessNormal)
	d.SetPixel(8, 3, 255)

	if err := d.Show(); err != nil {
		t.Fatalf("Show failed: %v", err)
	}
	if mock.calls != 1 {
		t.Errorf("ShowPWM calls = %d, want 1", mock.calls)
	}
	if len(mock.lastPWM) != PWMBufSize {
		t.Fatalf("PWM buffer length = %d, want %d", len(mock.lastPWM), PWMBufSize)
	}

	// Pixel (8,3) rotated 180° is (Width-1-8, Height-1-3) = (8, 3).
	idx := pixelAddr(8, 3)
	if mock.lastPWM[idx] == 0 {
		t.Errorf("expected non-zero PWM at index %d for pixel (8,3)", idx)
	}
}

func TestShowError(t *testing.T) {
	mock := &mockPWMWriter{err: errors.New("hw error")}
	d := New(mock, BrightnessNormal)
	d.SetPixel(0, 0, 255)

	if err := d.Show(); err == nil {
		t.Fatal("Show should propagate PWMWriter error")
	}
}

func TestSetBrightness(t *testing.T) {
	mock := &mockPWMWriter{}
	d := New(mock, BrightnessBright)
	d.SetPixel(0, 0, 255)
	if err := d.Show(); err != nil {
		t.Fatalf("Show failed: %v", err)
	}
	brightPWM := make([]byte, len(mock.lastPWM))
	copy(brightPWM, mock.lastPWM)

	d.SetBrightness(BrightnessDark)
	d.SetPixel(0, 0, 255)
	if err := d.Show(); err != nil {
		t.Fatalf("Show failed: %v", err)
	}

	// Dark should produce dimmer output at the same pixel.
	idx := pixelAddr(Width-1, Height-1) // rotated (0,0)
	if mock.lastPWM[idx] >= brightPWM[idx] {
		t.Errorf("dark PWM[%d] = %d should be less than bright %d", idx, mock.lastPWM[idx], brightPWM[idx])
	}
}

func TestSetBrightnessConcurrent(t *testing.T) { //nolint:revive // race detector uses t implicitly
	mock := &mockPWMWriter{}
	d := New(mock, BrightnessNormal)

	// Fill buffer so Show does real work
	for y := range Height {
		for x := range Width {
			d.SetPixel(x, y, 128)
		}
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 1000 {
			d.SetBrightness(BrightnessDark)
			d.SetBrightness(BrightnessBright)
		}
	}()

	for range 1000 {
		_ = d.Show()
	}
	<-done
}

func TestShowClearBuffer(t *testing.T) {
	mock := &mockPWMWriter{}
	d := New(mock, BrightnessNormal)

	// Show with empty buffer should produce all-zero PWM.
	if err := d.Show(); err != nil {
		t.Fatalf("Show failed: %v", err)
	}
	for i, v := range mock.lastPWM {
		if v != 0 {
			t.Errorf("empty display: PWM[%d] = %d, want 0", i, v)
			break
		}
	}
}
