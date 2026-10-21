package core

import (
	"math"
	"testing"
)

func TestFitOLSSimpleLine(t *testing.T) {
	// y = 10 + 2 x
	x := NewMat2f(4, 2)
	y := NewVecNf(4)
	xs := []float64{1, 2, 3, 4}
	for i, xv := range xs {
		x.Put(i, 0, 1)
		x.Put(i, 1, xv)
		y.Put(i, Float(10+2*xv))
	}
	m := FitOLS(y, x)
	if m.n != 4 || m.p != 2 {
		t.Fatalf("n,p = %d,%d", m.n, m.p)
	}
	if math.Abs(float64(m.beta.At(0))-10) > 1e-9 {
		t.Fatalf("β0 = %v, want 10", m.beta.At(0))
	}
	if math.Abs(float64(m.beta.At(1))-2) > 1e-9 {
		t.Fatalf("β1 = %v, want 2", m.beta.At(1))
	}
	if m.sse > 1e-18 {
		t.Fatalf("SSE = %g, want ~0", m.sse)
	}
	if math.Abs(m.R2()-1) > 1e-12 {
		t.Fatalf("R² = %g, want 1", m.R2())
	}
}

func TestCompareOLSMeanVsSlope(t *testing.T) {
	x0 := NewMat2f(5, 1)
	x1 := NewMat2f(5, 2)
	y := NewVecNf(5)
	for i := 0; i < 5; i++ {
		xv := float64(i + 1)
		x0.Put(i, 0, 1)
		x1.Put(i, 0, 1)
		x1.Put(i, 1, xv)
		y.Put(i, Float(3+4*xv))
	}
	c := FitOLS(y, x0)
	a := FitOLS(y, x1)
	r := CompareOLS(c, a)
	if r.df1 != 1 || r.df2 != 3 {
		t.Fatalf("dfs = %d,%d", r.df1, r.df2)
	}
	if r.pre < 0.99 {
		t.Fatalf("PRE = %g, want near 1", r.pre)
	}
	if r.f < 100 {
		t.Fatalf("F = %g, want large", r.f)
	}
}

func TestOLSModelPredict(t *testing.T) {
	x := NewMat2f(3, 2)
	y := NewVecNf(3)
	for i := 0; i < 3; i++ {
		x.Put(i, 0, 1)
		x.Put(i, 1, float64(i+1))
		y.Put(i, Float(5+i+1))
	}
	m := FitOLS(y, x)
	xNew := NewMat2f(1, 2)
	xNew.Put(0, 0, 1)
	xNew.Put(0, 1, 10)
	yHat := m.Predict(xNew)
	if yHat.Len() != 1 {
		t.Fatalf("len = %d", yHat.Len())
	}
	if math.Abs(float64(yHat.At(0))-15) > 1e-9 {
		t.Fatalf("ŷ = %v, want 15", yHat.At(0))
	}
}
