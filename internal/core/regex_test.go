package core

import (
	"strings"
	"testing"
)

func testRegexType(t *testing.T) *Type {
	t.Helper()
	typ := GetTypeRegistrar().Get("Regex")
	if typ == nil {
		t.Fatal("Regex type not registered")
	}
	return typ
}

func invokeRegexClass(t *testing.T, name string, pushArgs func(*Stack[any])) *RunProvider {
	t.Helper()
	mm := testRegexType(t).ClassMethodByName(name)
	if mm == nil {
		t.Fatalf("missing Regex.%s class method", name)
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, testRegexType(t))
	if pushArgs != nil {
		pushArgs(rp.DataStack)
	}
	mm.invokeCall(f, ths)
	return rp
}

func invokeRegexInst(t *testing.T, re *Regex, name string, pushArgs func(*Stack[any])) *RunProvider {
	t.Helper()
	mm := testRegexType(t).MethodByName(name)
	if mm == nil {
		t.Fatalf("missing Regex.%s method", name)
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, re)
	if pushArgs != nil {
		pushArgs(rp.DataStack)
	}
	mm.invokeCall(f, ths)
	return rp
}

func TestRegexNewAndProperties(t *testing.T) {
	ResetTestRuntime()
	rp := invokeRegexClass(t, "new", func(ds *Stack[any]) {
		ds.Push(String(`(?P<code>[A-Z]+)-(\d+)`))
	})
	re := rp.DataStack.Pop().(*Regex)
	if re.re.String() != `(?P<code>[A-Z]+)-(\d+)` {
		t.Fatalf("pattern = %q", re.re.String())
	}

	rt := testRegexType(t)
	f, ths := memberRunFrame(NewTestRunProvider(), re)

	pm := rt.PropertyByName("pattern")
	pm.invokeGet(f, ths)
	if got := f.RP.DataStack.Pop().(String); got != String(`(?P<code>[A-Z]+)-(\d+)`) {
		t.Fatalf(".pattern = %q", got)
	}

	f, ths = memberRunFrame(NewTestRunProvider(), re)
	pm = rt.PropertyByName("numGroups")
	pm.invokeGet(f, ths)
	if got := f.RP.DataStack.Pop().(Int); got != Int(2) {
		t.Fatalf(".numGroups = %v, want 2", got)
	}

	f, ths = memberRunFrame(NewTestRunProvider(), re)
	pm = rt.PropertyByName("groupNames")
	pm.invokeGet(f, ths)
	names := f.RP.DataStack.Pop().(*StringList)
	if names.Len() != 3 {
		t.Fatalf("groupNames len = %d, want 3 (whole match + 2 groups)", names.Len())
	}
	if names.At(0) != "0" {
		t.Fatalf("groupNames[0] = %q, want 0 (whole match key)", names.At(0))
	}
	if names.At(1) != "code" {
		t.Fatalf("groupNames[1] = %q, want code", names.At(1))
	}
	if names.At(2) != "2" {
		t.Fatalf("groupNames[2] = %q, want 2 (unnamed)", names.At(2))
	}
}

func TestRegexNewBadPatternPanics(t *testing.T) {
	ResetTestRuntime()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on bad pattern")
		}
	}()
	invokeRegexClass(t, "new", func(ds *Stack[any]) {
		ds.Push(String(`(unclosed`))
	})
}

func TestRegexIsPatternValid(t *testing.T) {
	ResetTestRuntime()
	rp := invokeRegexClass(t, "isPatternValid?", func(ds *Stack[any]) {
		ds.Push(String(`^\d+$`))
	})
	if got := rp.DataStack.Pop().(Bool); !bool(got) {
		t.Fatal("valid pattern → false")
	}
	rp = invokeRegexClass(t, "isPatternValid?", func(ds *Stack[any]) {
		ds.Push(String(`(unclosed`))
	})
	if got := rp.DataStack.Pop().(Bool); bool(got) {
		t.Fatal("invalid pattern → true")
	}
}

func TestRegexClassMatch(t *testing.T) {
	ResetTestRuntime()
	rp := invokeRegexClass(t, "match", func(ds *Stack[any]) {
		ds.Push(String("42"))
		ds.Push(String(`^\d+$`))
	})
	if got := rp.DataStack.Pop().(Bool); !bool(got) {
		t.Fatal("want match true")
	}
	rp = invokeRegexClass(t, "match", func(ds *Stack[any]) {
		ds.Push(String("x"))
		ds.Push(String(`^\d+$`))
	})
	if got := rp.DataStack.Pop().(Bool); bool(got) {
		t.Fatal("want match false")
	}
}

