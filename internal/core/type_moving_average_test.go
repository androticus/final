package core

import (
	"math"
	"testing"
)

func TestMovingAverageSpanClamp(t *testing.T) {
	ResetTestRuntime()
	m := NewMovingAverage(0)
	if m.Span() != 1 {
		t.Fatalf("span=%d want 1", m.Span())
	}
	if m.alpha != 1.0 {
		t.Fatalf("alpha=%v want 1", m.alpha)
	}
}

func TestMovingAverageFirstAndEMA(t *testing.T) {
	ResetTestRuntime()
	// span 3 → α = 2/4 = 0.5
	m := NewMovingAverage(3)
	if math.Abs(m.alpha-0.5) > 1e-12 {
		t.Fatalf("alpha=%v want 0.5", m.alpha)
	}
	m.Add(10)
	if m.N() != 1 || m.Mean() != 10 {
		t.Fatalf("first: mean=%v n=%d", m.Mean(), m.N())
	}
	m.Add(0)
	// 0.5*0 + 0.5*10 = 5
	if m.N() != 2 || math.Abs(m.Mean()-5) > 1e-12 {
		t.Fatalf("second: mean=%v n=%d", m.Mean(), m.N())
	}
}

func TestMovingAverageResetAndGetMeanAndN(t *testing.T) {
	ResetTestRuntime()
	m := NewMovingAverage(20)
	m.Add(100)
	m.Add(200)
	mean, n := m.GetMeanAndN()
	if n != 2 || mean == 0 {
		t.Fatalf("got mean=%v n=%d", mean, n)
	}
	m.Reset()
	if m.N() != 0 || m.Mean() != 0 {
		t.Fatal("reset should clear mean and n")
	}
	if m.Span() != 20 {
		t.Fatal("reset must not change span")
	}
}
