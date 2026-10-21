package core

import (
	"bufio"
	"fmt"
	"math"
	"strings"
	"sync"
	"testing"
)

func TestFormatErrorCell_numericTypes(t *testing.T) {
	cases := []struct {
		v    any
		want string
	}{
		{Int(-3), "-3"},
		{Uint(42), "42u"},
		{Float(1), "1.0"},
		{Float(math.Copysign(0, -1)), "-0.0"},
		{Float(1.5), "1.5"},
		{Float(1e20), "1e+20"},
		{Float(math.Inf(1)), "+Inf"},
		{Float(math.NaN()), "NaN"},
	}
	for _, tc := range cases {
		if got := FormatErrorCell(tc.v); got != tc.want {
			t.Fatalf("FormatErrorCell(%#v) = %q, want %q", tc.v, got, tc.want)
		}
	}
}

func TestFormatErrorCell_elidesLongString(t *testing.T) {
	long := String(strings.Repeat("x", 80))
	got := FormatErrorCell(long)
	if !strings.Contains(got, "…") {
		t.Fatalf("got %q, want elision", got)
	}
}

func TestFormatErrorCell_unknownType(t *testing.T) {
	got := FormatErrorCell(struct{ x int }{1})
	if got != "<struct { x int }>" {
		t.Fatalf("got %q", got)
	}
}

func TestElideStackLines(t *testing.T) {
	lines := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	got := elideStackLines(lines, faultStackReportHead, faultStackReportTail)
	if len(got) != faultStackReportHead+1+faultStackReportTail || got[faultStackReportHead] != "..." {
		t.Fatalf("got %v", got)
	}
}

func TestFaultReportTemplate_emptyDataAndTempStack(t *testing.T) {
	sb := &StringBuilder{}
	rp := NewTestRunProvider()
	rp.Stdout = sb
	WriteRunFaultReport(rp, &RunFault{
		Message:     "test",
		ReturnStack: []string{"*: at nop"},
	})
	out := sb.ToString()
	if !strings.Contains(out, "Data Stack (tos first):\n  (empty)\n") {
		t.Fatalf("want empty data stack marker:\n%s", out)
	}
	if !strings.Contains(out, "Temp Stack (tos first):\n  (empty)\n") {
		t.Fatalf("want empty temp stack marker:\n%s", out)
	}
	if !strings.Contains(out, "X Stack (tos first):\n  (empty)\n") {
		t.Fatalf("want empty x stack marker:\n%s", out)
	}
	if !strings.Contains(out, "Y Stack (tos first):\n  (empty)\n") {
		t.Fatalf("want empty y stack marker:\n%s", out)
	}
}

func TestFaultReportTemplate_indentsEntries(t *testing.T) {
	sb := &StringBuilder{}
	rp := NewTestRunProvider()
	rp.Stdout = sb
	WriteRunFaultReport(rp, &RunFault{
		Message:     "test",
		ReturnStack: []string{"0: a"},
		DataStack:   []string{"0: 1"},
		TempStack:   []string{"0: t"},
	})
	out := sb.ToString()
	if !strings.Contains(out, "Return Stack (tos first, * = fault site):") || !strings.Contains(out, "Data Stack (tos first)") {
		t.Fatalf("want tos-first headings:\n%s", out)
	}
	if !strings.Contains(out, "  0: a\n") || !strings.Contains(out, "  0: 1\n") || !strings.Contains(out, "  0: t\n") {
		t.Fatalf("want two-space indent and index per entry:\n%s", out)
	}
}

func TestResolveWordLabel_replLoopNotWarmBoot(t *testing.T) {
	ResetTestRuntime()
	// OpExec in ColdBoot repl body is around CT 12; return ip after Nest often lands nearby.
	off := 12
	got := resolveWordLabelAtCT(globalDictionaryContext, off, OpINVALID)
	if got == bootWordWarmBoot {
		t.Fatalf("ct:%d resolved to %q, want repl driver not boot word entry", off, got)
	}
	if got != "repl" {
		t.Fatalf("ct:%d = %q, want repl", off, got)
	}
}

func TestResolveWordLabel_colonBody(t *testing.T) {
	ResetTestRuntime()
	tk, rp := NewTestTask(), NewTestRunProvider()
	cs := tk.compilerForTarget()
	bootstrapCreate(cs, "div0")
	bootstrapEmitLiteral(cs, Int(1))
	bootstrapEmitLiteral(cs, Int(0))
	bootstrapEmitOpcode(cs, OpDiv)
	bootstrapFinalizeColonWord(cs)
	w := dictGet(t, "div0")
	if w == nil {
		t.Fatal("div0 not in dict")
	}
	off := int(w.colonCode) + 2 // past two literals toward OpDiv
	got := resolveWordLabelAtCT(globalDictionaryContext, off, OpDiv)
	if got != "div0" {
		t.Fatalf("got %q, want div0", got)
	}
	_ = rp
}

