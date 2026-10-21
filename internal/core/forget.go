package core

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

const leadingForgetPrefix = "forget: "

// Forget truncates this DC to the four keep-lengths stamped at mark create.
// Order: Final-method guard → defer hygiene → catalog → dict → CT → DT.
// Panics (no mutation) on sealed/locked/building, bad keeps, or Final-method refuse.
func (dc *DictionaryContext) Forget(dictKeep, ctKeep, dtKeep, typeKeep int) {
	if dc == nil {
		panic(fmt.Errorf("DictionaryContext.Forget: nil dc"))
	}
	dc.panicIfLockedOrSealed("forget")
	if dc.dict != nil && dc.dict.IsBuilding() {
		panic(fmt.Errorf("DictionaryContext %q: cannot forget while compiling", dc.name))
	}
	if dc.dict == nil || dc.ct == nil || dc.dt == nil || dc.tr == nil {
		panic(fmt.Errorf("DictionaryContext.Forget: incomplete DC %q", dc.name))
	}
	if dictKeep < 0 || dictKeep > dc.dict.Len() {
		panic(fmt.Errorf("forget: dictKeep %d out of range (len %d)", dictKeep, dc.dict.Len()))
	}
	if ctKeep < 0 || ctKeep > dc.ct.Len() {
		panic(fmt.Errorf("forget: ctKeep %d out of range (len %d)", ctKeep, dc.ct.Len()))
	}
	if dtKeep < 0 || dtKeep > dc.dt.Len() {
		panic(fmt.Errorf("forget: dtKeep %d out of range (len %d)", dtKeep, dc.dt.Len()))
	}
	if typeKeep < 0 || typeKeep > dc.tr.Len() {
		panic(fmt.Errorf("forget: typeKeep %d out of range (len %d)", typeKeep, dc.tr.Len()))
	}

	dc.refuseForgetIfSurvivingFinalMethods(ctKeep, typeKeep)
	dc.panicIfDeferTargetsWiped(dictKeep, ctKeep)

	dc.tr.RollbackTo(typeKeep)
	dc.dict.RollbackTo(dictKeep)
	dc.ct.Truncate(ctKeep)
	dc.dt.Truncate(dtKeep)
}

// refuseIfSurvivingFinalMethods panics when a type that survives the cut
// (typeIdx < typeKeep) still has a Final method whose body starts at/after ctKeep.
// errFmt must accept (typeName, methodName, ctKeep) like "%q %q %d".
func (dc *DictionaryContext) refuseIfSurvivingFinalMethods(ctKeep, typeKeep int, errFmt string) {
	if dc == nil || dc.tr == nil || typeKeep <= 0 {
		return
	}
	for i := 0; i < typeKeep; i++ {
		typ := dc.tr.GetByTypeIdx(i)
		if typ == nil {
			continue
		}
		if name := survivingFinalMethodPastCT(typ, ctKeep); name != "" {
			panic(fmt.Errorf(errFmt, typ.name, name, ctKeep))
		}
	}
}

// refuseForgetIfSurvivingFinalMethods refuses when a type that survives the cut
// (typeIdx < typeKeep) still has a Final method whose body starts at/after ctKeep.
func (dc *DictionaryContext) refuseForgetIfSurvivingFinalMethods(ctKeep, typeKeep int) {
	dc.refuseIfSurvivingFinalMethods(ctKeep, typeKeep,
		"forget: refused — type %q method %q has Final body past CT keep %d (mark not removed)")
}

func (dc *DictionaryContext) refuseCompileUnitRollbackIfSurvivingFinalMethods(ctKeep, typeKeep int) {
	dc.refuseIfSurvivingFinalMethods(ctKeep, typeKeep,
		"compile-unit rollback refused — type %q method %q has Final body past CT keep %d (unit not rolled back)")
}

func survivingFinalMethodPastCT(typ *Type, ctKeep int) string {
	if typ == nil {
		return ""
	}
	if name := finalMethodPastCT(typ.methodMetadata, ctKeep); name != "" {
		return name
	}
	return finalMethodPastCT(typ.classMethodMetadata, ctKeep)
}

func finalMethodPastCT(mm *MethodMetadataMap, ctKeep int) string {
	if mm == nil {
		return ""
	}
	for i := 0; i < mm.Len(); i++ {
		item := mm.ItemAt(i)
		m := item.Value
		if m == nil || !m.isFinalMethod {
			continue
		}
		if m.finalColonCode >= ctKeep {
			return item.Name
		}
	}
	return ""
}

// panicIfDeferTargetsWiped panics if a surviving defer word's DT XT points into
// the wiped dict or CT region of this DC.
func (dc *DictionaryContext) panicIfDeferTargetsWiped(dictKeep, ctKeep int) {
	for i := 0; i < dictKeep; i++ {
		w := dc.dict.GetByOrdinal(i)
		if w == nil || w.flags&WordFlagDefer == 0 || w.dtIdx < 0 {
			continue
		}
		raw := dc.dt.Read(w.dtIdx)
		xtInt, ok := raw.(Int)
		if !ok {
			continue
		}
		xt := VirtualAddr(xtInt)
		if xt == 0 {
			continue
		}
		moduleId, packedOff := DecodeVirt(xt)
		var base int
		switch {
		case moduleId == ModuleIDThis:
			base = packedOff
		case moduleId == dc.moduleId:
			base = packedOff - int(moduleId)
		default:
			continue // foreign module XT — out of scope for this DC's forget
		}
		if base < 0 {
			continue
		}
		if base >= dictKeep {
			panic(fmt.Errorf(
				"forget: refused — defer %q XT points into wiped dict (ord %d ≥ keep %d)",
				w.name, base, dictKeep,
			))
		}
		tw := dc.dict.GetByOrdinal(base)
		if tw != nil && tw.colonCode >= 0 && int(tw.colonCode) >= ctKeep {
			panic(fmt.Errorf(
				"forget: refused — defer %q target %q body past CT keep %d",
				w.name, tw.name, ctKeep,
			))
		}
	}
}

