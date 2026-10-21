package core

import "fmt"

const (
	subwordOpenMsg  = `":s requires an open colon definition"`
	subwordCloseMsg = `";s requires a preceding :s"`
	subwordLateMsg  = `":s must appear before the main body of the word"`
	subwordNestMsg  = `"subword definitions cannot be nested"`
)

func (w *Word) clearSubwordCompileState() {
	w.subwords = nil
	w.subSkipMark = -1
	w.openSubStart = -1
}

// findSubword returns the index of name in w.subwords, or -1.
func (w *Word) findSubword(name string) int {
	for i := range w.subwords {
		if w.subwords[i].name == name {
			return i
		}
	}
	return -1
}

// maybeResolveSubwordSkip patches the prologue skip branch when the first main-body
// CT cell is about to be emitted (openSubStart < 0 and skip still unresolved).
func (cs *CompilerState) maybeResolveSubwordSkip() {
	w := cs.DC.dict.PeekBuilding()
	if w == nil || w.subSkipMark < 0 || w.openSubStart >= 0 {
		return
	}
	here := cs.ct().Len()
	j := here - (w.subSkipMark + 1)
	cs.ct().Write(w.subSkipMark, j)
	w.subSkipMark = -1
}

// tryCompileSubwordRef emits brsub for a closed subword name.
// Returns handled=true when the name is in the building word's subword list.
func tryCompileSubwordRef(cs *CompilerState, name string) (handled bool, err error) {
	w := cs.DC.dict.PeekBuilding()
	if w == nil {
		return false, nil
	}
	idx := w.findSubword(name)
	if idx < 0 {
		return false, nil
	}
	sw := &w.subwords[idx]
	if sw.entry < 0 {
		return true, fmt.Errorf("subword %q is not yet closed", name)
	}
	bootstrapEmitBrSub(cs, sw.entry)
	return true, nil
}

// bootstrapEmitBrSub emits OpBrSub + relative offset to entry (same shape as OpBranch).
func bootstrapEmitBrSub(cs *CompilerState, entry IP) {
	bootstrapEmitOpcode(cs, OpBrSub)
	jSlot := decodeCTOffset(cs.emitCT(0))
	j := int(entry) - (jSlot + 1)
	cs.ct().Write(jSlot, j)
}

// bootstrapSubwordStart handles compile-time `:s` — reads name, optional skip branch, opens BlockKindSubword.
func bootstrapSubwordStart(p *Parser, t *Task, rp *RunProvider) error {
	cs := t.compilerForTarget()
	if !GetCompileState(cs) {
		return fmt.Errorf("%s", subwordOpenMsg)
	}
	w := cs.DC.dict.PeekBuilding()
	if w == nil {
		return fmt.Errorf("%s", subwordOpenMsg)
	}
	if w.openSubStart >= 0 {
		return fmt.Errorf("%s", subwordNestMsg)
	}
	ctLen := cs.ct().Len()
	if ctLen != int(w.colonCode) && w.subSkipMark < 0 {
		return fmt.Errorf("%s", subwordLateMsg)
	}

	val, typ, err := p.Next()
	if err != nil {
		return err
	}
	if !IsPlainIdentToken(typ) {
		return fmt.Errorf(":s expected subword name, got %v", typ)
	}
	name, ok := val.(String)
	if !ok {
		return fmt.Errorf(":s expected String name, got %T", val)
	}
	nameStr := string(name)
	if nameStr == w.name {
		return fmt.Errorf("subword name %q collides with the word being defined", nameStr)
	}
	if w.findSubword(nameStr) >= 0 {
		return fmt.Errorf("duplicate subword name %q", nameStr)
	}

	if ctLen == int(w.colonCode) {
		// Prologue skip so nest/recurse still enter at colonCode.
		bootstrapEmitOpcode(cs, OpBranch)
		w.subSkipMark = decodeCTOffset(cs.emitCT(0))
	}

	w.subwords = append(w.subwords, Subword{name: nameStr, entry: -1})
	w.openSubStart = cs.ct().Len()
	rp.TempStack.Push(NewBlockToken(BlockKindSubword))
	return nil
}

// bootstrapSubwordEnd handles compile-time `;s` — emits retsub, closes the open subword.
func bootstrapSubwordEnd(t *Task, rp *RunProvider) error {
	cs := t.compilerForTarget()
	if !GetCompileState(cs) {
		return fmt.Errorf("%s", subwordCloseMsg)
	}
	w := cs.DC.dict.PeekBuilding()
	if w == nil || w.openSubStart < 0 || len(w.subwords) == 0 {
		return fmt.Errorf("%s", subwordCloseMsg)
	}
	ts := rp.TempStack
	if ts.Len() == 0 {
		return fmt.Errorf("%s", subwordCloseMsg)
	}
	block, ok := ts.Peek().(*BlockToken)
	if !ok || block.Kind() != BlockKindSubword {
		return fmt.Errorf("%s", subwordCloseMsg)
	}

	sw := &w.subwords[len(w.subwords)-1]
	if sw.entry >= 0 {
		return fmt.Errorf("%s", subwordCloseMsg)
	}
	// Emit retsub while openSubStart is still set so maybeResolveSubwordSkip
	// does not treat retsub as the first main-body cell.
	bootstrapEmitOpcode(cs, OpRetSub)
	sw.entry = IP(w.openSubStart)
	w.openSubStart = -1
	ts.Pop()
	return nil
}