func dictGet(t *testing.T, name string) *Word {
	t.Helper()
	w := globalDictionaryContext.dict.GetByName(name)
	return w
}

func TestBuildFaultStackSnapshot_indexZeroIsBottom(t *testing.T) {
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(10))
	rp.DataStack.Push(Int(20))
	lines := buildFaultStackSnapshot(rp.DataStack)
	if len(lines) != 2 || lines[0] != "1: 20" || lines[1] != "0: 10" {
		t.Fatalf("got %v, want TOS-first", lines)
	}
}

func TestCaptureRunFault_divideByZero(t *testing.T) {
	ResetTestRuntime()
	task, rp := NewTestTask(), NewTestRunProvider()
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(Int(0))
	f := captureRunFault(task, rp, fmt.Errorf("divide by zero"), OpDiv, 0)
	if f.Message != "divide by zero" {
		t.Fatalf("message = %q", f.Message)
	}
	if len(f.DataStack) != 2 || f.DataStack[0] != "1: 0" || f.DataStack[1] != "0: 1" {
		t.Fatalf("data stack = %v, want TOS-first 1:0 then 0:1", f.DataStack)
	}
	if len(f.ReturnStack) < 1 || !strings.HasPrefix(f.ReturnStack[0], "*:") {
		t.Fatalf("return stack = %v, want *: fault site first", f.ReturnStack)
	}
}

func TestWriteRunFaultReportTemplate(t *testing.T) {
	rp := NewTestRunProvider()
	f := &RunFault{
		Message:     "divide by zero",
		ReturnStack: []string{"at / (ct:0)"},
		DataStack:   []string{"0", "1"},
		TempStack:   []string{"List(len=2)"},
	}
	WriteRunFaultReport(rp, f)
}

func TestCaptureRunFault_includesTempStack(t *testing.T) {
	ResetTestRuntime()
	task, rp := NewTestTask(), NewTestRunProvider()
	rp.TempStack.Push(Int(42))
	f := captureRunFault(task, rp, "test", OpNop, 0)
	if len(f.TempStack) != 1 || f.TempStack[0] != "0: 42" {
		t.Fatalf("TempStack = %v", f.TempStack)
	}
}

func TestBootstrapREPLSurvivesPanic(t *testing.T) {
	ResetTestRuntime()
	tk, _ := NewTask(false)
	rp := NewTestRunProvider()
	sb := &StringBuilder{}
	rp.Stdout = sb
	rp.Stdin = &TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("1 0 /\n2 3 +\n")),
		label: "test",
	}}
	if err := BootstrapREPL(tk, rp); err != nil {
		t.Fatal(err)
	}
	report := sb.ToString()
	if !strings.Contains(report, "Fatal Error:") {
		t.Fatalf("stdout = %q", report)
	}
	if !strings.Contains(report, "divide by zero") {
		t.Fatalf("stdout = %q", report)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 5 {
		t.Fatalf("got %v, want 5", got)
	}
}

// TestConcurrentRunFaultSites proves each Task's doneChan fault reports its own LastOp
// (no package-global fault site).
func TestConcurrentRunFaultSites(t *testing.T) {
	ResetTestRuntime()
	// Build CT programs serially (global CT is not concurrent-safe to Append).
	ctDiv := &VirtCTResolver{}
	ipDiv := appendOpsWithCSArming(ctDiv, csDataStack, OpDiv, OpStop)
	ctDup := &VirtCTResolver{}
	ipDup := appendOpsWithCSArming(ctDup, csDataStack, OpDup, OpStop)

	const n = 32
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			task := NewTestTask()
			rp := NewTestRunProvider()
			wantOp := OpDiv
			task.ip = ipDiv
			if i%2 == 1 {
				wantOp = OpDup
				task.ip = ipDup
			} else {
				rp.DataStack.Push(Int(1))
				rp.DataStack.Push(Int(0))
			}
			done := make(chan error, 1)
			task.doneChan = done
			Run(task, rp)
			err := <-done
			if err == nil {
				errs[i] = fmt.Errorf("goroutine %d: expected fault", i)
				return
			}
			fault, ok := err.(*RunFault)
			if !ok {
				errs[i] = fmt.Errorf("goroutine %d: got %T %v", i, err, err)
				return
			}
			if fault.LastOp != wantOp {
				errs[i] = fmt.Errorf("goroutine %d: LastOp=%v want %v", i, fault.LastOp, wantOp)
				return
			}
			if rp.LastOp != wantOp {
				errs[i] = fmt.Errorf("goroutine %d: rp.LastOp=%v want %v", i, rp.LastOp, wantOp)
			}
		}(i)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestWarmBootResetsStacks(t *testing.T) {
	ResetTestRuntime()
	task, rp := NewTestTask(), NewTestRunProvider()
	rp.DataStack.Push(Int(99))
	rp.ReturnStack.PushNew(int(ReservedCTBootStop))
	rp.TempStack.Push(Int(1))
	rp.XStack.Push(Int(2))
	rp.YStack.Push(Int(3))
	rp.ZStack.Push(Int(4))
	rp.OStack.Push(Int(5))
	warmBootReset(task, rp)
	if rp.DataStack.Len() != 0 || rp.ReturnStack.Len() != 0 || rp.TempStack.Len() != 0 {
		t.Fatal("DS/RS/TS should be empty after warmBootReset")
	}
	if rp.XStack.Len() != 0 || rp.YStack.Len() != 0 || rp.ZStack.Len() != 0 || rp.OStack.Len() != 0 {
		t.Fatal("XS/YS/ZS/OS should be empty after warmBootReset")
	}
}

