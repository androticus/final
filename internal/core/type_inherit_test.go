package core

import "testing"

func TestGetParentImplicitProtoType(t *testing.T) {
	if protoTypeType == nil {
		t.Fatal("protoTypeType not registered")
	}
	intT := GetTypeRegistrar().Get("Int")
	if intT == nil {
		t.Fatal("Int type not registered")
	}
	if intT.parent != nil {
		t.Fatalf("Int.parent = %v, want nil (implicit ProtoType)", intT.parent)
	}
	if got := intT.GetParent(); got != protoTypeType {
		t.Fatalf("Int.GetParent() = %v, want ProtoType", got)
	}
	if got := protoTypeType.GetParent(); got != nil {
		t.Fatalf("ProtoType.GetParent() = %v, want nil", got)
	}
}

func TestGetClassMethodMetadataInheritsFromParent(t *testing.T) {
	parent := &Type{name: "InheritClassParent"}
	want := &MethodMetadata{name: "factory", call: func(*RunFrame, *TypeObjPair) {}}
	parent.AddClassMethod("factory", want)
	child := &Type{name: "InheritClassChild"}
	child.SetParent(parent)
	if !child.HasClassMethod("factory") {
		t.Fatal("child should inherit class method factory from parent")
	}
	mm := child.GetClassMethodMetadata("factory")
	if mm != want {
		t.Fatal("expected same MethodMetadata as on parent")
	}
}

func TestSetParentCyclePanics(t *testing.T) {
	a := &Type{name: "A"}
	b := &Type{name: "B"}
	a.SetParent(b)
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for SetParent cycle")
		}
	}()
	b.SetParent(a)
}

func TestGetParentExplicitChain(t *testing.T) {
	root := &Type{name: "Root"}
	child := &Type{name: "Child"}
	leaf := &Type{name: "Leaf"}
	child.SetParent(root)
	leaf.SetParent(child)

	if leaf.GetParent() != child {
		t.Fatalf("Leaf.GetParent() = %v, want Child", leaf.GetParent())
	}
	if child.GetParent() != root {
		t.Fatalf("Child.GetParent() = %v, want Root", child.GetParent())
	}
	if root.GetParent() != protoTypeType {
		t.Fatalf("Root.GetParent() = %v, want ProtoType", root.GetParent())
	}
}

func TestGetItemsMetadataInheritsAllOrNothing(t *testing.T) {
	parent := &Type{name: "Parent"}
	child := &Type{name: "Child"}
	child.SetParent(parent)
	parent.SetItemsMetadata(&ItemsMetadata{
		getByOrdinal: func(inst Any, idx Any) any { return idx },
	})
	im := child.GetItemsMetadata()
	if im == nil || im.getByOrdinal == nil {
		t.Fatal("child should inherit parent itemsMetadata")
	}
}

func TestGetMethodMetadataWalksExplicitParent(t *testing.T) {
	parent := &Type{name: "Parent"}
	child := &Type{name: "Child"}
	child.SetParent(parent)
	parent.AddMethod("m", &MethodMetadata{name: "m", call: func(f *RunFrame, ths *TypeObjPair) {}})

	mm := child.GetMethodMetadata("m")
	if mm == nil {
		t.Fatal("child should inherit parent method m")
	}
}

func TestTypeGetPropertyMethod(t *testing.T) {
	wordT := GetTypeRegistrar().Get("Word")
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, wordT)
	mm := ths.Typ.GetMethodMetadata("getProperty")
	if mm == nil {
		t.Fatal("missing getProperty on Type chain")
	}
	rp.DataStack.Push(String("name"))
	mm.invokeCall(f, ths)
	if !rp.DataStack.Pop().(Bool) {
		t.Fatal("want true for Word.name")
	}
	pm := rp.DataStack.Pop().(*PropertyMetadata)
	if pm.name != "name" {
		t.Fatalf("property name = %q, want name", pm.name)
	}

	rp.DataStack.Push(String("nosuch"))
	mm.invokeCall(f, ths)
	if rp.DataStack.Pop().(Bool) {
		t.Fatal("want false for missing property")
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0 after false", rp.DataStack.Len())
	}
}

