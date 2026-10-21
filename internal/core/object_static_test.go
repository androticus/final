package core

import (
	"strings"
	"testing"
)

func TestStaticObjectFinalTypeAndParent(t *testing.T) {
	ResetTestRuntime()
	proto := NewDynamicObject()
	proto.Set("a", Int(1))
	o := NewStaticObject(proto)
	if o.FinalType() != staticObjectType {
		t.Fatalf("FinalType = %v, want StaticObject", o.FinalType())
	}
	if staticObjectType.GetParent() != mapProtoType {
		t.Fatalf("parent = %v, want ProtoMap", staticObjectType.GetParent())
	}
}

func TestStaticObjectNewStaticShallowCopy(t *testing.T) {
	ResetTestRuntime()
	proto := NewDynamicObject()
	proto.Set("mass", Float(1.5))
	proto.Set("vx", Float(0))

	rp := NewTestRunProvider()
	mm := dynamicObjectType.MethodByName("newStatic")
	f, ths := memberRunFrame(rp, proto)
	mm.invokeCall(f, ths)
	o, ok := rp.DataStack.Pop().(*StaticObject)
	if !ok || o == nil {
		t.Fatalf("newStatic = %T", rp.DataStack.Peek())
	}
	if o.GetPrototype() != proto {
		t.Fatal("prototype pointer mismatch")
	}
	if o.Len() != 2 || o.Get("mass") != Float(1.5) || o.Get("vx") != Float(0) {
		t.Fatalf("seeded values mass=%v vx=%v len=%d", o.Get("mass"), o.Get("vx"), o.Len())
	}
	o.Set("mass", Float(9))
	if proto.Get("mass") != Float(1.5) {
		t.Fatal("instance Set must not mutate prototype values")
	}
}

func TestStaticObjectSetUnknownPanics(t *testing.T) {
	proto := NewDynamicObject()
	proto.Set("a", Int(1))
	o := NewStaticObject(proto)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for unknown field Set")
		}
		msg, ok := r.(error)
		if !ok || !strings.Contains(msg.Error(), "not found") {
			t.Fatalf("panic = %v, want not found", r)
		}
	}()
	o.Set("missing", Int(2))
}

func TestStaticObjectDotFieldImpliedRead(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	proto := NewDynamicObject()
	proto.Set("myField", Int(42))
	o := NewStaticObject(proto)
	rp.DataStack.Push(o)
	runHandleCompoundWithSegs(task, rp, []string{".", "myField"})
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want 2 (receiver + value)", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 42 {
		t.Fatalf("TOS = %v, want Int(42)", got)
	}
}

func TestStaticObjectDotFieldStoreNoGrow(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	proto := NewDynamicObject()
	proto.Set("greeting", String(""))
	o := NewStaticObject(proto)
	rp.TempStack.Push(o)
	rp.DataStack.Push(String("hello"))
	cp := CompoundPathWithSelectorForTest([]string{".", "greeting", "!"}, csTStack)
	if err := dispatchCompoundToken(p, task, rp, cp); err != nil {
		t.Fatal(err)
	}
	if got := o.Get("greeting"); got != String("hello") {
		t.Fatalf("greeting = %v, want hello", got)
	}
	if proto.Len() != 1 {
		t.Fatalf("prototype grew: len=%d", proto.Len())
	}
}

func TestStaticObjectDotFieldMissingPanics(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	proto := NewDynamicObject()
	proto.Set("a", Int(1))
	o := NewStaticObject(proto)
	rp.DataStack.Push(o)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for missing field")
		}
		msg, ok := r.(error)
		if !ok || !strings.Contains(msg.Error(), "not found") {
			t.Fatalf("panic = %v, want missing-key error", r)
		}
	}()
	runHandleCompoundWithSegs(task, rp, []string{".", "missing"})
}

