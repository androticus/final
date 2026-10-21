package core

import (
	"fmt"
	"strings"
	"testing"
)

func makeTypedObjectTypeForTest(t *testing.T, dc *DictionaryContext, typeName string, fieldNames ...string) *Type {
	t.Helper()
	tr := dc.tr
	def := NewTypeDef()
	def.Set("name", String(typeName))
	fields := def.Get("fields").(*DynamicObject)
	for _, name := range fieldNames {
		switch name {
		case "alpha":
			fields.Set(name, tr.Get("String"))
		case "beta":
			fields.Set(name, tr.Get("Int"))
		case "gamma":
			fields.Set(name, tr.Get("Bool"))
		default:
			t.Fatalf("unsupported test field %q", name)
		}
	}
	return CreateTypeFromDef(dc, def)
}

func typedObjectPropertyGet(t *testing.T, inst *TypedObject, name string) any {
	t.Helper()
	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, inst)
	rp.DataStack.Push(String(name))
	RunOpsWithSubCode(NewTestTask(), rp, int(TypeDispatchKindInstanceGet), OpTypeMemberDispatch, OpStop)
	return rp.DataStack.Pop()
}

func typedObjectPropertySet(t *testing.T, inst *TypedObject, name string, value any) {
	t.Helper()
	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, inst)
	rp.DataStack.Push(value)
	rp.DataStack.Push(String(name))
	RunOpsWithSubCode(NewTestTask(), rp, int(RuntimeDispatchKindInstanceSetByName), OpNativeInvoke, OpStop)
}

func TestTypedObjectNewInstanceFieldsExistWithZeroValues(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmt.Sprintf("FieldsExist_%s", t.Name()), "alpha", "beta", "gamma")
	inst := MakeNewTypedObject(typ)

	want := map[string]any{
		"alpha": String(""),
		"beta":  Int(0),
		"gamma": Bool(false),
	}
	for name, zero := range want {
		got := typedObjectPropertyGet(t, inst, name)
		if got != zero {
			t.Fatalf(".%s on new instance = %v, want zero %v", name, got, zero)
		}
	}
}

func TestTypedObjectPropertyReadWrite(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmt.Sprintf("PropRW_%s", t.Name()), "alpha", "beta", "gamma")
	inst := MakeNewTypedObject(typ)

	cases := []struct {
		name  string
		value any
	}{
		{"alpha", String("hello")},
		{"beta", Int(42)},
		{"gamma", Bool(true)},
	}
	for _, tc := range cases {
		typedObjectPropertySet(t, inst, tc.name, tc.value)
		got := typedObjectPropertyGet(t, inst, tc.name)
		if got != tc.value {
			t.Fatalf(".%s after set = %v, want %v", tc.name, got, tc.value)
		}
	}
}

func TestTypedObjectPropertyGetUnknownPanics(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmt.Sprintf("PropMiss_%s", t.Name()), "alpha")
	inst := MakeNewTypedObject(typ)
	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, inst)
	rp.DataStack.Push(String("nosuch"))
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for unknown property")
		}
	}()
	RunOpsWithSubCode(NewTestTask(), rp, int(TypeDispatchKindInstanceGet), OpTypeMemberDispatch, OpStop)
}

func TestTypedObjectItemsGetByKey(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmt.Sprintf("ItemsGet_%s", t.Name()), "alpha", "beta")
	inst := MakeNewTypedObject(typ)
	inst.SetFieldByOrdinal(0, String("read-me"))
	inst.SetFieldByOrdinal(1, Int(9))

	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, inst)
	rp.DataStack.Push(String("alpha"))
	RunOps(NewTestTask(), rp, OpAtBrace, OpStop)
	if got := rp.DataStack.Pop(); got != String("read-me") {
		t.Fatalf("@{} alpha = %v, want read-me", got)
	}
	if rp.DataStack.Pop() != inst {
		t.Fatal("receiver not preserved on DS after @{}")
	}

	rp = NewTestRunProvider()
	rp.DataStack.Push(inst)
	rp.DataStack.Push(String("beta"))
	RunOpsWithSubCode(NewTestTask(), rp, int(TypeDispatchKindInstanceItemGetByKey), OpTypeMemberDispatch, OpStop)
	if got := rp.DataStack.Pop(); got != Int(9) {
		t.Fatalf("itemsGetByKey beta = %v, want 9", got)
	}
}

func TestTypedObjectItemsGetByKeyMissingPanics(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmt.Sprintf("ItemsGetMiss_%s", t.Name()), "alpha")
	inst := MakeNewTypedObject(typ)
	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, inst)
	rp.DataStack.Push(String("missing"))
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected @{} panic on missing field")
		}
		if !strings.Contains(fmt.Sprint(r), "unknown name") {
			t.Fatalf("panic = %v, want unknown name", r)
		}
	}()
	RunOps(NewTestTask(), rp, OpAtBrace, OpStop)
}

