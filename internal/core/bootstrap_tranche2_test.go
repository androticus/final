package core

import (
	"testing"
)

func TestParseBootstrapManifestLines(t *testing.T) {
	got := parseBootstrapManifestLines(`
// comment
internal/final/core/a.fnl

internal/final/core/b.fnl
// trailing comment
`)
	want := []string{
		"internal/final/core/a.fnl",
		"internal/final/core/b.fnl",
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestRunBootstrapTranche2RollbackOnError(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	dictBefore := len(globalDictionaryContext.dict.byOrdinal)

	manifest := `
` + writeTempFinalSource(t, ": good ;") + `
` + writeTempFinalSource(t, ": bad nosuchword ;") + `
`
	err := runBootstrapTranche2Manifest(task, rp, manifest)
	if err == nil {
		t.Fatal("expected compile error")
	}
	if globalDictionaryContext.dict.GetByName("good") == nil {
		t.Fatal("good should remain after first file compiled successfully")
	}
	if globalDictionaryContext.dict.GetByName("bad") != nil {
		t.Fatal("bad should be rolled back after bad file fails")
	}
	if len(globalDictionaryContext.dict.byOrdinal) != dictBefore+1 {
		t.Fatalf("dict len = %d, want %d (good committed, bad rolled back)", len(globalDictionaryContext.dict.byOrdinal), dictBefore+1)
	}
}
