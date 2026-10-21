package core

import (
	"fmt"
	"strings"
)

func compoundPathSegments(path String) []string {
	s := strings.TrimSpace(string(path))
	if s == "" {
		return nil
	}
	return strings.Split(s, " ")
}

func dispatchSimplePrefixedOpInterpret(p *Parser, t *Task, rp *RunProvider, sel ContextStackIndex, ident string) error {
	w, _, found := FindNameInTask(t, rp, String(ident))
	if !found {
		return fmt.Errorf("unknown word %q", ident)
	}
	if !w.RequiresStackContext() {
		return fmt.Errorf("%q does not accept a stack selector prefix", ident)
	}
	if err := checkInterpretWordAllowed(t, w); err != nil {
		return err
	}
	return runExecWordWithCS(t, rp, w, sel)
}

func compileSimplePrefixedOp(_ *Parser, t *Task, rp *RunProvider, cs *CompilerState, cp *compoundPathValue, ident string) error {
	w, _, found := FindNameInTask(t, rp, String(ident))
	if !found {
		return fmt.Errorf("unknown word %q", ident)
	}
	if !w.RequiresStackContext() {
		return fmt.Errorf("%q does not accept a stack selector prefix", ident)
	}
	if err := checkOStackExecAllowed(cp.stackSelector, w); err != nil {
		return err
	}
	return emitCompoundPathCSSandwichCompile(t, cp, func() error {
		if w.IsImmediate() {
			return dispatchExecWord(t, rp, w, false)
		}
		return bootstrapEmitWordReference(cs, w, cp.stackSelector)
	})
}
