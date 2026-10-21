package finalpi

import (
	"sync"
	"time"

	"github.com/androticus/final/internal/core"
)

const gpioPollInterval = 2 * time.Millisecond

type pinWatch struct {
	pin    *GpioPin
	filter debounceFilter
}

// gpioPoller is the shared Phase 7 sample loop for all watched GpioPins.
type gpioPoller struct {
	mu      sync.Mutex
	watches map[*GpioPin]*pinWatch
	worker  *Worker
}

func newGpioPoller() *gpioPoller {
	return &gpioPoller{
		watches: make(map[*GpioPin]*pinWatch),
		worker:  NewWorker(),
	}
}

func (p *gpioPoller) syncWatch(pin *GpioPin) {
	p.mu.Lock()
	if !pin.wantsWatch() {
		delete(p.watches, pin)
		needStop := len(p.watches) == 0
		w := p.worker
		p.mu.Unlock()
		if needStop {
			_ = w.Stop(2 * time.Second)
		}
		return
	}
	w, ok := p.watches[pin]
	if !ok {
		w = &pinWatch{pin: pin}
		p.watches[pin] = w
	}
	w.filter.setClampMs(int(pin.DebounceMs))
	needStart := !p.worker.Running()
	p.mu.Unlock()
	if needStart {
		// Re-check under lock in Go path — ensureRunning may race with remove.
		p.mu.Lock()
		if !p.worker.Running() && len(p.watches) > 0 {
			p.worker.Go(p.loop)
		}
		p.mu.Unlock()
	}
}

func (p *gpioPoller) remove(pin *GpioPin) {
	p.mu.Lock()
	delete(p.watches, pin)
	needStop := len(p.watches) == 0
	w := p.worker
	p.mu.Unlock()
	if needStop {
		_ = w.Stop(2 * time.Second)
	}
}

func (p *gpioPoller) stop() {
	p.mu.Lock()
	p.watches = make(map[*GpioPin]*pinWatch)
	w := p.worker
	p.mu.Unlock()
	_ = w.Stop(2 * time.Second)
}

func (p *gpioPoller) loop(cancel <-chan struct{}) {
	for {
		if WaitCancel(cancel, gpioPollInterval) == WaitCanceled {
			return
		}
		p.tick(time.Now())
	}
}

func (p *gpioPoller) tick(now time.Time) {
	p.mu.Lock()
	list := make([]*pinWatch, 0, len(p.watches))
	for _, w := range p.watches {
		list = append(list, w)
	}
	p.mu.Unlock()

	for _, w := range list {
		pin := w.pin
		if !pin.Active() || pin.Direction != GpioIn {
			continue
		}
		raw, ok := pin.sampleRaw()
		if !ok {
			continue
		}
		logical, rose, fell := w.filter.Sample(raw, now)
		pin.setLogical(logical)
		pin.mu.Lock()
		onHigh := pin.onHigh
		onLow := pin.onLow
		pin.mu.Unlock()
		if rose && onHigh != nil {
			onHigh(pin)
		}
		if fell && onLow != nil {
			onLow(pin)
		}
	}
}

func (p *gpioPoller) logicalFor(pin *GpioPin) (bool, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	w, ok := p.watches[pin]
	if !ok {
		return false, false
	}
	return w.filter.Logical()
}

func mintPinHandler(handler *core.MethodRef, index int) core.HandlerMintEventFn {
	return MintEvent(handler, index)
}
