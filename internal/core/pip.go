package core

import (
	"fmt"
	"strconv"
)

// Pip is a registered Final type and the Go handle for one child Task in a parent’s
// children map. Phase 2 registers almost no members — only Pip.newPipParams^.
type Pip struct {
	id   int
	task *Task
}

func (p *Pip) FinalType() *Type { return pipType }

func (p *Pip) ID() int {
	if p == nil {
		panic("Pip.ID: nil Pip")
	}
	return p.id
}

func (p *Pip) Task() *Task {
	if p == nil {
		return nil
	}
	return p.task
}

// treeRoot walks to the Session/root Task (parent == nil).
func (t *Task) treeRoot() *Task {
	if t == nil {
		return nil
	}
	for t.parent != nil {
		t = t.parent
	}
	return t
}

// incDescendants walks parent → root and adds 1 to each ancestor’s descendantCount.
func (t *Task) incDescendants() {
	for a := t; a != nil; a = a.parent {
		a.descendantCount.Add(1)
	}
}

// decDescendants walks parent → root and subtracts 1 from each ancestor’s descendantCount.
func (t *Task) decDescendants() {
	for a := t; a != nil; a = a.parent {
		a.descendantCount.Add(-1)
	}
}

// notePipRunFinished marks Run exit and reaps this pip when it has no live descendants.
func (t *Task) notePipRunFinished() {
	if t == nil {
		return
	}
	t.childrenMu.Lock()
	t.runFinished = true
	t.childrenMu.Unlock()
	t.tryReap()
}

// tryReap removes this Task from its parent map and decrements ancestor counts when
// Run has returned and descendantCount == 0; then cascades to the parent.
func (t *Task) tryReap() {
	if t == nil || t.parent == nil {
		return
	}
	t.childrenMu.Lock()
	finished := t.runFinished
	desc := t.descendantCount.Load()
	id := t.pipId
	parent := t.parent
	t.childrenMu.Unlock()
	if !finished || desc > 0 {
		return
	}
	parent.removePip(id)
	parent.decDescendants()
	parent.tryReap()
}

// escalateFaultUp stores the first fault at the tree root, fanouts Break down the tree,
// Breaks the root, then writes the fault report when the culprit is a child pip.
// Runs on the faulting Task’s goroutine (not the root) so a pip report can print even when
// the mothership is blocked (e.g. Pipe.Pull). Culprit: sessionFaultPipId (0 = root/self)
// and fault.TaskName (qualified pip name when spawned). First fault wins.
//
// Root/self faults (t.parent == nil) do not print here — the outer REPL / Session caller
// owns reporting via doneChan (avoids double Fatal Error on Go-hosted interpret).
// Child-pip reports use WriteRunFaultReport → culprit RP stdout.
func (t *Task) escalateFaultUp(fault *RunFault, rp *RunProvider) {
	if t == nil || fault == nil {
		return
	}
	root := t.treeRoot()
	if root == nil {
		return
	}
	culpritID := t.pipId
	root.sessionFaultOnce.Do(func() {
		root.sessionFault = fault
		root.sessionFaultPipId = culpritID
		root.propagateFaultDown()
		root.Break() // after children: top-level Task aborts too
		if t.parent != nil {
			WriteRunFaultReport(rp, fault)
		}
	})
}

// propagateFaultDown Breaks each direct child Task and recurses.
func (t *Task) propagateFaultDown() {
	if t == nil {
		return
	}
	t.childrenMu.Lock()
	kids := make([]*Pip, 0, len(t.children))
	for _, p := range t.children {
		kids = append(kids, p)
	}
	t.childrenMu.Unlock()
	for _, p := range kids {
		if p == nil || p.task == nil {
			continue
		}
		p.task.Break()
		p.task.propagateFaultDown()
	}
}

// CreateSubTask allocates a child Task sharing parent’s DC and registers a Pip handle.
func CreateSubTask(parent *Task) (*Task, *Pip) {
	if parent == nil {
		panic("CreateSubTask: nil parent")
	}
	if parent.dc == nil {
		panic("CreateSubTask: parent has nil dc")
	}
	child := &Task{
		name:          fmt.Sprintf("%s/pip", parent.name),
		ip:            EncodeVirt(ModuleIDThis, 0),
		compileModID:  parent.compileModID,
		dc:            parent.dc,
		compilerState: NewCompilerState(parent.dc),
		parent:        parent,
		restrict:      parent.restrict, // spawn-time snapshot; not updated if root Send changes SR
		// systemMode stays false: pips run user XT, they do not compile core.
		session:            parent.session,
		currentSearchOrder: parent.currentSearchOrder,
	}
	parent.childrenMu.Lock()
	defer parent.childrenMu.Unlock()
	parent.nextPipID++
	id := parent.nextPipID
	child.pipId = id
	if parent.children == nil {
		parent.children = make(map[int]*Pip)
	}
	if parent.dc.lockedFn == nil {
		root := parent.treeRoot()
		parent.dc.lockedFn = func() bool {
			return root.descendantCount.Load() > 0
		}
	}
	pip := &Pip{id: id, task: child}
	parent.children[id] = pip
	parent.incDescendants()
	return child, pip
}

