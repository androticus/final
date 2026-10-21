package finalpi

import (
	"strings"
	"testing"
)

func TestInstLazyOnce(t *testing.T) {
	a := GpioCtrlInst()
	b := GpioCtrlInst()
	if a != b {
		t.Fatal("GpioCtrlInst() returned different pointers")
	}
	if a.pwm0 == nil || a.gpio12 == nil {
		t.Fatal("gpio ctrl graph not populated")
	}
}

func TestGpioCtrlHeaderPins2to27(t *testing.T) {
	c := GpioCtrlInst()
	if len(c.gpios) != 26 {
		t.Fatalf("gpios len=%d want 26", len(c.gpios))
	}
	for bcm := 2; bcm <= 27; bcm++ {
		p := c.gpioByBCM(bcm)
		if p == nil || p.BCM != bcm {
			t.Fatalf("gpio%d missing or bcm=%d", bcm, p.BCM)
		}
		if c.gpios[bcm-2] != p {
			t.Fatalf("gpios[%d] != gpioByBCM(%d)", bcm-2, bcm)
		}
	}
}

func TestValidateGPIO4Alone(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	c.gpio4.Enabled = true
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.gpio4.Active() {
		t.Fatal("Validate must not activate hardware")
	}
}

func TestValidatePWM0AloneClaims12(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	c.pwm0.enabled = true
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.pwm0.Active() {
		t.Fatal("Validate must not activate PWM hardware")
	}
}

func TestValidateGPIO12AndPWM0Conflict(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	c.gpio12.Enabled = true
	c.pwm0.enabled = true
	err := c.Validate()
	if err == nil {
		t.Fatal("expected BCM 12 conflict")
	}
	if !strings.Contains(err.Error(), "12") {
		t.Fatalf("error should mention BCM 12: %v", err)
	}
}

func TestSetEnabledPWM0ConflictRollsBack(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	// Soft claim gpio12 without activate so conflict path runs on any host.
	c.gpio12.Enabled = true
	err := c.pwm0.SetEnabled(true)
	if err == nil {
		t.Fatal("expected BCM 12 conflict")
	}
	if !strings.Contains(err.Error(), "12") {
		t.Fatalf("error should mention BCM 12: %v", err)
	}
	if c.pwm0.Enabled() {
		t.Fatal("pwm0 enabled should roll back on conflict")
	}
	if !c.gpio12.Enabled {
		t.Fatal("gpio12 should remain enabled")
	}
}

func TestValidateReentrantPanics(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	c.deferDepth = 1
	defer func() {
		c.deferDepth = 0
		if r := recover(); r == nil {
			t.Fatal("expected panic on reentrant Validate")
		}
	}()
	_ = c.Validate()
}

func TestGpioPinValueBeforeActive(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	c.gpio4.setDirection(GpioOut)
	c.gpio4.SetValue(true)
	if !c.gpio4.Value() {
		t.Fatal("value not stored")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected raw panic when not active")
			}
		}()
		_ = c.gpio4.Raw()
	}()
}

func TestGpioPinValuePanicActiveInput(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	c.gpio4.setDirection(GpioIn)
	c.gpio4.active = true
	defer func() { c.gpio4.active = false }()
	defer func() {
		if recover() == nil {
			t.Fatal("expected value! panic on active input")
		}
	}()
	c.gpio4.SetValue(true)
}

func TestDefaultGpioPullTable(t *testing.T) {
	for _, bcm := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8} {
		if defaultGpioPull(bcm) != GpioPullUp {
			t.Fatalf("BCM %d want pull-up", bcm)
		}
	}
	for bcm := 9; bcm <= 27; bcm++ {
		if defaultGpioPull(bcm) != GpioPullDown {
			t.Fatalf("BCM %d want pull-down", bcm)
		}
	}
}

func TestResetRestoresPullDefaultsAndDisablesPeriphs(t *testing.T) {
	usePinmuxUnclaimed(t) // exercise full pad restore, not OS skip
	c := GpioCtrlInst()
	c.Reset()
	c.gpio4.setPull(GpioPullOff)
	c.gpio12.setDirection(GpioOut)
	c.gpio12.setPull(GpioPullUp)
	c.pwm0.enabled = true
	c.pwm1.enabled = true
	c.uart3.enabled = true
	c.Reset()
	if c.gpio4.Pull != GpioPullUp || c.gpio4.Direction != GpioIn {
		t.Fatalf("gpio4 defaults: pull=%v dir=%v", c.gpio4.Pull, c.gpio4.Direction)
	}
	if c.gpio12.Pull != GpioPullDown || c.gpio12.Direction != GpioIn {
		t.Fatalf("gpio12 defaults: pull=%v dir=%v", c.gpio12.Pull, c.gpio12.Direction)
	}
	if c.pwm0.Enabled() || c.pwm1.Enabled() || c.uart3.Enabled() {
		t.Fatal("peripherals should be disabled after reset")
	}
}

