package core

import (
	"fmt"
	"math"
)

/*
Cast operators, eg >Int?, >>String?.
Subcode workers on OpCast; dictionary names registered in registerCastOps.

Tranche 1 — exact type (Type?): runtime type match; s( v:a -- b ).
Tranche 2 — representation (>Type?): compatible primitives; modal s( v:a -- v':t true | false ).
Tranche 3 — semantic (>>Type?): tranche 2 plus lexer string parsing where noted; same modal stack.

Panic forms (>Int, >>String, …) are Final wrappers.

See docs/001-high-level/05-operators.md and 00-stacks.md (modal ? results).
*/

type SCCastOp = int

const (
	SC_cast_nullType SCCastOp = iota
	SC_cast_boolType
	SC_cast_intType
	SC_cast_uintType
	SC_cast_floatType
	SC_cast_stringType

	SC_cast_toBool
	SC_cast_toInt
	SC_cast_toUint
	SC_cast_toFloat

	SC_cast_softNull
	SC_cast_softBool
	SC_cast_softInt
	SC_cast_softUint
	SC_cast_softFloat
	SC_cast_softString

	SCCastOpCount
)

type CastOpFn func(ds *Stack[any])

var castOpFns = [SCCastOpCount]CastOpFn{
	SC_cast_nullType:   runCastNullType,
	SC_cast_boolType:   runCastBoolType,
	SC_cast_intType:    runCastIntType,
	SC_cast_uintType:   runCastUintType,
	SC_cast_floatType:  runCastFloatType,
	SC_cast_stringType: runCastStringType,

	SC_cast_toBool:  runCastToBool,
	SC_cast_toInt:   runCastToInt,
	SC_cast_toUint:  runCastToUint,
	SC_cast_toFloat: runCastToFloat,

	SC_cast_softNull:   runCastSoftNull,
	SC_cast_softBool:   runCastSoftBool,
	SC_cast_softInt:    runCastSoftInt,
	SC_cast_softUint:   runCastSoftUint,
	SC_cast_softFloat:  runCastSoftFloat,
	SC_cast_softString: runCastSoftString,
}

func registerCastOps(d *[]*Word) {
	regSubcodeAxiom(d, "Null?", OpCast, SC_cast_nullType, WordFlagAxiom)
	regSubcodeAxiom(d, "Bool?", OpCast, SC_cast_boolType, WordFlagAxiom)
	regSubcodeAxiom(d, "Int?", OpCast, SC_cast_intType, WordFlagAxiom)
	regSubcodeAxiom(d, "Uint?", OpCast, SC_cast_uintType, WordFlagAxiom)
	regSubcodeAxiom(d, "Float?", OpCast, SC_cast_floatType, WordFlagAxiom)
	regSubcodeAxiom(d, "String?", OpCast, SC_cast_stringType, WordFlagAxiom)

	regSubcodeAxiom(d, ">Bool?", OpCast, SC_cast_toBool, WordFlagAxiom)
	regSubcodeAxiom(d, ">Int?", OpCast, SC_cast_toInt, WordFlagAxiom)
	regSubcodeAxiom(d, ">Uint?", OpCast, SC_cast_toUint, WordFlagAxiom)
	regSubcodeAxiom(d, ">Float?", OpCast, SC_cast_toFloat, WordFlagAxiom)

	regSubcodeAxiom(d, ">>Null?", OpCast, SC_cast_softNull, WordFlagAxiom)
	regSubcodeAxiom(d, ">>Bool?", OpCast, SC_cast_softBool, WordFlagAxiom)
	regSubcodeAxiom(d, ">>Int?", OpCast, SC_cast_softInt, WordFlagAxiom)
	regSubcodeAxiom(d, ">>Uint?", OpCast, SC_cast_softUint, WordFlagAxiom)
	regSubcodeAxiom(d, ">>Float?", OpCast, SC_cast_softFloat, WordFlagAxiom)
	regSubcodeAxiom(d, ">>String?", OpCast, SC_cast_softString, WordFlagAxiom)
}

