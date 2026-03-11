package display

// PWMShower writes PWM data to the LED driver.
type PWMShower interface {
	ShowPWM(pwm []byte) error
}
