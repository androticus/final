package core

import (
	"strings"
	"testing"
	"time"
)

func TestGetPropertyNamesDynamicObject(t *testing.T) {
	o := NewDynamicObject()
	o.Set("b", Int(2))
	o.Set("a", Int(1))
	names := o.GetPropertyNames()
	if names.Len() != 2 || names.At(0) != "b" || names.At(1) != "a" {
		t.Fatalf("names = %v,%v len=%d", names.At(0), names.At(1), names.Len())
	}
}

func TestGetPropertyNamesStaticObject(t *testing.T) {
	proto := NewDynamicObject()
	proto.Set("x", Int(1))
	proto.Set("y", Int(2))
	o := NewStaticObject(proto)
	names := o.GetPropertyNames()
	if names.Len() != 2 || names.At(0) != "x" || names.At(1) != "y" {
		t.Fatalf("names = %v", []string{names.At(0), names.At(1)})
	}
}

func TestGetPropertyNamesTypedObjectAndType(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmtSprintfUnique(t, "PropNames"), "alpha", "beta")
	inst := MakeNewTypedObject(typ)
	names := inst.GetPropertyNames()
	if names.Len() != 2 || names.At(0) != "alpha" || names.At(1) != "beta" {
		t.Fatalf("inst names = %v,%v", names.At(0), names.At(1))
	}
	tn := typ.GetPropertyNames()
	if tn.Len() != 2 || tn.At(0) != "alpha" || tn.At(1) != "beta" {
		t.Fatalf("type names = %v,%v", tn.At(0), tn.At(1))
	}
	if dynamicObjectType.GetPropertyNames().Len() != 0 {
		t.Fatal("DynamicObject type facet should have empty GetPropertyNames")
	}
}

func fmtSprintfUnique(t *testing.T, prefix string) string {
	t.Helper()
	return prefix + "_" + strings.ReplaceAll(t.Name(), "/", "_")
}

func TestTableFromDynamicObjectSortOrdinal(t *testing.T) {
	ResetTestRuntime()
	proto := NewDynamicObject()
	proto.Set("name", String(""))
	proto.Set("score", Int(0))
	tab := NewTableFromDynamicObject(proto)
	if !tab.canAccessOrdinally {
		t.Fatal("expected ordinal access")
	}

	r1 := NewStaticObject(proto)
	r1.Set("name", String("b"))
	r1.Set("score", Int(10))
	r2 := proto.Clone()
	r2.Set("name", String("a"))
	r2.Set("score", Int(20))
	r3 := NewStaticObject(proto)
	r3.Set("name", String("c"))
	r3.Set("score", Int(5))
	tab.Append(r1)
	tab.Append(r2)
	tab.Append(r3)

	tab.SortByOrd(1, false) // score
	if tab.At(0).(*StaticObject).Get("score") != Int(5) {
		t.Fatalf("after score sort first=%v", tab.At(0))
	}
	tab.SortByOrd(0, false) // name (stable preserves score order within ties — here unique names)
	got := []string{
		string(rowName(tab.At(0))),
		string(rowName(tab.At(1))),
		string(rowName(tab.At(2))),
	}
	if got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("after name sort = %v", got)
	}
}

func rowName(row any) String {
	switch o := row.(type) {
	case *DynamicObject:
		return o.Get("name").(String)
	case *StaticObject:
		return o.Get("name").(String)
	default:
		panic("bad row")
	}
}

func TestTableFromColumnNamesSortByName(t *testing.T) {
	ResetTestRuntime()
	cols := NewStringListFromSlice([]string{"k", "v"})
	tab := NewTableFromColumnNames(cols)
	if tab.canAccessOrdinally {
		t.Fatal("column-names factory leaves canAccessOrdinally false (resolve at Sort)")
	}
	a := NewDynamicObject()
	a.Set("k", String("z"))
	a.Set("v", Int(1))
	b := NewDynamicObject()
	b.Set("k", String("a"))
	b.Set("v", Int(2))
	tab.Append(a)
	tab.Append(b)
	tab.SortByOrd(0, false)
	if tab.At(0).(*DynamicObject).Get("k") != String("a") {
		t.Fatalf("first k = %v", tab.At(0).(*DynamicObject).Get("k"))
	}
}

