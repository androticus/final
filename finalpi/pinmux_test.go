package finalpi

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

var errNoPinsFixture = errors.New("no pins fixture")

const pinmuxFixtureHeader = `Pinmux settings per pin
Format: pin (name): mux_owner gpio_owner hog?
`

const pinmuxFixturePi5Sample = pinmuxFixtureHeader + `
pin 0 (gpio0): (MUX UNCLAIMED) (GPIO UNCLAIMED)
pin 1 (gpio1): (MUX UNCLAIMED) (GPIO UNCLAIMED)
pin 2 (gpio2): 1f00074000.i2c (GPIO UNCLAIMED) function i2c1 group gpio2
pin 3 (gpio3): 1f00074000.i2c (GPIO UNCLAIMED) function i2c1 group gpio3
pin 4 (gpio4): 1f00038000.serial (GPIO UNCLAIMED) function uart2 group gpio4
pin 5 (gpio5): 1f00038000.serial (GPIO UNCLAIMED) function uart2 group gpio5
pin 12 (gpio12): 1f00098000.pwm (GPIO UNCLAIMED) function pwm0 group gpio12
pin 13 (gpio13): 1f00098000.pwm (GPIO UNCLAIMED) function pwm0 group gpio13
pin 14 (gpio14): 1f00030000.serial (GPIO UNCLAIMED) function uart0 group gpio14
pin 15 (gpio15): 1f00030000.serial (GPIO UNCLAIMED) function uart0 group gpio15
pin 18 (gpio18): (MUX UNCLAIMED) (GPIO UNCLAIMED)
`

func usePinmuxFixture(t *testing.T, content string) {
	t.Helper()
	prev := pinmuxPinsRead
	pinmuxPinsRead = func() ([]byte, error) { return []byte(content), nil }
	t.Cleanup(func() { pinmuxPinsRead = prev })
	// Default: no pins fallback (avoids live /sys merge when tests run on a Pi).
	prevPins := pinctrlPinsRead
	pinctrlPinsRead = func() ([]byte, error) { return nil, errNoPinsFixture }
	t.Cleanup(func() { pinctrlPinsRead = prevPins })
}

func usePinmuxUnclaimed(t *testing.T) {
	t.Helper()
	var b strings.Builder
	b.WriteString(pinmuxFixtureHeader)
	for bcm := 0; bcm <= 27; bcm++ {
		n := strconv.Itoa(bcm)
		b.WriteString("pin ")
		b.WriteString(n)
		b.WriteString(" (gpio")
		b.WriteString(n)
		b.WriteString("): (MUX UNCLAIMED) (GPIO UNCLAIMED)\n")
	}
	usePinmuxFixture(t, b.String())
}

func TestParsePinmuxPinsLabels(t *testing.T) {
	labels := make([]string, osPinMapLen)
	parsePinmuxPins([]byte(pinmuxFixturePi5Sample), labels)
	if labels[0] != "" || labels[1] != "" || labels[18] != "" {
		t.Fatalf("unclaimed want empty: 0=%q 1=%q 18=%q", labels[0], labels[1], labels[18])
	}
	if labels[14] != "uart0" || labels[15] != "uart0" {
		t.Fatalf("uart0 pins: 14=%q 15=%q", labels[14], labels[15])
	}
	if labels[12] != "pwm0" || labels[13] != "pwm0" {
		t.Fatalf("pwm pins: 12=%q 13=%q", labels[12], labels[13])
	}
	if labels[2] != "i2c1" {
		t.Fatalf("i2c1: got %q", labels[2])
	}
	if labels[4] != "uart2" {
		t.Fatalf("uart2: got %q", labels[4])
	}
}

func TestParsePinmuxPinsOwnerFallback(t *testing.T) {
	labels := make([]string, osPinMapLen)
	parsePinmuxPins([]byte("pin 9 (gpio9): 1f00050000.spi (GPIO UNCLAIMED)\n"), labels)
	if labels[9] != "spi" {
		t.Fatalf("owner fallback: got %q want spi", labels[9])
	}
}

const pinmuxFixtureZ2Sample = `Pinmux settings per pin
Format: pin (name): mux_owner|gpio_owner (strict) hog?
pin 0 (gpio0): UNCLAIMED
pin 1 (gpio1): UNCLAIMED
pin 2 (gpio2): device 3f804000.i2c function alt0 group gpio2
pin 3 (gpio3): device 3f804000.i2c function alt0 group gpio3
pin 6 (gpio6): GPIO pinctrl-bcm2835:518
pin 7 (gpio7): device 3f204000.spi function gpio_out group gpio7
pin 8 (gpio8): device 3f204000.spi function gpio_out group gpio8
pin 9 (gpio9): device 3f204000.spi function alt0 group gpio9
pin 10 (gpio10): device 3f204000.spi function alt0 group gpio10
pin 11 (gpio11): device 3f204000.spi function alt0 group gpio11
pin 12 (gpio12): device 3f20c000.pwm function alt5 group gpio12
pin 13 (gpio13): device 3f20c000.pwm function alt5 group gpio13
pin 14 (gpio14): UNCLAIMED
pin 15 (gpio15): UNCLAIMED
pin 26 (gpio26): GPIO pinctrl-bcm2835:538
pin 27 (gpio27): GPIO pinctrl-bcm2835:539
`

