package core

import (
	"slices"
	"strings"
	"testing"
)

func listExpectPanic(t *testing.T, wantSubstr string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		msg := ""
		if err, ok := r.(error); ok {
			msg = err.Error()
		}
		if wantSubstr != "" && !strings.Contains(msg, wantSubstr) {
			t.Fatalf("panic = %v, want substring %q", r, wantSubstr)
		}
	}()
	fn()
}

func TestAnyListNewEmpty(t *testing.T) {
	l := NewAnyList()
	if l.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", l.Len())
	}
}

func TestAnyListHasItemFromAny(t *testing.T) {
	l := NewAnyList()
	if l.HasItemFromAny(Int(1)) {
		t.Fatal("empty list should not hasItem")
	}
	l.AppendFromAny(Int(10))
	l.AppendFromAny(Int(20))
	l.AppendFromAny(String("x"))
	if !l.HasItemFromAny(Int(20)) {
		t.Fatal("want hasItem 20")
	}
	if !l.HasItemFromAny(String("x")) {
		t.Fatal("want hasItem \"x\"")
	}
	if l.HasItemFromAny(Int(99)) {
		t.Fatal("must not hasItem 99")
	}
}

func TestIntListHasItemFromAny(t *testing.T) {
	l := NewIntList()
	l.AppendFromAny(Int(12))
	l.AppendFromAny(Int(5))
	if !l.HasItemFromAny(Int(12)) {
		t.Fatal("want hasItem 12")
	}
	if l.HasItemFromAny(Int(7)) {
		t.Fatal("must not hasItem 7")
	}
}

func TestAnyListDeleteItemFromAny(t *testing.T) {
	l := NewAnyList()
	if l.DeleteItemFromAny(Int(1)) {
		t.Fatal("empty list should not deleteItem")
	}
	l.AppendFromAny(Int(10))
	l.AppendFromAny(Int(20))
	l.AppendFromAny(Int(10))
	if !l.DeleteItemFromAny(Int(10)) {
		t.Fatal("want deleteItem 10")
	}
	if l.Len() != 2 || l.At(0) != Int(20) || l.At(1) != Int(10) {
		t.Fatalf("after first deleteItem: [%v %v], want [20 10]", l.At(0), l.At(1))
	}
	if l.DeleteItemFromAny(Int(99)) {
		t.Fatal("must not deleteItem 99")
	}
	if l.Len() != 2 {
		t.Fatalf("absent deleteItem changed len to %d", l.Len())
	}
}

func TestAnyListDeleteItemFromAnyPointerIdentity(t *testing.T) {
	type rec struct{ n int }
	a := &rec{1}
	b := &rec{1}
	l := NewAnyList()
	l.Append(a)
	l.Append(b)
	if !l.DeleteItemFromAny(a) {
		t.Fatal("want deleteItem pointer a")
	}
	if l.Len() != 1 || l.At(0) != b {
		t.Fatalf("after deleteItem a: len=%d at0=%v, want b", l.Len(), l.At(0))
	}
	if l.HasItemFromAny(a) {
		t.Fatal("a should be gone")
	}
	if !l.HasItemFromAny(b) {
		t.Fatal("b should remain")
	}
}

func TestAnyListDeleteItemDuringSnapshotIter(t *testing.T) {
	type rec struct{ n int }
	a := &rec{1}
	b := &rec{2}
	l := NewAnyList()
	l.Append(a)
	l.Append(b)
	it := l.GetIter()
	var removed int
	for it.GetNext() {
		if l.DeleteItemFromAny(it.Current()) {
			removed++
		}
	}
	if removed != 2 || l.Len() != 0 {
		t.Fatalf("snapshot deleteItem: removed=%d len=%d, want 2 / 0", removed, l.Len())
	}
}

func TestIntListDeleteItemFromAny(t *testing.T) {
	l := NewIntList()
	l.AppendFromAny(Int(12))
	l.AppendFromAny(Int(5))
	if !l.DeleteItemFromAny(Int(12)) {
		t.Fatal("want deleteItem 12")
	}
	if l.Len() != 1 || l.At(0) != 5 {
		t.Fatalf("after deleteItem: [%v], want [5]", l.At(0))
	}
}

