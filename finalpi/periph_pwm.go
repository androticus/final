package finalpi

import (
	"fmt"
	"math"

	"github.com/androticus/final/finalembed"
)

// PwmMode selects which facet may be read/written on a PwmPeriph.
type PwmMode byte

const (
	PwmModeRaw PwmMode = iota
	PwmModeLinear
	PwmModeServo
)

func (m PwmMode) String() string {
	switch m {
	case PwmModeRaw:
		return "raw"
	case PwmModeLinear:
		return "linear"
	case PwmModeServo:
		return "servo"
	default:
		return fmt.Sprintf("PwmMode(%d)", m)
	}
}

// PwmPeriph is a GpioCtrl-managed PWM channel (pwm0→12, pwm1→13).
//
// Backend: sysfs /sys/class/pwm/pwmchipN (pure Go). Sysfs "enable" is activate-only.
// Header pinmux: dtoverlay=pwm-2chan,pin=12,pin2=13,func=4,func2=4 (RP1 pwm@98000).
// Chip resolve is modal (Pi5 RP1 @98000 vs Z2 first non-fan); override via FINAL_PWMCHIP.
type PwmPeriph struct {
	PeripheralBase
	bcm  int
	mode PwmMode // default Raw

	periodNs uint64 // hardware truth; surfaced only via raw
	dutyNs   uint64

	linearFreqHz int     // Hz; 0 = unset (required before enable in Linear)
	linearDuty   float64 // 0..1; default 0

	servoFreqHz int     // Hz; default 50
	minUs       int     // µs; default 1000
	maxUs       int     // µs; default 2000
	asRange     float64 // definitive −1..1; default 0; asAngle derived

	raw    *PwmRaw
	linear *PwmLinear
	servo  *PwmServo
	hw     *pwmHandle
}

// Facet types hold a back-pointer; no independent period/duty state.
type PwmRaw struct{ p *PwmPeriph }
type PwmLinear struct{ p *PwmPeriph }
type PwmServo struct{ p *PwmPeriph }

func newPwmPeriph(name string, instance, bcm int) *PwmPeriph {
	p := &PwmPeriph{
		PeripheralBase: PeripheralBase{
			name:     name,
			instance: instance,
			reqPins:  []*LogiPin{{Name: "pwm", BCM: bcm}},
		},
		bcm:         bcm,
		mode:        PwmModeRaw,
		servoFreqHz: 50,
		minUs:       1000,
		maxUs:       2000,
		asRange:     0,
	}
	p.raw = &PwmRaw{p: p}
	p.linear = &PwmLinear{p: p}
	p.servo = &PwmServo{p: p}
	return p
}

func (p *PwmPeriph) SetEnabled(enabled bool) error {
	return GpioCtrlInst().applyPwmEnabled(p, enabled)
}

func (p *PwmPeriph) requireMode(want PwmMode, op string) {
	if p.mode != want {
		panic(fmt.Errorf("%s.%s: mode is %s (need %s)", p.name, op, p.mode, want))
	}
}

func (p *PwmPeriph) setMode(m PwmMode) {
	if p.active {
		panic(fmt.Errorf("%s.mode!: cannot change mode while active", p.name))
	}
	p.mode = m
}

func hzToPeriodNs(hz int) uint64 {
	if hz <= 0 {
		return 0
	}
	return uint64(math.Round(1e9 / float64(hz)))
}

func (p *PwmPeriph) pulseUsFromRange() float64 {
	mid := float64(p.minUs+p.maxUs) / 2
	half := float64(p.maxUs-p.minUs) / 2
	return mid + p.asRange*half
}

func (p *PwmPeriph) rangeFromAngle(deg float64) float64 {
	// 0° → −1, 90° → 0, 180° → +1
	return (deg - 90) / 90
}

func (p *PwmPeriph) angleFromRange() float64 {
	return 90 + p.asRange*90
}

