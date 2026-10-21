package core

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// StringRuneIter iterates UTF-8 runes over a String (Iterator protocol; foreach via ._tryGetIter).
type StringRuneIter struct {
	s    string
	off  int
	curr Rune
}

func (it *StringRuneIter) GetNext() bool {
	r, eof := readNextRuneFromString(it.s, &it.off, "StringRuneIter")
	if eof {
		return false
	}
	it.curr = Rune(r)
	return true
}

func (it *StringRuneIter) Current() any {
	return it.curr
}

func newStringRuneIter(s String) *StringRuneIter {
	return &StringRuneIter{s: string(s)}
}

var (
	boolType   *Type // set in CreateRegisterBoolType; valid at runtime after init
	intType    *Type
	uintType   *Type
	floatType  *Type
	stringType *Type
	nullType   *Type
)

// shared by the primitives, delegates to their _toString method
// ( sb:StringBuilder -- sb)
func nativeRenderJson(f *RunFrame, ths *TypeObjPair) {
	typ := ths.Typ
	if typ == nil {
		panic("nativeRenderJson: nil type")
	}
	typ.InvokeNamed(f, f.RP.DataStack, ths, RuntimeDispatchKindInstanceMethodByName, "_toString")
	// ( -- s )
	sb := f.RP.ZStack.Peek().(*StringBuilder)
	sb.WriteString(string(f.RP.DataStack.Pop().(String)))
}

// parseJsonWireNumber converts one JSON number token to Int or Float (119 number policy).
func parseJsonWireNumber(token string) (any, bool) {
	if token == "" {
		return nil, false
	}
	if strings.ContainsAny(token, ".eE") {
		f, err := strconv.ParseFloat(token, 64)
		if err != nil {
			return nil, false
		}
		return Float(f), true
	}
	i, err := strconv.ParseInt(token, 10, 64)
	if err != nil {
		return nil, false
	}
	if i == int64(int(i)) {
		return Int(i), true
	}
	f, err := strconv.ParseFloat(token, 64)
	if err != nil {
		return nil, false
	}
	return Float(f), true
}

// ( -- n:a ok:b )
func nativeParseJsonNumber(f *RunFrame, ths *TypeObjPair) {
	val, ok := parseJsonWireNumber(string(ths.Obj.(String)))
	if ok {
		f.RP.DataStack.Push(val)
		f.RP.DataStack.Push(Bool(true))
	} else {
		f.RP.DataStack.Push(Bool(false))
	}
}

// stringJsonWire returns s as a JSON string fragment (quoted, escaped).
func stringJsonWire(s String) string {
	encoded, err := json.Marshal(string(s))
	if err != nil {
		panic(fmt.Errorf("stringJsonWire: %w", err))
	}
	return string(encoded)
}

// String JSON fragment via encoding/json (quoted, escaped).
// ( sb:StringBuilder -- sb )
func nativeRenderJsonString(f *RunFrame, ths *TypeObjPair) {
	sb := f.RP.ZStack.Peek().(*StringBuilder)
	sb.WriteString(stringJsonWire(ths.Obj.(String)))
}

func CreateRegisterAnyType() {
	t := &Type{
		name: "Any",
		getZero: func() any {
			return nil
		},
		checkCoerceType: func(value any) any {
			return value
		},
	}
	GetTypeRegistrar().RegisterCurrentType(t)
}

func CreateRegisterBoolType() {
	t := &Type{
		name: "Bool",
		getZero: func() any {
			return Bool(false)
		},
		checkCoerceType: func(value any) any {
			return AnyToTruthy(value)
		},
	}
	t.registerNativeMethod("_toString", func(f *RunFrame, ths *TypeObjPair) {
		if ths.Obj.(Bool) {
			f.RP.DataStack.Push(String("true"))
		} else {
			f.RP.DataStack.Push(String("false"))
		}
	})
	t.registerNativeMethod("_renderJson", nativeRenderJson)
	GetTypeRegistrar().RegisterCurrentType(t)
	boolType = t
}

func CreateRegisterStringType() {
	t := &Type{
		name: "String",
		getZero: func() any {
			return String("")
		},
		checkCoerceType: func(value any) any {
			s, ok := value.(String)
			if !ok {
				panic(fmt.Errorf("String.checkCoerceType: value %T is not String", value))
			}
			return s
		},
	}
	t.registerNativeMethod("_toString", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(String))
	})
	t.registerNativeMethod("_renderJson", nativeRenderJsonString)
	// ( provider -- iter true ) — provider elided from DS when present (compound ^ path)
	t.registerNativeMethod("_tryGetIter", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		if ds.Len() > 0 && ds.Peek() == any(ths.Obj) {
			ds.Pop()
		}
		ds.Push(newStringRuneIter(ths.Obj.(String)))
		ds.Push(true)
	})
	t.registerNativeMethod("parseJsonNumber", nativeParseJsonNumber)

	t.registerNativeClassMethod("isDangerousRune", func(f *RunFrame, ths *TypeObjPair) {
		rune_ := rune(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(Bool(IsDangerousRune(rune_)))
	})
	registerStringMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	stringType = t
}

func CreateRegisterIntType() {
	t := &Type{
		name: "Int",
		getZero: func() any {
			return Int(0)
		},
		checkCoerceType: func(value any) any {
			switch x := value.(type) {
			case Int:
				return x
			case Uint:
				return Int(x)
			default:
				panic(fmt.Errorf("Int.checkCoerceType: value %T is not integral", value))
			}
		},
	}
	t.registerNativeMethod("_toString", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(String(strconv.Itoa(int(ths.Obj.(Int)))))
	})
	t.registerNativeMethod("_renderJson", nativeRenderJson)
	registerIntClassConstants(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	intType = t
}

