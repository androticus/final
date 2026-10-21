package core

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/term"
)

func TestSysTypeRegistered(t *testing.T) {
	typ := GetTypeRegistrar().Get("Sys")
	if typ == nil {
		t.Fatal("Sys type not registered")
	}
	if typ != sysType {
		t.Fatal("Sys catalog type is not sysType")
	}
	for _, name := range []string{"argv", "params", "isTty", "coreCount"} {
		if !typ.HasProperty(name) {
			t.Fatalf("Sys missing property %q", name)
		}
		pm := typ.PropertyByName(name)
		if pm == nil || pm.set != nil {
			t.Fatalf("Sys.%s should be readonly", name)
		}
	}
	mm := typ.GetMethodMetadata("exec?")
	if mm == nil {
		t.Fatal("Sys missing method exec?")
	}
	if mm.restrict != SR_System {
		t.Fatalf("Sys.exec? restrict = %#x, want SR_System", mm.restrict)
	}
}

func invokeSysExecQ(t *testing.T, task *Task, program string, args *StringList, timeoutS int) *RunProvider {
	t.Helper()
	if sys == nil {
		t.Fatal("global sys is nil")
	}
	rp := NewTestRunProvider()
	rp.WithTaskFindScope(task)
	f, ths := memberRunFrame(rp, sys)
	f.Task = task
	rp.DataStack.Push(String(program))
	rp.DataStack.Push(args)
	rp.DataStack.Push(Int(timeoutS))
	sysType.GetMethodMetadata("exec?").invokeCall(f, ths)
	return rp
}

func TestSysExecQSuccess(t *testing.T) {
	InitTestRuntime()
	task := NewTestTask()
	rp := invokeSysExecQ(t, task, "echo", NewStringListFromSlice([]string{"hello-exec"}), 0)
	ok := rp.DataStack.Pop().(Bool)
	code := rp.DataStack.Pop().(Int)
	out := string(rp.DataStack.Pop().(String))
	if !ok {
		t.Fatalf("exec? modal false, out/err path unexpected")
	}
	if code != 0 {
		t.Fatalf("code = %d, want 0; out=%q", code, out)
	}
	if !strings.Contains(out, "hello-exec") {
		t.Fatalf("out = %q, want to contain hello-exec", out)
	}
}

func TestSysExecQNonzeroExit(t *testing.T) {
	InitTestRuntime()
	task := NewTestTask()
	rp := invokeSysExecQ(t, task, "false", NewStringList(), 0)
	ok := rp.DataStack.Pop().(Bool)
	code := rp.DataStack.Pop().(Int)
	_ = rp.DataStack.Pop().(String) // may be empty
	if !ok {
		t.Fatal("nonzero exit should still be modal true")
	}
	if code == 0 {
		t.Fatal("code = 0, want nonzero from false")
	}
}

func TestSysExecQLaunchFailure(t *testing.T) {
	InitTestRuntime()
	task := NewTestTask()
	rp := invokeSysExecQ(t, task, "/no/such/final-exec-binary-xyz", NewStringList(), 0)
	ok := rp.DataStack.Pop().(Bool)
	errStr := string(rp.DataStack.Pop().(String))
	_ = string(rp.DataStack.Pop().(String)) // out, usually empty
	if ok {
		t.Fatal("missing binary should be modal false")
	}
	if errStr == "" {
		t.Fatal("expected launch error string")
	}
}

func TestSysExecQTimeout(t *testing.T) {
	InitTestRuntime()
	task := NewTestTask()
	// Single process (not sh -c): CommandContext kill must hit the sleeper itself.
	// Python prints then sleeps in-process so we can assert partial out.
	rp := invokeSysExecQ(t, task, "python3", NewStringListFromSlice([]string{
		"-c", "import sys,time; sys.stdout.write('partial-out'); sys.stdout.flush(); time.sleep(5)",
	}), 1)
	ok := rp.DataStack.Pop().(Bool)
	errStr := string(rp.DataStack.Pop().(String))
	out := string(rp.DataStack.Pop().(String))
	if ok {
		t.Fatal("timeout should be modal false")
	}
	if errStr != sysExecTimeoutErr {
		t.Fatalf("err = %q, want %q", errStr, sysExecTimeoutErr)
	}
	if !strings.Contains(out, "partial-out") {
		t.Fatalf("out = %q, want to contain partial-out", out)
	}
}

func TestSysExecQRespectsSRSystem(t *testing.T) {
	InitTestRuntime()
	task := NewTestTask()
	task.restrict = SR_System
	rp := NewTestRunProvider()
	rp.WithTaskFindScope(task)
	f, ths := memberRunFrame(rp, sys)
	f.Task = task
	rp.DataStack.Push(String("echo"))
	rp.DataStack.Push(NewStringList())
	rp.DataStack.Push(Int(0))
	defer func() {
		r := recover()
		if r != "operation not allowed" {
			t.Fatalf("recover = %v, want operation not allowed", r)
		}
	}()
	sysType.GetMethodMetadata("exec?").invokeCall(f, ths)
	t.Fatal("expected SR_System panic")
}

func TestOpGetSysPushesGlobal(t *testing.T) {
	InitTestRuntime()
	if sys == nil {
		t.Fatal("global sys is nil")
	}
	rp := NewTestRunProvider()
	RunOps(NewTestTask(), rp, OpGetSys, OpStop)
	got, ok := rp.DataStack.Pop().(*Sys)
	if !ok || got != sys {
		t.Fatalf("sys axiom = %v, want global %v", got, sys)
	}
}

func TestSysArgvParamsEmpty(t *testing.T) {
	if sys == nil {
		t.Fatal("global sys is nil")
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, sys)
	sysType.PropertyByName("argv").invokeGet(f, ths)
	argv := rp.DataStack.Pop().(*StringList)
	if argv.Len() != 0 {
		t.Fatalf("sys.argv len = %d, want 0", argv.Len())
	}
	f, ths = memberRunFrame(rp, sys)
	sysType.PropertyByName("params").invokeGet(f, ths)
	params := rp.DataStack.Pop().(*StringList)
	if params.Len() != 0 {
		t.Fatalf("sys.params len = %d, want 0", params.Len())
	}
}

func TestSysIsTtySnapshot(t *testing.T) {
	if sys == nil {
		t.Fatal("global sys is nil")
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, sys)
	sysType.PropertyByName("isTty").invokeGet(f, ths)
	gotTty, ok := rp.DataStack.Pop().(Bool)
	if !ok {
		t.Fatalf("sys.isTty type %T, want Bool", rp.DataStack.Peek())
	}
	wantTty := Bool(term.IsTerminal(int(os.Stdin.Fd())))
	if gotTty != wantTty {
		t.Fatalf("sys.isTty = %v, want %v", gotTty, wantTty)
	}
}

func TestSysCoreCount(t *testing.T) {
	if sys == nil {
		t.Fatal("global sys is nil")
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, sys)
	sysType.PropertyByName("coreCount").invokeGet(f, ths)
	got, ok := rp.DataStack.Pop().(Int)
	if !ok {
		t.Fatalf("sys.coreCount type %T, want Int", rp.DataStack.Peek())
	}
	want := Int(runtime.NumCPU())
	if got != want {
		t.Fatalf("sys.coreCount = %v, want %v", got, want)
	}
	if got < 1 {
		t.Fatalf("sys.coreCount = %v, want >= 1", got)
	}
}
