package finalpi

import (
	"fmt"
	"strings"
	"testing"
)

func clearI2sReserved(c *GpioCtrl) {
	c.i2s.reserved = false
	c.Reset()
}

func TestI2sReservedClaimsVsGpio18(t *testing.T) {
	c := GpioCtrlInst()
	clearI2sReserved(c)
	defer clearI2sReserved(c)

	if err := c.i2s.SetReserved(true); err != nil {
		t.Fatal(err)
	}
	err := c.gpio18.SetEnabled(true)
	if err == nil {
		t.Fatal("expected BCM 18 conflict with reserved i2s")
	}
	if !strings.Contains(err.Error(), "18") {
		t.Fatalf("error should mention BCM 18: %v", err)
	}
	if c.gpio18.Enabled {
		t.Fatal("gpio18 enabled should roll back on conflict")
	}
	if !c.i2s.Reserved() {
		t.Fatal("i2s should stay reserved")
	}
}

func TestI2sReservedClaimsVsSpi1(t *testing.T) {
	c := GpioCtrlInst()
	clearI2sReserved(c)
	defer clearI2sReserved(c)

	if err := c.i2s.SetReserved(true); err != nil {
		t.Fatal(err)
	}
	err := c.spi1.SetEnabled(true)
	if err == nil {
		t.Fatal("expected conflict enabling spi1 while i2s reserved")
	}
	if !strings.Contains(err.Error(), "i2s") && !strings.Contains(err.Error(), "spi1") {
		t.Fatalf("error should mention owners: %v", err)
	}
	if c.spi1.Enabled() {
		t.Fatal("spi1 enabled should roll back")
	}
}

func TestI2sReservedRejectsWhenGpioAlreadyClaims(t *testing.T) {
	c := GpioCtrlInst()
	clearI2sReserved(c)
	defer clearI2sReserved(c)

	c.gpio18.Enabled = true
	err := c.i2s.SetReserved(true)
	if err == nil {
		t.Fatal("expected conflict reserving i2s while gpio18 enabled")
	}
	if c.i2s.Reserved() {
		t.Fatal("i2s reserved should roll back on conflict")
	}
}

func TestI2sResetSkipsReservedPinsAndKeepsReserved(t *testing.T) {
	usePinmuxUnclaimed(t) // only .reserved should skip — not live OS overlays
	c := GpioCtrlInst()
	clearI2sReserved(c)
	defer clearI2sReserved(c)

	if err := c.i2s.SetReserved(true); err != nil {
		t.Fatal(err)
	}
	c.gpio18.setDirection(GpioOut)
	c.gpio18.setPull(GpioPullUp)
	c.gpio18.value = true
	c.gpio4.setPull(GpioPullOff)
	c.pwm0.enabled = true

	c.Reset()

	if !c.i2s.Reserved() {
		t.Fatal("Reset must not clear i2s.reserved")
	}
	if c.gpio18.Direction != GpioOut || c.gpio18.Pull != GpioPullUp || !c.gpio18.value {
		t.Fatalf("reserved BCM18 in-memory state should be left alone: dir=%v pull=%v value=%v",
			c.gpio18.Direction, c.gpio18.Pull, c.gpio18.value)
	}
	if c.gpio4.Pull != GpioPullUp {
		t.Fatalf("non-reserved gpio4 should restore pull-up, got %v", c.gpio4.Pull)
	}
	if c.pwm0.Enabled() {
		t.Fatal("other peripherals should still be disabled after Reset")
	}
}

func TestI2sSetEnabledTruePanics(t *testing.T) {
	c := GpioCtrlInst()
	clearI2sReserved(c)
	defer clearI2sReserved(c)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on SetEnabled(true)")
		}
		if !strings.Contains(fmt.Sprint(r), "reserved") {
			t.Fatalf("panic should mention .reserved: %v", r)
		}
	}()
	_ = c.i2s.SetEnabled(true)
}
