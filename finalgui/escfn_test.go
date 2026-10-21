package finalgui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/internal/core"
	"fyne.io/fyne/v2"
)

func chdirModuleRoot(t *testing.T) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := orig
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if err := os.Chdir(dir); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chdir(orig) })
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found above %s", orig)
		}
		dir = parent
	}
}

func TestEscFnCodeFromFyneKeys(t *testing.T) {
	keys := []fyne.KeyName{
		fyne.KeyEscape,
		fyne.KeyF1, fyne.KeyF2, fyne.KeyF3, fyne.KeyF4,
		fyne.KeyF5, fyne.KeyF6, fyne.KeyF7, fyne.KeyF8,
		fyne.KeyF9, fyne.KeyF10, fyne.KeyF11, fyne.KeyF12,
	}
	if len(keys) != 13 {
		t.Fatalf("key table len %d, want 13", len(keys))
	}
	for i, kn := range keys {
		n, ok := escFnCodeFromFyneKey(kn)
		if !ok || n != i {
			t.Fatalf("%v -> %d %v, want %d true", kn, n, ok, i)
		}
	}
	if _, ok := escFnCodeFromFyneKey(fyne.KeyA); ok {
		t.Fatal("KeyA must not map")
	}
}

func TestLatchEscFnKey(t *testing.T) {
	core.ClearEscFn()
	if !latchEscFnKey(&fyne.KeyEvent{Name: fyne.KeyEscape}) {
		t.Fatal("expected latch Esc")
	}
	n, ok := core.TakeEscFn()
	if !ok || n != 0 {
		t.Fatalf("TakeEscFn = %d %v, want 0 true", n, ok)
	}
	if latchEscFnKey(&fyne.KeyEvent{Name: fyne.KeyA}) {
		t.Fatal("KeyA must not latch")
	}
}

func TestEscFnWordsAfterInit(t *testing.T) {
	chdirModuleRoot(t)
	core.ResetTestRuntime()
	core.ClearEscFn()
	sess, err := finalembed.NewPersistentSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	if err := Init(sess); err != nil {
		t.Fatalf("Init: %v", err)
	}
	core.LatchEscFn(0)
	out, err := sess.GetSender().Send(
		`: checkEscFn
		   escFn? if
		     0 == if
		       escFn? if drop "stale" else "ok" then
		     else
		       "badn"
		     then
		   else
		     "none"
		   then $
		 ; checkEscFn`,
		defaultSendTimeoutS,
	)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if out != "ok" {
		t.Fatalf("got %q, want ok", out)
	}
}
