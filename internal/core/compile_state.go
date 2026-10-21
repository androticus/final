package core

import "fmt"

type CompileStateFlags = int

const ( // NOTE: no iota, and values must remain idempotent
	CompileStateFlagCompiling  CompileStateFlags = 0x1
	CompileStateFlagInColonDef CompileStateFlags = 0x2
	CompileStateFlagEmitting   CompileStateFlags = 0x4
)

// GetCompileState reports whether compile-time emission into CT is active (Forth STATE).
func GetCompileState(cs *CompilerState) Bool {
	return (cs.compileState & CompileStateFlagCompiling) != 0
}

// SetCompileState sets compile-time emission mode on cs.
func SetCompileState(cs *CompilerState, compiling bool) {
	if compiling {
		if cs != nil && cs.DC != nil {
			cs.DC.panicIfLockedOrSealed("enter compile")
		}
		cs.compileState |= CompileStateFlagCompiling
	} else {
		cs.compileState &= ^CompileStateFlagCompiling
	}
}

func GetInColonDef(cs *CompilerState) bool {
	return (cs.compileState & CompileStateFlagInColonDef) != 0
}

func SetInColonDef(cs *CompilerState, inColonDef bool) {
	if inColonDef {
		cs.compileState |= CompileStateFlagInColonDef
	} else {
		cs.compileState &= ^CompileStateFlagInColonDef
	}
}

// GetEmitting reports whether [[ … ]] emit-program token dispatch is active.
func GetEmitting(cs *CompilerState) Bool {
	return (cs.compileState & CompileStateFlagEmitting) != 0
}

// SetEmitting sets [[ … ]] emit-program mode on cs (requires Compiling for meaningful use).
func SetEmitting(cs *CompilerState, emitting bool) {
	if emitting {
		cs.compileState |= CompileStateFlagEmitting
	} else {
		cs.compileState &= ^CompileStateFlagEmitting
	}
}

// Compiler Block Verification

type BlockKind = int

// NOTE: no iota, and values must remain idempotent
// since we use the literal int values in early bootstrap code
// even block closers get a value, even if they don't emit tokens

const (
	// Colon definition (: pushes, ; asserts)
	BlockKindColon BlockKind = 1

	// if / else / then
	BlockKindIf   BlockKind = 2
	BlockKindElse BlockKind = 3
	BlockKindThen BlockKind = 4

	// begin … again / until; begin … while … repeat
	BlockKindBegin    BlockKind = 5
	BlockKindBreak    BlockKind = 6 // mid-structure (forward exit mark)
	BlockKindContinue BlockKind = 7 // mid-structure (stretch: loop continue)
	BlockKindAgain    BlockKind = 8
	BlockKindUntil    BlockKind = 9
	BlockKindWhile    BlockKind = 10
	BlockKindRepeat   BlockKind = 11

	// foreach … /foreach (i/j/k are accessors — no BlockKind)
	BlockKindForeach    BlockKind = 12
	BlockKindEndForeach BlockKind = 13

	// switch … /switch (a/b/c accessors — no BlockKind); target on SwitchTargets
	BlockKindSwitch    BlockKind = 14
	BlockKindIfdo      BlockKind = 15
	BlockKindEndIfdo   BlockKind = 16
	// 17 was ExitSwitch (removed; mid-arm early exit via if)
	BlockKindElsedo    BlockKind = 18
	BlockKindEndElsedo BlockKind = 19
	BlockKindEndSwitch BlockKind = 20

	// varargs ( … ) — slice-006 step 6b
	BlockKindOpenParen  BlockKind = 21
	BlockKindCloseParen BlockKind = 22

	KindSemicolon BlockKind = 23

	BlockKind1OpenBracket  BlockKind = 24
	BlockKind1CloseBracket BlockKind = 25
	BlockKind2OpenBracket  BlockKind = 26
	BlockKind2CloseBracket BlockKind = 27

	// 28–30 reserved (was beginDefinerWord / endDefinerWord / does> tranche)

	// forever … /forever (infinite loop; continue-capable)
	BlockKindForever BlockKind = 31

	// try / catch / finally / /try
	BlockKindTry     BlockKind = 32
	BlockKindCatch   BlockKind = 33
	BlockKindFinally BlockKind = 34
	BlockKindEndTry  BlockKind = 35

	// :s … ;s subword body (187); sits on BlockKindColon
	BlockKindSubword BlockKind = 36

	// [imm imm] / [[ ]] — interpret-in-compile / emit islands (BlockKind 24–27)
)

