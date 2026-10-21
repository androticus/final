package core

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func waitPipUnregistered(t *testing.T, parent *Task, id int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for parent.hasPip(id) {
		if time.Now().After(deadline) {
			t.Fatal("timeout waiting for child pip unregister")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestSpawnPipHelloStdout(t *testing.T) {
	ResetTestRuntime()
	parent := CreateLocalTask()
	parent.restrict = SR_FileRead | SR_HttpDial
	rp := NewTestRunProvider()
	sb := NewStringBuilder()
	rp.Stdout = sb
	tr, err := NewTextReaderFromString(`: helloPip drop "Hello from pip!\n" $ ;`)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	rp.WithTaskFindScope(parent)
	if err := interpretSession(parent, rp); err != nil {
		t.Fatalf("compile helloPip: %v", err)
	}
	w := parent.dc.dict.GetByName("helloPip")
	if w == nil {
		t.Fatal("missing helloPip")
	}

	mm := pipType.ClassMethodByName("newPipParams")
	if mm == nil {
		t.Fatal("Pip.newPipParams missing")
	}
	f, ths := memberRunFrame(rp, pipType)
	mm.invokeCall(f, ths)
	pp, ok := rp.DataStack.Pop().(*AtomicDynamicObject)
	if !ok || pp == nil {
		t.Fatalf("newPipParams = %T, want *AtomicDynamicObject", rp.DataStack.Peek())
	}
	if pp.Get("_stdout") != sb {
		t.Fatal("PipParams._stdout should be parent RP.Stdout")
	}
	if pp.Get("_stderr") != rp.Stderr {
		t.Fatal("PipParams._stderr should be parent RP.Stderr")
	}

	pip := spawnPip(parent, w.XT(), pp)
	if pip == nil || pip.ID() < 1 {
		t.Fatalf("pip = %+v, want id >= 1", pip)
	}
	if !parent.hasPip(pip.ID()) {
		t.Fatal("pip should be registered after spawn")
	}
	wantName := "helloPip#" + strconv.Itoa(pip.ID())
	if got := pip.Task().name; got != wantName {
		t.Fatalf("Task.name = %q, want %q", got, wantName)
	}
	if got := pip.Task().restrict; got != parent.restrict {
		t.Fatalf("child.restrict = %#x, want %#x", got, parent.restrict)
	}

	waitPipUnregistered(t, parent, pip.ID())
	if got := sb.ToString(); got != "Hello from pip!\n" {
		t.Fatalf("stdout = %q, want %q", got, "Hello from pip!\n")
	}
}

func TestSpawnPipNameOverride(t *testing.T) {
	ResetTestRuntime()
	parent := CreateLocalTask()
	rp := NewTestRunProvider()

	tr, err := NewTextReaderFromString(`: namedPip drop ;`)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	rp.WithTaskFindScope(parent)
	if err := interpretSession(parent, rp); err != nil {
		t.Fatalf("compile: %v", err)
	}
	w := parent.dc.dict.GetByName("namedPip")
	if w == nil {
		t.Fatal("missing namedPip")
	}
	pp := pipParamsFromRP(t, rp)
	pp.Set("_name", String("worker"))

	pip := spawnPip(parent, w.XT(), pp)
	wantName := "worker#" + strconv.Itoa(pip.ID())
	if got := pip.Task().name; got != wantName {
		t.Fatalf("Task.name = %q, want %q", got, wantName)
	}
	waitPipUnregistered(t, parent, pip.ID())
}

func TestSpawnPipNameEmptyFallsBackToWord(t *testing.T) {
	ResetTestRuntime()
	parent := CreateLocalTask()
	rp := NewTestRunProvider()

	tr, err := NewTextReaderFromString(`: emptyNamePip drop ;`)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	rp.WithTaskFindScope(parent)
	if err := interpretSession(parent, rp); err != nil {
		t.Fatalf("compile: %v", err)
	}
	w := parent.dc.dict.GetByName("emptyNamePip")
	if w == nil {
		t.Fatal("missing emptyNamePip")
	}
	pp := pipParamsFromRP(t, rp)
	pp.Set("_name", String(""))

	pip := spawnPip(parent, w.XT(), pp)
	wantName := "emptyNamePip#" + strconv.Itoa(pip.ID())
	if got := pip.Task().name; got != wantName {
		t.Fatalf("Task.name = %q, want %q", got, wantName)
	}
	waitPipUnregistered(t, parent, pip.ID())
}

func TestOpSpawnPipHelloStdout(t *testing.T) {
	ResetTestRuntime()
	parent := CreateLocalTask()
	rp := NewTestRunProvider()
	sb := NewStringBuilder()
	rp.Stdout = sb
	tr, err := NewTextReaderFromString(`: helloPip drop "Hello from pip!\n" $ ;`)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	rp.WithTaskFindScope(parent)
	if err := interpretSession(parent, rp); err != nil {
		t.Fatalf("compile helloPip: %v", err)
	}
	w := parent.dc.dict.GetByName("helloPip")
	if w == nil {
		t.Fatal("missing helloPip")
	}

	mm := pipType.ClassMethodByName("newPipParams")
	f, ths := memberRunFrame(rp, pipType)
	mm.invokeCall(f, ths)
	pp := rp.DataStack.Pop().(*AtomicDynamicObject)

	rp.DataStack.Push(pp)
	rp.DataStack.Push(w.XT())
	RunOps(parent, rp, OpSpawnPip, OpStop)
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0 after (spawnPip)", rp.DataStack.Len())
	}
	waitPipUnregistered(t, parent, parent.nextPipID)
	if got := sb.ToString(); got != "Hello from pip!\n" {
		t.Fatalf("stdout = %q, want %q", got, "Hello from pip!\n")
	}
	ax := globalDictionaryContext.dict.GetByName("(spawnPip)")
	if ax == nil || ax.opCode != OpSpawnPip {
		t.Fatal("missing axiom (spawnPip)")
	}
}

func TestSpawnPipRequiresStdout(t *testing.T) {
	ResetTestRuntime()
	parent := CreateLocalTask()
	pp := NewAtomicDynamicObject()
	pp.Set("_stderr", NewRunProvider().Stderr)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic without ._stdout")
		}
	}()
	spawnPip(parent, XT(1), pp)
}