func TestAnyListPutFromAny(t *testing.T) {
	l := NewAnyList()
	l.AppendFromAny(Int(10))
	l.PutFromAny(0, Int(77))
	if l.At(0) != Int(77) {
		t.Fatalf("At(0) = %v, want 77", l.At(0))
	}
}

func TestAnyListInsertAt_middle(t *testing.T) {
	l := NewAnyList()
	l.Append(Int(1))
	l.Append(Int(3))
	l.InsertAt(1, Int(2))
	if l.Len() != 3 || l.At(1) != Int(2) {
		t.Fatalf("InsertAt middle failed: %v %v %v", l.At(0), l.At(1), l.At(2))
	}
}

func TestAnyListInsertAtFromAny_appendViaLen(t *testing.T) {
	l := NewAnyList()
	l.AppendFromAny(Int(1))
	l.InsertAtFromAny(l.Len(), Int(2))
	if l.Len() != 2 || l.At(0) != Int(1) || l.At(1) != Int(2) {
		t.Fatalf("list = [%v %v], want [1 2]", l.At(0), l.At(1))
	}
}

func TestAnyListDeleteAt(t *testing.T) {
	l := NewAnyList()
	l.AppendFromAny(Int(1))
	l.AppendFromAny(Int(2))
	l.DeleteAt(0)
	if l.Len() != 1 || l.At(0) != Int(2) {
		t.Fatalf("after delete-at 0: Len=%d At(0)=%v", l.Len(), l.At(0))
	}
}

func TestAnyListClear(t *testing.T) {
	l := NewAnyList()
	l.AppendFromAny(Int(1))
	l.Clear()
	if l.Len() != 0 {
		t.Fatalf("Len() = %d, want 0 after clear", l.Len())
	}
}

func TestAnyListAt_outOfRangePanics(t *testing.T) {
	l := NewAnyList()
	listExpectPanic(t, "out of range", func() { l.At(0) })
}

func TestAnyListDeleteAt_outOfRangePanics(t *testing.T) {
	l := NewAnyList()
	listExpectPanic(t, "out of range", func() { l.DeleteAt(0) })
}

func listType(t *testing.T) *Type {
	t.Helper()
	return GetTypeRegistrar().Get("List")
}

func stringListType(t *testing.T) *Type {
	t.Helper()
	return GetTypeRegistrar().Get("StringList")
}

func invokeListMethod[T any](t *testing.T, typ *Type, l *List[T], name string, setup func(*Stack[any])) {
	t.Helper()
	rp := NewTestRunProvider()
	if setup != nil {
		setup(rp.DataStack)
	}
	mm := typ.MethodByName(name)
	f, ths := memberRunFrame(rp, l)
	mm.invokeCall(f, ths)
}

func TestListTypeClassFacetNew(t *testing.T) {
	listT := listType(t)
	if !listT.HasClassMethod("new") {
		t.Fatal("List should have class method new")
	}
}

func TestListNativeClassMethodNewViaRunFrame(t *testing.T) {
	rp := NewTestRunProvider()
	mm := listType(t).ClassMethodByName("new")
	f, ths := memberRunFrame(rp, listType(t))
	mm.invokeCall(f, ths)
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want new list on DS", rp.DataStack.Len())
	}
	l, ok := rp.DataStack.Peek().(*AnyList)
	if !ok || l.Len() != 0 {
		t.Fatalf("new list = %v, want empty *AnyList", rp.DataStack.Peek())
	}
}

func TestListNativeMethodAppendViaRunFrame(t *testing.T) {
	l := NewAnyList()
	invokeListMethod(t, listType(t), l, "append", func(ds *Stack[any]) {
		ds.Push(Int(42))
	})
	if l.Len() != 1 || l.At(0) != Int(42) {
		t.Fatalf("list = %v, want [42]", l.At(0))
	}
}

func TestAnyListStackPushPopLIFO(t *testing.T) {
	l := NewAnyList()
	l.AppendFromAny(Int(1))
	l.AppendFromAny(Int(2))
	l.AppendFromAny(Int(3))
	if got := l.Dequeue(); got != Int(3) {
		t.Fatalf("pop = %v, want 3", got)
	}
	if got := l.Dequeue(); got != Int(2) {
		t.Fatalf("pop = %v, want 2", got)
	}
	if got := l.Dequeue(); got != Int(1) {
		t.Fatalf("pop = %v, want 1", got)
	}
	if l.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", l.Len())
	}
}

