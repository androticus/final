package core

import (
	"fmt"
	"math"
)

type compareFamily int

const (
	compareFamilyTruthy compareFamily = iota
	compareFamilyNumeric
	compareFamilyString
)

// compareOperandFamily classifies a dyadic operand pair for equate/order/compare.
func compareOperandFamily(a, b any) (compareFamily, bool) {
	_, aIsStr := a.(String)
	_, bIsStr := b.(String)
	if aIsStr || bIsStr {
		if aIsStr && bIsStr {
			return compareFamilyString, true
		}
		return 0, false
	}
	if IsNumeric(a) && IsNumeric(b) {
		return compareFamilyNumeric, true
	}
	if IsTruthy(a) && IsTruthy(b) {
		return compareFamilyTruthy, true
	}
	return 0, false
}

// pushZeroMatchingNumericTOS pushes a typed 0 matching DS TOS for ==0 / <>0 / <0 / …
// wrappers. Non-Numeric TOS still gets Int(0); subsequent compare uses Truthy/etc.
func pushZeroMatchingNumericTOS(ds *Stack[any]) {
	switch ds.Peek().(type) {
	case Uint:
		ds.Push(Uint(0))
	case Float:
		ds.Push(Float(0))
	default:
		ds.Push(Int(0))
	}
}

// promoteNumericPair promotes a,b in place to a common numeric type.
// Order after promote for dispatch: Int, then Float, then Uint.
// Panics with a readable message on Int/Uint mix or non-numeric pair.
func promoteNumericPair(a, b *any, op OpCode) {
	switch (*a).(type) {
	case Int:
		switch (*b).(type) {
		case Int:
			return
		case Uint:
			panic(fmt.Errorf("mixed Int/Uint not allowed for dyadic op %s: a:%T, b:%T", OpCodeName(op), *a, *b))
		case Float:
			*a = Float((*a).(Int))
			return
		}
	case Float:
		switch (*b).(type) {
		case Int:
			*b = Float((*b).(Int))
			return
		case Uint:
			*b = Float((*b).(Uint))
			return
		case Float:
			return
		}
	case Uint:
		switch (*b).(type) {
		case Int:
			panic(fmt.Errorf("mixed Int/Uint not allowed for dyadic op %s: a:%T, b:%T", OpCodeName(op), *a, *b))
		case Float:
			*a = Float((*a).(Uint))
			return
		case Uint:
			return
		}
	}
	panic(fmt.Errorf("Invalid type pair for dyadic arith op %s: a:%T, b:%T", OpCodeName(op), *a, *b))
}

func doDyadicArithOp(ds *Stack[any], op OpCode) {
	b := ds.Pop()
	a := ds.Pop()
	promoteNumericPair(&a, &b, op)
	switch a.(type) {
	case Int:
		ai, bi := a.(Int), b.(Int)
		switch op {
		case OpAdd:
			ds.Push(ai + bi)
		case OpSub:
			ds.Push(ai - bi)
		case OpMul:
			ds.Push(ai * bi)
		case OpDiv:
			if bi == 0 {
				panic(fmt.Errorf("divide by zero"))
			}
			ds.Push(ai / bi)
		default:
			panic(fmt.Errorf("Invalid dyadic arith op %s", OpCodeName(op)))
		}
	case Float:
		af, bf := a.(Float), b.(Float)
		switch op {
		case OpAdd:
			ds.Push(af + bf)
		case OpSub:
			ds.Push(af - bf)
		case OpMul:
			ds.Push(af * bf)
		case OpDiv:
			ds.Push(af / bf)
		default:
			panic(fmt.Errorf("Invalid dyadic arith op %s", OpCodeName(op)))
		}
	case Uint:
		au, bu := a.(Uint), b.(Uint)
		switch op {
		case OpAdd:
			ds.Push(au + bu)
		case OpSub:
			ds.Push(au - bu)
		case OpMul:
			ds.Push(au * bu)
		case OpDiv:
			if bu == 0 {
				panic(fmt.Errorf("divide by zero"))
			}
			ds.Push(au / bu)
		default:
			panic(fmt.Errorf("Invalid dyadic arith op %s", OpCodeName(op)))
		}
	}
}

