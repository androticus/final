package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// chdirRepoRootForTests sets cwd to the module root so bootstrap manifest paths
// (internal/final/...) resolve when go test runs from internal/core.
func chdirRepoRootForTests() {
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("chdirRepoRootForTests: %w", err))
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if err := os.Chdir(dir); err != nil {
				panic(fmt.Errorf("chdirRepoRootForTests: %w", err))
			}
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("chdirRepoRootForTests: go.mod not found")
		}
		dir = parent
	}
}

// InitTestRuntime confirms the global execution environment is ready for axiom tests.
// TestMain calls ResolveBuiltInTypeRegistrations, InitGlobalExecutionEnvironment, then this.
func InitTestRuntime() {
	if globalDictionaryContext == nil {
		panic(fmt.Errorf("InitTestRuntime: global execution environment not initialized"))
	}
}

// ResetTestRuntime restores global Dict/CT/DT to the post-init snapshot
// (axioms, bootstrap shell, boot CT, core0/core1/core2 compile, core3 if enabled).
func ResetTestRuntime() {
	ResetGlobalExecutionEnvironmentForTest()
}

// NewTestTask returns a fresh Task with zero defaults for one test case.
func NewTestTask() *Task {
	t, _ := NewTask(false)
	return t
}

// NewTestRunProvider returns fresh execution stacks and null stdio for one test Run.
func NewTestRunProvider() *RunProvider {
	return NewRunProvider()
}

// pushTypedOnContextStack places a Typed receiver on the context stack (CS aliases DS in axiom tests).
func pushTypedOnContextStack(rp *RunProvider, obj Typed) {
	rp.DataStack.Push(obj)
}

// memberRunFrame returns a RunFrame and ths pair for native member handler tests.
func memberRunFrame(rp *RunProvider, obj any) (*RunFrame, *TypeObjPair) {
	ths := rp.Resolve(obj)
	return &RunFrame{RP: rp}, ths
}

func appendOpsWithCSArming(ct *VirtCTResolver, sel ContextStackIndex, ops ...OpCode) IP {
	var ip IP
	started := false
	appendCell := func(val int) {
		if !started {
			ip = ct.Append(ModuleIDThis, val)
			started = true
		} else {
			ct.Append(ModuleIDThis, val)
		}
	}
	for _, op := range ops {
		if axiomWordForOp(op).RequiresStackContext() {
			appendCell(int(OpDoCSSelect))
			appendCell(int(sel))
		}
		appendCell(int(op))
	}
	return ip
}

// RunOpsArmedCS runs ops, emitting OpDoCSSelect · sel before each cs-using opcode.
func RunOpsArmedCS(t *Task, rp *RunProvider, sel ContextStackIndex, ops ...OpCode) {
	if rp == nil {
		panic(fmt.Errorf("RunOpsArmedCS: nil RunProvider"))
	}
	if len(ops) == 0 {
		panic(fmt.Errorf("RunOpsArmedCS: no opcodes"))
	}
	ct := &VirtCTResolver{}
	ip := appendOpsWithCSArming(ct, sel, ops...)
	t.ip = ip
	Run(t, rp)
}

// RunOpsWithSubCodeArmedCS runs OpDoSubCode · subCode then ops with per-op CS arming.
func RunOpsWithSubCodeArmedCS(t *Task, rp *RunProvider, sel ContextStackIndex, subCode int, ops ...OpCode) {
	if rp == nil {
		panic(fmt.Errorf("RunOpsWithSubCodeArmedCS: nil RunProvider"))
	}
	if len(ops) == 0 {
		panic(fmt.Errorf("RunOpsWithSubCodeArmedCS: no opcodes"))
	}
	ct := &VirtCTResolver{}
	ip := ct.Append(ModuleIDThis, int(OpDoSubCode))
	ct.Append(ModuleIDThis, subCode)
	appendOpsWithCSArming(ct, sel, ops...)
	t.ip = ip
	Run(t, rp)
}

// RunOps appends opcodes to global CT via VirtCTResolver, sets task ip to the program
// start, and runs synchronously. Typical axiom test:
// RunOps(t, rp, opUnderTest, OpStop).
func RunOps(t *Task, rp *RunProvider, ops ...OpCode) {
	if rp == nil {
		panic(fmt.Errorf("RunOps: nil RunProvider"))
	}
	if len(ops) == 0 {
		panic(fmt.Errorf("RunOps: no opcodes"))
	}
	ct := &VirtCTResolver{}
	ip := appendOpsWithCSArming(ct, csDataStack, ops...)
	t.ip = ip
	Run(t, rp)
}

// RunOpsWithSubCode prepends OpDoSubCode · subCode before ops (invoke workers use runSubCode).
func RunOpsWithSubCode(t *Task, rp *RunProvider, subCode int, ops ...OpCode) {
	if rp == nil {
		panic(fmt.Errorf("RunOpsWithSubCode: nil RunProvider"))
	}
	if len(ops) == 0 {
		panic(fmt.Errorf("RunOpsWithSubCode: no opcodes"))
	}
	ct := &VirtCTResolver{}
	ip := ct.Append(ModuleIDThis, int(OpDoSubCode))
	ct.Append(ModuleIDThis, subCode)
	appendOpsWithCSArming(ct, csDataStack, ops...)
	t.ip = ip
	Run(t, rp)
}
