package core

import (
	"fmt"
	"math"
	"testing"
)

func runMathMonadic(t *testing.T, subCode int, input any) Float {
	t.Helper()
	ResetTestRuntime()
	rp := NewTestRunProvider()
	rp.DataStack.Push(input)
	RunOpsWithSubCode(NewTestTask(), rp, subCode, OpMathMonadic, OpStop)
	got, ok := rp.DataStack.Pop().(Float)
	if !ok {
		t.Fatalf("result type = %T, want Float", got)
	}
	return got
}

func TestOpMathMonadic(t *testing.T) {
	cases := []struct {
		name    string
		subCode int
		fn      func(float64) float64
		inputs  []float64
	}{
		{"sin", SC_sin, math.Sin, []float64{0.31, 2.17}},
		{"cos", SC_cos, math.Cos, []float64{0.67, 1.23}},
		{"tan", SC_tan, math.Tan, []float64{0.41, 0.89}},
		{"asin", SC_asin, math.Asin, []float64{0.19, -0.47}},
		{"acos", SC_acos, math.Acos, []float64{0.33, -0.71}},
		{"atan", SC_atan, math.Atan, []float64{0.52, -0.84}},
		{"sinh", SC_sinh, math.Sinh, []float64{0.27, 1.41}},
		{"cosh", SC_cosh, math.Cosh, []float64{0.38, 0.93}},
		{"tanh", SC_tanh, math.Tanh, []float64{0.44, 1.62}},
		{"asinh", SC_asinh, math.Asinh, []float64{0.58, 2.11}},
		{"acosh", SC_acosh, math.Acosh, []float64{1.37, 4.2}},
		{"atanh", SC_atanh, math.Atanh, []float64{0.12, -0.29}},
		{"sqrt", SC_sqrt, math.Sqrt, []float64{5, 18}},
		{"cbrt", SC_cbrt, math.Cbrt, []float64{7, 64}},
		{"exp", SC_exp, math.Exp, []float64{0.63, 1.85}},
		{"log", SC_log, math.Log, []float64{2.7, 5.1}},
		{"log2", SC_log2, math.Log2, []float64{6, 32}},
		{"log10", SC_log10, math.Log10, []float64{3.5, 1000}},
		{"ceil", SC_ceil, math.Ceil, []float64{1.34, -3.56}},
		{"floor", SC_floor, math.Floor, []float64{2.91, -1.23}},
		{"round", SC_round, math.Round, []float64{3.49, -4.51}},
		{"trunc", SC_trunc, math.Trunc, []float64{6.78, -5.43}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, in := range tc.inputs {
				t.Run(fmt.Sprintf("%g", in), func(t *testing.T) {
					want := Float(tc.fn(in))
					got := runMathMonadic(t, int(tc.subCode), Float(in))
					if got != want {
						t.Fatalf("got %v, want %v", got, want)
					}
				})
			}
		})
	}

	t.Run("int_and_uint_promote_to_float", func(t *testing.T) {
		if got := runMathMonadic(t, int(SC_sqrt), Int(16)); got != Float(4) {
			t.Fatalf("sqrt(Int(16)) = %v, want Float(4)", got)
		}
		if got := runMathMonadic(t, int(SC_cbrt), Uint(27)); got != Float(3) {
			t.Fatalf("cbrt(Uint(27)) = %v, want Float(3)", got)
		}
	})
}