// Sibling "pins" dump: Z2 mini-UART is alt5 with no pinmux device owner.
const pinctrlPinsFixtureZ2Sample = `Pinmux settings per pin
Format: type is XX nnnnnnnn.nnnnnnnnnn (hex 8+4)~function~group; # defines max of 32 aliases per pin
pin 0 (gpio0) 0:gpio_in 1:gpio_out 2:alt5 3:alt4 4:alt0 5:alt1 6:alt2 7:alt3 function gpio_in in hi; irq 0 (none)
pin 12 (gpio12) 0:? function alt5 in lo; irq 0 (none)
pin 13 (gpio13) 0:? function alt5 in lo; irq 0 (none)
pin 14 (gpio14) 0:? function alt5 in hi; irq 0 (none)
pin 15 (gpio15) 0:? function alt5 in hi; irq 0 (none)
pin 22 (gpio22) 0:? function gpio_in in hi; irq 0 (none)
`

func TestParsePinmuxPinsBCM2835FriendlyLabels(t *testing.T) {
	labels := make([]string, osPinMapLen)
	parsePinmuxPins([]byte(pinmuxFixtureZ2Sample), labels)
	if labels[0] != "" || labels[14] != "" || labels[15] != "" {
		t.Fatalf("UNCLAIMED want empty: 0=%q 14=%q 15=%q", labels[0], labels[14], labels[15])
	}
	if labels[6] != "" || labels[26] != "" || labels[27] != "" {
		t.Fatalf("gpiochip-only claims must be empty (Reset must restore): 6=%q 26=%q 27=%q",
			labels[6], labels[26], labels[27])
	}
	if labels[2] != "i2c" || labels[12] != "pwm" || labels[9] != "spi" {
		t.Fatalf("device labels: i2c=%q pwm=%q spi=%q", labels[2], labels[12], labels[9])
	}
	if labels[7] != "spi" {
		t.Fatalf("spi CE gpio_out should still label spi, got %q", labels[7])
	}
}

func TestMergeAltFunctionLabelsZ2MiniUART(t *testing.T) {
	labels := make([]string, osPinMapLen)
	parsePinmuxPins([]byte(pinmuxFixtureZ2Sample), labels)
	mergeAltFunctionLabels([]byte(pinctrlPinsFixtureZ2Sample), labels)
	if labels[14] != "uart" || labels[15] != "uart" {
		t.Fatalf("mini-UART alt5: 14=%q 15=%q want uart", labels[14], labels[15])
	}
	// pinmux device claim wins over pins alt5 (pwm stays "pwm", not overwritten).
	if labels[12] != "pwm" || labels[13] != "pwm" {
		t.Fatalf("pwm device labels must win: 12=%q 13=%q", labels[12], labels[13])
	}
	if labels[22] != "" {
		t.Fatalf("gpio_in must stay empty, got %q", labels[22])
	}
}

func useZ2PinmuxAndPins(t *testing.T) {
	t.Helper()
	usePinmuxFixture(t, pinmuxFixtureZ2Sample)
	prev := pinctrlPinsRead
	pinctrlPinsRead = func() ([]byte, error) { return []byte(pinctrlPinsFixtureZ2Sample), nil }
	t.Cleanup(func() { pinctrlPinsRead = prev })
}

func TestResetRestoresGpiochipHeldPins(t *testing.T) {
	// After a demo, pinmux shows "GPIO pinctrl-…"; those must not skip Reset.
	usePinmuxFixture(t, pinmuxFixtureZ2Sample)
	c := GpioCtrlInst()
	c.i2s.reserved = false
	c.Reset()
	c.gpio6.setDirection(GpioOut)
	c.gpio6.setPull(GpioPullOff)
	c.gpio12.setDirection(GpioOut) // OS pwm — should skip
	c.gpio12.setPull(GpioPullUp)
	c.Reset()
	if c.gpio6.Direction != GpioIn || c.gpio6.Pull != GpioPullUp {
		t.Fatalf("gpio6 should full-restore: dir=%v pull=%v", c.gpio6.Direction, c.gpio6.Pull)
	}
	if c.gpio12.Direction != GpioOut || c.gpio12.Pull != GpioPullUp {
		t.Fatalf("pwm-mapped gpio12 should skip restore: dir=%v pull=%v", c.gpio12.Direction, c.gpio12.Pull)
	}
}