type BlockToken struct {
	kind BlockKind
	// mark is a CT index for continue back-target; -1 if unset.
	mark int
}

func NewBlockToken(kind BlockKind) *BlockToken {
	return &BlockToken{kind: kind, mark: -1}
}

func (bt *BlockToken) Kind() BlockKind {
	return bt.kind
}

func (bt *BlockToken) Mark() int {
	return bt.mark
}

func (bt *BlockToken) SetMark(mark int) {
	bt.mark = mark
}

// hasItemFromAny is satisfied by *List[T] for HasItemFromAny.
type hasItemFromAny interface {
	HasItemFromAny(v any) bool
}

// panicRequireNKind walks TS from TOS toward KindColon and requires n blocks of kind.
// Inside a subword (:s), the check is skipped — i/j/k and a/b/c are runtime-only there.
func panicRequireNKind(ts *Stack[any], n int, msg string, kind BlockKind) {
	for i := 0; i < ts.Len(); i++ {
		block := ts.At(i).(*BlockToken)
		if block.Kind() == BlockKindSubword {
			return
		}
		if block.Kind() == BlockKindColon {
			break
		}
	}
	remaining := n
	for i := 0; i < ts.Len() && remaining > 0; i++ {
		block := ts.At(i).(*BlockToken)
		if block.Kind() == kind {
			remaining--
		}
		if block.Kind() == BlockKindColon {
			break
		}
	}
	if remaining > 0 {
		panic(fmt.Errorf("%s", msg))
	}
}

// panicRequireOneOfKind walks TS from TOS toward KindColon. The first block whose kind
// is in kinds succeeds (innermost allow-listed frame). Otherwise panics with msg.
func panicRequireOneOfKind(ts *Stack[any], msg string, kinds hasItemFromAny) {
	for i := 0; i < ts.Len(); i++ {
		block := ts.At(i).(*BlockToken)
		if block.Kind() == BlockKindColon {
			break
		}
		if kinds.HasItemFromAny(Int(block.Kind())) {
			return
		}
	}
	panic(fmt.Errorf("%s", msg))
}

// panicIfSwitchNotInArm enforces [162]: a new switch may open outside any switch, or inside
// case/default — not directly under switch / between arms / after /default.
// Walks TS from TOS: case(15) or default(18) first → ok; switch(14) first → panic; colon → ok.
func panicIfSwitchNotInArm(ts *Stack[any], msg string) {
	for i := 0; i < ts.Len(); i++ {
		block := ts.At(i).(*BlockToken)
		switch block.Kind() {
		case BlockKindColon:
			return
		case BlockKindIfdo, BlockKindElsedo:
			return
		case BlockKindSwitch:
			panic(fmt.Errorf("%s", msg))
		}
	}
}

// acceptDropBlock drops TS TOS when it is a BlockToken of kind; no-op if TS empty or kind mismatch.
func acceptDropBlock(ts *Stack[any], kind BlockKind) {
	if ts.Len() == 0 {
		return
	}
	block, ok := ts.Peek().(*BlockToken)
	if !ok || block.Kind() != kind {
		return
	}
	ts.Pop()
}

// panicDropBlock panics with msg if TS is empty or top block kind mismatches; else drops top.
func panicDropBlock(ts *Stack[any], msg string, kind BlockKind) {
	if ts.Len() == 0 {
		panic(fmt.Errorf("%s", msg))
	}
	block := ts.Peek().(*BlockToken)
	if block.Kind() != kind {
		panic(fmt.Errorf("%s", msg))
	}
	ts.Pop()
}

/* The Compile Block State checks use the following helper axioms:

// push the block token onto ts
pushBlock ( kind:i -- ) t( -- bt )

// allow the block, and drop if found
acceptDropBlock ( kind:i -- ) t( bt -- ? )

// we standardize on this form, since chicken/egg 'if' not compiled yet
// for compiling if/else/then, so we don't define a bool flag version
panicDropBlock ( msg:s kind:i) t( bt -- )

// panic if the block was not found, but don't drop it
// eg use-case is 'else' that requires a matching 'if'
panicRequireBlock ( msg:s kind:i) t( bt -- )

// require n blocks of kind under colon; no-op inside :s (ijk/abc runtime-only in subwords)
panicRequireNKind ( n:i msg:s kind:i -- ) t( -- )

// require innermost TS block kind ∈ kinds (List); skip other kinds until KindColon
panicRequireOneOfKind ( kinds:List msg:s -- ) t( -- )

// set continue/back mark on TS TOS BlockToken
setBlockMark ( mark:i -- ) t( bt -- bt )
*/
