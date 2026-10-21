package core

import (
	"math"
	"testing"
)

func invokeClassConst(t *testing.T, typ *Type, name string) any {
	t.Helper()
	pm := typ.GetClassPropertyMetadata(name)
	if pm == nil {
		t.Fatalf("missing %s.%s class property", typ.name, name)
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, typ)
	pm.invokeGet(f, ths)
	return rp.DataStack.Pop()
}

func TestIntClassConstants(t *testing.T) {
	ResetTestRuntime()
	if got := invokeClassConst(t, intType, "minInt"); got != Int(math.MinInt64) {
		t.Fatalf("Int.minInt = %v, want %v", got, math.MinInt64)
	}
	if got := invokeClassConst(t, intType, "maxInt"); got != Int(math.MaxInt64) {
		t.Fatalf("Int.maxInt = %v, want %v", got, math.MaxInt64)
	}
}

func TestUintClassConstants(t *testing.T) {
	ResetTestRuntime()
	if got := invokeClassConst(t, uintType, "minUint"); got != Uint(0) {
		t.Fatalf("Uint.minUint = %v, want 0", got)
	}
	if got := invokeClassConst(t, uintType, "maxUint"); got != Uint(math.MaxUint64) {
		t.Fatalf("Uint.maxUint = %v, want %v", got, uint64(math.MaxUint64))
	}
}

func TestFloatClassConstants(t *testing.T) {
	ResetTestRuntime()
	if got := invokeClassConst(t, floatType, "maxFloat"); got != Float(math.MaxFloat64) {
		t.Fatalf("Float.maxFloat = %v, want %v", got, math.MaxFloat64)
	}
	if got := invokeClassConst(t, floatType, "minNonZero"); got != Float(math.SmallestNonzeroFloat64) {
		t.Fatalf("Float.minNonZero = %v, want %v", got, math.SmallestNonzeroFloat64)
	}
	got := invokeClassConst(t, floatType, "NaN")
	f, ok := got.(Float)
	if !ok || !math.IsNaN(float64(f)) {
		t.Fatalf("Float.NaN = %v (%T), want NaN Float", got, got)
	}
}

func invokeFloatClassMethod(t *testing.T, name string, pushArgs func(*Stack[any])) any {
	t.Helper()
	mm := floatType.ClassMethodByName(name)
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, floatType)
	if pushArgs != nil {
		pushArgs(rp.DataStack)
	}
	mm.invokeCall(f, ths)
	if rp.DataStack.Len() != 1 {
		t.Fatalf("Float.%s: DS len = %d, want 1", name, rp.DataStack.Len())
	}
	return rp.DataStack.Pop()
}

func TestFloatClassMethodIsNaN(t *testing.T) {
	ResetTestRuntime()
	cases := []struct {
		name string
		push func(*Stack[any])
		want Bool
	}{
		{"nan", func(ds *Stack[any]) { ds.Push(Float(math.NaN())) }, true},
		{"finite_float", func(ds *Stack[any]) { ds.Push(Float(1.5)) }, false},
		{"int", func(ds *Stack[any]) { ds.Push(Int(1)) }, false},
		{"uint", func(ds *Stack[any]) { ds.Push(Uint(1)) }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := invokeFloatClassMethod(t, "isNaN", tc.push)
			if got != tc.want {
				t.Fatalf("Float.isNaN = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFloatClassMethodIsInf(t *testing.T) {
	ResetTestRuntime()
	cases := []struct {
		name string
		push func(*Stack[any])
		want Bool
	}{
		{"pos_inf", func(ds *Stack[any]) { ds.Push(Float(math.Inf(1))) }, true},
		{"neg_inf", func(ds *Stack[any]) { ds.Push(Float(math.Inf(-1))) }, true},
		{"finite_float", func(ds *Stack[any]) { ds.Push(Float(1.5)) }, false},
		{"nan", func(ds *Stack[any]) { ds.Push(Float(math.NaN())) }, false},
		{"int", func(ds *Stack[any]) { ds.Push(Int(1)) }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := invokeFloatClassMethod(t, "isInf", tc.push)
			if got != tc.want {
				t.Fatalf("Float.isInf = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFloatClassMethodIsClose(t *testing.T) {
	ResetTestRuntime()
	cases := []struct {
		name string
		push func(*Stack[any])
		want Bool
	}{
		{"close_floats", func(ds *Stack[any]) {
			ds.Push(Float(1.0))
			ds.Push(Float(1.0001))
			ds.Push(Float(0.001))
		}, true},
		{"not_close_floats", func(ds *Stack[any]) {
			ds.Push(Float(1.0))
			ds.Push(Float(1.1))
			ds.Push(Float(0.01))
		}, false},
		{"mixed_numeric", func(ds *Stack[any]) {
			ds.Push(Int(10))
			ds.Push(Float(10.0))
			ds.Push(Uint(0))
		}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := invokeFloatClassMethod(t, "isClose", tc.push)
			if got != tc.want {
				t.Fatalf("Float.isClose = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFloatClassMethodPow(t *testing.T) {
	ResetTestRuntime()
	cases := []struct {
		name string
		push func(*Stack[any])
		want Float
	}{
		{"float_float", func(ds *Stack[any]) {
			ds.Push(Float(2))
			ds.Push(Float(3))
		}, Float(8)},
		{"int_int", func(ds *Stack[any]) {
			ds.Push(Int(2))
			ds.Push(Int(10))
		}, Float(1024)},
		{"mixed", func(ds *Stack[any]) {
			ds.Push(Uint(3))
			ds.Push(Float(2))
		}, Float(9)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := invokeFloatClassMethod(t, "_pow", tc.push)
			if got != tc.want {
				t.Fatalf("Float._pow = %v (%T), want %v", got, got, tc.want)
			}
		})
	}
}

func TestMathClassConstants(t *testing.T) {
	ResetTestRuntime()
	if mathType == nil {
		t.Fatal("Math type not registered")
	}
	cases := []struct {
		name string
		want float64
	}{
		{"e", math.E},
		{"pi", math.Pi},
		{"ln2", math.Ln2},
		{"ln10", math.Ln10},
	}
	for _, tc := range cases {
		got, ok := invokeClassConst(t, mathType, tc.name).(Float)
		if !ok {
			t.Fatalf("Math.%s type = %T, want Float", tc.name, got)
		}
		if float64(got) != tc.want {
			t.Fatalf("Math.%s = %v, want %v", tc.name, got, tc.want)
		}
	}
}