func TestTypedObjectItemsSetByKey(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmt.Sprintf("ItemsSet_%s", t.Name()), "alpha", "beta")
	inst := MakeNewTypedObject(typ)

	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, inst)
	rp.DataStack.Push(String("stored"))
	rp.DataStack.Push(String("alpha"))
	RunOps(NewTestTask(), rp, OpBangBrace, OpStop)
	if rp.DataStack.Pop() != inst {
		t.Fatal("receiver not preserved on DS after !{}")
	}
	if inst.GetFieldByOrdinal(0) != String("stored") {
		t.Fatalf("alpha after !{} = %v, want stored", inst.GetFieldByOrdinal(0))
	}

	rp = NewTestRunProvider()
	pushTypedOnContextStack(rp, inst)
	rp.DataStack.Push(Int(77))
	rp.DataStack.Push(String("beta"))
	RunOps(NewTestTask(), rp, OpBangBrace, OpStop)
	if inst.GetFieldByOrdinal(1) != Int(77) {
		t.Fatalf("beta after !{} = %v, want 77", inst.GetFieldByOrdinal(1))
	}
	if got := typedObjectPropertyGet(t, inst, "beta"); got != Int(77) {
		t.Fatalf(".beta after !{} = %v, want 77", got)
	}
}

func TestTypedObjectItemsSetByKeyMissingPanics(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmt.Sprintf("ItemsSetMiss_%s", t.Name()), "alpha")
	inst := MakeNewTypedObject(typ)
	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, inst)
	rp.DataStack.Push(String("x"))
	rp.DataStack.Push(String("missing"))
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected !{} panic on missing field")
		}
		if !strings.Contains(fmt.Sprint(r), "unknown name") {
			t.Fatalf("panic = %v, want unknown name", r)
		}
	}()
	RunOps(NewTestTask(), rp, OpBangBrace, OpStop)
}

func TestTypedObjectItemsHasByKey(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmt.Sprintf("ItemsHas_%s", t.Name()), "alpha")
	inst := MakeNewTypedObject(typ)

	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, inst)
	rp.DataStack.Push(String("alpha"))
	RunOps(NewTestTask(), rp, OpQueryBrace, OpStop)
	has, ok := rp.DataStack.Pop().(Bool)
	if !ok || !has {
		t.Fatalf("?{} alpha = %v, want true", has)
	}
	if rp.DataStack.Pop() != inst {
		t.Fatal("receiver not preserved on DS after ?{}")
	}

	rp = NewTestRunProvider()
	pushTypedOnContextStack(rp, inst)
	rp.DataStack.Push(String("missing"))
	RunOps(NewTestTask(), rp, OpQueryBrace, OpStop)
	has, ok = rp.DataStack.Pop().(Bool)
	if !ok || has {
		t.Fatalf("?{} missing = %v, want false", has)
	}
}

func TestTypedObjectCreateTypeFromDefRegistersItemsMetadata(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmt.Sprintf("ItemsMeta_%s", t.Name()), "alpha")
	im := typ.GetItemsMetadata()
	if im == nil || im.getByKey == nil || im.setByKey == nil || im.hasByKey == nil {
		t.Fatal("expected @{} !{} ?{} items metadata on dynamic TypedObject type")
	}
	if im.getByOrdinal == nil || im.setByOrdinal == nil {
		t.Fatal("expected @() !() items metadata on TypedObject type")
	}
}

func TestTypedObjectNewTypeDefHasConstantsNotMethods(t *testing.T) {
	ResetTestRuntime()
	def := NewTypeDef()
	if !def.HasField("constants") {
		t.Fatal("newTypeDef missing constants")
	}
	if def.HasField("methods") {
		t.Fatal("newTypeDef should not have methods")
	}
	c := def.Get("constants").(*DynamicObject)
	if c.Len() != 0 {
		t.Fatalf("constants len = %d, want 0", c.Len())
	}
}

func TestTypedObjectClassConstants(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	def := NewTypeDef()
	def.Set("name", String(fmt.Sprintf("ConstHost_%s", t.Name())))
	fields := def.Get("fields").(*DynamicObject)
	fields.Set("n", task.dc.tr.Get("Int"))
	constants := def.Get("constants").(*DynamicObject)
	constants.Set("defaultN", Int(18))
	constants.Set("label", String("anon"))
	typ := CreateTypeFromDef(task.dc, def)

	pm := typ.GetClassPropertyMetadata("defaultN")
	if pm == nil || pm.set != nil {
		t.Fatal("defaultN should be readonly class const")
	}
	if got := pm.get(typ); got != Int(18) {
		t.Fatalf("defaultN = %v, want 18", got)
	}
	if got := typ.GetClassPropertyMetadata("label").get(typ); got != String("anon") {
		t.Fatalf("label = %v, want anon", got)
	}
}

