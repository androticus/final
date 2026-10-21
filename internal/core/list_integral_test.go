package core

import (
	"reflect"
	"testing"
)

// Runtime DS cells for integral lists (wide int / wide uint on this platform).
func boxInt(v int) any   { return Int(v) }
func boxUint(v uint) any { return Uint(v) }

// goNarrow applies Go's conversion from a runtime integral cell to storage type T.
func goNarrow[T any](v any) T {
	var zero T
	rt := reflect.TypeOf(zero)
	switch x := v.(type) {
	case Int:
		return reflect.ValueOf(int(x)).Convert(rt).Interface().(T)
	case Uint:
		return reflect.ValueOf(uint(x)).Convert(rt).Interface().(T)
	default:
		panic("goNarrow: not an integral runtime cell")
	}
}

// goWidenUint8 is the wide unsigned cell expected when reading uint8 from a list.
func goWidenUint8(v uint8) uint {
	return uint(v)
}

// goWidenInt32 is the wide signed cell expected when reading int32 from a list.
func goWidenInt32(v int32) int {
	return int(v)
}

func atUint(t *testing.T, l *ByteList, i int) uint {
	t.Helper()
	return uint(l.At(i))
}

func atInt(t *testing.T, l *Int32List, i int) int {
	t.Helper()
	return int(l.At(i))
}

// ByteList exemplar: unsigned narrow storage, runtime Int/Uint operands.

func TestByteList_integralCoercion(t *testing.T) {
	narrowCases := []struct {
		name string
		in   any
	}{
		{"uint_truncates", boxUint(300)},
		{"int_truncates_bit_pattern", boxInt(-1)},
		{"int_in_range", boxInt(42)},
		{"uint_in_range", boxUint(42)},
	}
	for _, tc := range narrowCases {
		t.Run("narrow_"+tc.name, func(t *testing.T) {
			l := NewByteList()
			l.AppendFromAny(tc.in)
			want := goWidenUint8(goNarrow[uint8](tc.in))
			if got := atUint(t, l, 0); got != want {
				t.Fatalf("round-trip = %d, want %d (Go uint8 conversion)", got, want)
			}
			if got := l.AtAsAny(0).(Int); got != Int(want) {
				t.Fatalf("AtAsAny = %v, want Int(%d)", got, want)
			}
		})
	}

	t.Run("widen_stored_unsigned", func(t *testing.T) {
		l := NewByteList()
		l.Append(0)
		l.Append(255)
		if atUint(t, l, 0) != 0 || atUint(t, l, 1) != 255 {
			t.Fatalf("widened read = %d, %d", atUint(t, l, 0), atUint(t, l, 1))
		}
		if l.AtAsAny(0) != Int(0) || l.AtAsAny(1) != Int(255) {
			t.Fatalf("AtAsAny = %v, %v, want Int 0 and Int 255", l.AtAsAny(0), l.AtAsAny(1))
		}
	})

	t.Run("put_and_insert_from_runtime", func(t *testing.T) {
		l := NewByteList()
		l.AppendFromAny(boxInt(10))
		l.PutFromAny(0, boxUint(20))
		l.InsertAtFromAny(1, boxInt(30))
		want := []uint{
			goWidenUint8(goNarrow[uint8](boxUint(20))),
			goWidenUint8(goNarrow[uint8](boxInt(30))),
		}
		for i, w := range want {
			if got := atUint(t, l, i); got != w {
				t.Fatalf("[%d] = %d, want %d", i, got, w)
			}
		}
	})

	t.Run("rejects_non_integral", func(t *testing.T) {
		for _, v := range []any{nil, String("1"), Bool(true), Float(1.0), NewAnyList()} {
			expectPanic(t, func() { NewByteList().AppendFromAny(v) })
		}
	})
}

// Int32List exemplar: signed narrow storage, runtime Int/Uint operands.