func TestRegexClassFindModal(t *testing.T) {
	ResetTestRuntime()
	rp := invokeRegexClass(t, "find", func(ds *Stack[any]) {
		ds.Push(String("hello 42 world"))
		ds.Push(String(`\d+`))
	})
	ok := rp.DataStack.Pop().(Bool)
	found := rp.DataStack.Pop().(String)
	if !bool(ok) || found != String("42") {
		t.Fatalf("find = %q %v", found, ok)
	}

	rp = invokeRegexClass(t, "find", func(ds *Stack[any]) {
		ds.Push(String("hello world"))
		ds.Push(String(`\d+`))
	})
	if rp.DataStack.Len() != 1 {
		t.Fatalf("no-match DS len = %d, want 1", rp.DataStack.Len())
	}
	if got := rp.DataStack.Pop().(Bool); bool(got) {
		t.Fatal("want false")
	}
}

func TestRegexClassFindEmptyMatchIsTrue(t *testing.T) {
	ResetTestRuntime()
	// a* can match empty at start of "bbb"
	rp := invokeRegexClass(t, "find", func(ds *Stack[any]) {
		ds.Push(String("bbb"))
		ds.Push(String(`a*`))
	})
	ok := rp.DataStack.Pop().(Bool)
	found := rp.DataStack.Pop().(String)
	if !bool(ok) || found != String("") {
		t.Fatalf("empty match = %q %v, want \"\" true", found, ok)
	}
}

func TestRegexQuoteMeta(t *testing.T) {
	ResetTestRuntime()
	rp := invokeRegexClass(t, "quoteMeta", func(ds *Stack[any]) {
		ds.Push(String(`1.2`))
	})
	if got := string(rp.DataStack.Pop().(String)); got != `1\.2` {
		t.Fatalf("quoteMeta = %q, want 1\\.2", got)
	}
}

func TestRegexInstanceFindAllAndSplit(t *testing.T) {
	ResetTestRuntime()
	re := NewRegex(`\w+`)
	rp := invokeRegexInst(t, re, "findAll", func(ds *Stack[any]) {
		ds.Push(String("one two three"))
	})
	ok := rp.DataStack.Pop().(Bool)
	parts := rp.DataStack.Pop().(*StringList)
	if !bool(ok) || parts.Len() != 3 {
		t.Fatalf("findAll = len %d ok %v", parts.Len(), ok)
	}
	if parts.At(0) != "one" || parts.At(1) != "two" || parts.At(2) != "three" {
		t.Fatalf("parts = %v", []string{parts.At(0), parts.At(1), parts.At(2)})
	}

	rp = invokeRegexInst(t, re, "findAll", func(ds *Stack[any]) {
		ds.Push(String("!!!"))
	})
	if rp.DataStack.Len() != 1 || bool(rp.DataStack.Pop().(Bool)) {
		t.Fatal("findAll no matches want false only")
	}

	re = NewRegex(`\s+`)
	rp = invokeRegexInst(t, re, "split", func(ds *Stack[any]) {
		ds.Push(String("a  b"))
	})
	parts = rp.DataStack.Pop().(*StringList)
	if parts.Len() != 2 || parts.At(0) != "a" || parts.At(1) != "b" {
		t.Fatalf("split = %#v", []string{parts.At(0), parts.At(1)})
	}
}

func TestRegexFindAllNAndFindIndex(t *testing.T) {
	ResetTestRuntime()
	re := NewRegex(`\d+`)
	rp := invokeRegexInst(t, re, "findAllN", func(ds *Stack[any]) {
		ds.Push(String("1 22 333"))
		ds.Push(Int(2))
	})
	ok := rp.DataStack.Pop().(Bool)
	parts := rp.DataStack.Pop().(*StringList)
	if !bool(ok) || parts.Len() != 2 {
		t.Fatalf("findAllN = len %d ok %v", parts.Len(), ok)
	}

	rp = invokeRegexInst(t, re, "findIndex", func(ds *Stack[any]) {
		ds.Push(String("ab12cd"))
	})
	ok = rp.DataStack.Pop().(Bool)
	end := rp.DataStack.Pop().(Int)
	start := rp.DataStack.Pop().(Int)
	if !bool(ok) || start != Int(2) || end != Int(4) {
		t.Fatalf("findIndex = %v %v %v", start, end, ok)
	}

	rp = invokeRegexInst(t, re, "findIndex", func(ds *Stack[any]) {
		ds.Push(String("abcd"))
	})
	if rp.DataStack.Len() != 1 || bool(rp.DataStack.Pop().(Bool)) {
		t.Fatal("findIndex miss want false")
	}
}

