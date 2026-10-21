package core

import "fmt"

type compoundTerminalKind int

const (
	compoundTerminalNone compoundTerminalKind = iota
	compoundTerminalValue
	compoundTerminalPropertyRef
	compoundTerminalMethodRef
	compoundTerminalMethodDone
	compoundTerminalStoreDone
)

// compoundDotFollow describes the segment immediately after .<name> in a path.
type compoundDotFollow int

const (
	compoundDotFollowTerminal  compoundDotFollow = iota // end of path
	compoundDotFollowContinue                           // more path segments (e.g. .a.b)
	compoundDotFollowRefSuffix                          // .name~
	compoundDotFollowAtBang                             // .name@ / .name! — explicit @/! requires PropertyRef
	compoundDotFollowCaret                              // .name^ — MethodRef then invoke on ^
)

func compoundPathHasMemberAccess(segs []string) bool {
	if len(segs) > 0 && segs[0] == segAt {
		return true
	}
	for i := range segs {
		if compoundSegmentIs(segDot, segs, i) {
			return true
		}
		if compoundSegmentIsCollectionAffix(segs[i]) {
			return true
		}
	}
	return false
}

func compoundSegmentIsCollectionAffix(seg string) bool {
	switch seg {
	case segAtIndex, segBangIndex, segAtBrace, segBangBrace, segQueryBrace, segAtMulti, segBangMulti:
		return true
	default:
		return false
	}
}

func compoundPathHasTerminalStore(segs []string) bool {
	for i := len(segs) - 1; i >= 0; i-- {
		switch segs[i] {
		case segBang, segBangIndex, segBangBrace, segBangMulti:
			return true
		}
	}
	return false
}

// compoundPathHasIdentRoot is true when the walk starts by materializing a dictionary
// ident (sess.config…, TokenType.…), not a parked receiver (.name) or collection affix.
func compoundPathHasIdentRoot(segs []string) bool {
	if len(segs) == 0 {
		return false
	}
	s := segs[0]
	return s != segAt && s != segDot && !compoundSegmentIsCollectionAffix(s)
}

func compoundDotFollowMode(segs []string, i int) compoundDotFollow {
	if i >= len(segs) {
		return compoundDotFollowTerminal
	}
	switch segs[i] {
	case segTilde:
		return compoundDotFollowRefSuffix
	case segCaret:
		return compoundDotFollowCaret
	case segAt, segBang:
		return compoundDotFollowAtBang
	default:
		return compoundDotFollowContinue
	}
}

