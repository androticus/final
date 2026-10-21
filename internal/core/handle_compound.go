package core

import "strings"

// emitCompoundPathCSSandwichCompile wraps compile emission for d:/t:/x:/y: prefixed simple ops only.
// Arming is emitted by bootstrapEmitWordReference(cssel); this wrapper restores d: after the op
// by emitting OpDoCSSelect (not by Find/exec of System word cs,).
func emitCompoundPathCSSandwichCompile(t *Task, cp *compoundPathValue, fn func() error) error {
	armed := cp != nil && cp.stackSelector >= 0
	err := fn()
	if armed {
		cs := t.compilerForTarget()
		if GetCompileState(cs) {
			bootstrapEmitDoCSSelectPrefix(cs, csDataStack)
		}
	}
	return err
}

// getStack maps arming index to a physical stack on RunProvider (same indices as OpDoCSSelect).
func getStack(rp *RunProvider, sel ContextStackIndex) *Stack[any] {
	switch sel {
	case csTStack:
		return rp.TempStack
	case csXStack:
		return rp.XStack
	case csYStack:
		return rp.YStack
	case csZStack:
		return rp.ZStack
	case csOStack:
		return rp.OStack
	default:
		return rp.DataStack
	}
}

// segmentToStackSelector maps a path segment "d:" / "t:" / "x:" / "y:" to ContextStackIndex.
func segmentToStackSelector(seg string) (ContextStackIndex, bool) {
	switch seg {
	case "d:":
		return csDataStack, true
	case "t:":
		return csTStack, true
	case "x:":
		return csXStack, true
	case "y:":
		return csYStack, true
	case "z:":
		return csZStack, true
	case "o:":
		return csOStack, true
	default:
		return csDataStack, false
	}
}

// stackSelectorToSegment is the inverse of segmentToStackSelector for compile emit.
func stackSelectorToSegment(sel ContextStackIndex) string {
	switch sel {
	case csTStack:
		return "t:"
	case csXStack:
		return "x:"
	case csYStack:
		return "y:"
	case csZStack:
		return "z:"
	case csOStack:
		return "o:"
	default:
		return "d:"
	}
}

// resolveCompoundWalkSelector peels optional first segment; returns (sel, restSegs).
func resolveCompoundWalkSelector(segs []string) (ContextStackIndex, []string) {
	if len(segs) > 0 {
		if sel, ok := segmentToStackSelector(segs[0]); ok {
			return sel, segs[1:]
		}
	}
	return csDataStack, segs
}

// compoundFrozenPathLiteral embeds stack prefix as first path segment for OpHandleCompound.
func compoundFrozenPathLiteral(cp *compoundPathValue) String {
	segs := compoundPathSegments(cp.path)
	if cp.stackSelector >= 0 {
		segs = append([]string{stackSelectorToSegment(cp.stackSelector)}, segs...)
	}
	return String(strings.Join(segs, " "))
}

// compoundWalker is in-hand compound path state; sel names the receiver stack for the whole walk.
// localsRoot is set when the path begins with segment "l" (ColonFrameMark locals DO, not a stack).
// plsRoot is set when the path begins with segment "pls" (Task Pip Local Storage DO, not a stack).
// Ident-rooted paths (sess.config…) materialize the ident into focus and elide it; they may
// terminal-store without t:/x:/… (same exemption class as l./pls.).
// f is the run-loop RunFrame (singleton in run.go); never allocate a new RunFrame here.
type compoundWalker struct {
	f          *RunFrame
	sel        ContextStackIndex
	localsRoot bool
	plsRoot    bool
}

func (w *compoundWalker) rootStack() *Stack[any] {
	return getStack(w.f.RP, w.sel)
}

// handleCompound demuxes a frozen compound path on DS ( path:s -- … ).
func handleCompound(f *RunFrame) {
	if f == nil || f.RP == nil || f.Task == nil {
		return
	}
	ds := f.RP.DataStack
	if ds.Len() < 1 {
		return
	}
	path, ok := ds.Peek().(String)
	if !ok || path == "" {
		return
	}
	ds.Pop()
	sel, segs := resolveCompoundWalkSelector(strings.Split(string(path), " "))
	w := &compoundWalker{f: f, sel: sel}
	if err := walkCompoundPath(w, segs); err != nil {
		panic(err)
	}
}

func compoundSegmentIs(want string, segs []string, i int) bool {
	return i >= 0 && i < len(segs) && segs[i] == want
}

func compoundSegmentIdent(segs []string, i int) (ident string, consumed int) {
	if i < 0 || i >= len(segs) {
		return "", 0
	}
	return segs[i], 1
}
