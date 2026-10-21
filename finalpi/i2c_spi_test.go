package finalpi

import (
	"os"
	"strings"
	"testing"
)

func TestI2cSpiStubInstancesAndConflicts(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	defer c.Reset()

	if c.i2c1 == nil || c.i2c2 == nil || c.i2c3 == nil || c.spi0 == nil || c.spi1 == nil || c.spi5 == nil {
		t.Fatal("missing i2c/spi stubs")
	}
	if c.i2c1.bus != 1 || c.i2c1.path != "/dev/i2c-1" {
		t.Fatalf("i2c1 bus/path: %d %s", c.i2c1.bus, c.i2c1.path)
	}
	if c.spi0.path != "/dev/spidev0.0" || c.spi1.path != "/dev/spidev1.0" || c.spi5.path != "/dev/spidev5.0" {
		t.Fatalf("spi paths: %s %s %s", c.spi0.path, c.spi1.path, c.spi5.path)
	}

	c.i2c1.enabled = true
	c.gpio2.Enabled = true
	if err := c.Validate(); err == nil {
		t.Fatal("expected i2c1 vs gpio2 conflict")
	}
	c.gpio2.Enabled = false
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}

	c.i2c1.enabled = false
	c.i2c2.enabled = true
	c.uart2.enabled = true
	if err := c.Validate(); err == nil {
		t.Fatal("expected i2c2 vs uart2 conflict (GPIO4/5)")
	}
	c.i2c2.enabled = false
	c.uart2.enabled = false

	c.spi0.enabled = true
	c.gpio10.Enabled = true
	if err := c.Validate(); err == nil {
		t.Fatal("expected spi0 vs gpio10 conflict")
	}
	c.gpio10.Enabled = false
	c.uart3.enabled = true // GPIO8/9 — conflicts with spi0 CE0=8 and MISO=9
	if err := c.Validate(); err == nil {
		t.Fatal("expected spi0 vs uart3 conflict")
	}
	c.Reset()

	c.spi5.enabled = true
	c.pwm0.enabled = true // CE0=12
	if err := c.Validate(); err == nil {
		t.Fatal("expected spi5 vs pwm0 conflict (GPIO12)")
	}
	c.pwm0.enabled = false
	c.uart0.enabled = true // 14/15
	if err := c.Validate(); err == nil {
		t.Fatal("expected spi5 vs uart0 conflict")
	}
	c.Reset()
}

func TestI2cSpiSoftOpenMissingDevice(t *testing.T) {
	b := newI2cPeriph("i2cMissing", 99, 2, 3)
	b.path = "/dev/i2c_does_not_exist_phase8"
	err := b.activate()
	if err == nil {
		b.release()
		t.Fatal("expected I2C open error")
	}
	if !strings.Contains(err.Error(), "I2C open") {
		t.Fatalf("error=%v", err)
	}

	s := newSpiPeriph("spiMissing", 99, "/dev/spidev_does_not_exist_phase8", 10, 9, 11, []int{8})
	err = s.activate()
	if err == nil {
		s.release()
		t.Fatal("expected SPI open error")
	}
	if !strings.Contains(err.Error(), "SPI open") {
		t.Fatalf("error=%v", err)
	}
}