func TestStaticObjectOpAtIndexBangIndex(t *testing.T) {
	ResetTestRuntime()
	proto := NewDynamicObject()
	proto.Set("mass", Float(1.5))
	proto.Set("x", Float(10))
	proto.Set("y", Float(20))
	o := NewStaticObject(proto)

	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, o)
	rp.DataStack.Push(Int(0))
	RunOps(NewTestTask(), rp, OpAtIndex, OpStop)
	if got := rp.DataStack.Pop(); got != Float(1.5) {
		t.Fatalf("@() 0 = %v, want 1.5", got)
	}
	if rp.DataStack.Pop() != o {
		t.Fatal("receiver not preserved on DS after @()")
	}

	rp = NewTestRunProvider()
	pushTypedOnContextStack(rp, o)
	rp.DataStack.Push(Float(99.5))
	rp.DataStack.Push(Int(1))
	RunOps(NewTestTask(), rp, OpBangIndex, OpStop)
	if rp.DataStack.Pop() != o {
		t.Fatal("receiver not preserved on DS after !()")
	}
	if got := o.Get("x"); got != Float(99.5) {
		t.Fatalf("x after !() 1 = %v, want 99.5", got)
	}
	if proto.Get("x") != Float(10) {
		t.Fatal("!() must not mutate prototype")
	}
}

func TestStaticObjectItemsByKey(t *testing.T) {
	ResetTestRuntime()
	proto := NewDynamicObject()
	proto.Set("k", Int(1))
	o := NewStaticObject(proto)
	im := staticObjectType.GetItemsMetadata()
	if im == nil || im.getByKey == nil || im.setByKey == nil || im.hasByKey == nil {
		t.Fatal("missing @{} !{} ?{} items metadata")
	}
	if im.getByOrdinal == nil || im.setByOrdinal == nil {
		t.Fatal("missing @() !() items metadata")
	}
	im.SetByKey(o, String("k"), Int(9))
	if got := im.GetByKey(o, String("k")); got != Int(9) {
		t.Fatalf("GetByKey = %v, want 9", got)
	}
	if !im.HasByKey(o, String("k")) || im.HasByKey(o, String("missing")) {
		t.Fatal("HasByKey mismatch")
	}
}

func TestStaticObjectGetIter(t *testing.T) {
	proto := NewDynamicObject()
	proto.Set("a", Int(1))
	proto.Set("b", Int(2))
	o := NewStaticObject(proto)
	o.Set("a", Int(10))
	it := o.GetIter()
	var keys []string
	var vals []any
	for it.GetNext() {
		kv := it.Current().(*AnyKeyValue)
		keys = append(keys, string(kv.Key.(String)))
		vals = append(vals, kv.Value)
	}
	if len(keys) != 2 || keys[0] != "a" || keys[1] != "b" {
		t.Fatalf("keys = %v, want [a b]", keys)
	}
	if vals[0] != Int(10) || vals[1] != Int(2) {
		t.Fatalf("vals = %v, want [10 2]", vals)
	}
}

func TestStaticObjectFromLargePrototype(t *testing.T) {
	n := sfmMaxIteratedFields + 6
	proto := NewDynamicObject()
	fillDynamicObject(proto, n)
	o := NewStaticObject(proto)
	if o.Len() != n {
		t.Fatalf("Len = %d, want %d", o.Len(), n)
	}
	if o.Get(sfmTestKey(0)) != Int(0) || o.Get(sfmTestKey(n-1)) != Int(n-1) {
		t.Fatal("named Get mismatch on large prototype")
	}
	if o.GetPrototype() != proto {
		t.Fatal("prototype pointer mismatch")
	}
	o.Set(sfmTestKey(n-1), Int(99))
	if proto.Get(sfmTestKey(n-1)) != Int(n-1) {
		t.Fatal("instance Set must not mutate prototype values")
	}
	if o.Get(sfmTestKey(n-1)) != Int(99) {
		t.Fatalf("Get after Set = %v, want 99", o.Get(sfmTestKey(n-1)))
	}
}