func runCastTypeProbe(ds *Stack[any], op string, fn func(any) bool) {
	if ds.Len() == 0 {
		panic(fmt.Errorf("%s: empty stack", op))
	}
	v := ds.Pop()
	ds.Push(Bool(fn(v)))
}

func runCastConvertOp(ds *Stack[any], op string, fn func(any) (any, bool)) {
	if ds.Len() == 0 {
		panic(fmt.Errorf("%s: empty stack", op))
	}
	v := ds.Pop()
	if out, ok := fn(v); ok {
		ds.Push(out)
		ds.Push(Bool(true))
		return
	}
	ds.Push(Bool(false))
}

func castIsExactType[T any](v any) bool {
	_, ok := v.(T)
	return ok
}

func castIsNull(v any) bool {
	return v == nil
}

func castRepToBool(v any) (any, bool) {
	switch x := v.(type) {
	case Bool:
		return x, true
	case Int:
		return x != 0, true
	case Uint:
		return x != 0, true
	case Float:
		return x != 0, true
	default:
		return nil, false
	}
}

func castRepToInt(v any) (any, bool) {
	switch x := v.(type) {
	case Int:
		return x, true
	case Uint:
		return Int(x), true
	case Float:
		return Int(math.Trunc(x)), true
	default:
		return nil, false
	}
}

// castMagnitudeToInt accepts values that fit in Int without wrap (for >>Int?).
// Overflowing Uint fails (symmetric with >>Uint rejecting negatives); Float still
// truncates toward zero as in >Int.
func castMagnitudeToInt(v any) (any, bool) {
	switch x := v.(type) {
	case Int:
		return x, true
	case Uint:
		if x > Uint(math.MaxInt) {
			return nil, false
		}
		return Int(x), true
	case Float:
		return Int(math.Trunc(x)), true
	default:
		return nil, false
	}
}

// castRepToUint is the >Uint? representation cast: Go bit-cast wrap
// (Uint(i) / Uint(trunc(f))), including negatives.
func castRepToUint(v any) (any, bool) {
	switch x := v.(type) {
	case Uint:
		return x, true
	case Int:
		return Uint(x), true
	case Float:
		return Uint(Int(math.Trunc(x))), true
	default:
		return nil, false
	}
}

// castMagnitudeToUint accepts only non-negative Int/Uint/Float (for >>Uint?).
func castMagnitudeToUint(v any) (any, bool) {
	switch x := v.(type) {
	case Uint:
		return x, true
	case Int:
		if x < 0 {
			return nil, false
		}
		return Uint(x), true
	case Float:
		if x < 0 {
			return nil, false
		}
		return Uint(math.Trunc(x)), true
	default:
		return nil, false
	}
}

func castRepToFloat(v any) (any, bool) {
	switch x := v.(type) {
	case Float:
		return x, true
	case Int:
		return Float(x), true
	case Uint:
		return Float(x), true
	default:
		return nil, false
	}
}

func castSoftToBool(v any) (any, bool) {
	if out, ok := castRepToBool(v); ok {
		return out, true
	}
	s, ok := v.(String)
	if !ok {
		return nil, false
	}
	switch s {
	case "true":
		return Bool(true), true
	case "false":
		return Bool(false), true
	default:
		return nil, false
	}
}

func castParseIntegralString(s String) (Int, bool) {
	typ, val, err := parseNumberLiteral(string(s))
	if err != nil || !HasTokenFlag(typ, TokenInt) || HasTokenFlag(typ, TokenFloat) {
		return 0, false
	}
	return val.(Int), true
}

func castParseUintString(s String) (Uint, bool) {
	typ, val, err := parseNumberLiteral(string(s))
	if err != nil || HasTokenFlag(typ, TokenFloat) {
		return 0, false
	}
	if HasTokenFlag(typ, TokenUint) {
		return val.(Uint), true
	}
	if HasTokenFlag(typ, TokenInt) {
		i := val.(Int)
		if i < 0 {
			return 0, false
		}
		return Uint(i), true
	}
	return 0, false
}

