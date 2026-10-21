package finalgui

import (
	"strings"
	"testing"

	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/internal/core"
)

func TestTypeRegsNonEmpty(t *testing.T) {
	if len(TypeRegs()) == 0 {
		t.Fatal("TypeRegs empty")
	}
}

func TestInitRegistersFinalGui(t *testing.T) {
	core.ResetTestRuntime()
	sess, err := finalembed.NewPersistentSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	if err := Init(sess); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if FinalGui_.typ == nil || FinalGui_.typ.GetMethodMetadata("onSend") == nil {
		t.Fatal("FinalGui#onSend not registered")
	}
	for _, name := range []string{"onLoad", "onQuit", "onWords", "onAbout"} {
		if FinalGui_.typ.GetMethodMetadata(name) == nil {
			t.Fatalf("FinalGui#%s not registered", name)
		}
	}
	for _, name := range []string{"App", "Window", "Label", "Button", "Box", "CanvasImage", "Canvas", "Disableable"} {
		if core.GetTypeRegistrar().Get(name) == nil {
			t.Fatalf("missing fyneui type %q on core catalog", name)
		}
		if _, err := sess.GetSender().Send(name+" drop", defaultSendTimeoutS); err != nil {
			t.Fatalf("core type word %q: %v", name, err)
		}
	}
}

func TestInitAppHelpDocs(t *testing.T) {
	core.ResetTestRuntime()
	sess, err := finalembed.NewPersistentSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	if err := Init(sess); err != nil {
		t.Fatalf("Init: %v", err)
	}
	out, err := sess.GetSender().Send("help App", defaultSendTimeoutS)
	if err != nil {
		t.Fatalf("help App: %v", err)
	}
	if !strings.Contains(out, "Fyne App object") {
		t.Fatalf("help App = %q, want App docs from core3", out)
	}
	if strings.Contains(out, "help: no doc string") || strings.Contains(out, "word not found") {
		t.Fatalf("help App = %q, core3 App :doc missing", out)
	}
}

func TestTakeEventStringParam(t *testing.T) {
	core.ResetTestRuntime()
	rp := core.NewTestRunProvider()
	rp.DataStack.Push(&core.Event{Params: core.ParamsToList(core.String("1 2 + $"))})
	got := takeEventStringParam(&finalembed.RunFrame{RP: rp})
	if got != "1 2 + $" {
		t.Fatalf("src = %q, want %q", got, "1 2 + $")
	}
}
