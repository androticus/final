package finalpi

import (
	"fmt"
	"sync"

	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/internal/core"
)

type PiSupportType int

const (
	PiSupportNone PiSupportType = iota
	PiSupportPiZ2
	PiSupportPi5
)

var piSupport PiSupportType

// GpioCtrl owns header GPIO / peripheral configuration.
type GpioCtrl struct {
	pwm0 *PwmPeriph // BCM12
	pwm1 *PwmPeriph // BCM13
	// pwm2/pwm3 (BCM 14/15) not exposed — no stock pinmux recipe; 14/15 kept for uart0.

	// uart1 (GPIO0/1 HAT ID) and uart4 (GPIO12/13; clashes with spi5) are not supported.
	uart0 *UartPeriph // GPIO14/15 → ttyAMA0 (Pi5) / ttyS0 (Zero 2)
	uart2 *UartPeriph // GPIO4/5 → /dev/ttyAMA2 (uart2-pi5)
	uart3 *UartPeriph // GPIO8/9 → /dev/ttyAMA3 (uart3-pi5)

	// i2c0 (HAT ID EEPROM) not exposed. Header defaults: i2c1/2/3.
	i2c1 *I2cPeriph // GPIO2/3 → /dev/i2c-1
	i2c2 *I2cPeriph // GPIO4/5 → /dev/i2c-2
	i2c3 *I2cPeriph // GPIO6/7 → /dev/i2c-3

	spi0 *SpiPeriph // MOSI10 MISO9 SCLK11 CE0=8 CE1=7 → /dev/spidev0.0|.1 (max numCS=2)
	spi1 *SpiPeriph // MOSI20 MISO19 SCLK21 CE0=18 CE1=17 CE2=16 → spidev1.* (max numCS=3)
	spi5 *SpiPeriph // MOSI14 MISO13 SCLK15 CE0=12 CE1=26 → spidev5.0|.1 (max numCS=2; spi5-1cs/2cs)

	// Stub: header I2S (BCM 18–21). Final never drives audio; OS pinmux skip + .reserved stopgap.
	i2s *I2sPeriph

	// Header user GPIO: BCM 2–27 (0/1 HAT ID reserved — not exposed).
	gpio2, gpio3, gpio4, gpio5, gpio6, gpio7                     *GpioPin
	gpio8, gpio9, gpio10, gpio11, gpio12, gpio13, gpio14, gpio15 *GpioPin
	gpio16, gpio17, gpio18, gpio19, gpio20, gpio21               *GpioPin
	gpio22, gpio23, gpio24, gpio25, gpio26, gpio27               *GpioPin

	gpios   []*GpioPin
	periphs []Peripheral

	// osPinMap: BCM-indexed mapping labels (0..27); filled by reenumerateOsPinMap.
	osPinMap *core.StringList

	deferDepth int // Validate reentrancy guard
	poller     *gpioPoller
}

var (
	gpioCtrlInst *GpioCtrl
	gpioCtrlOnce sync.Once
)

// GpioCtrlInst returns the process singleton, creating the pin/periph graph on first access.
func GpioCtrlInst() *GpioCtrl {
	gpioCtrlOnce.Do(func() {
		gpioCtrlInst = newGpioCtrl()
	})
	return gpioCtrlInst
}