func (p *PwmPeriph) syncFromMode() error {
	switch p.mode {
	case PwmModeRaw:
		if p.periodNs == 0 {
			return fmt.Errorf("%s: set raw.periodNs before enabled", p.name)
		}
		if p.dutyNs > p.periodNs {
			return fmt.Errorf("%s: raw.dutyNs %d > periodNs %d", p.name, p.dutyNs, p.periodNs)
		}
	case PwmModeLinear:
		if p.linearFreqHz <= 0 {
			return fmt.Errorf("%s: set linear.freq before enabled", p.name)
		}
		p.periodNs = hzToPeriodNs(p.linearFreqHz)
		if p.linearDuty < 0 || p.linearDuty > 1 {
			return fmt.Errorf("%s: linear.duty must be 0..1, got %v", p.name, p.linearDuty)
		}
		p.dutyNs = uint64(math.Round(p.linearDuty * float64(p.periodNs)))
	case PwmModeServo:
		if p.servoFreqHz <= 0 {
			return fmt.Errorf("%s: servo.freq must be > 0", p.name)
		}
		if p.minUs >= p.maxUs {
			return fmt.Errorf("%s: servo minUs >= maxUs", p.name)
		}
		p.periodNs = hzToPeriodNs(p.servoFreqHz)
		pulseUs := p.pulseUsFromRange()
		duty := uint64(math.Round(pulseUs * 1000)) // µs → ns
		if duty > p.periodNs {
			return fmt.Errorf("%s: servo pulse %g µs exceeds period", p.name, pulseUs)
		}
		p.dutyNs = duty
	default:
		return fmt.Errorf("%s: unknown mode %v", p.name, p.mode)
	}
	return nil
}

func (p *PwmPeriph) applyDutyLive() {
	if p.active && p.hw != nil {
		if err := p.hw.setDuty(p.dutyNs); err != nil {
			panic(err)
		}
	}
}

func (p *PwmPeriph) activate() error {
	if p.active {
		return nil
	}
	if err := p.syncFromMode(); err != nil {
		return err
	}
	hw, err := openPWM(p.bcm)
	if err != nil {
		return err
	}
	if err := hw.applyPeriodDuty(p.periodNs, p.dutyNs); err != nil {
		_ = hw.Close()
		return err
	}
	p.hw = hw
	p.active = true
	return nil
}

func (p *PwmPeriph) release() {
	if p.hw != nil {
		_ = p.hw.Close()
		p.hw = nil
	}
	p.active = false
}

func (p *PwmPeriph) restoreResetDefaults() {
	p.release()
	p.enabled = false
	p.mode = PwmModeRaw
	p.periodNs = 0
	p.dutyNs = 0
	p.linearFreqHz = 0
	p.linearDuty = 0
	p.servoFreqHz = 50
	p.minUs = 1000
	p.maxUs = 2000
	p.asRange = 0
}

// --- raw ---

func (r *PwmRaw) periodNs() uint64 {
	r.p.requireMode(PwmModeRaw, "raw.periodNs@")
	return r.p.periodNs
}

func (r *PwmRaw) setPeriodNs(n uint64) {
	r.p.requireMode(PwmModeRaw, "raw.periodNs!")
	if r.p.active {
		panic(fmt.Errorf("%s.raw.periodNs!: active", r.p.name))
	}
	if n > 0 && r.p.dutyNs > n {
		panic(fmt.Errorf("%s.raw.periodNs!: dutyNs %d > periodNs %d", r.p.name, r.p.dutyNs, n))
	}
	r.p.periodNs = n
}

func (r *PwmRaw) dutyNs() uint64 {
	r.p.requireMode(PwmModeRaw, "raw.dutyNs@")
	return r.p.dutyNs
}

func (r *PwmRaw) setDutyNs(n uint64) {
	r.p.requireMode(PwmModeRaw, "raw.dutyNs!")
	if r.p.periodNs == 0 && n > 0 {
		panic(fmt.Errorf("%s.raw.dutyNs!: periodNs is 0", r.p.name))
	}
	if n > r.p.periodNs {
		panic(fmt.Errorf("%s.raw.dutyNs!: %d > periodNs %d", r.p.name, n, r.p.periodNs))
	}
	r.p.dutyNs = n
	r.p.applyDutyLive()
}

// --- linear ---

