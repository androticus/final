package core

import (
	"testing"
)

func runStrOp(t *testing.T, subCode int, push func(*Stack[any])) any {
	t.Helper()
	ResetTestRuntime()
	rp := NewTestRunProvider()
	push(rp.DataStack)
	RunOpsWithSubCode(NewTestTask(), rp, subCode, OpStrOp, OpStop)
	if rp.DataStack.Len() != 1 {
		t.Fatalf("stack depth = %d, want 1", rp.DataStack.Len())
	}
	return rp.DataStack.Pop()
}

func TestStrOpWordsRegistered(t *testing.T) {
	ResetTestRuntime()
	for _, name := range []string{
		"str::len", "str::trimStart", "str::trimEnd", "str::trim", "str::cat", "str::+",
		"str::isEmpty", "str::startsWith", "str::endsWith", "str::indexOf", "str::split", "str::splitLines", "str::join",
		"str::toLower", "str::toUpper", "str::sub",
	} {
		w := globalDictionaryContext.FindWordByName(name)
		if w == nil {
			t.Fatalf("missing word %q", name)
		}
		if w.opCode != OpStrOp {
			t.Fatalf("%q opCode = %v, want OpStrOp", name, w.opCode)
		}
	}
	cat := globalDictionaryContext.FindWordByName("str::cat")
	plus := globalDictionaryContext.FindWordByName("str::+")
	if cat.subCode != plus.subCode {
		t.Fatalf("str::cat subCode %d != str::+ subCode %d", cat.subCode, plus.subCode)
	}
}

func TestOpStrLen(t *testing.T) {
	got := runStrOp(t, SC_str_len, func(ds *Stack[any]) {
		ds.Push(String("hello"))
	})
	if got != Int(5) {
		t.Fatalf("len = %v, want 5", got)
	}
}

func TestOpStrTrimOps(t *testing.T) {
	cases := []struct {
		subCode int
		in      String
		want    String
	}{
		{SC_str_trimStart, "  ab  ", "ab  "},
		{SC_str_trimEnd, "  ab  ", "  ab"},
		{SC_str_trim, "  ab  ", "ab"},
	}
	for _, tc := range cases {
		got := runStrOp(t, tc.subCode, func(ds *Stack[any]) {
			ds.Push(tc.in)
		}).(String)
		if got != tc.want {
			t.Fatalf("subcode %d(%q) = %q, want %q", tc.subCode, tc.in, got, tc.want)
		}
	}
}

func TestOpStrCaseOps(t *testing.T) {
	cases := []struct {
		subCode int
		in      String
		want    String
	}{
		{SC_str_toLower, "Hello", "hello"},
		{SC_str_toLower, "ΑΒΓ", "αβγ"},
		{SC_str_toUpper, "Hello", "HELLO"},
		{SC_str_toUpper, "αβγ", "ΑΒΓ"},
	}
	for _, tc := range cases {
		got := runStrOp(t, tc.subCode, func(ds *Stack[any]) {
			ds.Push(tc.in)
		}).(String)
		if got != tc.want {
			t.Fatalf("subcode %d(%q) = %q, want %q", tc.subCode, tc.in, got, tc.want)
		}
	}
}

func TestOpStrSub(t *testing.T) {
	cases := []struct {
		s, find, replace String
		want             String
	}{
		{"banana", "a", "o", "bonono"},
		{"foo", "x", "y", "foo"},
		{"aa", "a", "aa", "aaaa"},
	}
	for _, tc := range cases {
		got := runStrOp(t, SC_str_sub, func(ds *Stack[any]) {
			ds.Push(tc.s)
			ds.Push(tc.find)
			ds.Push(tc.replace)
		}).(String)
		if got != tc.want {
			t.Fatalf("sub(%q, %q, %q) = %q, want %q", tc.s, tc.find, tc.replace, got, tc.want)
		}
	}
}