func newGpioCtrl() *GpioCtrl {
	c := &GpioCtrl{
		pwm0: newPwmPeriph("pwm0", 0, 12),
		pwm1: newPwmPeriph("pwm1", 1, 13),
		// TX/RX required; CTS/RTS hardwired pair claimed only when flowHW (overlay ctsrts).
		uart0:  newUartPeriph("uart0", 0, uartDevPath("FINAL_UART0", uart0DefaultPath(piSupport)), 14, 15, 16, 17),
		uart2:  newUartPeriph("uart2", 2, uartDevPath("FINAL_UART2", "/dev/ttyAMA2"), 4, 5, 6, 7),
		uart3:  newUartPeriph("uart3", 3, uartDevPath("FINAL_UART3", "/dev/ttyAMA3"), 8, 9, 10, 11),
		i2c1:   newI2cPeriph("i2c1", 1, 2, 3),
		i2c2:   newI2cPeriph("i2c2", 2, 4, 5),
		i2c3:   newI2cPeriph("i2c3", 3, 6, 7),
		spi0:   newSpiPeriph("spi0", 0, spiDevPath("FINAL_SPI0", "/dev/spidev0.0"), 10, 9, 11, []int{8, 7}),
		spi1:   newSpiPeriph("spi1", 1, spiDevPath("FINAL_SPI1", "/dev/spidev1.0"), 20, 19, 21, []int{18, 17, 16}),
		spi5:   newSpiPeriph("spi5", 5, spiDevPath("FINAL_SPI5", "/dev/spidev5.0"), 14, 13, 15, []int{12, 26}),
		i2s:    newI2sPeriph("i2s", 0, 18, 19, 20, 21),
		gpio2:  newGpioPin(2),
		gpio3:  newGpioPin(3),
		gpio4:  newGpioPin(4),
		gpio5:  newGpioPin(5),
		gpio6:  newGpioPin(6),
		gpio7:  newGpioPin(7),
		gpio8:  newGpioPin(8),
		gpio9:  newGpioPin(9),
		gpio10: newGpioPin(10),
		gpio11: newGpioPin(11),
		gpio12: newGpioPin(12),
		gpio13: newGpioPin(13),
		gpio14: newGpioPin(14),
		gpio15: newGpioPin(15),
		gpio16: newGpioPin(16),
		gpio17: newGpioPin(17),
		gpio18: newGpioPin(18),
		gpio19: newGpioPin(19),
		gpio20: newGpioPin(20),
		gpio21: newGpioPin(21),
		gpio22: newGpioPin(22),
		gpio23: newGpioPin(23),
		gpio24: newGpioPin(24),
		gpio25: newGpioPin(25),
		gpio26: newGpioPin(26),
		gpio27: newGpioPin(27),
		poller: newGpioPoller(),
	}
	c.gpios = make([]*GpioPin, 0, 26)
	for bcm := 2; bcm <= 27; bcm++ {
		c.gpios = append(c.gpios, c.gpioByBCM(bcm))
	}
	c.periphs = []Peripheral{
		c.pwm0, c.pwm1,
		c.uart0, c.uart2, c.uart3,
		c.i2c1, c.i2c2, c.i2c3,
		c.spi0, c.spi1, c.spi5,
		c.i2s,
	}
	return c
}

// gpioByBCM returns the stub GpioPin for header BCM 2–27.
func (c *GpioCtrl) gpioByBCM(bcm int) *GpioPin {
	switch bcm {
	case 2:
		return c.gpio2
	case 3:
		return c.gpio3
	case 4:
		return c.gpio4
	case 5:
		return c.gpio5
	case 6:
		return c.gpio6
	case 7:
		return c.gpio7
	case 8:
		return c.gpio8
	case 9:
		return c.gpio9
	case 10:
		return c.gpio10
	case 11:
		return c.gpio11
	case 12:
		return c.gpio12
	case 13:
		return c.gpio13
	case 14:
		return c.gpio14
	case 15:
		return c.gpio15
	case 16:
		return c.gpio16
	case 17:
		return c.gpio17
	case 18:
		return c.gpio18
	case 19:
		return c.gpio19
	case 20:
		return c.gpio20
	case 21:
		return c.gpio21
	case 22:
		return c.gpio22
	case 23:
		return c.gpio23
	case 24:
		return c.gpio24
	case 25:
		return c.gpio25
	case 26:
		return c.gpio26
	case 27:
		return c.gpio27
	default:
		panic(fmt.Errorf("GpioCtrl: BCM %d not in stub set (2–27)", bcm))
	}
}

// Reset releases held lines (as inputs with Pi reset pulls), disables all
// peripherals, and restores in-process pin/PWM/UART defaults. Always valid.
// Skips GPIO pad restore for OS pinmux-claimed BCMs and .reserved periph req pins
// (see 148.5). Reserved peripherals keep .reserved.
func (c *GpioCtrl) Reset() {
	skipBCM := make(map[int]struct{})
	c.reenumerateOsPinMap()
	for bcm := 2; bcm <= 27; bcm++ {
		if c.osPinMap.At(bcm) != "" {
			skipBCM[bcm] = struct{}{}
		}
	}
	for _, p := range c.periphs {
		if !p.Reserved() {
			continue
		}
		for _, lp := range p.ReqPins() {
			if lp != nil {
				skipBCM[lp.BCM] = struct{}{}
			}
		}
	}
	for _, g := range c.gpios {
		if _, skip := skipBCM[g.BCM]; skip {
			g.releaseSkipPad() // drop Final hold; do not steal kernel pinmux
			continue
		}
		g.restoreResetDefaults()
	}
	for _, p := range c.periphs {
		switch x := p.(type) {
		case *PwmPeriph:
			x.restoreResetDefaults()
		case *UartPeriph:
			x.restoreResetDefaults()
		case *I2cPeriph:
			x.restoreResetDefaults()
		case *SpiPeriph:
			x.restoreResetDefaults()
		}
		// *I2sPeriph: leave .reserved alone; no other state to restore.
	}
	c.poller.stop()
}

