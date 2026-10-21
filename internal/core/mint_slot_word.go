package core

import "fmt"

// Slot-word mint kinds for primitive var: / const: / defer: (no does>).
type mintSlotKind int

const (
	mintSlotVar mintSlotKind = iota
	mintSlotConst
	mintSlotDefer
)

// requireInterpretOnlyCreate panics if compile-time emission is active.
// var: / const: / defer: may only create words in the interpreter.
func requireInterpretOnlyCreate(t *Task, form string) {
	if t == nil {
		panic(fmt.Errorf("%s: nil Task", form))
	}
	cs := t.compilerForTarget()
	if cs != nil && GetCompileState(cs) {
		panic(fmt.Errorf("%s: only valid in the interpreter (not while compiling)", form))
	}
}

// mintSlotWordFromStacks pops name (and init for var/const), allots host DT, bakes
// instance CT, and Adds the word to the host dictionary.
// Stack: var/const ( init:a name:s -- ); defer ( name:s -- ).
func mintSlotWordFromStacks(t *Task, rp *RunProvider, kind mintSlotKind) {
	form := mintSlotFormName(kind)
	requireInterpretOnlyCreate(t, form)
	if rp == nil {
		panic(fmt.Errorf("%s: nil RunProvider", form))
	}
	ds := rp.DataStack
	if ds.Len() < 1 {
		panic(fmt.Errorf("%s: missing name", form))
	}
	nameRaw := ds.Pop()
	name, ok := nameRaw.(String)
	if !ok {
		panic(fmt.Errorf("%s: name is %T, want String", form, nameRaw))
	}
	if err := ValidatePlainIdentName(name); err != nil {
		panic(err)
	}

	var init any
	switch kind {
	case mintSlotVar, mintSlotConst:
		if ds.Len() < 1 {
			panic(fmt.Errorf("%s: missing init value", form))
		}
		init = ds.Pop()
	case mintSlotDefer:
		init = Int(0) // unset VirtXT; OpDefer / exec panics until is
	default:
		panic(fmt.Errorf("%s: bad mint kind", form))
	}

	host := instanceHostDC(t)
	host.panicIfLockedOrSealed("mint slot word")
	idx := host.AllotDTIndex()
	host.dt.Write(idx, init)

	flags := WordFlagDefinedWord
	if kind == mintSlotDefer {
		flags |= WordFlagDefer
	}
	w := NewWord(string(name), flags)
	start := host.ct.Len()
	switch kind {
	case mintSlotConst:
		// OpDoLit · dtIdx · OpUnnest — idx only in CT (Word.dtIdx stays -1 so
		// armExecWord / compile emit do not also push a dtIdx prefix).
		host.ct.Append(int(OpDoLit))
		host.ct.Append(idx)
		host.ct.Append(int(OpUnnest))
	case mintSlotVar:
		w.dtIdx = idx
		host.ct.Append(int(OpDTRef))
		host.ct.Append(int(OpUnnest))
	case mintSlotDefer:
		w.dtIdx = idx
		host.ct.Append(int(OpDefer))
		host.ct.Append(int(OpUnnest))
	}
	w.colonCode = CTP(start)
	host.dict.Add(w)
}

func mintSlotFormName(kind mintSlotKind) string {
	switch kind {
	case mintSlotVar:
		return "var:"
	case mintSlotConst:
		return "const:"
	case mintSlotDefer:
		return "defer:"
	default:
		return "mintSlotWord"
	}
}

func bootstrapMintVarWord(f *RunFrame) {
	mintSlotWordFromStacks(f.Task, f.RP, mintSlotVar)
}

func bootstrapMintConstWord(f *RunFrame) {
	mintSlotWordFromStacks(f.Task, f.RP, mintSlotConst)
}

func bootstrapMintDeferWord(f *RunFrame) {
	mintSlotWordFromStacks(f.Task, f.RP, mintSlotDefer)
}
