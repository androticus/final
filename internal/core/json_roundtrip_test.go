package core

import (
	"math"
	"testing"
)

func runParseJson(t *testing.T, task *Task, rp *RunProvider, json string) (root any, ok bool) {
	t.Helper()
	w := globalDictionaryContext.dict.GetByName("parseJson")
	if w == nil {
		t.Fatal("missing parseJson (bootstrap parse.fnl?)")
	}
	rp.DataStack.Push(String(json))
	if err := runWordInlineExec(task, rp, w); err != nil {
		t.Fatalf("parseJson: %v", err)
	}
	if rp.DataStack.Len() != 2 {
		t.Fatalf("parseJson DS len = %d, want 2 (root ok)", rp.DataStack.Len())
	}
	okVal, okBool := rp.DataStack.Pop().(Bool)
	if !okBool {
		t.Fatalf("parseJson ok = %T, want Bool", okVal)
	}
	if !okVal {
		errVal, _ := rp.DataStack.Pop().(String)
		t.Fatalf("parseJson failed: %s", errVal)
	}
	return rp.DataStack.Pop(), true
}

func runFromJson(t *testing.T, task *Task, rp *RunProvider, json string) any {
	t.Helper()
	root, ok := runParseJson(t, task, rp, json)
	if !ok {
		t.Fatal("parseJson returned false")
	}
	return root
}

func assertJsonRoundTrip(t *testing.T, task *Task, rp *RunProvider, val any) {
	t.Helper()
	wire := runToJson(t, task, rp, val)
	got := runFromJson(t, task, rp, wire)
	want := jsonDecodeExpected(val)
	if !jsonValueEqual(got, want) {
		t.Fatalf("roundtrip mismatch:\n  wire: %q\n  got:  %#v\n  want: %#v", wire, got, want)
	}
}

// jsonDecodeExpected applies post-decode relaxations: JSON has no Uint type (119 policy).
func jsonDecodeExpected(val any) any {
	switch v := val.(type) {
	case Uint:
		return Int(v)
	case *DynamicObject:
		out := NewDynamicObject()
		for i := 0; i < v.Len(); i++ {
			item := v.ItemAt(i)
			out.Set(item.Name, jsonDecodeExpected(item.Value))
		}
		return out
	case *List[any]:
		out := NewAnyList()
		for i := 0; i < v.Len(); i++ {
			out.Append(jsonDecodeExpected(v.AtAsAny(i)))
		}
		return out
	default:
		return val
	}
}

func jsonValueEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	switch av := a.(type) {
	case Bool:
		bv, ok := b.(Bool)
		return ok && av == bv
	case Int:
		return jsonIntEqual(av, b)
	case Uint:
		return jsonIntEqual(Int(av), b)
	case Float:
		return jsonFloatEqual(av, b)
	case String:
		bv, ok := b.(String)
		return ok && av == bv
	case *DynamicObject:
		bv, ok := b.(*DynamicObject)
		if !ok || av.Len() != bv.Len() {
			return false
		}
		for i := 0; i < av.Len(); i++ {
			item := av.ItemAt(i)
			if !bv.HasField(item.Name) {
				return false
			}
			if !jsonValueEqual(item.Value, bv.Get(item.Name)) {
				return false
			}
		}
		return true
	case *List[any]:
		bv, ok := b.(*List[any])
		if !ok || av.Len() != bv.Len() {
			return false
		}
		for i := 0; i < av.Len(); i++ {
			if !jsonValueEqual(av.AtAsAny(i), bv.AtAsAny(i)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func jsonIntEqual(a Int, b any) bool {
	switch bv := b.(type) {
	case Int:
		return a == bv
	case Uint:
		return a == Int(bv)
	default:
		return false
	}
}

func jsonFloatEqual(a Float, b any) bool {
	switch bv := b.(type) {
	case Float:
		return math.Abs(float64(a-bv)) < 1e-9
	case Int:
		return math.Abs(float64(a)-float64(bv)) < 1e-9
	default:
		return false
	}
}

func makeJsonNestedObject() *DynamicObject {
	o := NewDynamicObject()
	o.Set("label", String("nested"))
	o.Set("count", Int(2))
	return o
}

func makeJsonNestedArray() *List[any] {
	lst := NewAnyList()
	lst.Append(Bool(false))
	lst.Append(String("item"))
	return lst
}

func makeJsonKitchenSinkObject() *DynamicObject {
	do := NewDynamicObject()
	do.Set("Null", any(nil))
	do.Set("Bool", Bool(true))
	do.Set("Int", Int(-49))
	do.Set("Uint", Uint(0xFF))
	do.Set("Float", Float(3.14159256))
	do.Set("String", String("the rain\nfalls in the plane\n"))
	do.Set("nestedObject", makeJsonNestedObject())
	do.Set("nestedArray", makeJsonNestedArray())

	mixed := NewAnyList()
	mixed.Append(makeJsonNestedObject())
	inner := NewAnyList()
	inner.Append(Int(1))
	inner.Append(Int(2))
	mixed.Append(inner)
	mixed.Append(String("tail"))
	do.Set("arrayField", mixed)
	return do
}

func makeJsonKitchenSinkArray() *List[any] {
	lst := NewAnyList()
	lst.Append(any(nil))
	lst.Append(Bool(false))
	lst.Append(Int(7))
	lst.Append(Uint(42))
	lst.Append(Float(2.71828))
	lst.Append(String("root"))
	lst.Append(makeJsonNestedObject())
	inner := NewAnyList()
	inner.Append(Bool(true))
	inner.Append(makeJsonNestedArray())
	lst.Append(inner)
	return lst
}

// Disabled while parse.fnl (and lex.fnl/session.fnl) are commented out of bootstrapTranche2Manifest.
/*
func TestJsonRoundTripPrimitiveRoots(t *testing.T) {
	cases := []struct {
		name string
		val  any
	}{
		{"null", any(nil)},
		{"true", Bool(true)},
		{"false", Bool(false)},
		{"int", Int(-49)},
		{"uint", Uint(0xFF)},
		{"float", Float(3.14159256)},
		{"string", String("the rain\nfalls in the plane\n")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ResetTestRuntime()
			assertJsonRoundTrip(t, NewTestTask(), NewTestRunProvider(), tc.val)
		})
	}
}
*/

/*
func TestJsonRoundTripObjectRoot(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	assertJsonRoundTrip(t, task, rp, makeJsonKitchenSinkObject())
}
*/

/*
func TestJsonRoundTripArrayRoot(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	assertJsonRoundTrip(t, task, rp, makeJsonKitchenSinkArray())
}
*/

/*
func TestJsonRoundTripPreservesNestedShapes(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	val := makeJsonKitchenSinkObject()
	wire := runToJson(t, task, rp, val)
	got := runFromJson(t, task, rp, wire)
	want := jsonDecodeExpected(val).(*DynamicObject)

	arrayField, ok := got.(*DynamicObject).Get("arrayField").(*List[any])
	if !ok {
		t.Fatalf("arrayField = %T, want *List[any]", got.(*DynamicObject).Get("arrayField"))
	}
	if arrayField.Len() != 3 {
		t.Fatalf("arrayField len = %d, want 3", arrayField.Len())
	}
	if _, ok := arrayField.AtAsAny(0).(*DynamicObject); !ok {
		t.Fatalf("arrayField[0] = %T, want *DynamicObject", arrayField.AtAsAny(0))
	}
	if _, ok := arrayField.AtAsAny(1).(*List[any]); !ok {
		t.Fatalf("arrayField[1] = %T, want *List[any]", arrayField.AtAsAny(1))
	}

	nestedObject := want.Get("nestedObject")
	if !jsonValueEqual(got.(*DynamicObject).Get("nestedObject"), nestedObject) {
		t.Fatalf("nestedObject mismatch: got %#v want %#v", got.(*DynamicObject).Get("nestedObject"), nestedObject)
	}
}
*/