func doMonadicNumericOp(ds *Stack[any], op OpCode) {
	v := ds.Pop()
	switch x := v.(type) {
	case Int:
		switch op {
		case OpAbs:
			if x < 0 {
				ds.Push(-x)
			} else {
				ds.Push(x)
			}
		case OpNeg:
			ds.Push(-x)
		default:
			panic(fmt.Errorf("Invalid monadic numeric op %s", OpCodeName(op)))
		}
	case Float:
		switch op {
		case OpAbs:
			ds.Push(Float(math.Abs(x)))
		case OpNeg:
			ds.Push(-x)
		default:
			panic(fmt.Errorf("Invalid monadic numeric op %s", OpCodeName(op)))
		}
	case Uint:
		switch op {
		case OpAbs:
			ds.Push(x) // nop
		case OpNeg:
			panic(fmt.Errorf("neg not allowed on Uint"))
		default:
			panic(fmt.Errorf("Invalid monadic numeric op %s", OpCodeName(op)))
		}
	default:
		panic(fmt.Errorf("Invalid type for monadic numeric op %s: %T", OpCodeName(op), v))
	}
}

func doDyadicEquateOp(ds *Stack[any], op OpCode) {
	b := ds.Pop()
	a := ds.Pop()
	family, ok := compareOperandFamily(a, b)
	if !ok {
		panic(fmt.Errorf("Incompatible types for compare op %s: a:%T, b:%T", OpCodeName(op), a, b))
	}
	switch family {
	case compareFamilyString:
		as, bs := a.(String), b.(String)
		switch op {
		case OpEqual:
			ds.Push(Bool(as == bs))
		case OpNotEqual:
			ds.Push(Bool(as != bs))
		default:
			panic(fmt.Errorf("Invalid equate op %s", OpCodeName(op)))
		}
	case compareFamilyNumeric:
		promoteNumericPair(&a, &b, op)
		switch a.(type) {
		case Int:
			ai, bi := a.(Int), b.(Int)
			switch op {
			case OpEqual:
				ds.Push(Bool(ai == bi))
			case OpNotEqual:
				ds.Push(Bool(ai != bi))
			}
		case Float:
			af, bf := a.(Float), b.(Float)
			// IEEE: NaN==x false; NaN<>x true (incl. NaN<>NaN)
			switch op {
			case OpEqual:
				ds.Push(Bool(af == bf))
			case OpNotEqual:
				ds.Push(Bool(af != bf))
			}
		case Uint:
			au, bu := a.(Uint), b.(Uint)
			switch op {
			case OpEqual:
				ds.Push(Bool(au == bu))
			case OpNotEqual:
				ds.Push(Bool(au != bu))
			}
		}
	case compareFamilyTruthy:
		ab, bb := AnyToTruthy(a), AnyToTruthy(b)
		switch op {
		case OpEqual:
			ds.Push(Bool(ab == bb))
		case OpNotEqual:
			ds.Push(Bool(ab != bb))
		default:
			panic(fmt.Errorf("Invalid equate op %s", OpCodeName(op)))
		}
	}
}

