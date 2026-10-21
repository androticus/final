package finalpi

import "time"

// debounceFilter implements leading-edge + clamp (148.1 / 148.2).
// clampMs 0 → logical tracks raw immediately; first sample never emits an edge.
// clampMs > 0 → first raw change commits logical + edge immediately, then blank
// for clampMs so bounce cannot emit further edges (value stays clamped).
type debounceFilter struct {
	clampMs    int
	logical    bool
	have       bool
	clampUntil time.Time
	clamping   bool
}

func (f *debounceFilter) setClampMs(ms int) {
	if ms < 0 {
		ms = 0
	}
	f.clampMs = ms
	f.clamping = false
}

// Sample updates the filter. rose/fell are logical edges (not first sample).
func (f *debounceFilter) Sample(raw bool, now time.Time) (logical bool, rose, fell bool) {
	if f.clampMs <= 0 {
		if !f.have {
			f.logical = raw
			f.have = true
			return f.logical, false, false
		}
		if raw == f.logical {
			return f.logical, false, false
		}
		prev := f.logical
		f.logical = raw
		return f.logical, raw && !prev, !raw && prev
	}

	if !f.have {
		f.logical = raw
		f.have = true
		f.clamping = false
		return f.logical, false, false
	}

	if f.clamping {
		if now.Before(f.clampUntil) {
			return f.logical, false, false
		}
		f.clamping = false
	}

	if raw == f.logical {
		return f.logical, false, false
	}

	prev := f.logical
	f.logical = raw
	f.clamping = true
	f.clampUntil = now.Add(time.Duration(f.clampMs) * time.Millisecond)
	return f.logical, raw && !prev, !raw && prev
}

func (f *debounceFilter) Logical() (bool, bool) {
	return f.logical, f.have
}
