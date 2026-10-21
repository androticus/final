package core

import (
	"strings"
	"testing"
)

func testPropertyRef(obj Typed, name string) *PropertyRef {
	pm := obj.FinalType().PropertyByName(name)
	return &PropertyRef{Parent: NewTypeObjPair(obj), PM: pm}
}

func TestIntTypeCheckCoerceAcceptsUint(t *testing.T) {
	got := GetTypeRegistrar().Get("Int").CheckCoerceType(Uint(7))
	if got != Int(7) {
		t.Fatalf("CheckCoerceType(Uint(7)) = %v, want Int(7)", got)
	}
}

func TestIntTypeCheckCoerceRejectsString(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for String on IntType")
		}
	}()
	GetTypeRegistrar().Get("Int").CheckCoerceType(String("x"))
}

func TestBoolTypeCheckCoerce(t *testing.T) {
	if got := GetTypeRegistrar().Get("Bool").CheckCoerceType(Bool(true)); got != Bool(true) {
		t.Fatalf("got %v, want true", got)
	}
}

func TestStringTypeZero(t *testing.T) {
	if got := GetTypeRegistrar().Get("String").Zero(); got != String("") {
		t.Fatalf("Zero = %q, want empty string", got)
	}
}

func TestAnyTypeCheckCoerceIdentity(t *testing.T) {
	v := Int(42)
	if got := GetTypeRegistrar().Get("Any").CheckCoerceType(v); got != v {
		t.Fatalf("Any identity failed: %v", got)
	}
}

func TestIntTypeNoClassMetadata(t *testing.T) {
	intT := GetTypeRegistrar().Get("Int")
	if intT.classMethodMetadata.Len() != 0 {
		t.Fatal("Int should have no class methods")
	}
	if intT.HasProperty("x") {
		t.Fatal("Int should have no properties")
	}
}

func TestSetItemsMetadataTwicePanics(t *testing.T) {
	t.Parallel()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for second SetItemsMetadata")
		}
	}()
	probe := &Type{name: "Probe"}
	probe.SetItemsMetadata(&ItemsMetadata{})
	probe.SetItemsMetadata(&ItemsMetadata{})
}

func TestWordTypeGlobalIdempotent(t *testing.T) {
	wordT := GetTypeRegistrar().Get("Word")
	if wordT == nil || wordT.Name() != "Word" {
		t.Fatalf("Word type = %v", wordT)
	}
	if wordT != GetTypeRegistrar().Get("Word") {
		t.Fatal("Word catalog lookup not stable")
	}
}

func TestWordFinalType(t *testing.T) {
	w := NewWord("probe", WordFlagColon)
	if w.FinalType() != GetTypeRegistrar().Get("Word") {
		t.Fatal("Word.FinalType should return registered Word type")
	}
}

func TestPropertyRefUnknownPropertyPanics(t *testing.T) {
	w := NewWord("alpha", 0)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for unknown Word property")
		}
		msg, ok := r.(error)
		if !ok {
			t.Fatalf("panic type %T", r)
		}
		if !strings.Contains(msg.Error(), "unknown property") || !strings.Contains(msg.Error(), "nosuch") {
			t.Fatalf("panic = %q, want unknown property nosuch", msg.Error())
		}
	}()
	testPropertyRef(w, "nosuch")
}

func TestPropertyRefFetchName(t *testing.T) {
	w := NewWord("alpha", 0)
	ref := testPropertyRef(w, "name")
	got, ok := ref.Fetch().(String)
	if !ok || got != "alpha" {
		t.Fatalf("name = %q, want alpha", got)
	}
}

func TestPropertyRefStoreName(t *testing.T) {
	w := NewWord("old", 0)
	ref := testPropertyRef(w, "name")
	ref.Store(String("new"))
	if w.name != "new" {
		t.Fatalf("w.name = %q, want new", w.name)
	}
}

func TestPropertyRefStoreDtIdx(t *testing.T) {
	w := NewWord("probe", 0)
	ref := testPropertyRef(w, "dtIdx")
	ref.Store(Int(9))
	if w.dtIdx != 9 {
		t.Fatalf("dtIdx = %d, want 9", w.dtIdx)
	}
}

func TestWordTypePropertyOrder(t *testing.T) {
	wordT := GetTypeRegistrar().Get("Word")
	want := []string{
		"name", "ordIdx", "flags", "opCode", "dtIdx",
		"colonCode", "doc",
		"isAxiom", "isColon", "isDefined", "isType", "xt",
	}
	if got := wordT.propertyMetadata.Len(); got != len(want) {
		t.Fatalf("property count = %d, want %d", got, len(want))
	}
	for i, name := range want {
		if got := wordT.propertyMetadata.ItemAt(i).Name; got != name {
			t.Fatalf("ordinal %d = %q, want %q", i, got, name)
		}
	}
}

func TestPropertyRefIsAxiomFetch(t *testing.T) {
	axiom := NewAxiomWord("plus", OpAdd, 0)
	ref := testPropertyRef(axiom, "isAxiom")
	if got := ref.Fetch().(Bool); !got {
		t.Fatalf("isAxiom for axiom word = %v, want true", got)
	}
	compiled := NewWord("double", WordFlagColon)
	ref2 := testPropertyRef(compiled, "isAxiom")
	if got := ref2.Fetch().(Bool); got {
		t.Fatalf("isAxiom for colon word = %v, want false", got)
	}
}

