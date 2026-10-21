//go:build linux

package finalpi

import (
	"strconv"
	"testing"
)

func requirePwmChip(t *testing.T) {
	t.Helper()
	h, err := openPWM(12)
	if err != nil {
		t.Skipf("pwmchip unavailable: %v", err)
	}
	_ = h.Close()
}

func TestPwmActivateRawPeriodDutyAndLiveDuty(t *testing.T) {
	requirePwmChip(t)
	c := GpioCtrlInst()
	c.Reset()
	c.pwm0.setMode(PwmModeRaw)
	c.pwm0.raw.setPeriodNs(1_000_000)
	c.pwm0.raw.setDutyNs(250_000)
	if err := c.pwm0.SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	defer c.Reset()
	if !c.pwm0.Active() {
		t.Fatal("expected active after SetEnabled(true)")
	}
	period, err := c.pwm0.hw.readAttr("period")
	if err != nil {
		t.Fatal(err)
	}
	if period != "1000000" {
		t.Fatalf("sysfs period=%s", period)
	}
	c.pwm0.raw.setDutyNs(0)
	duty, err := c.pwm0.hw.readAttr("duty_cycle")
	if err != nil {
		t.Fatal(err)
	}
	if duty != "0" {
		t.Fatalf("live duty 0 got %s", duty)
	}
	c.pwm0.raw.setDutyNs(500_000)
	duty, err = c.pwm0.hw.readAttr("duty_cycle")
	if err != nil {
		t.Fatal(err)
	}
	if duty != "500000" {
		t.Fatalf("live duty got %s", duty)
	}
	if err := c.pwm0.SetEnabled(false); err != nil {
		t.Fatal(err)
	}
	if c.pwm0.Active() {
		t.Fatal("expected released")
	}
}

func TestPwmActivateLinear(t *testing.T) {
	requirePwmChip(t)
	c := GpioCtrlInst()
	c.Reset()
	c.pwm0.setMode(PwmModeLinear)
	c.pwm0.linear.setFreq(1000)
	c.pwm0.linear.setDuty(0.25)
	if err := c.pwm0.SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	defer c.Reset()
	period, _ := c.pwm0.hw.readAttr("period")
	if period != "1000000" {
		t.Fatalf("period=%s want 1000000", period)
	}
	duty, _ := c.pwm0.hw.readAttr("duty_cycle")
	if duty != "250000" {
		t.Fatalf("duty=%s want 250000", duty)
	}
}

func TestPwm1ActivateChannel(t *testing.T) {
	requirePwmChip(t)
	c := GpioCtrlInst()
	c.Reset()
	c.pwm1.setMode(PwmModeRaw)
	c.pwm1.raw.setPeriodNs(2_000_000)
	c.pwm1.raw.setDutyNs(1_000_000)
	if err := c.pwm1.SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	defer c.Reset()
	if c.pwm1.hw.channel != 1 {
		t.Fatalf("pwm1 channel=%d", c.pwm1.hw.channel)
	}
	got, err := c.pwm1.hw.readAttr("period")
	if err != nil {
		t.Fatal(err)
	}
	if got != strconv.FormatUint(2_000_000, 10) {
		t.Fatalf("period=%s", got)
	}
}

func TestPwmActivateRequiresPeriodInRaw(t *testing.T) {
	requirePwmChip(t)
	c := GpioCtrlInst()
	c.Reset()
	c.pwm0.setMode(PwmModeRaw)
	c.pwm0.periodNs = 0
	c.pwm0.dutyNs = 0
	err := c.pwm0.SetEnabled(true)
	if err == nil {
		c.Reset()
		t.Fatal("expected error when periodNs==0")
	}
	if c.pwm0.Enabled() || c.pwm0.Active() {
		t.Fatal("should roll back on activate failure")
	}
}

func TestPwmConflictWithGpio12(t *testing.T) {
	requirePwmChip(t)
	c := GpioCtrlInst()
	c.Reset()
	c.pwm0.setMode(PwmModeRaw)
	c.pwm0.raw.setPeriodNs(1_000_000)
	if err := c.pwm0.SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	defer c.Reset()
	err := c.gpio12.SetEnabled(true)
	if err == nil {
		t.Fatal("expected conflict")
	}
	if !c.pwm0.Active() {
		t.Fatal("pwm0 should stay active")
	}
}