func TestSpawnPipRequiresStderr(t *testing.T) {
	ResetTestRuntime()
	parent := CreateLocalTask()
	pp := NewAtomicDynamicObject()
	pp.Set("_stdout", NewRunProvider().Stdout)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic without ._stderr")
		}
	}()
	spawnPip(parent, XT(1), pp)
}

func pipParamsFromRP(t *testing.T, rp *RunProvider) *AtomicDynamicObject {
	t.Helper()
	mm := pipType.ClassMethodByName("newPipParams")
	if mm == nil {
		t.Fatal("Pip.newPipParams missing")
	}
	f, ths := memberRunFrame(rp, pipType)
	mm.invokeCall(f, ths)
	pp, ok := rp.DataStack.Pop().(*AtomicDynamicObject)
	if !ok || pp == nil {
		t.Fatalf("newPipParams = %T", rp.DataStack.Peek())
	}
	return pp
}

func TestNestedSpawnKeepsDCLockedUntilGrandchildDone(t *testing.T) {
	ResetTestRuntime()
	root := CreateLocalTask()
	rp := NewTestRunProvider()
	sb := NewStringBuilder()
	rp.Stdout = sb
	src := `
: hangGrand drop begin again ;
: midPip
  drop
  Pip.newPipParams^ spawn: hangGrand
;
`
	tr, err := NewTextReaderFromString(src)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	rp.WithTaskFindScope(root)
	if err := interpretSession(root, rp); err != nil {
		t.Fatalf("compile: %v", err)
	}
	midW := root.dc.dict.GetByName("midPip")
	if midW == nil {
		t.Fatal("missing midPip")
	}

	pp := pipParamsFromRP(t, rp)
	midPip := spawnPip(root, midW.XT(), pp)
	midID := midPip.ID()

	deadline := time.Now().Add(2 * time.Second)
	for root.descendantCount.Load() < 2 {
		if time.Now().After(deadline) {
			t.Fatalf("descendantCount = %d, want >= 2", root.descendantCount.Load())
		}
		time.Sleep(time.Millisecond)
	}
	if !root.dc.IsLockedOrSealed() {
		t.Fatal("DC should be locked while grandchild lives")
	}

	// Mid's Run returns but stays registered while grandchild hangs.
	deadline = time.Now().Add(2 * time.Second)
	for {
		midTask := midPip.Task()
		midTask.childrenMu.Lock()
		finished := midTask.runFinished
		midTask.childrenMu.Unlock()
		if finished && root.hasPip(midID) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timeout waiting for mid Run finish while still registered")
		}
		time.Sleep(time.Millisecond)
	}
	if !root.dc.IsLockedOrSealed() {
		t.Fatal("DC must stay locked after mid exits with live grandchild")
	}
	if root.descendantCount.Load() < 1 {
		t.Fatal("root descendantCount should remain > 0")
	}

	// Break the hanging grandchild via root fanout.
	root.propagateFaultDown()
	waitPipUnregistered(t, root, midID)
	deadline = time.Now().Add(2 * time.Second)
	for root.descendantCount.Load() != 0 || root.dc.IsLockedOrSealed() {
		if time.Now().After(deadline) {
			t.Fatalf("after reap: count=%d locked=%v", root.descendantCount.Load(), root.dc.IsLockedOrSealed())
		}
		time.Sleep(time.Millisecond)
	}
}

