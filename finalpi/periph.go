package finalpi

import (
	"fmt"

	"github.com/androticus/final/finalembed"
)

type PeripheralKind byte

const (
	PeripheralKindUART PeripheralKind = iota
	PeripheralKindSPI
	PeripheralKindI2C
	PeripheralKindPWM
	PeripheralKindI2S
)

// Peripheral is the Go + Final shared surface for GpioCtrl-managed peripherals.
type Peripheral interface {
	Name() string
	Instance() int
	ReqPins() []*LogiPin
	OptPins() []*LogiPin
	Enabled() bool
	SetEnabled(bool) error
	Active() bool
	Reserved() bool
}

type PeripheralBase struct {
	name     string // eg i2c1
	instance int
	reqPins  []*LogiPin // fixed from mux
	optPins  []*LogiPin // eg RTS/CTS; fixed from mux when used
	enabled  bool
	reserved bool // Final exposes only on I2sPeriph; others stay false
	active   bool // Go-only: hardware claimed; Final not exposed yet
	needsPi5 bool // uart2/3, i2c2/3, spi1/spi5 — refuse enable unless PiSupportPi5
}

// checkNeedsPi5 refuses Pi 5–only alts on classic / none profiles before any Linux I/O.
func checkNeedsPi5(p *PeripheralBase) error {
	if p.needsPi5 && piSupport != PiSupportPi5 {
		return fmt.Errorf("%s: requires Pi 5 (piSupport=%d)", p.name, piSupport)
	}
	return nil
}

func (p *PeripheralBase) Name() string {
	return p.name
}
func (p *PeripheralBase) Instance() int {
	return p.instance
}
func (p *PeripheralBase) ReqPins() []*LogiPin {
	return p.reqPins
}
func (p *PeripheralBase) OptPins() []*LogiPin {
	return p.optPins
}
func (p *PeripheralBase) Enabled() bool {
	return p.enabled
}
func (p *PeripheralBase) Reserved() bool {
	return p.reserved
}
func (p *PeripheralBase) Active() bool {
	return p.active
}

func (p *PeripheralBase) SetEnabled(enabled bool) error {
	return GpioCtrlInst().applyPeripheralEnabled(p, enabled)
}

func registerPeripheral(ty *finalembed.Type) {
	ty.RegisterProperty("enabled", "Bool",
		func(inst any) any { return finalembed.Bool(inst.(Peripheral).Enabled()) },
		func(inst any, v any) {
			if err := inst.(Peripheral).SetEnabled(bool(v.(finalembed.Bool))); err != nil {
				panic(err)
			}
		},
	)
	ty.RegisterProperty("osMapped", "String",
		func(inst any) any {
			return finalembed.String(GpioCtrlInst().osMappedForPeriph(inst.(Peripheral)))
		},
		nil,
	)
}