func TestAnyListPeek_returnsTopWithoutRemoving(t *testing.T) {
	l := NewAnyList()
	l.AppendFromAny(Int(10))
	l.AppendFromAny(Int(20))
	if got := l.Peek(); got != Int(20) {
		t.Fatalf("Peek() = %v, want 20", got)
	}
	if l.Len() != 2 || l.At(1) != Int(20) {
		t.Fatalf("list mutated by Peek: Len=%d At(1)=%v", l.Len(), l.At(1))
	}
}

func TestAnyListPeek_emptyPanics(t *testing.T) {
	l := NewAnyList()
	listExpectPanic(t, "empty", func() { l.Peek() })
}

func TestAnyListPop_emptyPanics(t *testing.T) {
	l := NewAnyList()
	listExpectPanic(t, "empty", func() { l.Dequeue() })
}

func TestListTypeHasStackMethods(t *testing.T) {
	lt := listType(t)
	for _, name := range []string{"push", "pop", "peek", "depth"} {
		if !lt.HasMethod(name) {
			t.Fatalf("List missing instance method %q", name)
		}
	}
}

func TestListNativeMethodPushViaRunFrame(t *testing.T) {
	l := NewAnyList()
	invokeListMethod(t, listType(t), l, "push", func(ds *Stack[any]) {
		ds.Push(Int(7))
	})
	if l.Len() != 1 || l.At(0) != Int(7) {
		t.Fatalf("after push: Len=%d At(0)=%v", l.Len(), l.At(0))
	}
}

func TestListNativeMethodPopViaRunFrame(t *testing.T) {
	l := NewAnyList()
	l.AppendFromAny(Int(1))
	l.AppendFromAny(Int(2))
	rp := NewTestRunProvider()
	mm := listType(t).MethodByName("pop")
	f, ths := memberRunFrame(rp, l)
	mm.invokeCall(f, ths)
	if l.Len() != 1 || l.At(0) != Int(1) {
		t.Fatalf("after pop: Len=%d At(0)=%v, want Len=1 At(0)=1", l.Len(), l.At(0))
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 2 {
		t.Fatalf("DS pop result = %v, want 2", got)
	}
}

func TestListNativeMethodPeekViaRunFrame(t *testing.T) {
	l := NewAnyList()
	l.AppendFromAny(Int(99))
	rp := NewTestRunProvider()
	mm := listType(t).MethodByName("peek")
	f, ths := memberRunFrame(rp, l)
	mm.invokeCall(f, ths)
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 99 {
		t.Fatalf("peek on DS = %v, want 99", got)
	}
	if l.Len() != 1 {
		t.Fatalf("Len() = %d, want 1 after peek", l.Len())
	}
}

func TestListNativeMethodDepthViaRunFrame(t *testing.T) {
	l := NewAnyList()
	l.AppendFromAny(Int(1))
	l.AppendFromAny(Int(2))
	l.AppendFromAny(Int(3))
	rp := NewTestRunProvider()
	mm := listType(t).MethodByName("depth")
	f, ths := memberRunFrame(rp, l)
	mm.invokeCall(f, ths)
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 3 {
		t.Fatalf("depth = %v, want 3", got)
	}
}

func TestListNativeMethodPopEmptyPanics(t *testing.T) {
	l := NewAnyList()
	listExpectPanic(t, "empty", func() {
		invokeListMethod(t, listType(t), l, "pop", nil)
	})
}

func TestListNativeMethodPeekEmptyPanics(t *testing.T) {
	l := NewAnyList()
	listExpectPanic(t, "empty", func() {
		invokeListMethod(t, listType(t), l, "peek", nil)
	})
}

func TestListNativeMethodPushMatchesAppendViaRunFrame(t *testing.T) {
	pushL := NewAnyList()
	appendL := NewAnyList()
	lt := listType(t)
	for _, v := range []Int{1, 2, 3} {
		invokeListMethod(t, lt, pushL, "push", func(ds *Stack[any]) { ds.Push(v) })
		invokeListMethod(t, lt, appendL, "append", func(ds *Stack[any]) { ds.Push(v) })
	}
	for i := 0; i < 3; i++ {
		if pushL.At(i) != appendL.At(i) {
			t.Fatalf("At(%d): push=%v append=%v", i, pushL.At(i), appendL.At(i))
		}
	}
}

func TestListNativeMethodDepthMatchesLenProperty(t *testing.T) {
	l := NewAnyList()
	l.AppendFromAny(Int(1))
	l.AppendFromAny(Int(2))
	rp := NewTestRunProvider()
	lt := listType(t)
	mm := lt.MethodByName("depth")
	f, ths := memberRunFrame(rp, l)
	mm.invokeCall(f, ths)
	depth, ok := rp.DataStack.Pop().(Int)
	if !ok {
		t.Fatalf("depth type = %T", rp.DataStack.Peek())
	}
	pm := lt.PropertyByName("len")
	f, ths = memberRunFrame(rp, l)
	pm.invokeGet(f, ths)
	length, ok := rp.DataStack.Pop().(Int)
	if !ok {
		t.Fatalf("len type = %T", rp.DataStack.Peek())
	}
	if depth != length || depth != 2 {
		t.Fatalf("depth=%v len=%v, want both 2", depth, length)
	}
}

func TestTypeMemberDispatchListPushByName(t *testing.T) {
	ResetTestRuntime()
	l := NewAnyList()
	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, l)
	rp.DataStack.Push(Int(55))
	rp.DataStack.Push(String("push"))
	RunOpsWithSubCode(NewTestTask(), rp, int(TypeDispatchKindInstanceMethod), OpTypeMemberDispatch, OpStop)
	if l.Len() != 1 || l.At(0) != Int(55) {
		t.Fatalf("list = %v, want [55]", l.At(0))
	}
}