func TestPipFaultBreaksSibling(t *testing.T) {
	ResetTestRuntime()
	root := CreateLocalTask()
	rp := NewTestRunProvider()
	sb := NewStringBuilder()
	rp.Stdout = sb
	src := `
: hangPip drop begin again ;
: boomPip drop 1 0 / ;
`
	tr, err := NewTextReaderFromString(src)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	rp.WithTaskFindScope(root)
	if err := interpretSession(root, rp); err != nil {
		t.Fatalf("compile: %v", err)
	}
	hangW := root.dc.dict.GetByName("hangPip")
	boomW := root.dc.dict.GetByName("boomPip")
	if hangW == nil || boomW == nil {
		t.Fatal("missing hangPip/boomPip")
	}

	hangPip := spawnPip(root, hangW.XT(), pipParamsFromRP(t, rp))
	boomPip := spawnPip(root, boomW.XT(), pipParamsFromRP(t, rp))
	hangID, boomID := hangPip.ID(), boomPip.ID()

	waitPipUnregistered(t, root, boomID)
	waitPipUnregistered(t, root, hangID)

	if root.sessionFault == nil {
		t.Fatal("root should store sessionFault from boomPip")
	}
	if root.sessionFaultPipId != boomID {
		t.Fatalf("sessionFaultPipId = %d, want boom %d", root.sessionFaultPipId, boomID)
	}
	report := sb.ToString()
	if !strings.Contains(report, "Fatal Error:") {
		t.Fatalf("stdout missing Fatal Error:\n%s", report)
	}
	if n := strings.Count(report, "Fatal Error:"); n != 1 {
		t.Fatalf("Fatal Error printed %d times, want 1:\n%s", n, report)
	}
	wantTag := "boomPip#" + strconv.Itoa(boomID)
	if !strings.Contains(report, "["+wantTag+"]") {
		t.Fatalf("stdout missing task tag [%s]:\n%s", wantTag, report)
	}
	deadline := time.Now().Add(2 * time.Second)
	for root.descendantCount.Load() != 0 {
		if time.Now().After(deadline) {
			t.Fatalf("descendantCount = %d after fault fanout", root.descendantCount.Load())
		}
		time.Sleep(time.Millisecond)
	}
}
