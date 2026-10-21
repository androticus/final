package finalpi

import (
	"fmt"
	"sync"

	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/internal/core"
)

type LogiPin struct {
	Name string
	BCM  int
}

type GpioPull byte

const (
	GpioPullOff GpioPull = iota
	GpioPullDown
	GpioPullUp
)

type GpioDirection byte

const (
	GpioIn GpioDirection = iota
	GpioOut
)

// defaultGpioPull is the Pi SoC reset bias for a BCM header line (0–27).
func defaultGpioPull(bcm int) GpioPull {
	if bcm >= 0 && bcm <= 8 {
		return GpioPullUp
	}
	if bcm >= 9 && bcm <= 27 {
		return GpioPullDown
	}
	return GpioPullOff
}

func newGpioPin(bcm int) *GpioPin {
	return &GpioPin{
		BCM:       bcm,
		Direction: GpioIn,
		Pull:      defaultGpioPull(bcm),
	}
}

// hint values for debounceMs
type DebounceValue byte

const (
	DebounceOff    DebounceValue = 0
	DebounceSensor DebounceValue = 5
	DebounceButton DebounceValue = 20
	DebounceSwitch DebounceValue = 30 // eg a limit switch
)

// GpioPin is a GpioCtrl-managed BCM GPIO line (config + gpiochip claim while active).
type GpioPin struct {
	BCM        int
	Enabled    bool
	active     bool // Go-only: hardware claimed; Final not exposed yet
	Direction  GpioDirection
	Pull       GpioPull
	DebounceMs DebounceValue // 0 = off; else leading-edge + clamp blanking (ms)
	value      bool          // out latch; input get uses Value() (logical or live)
	logical    bool          // last stable logical (inputs)
	haveLogic  bool
	hw         *Pin

	mu     sync.Mutex
	onHigh core.HandlerMintEventFn
	onLow  core.HandlerMintEventFn
}

func (p *GpioPin) SetEnabled(enabled bool) error {
	return GpioCtrlInst().applyPinEnabled(p, enabled)
}

func (p *GpioPin) Active() bool {
	return p.active
}

func (p *GpioPin) Value() bool {
	if p.active && p.Direction == GpioIn {
		// Debounce on (or level handlers): poller owns last-stable logical.
		if log, ok := GpioCtrlInst().poller.logicalFor(p); ok {
			return log
		}
		if p.haveLogic {
			return p.logical
		}
		// Debounce off / not watched: instantaneous level (same as raw).
		return p.Raw()
	}
	return p.value
}

func (p *GpioPin) SetValue(v bool) {
	if p.active && p.Direction == GpioIn {
		panic(fmt.Errorf("GpioPin.value!: gpio%d is active input", p.BCM))
	}
	p.value = v
	if p.active && p.Direction == GpioOut && p.hw != nil {
		level := 0
		if v {
			level = 1
		}
		if err := p.hw.setLevel(level); err != nil {
			panic(err)
		}
	}
}

func (p *GpioPin) Raw() bool {
	if !p.active || p.hw == nil {
		panic(fmt.Errorf("GpioPin.raw: gpio%d is not active", p.BCM))
	}
	high, err := p.hw.isHigh()
	if err != nil {
		panic(err)
	}
	return high
}

func (p *GpioPin) sampleRaw() (bool, bool) {
	if !p.active || p.hw == nil {
		return false, false
	}
	high, err := p.hw.isHigh()
	if err != nil {
		return false, false
	}
	return high, true
}

func (p *GpioPin) setLogical(v bool) {
	p.logical = v
	p.haveLogic = true
	if p.Direction == GpioIn {
		p.value = v
	}
}

func (p *GpioPin) wantsWatch() bool {
	return p.active && p.Direction == GpioIn &&
		(p.DebounceMs > 0 || p.onHigh != nil || p.onLow != nil)
}

func (p *GpioPin) setDirection(d GpioDirection) {
	if p.active {
		panic(fmt.Errorf("GpioPin.direction!: gpio%d is active", p.BCM))
	}
	p.Direction = d
}

func (p *GpioPin) setPull(pull GpioPull) {
	if p.active {
		panic(fmt.Errorf("GpioPin.pull!: gpio%d is active", p.BCM))
	}
	p.Pull = pull
}

func (p *GpioPin) setDebounceMs(ms DebounceValue) {
	p.DebounceMs = ms
	GpioCtrlInst().poller.syncWatch(p)
}

func (p *GpioPin) setOnHigh(ref *core.MethodRef, index int) {
	p.mu.Lock()
	p.onHigh = mintPinHandler(ref, index)
	p.mu.Unlock()
	GpioCtrlInst().poller.syncWatch(p)
}

func (p *GpioPin) setOnLow(ref *core.MethodRef, index int) {
	p.mu.Lock()
	p.onLow = mintPinHandler(ref, index)
	p.mu.Unlock()
	GpioCtrlInst().poller.syncWatch(p)
}

func (p *GpioPin) activate() error {
	if p.active {
		return nil
	}
	hw, err := openGpioLine(p.BCM, p.Direction, p.Pull, p.value)
	if err != nil {
		return err
	}
	p.hw = hw
	p.active = true
	p.haveLogic = false
	GpioCtrlInst().poller.syncWatch(p)
	return nil
}

