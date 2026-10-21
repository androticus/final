package finaltui_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/finaltui"
)

// Init loads finaltui/final/*.fnl cwd-relative (repo root or unzipped Final/).
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

func TestInitAndShowDefaultPrompt(t *testing.T) {
	chdirModuleRoot(t)

	sess, err := finalembed.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	if err := finaltui.Init(sess); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if _, err := sess.GetSender().Send(`Tui_ drop`, 30); err != nil {
		t.Fatalf("Tui_ resolve: %v", err)
	}
	if _, err := sess.GetSender().Send(`Dialog.btnOk Dialog.btnCancel drop drop`, 30); err != nil {
		t.Fatalf("Dialog consts: %v", err)
	}

	finaltui.Tui_.ArmUI()
	go func() {
		req := <-finaltui.Tui_.UIRequests()
		if req.Cmd != finaltui.UiCmdPrompt {
			t.Errorf("cmd=%v", req.Cmd)
		}
		if req.Params == nil {
			t.Error("nil params")
		}
		if got := string(req.Params.Get(finaltui.DOFieldTitle.String()).(finalembed.String)); got != "Prompt" {
			t.Errorf("title=%q", got)
		}
		if got := string(req.Params.Get(finaltui.DOFieldBodyText.String()).(finalembed.String)); got != "Sure?" {
			t.Errorf("body=%q", got)
		}
		def := req.Params.Get(finaltui.DOFieldDefaultButton.String()).(finalembed.Int)
		if finaltui.Btn(def) != finaltui.BtnOk {
			t.Errorf("defaultButton=%v", def)
		}
		finaltui.Tui_.Complete(finaltui.CompleteChoice(finaltui.BtnCancel))
	}()

	out, err := sess.GetSender().Send(`"Sure?" Dialog.showDefaultPrompt^ Dialog.btnCancel ==`, 30)
	if err != nil {
		t.Fatalf("Send: %v\nout=%q", err, out)
	}
	// Bool result stays on DS; Send does not require printing. Round-trip already
	// asserted params in the UI mock; CompleteChoice(BtnCancel) is what we sent.
	_ = out
}