func TestSpiNumCSOptPinsAndLimits(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	defer c.Reset()

	if c.spi0.maxNumCS != 2 || c.spi1.maxNumCS != 3 || c.spi5.maxNumCS != 2 {
		t.Fatalf("maxNumCS spi0=%d spi1=%d spi5=%d", c.spi0.maxNumCS, c.spi1.maxNumCS, c.spi5.maxNumCS)
	}
	if len(c.spi0.OptPins()) != 0 {
		t.Fatal("default numCS=1 should leave optPins empty")
	}

	c.spi0.setNumCS(2)
	opts := c.spi0.OptPins()
	if len(opts) != 1 || opts[0].Name != "ce1" || opts[0].BCM != 7 {
		t.Fatalf("spi0 numCS=2 optPins=%v", opts)
	}
	c.spi0.enabled = true
	c.gpio7.Enabled = true
	if err := c.Validate(); err == nil {
		t.Fatal("expected spi0 CE1 vs gpio7 conflict")
	}
	c.gpio7.Enabled = false
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	c.Reset()

	c.spi5.setNumCS(2)
	opts = c.spi5.OptPins()
	if len(opts) != 1 || opts[0].Name != "ce1" || opts[0].BCM != 26 {
		t.Fatalf("spi5 numCS=2 optPins=%v", opts)
	}
	c.spi5.enabled = true
	c.gpio26.Enabled = true
	if err := c.Validate(); err == nil {
		t.Fatal("expected spi5 CE1 vs gpio26 conflict")
	}
	c.Reset()

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected spi5.numCS=3 panic")
			}
		}()
		c.spi5.setNumCS(3)
	}()

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected spi0.numCS=3 panic")
			}
		}()
		c.spi0.setNumCS(3)
	}()
}

func TestSpiSelCSBounds(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	defer c.Reset()

	c.spi0.setNumCS(2)
	c.spi0.setSelCS(1)
	if c.spi0.selCS != 1 {
		t.Fatal("selCS=1")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected selCS=2 panic with numCS=2")
			}
		}()
		c.spi0.setSelCS(2)
	}()
	c.spi0.setNumCS(1)
	if c.spi0.selCS != 0 {
		t.Fatalf("numCS shrink should reset selCS, got %d", c.spi0.selCS)
	}
}

func TestSpiPathForCE(t *testing.T) {
	s := newSpiPeriph("spi0", 0, "/dev/spidev0.0", 10, 9, 11, []int{8, 7})
	if got := s.pathForCE(0); got != "/dev/spidev0.0" {
		t.Fatalf("ce0 path=%s", got)
	}
	if got := s.pathForCE(1); got != "/dev/spidev0.1" {
		t.Fatalf("ce1 path=%s", got)
	}
}

func TestI2cSpiResetResets(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	c.i2c1.enabled = true
	c.spi0.enabled = true
	c.spi0.mode = 3
	c.spi0.numCS = 2
	c.spi0.selCS = 1
	c.spi0.syncCSOptPins()
	c.Reset()
	if c.i2c1.Enabled() || c.spi0.Enabled() || c.spi0.mode != 0 {
		t.Fatal("Reset should reset i2c/spi")
	}
	if c.spi0.numCS != 1 || c.spi0.selCS != 0 || len(c.spi0.OptPins()) != 0 {
		t.Fatalf("Reset should reset numCS/selCS/optPins: numCS=%d selCS=%d opts=%v",
			c.spi0.numCS, c.spi0.selCS, c.spi0.OptPins())
	}
}

func TestI2cHardwareOpenIfPresent(t *testing.T) {
	if _, err := os.Stat("/dev/i2c-1"); err != nil {
		t.Skip("no /dev/i2c-1 (enable i2c_arm / i2c1-pi5)")
	}
	c := GpioCtrlInst()
	c.Reset()
	defer c.Reset()
	if err := c.i2c1.SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	if !c.i2c1.Active() {
		t.Fatal("expected active")
	}
	if err := c.i2c1.SetEnabled(false); err != nil {
		t.Fatal(err)
	}
}

func TestSpiHardwareOpenIfPresent(t *testing.T) {
	if _, err := os.Stat("/dev/spidev0.0"); err != nil {
		t.Skip("no /dev/spidev0.0 (enable dtparam=spi / spi0-2cs)")
	}
	c := GpioCtrlInst()
	c.Reset()
	defer c.Reset()
	if err := c.spi0.SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	if !c.spi0.Active() {
		t.Fatal("expected active")
	}
	// Full-duplex transfer with no slave: still exercises ioctl path.
	if _, err := c.spi0.Transfer([]byte{0x00}); err != nil {
		t.Logf("transfer (may fail without slave): %v", err)
	}
	if err := c.spi0.SetEnabled(false); err != nil {
		t.Fatal(err)
	}
}