func (l *PwmLinear) freq() int {
	l.p.requireMode(PwmModeLinear, "linear.freq@")
	return l.p.linearFreqHz
}

func (l *PwmLinear) setFreq(hz int) {
	l.p.requireMode(PwmModeLinear, "linear.freq!")
	if l.p.active {
		panic(fmt.Errorf("%s.linear.freq!: active", l.p.name))
	}
	if hz <= 0 {
		panic(fmt.Errorf("%s.linear.freq!: must be > 0", l.p.name))
	}
	l.p.linearFreqHz = hz
}

func (l *PwmLinear) duty() float64 {
	l.p.requireMode(PwmModeLinear, "linear.duty@")
	return l.p.linearDuty
}

func (l *PwmLinear) setDuty(d float64) {
	l.p.requireMode(PwmModeLinear, "linear.duty!")
	if d < 0 || d > 1 {
		panic(fmt.Errorf("%s.linear.duty!: must be 0..1, got %v", l.p.name, d))
	}
	l.p.linearDuty = d
	if l.p.linearFreqHz > 0 {
		l.p.periodNs = hzToPeriodNs(l.p.linearFreqHz)
		l.p.dutyNs = uint64(math.Round(d * float64(l.p.periodNs)))
		l.p.applyDutyLive()
	}
}

// --- servo ---

func (s *PwmServo) freq() int {
	s.p.requireMode(PwmModeServo, "servo.freq@")
	return s.p.servoFreqHz
}

func (s *PwmServo) setFreq(hz int) {
	s.p.requireMode(PwmModeServo, "servo.freq!")
	if s.p.active {
		panic(fmt.Errorf("%s.servo.freq!: active", s.p.name))
	}
	if hz <= 0 {
		panic(fmt.Errorf("%s.servo.freq!: must be > 0", s.p.name))
	}
	s.p.servoFreqHz = hz
}

func (s *PwmServo) minUs() int {
	s.p.requireMode(PwmModeServo, "servo.minUs@")
	return s.p.minUs
}

func (s *PwmServo) maxUs() int {
	s.p.requireMode(PwmModeServo, "servo.maxUs@")
	return s.p.maxUs
}

func (s *PwmServo) setMinMaxUs(minUs, maxUs int) {
	s.p.requireMode(PwmModeServo, "servo.setMinMaxUs")
	if s.p.active {
		panic(fmt.Errorf("%s.servo.setMinMaxUs: active", s.p.name))
	}
	if minUs >= maxUs {
		panic(fmt.Errorf("%s.servo.setMinMaxUs: minUs >= maxUs", s.p.name))
	}
	s.p.minUs = minUs
	s.p.maxUs = maxUs
}

func (s *PwmServo) asRange() float64 {
	s.p.requireMode(PwmModeServo, "servo.asRange@")
	return s.p.asRange
}

func (s *PwmServo) setAsRange(r float64) {
	s.p.requireMode(PwmModeServo, "servo.asRange!")
	if r < -1 || r > 1 {
		panic(fmt.Errorf("%s.servo.asRange!: must be -1..1, got %v", s.p.name, r))
	}
	s.p.asRange = r
	if s.p.servoFreqHz > 0 {
		s.p.periodNs = hzToPeriodNs(s.p.servoFreqHz)
		pulseUs := s.p.pulseUsFromRange()
		s.p.dutyNs = uint64(math.Round(pulseUs * 1000))
		if s.p.dutyNs > s.p.periodNs {
			panic(fmt.Errorf("%s.servo.asRange!: pulse exceeds period", s.p.name))
		}
		s.p.applyDutyLive()
	}
}

func (s *PwmServo) asAngle() float64 {
	s.p.requireMode(PwmModeServo, "servo.asAngle@")
	return s.p.angleFromRange()
}

func (s *PwmServo) setAsAngle(deg float64) {
	s.p.requireMode(PwmModeServo, "servo.asAngle!")
	if deg < 0 || deg > 180 {
		panic(fmt.Errorf("%s.servo.asAngle!: must be 0..180, got %v", s.p.name, deg))
	}
	s.setAsRange(s.p.rangeFromAngle(deg))
}