func TestTypedObjectConstantFieldNameClashPanics(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	def := NewTypeDef()
	def.Set("name", String(fmt.Sprintf("Clash_%s", t.Name())))
	fields := def.Get("fields").(*DynamicObject)
	fields.Set("age", task.dc.tr.Get("Int"))
	constants := def.Get("constants").(*DynamicObject)
	constants.Set("age", Int(1))
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected clash panic")
		}
		if !strings.Contains(fmt.Sprint(r), "clashes") {
			t.Fatalf("panic = %v, want clashes", r)
		}
	}()
	_ = CreateTypeFromDef(task.dc, def)
}

func TestTypedObjectOpAtIndexByFieldOrder(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmt.Sprintf("OrdGet_%s", t.Name()), "alpha", "beta", "gamma")
	inst := MakeNewTypedObject(typ)
	typedObjectPropertySet(t, inst, "alpha", String("a"))
	typedObjectPropertySet(t, inst, "beta", Int(2))
	typedObjectPropertySet(t, inst, "gamma", Bool(true))

	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, inst)
	rp.DataStack.Push(Int(0))
	RunOps(NewTestTask(), rp, OpAtIndex, OpStop)
	if got := rp.DataStack.Pop(); got != String("a") {
		t.Fatalf("@() 0 = %v, want a", got)
	}
	if rp.DataStack.Pop() != inst {
		t.Fatal("receiver not preserved after @()")
	}

	rp = NewTestRunProvider()
	pushTypedOnContextStack(rp, inst)
	rp.DataStack.Push(Int(1))
	RunOps(NewTestTask(), rp, OpAtIndex, OpStop)
	if got := rp.DataStack.Pop(); got != Int(2) {
		t.Fatalf("@() 1 = %v, want 2", got)
	}
}

func TestTypedObjectOpBangIndexByFieldOrder(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmt.Sprintf("OrdSet_%s", t.Name()), "alpha", "beta")
	inst := MakeNewTypedObject(typ)

	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, inst)
	rp.DataStack.Push(String("hi"))
	rp.DataStack.Push(Int(0))
	RunOps(NewTestTask(), rp, OpBangIndex, OpStop)
	if rp.DataStack.Pop() != inst {
		t.Fatal("receiver not preserved after !()")
	}
	if inst.GetFieldByOrdinal(0) != String("hi") {
		t.Fatalf("alpha after !() = %v, want hi", inst.GetFieldByOrdinal(0))
	}

	rp = NewTestRunProvider()
	pushTypedOnContextStack(rp, inst)
	rp.DataStack.Push(Int(9))
	rp.DataStack.Push(Int(1))
	RunOps(NewTestTask(), rp, OpBangIndex, OpStop)
	if inst.GetFieldByOrdinal(1) != Int(9) {
		t.Fatalf("beta after !() = %v, want 9", inst.GetFieldByOrdinal(1))
	}
}

func TestTypedObjectClassMethodTearOff(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	name := fmt.Sprintf("Cls_%s", t.Name())
	src := fmt.Sprintf(`
TypedObject.newTypeDef^ >ys
"%s" y:.name!
y:.fields@ >xs
Int x:.n!
x:drop
y:.constants@ >xs
5 x:.k!
x:drop
ys> TypedObject.makeType^

: %s##make
  o:.new^ >ts
  %s.k@ t:.n!
  ts>
;
`, name, name, name)
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatalf("compile: %v", err)
	}
	typ := task.dc.tr.GetLocal(name)
	if typ == nil {
		t.Fatal("type missing")
	}
	mm := typ.GetClassMethodMetadata("make")
	if mm == nil || !mm.isFinalMethod {
		t.Fatal("##make missing")
	}
	rp.DataStack.Push(typ)
	runHandleCompoundWithSegs(task, rp, []string{".", "make", "^"})
	inst, ok := rp.DataStack.Pop().(*TypedObject)
	if !ok {
		t.Fatalf("got %T, want *TypedObject", rp.DataStack.Peek())
	}
	if inst.GetFieldByOrdinal(0) != Int(5) {
		t.Fatalf("n = %v, want 5", inst.GetFieldByOrdinal(0))
	}
	if rp.OStack.Len() != 0 {
		t.Fatalf("OS len = %d", rp.OStack.Len())
	}
}
