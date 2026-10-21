//go:build linux

package finalpi

import "testing"

func requireGPIOChip(t *testing.T) {
	t.Helper()
	// Use BCM26 — not uart0/uart2 header pins (4/5/14/15) when UART overlays + breadboard are live.
	p, err := openGpioLine(26, GpioOut, GpioPullOff, false)
	if err != nil {
		t.Skipf("gpiochip unavailable: %v", err)
	}
	_ = p.close()
}

func TestGpioPinActivateOutValueAndRaw(t *testing.T) {
	requireGPIOChip(t)
	c := GpioCtrlInst()
	c.Reset()
	c.gpio26.setDirection(GpioOut)
	c.gpio26.setPull(GpioPullOff)
	c.gpio26.SetValue(true)
	if err := c.gpio26.SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	defer c.Reset()
	if !c.gpio26.Active() {
		t.Fatal("expected active after SetEnabled(true)")
	}
	if !c.gpio26.Raw() {
		t.Fatal("raw should read driven high")
	}
	c.gpio26.SetValue(false)
	if c.gpio26.Raw() {
		t.Fatal("raw should read driven low")
	}
	if err := c.gpio26.SetEnabled(false); err != nil {
		t.Fatal(err)
	}
	if c.gpio26.Active() {
		t.Fatal("expected released after SetEnabled(false)")
	}
}

func TestGpioPinActivateInputPullUp(t *testing.T) {
	requireGPIOChip(t)
	c := GpioCtrlInst()
	c.Reset()
	c.gpio27.setDirection(GpioIn)
	c.gpio27.setPull(GpioPullUp)
	if err := c.gpio27.SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	defer c.Reset()
	if !c.gpio27.Active() {
		t.Fatal("expected active input")
	}
	_ = c.gpio27.Raw() // floating with pull-up; just ensure read works
}

func TestGpioPinSetEnabledConflictBeforeActivate(t *testing.T) {
	requireGPIOChip(t)
	c := GpioCtrlInst()
	c.Reset()
	if err := c.gpio12.SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	defer c.Reset()
	err := c.pwm0.SetEnabled(true)
	if err == nil {
		t.Fatal("expected conflict with gpio12")
	}
	if c.pwm0.Enabled() {
		t.Fatal("pwm0 should roll back")
	}
	if !c.gpio12.Active() {
		t.Fatal("gpio12 should stay active")
	}
}

func TestGpioPinResetReleasesToInput(t *testing.T) {
	requireGPIOChip(t)
	usePinmuxUnclaimed(t) // assert in-memory defaults for gpio4/12 after Reset
	c := GpioCtrlInst()
	c.Reset()
	c.gpio26.setDirection(GpioOut)
	c.gpio26.setPull(GpioPullOff)
	c.gpio26.SetValue(true)
	if err := c.gpio26.SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	if !c.gpio26.Raw() {
		t.Fatal("expected driven high before reset")
	}
	c.Reset()
	if c.gpio26.Active() || c.gpio26.Enabled {
		t.Fatal("expected released after reset")
	}
	if c.gpio26.Direction != GpioIn {
		t.Fatal("reset should restore direction to in")
	}
	if c.gpio4.Pull != GpioPullUp {
		t.Fatalf("gpio4 reset pull want up, got %v", c.gpio4.Pull)
	}
	if c.gpio12.Pull != GpioPullDown {
		t.Fatalf("gpio12 reset pull want down, got %v", c.gpio12.Pull)
	}
	// After release-as-input+default pull, reclaim with pull-down: pad should not stay pushed high.
	c.gpio26.setDirection(GpioIn)
	c.gpio26.setPull(GpioPullDown)
	if err := c.gpio26.SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	defer c.Reset()
	if c.gpio26.Raw() {
		t.Fatal("after reset→input, line should read low with pull-down")
	}
}
