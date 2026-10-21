package core

import (
	"fmt"
	"math"
)

// compileEmitMode reports compile-time emission into the open colon body (STATE true).
func compileEmitMode(t *Task) bool {
	if t == nil || t.dc == nil {
		return false
	}
	return GetCompileState(t.compilerForTarget())
}

// emitBracketLiteral freezes val in DT and appends an emit-program fragment to the open colon:
// OpDoLit* (push frozen value when the host word runs) then OpCompileLit (pop and emit into client CT).
// Used for [[ … ]] regions; distinct from bootstrapEmitLiteral alone (runtime push only).
func emitBracketLiteral(cs *CompilerState, val any) {
	bootstrapEmitLiteral(cs, val)
	bootstrapEmitOpcode(cs, OpCompileLit)
}

// emitDoEmitWord appends OpDoEmit · VirtXT into the open colon (Emitting mode).
// Also virally ORs w's SR onto the building word (same as bootstrapEmitWordReference),
// so [[ gatedWord ]] stamps capability bits on the host colon being defined.
func emitDoEmitWord(cs *CompilerState, w *Word) error {
	if cs == nil || w == nil {
		return fmt.Errorf("emitDoEmitWord: nil argument")
	}
	if !GetCompileState(cs) {
		return fmt.Errorf("emitDoEmitWord: not in compile state")
	}
	if b := cs.DC.dict.PeekBuilding(); b != nil {
		accumulateSRFrom(b, w)
	}
	xt, ok := cs.DC.dict.XTOf(w)
	if !ok {
		return fmt.Errorf("emitDoEmitWord: no xt for word %q", w.name)
	}
	bootstrapEmitOpcode(cs, OpDoEmit)
	_ = cs.emitCT(int(xt))
	return nil
}

// compileHere is the next CT append index for the in-flight colon body.
func (t *Task) compileHere(cs *CompilerState) int {
	return cs.ct().Len()
}

// compileAppend emits one int cell into the in-flight colon body.
func (t *Task) compileAppend(cs *CompilerState, val int) VirtualAddr {
	cs.DC.panicIfLockedOrSealed("emit CT")
	return cs.emitCT(val)
}

// compileEmitLiteralCells appends CT loader cells for v (OpCompileLit / compileLiteral path).
func compileEmitLiteralCells(t *Task, cs *CompilerState, dc *DictionaryContext, v any) {
	dc.panicIfLockedOrSealed("emit literal")
	dt := dc.dt
	ct := dc.ct
	emit := func(val int) {
		if compileEmitMode(t) {
			t.compileAppend(cs, val)
			return
		}
		ct.Append(val)
	}
	if v == nil {
		emit(int(OpDoLitNull))
		return
	}
	switch v := v.(type) {
	case Bool:
		emit(int(OpDoLitB))
		b := 0
		if v {
			b = 1
		}
		emit(b)
	case Int:
		if v == 0 {
			emit(int(OpDoLit0))
		} else {
			emit(int(OpDoLitI))
			emit(int(v))
		}
	case Uint:
		emit(int(OpDoLitU))
		emit(int(v))
	case Float:
		emit(int(OpDoLitF))
		emit(int(math.Float64bits(v)))
	case String:
		idx := dt.Append(v)
		emit(int(OpDoLit))
		emit(idx)
	default:
		panic(fmt.Errorf("Invalid literal, type: %T", v))
	}
}