func registerIntClassConstants(t *Type) {
	t.registerClassProperty("minInt", "Int", func(_ Any) Any { return Int(math.MinInt64) }, nil)
	t.registerClassProperty("maxInt", "Int", func(_ Any) Any { return Int(math.MaxInt64) }, nil)
}

func CreateRegisterUintType() {
	t := &Type{
		name: "Uint",
		getZero: func() any {
			return Uint(0)
		},
		checkCoerceType: func(value any) any {
			switch x := value.(type) {
			case Uint:
				return x
			case Int:
				return Uint(x)
			default:
				panic(fmt.Errorf("Uint.checkCoerceType: value %T is not integral", value))
			}
		},
	}
	t.registerNativeMethod("_toString", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(String(strconv.FormatUint(uint64(ths.Obj.(Uint)), 10)))
	})
	t.registerNativeMethod("_renderJson", nativeRenderJson)
	registerUintClassConstants(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	uintType = t
}

func registerUintClassConstants(t *Type) {
	// Go has no MinUint64; unsigned minimum is 0.
	t.registerClassProperty("minUint", "Uint", func(_ Any) Any { return Uint(0) }, nil)
	t.registerClassProperty("maxUint", "Uint", func(_ Any) Any { return Uint(math.MaxUint64) }, nil)
}

func checkCoerceNumericToFloat(value any) Float {
	switch x := value.(type) {
	case Float:
		return x
	case Int:
		return Float(x)
	case Uint:
		return Float(x)
	default:
		panic(fmt.Errorf("Float.checkCoerceNumericToFloat: value %T is not numeric", value))
	}
}

// CheckCoerceNumericToFloat converts Int, Uint, or Float to Float; panics otherwise.
func CheckCoerceNumericToFloat(value any) Float {
	return checkCoerceNumericToFloat(value)
}
func CreateRegisterFloatType() {
	t := &Type{
		name: "Float",
		getZero: func() any {
			return Float(0)
		},
		checkCoerceType: func(value any) any {
			switch x := value.(type) {
			case Float:
				return x
			case Int:
				return Float(x)
			case Uint:
				return Float(x)
			default:
				panic(fmt.Errorf("Float.checkCoerceType: value %T is not numeric", value))
			}
		},
	}
	t.registerNativeMethod("_toString", func(f *RunFrame, ths *TypeObjPair) {
		// prec -1: smallest digit count for exact float64 round-trip (strconv.ParseFloat).
		s := strconv.FormatFloat(float64(ths.Obj.(Float)), 'g', -1, 64)
		f.RP.DataStack.Push(String(s))
	})
	t.registerNativeMethod("_renderJson", nativeRenderJson)
	registerFloatClassConstants(t)
	t.registerNativeClassMethod("isClose", func(f *RunFrame, ths *TypeObjPair) {
		epsilon := checkCoerceNumericToFloat(f.RP.DataStack.Pop())
		target := checkCoerceNumericToFloat(f.RP.DataStack.Pop())
		value := checkCoerceNumericToFloat(f.RP.DataStack.Pop())
		f.RP.DataStack.Push(Bool(math.Abs(target-value) <= epsilon))
	})
	t.registerNativeClassMethod("isInf", func(f *RunFrame, ths *TypeObjPair) {
		x := checkCoerceNumericToFloat(f.RP.DataStack.Pop())
		f.RP.DataStack.Push(Bool(math.IsInf(x, 0)))
	})
	t.registerNativeClassMethod("isNaN", func(f *RunFrame, ths *TypeObjPair) {
		x := checkCoerceNumericToFloat(f.RP.DataStack.Pop())
		f.RP.DataStack.Push(Bool(math.IsNaN(x)))
	})
	t.registerNativeClassMethod("_pow", func(f *RunFrame, ths *TypeObjPair) {
		y := checkCoerceNumericToFloat(f.RP.DataStack.Pop())
		x := checkCoerceNumericToFloat(f.RP.DataStack.Pop())
		f.RP.DataStack.Push(Float(math.Pow(x, y)))
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	floatType = t
}

func registerFloatClassConstants(t *Type) {
	t.registerClassProperty("maxFloat", "Float", func(_ Any) Any { return Float(math.MaxFloat64) }, nil)
	t.registerClassProperty("minNonZero", "Float", func(_ Any) Any { return Float(math.SmallestNonzeroFloat64) }, nil)
	t.registerClassProperty("+Inf", "Float", func(_ Any) Any { return Float(math.Inf(1)) }, nil)
	t.registerClassProperty("-Inf", "Float", func(_ Any) Any { return Float(math.Inf(-1)) }, nil)
	t.registerClassProperty("NaN", "Float", func(_ Any) Any { return Float(math.NaN()) }, nil)
}

func CreateRegisterNullType() {
	t := &Type{
		name:            "Null",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: nullCheckCoerceType,
	}
	t.registerNativeMethod("_toString", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(String("null"))
	})
	t.registerNativeMethod("_renderJson", nativeRenderJson)
	GetTypeRegistrar().RegisterCurrentType(t)
	nullType = t
}

func nullCheckCoerceType(value any) any {
	if value != nil {
		panic(fmt.Errorf("Null.checkCoerceType: expected nil, got %T", value))
	}
	return nil
}

func init() {
	CreateRegisterAnyType()
	CreateRegisterBoolType()
	CreateRegisterStringType()
	CreateRegisterIntType()
	CreateRegisterUintType()
	CreateRegisterFloatType()
	CreateRegisterNullType()
}