func TestOpStrCat(t *testing.T) {
	catWord := globalDictionaryContext.FindWordByName("str::cat")
	plusWord := globalDictionaryContext.FindWordByName("str::+")
	for _, tc := range []struct {
		name    string
		subCode int
		a, b    String
		want    String
	}{
		{"str::cat", catWord.subCode, "foo", "bar", "foobar"},
		{"str::+", plusWord.subCode, "x", "y", "xy"},
	} {
		got := runStrOp(t, tc.subCode, func(ds *Stack[any]) {
			ds.Push(tc.a)
			ds.Push(tc.b)
		}).(String)
		if got != tc.want {
			t.Fatalf("%s = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestOpStrOpBadTypePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	runStrOp(t, SC_str_len, func(ds *Stack[any]) {
		ds.Push(Int(42))
	})
}

func TestOpStrOtherOps(t *testing.T) {
	s := String("internal/final/core/0040-json/0046-parse.fnl")

	if got := runStrOp(t, SC_str_isEmpty, func(ds *Stack[any]) {
		ds.Push(String(""))
	}); got != Bool(true) {
		t.Fatalf("isEmpty empty = %v, want true", got)
	}
	if got := runStrOp(t, SC_str_isEmpty, func(ds *Stack[any]) {
		ds.Push(String("x"))
	}); got != Bool(false) {
		t.Fatalf("isEmpty x = %v, want false", got)
	}

	if got := runStrOp(t, SC_str_startsWith, func(ds *Stack[any]) {
		ds.Push(s)
		ds.Push(String("internal/"))
	}); got != Bool(true) {
		t.Fatalf("startsWith = %v, want true", got)
	}

	if got := runStrOp(t, SC_str_endsWith, func(ds *Stack[any]) {
		ds.Push(s)
		ds.Push(String(".fnl"))
	}); got != Bool(true) {
		t.Fatalf("endsWith = %v, want true", got)
	}

	if got := runStrOp(t, SC_str_indexOf, func(ds *Stack[any]) {
		ds.Push(String("banana"))
		ds.Push(String("an"))
	}); got != Int(1) {
		t.Fatalf("indexOf = %v, want 1", got)
	}

	list := runStrOp(t, SC_str_split, func(ds *Stack[any]) {
		ds.Push(String("a,b,c"))
		ds.Push(String(","))
	}).(*StringList)
	if list.Len() != 3 || list.At(0) != "a" || list.At(1) != "b" || list.At(2) != "c" {
		t.Fatalf("split = %#v", list.items)
	}

	lines := runStrOp(t, SC_str_splitLines, func(ds *Stack[any]) {
		ds.Push(String("a\nb\n"))
	}).(*StringList)
	if lines.Len() != 2 || lines.At(0) != "a" || lines.At(1) != "b" {
		t.Fatalf("splitLines = %#v", lines.items)
	}
}

func TestOpStrJoin(t *testing.T) {
	cases := []struct {
		name  string
		build func() any
		sep   String
		want  String
	}{
		{
			name: "StringList",
			build: func() any {
				l := NewStringList()
				l.Append("a")
				l.Append("b")
				l.Append("c")
				return l
			},
			sep:  ",",
			want: "a,b,c",
		},
		{
			name: "List",
			build: func() any {
				l := NewAnyList()
				l.AppendFromAny(String("x"))
				l.AppendFromAny(String("y"))
				return l
			},
			sep:  "/",
			want: "x/y",
		},
		{
			name: "empty StringList",
			build: func() any {
				return NewStringList()
			},
			sep:  ",",
			want: "",
		},
		{
			name: "single item",
			build: func() any {
				l := NewStringList()
				l.Append("only")
				return l
			},
			sep:  ",",
			want: "only",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := runStrOp(t, SC_str_join, func(ds *Stack[any]) {
				ds.Push(tc.build())
				ds.Push(tc.sep)
			}).(String)
			if got != tc.want {
				t.Fatalf("join = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestOpStrJoinBadItemPanics(t *testing.T) {
	l := NewAnyList()
	l.AppendFromAny(Int(42))
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	runStrOp(t, SC_str_join, func(ds *Stack[any]) {
		ds.Push(l)
		ds.Push(String(","))
	})
}
