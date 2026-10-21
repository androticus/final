package core

import (
	"math"
	"testing"
)

func execCastTypeProbe(t *testing.T, rp *RunProvider, subCode int, in any) Bool {
	t.Helper()
	rp.DataStack.Push(in)
	RunOpsWithSubCode(NewTestTask(), rp, subCode, OpCast, OpStop)
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1", rp.DataStack.Len())
	}
	return rp.DataStack.Pop().(Bool)
}

func runCastConvert(t *testing.T, rp *RunProvider, subCode int, in any) (any, Bool, int) {
	t.Helper()
	rp.DataStack.Push(in)
	RunOpsWithSubCode(NewTestTask(), rp, subCode, OpCast, OpStop)
	n := rp.DataStack.Len()
	if n != 1 && n != 2 {
		t.Fatalf("DS len = %d, want 1 or 2", n)
	}
	if n == 1 {
		ok := rp.DataStack.Pop().(Bool)
		if ok {
			t.Fatalf("DS len 1 but flag = true")
		}
		return nil, false, 1
	}
	ok := rp.DataStack.Pop().(Bool)
	if !ok {
		t.Fatalf("DS len 2 but flag = false")
	}
	return rp.DataStack.Pop(), true, 2
}

func TestCastExactTypeProbes(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()

	cases := []struct {
		subCode int
		in      any
		want    Bool
	}{
		{SC_cast_nullType, nil, true},
		{SC_cast_nullType, Int(0), false},
		{SC_cast_boolType, Bool(true), true},
		{SC_cast_boolType, Int(1), false},
		{SC_cast_intType, Int(42), true},
		{SC_cast_intType, Uint(1), false},
		{SC_cast_uintType, Uint(1), true},
		{SC_cast_floatType, Float(1.5), true},
		{SC_cast_stringType, String("x"), true},
		{SC_cast_stringType, String(""), true},
	}
	for _, tc := range cases {
		got := execCastTypeProbe(t, rp, tc.subCode, tc.in)
		if got != tc.want {
			t.Fatalf("subcode %d(%v): got %v, want %v", tc.subCode, tc.in, got, tc.want)
		}
	}
}

func TestCastRepToInt(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()

	out, ok, n := runCastConvert(t, rp, SC_cast_toInt, Uint(7))
	if !ok || n != 2 || out.(Int) != 7 {
		t.Fatalf("Uint→Int = %v ok=%v depth=%d", out, ok, n)
	}
	out, ok, n = runCastConvert(t, rp, SC_cast_toInt, Float(3.9))
	if !ok || n != 2 || out.(Int) != 3 {
		t.Fatalf("Float→Int = %v ok=%v depth=%d", out, ok, n)
	}
	_, ok, n = runCastConvert(t, rp, SC_cast_toInt, Bool(true))
	if ok || n != 1 {
		t.Fatalf("Bool→Int should fail with depth 1, got ok=%v depth=%d", ok, n)
	}
	// >Int wraps overflowing Uint (Go bit-cast)
	big := Uint(math.MaxInt) + 1
	out, ok, n = runCastConvert(t, rp, SC_cast_toInt, big)
	if !ok || n != 2 || out.(Int) != Int(big) {
		t.Fatalf(">Int overflowing Uint wrap = %v ok=%v depth=%d", out, ok, n)
	}
}

func TestCastSoftIntRejectsOverflowingUint(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()

	big := Uint(math.MaxInt) + 1
	_, ok, n := runCastConvert(t, rp, SC_cast_softInt, big)
	if ok || n != 1 {
		t.Fatal(">>Int? should reject Uint that overflows Int")
	}
	out, ok, n := runCastConvert(t, rp, SC_cast_softInt, Uint(math.MaxInt))
	if !ok || n != 2 || out.(Int) != math.MaxInt {
		t.Fatalf(">>Int? MaxInt Uint = %v ok=%v depth=%d", out, ok, n)
	}
}