func TestStringListNativeMethodPushPopViaRunFrame(t *testing.T) {
	l := NewStringList()
	slt := stringListType(t)
	rp := NewTestRunProvider()
	invokeListMethod(t, slt, l, "push", func(ds *Stack[any]) {
		ds.Push(String("a"))
	})
	invokeListMethod(t, slt, l, "push", func(ds *Stack[any]) {
		ds.Push(String("b"))
	})
	mm := slt.MethodByName("pop")
	f, ths := memberRunFrame(rp, l)
	mm.invokeCall(f, ths)
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "b" {
		t.Fatalf("pop = %v, want b", got)
	}
	if l.Len() != 1 || l.At(0) != "a" {
		t.Fatalf("list = %q, want [a]", l.At(0))
	}
}

func TestListNativeMethodInsertAtViaRunFrame(t *testing.T) {
	l := NewAnyList()
	l.AppendFromAny(Int(1))
	invokeListMethod(t, listType(t), l, "insertAt", func(ds *Stack[any]) {
		ds.Push(Int(2))
		ds.Push(Int(l.Len()))
	})
	if l.Len() != 2 || l.At(0) != Int(1) || l.At(1) != Int(2) {
		t.Fatalf("list = [%v %v], want [1 2]", l.At(0), l.At(1))
	}
}

func TestListNativeMethodDeleteAtViaRunFrame(t *testing.T) {
	l := NewAnyList()
	l.AppendFromAny(Int(1))
	l.AppendFromAny(Int(2))
	invokeListMethod(t, listType(t), l, "deleteAt", func(ds *Stack[any]) {
		ds.Push(Int(0))
	})
	if l.Len() != 1 || l.At(0) != Int(2) {
		t.Fatalf("after deleteAt: Len=%d At(0)=%v", l.Len(), l.At(0))
	}
}

