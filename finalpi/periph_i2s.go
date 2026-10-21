package finalpi

import (
	"fmt"

	"github.com/androticus/final/finalembed"
)

// I2sPeriph is a stub for header I2S (BCM 18–21). Final never drives audio;
// set .reserved so Reset leaves those pins alone and Validate blocks Final claims.
type I2sPeriph struct {
	PeripheralBase
}

func newI2sPeriph(name string, instance int, bclk, lrclk, din, dout int) *I2sPeriph {
	return &I2sPeriph{
		PeripheralBase: PeripheralBase{
			name:     name,
			instance: instance,
			reqPins: []*LogiPin{
				{Name: "bclk", BCM: bclk},
				{Name: "lrclk", BCM: lrclk},
				{Name: "din", BCM: din},
				{Name: "dout", BCM: dout},
			},
		},
	}
}

func (p *I2sPeriph) Enabled() bool {
	return false
}

func (p *I2sPeriph) SetEnabled(enabled bool) error {
	if enabled {
		panic(fmt.Errorf("%s: cannot enable; use .reserved for ALSA coexistence", p.name))
	}
	return nil
}

func (p *I2sPeriph) SetReserved(reserved bool) error {
	return GpioCtrlInst().applyPeripheralReserved(&p.PeripheralBase, reserved)
}

func registerI2sPeriph(ty *finalembed.Type) {
	// No AddClassResolution("Peripheral") — avoid inheriting .enabled.
	ty.RegisterProperty("osMapped", "String",
		func(inst any) any {
			return finalembed.String(GpioCtrlInst().osMappedForPeriph(inst.(*I2sPeriph)))
		},
		nil,
	)
	ty.RegisterProperty("reserved", "Bool",
		func(inst any) any { return finalembed.Bool(inst.(*I2sPeriph).Reserved()) },
		func(inst any, v any) {
			if err := inst.(*I2sPeriph).SetReserved(bool(v.(finalembed.Bool))); err != nil {
				panic(err)
			}
		},
	)
}
