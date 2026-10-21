package finalgui

import (
	"strings"
	"testing"

	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/internal/core"
)

func TestCompileWorldDemoExample(t *testing.T) {
	core.ResetTestRuntime()
	sess, err := finalembed.NewPersistentSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	if err := Init(sess); err != nil {
		t.Fatalf("Init: %v", err)
	}
	out, err := sess.GetSender().Send(
		"`examples/gui/simgame/world-demo.fnl` compileFromFile",
		defaultSendTimeoutS,
	)
	if err != nil {
		t.Fatalf("compile world-demo.fnl: %v\nout=%q", err, out)
	}
	out, err = sess.GetSender().Send("help WD", defaultSendTimeoutS)
	if err != nil {
		t.Fatalf("help WD: %v", err)
	}
	if strings.Contains(out, "word not found") {
		t.Fatalf("WD type missing after compile: %q", out)
	}
}
