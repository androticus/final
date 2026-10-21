package core

import (
	"strings"
	"testing"
)

func TestStringBuilderWriteAndToString(t *testing.T) {
	b := NewStringBuilder()
	b.WriteString("hello")
	b.WriteRune(' ')
	b.WriteLine("world")
	if got := b.ToString(); got != "hello world\n" {
		t.Fatalf("ToString() = %q, want %q", got, "hello world\n")
	}
}

func TestNewStringBuilderFinalType(t *testing.T) {
	sbT := GetTypeRegistrar().Get("StringBuilder")
	if sbT == nil {
		t.Fatal("StringBuilder type not registered")
	}
	b := NewStringBuilder()
	if b.FinalType() != sbT {
		t.Fatalf("FinalType() = %v, want StringBuilder", b.FinalType())
	}
}

func TestStringBuilderWriteRuneFromIntCoerce(t *testing.T) {
	b := NewStringBuilder()
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(65))
	mm := GetTypeRegistrar().Get("StringBuilder").MethodByName("writeRune")
	f, ths := memberRunFrame(rp, b)
	mm.invokeCall(f, ths)
	if b.ToString() != "A" {
		t.Fatalf("ToString() = %q, want %q", b.ToString(), "A")
	}
}

func TestStringBuilderCheckCoerceTypeRejectsWrongType(t *testing.T) {
	sbT := GetTypeRegistrar().Get("StringBuilder")
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if !strings.Contains(panicMsg(r), "expected StringBuilder") {
			t.Fatalf("panic = %v", r)
		}
	}()
	sbT.checkCoerceType(Int(1))
}

func TestInterpretStringBuilderWriteAndToString(t *testing.T) {
	ResetTestRuntime()
	task, rp, _ := testParserWithInput(t, "StringBuilder.new^ >ts \"hello\" t:.writeString^ 32 t:.writeRune^ t:.toString^\n")
	if err := interpretSession(task, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "hello " {
		t.Fatalf("result = %q, want %q", got, "hello ")
	}
}

func panicMsg(r any) string {
	if err, ok := r.(error); ok {
		return err.Error()
	}
	return ""
}
