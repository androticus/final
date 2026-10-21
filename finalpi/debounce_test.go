package finalpi

import (
	"testing"
	"time"
)

func TestDebounceOffImmediateEdges(t *testing.T) {
	var f debounceFilter
	now := time.Now()
	_, rose, fell := f.Sample(false, now)
	if rose || fell {
		t.Fatal("no edge on first sample")
	}
	_, rose, fell = f.Sample(true, now)
	if !rose || fell {
		t.Fatalf("want rise, rose=%v fell=%v", rose, fell)
	}
	_, rose, fell = f.Sample(false, now)
	if rose || !fell {
		t.Fatalf("want fall, rose=%v fell=%v", rose, fell)
	}
}

func TestDebounceLeadingClamp(t *testing.T) {
	f := debounceFilter{}
	f.setClampMs(20)
	t0 := time.Unix(0, 0)
	f.Sample(false, t0)

	logical, rose, fell := f.Sample(true, t0.Add(time.Millisecond))
	if !logical || !rose || fell {
		t.Fatalf("want immediate rise: log=%v r=%v f=%v", logical, rose, fell)
	}

	logical, rose, fell = f.Sample(false, t0.Add(5*time.Millisecond))
	if !logical || rose || fell {
		t.Fatalf("bounce ignored during clamp: log=%v r=%v f=%v", logical, rose, fell)
	}
	logical, rose, fell = f.Sample(true, t0.Add(10*time.Millisecond))
	if !logical || rose || fell {
		t.Fatalf("still clamped high: log=%v r=%v f=%v", logical, rose, fell)
	}

	logical, rose, fell = f.Sample(true, t0.Add(25*time.Millisecond))
	if !logical || rose || fell {
		t.Fatalf("after clamp same level: log=%v r=%v f=%v", logical, rose, fell)
	}

	logical, rose, fell = f.Sample(false, t0.Add(26*time.Millisecond))
	if logical || rose || !fell {
		t.Fatalf("want fall after clamp: log=%v r=%v f=%v", logical, rose, fell)
	}
}
