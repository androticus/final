package core

import "fmt"

const semicolonBlockCheckMsg = `": (semicolon) requires all code blocks to be closed; eg if ... then"`

// runCompileImmediate executes a compile-only immediate bootstrap word (`;`, `:s`, `;s`).
func runCompileImmediate(t *Task, rp *RunProvider, name string, cs *CompilerState) error {
	if cs == nil {
		return fmt.Errorf("%q outside colon definition", name)
	}
	switch name {
	case ";":
		if !GetCompileState(cs) {
			return fmt.Errorf("; outside colon definition")
		}
		panicDropBlock(rp.TempStack, semicolonBlockCheckMsg, BlockKindColon)
		_ = rp.Parser.GetClearLastComment()
		bootstrapFinalizeColonWord(cs)
		return nil
	case ":s":
		return bootstrapSubwordStart(rp.Parser, t, rp)
	case ";s":
		return bootstrapSubwordEnd(t, rp)
	default:
		if !GetCompileState(cs) {
			return fmt.Errorf("%q outside colon definition", name)
		}
		return fmt.Errorf("unhandled compile immediate %q", name)
	}
}
