package core

import (
	"testing"
)

func runToJson(t *testing.T, task *Task, rp *RunProvider, val any) string {
	t.Helper()
	w := globalDictionaryContext.dict.GetByName(">json")
	if w == nil {
		t.Fatal("missing >json (bootstrap render.fnl?)")
	}
	rp.DataStack.Push(val)
	if err := runWordInlineExec(task, rp, w); err != nil {
		t.Fatalf(">json: %v", err)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok {
		t.Fatalf("result = %T, want String", got)
	}
	return string(got)
}

func TestJsonRenderPrimitiveOutputs(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	cases := []struct {
		name string
		val  any
		want string
	}{
		{"null", any(nil), "null"},
		{"true", Bool(true), "true"},
		{"false", Bool(false), "false"},
		{"int", Int(-49), "-49"},
		{"uint", Uint(0xFF), "255"},
		{"string", String("the rain\nfalls in the plane\n"), `"the rain\nfalls in the plane\n"`},
		{"float_non_integral", Float(3.14159), "3.14159"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := runToJson(t, task, rp, tc.val)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestJsonRenderNumberWireShape(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	for _, tc := range []struct {
		name    string
		val     any
		wantDot bool
	}{
		{"int", Int(-49), false},
		{"uint", Uint(0xFF), false},
		{"float_non_integral", Float(3.14159), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := runToJson(t, task, rp, tc.val)
			hasDot := false
			for i := 0; i < len(got); i++ {
				if got[i] == '.' {
					hasDot = true
					break
				}
			}
			if hasDot != tc.wantDot {
				t.Fatalf("got %q, wantDot=%v", got, tc.wantDot)
			}
		})
	}
}

func TestJsonRenderAnyListCompact(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	lst := NewAnyList()
	lst.Append(any(nil))
	lst.Append(Bool(true))
	lst.Append(Int(-49))
	lst.Append(Uint(0xFF))
	lst.Append(String("the rain\nfalls in the plane\n"))
	got := runToJson(t, task, rp, lst)
	want := `[null,true,-49,255,"the rain\nfalls in the plane\n"]`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestJsonRenderDynamicObjectCompact(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	do := NewDynamicObject()
	do.Set("Null", any(nil))
	do.Set("Bool", Bool(true))
	do.Set("Int", Int(-49))
	do.Set("Uint", Uint(0xFF))
	do.Set("Float", Float(3.14159256))
	do.Set("String", String("the rain\nfalls in the plane\n"))
	got := runToJson(t, task, rp, do)
	want := `{"Null":null,"Bool":true,"Int":-49,"Uint":255,"Float":3.14159256,"String":"the rain\nfalls in the plane\n"}`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
