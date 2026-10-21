package finaltui

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Caller")
	}
	return filepath.Dir(filepath.Dir(file))
}

func buildFinalCmd(t *testing.T) string {
	t.Helper()
	exe := filepath.Join(t.TempDir(), "final")
	cmd := exec.Command("go", "build", "-o", exe, ".")
	cmd.Dir = filepath.Join(repoRoot(t), "cmd", "final")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build cmd/final: %v\n%s", err, out)
	}
	return exe
}

func TestFinalCmdPipeDoesNotRunPositionalFile(t *testing.T) {
	exe := buildFinalCmd(t)
	root := repoRoot(t)
	suspect := filepath.Join(t.TempDir(), "suspect.fnl")
	if err := os.WriteFile(suspect, []byte("this-word-does-not-exist\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, suspect)
	cmd.Dir = root
	cmd.Stdin = strings.NewReader("#!/usr/bin/env final\nsys.params >json $\n")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("run: %v\nstderr: %s\nstdout: %s", err, stderr.String(), stdout.String())
	}
	out := stdout.String()
	if strings.Contains(out, "Final v") || strings.Contains(out, "(ok)") {
		t.Fatalf("script mode leaked REPL chrome: %q", out)
	}
	var params []string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &params); err != nil {
		t.Fatalf("params json = %q: %v", out, err)
	}
	if len(params) != 2 || params[0] != "-" || params[1] != suspect {
		t.Fatalf("params = %v, want [-, %s]", params, suspect)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty (did not run suspect.fnl)", stderr.String())
	}
}
