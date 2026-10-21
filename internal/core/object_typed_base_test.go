package core

import (
	"fmt"
	"strings"
	"testing"
)

// testBaseHost is a minimal Base-capable catalog type for TypedObject baseType tests.
type testBaseHost struct {
	del *TypedObject
	tag string
}

func (h *testBaseHost) FinalType() *Type { return testBaseHostType }

var testBaseHostType *Type

func ensureTestBaseHostType(t *testing.T) *Type {
	t.Helper()
	if testBaseHostType != nil && GetTypeRegistrar().Get(testBaseHostType.Name()) == testBaseHostType {
		return testBaseHostType
	}
	name := fmt.Sprintf("TestBaseHost_%s", strings.ReplaceAll(t.Name(), "/", "_"))
	typ := NewRegisteringType(name, ReferenceTypeZeroFn, MakePointerTypeCheckCoerce[testBaseHost](name))
	typ.registerNativeClassMethod("_newWithDelegate", func(f *RunFrame, _ *TypeObjPair) {
		del := f.RP.DataStack.Pop().(*TypedObject)
		f.RP.DataStack.Push(&testBaseHost{del: del, tag: "wired"})
	})
	GetTypeRegistrar().RegisterCurrentType(typ)
	testBaseHostType = typ
	return typ
}

func TestTypedObjectBaseTypeMintWiresBaseAndDel(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	base := ensureTestBaseHostType(t)

	childName := fmt.Sprintf("BasedChild_%s", t.Name())
	def := NewTypeDef()
	def.Set("name", String(childName))
	fields := def.Get("fields").(*DynamicObject)
	fields.Set("score", GetTypeRegistrar().Get("Int"))
	def.Set("baseType", base)
	child := CreateTypeFromDef(task.dc, def)

	if child.baseType != base {
		t.Fatalf("baseType = %v, want %v", child.baseType, base)
	}
	if child.GetPropertyMetadata("base") == nil {
		t.Fatal("missing auto field base")
	}

	f := &RunFrame{Task: task, RP: rp}
	to := MakeNewTypedObject(child)
	FinishNewTypedObject(f, to)

	raw := to.GetFieldByName("base")
	host, ok := raw.(*testBaseHost)
	if !ok {
		t.Fatalf("base = %T, want *testBaseHost", raw)
	}
	if host.del != to {
		t.Fatal("base.del is not the TypedObject")
	}
	if host.tag != "wired" {
		t.Fatalf("tag = %q", host.tag)
	}
	if to.GetFieldByName("score") != Int(0) {
		t.Fatalf("score = %v, want 0", to.GetFieldByName("score"))
	}
}

func TestTypedObjectBaseTypeProtoNew(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	base := ensureTestBaseHostType(t)

	childName := fmt.Sprintf("BasedNew_%s", t.Name())
	def := NewTypeDef()
	def.Set("name", String(childName))
	fields := def.Get("fields").(*DynamicObject)
	fields.Set("n", GetTypeRegistrar().Get("Int"))
	def.Set("baseType", base)
	child := CreateTypeFromDef(task.dc, def)

	rp.DataStack.Push(child)
	runHandleCompoundWithSegs(task, rp, []string{".", "new", "^"})
	inst, ok := rp.DataStack.Pop().(*TypedObject)
	if !ok {
		t.Fatalf("got %T, want *TypedObject", rp.DataStack.Peek())
	}
	host, ok := inst.GetFieldByName("base").(*testBaseHost)
	if !ok || host.del != inst {
		t.Fatalf("base not wired: %T %#v", inst.GetFieldByName("base"), host)
	}
}

func TestTypedObjectBaseTypeRejectsUserBaseField(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	base := ensureTestBaseHostType(t)

	def := NewTypeDef()
	def.Set("name", String(fmt.Sprintf("ClashBase_%s", t.Name())))
	fields := def.Get("fields").(*DynamicObject)
	fields.Set("base", GetTypeRegistrar().Get("Int"))
	def.Set("baseType", base)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for user field base")
		}
		msg := fmt.Sprint(r)
		if !strings.Contains(msg, "must not declare") {
			t.Fatalf("panic = %v", r)
		}
	}()
	CreateTypeFromDef(task.dc, def)
}

func TestTypedObjectBaseTypeRejectsNonBase(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()

	def := NewTypeDef()
	def.Set("name", String(fmt.Sprintf("BadBase_%s", t.Name())))
	fields := def.Get("fields").(*DynamicObject)
	fields.Set("n", GetTypeRegistrar().Get("Int"))
	def.Set("baseType", GetTypeRegistrar().Get("Int"))

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for Int as baseType")
		}
		msg := fmt.Sprint(r)
		if !strings.Contains(msg, "_newWithDelegate") {
			t.Fatalf("panic = %v", r)
		}
	}()
	CreateTypeFromDef(task.dc, def)
}