func walkCompoundPath(w *compoundWalker, segs []string) error {
	if w == nil || w.f == nil {
		panic("walkCompoundPath: nil walker or RunFrame")
	}
	if w.f.Task == nil || w.f.RP == nil {
		panic("walkCompoundPath: nil task or RunProvider")
	}
	p := w.f.RP.Parser

	// Leading "l" = locals container root (not OpLocals / DS). Bare axiom "l" stays a separate token.
	// Leading "pls" = Pip Local Storage root (not OpPls / DS). Bare axiom "pls" stays a separate token.
	if len(segs) > 0 && segs[0] == "l" {
		w.localsRoot = true
		segs = segs[1:]
	} else if len(segs) > 0 && segs[0] == "pls" {
		w.plsRoot = true
		segs = segs[1:]
	}

	if len(segs) == 0 {
		return fmt.Errorf("compound: empty path")
	}
	if !compoundPathHasMemberAccess(segs) {
		return fmt.Errorf("compound: not a member-access path")
	}
	if err := validateCompoundPathSegments(segs); err != nil {
		return err
	}
	if w.sel == csDataStack && !w.localsRoot && !w.plsRoot && !compoundPathHasIdentRoot(segs) && compoundPathHasTerminalStore(segs) {
		return fmt.Errorf("compound: store path requires t:/x:/y:/z:/o: prefix")
	}

	var focus any
	i := 0

	// Leading @: fetch from PropertyRef on receiver stack; value becomes path root (ref elided).
	if segs[0] == segAt {
		var err error
		focus, err = compoundFetchInitialAt(w)
		if err != nil {
			return err
		}
		i = 1
		if i >= len(segs) {
			compoundTerminalSync(w, focus, compoundTerminalValue)
			return nil
		}
	} else if segs[0] == segDot {
		var err error
		if w.localsRoot {
			focus, err = compoundBindLocalsRoot(w)
		} else if w.plsRoot {
			focus, err = compoundBindPlsRoot(w)
		} else {
			focus, err = compoundBindRootReceiver(w)
		}
		if err != nil {
			return err
		}
	} else if compoundSegmentIsCollectionAffix(segs[0]) {
		var err error
		if w.localsRoot {
			focus, err = compoundBindLocalsRoot(w)
		} else if w.plsRoot {
			focus, err = compoundBindPlsRoot(w)
		} else {
			focus, err = compoundBindCollectionRoot(w)
		}
		if err != nil {
			return err
		}
	} else {
		if w.localsRoot {
			return fmt.Errorf("compound: l. requires .member or collection affix")
		}
		if w.plsRoot {
			return fmt.Errorf("compound: pls. requires .member or collection affix")
		}
		v, err := compoundMaterializeIdent(w, p, w.f.Task, segs[0])
		if err != nil {
			return err
		}
		focus = v
		i = 1
	}

	terminal := compoundTerminalNone

	for i < len(segs) {
		seg := segs[i]

		switch seg {
		case segDot:
			i++
			if i >= len(segs) {
				return fmt.Errorf("compound: %q missing member name", segDot)
			}
			name := segs[i]
			i++
			follow := compoundDotFollowMode(segs, i)
			var err error
			focus, terminal, err = compoundDotMember(w, focus, name, follow)
			if err != nil {
				return err
			}
			if follow == compoundDotFollowRefSuffix {
				i++ // consume ~
				if i >= len(segs) {
					if _, ok := focus.(*MethodRef); ok {
						compoundTerminalRef(w, focus, compoundTerminalMethodRef)
					} else {
						compoundTerminalRef(w, focus, compoundTerminalPropertyRef)
					}
					return nil
				}
				return fmt.Errorf("compound: %q must be the terminal segment", segTilde)
			}
			if terminal == compoundTerminalMethodDone {
				return nil
			}
		case segTilde:
			return fmt.Errorf("compound: %q must immediately follow .<name>", segTilde)
		case segCaret:
			if i != len(segs)-1 {
				return fmt.Errorf("compound: %q must be the terminal segment", segCaret)
			}
			if err := compoundInvokeMethodRef(w, focus); err != nil {
				return err
			}
			return nil
		case segAt:
			var err error
			focus, err = compoundFetchFocusPath(w, focus)
			if err != nil {
				return err
			}
			i++
			if i >= len(segs) {
				compoundTerminalSync(w, focus, compoundTerminalValue)
				return nil
			}
			terminal = compoundTerminalNone
		case segBang:
			if err := compoundStoreFocusPath(w, focus); err != nil {
				return err
			}
			i++
			return nil
		case segAtIndex:
			var err error
			focus, err = compoundFetchCollection(w, focus, true, false)
			if err != nil {
				return err
			}
			i++
			if i >= len(segs) {
				compoundTerminalSync(w, focus, compoundTerminalValue)
				return nil
			}
		case segBangIndex:
			if err := compoundStoreCollection(w, focus, true, false); err != nil {
				return err
			}
			i++
			return nil
		case segAtMulti:
			var err error
			focus, err = compoundFetchMultiIndex(w, focus)
			if err != nil {
				return err
			}
			i++
			if i >= len(segs) {
				compoundTerminalSync(w, focus, compoundTerminalValue)
				return nil
			}
		case segBangMulti:
			if err := compoundStoreMultiIndex(w, focus); err != nil {
				return err
			}
			i++
			return nil
		case segAtBrace:
			var err error
			focus, err = compoundFetchCollection(w, focus, false, true)
			if err != nil {
				return err
			}
			i++
			if i >= len(segs) {
				compoundTerminalSync(w, focus, compoundTerminalValue)
				return nil
			}
		case segQueryBrace:
			var err error
			focus, err = compoundQueryCollection(w, focus, false, true)
			if err != nil {
				return err
			}
			i++
			if i >= len(segs) {
				compoundTerminalSync(w, focus, compoundTerminalValue)
				return nil
			}
		case segBangBrace:
			if err := compoundStoreCollection(w, focus, false, true); err != nil {
				return err
			}
			i++
			return nil
		default:
			return fmt.Errorf("compound: unexpected segment %q in member path", seg)
		}
	}

	compoundTerminalSync(w, focus, terminal)
	return nil
}

// compoundBindRootReceiver binds the path root receiver into focus; receiver stays on root stack.
func compoundBindRootReceiver(w *compoundWalker) (focus any, err error) {
	return compoundCopyRootReceiver(w)
}