// stampForgetKeeps returns dict/CT/DT/type catalog lengths for compileForget.
func (dc *DictionaryContext) stampForgetKeeps() (dictKeep, ctKeep, dtKeep, typeKeep int) {
	return dc.dict.Len(), dc.ct.Len(), dc.dt.Len(), dc.tr.Len()
}

// emitCompileForget stamps four keeps into CT as OpDoLitI literals, then OpDoForget.
func emitCompileForget(cs *CompilerState) {
	if cs == nil || cs.DC == nil {
		panic(fmt.Errorf("compileForget: nil compiler state"))
	}
	if !GetCompileState(cs) {
		panic(fmt.Errorf("compileForget: not compiling"))
	}
	if cs.DC.dict.PeekBuilding() == nil {
		panic(fmt.Errorf("compileForget: no building word"))
	}
	cs.DC.panicIfLockedOrSealed("compileForget")
	dictKeep, ctKeep, dtKeep, typeKeep := cs.DC.stampForgetKeeps()
	bootstrapEmitOpcode(cs, OpDoLitI)
	_ = cs.emitCT(dictKeep)
	bootstrapEmitOpcode(cs, OpDoLitI)
	_ = cs.emitCT(ctKeep)
	bootstrapEmitOpcode(cs, OpDoLitI)
	_ = cs.emitCT(dtKeep)
	bootstrapEmitOpcode(cs, OpDoLitI)
	_ = cs.emitCT(typeKeep)
	bootstrapEmitOpcode(cs, OpDoForget)
}

// runCompileForget is the immediate compile-time axiom body.
func runCompileForget(t *Task) {
	if t == nil {
		panic(fmt.Errorf("compileForget: nil task"))
	}
	emitCompileForget(t.compilerForTarget())
}

// runDoForget pops typeKeep dtKeep ctKeep dictKeep (TOS first) and forgets.
func runDoForget(t *Task, rp *RunProvider) {
	if t == nil || rp == nil || t.dc == nil {
		panic(fmt.Errorf("doForget: nil task or provider"))
	}
	ds := rp.DataStack
	typeKeep := int(ds.Pop().(Int))
	dtKeep := int(ds.Pop().(Int))
	ctKeep := int(ds.Pop().(Int))
	dictKeep := int(ds.Pop().(Int))
	t.dc.Forget(dictKeep, ctKeep, dtKeep, typeKeep)
}

// parseLeadingForgetMarkLine reports a mark name when line is exactly "forget: name"
// (no leading whitespace, no trailing text). The newline from a file read is stripped.
func parseLeadingForgetMarkLine(line string) (name string, ok bool) {
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	rest, found := strings.CutPrefix(line, leadingForgetPrefix)
	if !found || rest == "" {
		return "", false
	}
	if strings.IndexFunc(rest, unicode.IsSpace) >= 0 {
		return "", false
	}
	if err := ValidatePlainIdentName(String(rest)); err != nil {
		return "", false
	}
	return rest, true
}

// peekLeadingForgetMarkName reads the first line of path and returns a forget: mark
// name when that line is exactly "forget: name".
func peekLeadingForgetMarkName(path string) (name string, ok bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()
	line, _ := bufio.NewReader(f).ReadString('\n')
	if len(line) == 0 {
		return "", false
	}
	return parseLeadingForgetMarkLine(line)
}

// maybeRunLeadingForgetMark runs an existing interpret-only mark named on line 1
// as "forget: name", before the compile-unit checkpoint. No-op when the line
// does not match, the mark is missing, a compile txn is already active, or STATE
// is true.
func maybeRunLeadingForgetMark(t *Task, rp *RunProvider, path string) error {
	if t == nil || t.dc == nil || t.dc.dict == nil || rp == nil {
		return nil
	}
	cs := t.compilerForTarget()
	if cs != nil && (cs.txnActive || GetCompileState(cs)) {
		return nil
	}
	name, ok := peekLeadingForgetMarkName(path)
	if !ok {
		return nil
	}
	w := t.dc.dict.GetByName(name)
	if w == nil || w.flags&WordFlagInterpretOnly == 0 {
		return nil
	}
	return runWordInlineExec(t, rp, w)
}

// bootstrapForgetStart handles interpret-time forget: create — read name, open mark word,
// stamp+emit via compileForget, finalize. Mark is InterpretOnly.
func bootstrapForgetStart(p *Parser, t *Task, rp *RunProvider) error {
	if p == nil || t == nil || rp == nil {
		return fmt.Errorf("forget:: nil parser, task, or provider")
	}
	cs := t.compilerForTarget()
	if GetCompileState(cs) {
		return fmt.Errorf("forget: while already compiling")
	}
	val, typ, err := p.Next()
	if err != nil {
		return err
	}
	if !IsPlainIdentToken(typ) {
		return fmt.Errorf("forget: expected mark name, got %v", typ)
	}
	name, ok := val.(String)
	if !ok {
		return fmt.Errorf("forget: expected String name, got %T", val)
	}
	nameStr := string(name)
	if err := ValidatePlainIdentName(name); err != nil {
		return err
	}
	w := bootstrapCreate(cs, nameStr)
	w.flags |= WordFlagInterpretOnly
	w.doc = p.GetClearLastComment()
	emitCompileForget(cs)
	// No OpUnnest in the mark body — doForget truncates CT then unnests at invoke.
	SetInColonDef(cs, false)
	cs.DC.dict.PopAddBuildingWordToParent()
	SetCompileState(cs, false)
	return nil
}