func TestListNativeMethodDeleteItemViaRunFrame(t *testing.T) {
	l := NewAnyList()
	l.AppendFromAny(Int(1))
	l.AppendFromAny(Int(2))
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(1))
	mm := listType(t).MethodByName("deleteItem")
	f, ths := memberRunFrame(rp, l)
	mm.invokeCall(f, ths)
	got, ok := rp.DataStack.Pop().(Bool)
	if !ok || !bool(got) {
		t.Fatalf("deleteItem = %v, want true", got)
	}
	if l.Len() != 1 || l.At(0) != Int(2) {
		t.Fatalf("after deleteItem: Len=%d At(0)=%v", l.Len(), l.At(0))
	}
	rp.DataStack.Push(Int(99))
	mm.invokeCall(f, ths)
	got, ok = rp.DataStack.Pop().(Bool)
	if !ok || bool(got) {
		t.Fatalf("deleteItem absent = %v, want false", got)
	}
}

func TestListNativeMethodClearViaRunFrame(t *testing.T) {
	l := NewAnyList()
	l.AppendFromAny(Int(1))
	invokeListMethod(t, listType(t), l, "clear", nil)
	if l.Len() != 0 {
		t.Fatalf("Len() = %d, want 0 after clear", l.Len())
	}
}

func TestListNativePropertyLenViaRunFrame(t *testing.T) {
	rp := NewTestRunProvider()
	l := NewAnyList()
	l.Append(Int(1))
	l.Append(Int(2))
	pm := listType(t).PropertyByName("len")
	f, ths := memberRunFrame(rp, l)
	pm.invokeGet(f, ths)
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 2 {
		t.Fatalf("len = %v, want 2", got)
	}
}

func TestListNativeItemDataViaRunFrame(t *testing.T) {
	rp := NewTestRunProvider()
	l := NewAnyList()
	l.Append(String("a"))
	l.Append(String("b"))
	im := listType(t).itemsMetadata
	rp.DataStack.Push(Int(1))
	_, ths := memberRunFrame(rp, l)
	got := im.GetByOrdinal(ths.Obj, Int(1))
	if s, ok := got.(String); !ok || s != "b" {
		t.Fatalf("At(1) = %v, want b", got)
	}
	im.SetByOrdinal(ths.Obj, Int(0), String("z"))
	if l.At(0) != String("z") {
		t.Fatalf("At(0) = %v, want z", l.At(0))
	}
}

func TestTypeMemberDispatchListAppendByName(t *testing.T) {
	ResetTestRuntime()
	l := NewAnyList()
	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, l)
	rp.DataStack.Push(Int(99))
	rp.DataStack.Push(String("append"))
	RunOpsWithSubCode(NewTestTask(), rp, int(TypeDispatchKindInstanceMethod), OpTypeMemberDispatch, OpStop)
	if l.Len() != 1 || l.At(0) != Int(99) {
		t.Fatalf("list = %v, want [99]", l.At(0))
	}
}

func TestInterpretSessionListAppendDotLenFetch(t *testing.T) {
	t.Skip("pending compound interpret path coverage")
}

func TestColonListAppendCompileRun(t *testing.T) {
	t.Skip("pending compound compile path coverage")
}

func TestStringListNativeClassMethodNewViaRunFrame(t *testing.T) {
	rp := NewTestRunProvider()
	mm := stringListType(t).ClassMethodByName("new")
	f, ths := memberRunFrame(rp, stringListType(t))
	mm.invokeCall(f, ths)
	l, ok := rp.DataStack.Peek().(*StringList)
	if !ok || l.Len() != 0 {
		t.Fatalf("new list = %v, want empty *StringList", rp.DataStack.Peek())
	}
}

