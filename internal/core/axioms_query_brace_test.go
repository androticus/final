package core

import (
	"fmt"
	"strings"
	"testing"
)

func TestSplitCompoundSegmentsQueryBrace(t *testing.T) {
	got, err := splitCompoundSegments("m?{}")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"m", segQueryBrace}
	if len(got) != len(want) {
		t.Fatalf("segments = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("segments[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestOpQueryBraceDynamicObjectMissing(t *testing.T) {
	ResetTestRuntime()
	do := NewDynamicObject()
	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, do)
	rp.DataStack.Push(String("counter"))
	RunOps(NewTestTask(), rp, OpQueryBrace, OpStop)
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want 2 (receiver + has)", rp.DataStack.Len())
	}
	has, ok := rp.DataStack.Pop().(Bool)
	if !ok || has {
		t.Fatalf("has = %v, want false", has)
	}
	if rp.DataStack.Pop() != do {
		t.Fatalf("receiver not preserved on DS")
	}
}

func TestOpQueryBraceDynamicObjectPresent(t *testing.T) {
	ResetTestRuntime()
	do := NewDynamicObject()
	do.Set("counter", Int(7))
	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, do)
	rp.DataStack.Push(String("counter"))
	RunOps(NewTestTask(), rp, OpQueryBrace, OpStop)
	has, ok := rp.DataStack.Pop().(Bool)
	if !ok || !has {
		t.Fatalf("has = %v, want true", has)
	}
}

func TestOpAtBraceDynamicObjectMissingPanics(t *testing.T) {
	ResetTestRuntime()
	do := NewDynamicObject()
	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, do)
	rp.DataStack.Push(String("counter"))
	defer func() {
		if recover() == nil {
			t.Fatal("expected @{} panic on missing field")
		}
	}()
	RunOps(NewTestTask(), rp, OpAtBrace, OpStop)
}

func TestOpQueryBraceMap(t *testing.T) {
	ResetTestRuntime()
	m := NewStringMap()
	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, m)
	rp.DataStack.Push(String("k"))
	RunOps(NewTestTask(), rp, OpQueryBrace, OpStop)
	has, ok := rp.DataStack.Pop().(Bool)
	if !ok || has {
		t.Fatalf("empty map has k = %v, want false", has)
	}
	m.Put(String("k"), Int(1))
	pushTypedOnContextStack(rp, m)
	rp.DataStack.Push(String("k"))
	RunOps(NewTestTask(), rp, OpQueryBrace, OpStop)
	has, ok = rp.DataStack.Pop().(Bool)
	if !ok || !has {
		t.Fatalf("map with k has = %v, want true", has)
	}
}

func TestOpQueryBraceListUnsupported(t *testing.T) {
	ResetTestRuntime()
	l := NewAnyList()
	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, l)
	rp.DataStack.Push(String("nope"))
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for List ?{}")
		}
		msg := fmt.Sprint(r)
		if !strings.Contains(msg, "hasByKey") {
			t.Fatalf("panic = %v, want hasByKey", r)
		}
	}()
	RunOps(NewTestTask(), rp, OpQueryBrace, OpStop)
}