func doDyadicCompareOp(ds *Stack[any], op OpCode) {
	b := ds.Pop()
	a := ds.Pop()
	family, ok := compareOperandFamily(a, b)
	if !ok {
		panic(fmt.Errorf("Incompatible types for compare op %s: a:%T, b:%T", OpCodeName(op), a, b))
	}
	switch family {
	case compareFamilyString:
		as, bs := a.(String), b.(String)
		switch op {
		case OpLessThan:
			ds.Push(Bool(as < bs))
		case OpLessThanOrEqual:
			ds.Push(Bool(as <= bs))
		case OpGreaterThan:
			ds.Push(Bool(as > bs))
		case OpGreaterThanOrEqual:
			ds.Push(Bool(as >= bs))
		default:
			panic(fmt.Errorf("Invalid order op %s", OpCodeName(op)))
		}
	case compareFamilyNumeric:
		promoteNumericPair(&a, &b, op)
		switch a.(type) {
		case Int:
			ai, bi := a.(Int), b.(Int)
			switch op {
			case OpLessThan:
				ds.Push(Bool(ai < bi))
			case OpLessThanOrEqual:
				ds.Push(Bool(ai <= bi))
			case OpGreaterThan:
				ds.Push(Bool(ai > bi))
			case OpGreaterThanOrEqual:
				ds.Push(Bool(ai >= bi))
			}
		case Float:
			af, bf := a.(Float), b.(Float)
			// IEEE unordered: all four relational ops false when NaN involved
			switch op {
			case OpLessThan:
				ds.Push(Bool(af < bf))
			case OpLessThanOrEqual:
				ds.Push(Bool(af <= bf))
			case OpGreaterThan:
				ds.Push(Bool(af > bf))
			case OpGreaterThanOrEqual:
				ds.Push(Bool(af >= bf))
			}
		case Uint:
			au, bu := a.(Uint), b.(Uint)
			switch op {
			case OpLessThan:
				ds.Push(Bool(au < bu))
			case OpLessThanOrEqual:
				ds.Push(Bool(au <= bu))
			case OpGreaterThan:
				ds.Push(Bool(au > bu))
			case OpGreaterThanOrEqual:
				ds.Push(Bool(au >= bu))
			}
		}
	case compareFamilyTruthy:
		ab, bb := AnyToTruthy(a), AnyToTruthy(b)
		// false < true
		switch op {
		case OpLessThan:
			ds.Push(Bool(!ab && bb))
		case OpLessThanOrEqual:
			ds.Push(Bool(ab == bb || (!ab && bb)))
		case OpGreaterThan:
			ds.Push(Bool(ab && !bb))
		case OpGreaterThanOrEqual:
			ds.Push(Bool(ab == bb || (ab && !bb)))
		default:
			panic(fmt.Errorf("Invalid order op %s", OpCodeName(op)))
		}
	}
}

func doCompareOp(ds *Stack[any]) {
	b := ds.Pop()
	a := ds.Pop()
	family, ok := compareOperandFamily(a, b)
	if !ok {
		panic(fmt.Errorf("Incompatible types for compare op %s: a:%T, b:%T", OpCodeName(OpCompare), a, b))
	}
	switch family {
	case compareFamilyString:
		as, bs := a.(String), b.(String)
		switch {
		case as < bs:
			ds.Push(Int(-1))
		case as > bs:
			ds.Push(Int(1))
		default:
			ds.Push(Int(0))
		}
	case compareFamilyNumeric:
		promoteNumericPair(&a, &b, OpCompare)
		switch a.(type) {
		case Int:
			ai, bi := a.(Int), b.(Int)
			switch {
			case ai < bi:
				ds.Push(Int(-1))
			case ai > bi:
				ds.Push(Int(1))
			default:
				ds.Push(Int(0))
			}
		case Float:
			af, bf := a.(Float), b.(Float)
			aNaN, bNaN := math.IsNaN(float64(af)), math.IsNaN(float64(bf))
			switch {
			case aNaN && bNaN:
				ds.Push(Int(0))
			case aNaN:
				ds.Push(Int(-1)) // NaN always less
			case bNaN:
				ds.Push(Int(1))
			case af < bf:
				ds.Push(Int(-1))
			case af > bf:
				ds.Push(Int(1))
			default:
				ds.Push(Int(0))
			}
		case Uint:
			au, bu := a.(Uint), b.(Uint)
			switch {
			case au < bu:
				ds.Push(Int(-1))
			case au > bu:
				ds.Push(Int(1))
			default:
				ds.Push(Int(0))
			}
		}
	case compareFamilyTruthy:
		ab, bb := AnyToTruthy(a), AnyToTruthy(b)
		switch {
		case !ab && bb:
			ds.Push(Int(-1))
		case ab && !bb:
			ds.Push(Int(1))
		default:
			ds.Push(Int(0))
		}
	}
}