// compoundBindLocalsRoot binds ColonFrameMark.GetLocals() into focus (no stack / OpLocals / DS).
func compoundBindLocalsRoot(w *compoundWalker) (focus any, err error) {
	rs := w.f.RP.ReturnStack
	return liftValue(w.f.RP, rs.Peek().GetLocals()), nil
}

// compoundBindPlsRoot binds Task.GetPls() into focus (no stack / OpPls / DS).
func compoundBindPlsRoot(w *compoundWalker) (focus any, err error) {
	if w == nil || w.f == nil || w.f.Task == nil {
		return nil, fmt.Errorf("compound: pls. requires Task")
	}
	return liftValue(w.f.RP, w.f.Task.GetPls()), nil
}

func compoundPeekRootCell(w *compoundWalker) (any, bool) {
	stack := w.rootStack()
	for tosIdx := 0; tosIdx < stack.Len(); tosIdx++ {
		switch (*stack.AddrAt(tosIdx)).(type) {
		case *PropertyRef, *MethodRef:
			continue
		default:
			return *stack.AddrAt(tosIdx), true
		}
	}
	return nil, false
}

// compoundCopyRootReceiver copies the first non-Ref cell at TOS on the receiver stack into focus.
func compoundCopyRootReceiver(w *compoundWalker) (focus any, err error) {
	cell, ok := compoundPeekRootCell(w)
	if !ok {
		return nil, fmt.Errorf("compound: no receiver on context stack")
	}
	return liftValue(w.f.RP, cell), nil
}

// compoundBindCollectionRoot binds the collection receiver into focus at path entry.
// Receiver is TOS on getStack(rp, sel); after copy to focus it is elided from the root stack
// so index/key operands on DS are at TOS for collection invoke.
func compoundBindCollectionRoot(w *compoundWalker) (focus any, err error) {
	focus, err = compoundBindRootReceiver(w)
	if err != nil {
		return nil, err
	}
	w.rootStack().Pop()
	return focus, nil
}

// compoundFetchInitialAt implements a path-initial @: ( PropertyRef -- value ), ref elided.
func compoundFetchInitialAt(w *compoundWalker) (any, error) {
	stack := w.rootStack()
	if stack.Len() < 1 {
		return nil, fmt.Errorf("compound: leading %q requires PropertyRef on context stack", segAt)
	}
	ref, ok := stack.Peek().(*PropertyRef)
	if !ok {
		return nil, fmt.Errorf("compound: leading %q requires PropertyRef, got %T", segAt, stack.Peek())
	}
	stack.Pop()
	return compoundFetchPropertyValue(w, ref.Parent, ref.PM), nil
}

func compoundMaterializeIdent(w *compoundWalker, p *Parser, t *Task, ident string) (any, error) {
	rp := w.f.RP
	before := rp.DataStack.Len()
	word, xt, found := FindNameInTask(t, rp, String(ident))
	if !found {
		return nil, fmt.Errorf("unknown word %q", ident)
	}
	var err error
	if word.RequiresStackContext() {
		err = compoundExecWordWithSel(w, word)
	} else {
		err = execInterpretWord(t, rp, word, xt)
	}
	if err != nil {
		return nil, err
	}
	if rp.DataStack.Len() <= before {
		return nil, fmt.Errorf("compound ident %q left no value on DS", ident)
	}
	return rp.DataStack.Pop(), nil
}

// compoundExecWordWithSel arms loop cs from walker sel then execs one ContextRequired word.
func compoundExecWordWithSel(w *compoundWalker, word *Word) error {
	if w == nil || w.f == nil || w.f.RP == nil || w.f.Task == nil {
		return fmt.Errorf("compoundExecWordWithSel: nil walker")
	}
	if word == nil {
		return fmt.Errorf("compoundExecWordWithSel: nil word")
	}
	if err := checkOStackExecAllowed(w.sel, word); err != nil {
		return err
	}
	rp := w.f.RP
	return runExecWordWithCS(w.f.Task, rp, word, w.sel)
}

func compoundFocusThs(w *compoundWalker, focus any) *TypeObjPair {
	switch v := focus.(type) {
	case *TypeObjPair:
		return v
	case *PropertyRef:
		return v.Parent
	case *MethodRef:
		return v.Parent
	default:
		return liftValue(w.f.RP, focus)
	}
}

