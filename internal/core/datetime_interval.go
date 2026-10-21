package core

import (
	"cmp"
	"encoding/json"
	"fmt"
	"time"
)

// DateTime is Final's sealed type name for Go time.Time. Interval is a sealed
// wrapper embedding time.Duration. DS cells hold both by value.

// Interval represents a duration of time (opaque embed of Go time.Duration).
type Interval struct {
	time.Duration
}

func (iv Interval) FinalType() *Type { return intervalType }

func (iv Interval) NaturalString() string {
	return iv.Duration.String()
}

var (
	dateTimeType *Type
	intervalType *Type
)

// dateTimeTimeSentinel is the calendar anchor for DateTime.newTime (clock-only values).
func isDateTimeTimeSentinel(t time.Time) bool {
	return t.Year() == 1 && t.Month() == time.January && t.Day() == 1
}

// formatDateTimeTimeOnly formats clock fields as HH.mm.SS.msec (UTC wall time).
func formatDateTimeTimeOnly(t time.Time) string {
	ms := t.Nanosecond() / int(time.Millisecond)
	return fmt.Sprintf("%02d.%02d.%02d.%03d", t.Hour(), t.Minute(), t.Second(), ms)
}

// formatDateTimeISO formats calendar + clock as ISO-8601 without a timezone suffix.
func formatDateTimeISO(t time.Time) string {
	if t.Nanosecond() == 0 {
		return t.Format("2006-01-02T15:04:05")
	}
	if t.Nanosecond()%int(time.Millisecond) == 0 {
		return fmt.Sprintf("%s.%03d", t.Format("2006-01-02T15:04:05"), t.Nanosecond()/int(time.Millisecond))
	}
	return t.Format("2006-01-02T15:04:05.999999999")
}

func dateTimeToString(t time.Time) String {
	if isDateTimeTimeSentinel(t) {
		return String(formatDateTimeTimeOnly(t))
	}
	return String(formatDateTimeISO(t))
}

func listIntAt(parts *List[any], i int) int {
	v, ok := parts.AtAsAny(i).(Int)
	if !ok {
		panic(fmt.Errorf("DateTime factory: list[%d] want Int, got %T", i, parts.AtAsAny(i)))
	}
	return int(v)
}

func listLenOneOf(parts *List[any], allowed ...int) {
	n := parts.Len()
	for _, want := range allowed {
		if n == want {
			return
		}
	}
	panic(fmt.Errorf("DateTime factory: list len %d, want one of %v", n, allowed))
}

func optionalListMillisAsNsec(parts *List[any], idx int) int {
	if parts.Len() <= idx {
		return 0
	}
	return int(listIntAt(parts, idx)) * int(time.Millisecond)
}

func popIntList(f *RunFrame) *List[any] {
	v := f.RP.DataStack.Pop()
	parts, ok := v.(*List[any])
	if !ok {
		panic(fmt.Errorf("DateTime factory: want List, got %T", v))
	}
	return parts
}

// ( sb:StringBuilder -- sb )
func nativeRenderJsonDateTime(f *RunFrame, ths *TypeObjPair) {
	tm := ths.Obj.(time.Time)
	encoded, err := json.Marshal(tm.UTC().Format(time.RFC3339))
	if err != nil {
		panic(fmt.Errorf("DateTime._renderJson: %w", err))
	}
	sb := f.RP.ZStack.Peek().(*StringBuilder)
	sb.WriteString(string(encoded))
}