func doMonadicCompare0Op(ds *Stack[any], op OpCode) {
	pushZeroMatchingNumericTOS(ds)
	switch op {
	case OpZeroEqual:
		doDyadicEquateOp(ds, OpEqual)
	case OpZeroNotEqual:
		doDyadicEquateOp(ds, OpNotEqual)
	case OpZeroLessThan:
		doDyadicCompareOp(ds, OpLessThan)
	case OpZeroLessThanOrEqual:
		doDyadicCompareOp(ds, OpLessThanOrEqual)
	case OpZeroGreaterThan:
		doDyadicCompareOp(ds, OpGreaterThan)
	case OpZeroGreaterThanOrEqual:
		doDyadicCompareOp(ds, OpGreaterThanOrEqual)
	default:
		panic(fmt.Errorf("Invalid zero-compare op %s", OpCodeName(op)))
	}
}

func doDyadicBooleanOp(ds *Stack[any], op OpCode) {
	b := AnyToTruthy(ds.Pop())
	a := AnyToTruthy(ds.Pop())
	switch op {
	case OpBoolAnd:
		ds.Push(Bool(a && b))
	case OpBoolOr:
		ds.Push(Bool(a || b))
	default:
		panic(fmt.Errorf("Invalid dyadic boolean op %s", OpCodeName(op)))
	}
}

func doMonadicBooleanOp(ds *Stack[any], op OpCode) {
	a := AnyToTruthy(ds.Pop())
	switch op {
	case OpBoolNot:
		ds.Push(Bool(!a))
	default:
		panic(fmt.Errorf("Invalid monadic boolean op %s", OpCodeName(op)))
	}
}

// doIntegralBitMonadic executes unary OpBitNot.
// ( a:iu -- c:iu ); result type matches operand.
func doIntegralBitMonadic(ds *Stack[any], op OpCode) {
	if op != OpBitNot {
		panic(fmt.Errorf("Invalid monadic bit op %s", OpCodeName(op)))
	}
	v := ds.Pop()
	switch x := v.(type) {
	case Int:
		ds.Push(^x)
	case Uint:
		ds.Push(^x)
	default:
		panic(fmt.Errorf("Invalid type for unary bit op %s: %T", OpCodeName(op), v))
	}
}

// doIntegralBitDyadic executes polymorphic integral bit/flag axioms.
// ( flags:iu mask:iu -- c:iu ) with mask on TOS; both Uint → Uint, else Int.
func doIntegralBitDyadic(ds *Stack[any], op OpCode) {
	mask := ds.Pop()
	flags := ds.Pop()

	fu, fIsU := flags.(Uint)
	mu, mIsU := mask.(Uint)
	if fIsU && mIsU {
		ds.Push(applyIntegralBitDyadicU(op, fu, mu))
		return
	}

	fi, fOk := integralToInt(flags)
	mi, mOk := integralToInt(mask)
	if !fOk || !mOk {
		panic(fmt.Errorf("Invalid type pair for dyadic bit op %s: flags:%T, mask:%T", OpCodeName(op), flags, mask))
	}
	ds.Push(applyIntegralBitDyadicI(op, fi, mi))
}

func integralToInt(v any) (Int, bool) {
	switch x := v.(type) {
	case Int:
		return x, true
	case Uint:
		return Int(x), true
	default:
		return 0, false
	}
}

// integralToUint32 coerces Final Int/Uint for WordFlags / SR sinks (iu).
func integralToUint32(v any) (uint32, bool) {
	switch x := v.(type) {
	case Int:
		return uint32(x), true
	case Uint:
		return uint32(x), true
	default:
		return 0, false
	}
}

func applyIntegralBitDyadicU(op OpCode, flags, mask Uint) Uint {
	switch op {
	case OpBitAnd:
		return flags & mask
	case OpBitOr, OpBitSet:
		return flags | mask
	case OpBitXor, OpBitToggle:
		return flags ^ mask
	case OpBitClear:
		return flags & ^mask
	default:
		panic(fmt.Errorf("Invalid dyadic bit op %s", OpCodeName(op)))
	}
}

func applyIntegralBitDyadicI(op OpCode, flags, mask Int) Int {
	switch op {
	case OpBitAnd:
		return flags & mask
	case OpBitOr, OpBitSet:
		return flags | mask
	case OpBitXor, OpBitToggle:
		return flags ^ mask
	case OpBitClear:
		return flags & ^mask
	default:
		panic(fmt.Errorf("Invalid dyadic bit op %s", OpCodeName(op)))
	}
}
