package core

import (
	"fmt"
	"math"
)

// Math provides common math operations like cos and sin.
// These are registered as subcode words.
// They share 2 opcodes: OpMathDyadic and OpMathMonadic.
// The axioms accept any numeric type, and return Float.
//
// Math is also a class-only catalog type exposing Go math constants (e, pi, …).

var mathType *Type

func CreateRegisterMathType() {
	t := &Type{
		name:    "Math",
		getZero: ReferenceTypeZeroFn,
		checkCoerceType: func(value any) any {
			panic(fmt.Errorf("Math.checkCoerceType: Math is a class-only type (got %T)", value))
		},
	}
	registerMathClassConstants(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	mathType = t
}

func registerMathClassConstants(t *Type) {
	t.registerClassProperty("e", "Float", func(_ Any) Any { return Float(math.E) }, nil)
	t.registerClassProperty("pi", "Float", func(_ Any) Any { return Float(math.Pi) }, nil)
	t.registerClassProperty("ln2", "Float", func(_ Any) Any { return Float(math.Ln2) }, nil)
	t.registerClassProperty("ln10", "Float", func(_ Any) Any { return Float(math.Ln10) }, nil)
}

func init() {
	CreateRegisterMathType()
}

type SCMathMonadicF = int

const (
	SC_sin SCMathMonadicF = iota
	SC_cos
	SC_tan
	SC_asin
	SC_acos
	SC_atan
	SC_sinh
	SC_cosh
	SC_tanh
	SC_asinh
	SC_acosh
	SC_atanh
	SC_sqrt
	SC_cbrt
	SC_exp
	SC_log
	SC_log2
	SC_log10
	SC_ceil
	SC_floor
	SC_round
	SC_trunc

	SCMathMonadicFCount
)

type MathMonadicFn func(x float64) float64

var mathMonadicFns = [SCMathMonadicFCount]MathMonadicFn{
	SC_sin:   math.Sin,
	SC_cos:   math.Cos,
	SC_tan:   math.Tan,
	SC_asin:  math.Asin,
	SC_acos:  math.Acos,
	SC_atan:  math.Atan,
	SC_sinh:  math.Sinh,
	SC_cosh:  math.Cosh,
	SC_tanh:  math.Tanh,
	SC_asinh: math.Asinh,
	SC_acosh: math.Acosh,
	SC_atanh: math.Atanh,
	SC_sqrt:  math.Sqrt,
	SC_cbrt:  math.Cbrt,
	SC_exp:   math.Exp,
	SC_log:   math.Log,
	SC_log2:  math.Log2,
	SC_log10: math.Log10,
	SC_ceil:  math.Ceil,
	SC_floor: math.Floor,
	SC_round: math.Round,
	SC_trunc: math.Trunc,
}

func registerMathMonadics(d *[]*Word) {
	regSubcodeAxiom(d, "sin", OpMathMonadic, SC_sin, WordFlagAxiom)
	regSubcodeAxiom(d, "cos", OpMathMonadic, SC_cos, WordFlagAxiom)
	regSubcodeAxiom(d, "tan", OpMathMonadic, SC_tan, WordFlagAxiom)
	regSubcodeAxiom(d, "asin", OpMathMonadic, SC_asin, WordFlagAxiom)
	regSubcodeAxiom(d, "acos", OpMathMonadic, SC_acos, WordFlagAxiom)
	regSubcodeAxiom(d, "atan", OpMathMonadic, SC_atan, WordFlagAxiom)
	regSubcodeAxiom(d, "sinh", OpMathMonadic, SC_sinh, WordFlagAxiom)
	regSubcodeAxiom(d, "cosh", OpMathMonadic, SC_cosh, WordFlagAxiom)
	regSubcodeAxiom(d, "tanh", OpMathMonadic, SC_tanh, WordFlagAxiom)
	regSubcodeAxiom(d, "asinh", OpMathMonadic, SC_asinh, WordFlagAxiom)
	regSubcodeAxiom(d, "acosh", OpMathMonadic, SC_acosh, WordFlagAxiom)
	regSubcodeAxiom(d, "atanh", OpMathMonadic, SC_atanh, WordFlagAxiom)
	regSubcodeAxiom(d, "sqrt", OpMathMonadic, SC_sqrt, WordFlagAxiom)
	regSubcodeAxiom(d, "cbrt", OpMathMonadic, SC_cbrt, WordFlagAxiom)
	regSubcodeAxiom(d, "exp", OpMathMonadic, SC_exp, WordFlagAxiom)
	regSubcodeAxiom(d, "log", OpMathMonadic, SC_log, WordFlagAxiom)
	regSubcodeAxiom(d, "log2", OpMathMonadic, SC_log2, WordFlagAxiom)
	regSubcodeAxiom(d, "log10", OpMathMonadic, SC_log10, WordFlagAxiom)
	regSubcodeAxiom(d, "ceil", OpMathMonadic, SC_ceil, WordFlagAxiom)
	regSubcodeAxiom(d, "floor", OpMathMonadic, SC_floor, WordFlagAxiom)
	regSubcodeAxiom(d, "round", OpMathMonadic, SC_round, WordFlagAxiom)
	regSubcodeAxiom(d, "trunc", OpMathMonadic, SC_trunc, WordFlagAxiom)
}
