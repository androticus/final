package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite" // registers "sqlite" driver (cznic / modernc, cgo-free)
)

func testEnvType(t *testing.T) *Type {
	t.Helper()
	typ := GetTypeRegistrar().Get("Env")
	if typ == nil {
		t.Fatal("Env type not registered")
	}
	return typ
}

func invokeEnvGet(t *testing.T, key string) *RunProvider {
	t.Helper()
	mm := testEnvType(t).ClassMethodByName("get")
	if mm == nil {
		t.Fatal("missing Env.get class method")
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, testEnvType(t))
	rp.DataStack.Push(String(key))
	mm.invokeCall(f, ths)
	return rp
}

func writeDotEnv(t *testing.T, dir, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestParseDotEnvGrammar(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	body := "" +
		"\n" +
		"# comment\n" +
		"// also comment\n" +
		"  SIMPLE = hello  \n" +
		"export EXPORTED=from-export\n" +
		"QUOTED=\"  spaced  \"\n" +
		"EMPTY=\n" +
		"INTERIOR=My Val\n" +
		"CASE=lower\n" +
		"case=upper-wins-not\n" // different key (case-sensitive)
	writeDotEnv(t, dir, body)

	m := parseDotEnvFile(".env")
	want := map[string]string{
		"SIMPLE":   "hello",
		"EXPORTED": "from-export",
		"QUOTED":   "  spaced  ",
		"EMPTY":    "",
		"INTERIOR": "My Val",
		"CASE":     "lower",
		"case":     "upper-wins-not",
	}
	if len(m) != len(want) {
		t.Fatalf("map len = %d, want %d; got %#v", len(m), len(want), m)
	}
	for k, v := range want {
		if m[k] != v {
			t.Fatalf("m[%q] = %q, want %q", k, m[k], v)
		}
	}
}

func TestParseDotEnvMissingFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	m := parseDotEnvFile(".env")
	if len(m) != 0 {
		t.Fatalf("missing .env should yield empty map, got %#v", m)
	}
}

func TestParseDotEnvStopsAtFirstDuplicate(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeDotEnv(t, dir, "A=1\nB=2\nA=3\nC=4\n")
	m := parseDotEnvFile(".env")
	if m["A"] != "1" {
		t.Fatalf("A = %q, want first binding 1", m["A"])
	}
	if m["B"] != "2" {
		t.Fatalf("B = %q, want 2", m["B"])
	}
	if _, ok := m["C"]; ok {
		t.Fatal("C should not be parsed after duplicate A")
	}
}

func TestParseDotEnvSkipsMalformed(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeDotEnv(t, dir, "NOEQ\n=novaluekey\nOK=yes\n")
	m := parseDotEnvFile(".env")
	if len(m) != 1 || m["OK"] != "yes" {
		t.Fatalf("got %#v, want only OK=yes", m)
	}
}

func TestParseDotEnvCRLF(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeDotEnv(t, dir, "A=1\r\nB=2\r\n")
	m := parseDotEnvFile(".env")
	if m["A"] != "1" || m["B"] != "2" {
		t.Fatalf("got %#v", m)
	}
}

func TestEnvLookupDotEnvBeatsProcess(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	const key = "FINAL_ENV_TEST_OVERRIDE"
	t.Setenv(key, "from-process")
	writeDotEnv(t, dir, key+"=from-dotenv\n")

	val, ok := envLookup(key)
	if !ok || val != "from-dotenv" {
		t.Fatalf("got %q ok=%v, want from-dotenv", val, ok)
	}
}

func TestEnvLookupFallsBackToProcess(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	const key = "FINAL_ENV_TEST_PROCESS_ONLY"
	t.Setenv(key, "proc-val")

	val, ok := envLookup(key)
	if !ok || val != "proc-val" {
		t.Fatalf("got %q ok=%v, want proc-val", val, ok)
	}
}

func TestEnvLookupNotFound(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	const key = "FINAL_ENV_TEST_MISSING_XYZ_9f3a"
	os.Unsetenv(key)

	if _, ok := envLookup(key); ok {
		t.Fatal("expected not found")
	}
}

func TestEnvLookupEmptyProcessValueIsFound(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	const key = "FINAL_ENV_TEST_EMPTY"
	t.Setenv(key, "")

	val, ok := envLookup(key)
	if !ok || val != "" {
		t.Fatalf("empty process value should be found; got %q ok=%v", val, ok)
	}
}

func TestEnvGetClassMethodModal(t *testing.T) {
	ResetTestRuntime()
	dir := t.TempDir()
	t.Chdir(dir)
	const key = "FINAL_ENV_TEST_CLASS"
	writeDotEnv(t, dir, key+"=via-class\n")

	rp := invokeEnvGet(t, key)
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want 2 (val true)", rp.DataStack.Len())
	}
	if rp.DataStack.Pop() != Bool(true) {
		t.Fatal("want true")
	}
	if got := rp.DataStack.Pop().(String); got != String("via-class") {
		t.Fatalf("val = %q", got)
	}

	rp = invokeEnvGet(t, "FINAL_ENV_TEST_NO_SUCH_KEY_zzz")
	if rp.DataStack.Len() != 1 || rp.DataStack.Pop() != Bool(false) {
		t.Fatal("want false alone when missing")
	}
}