func TestOsMappedPeriphClassicLabels(t *testing.T) {
	useZ2PinmuxAndPins(t)
	c := GpioCtrlInst()
	if got := c.osMappedForPeriph(c.pwm0); got != "pwm0" {
		t.Fatalf("pwm0.osMapped=%q", got)
	}
	if got := c.osMappedForPeriph(c.i2c1); got != "i2c1" {
		t.Fatalf("i2c1.osMapped=%q", got)
	}
	if got := c.osMappedForPeriph(c.spi0); got != "spi0" {
		t.Fatalf("spi0.osMapped=%q", got)
	}
	if got := c.osMappedForPeriph(c.uart0); got != "uart0" {
		t.Fatalf("uart0.osMapped=%q (Z2 mini-UART alt5 fallback)", got)
	}
}

func TestReenumerateZ2MiniUARTViaPins(t *testing.T) {
	useZ2PinmuxAndPins(t)
	c := GpioCtrlInst()
	list := c.reenumerateOsPinMap()
	if list.At(14) != "uart" || list.At(15) != "uart" {
		t.Fatalf("osPinMap uart: 14=%q 15=%q", list.At(14), list.At(15))
	}
	if got := c.osMappedForBCM(14); got != "uart" {
		t.Fatalf("osMappedForBCM(14)=%q", got)
	}
}

func TestReenumerateOsPinMapAndGetters(t *testing.T) {
	usePinmuxFixture(t, pinmuxFixturePi5Sample)
	c := GpioCtrlInst()
	list := c.reenumerateOsPinMap()
	if list.Len() != osPinMapLen {
		t.Fatalf("len=%d want %d", list.Len(), osPinMapLen)
	}
	if list.At(0) != "" || list.At(14) != "uart0" {
		t.Fatalf("list[0]=%q list[14]=%q", list.At(0), list.At(14))
	}
	if got := c.osMappedForBCM(14); got != "uart0" {
		t.Fatalf("osMappedForBCM(14)=%q", got)
	}
	if got := c.osMappedForPeriph(c.uart0); got != "uart0" {
		t.Fatalf("uart0.osMapped=%q", got)
	}
	if got := c.osMappedForPeriph(c.pwm0); got != "pwm0" {
		t.Fatalf("pwm0.osMapped=%q", got)
	}
	if got := c.osMappedForPeriph(c.pwm1); got != "pwm1" {
		t.Fatalf("pwm1.osMapped=%q (kernel function pwm0 still matches)", got)
	}
	if got := c.osMappedForPeriph(c.uart2); got != "uart2" {
		t.Fatalf("uart2.osMapped=%q", got)
	}
	if got := c.osMappedForPeriph(c.i2s); got != "" {
		t.Fatalf("i2s.osMapped want empty, got %q", got)
	}
	if got := c.osMappedForPeriph(c.uart3); got != "" {
		t.Fatalf("uart3.osMapped want empty, got %q", got)
	}
}

func TestResetSkipsOsMappedPins(t *testing.T) {
	usePinmuxFixture(t, pinmuxFixturePi5Sample)
	c := GpioCtrlInst()
	c.i2s.reserved = false
	c.Reset()

	c.gpio14.setDirection(GpioOut)
	c.gpio14.setPull(GpioPullUp)
	c.gpio14.value = true
	c.gpio22.setPull(GpioPullOff)

	c.Reset()

	if c.gpio14.Direction != GpioOut || c.gpio14.Pull != GpioPullUp || !c.gpio14.value {
		t.Fatalf("OS-mapped BCM14 should skip restore: dir=%v pull=%v value=%v",
			c.gpio14.Direction, c.gpio14.Pull, c.gpio14.value)
	}
	if c.gpio22.Pull != GpioPullDown {
		t.Fatalf("unmapped gpio22 should restore pull-down, got %v", c.gpio22.Pull)
	}
}

func TestResetOsMappedUnionReserved(t *testing.T) {
	usePinmuxUnclaimed(t)
	c := GpioCtrlInst()
	c.i2s.reserved = false
	c.Reset()
	if err := c.i2s.SetReserved(true); err != nil {
		t.Fatal(err)
	}
	c.gpio18.setDirection(GpioOut)
	c.gpio18.setPull(GpioPullUp)
	c.Reset()
	if c.gpio18.Direction != GpioOut {
		t.Fatal("reserved i2s should still skip BCM18 with empty pinmux")
	}
}