func TestPropertyRefIsDefinedFetch(t *testing.T) {
	colon := NewWord("plain", WordFlagColon)
	ref := testPropertyRef(colon, "isDefined")
	if got := ref.Fetch().(Bool); got {
		t.Fatalf("isDefined for plain colon = %v, want false", got)
	}
	defined := NewWord("myvar", WordFlagColon|WordFlagDefined)
	ref2 := testPropertyRef(defined, "isDefined")
	if got := ref2.Fetch().(Bool); !got {
		t.Fatalf("isDefined for definer instance = %v, want true", got)
	}
}

func TestPropertyRefIsAxiomStorePanics(t *testing.T) {
	w := NewWord("probe", 0)
	ref := testPropertyRef(w, "isAxiom")
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic storing read-only isAxiom")
		}
		msg, ok := r.(error)
		if !ok {
			t.Fatalf("panic type %T", r)
		}
		if !strings.Contains(msg.Error(), "read-only") || !strings.Contains(msg.Error(), "isAxiom") {
			t.Fatalf("panic = %q, want read-only isAxiom message", msg.Error())
		}
	}()
	ref.Store(Bool(true))
}

func TestOpFetchPropertyRefName(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	w := NewWord("hello", 0)
	ths := NewTypeObjPair(w)
	pm := GetTypeRegistrar().Get("Word").PropertyByName("name")
	rp.DataStack.Push(w)
	rp.DataStack.Push(&PropertyRef{Parent: ths, PM: pm})
	RunOps(NewTestTask(), rp, OpFetch, OpStop)
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want word + value", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "hello" {
		t.Fatalf("fetch name = %q, want hello", got)
	}
	if _, ok := rp.DataStack.Peek().(*Word); !ok {
		t.Fatalf("receiver should remain, TOS = %T", rp.DataStack.Peek())
	}
}

func TestOpStorePropertyRefName(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	w := NewWord("old", 0)
	ths := NewTypeObjPair(w)
	pm := GetTypeRegistrar().Get("Word").PropertyByName("name")
	rp.DataStack.Push(w)
	rp.DataStack.Push(String("fresh"))
	rp.DataStack.Push(&PropertyRef{Parent: ths, PM: pm})
	RunOps(NewTestTask(), rp, OpStore, OpStop)
	if w.name != "fresh" {
		t.Fatalf("w.name = %q, want fresh", w.name)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want word only", rp.DataStack.Len())
	}
	if _, ok := rp.DataStack.Peek().(*Word); !ok {
		t.Fatalf("receiver should remain, TOS = %T", rp.DataStack.Peek())
	}
}

func TestWordPropertyRefFetchName(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	w := NewWord("layer1", 0)
	ths := NewTypeObjPair(w)
	pm := GetTypeRegistrar().Get("Word").PropertyByName("name")
	rp.DataStack.Push(w)
	rp.DataStack.Push(&PropertyRef{Parent: ths, PM: pm})
	RunOps(NewTestTask(), rp, OpFetch, OpStop)
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "layer1" {
		t.Fatalf("property name = %q", got)
	}
}

func TestListTypeLenPropertyReadOnly(t *testing.T) {
	l := NewAnyList()
	l.Append(Int(1))
	ref := testPropertyRef(l, "len")
	if got := ref.Fetch().(Int); got != 1 {
		t.Fatalf("len = %v, want 1", got)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic storing read-only len")
		}
	}()
	ref.Store(Int(99))
}

type nativeTestInst struct{ v Int }

func (nativeTestInst) FinalType() *Type { return GetTypeRegistrar().Get("Int") }

func TestWrapNativeSet(t *testing.T) {
	rp := NewTestRunProvider()
	var stored any
	f, ths := memberRunFrame(rp, nativeTestInst{1})
	rp.DataStack.Push(Uint(7))
	pm := &PropertyMetadata{
		name:  "v",
		_type: GetTypeRegistrar().Get("Int"),
		set: func(obj any, newValue any) {
			stored = obj.(nativeTestInst).v + newValue.(Int)
		},
	}
	pm.invokeSet(f, ths)
	if stored != Int(8) {
		t.Fatalf("stored = %v, want Int(8)", stored)
	}
}

func TestWrapNativeMethod(t *testing.T) {
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, nativeTestInst{10})
	rp.DataStack.Push(Int(5))
	wrapNativeMethod(func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		ds.Push(ths.Obj.(nativeTestInst).v + ds.Pop().(Int))
	})(f, ths)
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 15 {
		t.Fatalf("DS = %v, want Int(15)", got)
	}
}

func TestWordNativePropertyGetViaRunFrame(t *testing.T) {
	rp := NewTestRunProvider()
	w := NewWord("hello", 0)
	pm := GetTypeRegistrar().Get("Word").PropertyByName("name")
	f, ths := memberRunFrame(rp, w)
	pm.invokeGet(f, ths)
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "hello" {
		t.Fatalf("name = %q, want hello", got)
	}
}

func TestWordNativePropertySetViaRunFrame(t *testing.T) {
	rp := NewTestRunProvider()
	w := NewWord("old", 0)
	pm := GetTypeRegistrar().Get("Word").PropertyByName("name")
	rp.DataStack.Push(String("fresh"))
	f, ths := memberRunFrame(rp, w)
	pm.invokeSet(f, ths)
	if w.name != "fresh" {
		t.Fatalf("w.name = %q, want fresh", w.name)
	}
}
