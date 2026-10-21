package finalembed_test

import (
	"strings"
	"testing"

	"github.com/androticus/final/finalembed"
)

func TestEmbedSysArgvParamsEmpty(t *testing.T) {
	sess, err := finalembed.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	out, err := sess.GetSender().Send("sys.argv.len $ nl sys.params.len $ nl", 0)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(out)
	if got != "0\n0" {
		t.Fatalf("argv/params lens = %q, want 0\\n0", out)
	}
}

func TestShebangHashCol1Runs(t *testing.T) {
	sess, err := finalembed.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	src := "#!/usr/bin/env final\n1 2 + $\n"
	out, err := sess.GetSender().Send(src, 0)
	if err != nil {
		t.Fatalf("shebang script: %v (stdout %q)", err, out)
	}
	if strings.TrimSpace(out) != "3" {
		t.Fatalf("stdout = %q, want 3", out)
	}
	if strings.Contains(out, "Final v") || strings.Contains(out, "(ok)") {
		t.Fatalf("unexpected REPL chrome: %q", out)
	}
}