func registerDateTimeMembers(t *Type) {
	t.registerProperty("year", "Int",
		func(inst Any) Any { return Int(inst.(time.Time).Year()) },
		nil,
	)
	t.registerProperty("month", "Int",
		func(inst Any) Any { return Int(inst.(time.Time).Month()) },
		nil,
	)
	t.registerProperty("day", "Int",
		func(inst Any) Any { return Int(inst.(time.Time).Day()) },
		nil,
	)
	t.registerProperty("hour", "Int",
		func(inst Any) Any { return Int(inst.(time.Time).Hour()) },
		nil,
	)
	t.registerProperty("minute", "Int",
		func(inst Any) Any { return Int(inst.(time.Time).Minute()) },
		nil,
	)
	t.registerProperty("second", "Int",
		func(inst Any) Any { return Int(inst.(time.Time).Second()) },
		nil,
	)
	t.registerProperty("nanosecond", "Int",
		func(inst Any) Any { return Int(inst.(time.Time).Nanosecond()) },
		nil,
	)
	t.registerProperty("isZero", "Bool",
		func(inst Any) Any { return Bool(inst.(time.Time).IsZero()) },
		nil,
	)
	t.registerNativeMethod("since", func(f *RunFrame, ths *TypeObjPair) {
		inst := f.RP.DataStack.Pop().(time.Time)
		f.RP.DataStack.Push(Interval{time.Since(inst)})
	})
	t.registerNativeMethod("_toString", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(dateTimeToString(ths.Obj.(time.Time)))
	})
	t.registerNativeMethod("_renderJson", nativeRenderJsonDateTime)

	t.registerNativeClassMethod("now", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(time.Now())
	})

	// ( parts:List -- dt:DateTime )
	t.registerNativeClassMethod("newDate", func(f *RunFrame, _ *TypeObjPair) {
		parts := popIntList(f)
		listLenOneOf(parts, 3)
		year := listIntAt(parts, 0)
		month := listIntAt(parts, 1)
		day := listIntAt(parts, 2)
		f.RP.DataStack.Push(time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC))
	})
	// ( parts:List -- dt:DateTime ) — 3 or 4 elements; optional ms at [3]
	t.registerNativeClassMethod("newTime", func(f *RunFrame, _ *TypeObjPair) {
		parts := popIntList(f)
		listLenOneOf(parts, 3, 4)
		hour := listIntAt(parts, 0)
		min := listIntAt(parts, 1)
		sec := listIntAt(parts, 2)
		nsec := optionalListMillisAsNsec(parts, 3)
		f.RP.DataStack.Push(time.Date(1, 1, 1, hour, min, sec, nsec, time.UTC))
	})
	// ( parts:List -- dt:DateTime ) — 6 or 7 elements; optional ms at [6]
	t.registerNativeClassMethod("newDateTime", func(f *RunFrame, _ *TypeObjPair) {
		parts := popIntList(f)
		listLenOneOf(parts, 6, 7)
		year := listIntAt(parts, 0)
		month := listIntAt(parts, 1)
		day := listIntAt(parts, 2)
		hour := listIntAt(parts, 3)
		min := listIntAt(parts, 4)
		sec := listIntAt(parts, 5)
		nsec := optionalListMillisAsNsec(parts, 6)
		f.RP.DataStack.Push(time.Date(year, time.Month(month), day, hour, min, sec, nsec, time.UTC))
	})

	registerDateTimeFormatConstants(t)
}

func registerIntervalClassConstants(t *Type) {
	for _, c := range []struct {
		name string
		d    time.Duration
	}{
		{"zero", time.Duration(0)},
		{"nanosecond", time.Nanosecond},
		{"microsecond", time.Microsecond},
		{"millisecond", time.Millisecond},
		{"second", time.Second},
		{"minute", time.Minute},
		{"hour", time.Hour},
	} {
		d := c.d
		name := c.name
		t.registerClassProperty(name, "Interval", func(_ Any) Any {
			return Interval{d}
		}, nil)
	}
}

func registerIntervalUnitProperties(t *Type) {
	t.registerProperty("nanoseconds", "Int",
		func(inst Any) Any { return Int(inst.(Interval).Nanoseconds()) },
		nil,
	)
	t.registerProperty("microseconds", "Int",
		func(inst Any) Any { return Int(inst.(Interval).Microseconds()) },
		nil,
	)
	t.registerProperty("milliseconds", "Int",
		func(inst Any) Any { return Int(inst.(Interval).Milliseconds()) },
		nil,
	)
	t.registerProperty("seconds", "Float",
		func(inst Any) Any { return Float(inst.(Interval).Seconds()) },
		nil,
	)
	t.registerProperty("minutes", "Float",
		func(inst Any) Any { return Float(inst.(Interval).Minutes()) },
		nil,
	)
	t.registerProperty("hours", "Float",
		func(inst Any) Any { return Float(inst.(Interval).Hours()) },
		nil,
	)
	// Go time.Duration has no Days(); Hours()/24 matches the Hours/Minutes/Seconds float64 shape.
	t.registerProperty("days", "Float",
		func(inst Any) Any { return Float(inst.(Interval).Hours() / 24) },
		nil,
	)
}

func registerIntervalMembers(t *Type) {
	registerIntervalClassConstants(t)
	registerIntervalUnitProperties(t)
	registerIntervalFormatConstants(t)
	t.registerNativeMethod("_toString", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(String(ths.Obj.(Interval).Duration.String()))
	})
	t.registerNativeMethod("_renderJson", nativeRenderJson)

	// ( parts:List -- iv:Interval ) — 3 or 4 elements; optional ms at [3]
	t.registerNativeClassMethod("newInterval", func(f *RunFrame, _ *TypeObjPair) {
		parts := popIntList(f)
		listLenOneOf(parts, 3, 4)
		hour := listIntAt(parts, 0)
		min := listIntAt(parts, 1)
		sec := listIntAt(parts, 2)
		nsec := optionalListMillisAsNsec(parts, 3)
		iv := (time.Duration(hour) * time.Hour) + (time.Duration(min) * time.Minute) + (time.Duration(sec) * time.Second) + (time.Duration(nsec) * time.Nanosecond)
		f.RP.DataStack.Push(Interval{iv})
	})
}

