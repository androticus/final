package core

import "fmt"

// registerBootstrapTransitionWord adds a hidden OpTransition worker: subCode indexes runtimeDispatchTable.
func registerBootstrapTransitionWord(name string, id RuntimeDispatchID, flags WordFlags) *Word {
	w := NewAxiomWord(name, OpTransition, flags)
	w.subCode = int(id)
	globalDictionaryContext.dict.Add(w)
	return w
}

// registerBootDeferWord adds a boot defer proxy: OpBootstrapDefer with a fixed reserved DT slot (no dt.Append).
// DT[slot] is written by wireBootstrapBootTargets or Final via coldBootTarget / warmBootTarget.
func registerBootDeferWord(name string, flags WordFlags, slot ReservedDTSlot) *Word {
	w := NewAxiomWord(name, OpBootstrapDefer, flags)
	w.dtIdx = slot
	globalDictionaryContext.dict.Add(w)
	return w
}

// registerDeferWord adds a defer proxy: OpBootstrapDefer worker with dtIdx = DT slot holding target XT.
func registerDeferWord(name string, flags WordFlags, initTarget *Word) *Word {
	if initTarget == nil {
		panic("registerDeferWord: nil initTarget")
	}
	d := globalDictionaryContext.dict
	targetXT, ok := d.XTOf(initTarget)
	if !ok {
		panic(fmt.Sprintf("registerDeferWord(%q): initTarget not in dictionary", name))
	}
	slot := globalDictionaryContext.dt.Append(Int(targetXT))
	w := NewAxiomWord(name, OpBootstrapDefer, flags)
	w.dtIdx = slot
	d.Add(w)
	return w
}
