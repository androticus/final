package core

import "fmt"

func init() {
	runtimeDispatchTable[TrFind] = bootstrapFinalFind
	runtimeDispatchTable[TrNextToken] = bootstrapFinalGetNextToken
	runtimeDispatchTable[TrSessionDispatch] = bootstrapDispatchSessionToken
	runtimeDispatchTable[TrReplError] = bootstrapFinalReplError
	runtimeDispatchTable[TrWarmReset] = bootstrapDispatchWarmReset
	runtimeDispatchTable[TrCheckResetByeReq] = bootstrapDispatchCheckResetByeReq
	runtimeDispatchTable[TrBootColon] = bootstrapDispatchBootColon
	runtimeDispatchTable[TrBootSemicolon] = bootstrapDispatchBootSemicolon
	runtimeDispatchTable[TrBootSubwordOpen] = bootstrapDispatchBootSubwordOpen
	runtimeDispatchTable[TrBootSubwordClose] = bootstrapDispatchBootSubwordClose
	runtimeDispatchTable[FinalFind] = bootstrapFinalFind
	runtimeDispatchTable[FinalCurrentDictHasWord] = bootstrapFinalCurrentDictHasWord
	runtimeDispatchTable[FinalReplError] = bootstrapFinalReplError
	runtimeDispatchTable[FinalGetNextToken] = bootstrapFinalGetNextToken
	runtimeDispatchTable[FinalReadIdentifier] = bootstrapFinalReadIdentifier
	runtimeDispatchTable[FinalValidatePlainIdent] = bootstrapFinalValidatePlainIdent
	runtimeDispatchTable[FinalHandleCompound] = bootstrapFinalHandleCompound
	runtimeDispatchTable[FinalCompileFromFile] = bootstrapFinalCompileFromFile
	runtimeDispatchTable[FinalCompileFromManifest] = bootstrapFinalCompileFromManifest
	runtimeDispatchTable[FinalSetCompileState] = bootstrapFinalSetCompileState
	runtimeDispatchTable[FinalFmt] = bootstrapFinalFmt
	runtimeDispatchTable[FinalMintVarWord] = bootstrapMintVarWord
	runtimeDispatchTable[FinalMintConstWord] = bootstrapMintConstWord
	runtimeDispatchTable[FinalMintDeferWord] = bootstrapMintDeferWord
	runtimeDispatchTable[FinalForgetMark] = bootstrapDispatchForgetMark
}

// bootstrapDispatchForgetMark handles interpret-time mark create for forget: (core1 wrapper).
func bootstrapDispatchForgetMark(f *RunFrame) {
	if err := bootstrapForgetStart(f.RP.Parser, f.Task, f.RP); err != nil {
		bootstrapHandleReplError(f, err)
	}
}

// OpTransition handlers for the in-VM REPL driver (Phase 2c).
// Parser is on RunProvider (rp.Parser); core handlers do not write terminal I/O.

func bootstrapHandleReplError(f *RunFrame, err error) {
	if err == nil {
		return
	}
	bootstrapWriteREPLError(f.RP, err)
	skipParserLine(f.RP.Parser)
	if f.Task != nil && f.Task.dc != nil {
		SetCompileState(f.Task.compilerForTarget(), false)
	}
}

// bootstrapDispatchSessionToken pops ( value type -- ) and dispatches one token (reentrant handler).
func bootstrapDispatchSessionToken(f *RunFrame) {
	ds := f.RP.DataStack
	if ds.Len() < 2 {
		bootstrapHandleReplError(f, fmt.Errorf("session dispatch: expected value and TokenType on stack"))
		return
	}
	typ, err := popTokenType(ds)
	if err != nil {
		bootstrapHandleReplError(f, err)
		return
	}
	val := ds.Pop()
	if err := dispatchSessionParsedInRun(f.RP.Parser, f.Task, f.RP, val, typ, f); err != nil {
		bootstrapHandleReplError(f, err)
	}
}

// bootstrapDispatchCheckResetByeReq implements test-and-clear for the in-VM repl loop head.
func bootstrapDispatchCheckResetByeReq(f *RunFrame) {
	if f == nil || f.RP == nil {
		panic("bootstrapDispatchCheckResetByeReq: nil RunFrame or RunProvider")
	}
	f.RP.DataStack.Push(CheckResetCoreByeReq())
}

// bootstrapDispatchBootColon handles interpret-time `:` (defer target for ":").
func bootstrapDispatchBootColon(f *RunFrame) {
	if err := bootstrapColonStart(f.RP.Parser, f.Task, f.RP); err != nil {
		bootstrapHandleReplError(f, err)
	}
}

// bootstrapDispatchCompilePunct runs a compile punctuation defer target; errors propagate to the Run loop.
func bootstrapDispatchCompilePunct(f *RunFrame, name string) {
	cs := f.Task.compilerForTarget()
	if err := runCompileImmediate(f.Task, f.RP, name, cs); err != nil {
		panic(err)
	}
}

// bootstrapDispatchBootSemicolon handles compile-time `;` (defer target).
func bootstrapDispatchBootSemicolon(f *RunFrame) {
	bootstrapDispatchCompilePunct(f, ";")
}

// bootstrapDispatchBootSubwordOpen handles compile-time `:s` (defer target).
func bootstrapDispatchBootSubwordOpen(f *RunFrame) {
	bootstrapDispatchCompilePunct(f, ":s")
}

// bootstrapDispatchBootSubwordClose handles compile-time `;s` (defer target).
func bootstrapDispatchBootSubwordClose(f *RunFrame) {
	bootstrapDispatchCompilePunct(f, ";s")
}

func runRuntimeDispatch(id RuntimeDispatchID, f *RunFrame) {
	if id < 0 || id >= runtimeDispatchCount {
		panic(fmt.Errorf("runRuntimeDispatch: invalid id %d", id))
	}
	fn := runtimeDispatchTable[id]
	if fn == nil {
		panic(fmt.Errorf("runRuntimeDispatch: unset handler for id %d", id))
	}
	fn(f)
}
