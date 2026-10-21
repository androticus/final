package core

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestSessionSendAdd(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	out, err := s.Send("1 2 +", 0)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if out != "" {
		t.Fatalf("output = %q, want empty", out)
	}
}

func TestSessionSendWithOutput(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	out, err := s.Send("1 2 ds$", 0)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if out != "1 2 " {
		t.Fatalf("output = %q, want %q", out, "1 2 ")
	}
}

func TestSessionSendDictionaryPersists(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if _, err := s.Send(": double 2 * ;", 0); err != nil {
		t.Fatalf("define: %v", err)
	}
	if _, err := s.Send("4 double", 0); err != nil {
		t.Fatalf("run double: %v", err)
	}
}

func TestSessionSendInterpretError(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	_, err = s.Send("nosuchword", 0)
	if err == nil {
		t.Fatal("expected interpret error")
	}
}

func TestSessionSendFault(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	_, err = s.Send("1 0 /", 0)
	if err == nil {
		t.Fatal("expected fault")
	}
}

// compileFromFile must panic on failure so Session.Send returns err (not silent null-stderr).
func TestSessionSendCompileFromFileFault(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	path := writeTempFinalSource(t, ": badfile nosuchword ;\n")
	script := "`" + filepath.Clean(path) + "` compileFromFile"
	_, err = s.Send(script, 0)
	if err == nil {
		t.Fatal("expected fault from compileFromFile")
	}
	if _, ok := err.(*RunFault); !ok {
		t.Fatalf("err = %v (%T), want *RunFault", err, err)
	}
}

func TestSessionSendEmptyScript(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	_, err = s.Send("", 0)
	if err != ErrEmptyScript {
		t.Fatalf("err = %v, want ErrEmptyScript", err)
	}
}

func TestSessionSendClosed(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = s.Send("1", 0)
	if err != ErrSessionClosed {
		t.Fatalf("err = %v, want ErrSessionClosed", err)
	}
}

func TestSessionSendConcurrentSerializes(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range 2 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, errs[idx] = s.Send("1 2 +", 0)
		}(i)
	}
	wg.Wait()
	for i, e := range errs {
		if e != nil {
			t.Fatalf("goroutine %d: %v", i, e)
		}
	}
}

func TestSessionSendStacksEphemeral(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if _, err := s.Send("1", 0); err != nil {
		t.Fatalf("push: %v", err)
	}
	out, err := s.Send("ds$", 0)
	if err != nil {
		t.Fatalf("ds$: %v", err)
	}
	if out != "" {
		t.Fatalf("ephemeral ds$ = %q, want empty (fresh stack)", out)
	}
}

func TestSessionSendStacksPersistent(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if _, err := s.Send("1", 0); err != nil {
		t.Fatalf("push: %v", err)
	}
	if _, err := s.Send("dup +", 0); err != nil {
		t.Fatalf("dup +: %v", err)
	}
	out, err := s.Send("ds$", 0)
	if err != nil {
		t.Fatalf("ds$: %v", err)
	}
	if out != "2 " {
		t.Fatalf("persistent ds$ = %q, want %q", out, "2 ")
	}
}

func TestSessionSendIdentRootedConfigStore(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if _, err := s.Send("1 sess.config.paramA!", 0); err != nil {
		t.Fatalf("ident-rooted store: %v", err)
	}
	got := s.config.Get("paramA")
	if got != Int(1) {
		t.Fatalf("config.paramA = %v, want 1", got)
	}
	out, err := s.Send("ds$", 0)
	if err != nil {
		t.Fatalf("ds$: %v", err)
	}
	if out != "" {
		t.Fatalf("ds$ after store = %q, want empty (root elided)", out)
	}

	if _, err := s.Send(": setB 2 sess.config.paramB! ; setB", 0); err != nil {
		t.Fatalf("compiled ident-rooted store: %v", err)
	}
	if got := s.config.Get("paramB"); got != Int(2) {
		t.Fatalf("config.paramB = %v, want 2", got)
	}

	if _, err := s.Send(`"cfg" sess.configJson!`, 0); err != nil {
		t.Fatalf("configJson store: %v", err)
	}
	if s.configJson != "cfg" {
		t.Fatalf("configJson = %q, want cfg", s.configJson)
	}

	if _, err := s.Send(`3 "paramC" sess.config!{}`, 0); err != nil {
		t.Fatalf("ident-rooted !{}: %v", err)
	}
	if got := s.config.Get("paramC"); got != Int(3) {
		t.Fatalf("config.paramC = %v, want 3", got)
	}
}

func TestSessionSendConfigPropertyReadonly(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	_, err = s.Send("DynamicObject.new^ sess.config!", 0)
	if err == nil {
		t.Fatal("expected error replacing readonly sess.config")
	}
	if !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("err = %v, want read-only", err)
	}
}

func TestSetStdoutStandingAndInterpretCapture(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	log := NewLineBufferTextWriterSync(8)
	s.SetStdout(log)
	s.rp.Stdout.WriteString("app\n")
	if got := log.ToString(); got != "app\n" {
		t.Fatalf("standing write = %q, want %q", got, "app\n")
	}

	out, err := InterpretSource(s.task, s.rp, "1 2 + $")
	if err != nil {
		t.Fatal(err)
	}
	if out != "3" {
		t.Fatalf("capture = %q, want 3", out)
	}
	if got := log.ToString(); got != "app\n" {
		t.Fatalf("standing during capture = %q, want %q", got, "app\n")
	}

	s.rp.Stdout.WriteString(out)
	if got := log.ToString(); got != "app\n3" {
		t.Fatalf("after replay = %q, want %q", got, "app\n3")
	}
}