// Validate checks pin/periph enabled claims for BCM conflicts.
// Panics if called reentrantly (deferDepth != 0 on entry).
func (c *GpioCtrl) Validate() error {
	if c.deferDepth != 0 {
		panic("GpioCtrl.Validate: reentrant call")
	}
	c.deferDepth++
	defer func() { c.deferDepth-- }()

	claims := make(map[int]string) // BCM → owner label
	claim := func(bcm int, owner string) error {
		if prev, ok := claims[bcm]; ok {
			return fmt.Errorf("BCM %d claimed by both %s and %s", bcm, prev, owner)
		}
		claims[bcm] = owner
		return nil
	}

	for _, g := range c.gpios {
		if !g.Enabled {
			continue
		}
		if err := claim(g.BCM, fmt.Sprintf("gpio%d", g.BCM)); err != nil {
			return err
		}
	}
	for _, p := range c.periphs {
		if !p.Enabled() && !p.Reserved() {
			continue
		}
		for _, lp := range p.ReqPins() {
			if lp == nil {
				continue
			}
			if err := claim(lp.BCM, p.Name()); err != nil {
				return err
			}
		}
		for _, lp := range p.OptPins() {
			if lp == nil {
				continue
			}
			if err := claim(lp.BCM, p.Name()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *GpioCtrl) applyPinEnabled(p *GpioPin, enabled bool) error {
	prev := p.Enabled
	p.Enabled = enabled
	if !enabled {
		p.release()
		if c.deferDepth != 0 {
			return nil
		}
		return c.Validate()
	}
	if c.deferDepth != 0 {
		return nil
	}
	if err := c.Validate(); err != nil {
		p.Enabled = prev
		return err
	}
	if p.active {
		return nil
	}
	if err := p.activate(); err != nil {
		p.Enabled = prev
		return err
	}
	return nil
}

func (c *GpioCtrl) applyPeripheralEnabled(p *PeripheralBase, enabled bool) error {
	prev := p.enabled
	p.enabled = enabled
	if !enabled {
		p.active = false
	}
	if c.deferDepth != 0 {
		return nil
	}
	if err := c.Validate(); err != nil {
		p.enabled = prev
		return err
	}
	return nil
}

func (c *GpioCtrl) applyPeripheralReserved(p *PeripheralBase, reserved bool) error {
	prev := p.reserved
	p.reserved = reserved
	if c.deferDepth != 0 {
		return nil
	}
	if err := c.Validate(); err != nil {
		p.reserved = prev
		return err
	}
	return nil
}

func (c *GpioCtrl) applyPwmEnabled(p *PwmPeriph, enabled bool) error {
	prev := p.enabled
	p.enabled = enabled
	if !enabled {
		p.release()
		if c.deferDepth != 0 {
			return nil
		}
		return c.Validate()
	}
	if c.deferDepth != 0 {
		return nil
	}
	if err := c.Validate(); err != nil {
		p.enabled = prev
		return err
	}
	if p.active {
		return nil
	}
	if err := p.activate(); err != nil {
		p.enabled = prev
		return err
	}
	return nil
}

func (c *GpioCtrl) applyUartEnabled(u *UartPeriph, enabled bool) error {
	if enabled {
		if err := checkNeedsPi5(&u.PeripheralBase); err != nil {
			return err
		}
	}
	prev := u.enabled
	u.enabled = enabled
	if !enabled {
		u.release()
		if c.deferDepth != 0 {
			return nil
		}
		return c.Validate()
	}
	if c.deferDepth != 0 {
		return nil
	}
	if err := c.Validate(); err != nil {
		u.enabled = prev
		return err
	}
	if u.active {
		return nil
	}
	if err := u.activate(); err != nil {
		u.enabled = prev
		return err
	}
	return nil
}

func (c *GpioCtrl) applyI2cEnabled(b *I2cPeriph, enabled bool) error {
	if enabled {
		if err := checkNeedsPi5(&b.PeripheralBase); err != nil {
			return err
		}
	}
	prev := b.enabled
	b.enabled = enabled
	if !enabled {
		b.release()
		if c.deferDepth != 0 {
			return nil
		}
		return c.Validate()
	}
	if c.deferDepth != 0 {
		return nil
	}
	if err := c.Validate(); err != nil {
		b.enabled = prev
		return err
	}
	if b.active {
		return nil
	}
	if err := b.activate(); err != nil {
		b.enabled = prev
		return err
	}
	return nil
}

func (c *GpioCtrl) applySpiEnabled(s *SpiPeriph, enabled bool) error {
	if enabled {
		if err := checkNeedsPi5(&s.PeripheralBase); err != nil {
			return err
		}
	}
	prev := s.enabled
	s.enabled = enabled
	if !enabled {
		s.release()
		if c.deferDepth != 0 {
			return nil
		}
		return c.Validate()
	}
	if c.deferDepth != 0 {
		return nil
	}
	if err := c.Validate(); err != nil {
		s.enabled = prev
		return err
	}
	if s.active {
		return nil
	}
	if err := s.activate(); err != nil {
		s.enabled = prev
		return err
	}
	return nil
}

func registerGpioCtrl(ty *finalembed.Type) {
	ty.RegisterClassProperty("inst", "GpioCtrl",
		func(_ any) any { return GpioCtrlInst() },
		nil,
	)
	// Class property so GpioCtrl.osPinMap works (like .inst / .piSupport); get re-enumerates.
	ty.RegisterClassProperty("osPinMap", "StringList",
		func(_ any) any { return GpioCtrlInst().reenumerateOsPinMap() },
		nil,
	)
	ty.RegisterProperty("pwm0", "PwmPeriph",
		func(inst any) any { return inst.(*GpioCtrl).pwm0 },
		nil,
	)
	ty.RegisterProperty("pwm1", "PwmPeriph",
		func(inst any) any { return inst.(*GpioCtrl).pwm1 },
		nil,
	)
	ty.RegisterProperty("uart0", "UartPeriph",
		func(inst any) any { return inst.(*GpioCtrl).uart0 },
		nil,
	)
	ty.RegisterProperty("uart2", "UartPeriph",
		func(inst any) any { return inst.(*GpioCtrl).uart2 },
		nil,
	)
	ty.RegisterProperty("uart3", "UartPeriph",
		func(inst any) any { return inst.(*GpioCtrl).uart3 },
		nil,
	)
	ty.RegisterProperty("i2c1", "I2cPeriph",
		func(inst any) any { return inst.(*GpioCtrl).i2c1 },
		nil,
	)
	ty.RegisterProperty("i2c2", "I2cPeriph",
		func(inst any) any { return inst.(*GpioCtrl).i2c2 },
		nil,
	)
	ty.RegisterProperty("i2c3", "I2cPeriph",
		func(inst any) any { return inst.(*GpioCtrl).i2c3 },
		nil,
	)
	ty.RegisterProperty("spi0", "SpiPeriph",
		func(inst any) any { return inst.(*GpioCtrl).spi0 },
		nil,
	)
	ty.RegisterProperty("spi1", "SpiPeriph",
		func(inst any) any { return inst.(*GpioCtrl).spi1 },
		nil,
	)
	ty.RegisterProperty("spi5", "SpiPeriph",
		func(inst any) any { return inst.(*GpioCtrl).spi5 },
		nil,
	)
	ty.RegisterProperty("i2s", "I2sPeriph",
		func(inst any) any { return inst.(*GpioCtrl).i2s },
		nil,
	)
	for bcm := 2; bcm <= 27; bcm++ {
		bcm := bcm
		ty.RegisterProperty(fmt.Sprintf("gpio%d", bcm), "GpioPin",
			func(inst any) any { return inst.(*GpioCtrl).gpioByBCM(bcm) },
			nil,
		)
	}
	ty.RegisterClassProperty("piSupport", "Int",
		func(inst any) any { return core.Int(piSupport) },
		nil,
	)
	ty.RegisterNativeMethod("validate", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		if err := ths.Obj.(*GpioCtrl).Validate(); err != nil {
			panic(err)
		}
	})
	ty.RegisterNativeMethod("reset", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		ths.Obj.(*GpioCtrl).Reset()
	})
}