func compoundFetchPropertyValue(w *compoundWalker, ths *TypeObjPair, pm *PropertyMetadata) any {
	ds := w.f.RP.DataStack
	before := ds.Len()
	pm.invokeGet(w.f, ths)
	if ds.Len() <= before {
		panic(fmt.Errorf("compound: property %q get left no value on DS", pm.name))
	}
	return ds.Pop()
}

func compoundDotMember(w *compoundWalker, focus any, name string, follow compoundDotFollow) (any, compoundTerminalKind, error) {
	ths := compoundFocusThs(w, focus)
	t := ths.Typ
	host := classDispatchHost(ths)

	if host != nil {
		if v, kind, ok := compoundClassAccessor(w, ths, host, name, follow); ok {
			return v, kind, nil
		}
	}

	if pm := t.GetPropertyMetadata(name); pm != nil {
		return compoundDotProperty(w, ths, pm, follow)
	}
	if mm := t.GetMethodMetadata(name); mm != nil {
		return compoundDotMethod(w, ths, mm, follow)
	}
	errType := t.name
	if host != nil {
		errType = host.name
	}
	return nil, compoundTerminalNone, fmt.Errorf("%s: no accessor %q", errType, name)
}

// assertCompoundMemberSR denies a restricted property/method against the task mask.
func assertCompoundMemberSR(w *compoundWalker, restrict SecurityRestrictions) {
	if w == nil {
		assertMemberRestrict(nil, restrict)
		return
	}
	assertMemberRestrict(w.f, restrict)
}

func compoundDotProperty(w *compoundWalker, ths *TypeObjPair, pm *PropertyMetadata, follow compoundDotFollow) (any, compoundTerminalKind, error) {
	assertCompoundMemberSR(w, pm.restrict)
	switch follow {
	case compoundDotFollowRefSuffix, compoundDotFollowAtBang:
		return &PropertyRef{Parent: ths, PM: pm}, compoundTerminalNone, nil
	default:
		v := compoundFetchPropertyValue(w, ths, pm)
		if follow == compoundDotFollowTerminal {
			return v, compoundTerminalValue, nil
		}
		return v, compoundTerminalNone, nil
	}
}

func compoundDotMethod(w *compoundWalker, ths *TypeObjPair, mm *MethodMetadata, follow compoundDotFollow) (any, compoundTerminalKind, error) {
	assertCompoundMemberSR(w, mm.restrict)
	switch follow {
	case compoundDotFollowRefSuffix:
		return &MethodRef{Parent: ths, MM: mm}, compoundTerminalNone, nil
	case compoundDotFollowCaret:
		return &MethodRef{Parent: ths, MM: mm}, compoundTerminalNone, nil
	default:
		return nil, compoundTerminalNone, fmt.Errorf("%s: bare .%s requires %q to invoke", ths.Typ.name, mm.name, segCaret)
	}
}

func compoundClassAccessor(w *compoundWalker, ths *TypeObjPair, host *Type, name string, follow compoundDotFollow) (any, compoundTerminalKind, bool) {
	if pm := host.GetClassPropertyMetadata(name); pm != nil {
		v, kind, err := compoundDotProperty(w, ths, pm, follow)
		if err != nil {
			panic(err)
		}
		return v, kind, true
	}
	if mm := host.GetClassMethodMetadata(name); mm != nil {
		v, kind, err := compoundDotMethod(w, ths, mm, follow)
		if err != nil {
			return nil, compoundTerminalNone, false
		}
		return v, kind, true
	}
	return nil, compoundTerminalNone, false
}

func compoundInvokeMethodRef(w *compoundWalker, focus any) error {
	ref, ok := focus.(*MethodRef)
	if !ok {
		panic(fmt.Errorf("^: focus is %T, want MethodRef", focus))
	}
	ref.MM.invokeCall(w.f, ref.Parent)
	return nil
}

// compoundTerminalRef publishes a terminal PropertyRef or MethodRef on DS (focus-owned ref).
func compoundTerminalRef(w *compoundWalker, focus any, kind compoundTerminalKind) {
	if kind != compoundTerminalPropertyRef && kind != compoundTerminalMethodRef {
		panic(fmt.Errorf("compoundTerminalRef: kind %v", kind))
	}
	compoundTerminalSync(w, focus, kind)
}