func CreateRegisterDateTimeType() {
	t := &Type{
		name: "DateTime",
		getZero: func() any {
			return time.Time{}
		},
		checkCoerceType: func(value any) any {
			tm, ok := value.(time.Time)
			if !ok {
				panic(fmt.Errorf("DateTime.checkCoerceType: value %T is not time.Time", value))
			}
			return tm
		},
	}
	registerDateTimeMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	dateTimeType = t
}

func CreateRegisterIntervalType() {
	t := &Type{
		name: "Interval",
		getZero: func() any {
			return Interval{}
		},
		checkCoerceType: func(value any) any {
			iv, ok := value.(Interval)
			if !ok {
				panic(fmt.Errorf("Interval.checkCoerceType: value %T is not Interval", value))
			}
			return iv
		},
	}
	registerIntervalMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	intervalType = t
}

func init() {
	CreateRegisterIntervalType()
	CreateRegisterDateTimeType()
}

// --- DateTime/Interval dyadic ops (OpDtDyadic subcodes) ---
// Words register as dt::+, dt::-, … (quasi-module prefix until module renaming exists).

type SCDtDyadic = int

const (
	SC_dt_add SCDtDyadic = iota
	SC_dt_sub
	SC_dt_mul
	SC_dt_compare
	SC_dt_lt
	SC_dt_lte
	SC_dt_gt
	SC_dt_gte
	SC_dt_eq
	SC_dt_ne
	SC_dt_round
	SC_dt_div
	SC_dt_fmt

	SCDtDyadicCount
)

type DtDyadicFn func(ds *Stack[any])

var dtDyadicFns = [SCDtDyadicCount]DtDyadicFn{
	SC_dt_add:     runDtDyadicAdd,
	SC_dt_sub:     runDtDyadicSub,
	SC_dt_mul:     runDtDyadicMul,
	SC_dt_compare: runDtDyadicCompare,
	SC_dt_lt:      runDtDyadicLT,
	SC_dt_lte:     runDtDyadicLTE,
	SC_dt_gt:      runDtDyadicGT,
	SC_dt_gte:     runDtDyadicGTE,
	SC_dt_eq:      runDtDyadicEQ,
	SC_dt_ne:      runDtDyadicNE,
	SC_dt_round:   runDtDyadicRound,
	SC_dt_div:     runDtDyadicDiv,
	SC_dt_fmt:     runDtDyadicFmt,
}

func registerDtDyadics(d *[]*Word) {
	regSubcodeAxiom(d, "dt::+", OpDtDyadic, SC_dt_add, WordFlagAxiom)
	regSubcodeAxiom(d, "dt::-", OpDtDyadic, SC_dt_sub, WordFlagAxiom)
	regSubcodeAxiom(d, "dt::*", OpDtDyadic, SC_dt_mul, WordFlagAxiom)
	regSubcodeAxiom(d, "dt::/", OpDtDyadic, SC_dt_div, WordFlagAxiom)
	regSubcodeAxiom(d, "dt::compare", OpDtDyadic, SC_dt_compare, WordFlagAxiom)
	regSubcodeAxiom(d, "dt::<", OpDtDyadic, SC_dt_lt, WordFlagAxiom)
	regSubcodeAxiom(d, "dt::<=", OpDtDyadic, SC_dt_lte, WordFlagAxiom)
	regSubcodeAxiom(d, "dt::>", OpDtDyadic, SC_dt_gt, WordFlagAxiom)
	regSubcodeAxiom(d, "dt::>=", OpDtDyadic, SC_dt_gte, WordFlagAxiom)
	regSubcodeAxiom(d, "dt::==", OpDtDyadic, SC_dt_eq, WordFlagAxiom)
	regSubcodeAxiom(d, "dt::<>", OpDtDyadic, SC_dt_ne, WordFlagAxiom)
	regSubcodeAxiom(d, "dt::round", OpDtDyadic, SC_dt_round, WordFlagAxiom)
	regSubcodeAxiom(d, "dt::fmt", OpDtDyadic, SC_dt_fmt, WordFlagAxiom)
}

func panicDtDyadic(op string, op1, op2 any) {
	panic(fmt.Errorf("dt::%s: invalid operands %T, %T", op, op1, op2))
}