func TestListNativeClassMethodNewWithCapacity(t *testing.T) {
	cases := []struct {
		typeName string
		check    func(t *testing.T, v any)
	}{
		{"List", func(t *testing.T, v any) {
			l := v.(*AnyList)
			if l.Len() != 0 || cap(l.items) != 64 {
				t.Fatalf("List: len=%d cap=%d, want 0/64", l.Len(), cap(l.items))
			}
		}},
		{"StringList", func(t *testing.T, v any) {
			l := v.(*StringList)
			if l.Len() != 0 || cap(l.items) != 64 {
				t.Fatalf("StringList: len=%d cap=%d, want 0/64", l.Len(), cap(l.items))
			}
		}},
		{"Int32List", func(t *testing.T, v any) {
			l := v.(*Int32List)
			if l.Len() != 0 || cap(l.items) != 64 {
				t.Fatalf("Int32List: len=%d cap=%d, want 0/64", l.Len(), cap(l.items))
			}
		}},
		{"FloatList", func(t *testing.T, v any) {
			l := v.(*FloatList)
			if l.Len() != 0 || cap(l.items) != 64 {
				t.Fatalf("FloatList: len=%d cap=%d, want 0/64", l.Len(), cap(l.items))
			}
		}},
		{"AtomicList", func(t *testing.T, v any) {
			l := v.(*AtomicAnyList)
			if l.Len() != 0 || cap(l.items) != 64 {
				t.Fatalf("AtomicList: len=%d cap=%d, want 0/64", l.Len(), cap(l.items))
			}
		}},
		{"AtomicStringList", func(t *testing.T, v any) {
			l := v.(*AtomicStringList)
			if l.Len() != 0 || cap(l.items) != 64 {
				t.Fatalf("AtomicStringList: len=%d cap=%d, want 0/64", l.Len(), cap(l.items))
			}
		}},
		{"StringKeyValueList", func(t *testing.T, v any) {
			l := v.(*StringKeyValueList)
			if l.Len() != 0 || cap(l.items) != 64 {
				t.Fatalf("StringKeyValueList: len=%d cap=%d, want 0/64", l.Len(), cap(l.items))
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.typeName, func(t *testing.T) {
			typ := GetTypeRegistrar().Get(tc.typeName)
			if !typ.HasClassMethod("newWithCapacity") {
				t.Fatalf("%s missing class method newWithCapacity", tc.typeName)
			}
			rp := NewTestRunProvider()
			rp.DataStack.Push(Int(64))
			mm := typ.ClassMethodByName("newWithCapacity")
			f, ths := memberRunFrame(rp, typ)
			mm.invokeCall(f, ths)
			tc.check(t, rp.DataStack.Peek())
		})
	}
}

func TestListNativeClassMethodNewWithLen(t *testing.T) {
	cases := []struct {
		typeName string
		check    func(t *testing.T, v any)
	}{
		{"List", func(t *testing.T, v any) {
			l := v.(*AnyList)
			if l.Len() != 64 || cap(l.items) != 64 {
				t.Fatalf("List: len=%d cap=%d, want 64/64", l.Len(), cap(l.items))
			}
		}},
		{"StringList", func(t *testing.T, v any) {
			l := v.(*StringList)
			if l.Len() != 64 || cap(l.items) != 64 || l.At(0) != "" {
				t.Fatalf("StringList: len=%d cap=%d At(0)=%q, want 64/64/\"\"", l.Len(), cap(l.items), l.At(0))
			}
		}},
		{"Int32List", func(t *testing.T, v any) {
			l := v.(*Int32List)
			if l.Len() != 64 || cap(l.items) != 64 || l.At(0) != 0 {
				t.Fatalf("Int32List: len=%d cap=%d At(0)=%d, want 64/64/0", l.Len(), cap(l.items), l.At(0))
			}
		}},
		{"FloatList", func(t *testing.T, v any) {
			l := v.(*FloatList)
			if l.Len() != 64 || cap(l.items) != 64 || l.At(0) != 0 {
				t.Fatalf("FloatList: len=%d cap=%d At(0)=%v, want 64/64/0", l.Len(), cap(l.items), l.At(0))
			}
		}},
		{"AtomicList", func(t *testing.T, v any) {
			l := v.(*AtomicAnyList)
			if l.Len() != 64 || cap(l.items) != 64 {
				t.Fatalf("AtomicList: len=%d cap=%d, want 64/64", l.Len(), cap(l.items))
			}
		}},
		{"AtomicStringList", func(t *testing.T, v any) {
			l := v.(*AtomicStringList)
			if l.Len() != 64 || cap(l.items) != 64 || l.At(0) != "" {
				t.Fatalf("AtomicStringList: len=%d cap=%d At(0)=%q, want 64/64/\"\"", l.Len(), cap(l.items), l.At(0))
			}
		}},
		{"StringKeyValueList", func(t *testing.T, v any) {
			l := v.(*StringKeyValueList)
			if l.Len() != 64 || cap(l.items) != 64 {
				t.Fatalf("StringKeyValueList: len=%d cap=%d, want 64/64", l.Len(), cap(l.items))
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.typeName, func(t *testing.T) {
			typ := GetTypeRegistrar().Get(tc.typeName)
			if !typ.HasClassMethod("newWithLen") {
				t.Fatalf("%s missing class method newWithLen", tc.typeName)
			}
			rp := NewTestRunProvider()
			rp.DataStack.Push(Int(64))
			mm := typ.ClassMethodByName("newWithLen")
			f, ths := memberRunFrame(rp, typ)
			mm.invokeCall(f, ths)
			tc.check(t, rp.DataStack.Peek())
		})
	}
}

func TestStringListNativeMethodAppendViaRunFrame(t *testing.T) {
	l := NewStringList()
	invokeListMethod(t, stringListType(t), l, "append", func(ds *Stack[any]) {
		ds.Push(String("hi"))
	})
	if l.At(0) != "hi" {
		t.Fatalf("At(0) = %q, want hi", l.At(0))
	}
}

func TestStringListNativeItemDataViaRunFrame(t *testing.T) {
	rp := NewTestRunProvider()
	l := NewStringList()
	l.Append("a")
	im := stringListType(t).itemsMetadata
	rp.DataStack.Push(Int(0))
	_, ths := memberRunFrame(rp, l)
	got := im.GetByOrdinal(ths.Obj, Int(0))
	if s, ok := got.(string); !ok || s != "a" {
		t.Fatalf("At(0) = %v, want a", got)
	}
}

func TestTypeMemberDispatchStringListAppendByName(t *testing.T) {
	ResetTestRuntime()
	l := NewStringList()
	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, l)
	rp.DataStack.Push(String("x"))
	rp.DataStack.Push(String("append"))
	RunOpsWithSubCode(NewTestTask(), rp, int(TypeDispatchKindInstanceMethod), OpTypeMemberDispatch, OpStop)
	if l.At(0) != "x" {
		t.Fatalf("list = %q, want [x]", l.At(0))
	}
}

func TestInterpretListNewTsAppend(t *testing.T) {
	ResetTestRuntime()
	task, err := newSessionTask(SessionKindLocal)
	if err != nil {
		t.Fatal(err)
	}
	tr, err := NewTextReaderFromString("List.new^\n>ts\n1 t:.append^\n")
	if err != nil {
		t.Fatal(err)
	}
	rp := NewTestRunProvider()
	rp.SetParser(NewParserFromTextRead(tr))
	rp.WithTaskFindScope(task)
	if err := interpretSession(task, rp); err != nil {
		t.Fatal(err)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0 (append arg consumed)", rp.DataStack.Len())
	}
	if rp.TempStack.Len() != 1 {
		t.Fatalf("TS len = %d, want 1 (list handle kept)", rp.TempStack.Len())
	}
	l, ok := rp.TempStack.Peek().(*AnyList)
	if !ok {
		t.Fatalf("TS TOS = %T, want *AnyList", rp.TempStack.Peek())
	}
	if l.Len() != 1 || l.At(0) != Int(1) {
		t.Fatalf("list = [%v], want [1]", l.At(0))
	}
}

func TestInterpretListNewAppendOnDSWithoutTs(t *testing.T) {
	ResetTestRuntime()
	task, rp, _ := testParserWithInput(t, "List.new^ >ts 1 t:.append^ t:.len@\n")
	if err := interpretSession(task, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 1 {
		t.Fatalf("len = %v, want 1", got)
	}
}

func TestInterpretListStackPushPopPeekDepth(t *testing.T) {
	ResetTestRuntime()
	task, rp, _ := testParserWithInput(t, "List.new^ >ts 1 t:.push^ 2 t:.push^ t:.peek^ t:.depth^ t:.pop^ t:.pop^\n")
	if err := interpretSession(task, rp); err != nil {
		t.Fatal(err)
	}
	if rp.DataStack.Len() != 4 {
		t.Fatalf("DS len = %d, want 4 (peek, depth, pop, pop)", rp.DataStack.Len())
	}
	pop1, ok := rp.DataStack.Pop().(Int)
	if !ok || pop1 != 1 {
		t.Fatalf("first pop = %v, want 1", pop1)
	}
	pop2, ok := rp.DataStack.Pop().(Int)
	if !ok || pop2 != 2 {
		t.Fatalf("second pop = %v, want 2", pop2)
	}
	depth, ok := rp.DataStack.Pop().(Int)
	if !ok || depth != 2 {
		t.Fatalf("depth = %v, want 2", depth)
	}
	peek, ok := rp.DataStack.Pop().(Int)
	if !ok || peek != 2 {
		t.Fatalf("peek = %v, want 2", peek)
	}
	if rp.TempStack.Len() != 1 {
		t.Fatalf("TS len = %d, want 1 (list handle)", rp.TempStack.Len())
	}
	l, ok := rp.TempStack.Peek().(*AnyList)
	if !ok || l.Len() != 0 {
		t.Fatalf("list Len = %d, want 0 after two pops", l.Len())
	}
}

func TestListPrototypeChain(t *testing.T) {
	listT := GetTypeRegistrar().Get("List")
	stringListT := GetTypeRegistrar().Get("StringList")
	int32ListT := GetTypeRegistrar().Get("Int32List")
	if listT == nil || stringListT == nil || int32ListT == nil {
		t.Fatal("List family types not registered")
	}
	if listT.GetParent() != listProtoType {
		t.Fatalf("List.GetParent() = %v, want ProtoList", listT.GetParent())
	}
	if stringListT.parent != listProtoType {
		t.Fatalf("StringList.parent = %v, want ProtoList", stringListT.parent)
	}
	if int32ListT.parent != listProtoType {
		t.Fatalf("Int32List.parent = %v, want ProtoList", int32ListT.parent)
	}
}

func TestListHasShuffleNotSort(t *testing.T) {
	lt := listType(t)
	if !lt.HasMethod("shuffle") {
		t.Fatal("List should have shuffle")
	}
	// ProtoList carries a doc-only sort stub for help/:doc; List must not register a native sort.
	if lt.methodMetadata != nil && lt.methodMetadata.HasKey("sort") {
		t.Fatal("List (Any) should not register local sort")
	}
	mm := lt.GetMethodMetadata("sort")
	if mm == nil || mm.call != nil {
		t.Fatal("List.sort should resolve to ProtoList doc stub (nil call)")
	}
}

func TestStringListHasShuffleAndSort(t *testing.T) {
	slt := stringListType(t)
	if !slt.HasMethod("shuffle") || !slt.HasMethod("sort") {
		t.Fatal("StringList should have shuffle and sort")
	}
}

func TestAnyListShufflePreservesMultiset(t *testing.T) {
	l := NewAnyList()
	for _, v := range []any{Int(1), Int(2), Int(3), Int(4), Int(5)} {
		l.AppendFromAny(v)
	}
	before := make([]Int, l.Len())
	for i := range before {
		before[i] = l.At(i).(Int)
	}
	l.Shuffle()
	after := make([]Int, l.Len())
	for i := range after {
		after[i] = l.At(i).(Int)
	}
	slices.Sort(before)
	slices.Sort(after)
	if !slices.Equal(before, after) {
		t.Fatalf("after shuffle: before=%v after=%v", before, after)
	}
}

func TestStringListSortViaRunFrame(t *testing.T) {
	l := NewStringList()
	for _, s := range []string{"cherry", "apple", "banana"} {
		l.Append(s)
	}
	invokeListMethod(t, stringListType(t), l, "sort", nil)
	if l.At(0) != "apple" || l.At(1) != "banana" || l.At(2) != "cherry" {
		t.Fatalf("sorted = [%q %q %q], want [apple banana cherry]", l.At(0), l.At(1), l.At(2))
	}
}

func TestInt32ListSortViaRunFrame(t *testing.T) {
	l := NewInt32List()
	for _, n := range []int32{30, 10, 20} {
		l.Append(n)
	}
	invokeListMethod(t, GetTypeRegistrar().Get("Int32List"), l, "sort", nil)
	if l.At(0) != 10 || l.At(1) != 20 || l.At(2) != 30 {
		t.Fatalf("sorted = [%d %d %d], want [10 20 30]", l.At(0), l.At(1), l.At(2))
	}
}