// wordFromXT resolves a VirtXT to *Word (same demux rules as OpExec).
func wordFromXT(t *Task, xt XT) *Word {
	if xt == 0 {
		panic(fmt.Errorf("wordFromXT: VirtXT was 0"))
	}
	moduleId, offset := DecodeVirt(VirtualAddr(xt))
	var dc *DictionaryContext
	var idx int
	switch moduleId {
	case ModuleIDThis:
		if t != nil && t.dc != nil {
			dc = t.dc
		} else {
			dc = globalDictionaryContext
		}
		idx = offset
	default:
		dc = ModuleTable.DCForModule(t, moduleId)
		idx = offset - int(moduleId)
	}
	if dc == nil || dc.dict == nil {
		panic(fmt.Errorf("wordFromXT: nil DC for xt %v", xt))
	}
	w := dc.dict.GetByOrdinal(idx)
	if w == nil {
		panic(fmt.Errorf("wordFromXT: invalid xt %v", xt))
	}
	return w
}

// pipNameBase returns PipParams ._name when present and non-empty; otherwise the spawn word name.
func pipNameBase(pp *AtomicDynamicObject, w *Word) string {
	if pp != nil && bool(pp.HasField("_name")) {
		if s, ok := pp.Get("_name").(String); ok && len(s) > 0 {
			return string(s)
		}
	}
	if w == nil {
		return "pip"
	}
	return w.name
}

// spawnPip starts the word named by xt on a new child Task.
// PipParams bag pp must provide ._stdout and ._stderr (TextWrite). Other fields are caller payload.
// Optional ._name overrides the diagnostic base (else spawn word name); Task.name is base#pipId.
// Returns the Pip handle. When the child Run exits, try reap (keep map entry while descendants live).
func spawnPip(parent *Task, xt XT, pp *AtomicDynamicObject) *Pip {
	if parent == nil {
		panic("spawnPip: nil parent")
	}
	if pp == nil {
		panic("spawnPip: nil PipParams AtomicDynamicObject")
	}
	stdoutVal := pp.Get("_stdout")
	stdout, ok := stdoutVal.(TextWrite)
	if !ok || stdoutVal == nil {
		panic(fmt.Errorf("spawnPip: PipParams._stdout must be TextWrite, got %T", stdoutVal))
	}
	stderrVal := pp.Get("_stderr")
	stderr, ok := stderrVal.(TextWrite)
	if !ok || stderrVal == nil {
		panic(fmt.Errorf("spawnPip: PipParams._stderr must be TextWrite, got %T", stderrVal))
	}
	w := wordFromXT(parent, xt)

	child, pip := CreateSubTask(parent)
	child.name = pipNameBase(pp, w) + "#" + strconv.Itoa(pip.ID())
	childRP := NewRunProvider()
	childRP.Stdout = stdout
	childRP.Stderr = stderr
	childRP.DataStack.Push(pp)

	go func() {
		_ = runExecWord(child, childRP, w)
		child.notePipRunFinished()
	}()
	return pip
}

func registerPipMembers(t *Type) {
	// ( -- pp:AtomicDynamicObject ) with ._stdout / ._stderr = current RP writers; caller adds payload fields.
	t.registerNativeClassMethod("newPipParams", func(f *RunFrame, _ *TypeObjPair) {
		if f == nil || f.RP == nil {
			panic("Pip.newPipParams: nil RunFrame/RP")
		}
		pp := NewAtomicDynamicObject()
		pp.Set("_stdout", f.RP.Stdout)
		pp.Set("_stderr", f.RP.Stderr)
		f.RP.DataStack.Push(pp)
	})
}

func CreateRegisterPipType() {
	t := &Type{
		name:            "Pip",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Pip]("Pip"),
		restrict:        SR_Threads,
	}
	registerPipMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	pipType = t
}

var pipType *Type

func init() {
	CreateRegisterPipType()
}

var _ Typed = (*Pip)(nil)