func registerPwmMode(ty *finalembed.Type) {
	ty.RegisterClassProperty("raw", "Int", func(_ any) any { return finalembed.Int(PwmModeRaw) }, nil)
	ty.RegisterClassProperty("linear", "Int", func(_ any) any { return finalembed.Int(PwmModeLinear) }, nil)
	ty.RegisterClassProperty("servo", "Int", func(_ any) any { return finalembed.Int(PwmModeServo) }, nil)
}

func registerPwmRaw(ty *finalembed.Type) {
	ty.RegisterProperty("periodNs", "Int",
		func(inst any) any { return finalembed.Int(inst.(*PwmRaw).periodNs()) },
		func(inst any, v any) { inst.(*PwmRaw).setPeriodNs(uint64(v.(finalembed.Int))) },
	)
	ty.RegisterProperty("dutyNs", "Int",
		func(inst any) any { return finalembed.Int(inst.(*PwmRaw).dutyNs()) },
		func(inst any, v any) { inst.(*PwmRaw).setDutyNs(uint64(v.(finalembed.Int))) },
	)
}

func registerPwmLinear(ty *finalembed.Type) {
	ty.RegisterProperty("freq", "Int",
		func(inst any) any { return finalembed.Int(inst.(*PwmLinear).freq()) },
		func(inst any, v any) { inst.(*PwmLinear).setFreq(int(v.(finalembed.Int))) },
	)
	ty.RegisterProperty("duty", "Float",
		func(inst any) any { return finalembed.Float(inst.(*PwmLinear).duty()) },
		func(inst any, v any) { inst.(*PwmLinear).setDuty(float64(v.(finalembed.Float))) },
	)
}

func registerPwmServo(ty *finalembed.Type) {
	ty.RegisterProperty("freq", "Int",
		func(inst any) any { return finalembed.Int(inst.(*PwmServo).freq()) },
		func(inst any, v any) { inst.(*PwmServo).setFreq(int(v.(finalembed.Int))) },
	)
	ty.RegisterProperty("minUs", "Int",
		func(inst any) any { return finalembed.Int(inst.(*PwmServo).minUs()) },
		nil,
	)
	ty.RegisterProperty("maxUs", "Int",
		func(inst any) any { return finalembed.Int(inst.(*PwmServo).maxUs()) },
		nil,
	)
	ty.RegisterProperty("asRange", "Float",
		func(inst any) any { return finalembed.Float(inst.(*PwmServo).asRange()) },
		func(inst any, v any) { inst.(*PwmServo).setAsRange(float64(v.(finalembed.Float))) },
	)
	ty.RegisterProperty("asAngle", "Float",
		func(inst any) any { return finalembed.Float(inst.(*PwmServo).asAngle()) },
		func(inst any, v any) { inst.(*PwmServo).setAsAngle(float64(v.(finalembed.Float))) },
	)
	ty.RegisterNativeMethod("setMinMaxUs", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		maxUs := int(f.RP.DataStack.Pop().(finalembed.Int))
		minUs := int(f.RP.DataStack.Pop().(finalembed.Int))
		ths.Obj.(*PwmServo).setMinMaxUs(minUs, maxUs)
	})
}

func registerPwmPeriph(ty *finalembed.Type) {
	ty.AddClassResolution("Peripheral")
	ty.RegisterProperty("bcm", "Int",
		func(inst any) any { return finalembed.Int(inst.(*PwmPeriph).bcm) },
		nil,
	)
	ty.RegisterProperty("mode", "Int",
		func(inst any) any { return finalembed.Int(inst.(*PwmPeriph).mode) },
		func(inst any, v any) {
			inst.(*PwmPeriph).setMode(PwmMode(v.(finalembed.Int)))
		},
	)
	ty.RegisterProperty("raw", "PwmRaw",
		func(inst any) any { return inst.(*PwmPeriph).raw },
		nil,
	)
	ty.RegisterProperty("linear", "PwmLinear",
		func(inst any) any { return inst.(*PwmPeriph).linear },
		nil,
	)
	ty.RegisterProperty("servo", "PwmServo",
		func(inst any) any { return inst.(*PwmPeriph).servo },
		nil,
	)
}
