package core

import (
	"testing"
)

func TestOpCodeNameAndByName(t *testing.T) {
	if OpCodeName(OpStop) != "Stop" {
		t.Fatalf("OpCodeName(OpStop) = %q", OpCodeName(OpStop))
	}
	op, ok := OpCodeByName("Add")
	if !ok || op != OpAdd {
		t.Fatalf("OpCodeByName(Add) = %v, %v", op, ok)
	}
	_, ok = OpCodeByName("NotAnOpCode")
	if ok {
		t.Fatal("expected miss for unknown name")
	}
}
