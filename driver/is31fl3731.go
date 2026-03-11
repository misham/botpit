package driver

import (
	"time"

	"periph.io/x/conn/v3/i2c"
)

// Register addresses.
const (
	RegCommand   = 0xFD // Bank selection register
	RegConfig    = 0x0B // Function/config bank
	RegMode      = 0x00 // Display mode (in config bank)
	RegFrame     = 0x01 // Active frame (in config bank)
	RegAudioSync = 0x06 // Audio sync (in config bank)
	RegShutdown  = 0x0A // Shutdown control (in config bank)

	OffsetEnable = 0x00 // LED enable start (in frame bank)
	OffsetPWM    = 0x24 // PWM data start (in frame bank)

	ModePicture = 0x00
	NumFrames   = 8
	NumLEDs     = 144
	EnableBytes = 18
)

// Device represents an IS31FL3731 connected via I2C.
type Device struct {
	dev   conn
	frame byte // current active frame (0 or 1)
}

// conn is the interface for I2C communication (satisfied by *i2c.Dev).
type conn interface {
	Tx(w []byte, r []byte) error
}

// NewDevice creates a new IS31FL3731 device handle.
func NewDevice(bus i2c.Bus, addr uint16) *Device {
	return &Device{
		dev: &i2c.Dev{Bus: bus, Addr: addr},
	}
}

// Init initializes the IS31FL3731: soft reset, picture mode, clear all frames.
func (d *Device) Init() error {
	// Soft reset: shutdown then wake
	if err := d.writeReg(RegShutdown, 0x00); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	if err := d.writeReg(RegShutdown, 0x01); err != nil {
		return err
	}

	// Picture mode, frame 0, no audio sync
	if err := d.writeReg(RegMode, ModePicture); err != nil {
		return err
	}
	if err := d.writeReg(RegFrame, 0x00); err != nil {
		return err
	}
	if err := d.writeReg(RegAudioSync, 0x00); err != nil {
		return err
	}

	// Initialize all 8 frames: enable Scroll pHAT HD LEDs, clear PWM
	for frame := byte(0); frame < NumFrames; frame++ {
		if err := d.selectBank(frame); err != nil {
			return err
		}
		// Enable LEDs: 0x7F per column (7 rows), last byte 0x00 (unused)
		enable := make([]byte, EnableBytes+1)
		enable[0] = OffsetEnable
		for i := 1; i <= 17; i++ {
			enable[i] = 0x7F
		}
		enable[18] = 0x00
		if err := d.dev.Tx(enable, nil); err != nil {
			return err
		}
		// Clear all PWM values
		pwm := make([]byte, NumLEDs+1)
		pwm[0] = OffsetPWM
		if err := d.dev.Tx(pwm, nil); err != nil {
			return err
		}
	}

	d.frame = 0
	return nil
}

// Shutdown clears the display and puts the chip into software shutdown.
func (d *Device) Shutdown() error {
	// Clear current frame
	if err := d.selectBank(d.frame); err != nil {
		return err
	}
	pwm := make([]byte, NumLEDs+1)
	pwm[0] = OffsetPWM
	if err := d.dev.Tx(pwm, nil); err != nil {
		return err
	}
	return d.writeReg(RegShutdown, 0x00)
}

// ShowPWM writes a full PWM buffer to the next frame and switches to it.
// pwm must be at least 135 bytes (indices 0-134 used by Scroll pHAT HD).
func (d *Device) ShowPWM(pwm []byte) error {
	next := (d.frame + 1) % 2

	// Write PWM data to inactive frame
	if err := d.selectBank(next); err != nil {
		return err
	}
	buf := make([]byte, len(pwm)+1)
	buf[0] = OffsetPWM
	copy(buf[1:], pwm)
	if err := d.dev.Tx(buf, nil); err != nil {
		return err
	}

	// Switch active frame
	if err := d.writeReg(RegFrame, next); err != nil {
		return err
	}
	d.frame = next
	return nil
}

// selectBank writes to the command register to select a bank.
func (d *Device) selectBank(bank byte) error {
	return d.dev.Tx([]byte{RegCommand, bank}, nil)
}

// writeReg writes a value to a register in the config bank.
func (d *Device) writeReg(reg, value byte) error {
	if err := d.selectBank(RegConfig); err != nil {
		return err
	}
	return d.dev.Tx([]byte{reg, value}, nil)
}
