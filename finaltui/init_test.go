package finaltui_test

import (
	"testing"

	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/finaltui"
	"github.com/androticus/final/internal/core"
)

func TestInitCompilesTuiRoot(t *testing.T) {
	chdirModuleRoot(t)
	core.ResetTestRuntime()
	sess, err := finalembed.NewPersistentSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	if err := finaltui.Init(sess); err != nil {
		t.Fatalf("Init: %v", err)
	}
	out, err := sess.GetSender().Send(
		`: checkTuiRoot "tuiRoot" find? if drop "ok" else "missing" then $ ; checkTuiRoot`,
		30,
	)
	if err != nil {
		t.Fatalf("probe tuiRoot: %v", err)
	}
	if out != "ok" {
		t.Fatalf("tuiRoot after Init: got %q, want ok (const %q)", out, finaltui.TuiRootWord)
	}
}
