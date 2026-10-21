package finaltui_test

import (
	"testing"

	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/finaltui"
	"github.com/androticus/final/internal/core"
)

func TestEscFnWordsAfterInit(t *testing.T) {
	chdirModuleRoot(t)
	core.ResetTestRuntime()
	core.ClearEscFn()
	sess, err := finalembed.NewPersistentSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	if err := finaltui.Init(sess); err != nil {
		t.Fatalf("Init: %v", err)
	}
	core.LatchEscFn(7)
	out, err := sess.GetSender().Send(
		`: checkEscFn
		   escFn? if
		     7 == if
		       clearEscFn
		       escFn? if drop "stale" else "ok" then
		     else
		       "badn"
		     then
		   else
		     "none"
		   then $
		 ; checkEscFn`,
		30,
	)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if out != "ok" {
		t.Fatalf("got %q, want ok", out)
	}
}
