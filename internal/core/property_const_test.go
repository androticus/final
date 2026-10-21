package core

import (
	"fmt"
	"testing"
)

type constHostInst struct{}

func (constHostInst) FinalType() *Type { return constHostType }

var (
	constHostType  *Type
	constHostChild *Word
)

func init() {
	constHostType = &Type{
		name: "ConstHost",
		getZero: func() any {
			return constHostInst{}
		},
		checkCoerceType: func(value any) any {
			if _, ok := value.(constHostInst); !ok {
				panic(fmt.Errorf("ConstHost.checkCoerceType: value %T is not constHostInst", value))
			}
			return value
		},
	}
	constHostType.registerProperty("child", "Word",
		func(_ Any) Any {
			if constHostChild == nil {
				panic("constHostChild unset")
			}
			return constHostChild
		},
		nil,
	)
	GetTypeRegistrar().RegisterCurrentType(constHostType)
}

func TestLiftReceiverFromCSSkipsRefsPrefersTOSWord(t *testing.T) {
	constHostChild = NewWord("x", 0)
	rp := NewTestRunProvider()
	rp.DataStack.Push(constHostInst{})
	rp.DataStack.Push(constHostChild)
	got := pairFromCS(rp, rp.DataStack, ReceiverLayoutSkipRefsTOS())
	w, ok := got.Obj.(*Word)
	if !ok || w != constHostChild {
		t.Fatalf("pairFromCS = %T %v, want *Word TOS", got.Obj, got.Obj)
	}
}

func TestWordImplementsTyped(t *testing.T) {
	w := NewWord("x", 0)
	var _ Typed = w
	if _, ok := any(w).(Typed); !ok {
		t.Fatal("*Word should satisfy Typed")
	}
}

// Mid-path readonly property: .child .name chain reads through to terminal value (implied @).
func TestConstPropertyMidPathCompoundChain(t *testing.T) {
	ResetTestRuntime()
	constHostChild = NewWord("childName", 0)
	task := NewTestTask()
	rp := NewTestRunProvider()
	rp.DataStack.Push(constHostInst{})
	runHandleCompoundWithSegs(task, rp, []string{".", "child", ".", "name"})
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want 2 (receiver + terminal value)", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "childName" {
		t.Fatalf("TOS = %v, want childName", got)
	}
}