func TestTableFromColumnNamesListRowsOrdinal(t *testing.T) {
	ResetTestRuntime()
	cols := NewStringListFromSlice([]string{"name", "score"})
	tab := NewTableFromColumnNames(cols)
	r1 := NewAnyList()
	r1.Append(String("b"))
	r1.Append(Int(10))
	r2 := NewAnyList()
	r2.Append(String("a"))
	r2.Append(Int(20))
	tab.Append(r1)
	tab.Append(r2)
	tab.SortByOrd(0, false)
	if tab.At(0).(*AnyList).At(0) != String("a") {
		t.Fatalf("first name = %v", tab.At(0).(*AnyList).At(0))
	}
}

func TestTableFromColumnNamesListRowsSortIntThenString(t *testing.T) {
	ResetTestRuntime()
	cols := NewStringListFromSlice([]string{"IntValue", "StringValue"})
	tab := NewTableFromColumnNames(cols)
	r1 := NewAnyList()
	r1.Append(Int(2))
	r1.Append(String("b"))
	r2 := NewAnyList()
	r2.Append(Int(1))
	r2.Append(String("a"))
	tab.Append(r1)
	tab.Append(r2)

	tab.SortByOrd(0, false) // IntValue
	if tab.At(0).(*AnyList).At(0) != Int(1) || tab.At(1).(*AnyList).At(0) != Int(2) {
		t.Fatalf("after Sort(0): got %v, %v", tab.At(0).(*AnyList).At(0), tab.At(1).(*AnyList).At(0))
	}
	tab.SortByOrd(1, false) // StringValue (stable; primary becomes StringValue)
	if tab.At(0).(*AnyList).At(1) != String("a") || tab.At(1).(*AnyList).At(1) != String("b") {
		t.Fatalf("after Sort(1): got %v, %v", tab.At(0).(*AnyList).At(1), tab.At(1).(*AnyList).At(1))
	}
}

func TestTableFromColumnNamesMapRowsByName(t *testing.T) {
	ResetTestRuntime()
	cols := NewStringListFromSlice([]string{"k", "v"})
	tab := NewTableFromColumnNames(cols)
	a := NewStringMap()
	a.Put(String("k"), String("z"))
	a.Put(String("v"), Int(1))
	b := NewStringMap()
	b.Put(String("k"), String("a"))
	b.Put(String("v"), Int(2))
	tab.Append(a)
	tab.Append(b)
	tab.SortByOrd(0, false)
	if tab.At(0).(*StringMap).Get(String("k")) != String("a") {
		t.Fatalf("first k = %v", tab.At(0).(*StringMap).Get(String("k")))
	}
}

func TestTableFromColumnNamesStringMapRowsSortIntThenString(t *testing.T) {
	ResetTestRuntime()
	cols := NewStringListFromSlice([]string{"IntValue", "StringValue"})
	tab := NewTableFromColumnNames(cols)
	a := NewStringMap()
	a.Put(String("IntValue"), Int(2))
	a.Put(String("StringValue"), String("b"))
	b := NewStringMap()
	b.Put(String("IntValue"), Int(1))
	b.Put(String("StringValue"), String("a"))
	tab.Append(a)
	tab.Append(b)

	tab.SortByOrd(0, false) // IntValue
	if tab.At(0).(*StringMap).Get(String("IntValue")) != Int(1) ||
		tab.At(1).(*StringMap).Get(String("IntValue")) != Int(2) {
		t.Fatalf("after Sort(0): got %v, %v",
			tab.At(0).(*StringMap).Get(String("IntValue")),
			tab.At(1).(*StringMap).Get(String("IntValue")))
	}
	tab.SortByOrd(1, false) // StringValue
	if tab.At(0).(*StringMap).Get(String("StringValue")) != String("a") ||
		tab.At(1).(*StringMap).Get(String("StringValue")) != String("b") {
		t.Fatalf("after Sort(1): got %v, %v",
			tab.At(0).(*StringMap).Get(String("StringValue")),
			tab.At(1).(*StringMap).Get(String("StringValue")))
	}
}

func TestTableSortTypeMismatchPanics(t *testing.T) {
	proto := NewDynamicObject()
	proto.Set("v", Int(0))
	tab := NewTableFromDynamicObject(proto)
	r1 := NewStaticObject(proto)
	r1.Set("v", Int(1))
	r2 := NewStaticObject(proto)
	r2.Set("v", String("x"))
	tab.Append(r1)
	tab.Append(r2)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		msg := r.(error).Error()
		if !strings.Contains(msg, "mismatch") {
			t.Fatalf("panic = %q", msg)
		}
	}()
	tab.SortByOrd(0, false)
}