func TestTypeGetMethodMethod(t *testing.T) {
	stringT := GetTypeRegistrar().Get("String")
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, stringT)
	mm := ths.Typ.GetMethodMetadata("getMethod")
	if mm == nil {
		t.Fatal("missing getMethod on Type chain")
	}
	rp.DataStack.Push(String("_toString"))
	mm.invokeCall(f, ths)
	if !rp.DataStack.Pop().(Bool) {
		t.Fatal("want true for String._toString")
	}
	got := rp.DataStack.Pop().(*MethodMetadata)
	if got.name != "_toString" {
		t.Fatalf("method name = %q, want _toString", got.name)
	}
}

func TestTypeTryGetItemsMetadataMethod(t *testing.T) {
	listT := GetTypeRegistrar().Get("List")
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, listT)
	mm := ths.Typ.GetMethodMetadata("tryGetItemsMetadata")
	if mm == nil {
		t.Fatal("missing tryGetItemsMetadata")
	}
	mm.invokeCall(f, ths)
	if !rp.DataStack.Pop().(Bool) {
		t.Fatal("want true for List items")
	}
	if _, ok := rp.DataStack.Pop().(*ItemsMetadata); !ok {
		t.Fatal("want *ItemsMetadata")
	}

	intT := GetTypeRegistrar().Get("Int")
	f, ths = memberRunFrame(rp, intT)
	mm.invokeCall(f, ths)
	if rp.DataStack.Pop().(Bool) {
		t.Fatal("want false for Int items")
	}
}

func TestTypeGetParentTypeMethod(t *testing.T) {
	intT := GetTypeRegistrar().Get("Int")
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, intT)
	mm := ths.Typ.GetMethodMetadata("getParentType")
	if mm == nil {
		t.Fatal("missing getParentType")
	}
	mm.invokeCall(f, ths)
	if !rp.DataStack.Pop().(Bool) {
		t.Fatal("Int.getParentType want true")
	}
	parent := rp.DataStack.Pop().(*Type)
	if parent != protoTypeType {
		t.Fatalf("Int parent = %v, want ProtoType", parent)
	}

	f, ths = memberRunFrame(rp, protoTypeType)
	mm.invokeCall(f, ths)
	if rp.DataStack.Pop().(Bool) {
		t.Fatal("ProtoType.getParentType want false")
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d after false", rp.DataStack.Len())
	}
}

func TestTypeHasNoParentProperty(t *testing.T) {
	if typeType.GetPropertyMetadata("parent") != nil {
		t.Fatal("Type must not expose parent property; use getParentType")
	}
}

func TestTypeHasNoMetadataProperties(t *testing.T) {
	for _, name := range []string{
		"propertyMetadata", "methodMetadata", "itemsMetadata",
		"classPropertyMetadata", "classMethodMetadata",
	} {
		if typeType.GetPropertyMetadata(name) != nil {
			t.Fatalf("Type must not expose %q property; use tryGet*", name)
		}
	}
}

func TestTypeTryGetPropertyMetadataMethod(t *testing.T) {
	// Word carries property metadata; String no longer does (its members are
	// now str:: operators), so it behaves like the zero-property case below.
	wordT := GetTypeRegistrar().Get("Word")
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, wordT)
	mm := ths.Typ.GetMethodMetadata("tryGetPropertyMetadata")
	if mm == nil {
		t.Fatal("missing tryGetPropertyMetadata")
	}
	mm.invokeCall(f, ths)
	n := rp.DataStack.Pop().(Int)
	if n <= 0 {
		t.Fatalf("Word property count = %v, want > 0", n)
	}
	if _, ok := rp.DataStack.Pop().(*PropertyMetadataMap); !ok {
		t.Fatal("want *PropertyMetadataMap")
	}

	intT := GetTypeRegistrar().Get("Int")
	f, ths = memberRunFrame(rp, intT)
	mm.invokeCall(f, ths)
	if got := rp.DataStack.Pop().(Int); got != 0 {
		t.Fatalf("Int tryGetPropertyMetadata = %v, want 0", got)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d after false", rp.DataStack.Len())
	}
}