func dtDyadicCompareInt(op1, op2 any) int {
	switch t1 := op1.(type) {
	case time.Time:
		t2, ok := op2.(time.Time)
		if !ok {
			panicDtDyadic("compare", op1, op2)
		}
		return t1.Compare(t2)
	case Interval:
		i2, ok := op2.(Interval)
		if !ok {
			panicDtDyadic("compare", op1, op2)
		}
		return cmp.Compare(t1.Duration, i2.Duration)
	default:
		panicDtDyadic("compare", op1, op2)
		return 0
	}
}

func runDtDyadicAdd(ds *Stack[any]) {
	op2 := ds.Pop()
	op1 := ds.Pop()
	switch t1 := op1.(type) {
	case time.Time:
		i2, ok := op2.(Interval)
		if !ok {
			panicDtDyadic("+", op1, op2)
		}
		ds.Push(t1.Add(i2.Duration))
	case Interval:
		switch t2 := op2.(type) {
		case time.Time:
			ds.Push(t2.Add(t1.Duration))
		case Interval:
			ds.Push(Interval{t1.Duration + t2.Duration})
		default:
			panicDtDyadic("+", op1, op2)
		}
	default:
		panicDtDyadic("+", op1, op2)
	}
}

func runDtDyadicSub(ds *Stack[any]) {
	op2 := ds.Pop()
	op1 := ds.Pop()
	switch t1 := op1.(type) {
	case time.Time:
		switch t2 := op2.(type) {
		case Interval:
			ds.Push(t1.Add(-t2.Duration))
		case time.Time:
			ds.Push(Interval{t1.Sub(t2)})
		default:
			panicDtDyadic("-", op1, op2)
		}
	case Interval:
		i2, ok := op2.(Interval)
		if !ok {
			panicDtDyadic("-", op1, op2)
		}
		ds.Push(Interval{t1.Duration - i2.Duration})
	default:
		panicDtDyadic("-", op1, op2)
	}
}

// runDtDyadicMul: Int × Interval (either order) → Interval.
func runDtDyadicMul(ds *Stack[any]) {
	op2 := ds.Pop()
	op1 := ds.Pop()
	switch t1 := op1.(type) {
	case Int:
		iv, ok := op2.(Interval)
		if !ok {
			panicDtDyadic("*", op1, op2)
		}
		ds.Push(Interval{iv.Duration * time.Duration(t1)})
	case Interval:
		n, ok := op2.(Int)
		if !ok {
			panicDtDyadic("*", op1, op2)
		}
		ds.Push(Interval{t1.Duration * time.Duration(n)})
	default:
		panicDtDyadic("*", op1, op2)
	}
}

// runDtDyadicDiv: Interval / Int → Interval. Panics if n <= 0.
func runDtDyadicDiv(ds *Stack[any]) {
	op2 := ds.Pop()
	op1 := ds.Pop()
	iv, ok := op1.(Interval)
	if !ok {
		panicDtDyadic("/", op1, op2)
	}
	n, ok := op2.(Int)
	if !ok {
		panicDtDyadic("/", op1, op2)
	}
	if n <= 0 {
		panic(fmt.Errorf("dt::/: divisor must be > 0, got %d", n))
	}
	ds.Push(Interval{iv.Duration / time.Duration(n)})
}

func runDtDyadicCompare(ds *Stack[any]) {
	op2 := ds.Pop()
	op1 := ds.Pop()
	ds.Push(Int(dtDyadicCompareInt(op1, op2)))
}

func runDtDyadicLT(ds *Stack[any]) {
	op2 := ds.Pop()
	op1 := ds.Pop()
	ds.Push(Bool(dtDyadicCompareInt(op1, op2) < 0))
}

func runDtDyadicLTE(ds *Stack[any]) {
	op2 := ds.Pop()
	op1 := ds.Pop()
	ds.Push(Bool(dtDyadicCompareInt(op1, op2) <= 0))
}

func runDtDyadicGT(ds *Stack[any]) {
	op2 := ds.Pop()
	op1 := ds.Pop()
	ds.Push(Bool(dtDyadicCompareInt(op1, op2) > 0))
}

func runDtDyadicGTE(ds *Stack[any]) {
	op2 := ds.Pop()
	op1 := ds.Pop()
	ds.Push(Bool(dtDyadicCompareInt(op1, op2) >= 0))
}

func runDtDyadicEQ(ds *Stack[any]) {
	op2 := ds.Pop()
	op1 := ds.Pop()
	ds.Push(Bool(dtDyadicCompareInt(op1, op2) == 0))
}

func runDtDyadicNE(ds *Stack[any]) {
	op2 := ds.Pop()
	op1 := ds.Pop()
	ds.Push(Bool(dtDyadicCompareInt(op1, op2) != 0))
}

func runDtDyadicRound(ds *Stack[any]) {
	op2 := ds.Pop().(Interval)
	op1 := ds.Pop().(Interval)
	ds.Push(Interval{op1.Duration.Round(op2.Duration)})
}