func TestRegexMatchObjectNamedAndUnnamed(t *testing.T) {
	ResetTestRuntime()
	re := NewRegex(`(?P<code>[A-Z]+)-(?P<num>\d+)`)
	rp := invokeRegexInst(t, re, "matchObject", func(ds *Stack[any]) {
		ds.Push(String("AB-12"))
	})
	ok := rp.DataStack.Pop().(Bool)
	o := rp.DataStack.Pop().(*DynamicObject)
	if !bool(ok) {
		t.Fatal("want match")
	}
	if o.Get("code") != String("AB") || o.Get("num") != String("12") {
		t.Fatalf("fields code=%v num=%v", o.Get("code"), o.Get("num"))
	}
	if o.Get("0") != String("AB-12") {
		t.Fatalf("whole match \"0\" = %v, want AB-12", o.Get("0"))
	}
	if o.Len() != 3 {
		t.Fatalf("len = %d, want 3", o.Len())
	}
	if o.ItemAt(0).Value != String("AB-12") || o.ItemAt(1).Value != String("AB") || o.ItemAt(2).Value != String("12") {
		t.Fatal("creation order wrong (expect Go group order 0,1,2)")
	}

	re = NewRegex(`(\d+)-(\d+)`)
	rp = invokeRegexInst(t, re, "matchObject", func(ds *Stack[any]) {
		ds.Push(String("3-4"))
	})
	ok = rp.DataStack.Pop().(Bool)
	o = rp.DataStack.Pop().(*DynamicObject)
	if !bool(ok) || o.Get("0") != String("3-4") || o.Get("1") != String("3") || o.Get("2") != String("4") {
		t.Fatalf("unnamed = 0=%v 1=%v 2=%v ok=%v", o.Get("0"), o.Get("1"), o.Get("2"), ok)
	}

	rp = invokeRegexInst(t, re, "matchObject", func(ds *Stack[any]) {
		ds.Push(String("nope"))
	})
	if rp.DataStack.Len() != 1 || bool(rp.DataStack.Pop().(Bool)) {
		t.Fatal("no match want false")
	}
}

func TestRegexMatchObjectZeroGroups(t *testing.T) {
	ResetTestRuntime()
	re := NewRegex(`hello`)
	rp := invokeRegexInst(t, re, "matchObject", func(ds *Stack[any]) {
		ds.Push(String("hello"))
	})
	ok := rp.DataStack.Pop().(Bool)
	o := rp.DataStack.Pop().(*DynamicObject)
	if !bool(ok) || o.Len() != 1 || o.Get("0") != String("hello") {
		t.Fatalf("no captures = len %d \"0\"=%v ok %v", o.Len(), o.Get("0"), ok)
	}
}

func TestRegexReplace(t *testing.T) {
	ResetTestRuntime()
	re := NewRegex(`(\d+)`)
	rp := invokeRegexInst(t, re, "replaceAll", func(ds *Stack[any]) {
		ds.Push(String("a1b2"))
		ds.Push(String(`N`))
	})
	if got := rp.DataStack.Pop().(String); got != String("aNbN") {
		t.Fatalf("replaceAll = %q", got)
	}

	re = NewRegex(`\$`)
	rp = invokeRegexInst(t, re, "replaceAllLiteral", func(ds *Stack[any]) {
		ds.Push(String("$1"))
		ds.Push(String(`$0`))
	})
	if got := rp.DataStack.Pop().(String); got != String("$01") {
		// ReplaceAllLiteralString("$1", "$0") replaces $ with $0 → "$01"
		t.Fatalf("replaceAllLiteral = %q", got)
	}
}

func TestRegexNaturalString(t *testing.T) {
	re := NewRegex(`\d+`)
	got := re.NaturalString()
	if !strings.HasPrefix(got, "Regex pattern: ") || !strings.Contains(got, `\d+`) {
		t.Fatalf("NaturalString = %q", got)
	}
	if formatNaturalString(re) != got {
		t.Fatalf("formatNaturalString mismatch: %q vs %q", formatNaturalString(re), got)
	}
}

func TestRegexRenderJson(t *testing.T) {
	ResetTestRuntime()
	re := NewRegex(`a"b`)
	rp := NewTestRunProvider()
	sb := NewStringBuilder()
	rp.ZStack.Push(sb)
	mm := testRegexType(t).MethodByName("_renderJson")
	f, ths := memberRunFrame(rp, re)
	mm.invokeCall(f, ths)
	got := sb.ToString()
	want := stringJsonWire(String(`a"b`))
	if got != want {
		t.Fatalf("_renderJson = %q, want %q", got, want)
	}
}