func TestCastRepToUintWrapsNegative(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()

	out, ok, n := runCastConvert(t, rp, SC_cast_toUint, Int(-1))
	if !ok || n != 2 || out.(Uint) != Uint(^uint(0)) {
		t.Fatalf("negative Int→Uint wrap = %v ok=%v depth=%d", out, ok, n)
	}
	out, ok, n = runCastConvert(t, rp, SC_cast_toUint, Float(-1.5))
	if !ok || n != 2 || out.(Uint) != Uint(^uint(0)) {
		t.Fatalf("negative Float→Uint wrap = %v ok=%v depth=%d", out, ok, n)
	}
}

func TestCastSoftUintRejectsNegative(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()

	_, ok, n := runCastConvert(t, rp, SC_cast_softUint, Int(-1))
	if ok || n != 1 {
		t.Fatal(">>Uint? should reject negative Int")
	}
	_, ok, n = runCastConvert(t, rp, SC_cast_softUint, Float(-0.5))
	if ok || n != 1 {
		t.Fatal(">>Uint? should reject negative Float")
	}
	out, ok, n := runCastConvert(t, rp, SC_cast_softUint, Int(7))
	if !ok || n != 2 || out.(Uint) != 7 {
		t.Fatalf(">>Uint? non-neg Int = %v ok=%v depth=%d", out, ok, n)
	}
}

func TestCastSoftBool(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()

	out, ok, n := runCastConvert(t, rp, SC_cast_softBool, String("true"))
	if !ok || n != 2 || out.(Bool) != true {
		t.Fatalf(">>Bool? true = %v ok=%v depth=%d", out, ok, n)
	}
	_, ok, n = runCastConvert(t, rp, SC_cast_softBool, String("1"))
	if ok || n != 1 {
		t.Fatal(">>Bool? should reject non-json strings")
	}
	_, ok, n = runCastConvert(t, rp, SC_cast_toBool, String("true"))
	if ok || n != 1 {
		t.Fatal(">Bool? should reject String")
	}
}

func TestCastSoftIntString(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()

	out, ok, n := runCastConvert(t, rp, SC_cast_softInt, String("0x10"))
	if !ok || n != 2 || out.(Int) != 16 {
		t.Fatalf(">>Int? hex = %v ok=%v depth=%d", out, ok, n)
	}
	_, ok, n = runCastConvert(t, rp, SC_cast_softInt, String("3.14"))
	if ok || n != 1 {
		t.Fatal(">>Int? should reject float strings")
	}
	_, ok, n = runCastConvert(t, rp, SC_cast_softInt, String("10u"))
	if ok || n != 1 {
		t.Fatal(">>Int? should reject Uint-suffixed strings")
	}
}

func TestCastSoftNullAndString(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()

	out, ok, n := runCastConvert(t, rp, SC_cast_softNull, String("null"))
	if !ok || n != 2 || out != nil {
		t.Fatalf(">>Null? = %v ok=%v depth=%d", out, ok, n)
	}
	_, ok, n = runCastConvert(t, rp, SC_cast_softNull, nil)
	if ok || n != 1 {
		t.Fatal(">>Null? should accept only string null")
	}

	out, ok, n = runCastConvert(t, rp, SC_cast_softString, Int(10))
	if !ok || n != 2 || out.(String) != "10" {
		t.Fatalf(">>String? Int = %v ok=%v depth=%d", out, ok, n)
	}
	_, ok, n = runCastConvert(t, rp, SC_cast_softString, String("hi"))
	if ok || n != 1 {
		t.Fatal(">>String? should reject String input")
	}
}

func TestCastOpsRegistered(t *testing.T) {
	ResetTestRuntime()
	names := []string{
		"Null?", "Bool?", "Int?", "Uint?", "Float?", "String?",
		">Bool?", ">Int?", ">Uint?", ">Float?",
		">>Null?", ">>Bool?", ">>Int?", ">>Uint?", ">>Float?", ">>String?",
	}
	for _, name := range names {
		w := globalDictionaryContext.FindWordByName(name)
		if w == nil {
			t.Fatalf("cast word %q not in dictionary", name)
		}
		if w.opCode != OpCast {
			t.Fatalf("%q opCode = %v, want OpCast", name, w.opCode)
		}
	}
}