func TestTableSortUnsupportedTypePanics(t *testing.T) {
	proto := NewDynamicObject()
	proto.Set("v", nil)
	tab := NewTableFromDynamicObject(proto)
	r := NewStaticObject(proto)
	r.Set("v", NewDynamicObject())
	tab.Append(r)
	tab.Append(NewStaticObject(proto))
	tab.At(1).(*StaticObject).Set("v", NewDynamicObject())
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if !strings.Contains(r.(error).Error(), "unsupported") {
			t.Fatalf("panic = %v", r)
		}
	}()
	tab.SortByOrd(0, false)
}

func TestTableSortDateTime(t *testing.T) {
	proto := NewDynamicObject()
	proto.Set("t", time.Time{})
	tab := NewTableFromDynamicObject(proto)
	early := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	late := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	r1 := NewStaticObject(proto)
	r1.Set("t", late)
	r2 := NewStaticObject(proto)
	r2.Set("t", early)
	tab.Append(r1)
	tab.Append(r2)
	tab.SortByOrd(0, false)
	if !tab.At(0).(*StaticObject).Get("t").(time.Time).Equal(early) {
		t.Fatal("expected earlier DateTime first")
	}
}

func TestTableFromTypeOrdinalFlag(t *testing.T) {
	ResetTestRuntime()
	// DynamicObject type has @() via items metadata
	tab := NewTableFromType(dynamicObjectType)
	if !tab.canAccessOrdinally {
		t.Fatal("DynamicObject type should set canAccessOrdinally")
	}
	// Word has properties but typically no @() items — flag false
	wordT := GetTypeRegistrar().Get("Word")
	tab2 := NewTableFromType(wordT)
	if tab2.canAccessOrdinally {
		t.Fatal("Word should not have ordinal @() for Table")
	}
	if tab2.columnNames.Len() == 0 {
		t.Fatal("Word should expose property names")
	}
}

func TestTableFromTypedObject(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmtSprintfUnique(t, "TOTable"), "alpha", "beta")
	tab := NewTableFromTypedObject(typ)
	if !tab.canAccessOrdinally {
		t.Fatal("expected ordinal access")
	}
	if tab.columnNames.Len() != 2 || tab.columnNames.At(0) != "alpha" || tab.columnNames.At(1) != "beta" {
		t.Fatalf("columns = %v,%v len=%d", tab.columnNames.At(0), tab.columnNames.At(1), tab.columnNames.Len())
	}

	r1 := MakeNewTypedObject(typ)
	r1.SetFieldByOrdinal(0, String("b"))
	r1.SetFieldByOrdinal(1, Int(10))
	r2 := MakeNewTypedObject(typ)
	r2.SetFieldByOrdinal(0, String("a"))
	r2.SetFieldByOrdinal(1, Int(20))
	tab.Append(r1)
	tab.Append(r2)
	tab.SortByOrd(0, false)
	if tab.At(0).(*TypedObject).GetFieldByOrdinal(0) != String("a") {
		t.Fatalf("first alpha = %v", tab.At(0).(*TypedObject).GetFieldByOrdinal(0))
	}
}

func TestTableFromTypedObjectNonTOParentPanics(t *testing.T) {
	ResetTestRuntime()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		msg := r.(error).Error()
		if !strings.Contains(msg, "parent is not TypedObject") {
			t.Fatalf("panic = %q", msg)
		}
	}()
	_ = NewTableFromTypedObject(dynamicObjectType)
}

func TestTableAppendAtIterAndFactoriesRegistered(t *testing.T) {
	ResetTestRuntime()
	proto := NewDynamicObject()
	proto.Set("n", Int(0))
	tab := NewTableFromDynamicObject(proto)
	tab.Append(NewStaticObject(proto))
	if tab.Len() != 1 || tab.At(0) == nil {
		t.Fatal("append/at")
	}
	it := tab.GetIter()
	if !it.GetNext() || it.Current() == nil {
		t.Fatal("iter")
	}
	for _, name := range []string{
		"newFromColumnNames", "newFromDynamicObject", "newFromTypedObject", "newFromType",
	} {
		if tableType.GetClassMethodMetadata(name) == nil {
			t.Fatalf("missing class method %s", name)
		}
	}
	for _, name := range []string{"append", "at", "sortByOrd", "sortByName"} {
		if tableType.GetMethodMetadata(name) == nil {
			t.Fatalf("missing method %s", name)
		}
	}
	for _, name := range []string{"SortAsc", "SortDesc"} {
		if tableType.GetClassPropertyMetadata(name) == nil {
			t.Fatalf("missing class property %s", name)
		}
	}
}