func (p *GpioPin) release() {
	p.releaseWithPull(p.Pull)
}

func (p *GpioPin) releaseWithPull(pull GpioPull) {
	GpioCtrlInst().poller.remove(p)
	if p.hw != nil {
		_ = p.hw.closeAsInput(pull)
		p.hw = nil
	}
	p.active = false
	p.haveLogic = false
}

func (p *GpioPin) restoreResetDefaults() {
	def := defaultGpioPull(p.BCM)
	p.releaseWithPull(def)
	p.Enabled = false
	p.Direction = GpioIn
	p.Pull = def
	p.value = false
	p.DebounceMs = DebounceOff
	p.mu.Lock()
	p.onHigh = nil
	p.onLow = nil
	p.mu.Unlock()
}

// releaseSkipPad drops Final's line hold / enable without Reconfigure (OS pinmux /
// .reserved skip path). Leaves Direction/Pull/value for inspection.
func (p *GpioPin) releaseSkipPad() {
	GpioCtrlInst().poller.remove(p)
	if p.hw != nil {
		_ = p.hw.closeRaw()
		p.hw = nil
	}
	p.active = false
	p.haveLogic = false
	p.Enabled = false
	p.mu.Lock()
	p.onHigh = nil
	p.onLow = nil
	p.mu.Unlock()
}

func registerGpioPin(ty *finalembed.Type) {
	ty.RegisterProperty("bcm", "Int",
		func(inst any) any { return finalembed.Int(inst.(*GpioPin).BCM) },
		nil,
	)
	ty.RegisterProperty("osMapped", "String",
		func(inst any) any {
			return finalembed.String(GpioCtrlInst().osMappedForBCM(inst.(*GpioPin).BCM))
		},
		nil,
	)
	ty.RegisterProperty("enabled", "Bool",
		func(inst any) any { return finalembed.Bool(inst.(*GpioPin).Enabled) },
		func(inst any, v any) {
			if err := inst.(*GpioPin).SetEnabled(bool(v.(finalembed.Bool))); err != nil {
				panic(err)
			}
		},
	)
	ty.RegisterProperty("direction", "Int",
		func(inst any) any { return finalembed.Int(inst.(*GpioPin).Direction) },
		func(inst any, v any) {
			inst.(*GpioPin).setDirection(GpioDirection(v.(finalembed.Int)))
		},
	)
	ty.RegisterProperty("pull", "Int",
		func(inst any) any { return finalembed.Int(inst.(*GpioPin).Pull) },
		func(inst any, v any) {
			inst.(*GpioPin).setPull(GpioPull(v.(finalembed.Int)))
		},
	)
	ty.RegisterProperty("debounceMs", "Int",
		func(inst any) any { return finalembed.Int(inst.(*GpioPin).DebounceMs) },
		func(inst any, v any) {
			inst.(*GpioPin).setDebounceMs(DebounceValue(v.(finalembed.Int)))
		},
	)
	ty.RegisterProperty("value", "Bool",
		func(inst any) any { return finalembed.Bool(inst.(*GpioPin).Value()) },
		func(inst any, v any) {
			inst.(*GpioPin).SetValue(bool(v.(finalembed.Bool)))
		},
	)
	ty.RegisterProperty("raw", "Bool",
		func(inst any) any { return finalembed.Bool(inst.(*GpioPin).Raw()) },
		nil,
	)
	// ( handler:MethodRef index:Int -- )
	ty.RegisterNativeMethod("setOnHigh", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		index := int(f.RP.DataStack.Pop().(finalembed.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		ths.Obj.(*GpioPin).setOnHigh(handler, index)
	})
	// ( handler:MethodRef index:Int -- )
	ty.RegisterNativeMethod("setOnLow", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		index := int(f.RP.DataStack.Pop().(finalembed.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		ths.Obj.(*GpioPin).setOnLow(handler, index)
	})

	ty.RegisterClassProperty("in", "Int", func(_ any) any { return finalembed.Int(GpioIn) }, nil)
	ty.RegisterClassProperty("out", "Int", func(_ any) any { return finalembed.Int(GpioOut) }, nil)
	ty.RegisterClassProperty("pullOff", "Int", func(_ any) any { return finalembed.Int(GpioPullOff) }, nil)
	ty.RegisterClassProperty("pullDown", "Int", func(_ any) any { return finalembed.Int(GpioPullDown) }, nil)
	ty.RegisterClassProperty("pullUp", "Int", func(_ any) any { return finalembed.Int(GpioPullUp) }, nil)
	ty.RegisterClassProperty("debounceOff", "Int", func(_ any) any { return finalembed.Int(DebounceOff) }, nil)
	ty.RegisterClassProperty("debounceSensor", "Int", func(_ any) any { return finalembed.Int(DebounceSensor) }, nil)
	ty.RegisterClassProperty("debounceButton", "Int", func(_ any) any { return finalembed.Int(DebounceButton) }, nil)
	ty.RegisterClassProperty("debounceSwitch", "Int", func(_ any) any { return finalembed.Int(DebounceSwitch) }, nil)
}