func TestEnvGetRereadsDotEnv(t *testing.T) {
	ResetTestRuntime()
	dir := t.TempDir()
	t.Chdir(dir)
	const key = "FINAL_ENV_TEST_REREAD"
	writeDotEnv(t, dir, key+"=v1\n")
	rp := invokeEnvGet(t, key)
	rp.DataStack.Pop() // true
	if got := rp.DataStack.Pop().(String); got != String("v1") {
		t.Fatalf("first = %q", got)
	}
	writeDotEnv(t, dir, key+"=v2\n")
	rp = invokeEnvGet(t, key)
	rp.DataStack.Pop()
	if got := rp.DataStack.Pop().(String); got != String("v2") {
		t.Fatalf("second = %q, want v2 (no cache)", got)
	}
}

func testMonikerType(t *testing.T) *Type {
	t.Helper()
	typ := GetTypeRegistrar().Get("Moniker")
	if typ == nil {
		t.Fatal("Moniker type not registered")
	}
	return typ
}

func invokeMonikerNew(t *testing.T, init string) *Moniker {
	t.Helper()
	mm := testMonikerType(t).ClassMethodByName("new")
	if mm == nil {
		t.Fatal("missing Moniker.new")
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, testMonikerType(t))
	rp.DataStack.Push(String(init))
	mm.invokeCall(f, ths)
	return rp.DataStack.Pop().(*Moniker)
}

func invokeEnvSubMoniker(t *testing.T, mon *Moniker, key string) Bool {
	t.Helper()
	mm := testEnvType(t).ClassMethodByName("subMoniker")
	if mm == nil {
		t.Fatal("missing Env.subMoniker")
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, testEnvType(t))
	rp.DataStack.Push(mon)
	rp.DataStack.Push(String(key))
	mm.invokeCall(f, ths)
	return rp.DataStack.Pop().(Bool)
}

func TestMonikerNewAndOpacity(t *testing.T) {
	ResetTestRuntime()
	mon := invokeMonikerNew(t, "secret-dsn")
	if mon.Value() != "secret-dsn" {
		t.Fatalf("Value = %q", mon.Value())
	}
	if mon.NaturalString() != "Moniker(<opaque>)" {
		t.Fatalf("NaturalString = %q", mon.NaturalString())
	}
	if strings.Contains(mon.NaturalString(), "secret") {
		t.Fatal("NaturalString must not leak value")
	}
}

func TestEnvSubMonikerSubstitutesAllAndFlag(t *testing.T) {
	ResetTestRuntime()
	dir := t.TempDir()
	t.Chdir(dir)
	const key = "FINAL_ENV_TEST_SUB"
	writeDotEnv(t, dir, key+"=XYZ\n")

	mon := invokeMonikerNew(t, "a=$FINAL_ENV_TEST_SUB;b=$FINAL_ENV_TEST_SUB")
	if invokeEnvSubMoniker(t, mon, key) != Bool(true) {
		t.Fatal("want true when key found")
	}
	if mon.Value() != "a=XYZ;b=XYZ" {
		t.Fatalf("Value = %q, want both placeholders replaced", mon.Value())
	}

	before := mon.Value()
	if invokeEnvSubMoniker(t, mon, "FINAL_ENV_TEST_MISSING_KEY") != Bool(false) {
		t.Fatal("want false when key missing")
	}
	if mon.Value() != before {
		t.Fatal("moniker must be unchanged on miss")
	}
}

func TestDsnFromStackCellStringOrMoniker(t *testing.T) {
	if got := dsnFromStackCell(String("plain")); got != "plain" {
		t.Fatalf("String = %q", got)
	}
	if got := dsnFromStackCell(NewMoniker("from-mon")); got != "from-mon" {
		t.Fatalf("Moniker = %q", got)
	}
}

func TestDbProviderNewAcceptsMonikerDSN(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	f := &RunFrame{Task: task, RP: rp}
	mm := GetTypeRegistrar().Get("DbProvider").GetClassMethodMetadata("new")
	if mm == nil {
		t.Fatal("missing DbProvider.new")
	}
	rp.DataStack.Push(String("sqlite"))
	rp.DataStack.Push(NewMoniker(":memory:"))
	mm.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("DbProvider")))
	ok := rp.DataStack.Pop().(Bool)
	if !ok {
		err := rp.DataStack.Pop().(String)
		t.Fatalf("DbProvider.new with Moniker failed: %s", err)
	}
	if _, ok := rp.DataStack.Pop().(*DbProvider); !ok {
		t.Fatal("want *DbProvider")
	}
}
