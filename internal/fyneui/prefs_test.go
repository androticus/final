package fyneui

import (
	"testing"

	"github.com/androticus/final/internal/core"
)

const (
	prefsTestKeyBool       = "final-go.test.Prefs.bool"
	prefsTestKeyInt        = "final-go.test.Prefs.int"
	prefsTestKeyFloat      = "final-go.test.Prefs.float"
	prefsTestKeyString     = "final-go.test.Prefs.string"
	prefsTestKeyIntList    = "final-go.test.Prefs.intList"
	prefsTestKeyFloatList  = "final-go.test.Prefs.floatList"
	prefsTestKeyStringList = "final-go.test.Prefs.stringList"
)

var prefsTestKeys = []string{
	prefsTestKeyBool,
	prefsTestKeyInt,
	prefsTestKeyFloat,
	prefsTestKeyString,
	prefsTestKeyIntList,
	prefsTestKeyFloatList,
	prefsTestKeyStringList,
}

func prefsType(t *testing.T) *core.Type {
	t.Helper()
	typ := core.GetTypeRegistrar().Get("Prefs")
	if typ == nil {
		t.Fatal("Prefs type not registered")
	}
	return typ
}

func callPrefs(typ *core.Type, f *core.RunFrame, name string) {
	typ.ClassMethodByName(name).Invoke(f, nil)
}

func removePrefsTestKeys(typ *core.Type, f *core.RunFrame) {
	ds := f.RP.DataStack
	for _, k := range prefsTestKeys {
		ds.Push(core.String(k))
		callPrefs(typ, f, "remove")
	}
}

