package finalpi

import (
	"fmt"
	"sync"
)

// line is the gpiochip handle (real on Linux, unused on stubs).
type line interface {
	SetValue(int) error
	Value() (int, error)
	Close() error
}

// Pin holds a gpiochip line fd while a GpioPin is active.
// Not registered as a Final type — use GpioPin / GpioCtrl.
type Pin struct {
	bcm    int
	output bool
	level  int // last driven level for outputs (0/1); used by toggle
	mu     sync.Mutex
	line   line
}

func (p *Pin) setLevel(v int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.output {
		return fmt.Errorf("pin gpio%d is not an output", p.bcm)
	}
	if p.line == nil {
		return fmt.Errorf("pin gpio%d is closed", p.bcm)
	}
	if err := p.line.SetValue(v); err != nil {
		return err
	}
	p.level = v
	return nil
}

func (p *Pin) toggle() error {
	p.mu.Lock()
	next := 1 - p.level
	p.mu.Unlock()
	return p.setLevel(next)
}

func (p *Pin) isHigh() (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.line == nil {
		return false, fmt.Errorf("pin gpio%d is closed", p.bcm)
	}
	v, err := p.line.Value()
	if err != nil {
		return false, err
	}
	return v != 0, nil
}

func (p *Pin) close() error {
	return p.closeAsInput(GpioPullOff)
}

// closeRaw frees the line fd without Reconfigure (keeps kernel pinmux alone).
func (p *Pin) closeRaw() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.line == nil {
		return nil
	}
	p.output = false
	err := p.line.Close()
	p.line = nil
	return err
}

// closeAsInput reconfigures to input with pull, then frees the line.
// Does not force a drive level (active-low loads must not be shoved low).
func (p *Pin) closeAsInput(pull GpioPull) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.line == nil {
		return nil
	}
	_ = lineAsInput(p.line, pull)
	p.output = false
	err := p.line.Close()
	p.line = nil
	return err
}
