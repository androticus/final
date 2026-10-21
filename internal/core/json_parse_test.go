package core

// JSON parse integration tests deferred until 0.0.4 Final-native test harness (not designed yet).
// Requires parse.fnl + lex.fnl + session.fnl in bootstrapTranche2Manifest.
/*
import (
	"path/filepath"
	"testing"
)

func compileJsonParseTestStack(t *testing.T, task *Task, rp *RunProvider) {
	t.Helper()
	cs := task.compilerForTarget()
	if GetCompileState(cs) {
		SetCompileState(cs, false)
	}
	globalDictionaryContext.dict.ClearBuilding()
	path := filepath.Join("internal", "final", "test", "json_parse_test.fnl")
	if err := compileBootstrapFile(task, rp, path); err != nil {
		t.Fatalf("compile %s: %v", path, err)
	}
}

func TestJsonParseFinalSource(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	compileJsonParseTestStack(t, task, rp)
	runDeferWordByName(t, task, rp, "jsonParseTests")
}
*/