// compoundTerminalSync pushes the terminal product on DS; rooted receivers remain on the receiver stack.
func compoundTerminalSync(w *compoundWalker, focus any, kind compoundTerminalKind) {
	if kind == compoundTerminalNone || kind == compoundTerminalMethodDone || kind == compoundTerminalStoreDone {
		if focus == nil {
			return
		}
		if kind == compoundTerminalNone {
			w.f.RP.DataStack.Push(focus)
		}
		return
	}
	ds := w.f.RP.DataStack
	switch kind {
	case compoundTerminalValue, compoundTerminalPropertyRef, compoundTerminalMethodRef:
		ds.Push(focus)
	}
}

func compoundFetchFocusPath(w *compoundWalker, focus any) (any, error) {
	ref, ok := focus.(*PropertyRef)
	if !ok {
		return nil, fmt.Errorf("@: focus is %T, want PropertyRef (use .prop~@ or implied read via .prop)", focus)
	}
	return compoundFetchPropertyValue(w, ref.Parent, ref.PM), nil
}

// compoundStoreFocusPath stores via an explicit ! segment; requires PropertyRef focus (path one-shot).
func compoundStoreFocusPath(w *compoundWalker, focus any) error {
	ref, ok := focus.(*PropertyRef)
	if !ok {
		return fmt.Errorf("!: focus is %T, want PropertyRef (use .prop~!)", focus)
	}
	ref.PM.invokeSet(w.f, ref.Parent)
	return nil
}

func compoundFetchCollection(w *compoundWalker, focus any, byOrdinal, byKey bool) (any, error) {
	ths := compoundFocusThs(w, focus)
	t := ths.Typ
	ds := w.f.RP.DataStack
	before := ds.Len()
	switch {
	case byOrdinal:
		t.invokeItemsGetByOrdinal(w.f, ths)
	case byKey:
		t.invokeItemsGetByKey(w.f, ths)
	default:
		panic("compoundFetchCollection: no accessor")
	}
	after := ds.Len()
	// Net DS change: @()/@{} pop 1 push 1 (delta 0).
	if after-before != 0 {
		panic(fmt.Errorf("compound: collection fetch DS delta %d, want 0", after-before))
	}
	if after == 0 {
		panic("compound: collection fetch left no value on DS")
	}
	return ds.Pop(), nil
}

func compoundFetchMultiIndex(w *compoundWalker, focus any) (any, error) {
	ths := compoundFocusThs(w, focus)
	im := ths.Typ.requireItemsMetadata()
	if im.getByMultiIndex == nil || im.multiIndexArity < 1 {
		panic(fmt.Errorf("compound: %s has no @[]", ths.Typ.name))
	}
	ds := w.f.RP.DataStack
	before := ds.Len()
	ths.Typ.invokeItemsGetByMultiIndex(w.f, ths)
	after := ds.Len()
	// Pop N indexes, push 1 value → delta 1−N.
	wantDelta := 1 - im.multiIndexArity
	if after-before != wantDelta {
		panic(fmt.Errorf("compound: @[] DS delta %d, want %d", after-before, wantDelta))
	}
	if after == 0 {
		panic("compound: @[] left no value on DS")
	}
	return ds.Pop(), nil
}

func compoundStoreMultiIndex(w *compoundWalker, focus any) error {
	ths := compoundFocusThs(w, focus)
	ths.Typ.invokeItemsSetByMultiIndex(w.f, ths)
	return nil
}

func compoundQueryCollection(w *compoundWalker, focus any, byOrdinal, byKey bool) (any, error) {
	ths := compoundFocusThs(w, focus)
	t := ths.Typ
	ds := w.f.RP.DataStack
	before := ds.Len()
	switch {
	case byOrdinal:
		panic("compoundQueryCollection: ordinal probe not supported")
	case byKey:
		t.invokeItemsHasByKey(w.f, ths)
	default:
		panic("compoundQueryCollection: no accessor")
	}
	after := ds.Len()
	if after < before-1 {
		panic("compound: collection query underflow on DS")
	}
	if after == before-1 {
		panic("compound: collection query consumed operand without result")
	}
	if after == 0 && before == 0 {
		panic("compound: collection query left no value on DS")
	}
	return ds.Pop(), nil
}

func compoundStoreCollection(w *compoundWalker, focus any, byOrdinal, byKey bool) error {
	ths := compoundFocusThs(w, focus)
	t := ths.Typ
	switch {
	case byOrdinal:
		t.invokeItemsSetByOrdinal(w.f, ths)
	case byKey:
		t.invokeItemsSetByKey(w.f, ths)
	default:
		panic("compoundStoreCollection: no accessor")
	}
	return nil
}