func castParseFloatString(s String) (Float, bool) {
	typ, val, err := parseNumberLiteral(string(s))
	if err != nil {
		return 0, false
	}
	switch {
	case HasTokenFlag(typ, TokenFloat):
		return val.(Float), true
	case HasTokenFlag(typ, TokenInt):
		return Float(val.(Int)), true
	case HasTokenFlag(typ, TokenUint):
		return Float(val.(Uint)), true
	default:
		return 0, false
	}
}

func castSoftToInt(v any) (any, bool) {
	if out, ok := castMagnitudeToInt(v); ok {
		return out, true
	}
	s, ok := v.(String)
	if !ok {
		return nil, false
	}
	i, ok := castParseIntegralString(s)
	if !ok {
		return nil, false
	}
	return i, true
}

func castSoftToUint(v any) (any, bool) {
	if out, ok := castMagnitudeToUint(v); ok {
		return out, true
	}
	s, ok := v.(String)
	if !ok {
		return nil, false
	}
	u, ok := castParseUintString(s)
	if !ok {
		return nil, false
	}
	return u, true
}

func castSoftToFloat(v any) (any, bool) {
	if out, ok := castRepToFloat(v); ok {
		return out, true
	}
	s, ok := v.(String)
	if !ok {
		return nil, false
	}
	f, ok := castParseFloatString(s)
	if !ok {
		return nil, false
	}
	return f, true
}

func castSoftToNull(v any) (any, bool) {
	s, ok := v.(String)
	if !ok || s != "null" {
		return nil, false
	}
	return nil, true
}

func castSoftToString(v any) (any, bool) {
	if _, ok := v.(String); ok {
		return nil, false
	}
	switch v.(type) {
	case Bool, Int, Uint, Float:
		return String(formatNaturalString(v)), true
	default:
		if v == nil {
			return String(formatNaturalString(v)), true
		}
		return nil, false
	}
}

func runCastNullType(ds *Stack[any])   { runCastTypeProbe(ds, "Null?", castIsNull) }
func runCastBoolType(ds *Stack[any])   { runCastTypeProbe(ds, "Bool?", castIsExactType[Bool]) }
func runCastIntType(ds *Stack[any])    { runCastTypeProbe(ds, "Int?", castIsExactType[Int]) }
func runCastUintType(ds *Stack[any])   { runCastTypeProbe(ds, "Uint?", castIsExactType[Uint]) }
func runCastFloatType(ds *Stack[any])  { runCastTypeProbe(ds, "Float?", castIsExactType[Float]) }
func runCastStringType(ds *Stack[any]) { runCastTypeProbe(ds, "String?", castIsExactType[String]) }

func runCastToBool(ds *Stack[any])  { runCastConvertOp(ds, ">Bool?", castRepToBool) }
func runCastToInt(ds *Stack[any])   { runCastConvertOp(ds, ">Int?", castRepToInt) }
func runCastToUint(ds *Stack[any])  { runCastConvertOp(ds, ">Uint?", castRepToUint) }
func runCastToFloat(ds *Stack[any]) { runCastConvertOp(ds, ">Float?", castRepToFloat) }

func runCastSoftNull(ds *Stack[any])   { runCastConvertOp(ds, ">>Null?", castSoftToNull) }
func runCastSoftBool(ds *Stack[any])   { runCastConvertOp(ds, ">>Bool?", castSoftToBool) }
func runCastSoftInt(ds *Stack[any])    { runCastConvertOp(ds, ">>Int?", castSoftToInt) }
func runCastSoftUint(ds *Stack[any])   { runCastConvertOp(ds, ">>Uint?", castSoftToUint) }
func runCastSoftFloat(ds *Stack[any])  { runCastConvertOp(ds, ">>Float?", castSoftToFloat) }
func runCastSoftString(ds *Stack[any]) { runCastConvertOp(ds, ">>String?", castSoftToString) }
