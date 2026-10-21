package finalpi

import (
	"strings"
	"testing"
)

func TestAltPeriphsNeedPi5Flag(t *testing.T) {
	c := GpioCtrlInst()
	alts := []struct {
		name string
		p    *PeripheralBase
		want bool
	}{
		{"uart0", &c.uart0.PeripheralBase, false},
		{"uart2", &c.uart2.PeripheralBase, true},
		{"uart3", &c.uart3.PeripheralBase, true},
		{"i2c1", &c.i2c1.PeripheralBase, false},
		{"i2c2", &c.i2c2.PeripheralBase, true},
		{"i2c3", &c.i2c3.PeripheralBase, true},
		{"spi0", &c.spi0.PeripheralBase, false},
		{"spi1", &c.spi1.PeripheralBase, true},
		{"spi5", &c.spi5.PeripheralBase, true},
		{"pwm0", &c.pwm0.PeripheralBase, false},
	}
	for _, tc := range alts {
		if tc.p.needsPi5 != tc.want {
			t.Errorf("%s.needsPi5=%v want %v", tc.name, tc.p.needsPi5, tc.want)
		}
	}
}

func TestAltEnableRefusedOnPi2ZProfile(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	defer c.Reset()

	prev := piSupport
	piSupport = PiSupportPiZ2
	defer func() { piSupport = prev }()

	alts := []struct {
		name string
		en   func(bool) error
	}{
		{"uart2", c.uart2.SetEnabled},
		{"uart3", c.uart3.SetEnabled},
		{"i2c2", c.i2c2.SetEnabled},
		{"i2c3", c.i2c3.SetEnabled},
		{"spi1", c.spi1.SetEnabled},
		{"spi5", c.spi5.SetEnabled},
	}
	for _, tc := range alts {
		err := tc.en(true)
		if err == nil {
			t.Fatalf("%s.SetEnabled(true): expected requires-Pi5 error under Pi2Z", tc.name)
		}
		if !strings.Contains(err.Error(), "requires Pi 5") {
			t.Fatalf("%s: want requires Pi 5, got %v", tc.name, err)
		}
		if tc.name == "uart2" && c.uart2.Enabled() {
			t.Fatal("uart2 must stay disabled after refused enable")
		}
	}

	// Defaults still allowed through the needsPi5 gate (may fail later on missing device).
	if err := checkNeedsPi5(&c.uart0.PeripheralBase); err != nil {
		t.Fatalf("uart0 should not need Pi5: %v", err)
	}
	if err := checkNeedsPi5(&c.i2c1.PeripheralBase); err != nil {
		t.Fatalf("i2c1 should not need Pi5: %v", err)
	}
	if err := checkNeedsPi5(&c.spi0.PeripheralBase); err != nil {
		t.Fatalf("spi0 should not need Pi5: %v", err)
	}
}

func TestAltEnableAllowedOnPi5Profile(t *testing.T) {
	prev := piSupport
	piSupport = PiSupportPi5
	defer func() { piSupport = prev }()

	c := GpioCtrlInst()
	for _, p := range []*PeripheralBase{
		&c.uart2.PeripheralBase,
		&c.i2c2.PeripheralBase,
		&c.spi1.PeripheralBase,
	} {
		if err := checkNeedsPi5(p); err != nil {
			t.Fatalf("%s: unexpected gate under Pi5: %v", p.name, err)
		}
	}
}
