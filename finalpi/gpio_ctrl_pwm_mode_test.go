package finalpi

import "testing"

func TestPwmModeAndFacetSoftRules(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	c.pwm0.setMode(PwmModeRaw)
	c.pwm0.raw.setPeriodNs(1_000_000)
	c.pwm0.raw.setDutyNs(250_000)
	if c.pwm0.periodNs != 1_000_000 || c.pwm0.dutyNs != 250_000 {
		t.Fatal("raw period/duty not stored")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected linear.freq! panic in Raw mode")
			}
		}()
		c.pwm0.linear.setFreq(1000)
	}()
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected linear.freq@ panic in Raw mode")
			}
		}()
		_ = c.pwm0.linear.freq()
	}()

	c.pwm0.setMode(PwmModeServo)
	if c.pwm0.servo.asAngle() != 90 {
		t.Fatalf("default asAngle want 90 got %v", c.pwm0.servo.asAngle())
	}
	c.pwm0.servo.setAsAngle(0)
	if c.pwm0.asRange != -1 {
		t.Fatalf("asRange want -1 got %v", c.pwm0.asRange)
	}
	c.pwm0.servo.setAsRange(1)
	if c.pwm0.servo.asAngle() != 180 {
		t.Fatalf("asAngle want 180 got %v", c.pwm0.servo.asAngle())
	}
	c.pwm0.servo.setMinMaxUs(500, 2500)
	if c.pwm0.minUs != 500 || c.pwm0.maxUs != 2500 {
		t.Fatal("setMinMaxUs failed")
	}

	c.pwm0.active = true
	defer func() { c.pwm0.active = false }()
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected mode! panic while active")
			}
		}()
		c.pwm0.setMode(PwmModeRaw)
	}()
}

func TestPwmLinearRequiresFreqOnEnable(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	c.pwm0.setMode(PwmModeLinear)
	// duty default 0; freq unset
	err := c.pwm0.SetEnabled(true)
	if err == nil {
		// may fail at openPWM on non-linux; still must fail sync first
		if c.pwm0.Active() {
			c.Reset()
			t.Fatal("linear enable without freq must fail")
		}
	}
	if c.pwm0.Enabled() {
		t.Fatal("enabled should roll back")
	}
}
