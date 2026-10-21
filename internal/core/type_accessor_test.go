package core

import "testing"

func TestMethodMetadata_NameAndAccessors(t *testing.T) {
	typ := GetTypeRegistrar().Get("Pixmap")
	if typ == nil {
		t.Fatal("Pixmap type missing")
	}
	mm := typ.GetMethodMetadata("clear")
	if mm == nil {
		t.Fatal("Pixmap#clear missing")
	}
	if mm.Name() != "clear" {
		t.Fatalf("Name = %q", mm.Name())
	}
	if mm.IsFinalMethod() {
		t.Fatal("clear should be native")
	}
	_ = mm.Doc()
	_ = mm.MixinType()
}

func TestPropertyMetadata_TypeAccessor(t *testing.T) {
	typ := GetTypeRegistrar().Get("Pixmap")
	pm := typ.GetPropertyMetadata("rows")
	if pm == nil {
		t.Fatal("Pixmap#rows missing")
	}
	if pm.PropertyName() != "rows" {
		t.Fatalf("PropertyName = %q", pm.PropertyName())
	}
	if pm.Type() == nil || pm.Type().Name() != "Int" {
		t.Fatalf("Type = %v", pm.Type())
	}
}

func TestNewMethodRef(t *testing.T) {
	pm := NewPixmap(2, 2)
	ref := NewMethodRef(pm, "clear")
	if ref == nil || ref.MM == nil || ref.Parent == nil {
		t.Fatal("nil ref")
	}
	if ref.MM.Name() != "clear" {
		t.Fatalf("bound %q", ref.MM.Name())
	}
	if ref.Parent.Obj != pm {
		t.Fatal("Parent.Obj mismatch")
	}
}

func TestNewMethodRef_MissingPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = NewMethodRef(NewPixmap(1, 1), "noSuchMethod")
}
