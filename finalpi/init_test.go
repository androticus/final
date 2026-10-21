package finalpi

import (
	"testing"

	"github.com/androticus/final/finalembed"
)

func TestCoreTranche4RegistersTypesAndFoundation(t *testing.T) {
	// Package init already registered HAL types + armed tranche4; NewSession seals core.
	sess, err := finalembed.NewPersistentSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	if _, err := sess.GetSender().Send(`GpioCtrl.inst drop`, 10); err != nil {
		t.Fatal(err)
	}
	if _, err := sess.GetSender().Send(`Pin.boardToBcm^`, 10); err == nil {
		t.Fatal("expected Pin type to be unregistered")
	}

	// PiSupport pseudoclass + gpio.fnl names in sealed core.
	if _, err := sess.GetSender().Send(
		`PiSupport.None drop PiSupport.PiZ2 drop PiSupport.Pi5 drop`, 10); err != nil {
		t.Fatal(err)
	}
	if _, err := sess.GetSender().Send(
		`gpio2 drop gpio7 drop gpio14 drop gpio20 drop gpio25 drop gpio27 drop  pwm0 drop pwm1 drop  uart0 drop uart3 drop  i2c1 drop spi0 drop i2s drop`, 10); err != nil {
		t.Fatal(err)
	}
	// utils.fnl (tranche4 after gpio.fnl): pinMap is ( -- )
	if _, err := sess.GetSender().Send(`pinMap`, 10); err != nil {
		t.Fatal(err)
	}
	// Soft PWM raw config without activate.
	if _, err := sess.GetSender().Send(
		`pwm0 >ts PwmMode.raw t:.mode! t:.raw >xs 1000000 x:.periodNs! 250000 x:.dutyNs! x:.periodNs@ drop x:drop ts>`, 10); err != nil {
		t.Fatal(err)
	}
	if _, err := sess.GetSender().Send(`GpioCtrl.inst.reset^`, 10); err != nil {
		t.Fatal(err)
	}
	// Soft config (any host): direction + value without activate.
	if _, err := sess.GetSender().Send(
		`gpio4 >ts GpioPin.out t:.direction! true t:.value! t:.value@ drop ts>`, 10); err != nil {
		t.Fatal(err)
	}
	// Soft conflict via Validate (no GPIO activate).
	if _, err := sess.GetSender().Send(`GpioCtrl.inst.reset^`, 10); err != nil {
		t.Fatal(err)
	}
	c := GpioCtrlInst()
	c.gpio12.Enabled = true
	c.pwm0.enabled = true
	if err := c.Validate(); err == nil {
		t.Fatal("expected Validate conflict gpio12+pwm0")
	}
	c.Reset()

	// Hardware activate when gpiochip works (Pi); skip soft-only hosts.
	// Prefer BCM26 — avoids uart2 header pins (4/5) when overlays + breadboard are live.
	probe, err := openGpioLine(26, GpioOut, GpioPullOff, false)
	if err != nil {
		t.Logf("skip Final GPIO activate smoke: %v", err)
		return
	}
	_ = probe.close()

	if _, err := sess.GetSender().Send(
		`gpio26 >ts GpioPin.out t:.direction! true t:.value! true t:.enabled! t:.raw@ drop false t:.value! t:.raw@ drop false t:.enabled! ts>`, 10); err != nil {
		t.Fatal(err)
	}
	if _, err := sess.GetSender().Send(`GpioCtrl.inst.reset^`, 10); err != nil {
		t.Fatal(err)
	}
	if _, err := sess.GetSender().Send(`gpio12 >ts true t:.enabled! ts>`, 10); err != nil {
		t.Fatal(err)
	}
	if _, err := sess.GetSender().Send(`pwm0 >ts true t:.enabled! ts>`, 10); err == nil {
		t.Fatal("expected Validate conflict enabling pwm0 while gpio12 enabled")
	}
	if _, err := sess.GetSender().Send(`GpioCtrl.inst.reset^`, 10); err != nil {
		t.Fatal(err)
	}

	// PWM activate smoke when pwmchip works.
	ph, err := openPWM(12)
	if err != nil {
		t.Logf("skip Final PWM activate smoke: %v", err)
		return
	}
	_ = ph.Close()
	if _, err := sess.GetSender().Send(
		`pwm0 >ts PwmMode.raw t:.mode! t:.raw >xs 1000000 x:.periodNs! 0 x:.dutyNs! x:drop true t:.enabled! t:.raw >xs 500000 x:.dutyNs! x:drop false t:.enabled! ts>`, 10); err != nil {
		t.Fatal(err)
	}
	if _, err := sess.GetSender().Send(`GpioCtrl.inst.reset^`, 10); err != nil {
		t.Fatal(err)
	}
}