func TestPrefsRoundTrip(t *testing.T) {
	typ := prefsType(t)
	rp := core.NewRunProvider()
	f := &core.RunFrame{RP: rp}
	ds := rp.DataStack

	t.Cleanup(func() {
		removePrefsTestKeys(typ, f)
	})

	// a. blanket remove on all test keys
	removePrefsTestKeys(typ, f)

	// b. missing keys: scalar defaults; lists not found
	ds.Push(core.String(prefsTestKeyBool))
	ds.Push(core.Bool(true))
	callPrefs(typ, f, "getBool")
	if got := bool(ds.Pop().(core.Bool)); !got {
		t.Fatalf("getBool missing: got %v, want default true", got)
	}

	ds.Push(core.String(prefsTestKeyInt))
	ds.Push(core.Int(42))
	callPrefs(typ, f, "getInt")
	if got := int(ds.Pop().(core.Int)); got != 42 {
		t.Fatalf("getInt missing: got %d, want default 42", got)
	}

	ds.Push(core.String(prefsTestKeyFloat))
	ds.Push(core.Float(1.5))
	callPrefs(typ, f, "getFloat")
	if got := float64(ds.Pop().(core.Float)); got != 1.5 {
		t.Fatalf("getFloat missing: got %v, want default 1.5", got)
	}

	ds.Push(core.String(prefsTestKeyString))
	ds.Push(core.String("default"))
	callPrefs(typ, f, "getString")
	if got := string(ds.Pop().(core.String)); got != "default" {
		t.Fatalf("getString missing: got %q, want default %q", got, "default")
	}

	assertListMissing := func(method, key string) {
		t.Helper()
		ds.Push(core.String(key))
		callPrefs(typ, f, method)
		if bool(ds.Pop().(core.Bool)) {
			t.Fatalf("%s missing: got true, want false", method)
		}
	}
	assertListMissing("getIntList", prefsTestKeyIntList)
	assertListMissing("getFloatList", prefsTestKeyFloatList)
	assertListMissing("getStringList", prefsTestKeyStringList)

	if ds.Len() != 0 {
		t.Fatalf("step b leftover DS depth %d", ds.Len())
	}

	// c. stored values different from step-b defaults
	ds.Push(core.Bool(false))
	ds.Push(core.String(prefsTestKeyBool))
	callPrefs(typ, f, "setBool")

	ds.Push(core.Int(7))
	ds.Push(core.String(prefsTestKeyInt))
	callPrefs(typ, f, "setInt")

	ds.Push(core.Float(9.25))
	ds.Push(core.String(prefsTestKeyFloat))
	callPrefs(typ, f, "setFloat")

	ds.Push(core.String("stored"))
	ds.Push(core.String(prefsTestKeyString))
	callPrefs(typ, f, "setString")

	il := core.NewIntList()
	il.Append(1)
	il.Append(2)
	il.Append(3)
	ds.Push(il)
	ds.Push(core.String(prefsTestKeyIntList))
	callPrefs(typ, f, "setIntList")

	fl := core.NewFloatList()
	fl.Append(0.5)
	fl.Append(1.25)
	ds.Push(fl)
	ds.Push(core.String(prefsTestKeyFloatList))
	callPrefs(typ, f, "setFloatList")

	sl := core.NewStringList()
	sl.Append("alpha")
	sl.Append("beta")
	ds.Push(sl)
	ds.Push(core.String(prefsTestKeyStringList))
	callPrefs(typ, f, "setStringList")

	ds.Push(core.String(prefsTestKeyBool))
	ds.Push(core.Bool(true))
	callPrefs(typ, f, "getBool")
	if got := bool(ds.Pop().(core.Bool)); got {
		t.Fatalf("getBool stored: got %v, want false", got)
	}

	ds.Push(core.String(prefsTestKeyInt))
	ds.Push(core.Int(42))
	callPrefs(typ, f, "getInt")
	if got := int(ds.Pop().(core.Int)); got != 7 {
		t.Fatalf("getInt stored: got %d, want 7", got)
	}

	ds.Push(core.String(prefsTestKeyFloat))
	ds.Push(core.Float(1.5))
	callPrefs(typ, f, "getFloat")
	if got := float64(ds.Pop().(core.Float)); got != 9.25 {
		t.Fatalf("getFloat stored: got %v, want 9.25", got)
	}

	ds.Push(core.String(prefsTestKeyString))
	ds.Push(core.String("default"))
	callPrefs(typ, f, "getString")
	if got := string(ds.Pop().(core.String)); got != "stored" {
		t.Fatalf("getString stored: got %q, want %q", got, "stored")
	}

	ds.Push(core.String(prefsTestKeyIntList))
	callPrefs(typ, f, "getIntList")
	if !bool(ds.Pop().(core.Bool)) {
		t.Fatal("getIntList stored: got false")
	}
	gotIL := ds.Pop().(*core.IntList)
	if gotIL.Len() != 3 || gotIL.At(0) != 1 || gotIL.At(1) != 2 || gotIL.At(2) != 3 {
		t.Fatalf("getIntList stored: %v", sliceFromIntList(gotIL))
	}

	ds.Push(core.String(prefsTestKeyFloatList))
	callPrefs(typ, f, "getFloatList")
	if !bool(ds.Pop().(core.Bool)) {
		t.Fatal("getFloatList stored: got false")
	}
	gotFL := ds.Pop().(*core.FloatList)
	if gotFL.Len() != 2 || gotFL.At(0) != 0.5 || gotFL.At(1) != 1.25 {
		t.Fatalf("getFloatList stored: %v", sliceFromFloatList(gotFL))
	}

	ds.Push(core.String(prefsTestKeyStringList))
	callPrefs(typ, f, "getStringList")
	if !bool(ds.Pop().(core.Bool)) {
		t.Fatal("getStringList stored: got false")
	}
	gotSL := ds.Pop().(*core.StringList)
	if gotSL.Len() != 2 || gotSL.At(0) != "alpha" || gotSL.At(1) != "beta" {
		t.Fatalf("getStringList stored: %v", sliceFromStringList(gotSL))
	}

	emptyIL := core.NewIntList()
	ds.Push(emptyIL)
	ds.Push(core.String(prefsTestKeyIntList))
	callPrefs(typ, f, "setIntList")
	ds.Push(core.String(prefsTestKeyIntList))
	callPrefs(typ, f, "getIntList")
	if !bool(ds.Pop().(core.Bool)) {
		t.Fatal("empty IntList should be found")
	}
	if ds.Pop().(*core.IntList).Len() != 0 {
		t.Fatal("empty IntList should have len 0")
	}

	if ds.Len() != 0 {
		t.Fatalf("step c leftover DS depth %d", ds.Len())
	}

	// d. blanket remove; keys missing again
	removePrefsTestKeys(typ, f)

	ds.Push(core.String(prefsTestKeyBool))
	ds.Push(core.Bool(true))
	callPrefs(typ, f, "getBool")
	if got := bool(ds.Pop().(core.Bool)); !got {
		t.Fatalf("after remove getBool: got %v, want default true", got)
	}
	assertListMissing("getIntList", prefsTestKeyIntList)
	assertListMissing("getFloatList", prefsTestKeyFloatList)
	assertListMissing("getStringList", prefsTestKeyStringList)
}