func TestEventLoopRunIsFinalMethod(t *testing.T) {
	ResetTestRuntime()
	mm := GetEventLoop().FinalType().GetMethodMetadata("run")
	if mm == nil || !mm.isFinalMethod || mm.finalColonCode < 0 {
		t.Fatal("EventLoop#run should be a torn-off Final method with colonCode")
	}
}

// Root faults must not print in escalateFaultUp — runInterpretLoop prints once via doneChan.
func TestRootFaultEscalateDoesNotPrintReport(t *testing.T) {
	ResetTestRuntime()
	task := CreateLocalTask()
	rp := NewTestRunProvider()
	outSB := NewStringBuilder()
	rp.Stdout = outSB
	src := `: boom "N must be >= 2" error ;`
	tr, err := NewTextReaderFromString(src)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	rp.WithTaskFindScope(task)
	if err := interpretSession(task, rp); err != nil {
		t.Fatalf("compile: %v", err)
	}
	w := task.dc.dict.GetByName("boom")
	if w == nil {
		t.Fatal("missing boom")
	}

	runErr := runExecWord(task, rp, w)
	if _, ok := runErr.(*RunFault); !ok {
		t.Fatalf("runExecWord err = %v, want *RunFault", runErr)
	}
	if got := outSB.ToString(); got != "" {
		t.Fatalf("escalate must not print root fault; stdout=%q", got)
	}

	// Mimic runInterpretLoop: one report from the outer caller.
	WriteRunFaultReport(rp, runErr.(*RunFault))
	if n := strings.Count(outSB.ToString(), "Fatal Error:"); n != 1 {
		t.Fatalf("after outer report: Fatal Error count = %d, want 1:\n%s", n, outSB.ToString())
	}
}

func TestFaultLabel_localOOBNotCoreWord(t *testing.T) {
	ResetTestRuntime()
	task := CreateLocalTask()
	got := formatReturnStackEntry(task.dc, 3906, OpINVALID, "at")
	if strings.Contains(got, "jsonParseValue") || strings.Contains(got, "logTrace") || strings.Contains(got, "list-os") {
		t.Fatalf("OOB local offset labeled as core word: %q", got)
	}
	if !strings.Contains(got, "ct:3906") || !strings.Contains(got, "len:") || !strings.Contains(got, "local") {
		t.Fatalf("got %q, want local/ct/len OOB form", got)
	}
	if strings.Contains(got, "local:") || strings.Contains(got, "module:") {
		t.Fatalf("got %q, want short local without id suffix or module: prefix", got)
	}
}

func TestFaultLabel_localColonWordUsesLocalDC(t *testing.T) {
	ResetTestRuntime()
	task := CreateLocalTask()
	rp := NewTestRunProvider()
	src := `: localProbe 1 ;`
	tr, err := NewTextReaderFromString(src)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	rp.WithTaskFindScope(task)
	if err := interpretSession(task, rp); err != nil {
		t.Fatalf("compile: %v", err)
	}
	w := task.dc.dict.GetByName("localProbe")
	if w == nil {
		t.Fatal("missing localProbe")
	}
	got := resolveWordLabelAtCT(task.dc, int(w.colonCode), OpINVALID)
	if got != "localProbe" {
		t.Fatalf("got %q, want localProbe", got)
	}
	line := formatReturnStackEntry(task.dc, IP(w.colonCode), OpINVALID, "at")
	if !strings.Contains(line, "(local ct:") {
		t.Fatalf("want (local ct:…) in non-core line: %q", line)
	}
	if strings.Contains(line, "local:") || strings.Contains(line, "module:") {
		t.Fatalf("want no local:<id> or module: prefix in line: %q", line)
	}
}

func TestFaultLabel_localOffsetNotRepl(t *testing.T) {
	ResetTestRuntime()
	task := CreateLocalTask()
	for i := 0; i < 20; i++ {
		task.dc.ct.Append(int(OpNop))
	}
	got := resolveWordLabelAtCT(task.dc, 12, OpINVALID)
	if got == "repl" {
		t.Fatalf("local ct:12 labeled repl")
	}
}
