package driver

import (
	"errors"
	"testing"
)

var errI2C = errors.New("i2c error")

// mockConn records Tx calls and optionally fails at call N.
type mockConn struct {
	writes [][]byte
	failAt int // fail on this call number (1-indexed); 0 = never fail
	calls  int
}

func (m *mockConn) Tx(w []byte, _ []byte) error {
	m.calls++
	if m.failAt > 0 && m.calls >= m.failAt {
		return errI2C
	}
	cp := make([]byte, len(w))
	copy(cp, w)
	m.writes = append(m.writes, cp)
	return nil
}

func newTestDevice(mock *mockConn) *Device {
	return &Device{dev: mock}
}

func TestInit(t *testing.T) {
	mock := &mockConn{}
	dev := newTestDevice(mock)

	if err := dev.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if len(mock.writes) == 0 {
		t.Fatal("Init produced no I2C writes")
	}

	// First write should be bank select to config (0x0B).
	if mock.writes[0][0] != RegCommand || mock.writes[0][1] != RegConfig {
		t.Errorf("first write = %v, want bank select to config", mock.writes[0])
	}

	// Verify frame state after init.
	if dev.frame != 0 {
		t.Errorf("frame after Init = %d, want 0", dev.frame)
	}
}

func TestInitEnablesAllFrames(t *testing.T) {
	mock := &mockConn{}
	dev := newTestDevice(mock)

	if err := dev.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Count enable writes (19-byte payloads starting with OffsetEnable).
	enableCount := 0
	for _, w := range mock.writes {
		if len(w) == EnableBytes+1 && w[0] == OffsetEnable {
			enableCount++
			// Verify LED enable pattern: 17 columns of 0x7F.
			for i := 1; i <= 17; i++ {
				if w[i] != 0x7F {
					t.Errorf("enable[%d] = 0x%02X, want 0x7F", i, w[i])
				}
			}
		}
	}
	if enableCount != NumFrames {
		t.Errorf("expected %d enable writes (one per frame), got %d", NumFrames, enableCount)
	}
}

func TestInitErrorAtEachStage(t *testing.T) {
	// Init makes many Tx calls. Count them to know the max.
	mock := &mockConn{}
	dev := newTestDevice(mock)
	if err := dev.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	totalCalls := mock.calls

	// Verify that failing at each call position returns an error.
	for failAt := 1; failAt <= totalCalls; failAt++ {
		mock := &mockConn{failAt: failAt}
		dev := newTestDevice(mock)
		if err := dev.Init(); err == nil {
			t.Errorf("Init should fail when Tx fails at call %d/%d", failAt, totalCalls)
		}
	}
}

func TestShowPWM(t *testing.T) {
	mock := &mockConn{}
	dev := newTestDevice(mock)

	pwm := make([]byte, 135)
	pwm[0] = 0xFF
	pwm[134] = 0xAA

	if err := dev.ShowPWM(pwm); err != nil {
		t.Fatalf("ShowPWM failed: %v", err)
	}

	if len(mock.writes) < 4 {
		t.Fatalf("expected at least 4 writes, got %d", len(mock.writes))
	}

	// First write: bank select to frame 1.
	if mock.writes[0][0] != RegCommand || mock.writes[0][1] != 1 {
		t.Errorf("first write = %v, want bank select to frame 1", mock.writes[0])
	}

	// Second write: PWM data (OffsetPWM prefix + data).
	if mock.writes[1][0] != OffsetPWM {
		t.Errorf("PWM write starts with %d, want %d", mock.writes[1][0], OffsetPWM)
	}
	if mock.writes[1][1] != 0xFF {
		t.Errorf("PWM[0] = %d, want 0xFF", mock.writes[1][1])
	}

	if dev.frame != 1 {
		t.Errorf("frame = %d, want 1", dev.frame)
	}
}

func TestShowPWMDoubleBuffer(t *testing.T) {
	mock := &mockConn{}
	dev := newTestDevice(mock)
	pwm := make([]byte, 135)

	// First call writes to frame 1.
	if err := dev.ShowPWM(pwm); err != nil {
		t.Fatalf("first ShowPWM failed: %v", err)
	}
	if dev.frame != 1 {
		t.Errorf("after first show: frame = %d, want 1", dev.frame)
	}

	mock.writes = nil

	// Second call writes to frame 0.
	if err := dev.ShowPWM(pwm); err != nil {
		t.Fatalf("second ShowPWM failed: %v", err)
	}
	if dev.frame != 0 {
		t.Errorf("after second show: frame = %d, want 0", dev.frame)
	}

	if mock.writes[0][1] != 0 {
		t.Errorf("second show bank select = %d, want 0", mock.writes[0][1])
	}
}

func TestShowPWMErrorAtEachStage(t *testing.T) {
	// ShowPWM makes several Tx calls. Count them.
	mock := &mockConn{}
	dev := newTestDevice(mock)
	pwm := make([]byte, 135)
	if err := dev.ShowPWM(pwm); err != nil {
		t.Fatalf("ShowPWM failed: %v", err)
	}
	totalCalls := mock.calls

	for failAt := 1; failAt <= totalCalls; failAt++ {
		mock := &mockConn{failAt: failAt}
		dev := newTestDevice(mock)
		if err := dev.ShowPWM(pwm); err == nil {
			t.Errorf("ShowPWM should fail when Tx fails at call %d/%d", failAt, totalCalls)
		}
	}
}

func TestShutdown(t *testing.T) {
	mock := &mockConn{}
	dev := newTestDevice(mock)

	if err := dev.Shutdown(); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	last := mock.writes[len(mock.writes)-1]
	if last[0] != RegShutdown || last[1] != 0x00 {
		t.Errorf("last write = %v, want shutdown register cleared", last)
	}
}

func TestShutdownClearsCurrentFrame(t *testing.T) {
	mock := &mockConn{}
	dev := newTestDevice(mock)

	// Verify shutdown writes PWM clear data.
	if err := dev.Shutdown(); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	// Should have a PWM clear write (OffsetPWM + zeros).
	foundPWMClear := false
	for _, w := range mock.writes {
		if len(w) == NumLEDs+1 && w[0] == OffsetPWM {
			foundPWMClear = true
			for i := 1; i < len(w); i++ {
				if w[i] != 0 {
					t.Errorf("PWM clear byte %d = %d, want 0", i, w[i])
				}
			}
		}
	}
	if !foundPWMClear {
		t.Error("Shutdown did not write PWM clear data")
	}
}

func TestShutdownErrorAtEachStage(t *testing.T) {
	mock := &mockConn{}
	dev := newTestDevice(mock)
	if err := dev.Shutdown(); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
	totalCalls := mock.calls

	for failAt := 1; failAt <= totalCalls; failAt++ {
		mock := &mockConn{failAt: failAt}
		dev := newTestDevice(mock)
		if err := dev.Shutdown(); err == nil {
			t.Errorf("Shutdown should fail when Tx fails at call %d/%d", failAt, totalCalls)
		}
	}
}
