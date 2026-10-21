package core

import (
	"fmt"
	"math/bits"
	"strconv"
)

const integralBitWidth = strconv.IntSize

func popIntegralInt(ds *Stack[any], op string) int {
	switch v := ds.Pop().(type) {
	case Int:
		return int(v)
	case Uint:
		return int(v)
	default:
		panic(fmt.Errorf("%s: expected Int or Uint, got %T", op, v))
	}
}

func popIntegralShiftOperands(ds *Stack[any]) (value any, by Uint) {
	_by := ds.Pop()
	value = ds.Pop()
	switch value.(type) {
	case Int, Uint:
	default:
		panic(fmt.Errorf("integral shift/rotate: invalid value type %T", value))
	}
	switch v :=_by.(type) {
	case Int:
		if v < 0 {
			panic(fmt.Errorf("integral shift/rotate count must be non-negative"))
		}
		return value, Uint(v)
	case Uint:
		return value, v
	default:
		panic(fmt.Errorf("integral shift/rotate: invalid count type %T", _by))
	}
}

func popIntegralDivOperands(ds *Stack[any]) (dividend, divisor any) {
	divisor = ds.Pop()
	dividend = ds.Pop()
	switch dividend.(type) {
	case Int, Uint:
	default:
		panic(fmt.Errorf("integral mod/rem: invalid dividend type %T", dividend))
	}
	switch divisor.(type) {
	case Int, Uint:
	default:
		panic(fmt.Errorf("integral mod/rem: invalid divisor type %T", divisor))
	}
	return dividend, divisor
}

func panicMixedIntUintModRem(dividend, divisor any) {
	panic(fmt.Errorf("mixed Int/Uint not allowed for %%/modrem: a:%T, b:%T", dividend, divisor))
}

func integralTruncRem(dividend, divisor any) any {
	du, dU := dividend.(Uint)
	su, sU := divisor.(Uint)
	if dU || sU {
		if !(dU && sU) {
			panicMixedIntUintModRem(dividend, divisor)
		}
		if su == 0 {
			panic(fmt.Errorf("divide by zero"))
		}
		return du % su
	}
	a := dividend.(Int)
	b := divisor.(Int)
	if b == 0 {
		panic(fmt.Errorf("divide by zero"))
	}
	return a % b
}

func integralTruncQuot(dividend, divisor any) any {
	du, dU := dividend.(Uint)
	su, sU := divisor.(Uint)
	if dU || sU {
		if !(dU && sU) {
			panicMixedIntUintModRem(dividend, divisor)
		}
		if su == 0 {
			panic(fmt.Errorf("divide by zero"))
		}
		return du / su
	}
	a := dividend.(Int)
	b := divisor.(Int)
	if b == 0 {
		panic(fmt.Errorf("divide by zero"))
	}
	return a / b
}

func integralShiftLeft(ds *Stack[any]) {
	value, n := popIntegralShiftOperands(ds)
	switch v := value.(type) {
	case Int:
		ds.Push(Int(v << n))
	case Uint:
		ds.Push(Uint(v << n))
	}
}

func integralShiftRight(ds *Stack[any]) {
	value, n := popIntegralShiftOperands(ds)
	switch v := value.(type) {
	case Int:
		ds.Push(Int(v >> n))
	case Uint:
		ds.Push(Uint(v >> n))
	}
}

func integralRotateLeft(ds *Stack[any]) {
	value, n := popIntegralShiftOperands(ds)
	switch v := value.(type) {
	case Int:
		ds.Push(Int(int(bits.RotateLeft(uint(v), int(n)))))
	case Uint:
		ds.Push(Uint(bits.RotateLeft(uint(v), int(n))))
	}
}

func integralRotateRight(ds *Stack[any]) {
	value, n := popIntegralShiftOperands(ds)
	switch v := value.(type) {
	case Int:
		ds.Push(Int(int(bits.RotateLeft(uint(v), -int(n)))))
	case Uint:
		ds.Push(Uint(bits.RotateLeft(uint(v), -int(n))))
	}
}

func integralRem(ds *Stack[any]) {
	dividend, divisor := popIntegralDivOperands(ds)
	ds.Push(integralTruncRem(dividend, divisor))
}

func integralModRem(ds *Stack[any]) {
	dividend, divisor := popIntegralDivOperands(ds)
	rem := integralTruncRem(dividend, divisor)
	quot := integralTruncQuot(dividend, divisor)
	ds.Push(quot)
	ds.Push(rem)
}