func TestInt32List_integralCoercion(t *testing.T) {
	narrowCases := []struct {
		name string
		in   any
	}{
		{"uint_truncates_bit_pattern", boxUint(0xffffffff)},
		{"int_truncates", boxInt(0x80000000)},
		{"int_in_range", boxInt(-1)},
		{"uint_small", boxUint(200)},
	}
	for _, tc := range narrowCases {
		t.Run("narrow_"+tc.name, func(t *testing.T) {
			l := NewInt32List()
			l.AppendFromAny(tc.in)
			want := goWidenInt32(goNarrow[int32](tc.in))
			if got := atInt(t, l, 0); got != want {
				t.Fatalf("round-trip = %d, want %d (Go int32 conversion)", got, want)
			}
		})
	}

	t.Run("widen_stored_signed", func(t *testing.T) {
		l := NewInt32List()
		l.Append(-1)
		l.Append(0x7fffffff)
		if atInt(t, l, 0) != -1 || atInt(t, l, 1) != 0x7fffffff {
			t.Fatalf("widened read = %d, %d", atInt(t, l, 0), atInt(t, l, 1))
		}
	})

	t.Run("put_and_insert_from_runtime", func(t *testing.T) {
		l := NewInt32List()
		l.AppendFromAny(boxInt(100))
		l.PutFromAny(0, boxUint(200))
		l.InsertAtFromAny(1, boxInt(-3))
		want := []int{
			goWidenInt32(goNarrow[int32](boxUint(200))),
			goWidenInt32(goNarrow[int32](boxInt(-3))),
		}
		for i, w := range want {
			if got := atInt(t, l, i); got != w {
				t.Fatalf("[%d] = %d, want %d", i, got, w)
			}
		}
	})

	t.Run("rejects_non_integral", func(t *testing.T) {
		for _, v := range []any{nil, String("1"), Float(1.0)} {
			expectPanic(t, func() { NewInt32List().AppendFromAny(v) })
		}
	})
}

func TestIntegralListTypeRegistration(t *testing.T) {
	cases := []struct {
		name     string
		newList  func() Typed
		typeName string
	}{
		{"ByteList", func() Typed { return NewByteList() }, "ByteList"},
		{"Int32List", func() Typed { return NewInt32List() }, "Int32List"},
		{"IntList", func() Typed { return NewIntList() }, "IntList"},
		{"UintList", func() Typed { return NewUintList() }, "UintList"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := tc.newList()
			wantT := GetTypeRegistrar().Get(tc.typeName)
			if wantT == nil {
				t.Fatalf("type %q not registered", tc.typeName)
			}
			if l.FinalType() != wantT {
				t.Fatalf("FinalType() = %v, want %q", l.FinalType(), tc.typeName)
			}
		})
	}
}

func TestByteListFromBytesBindsByteListType(t *testing.T) {
	byteT := GetTypeRegistrar().Get("ByteList")
	if byteT == nil {
		t.Fatal("ByteList must be registered")
	}
	fromBytes := NewByteListFromBytes([]byte("hi"))
	if fromBytes.FinalType() != byteT {
		t.Fatalf("NewByteListFromBytes FinalType = %v, want ByteList", fromBytes.FinalType())
	}
	if GetTypeRegistrar().Get("Uint8List") != nil {
		t.Fatal("Uint8List must not be registered")
	}
	if GetTypeRegistrar().Get("RuneList") != nil {
		t.Fatal("RuneList must not be registered")
	}
}

// @() must widen via AtAsAny — raw storage (int8 etc.) must not leak onto DS.
func TestInt8ListGetByOrdinalWidensToInt(t *testing.T) {
	l := NewInt8List()
	l.Append(1)
	l.Append(0)
	im := GetTypeRegistrar().Get("Int8List").itemsMetadata
	got := im.GetByOrdinal(l, Int(0))
	if v, ok := got.(Int); !ok || v != 1 {
		t.Fatalf("GetByOrdinal = %T(%v), want Int(1)", got, got)
	}
	if _, isInt8 := got.(int8); isInt8 {
		t.Fatal("GetByOrdinal leaked int8 storage type")
	}
}

func TestOpAtIndexInt8ListWidensToInt(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	l := NewInt8List()
	l.Append(7)
	rp.DataStack.Push(l)
	rp.DataStack.Push(Int(0))
	RunOps(task, rp, OpAtIndex, OpStop)
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 7 {
		t.Fatalf("OpAtIndex = %T(%v), want Int(7)", got, got)
	}
}

func expectPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}
