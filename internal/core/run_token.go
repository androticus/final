package core

import "fmt"

// handleParsedToken is the reentrant session token handler (parser-free for stream tokens).
// Compound expansion synthesizes (value, type) and recurses here. Compile-time definers
// that read ahead (e.g. ":") still use RunProvider.Parser when executed at interpret time.
// rf is the run-loop RunFrame when inRunLoop; required for CompileWord inside Run.
func handleParsedToken(p *Parser, t *Task, rp *RunProvider, val any, typ TokenType, inRunLoop bool, rf *RunFrame) error {
	cs := t.compilerForTarget()
	if GetCompileState(cs) {
		return handleParsedTokenCompile(p, t, rp, cs, val, typ, inRunLoop, rf)
	}
	return handleParsedTokenInterpret(p, t, rp, val, typ, inRunLoop)
}

func handleParsedTokenInterpret(p *Parser, t *Task, rp *RunProvider, val any, typ TokenType, inRunLoop bool) error {
	ds := rp.DataStack
	if typ == TokenEOF {
		return nil
	}
	if typ == TokenNewline {
		return nil
	}
	if HasTokenFlag(typ, TokenComment) {
		return nil
	}
	if IsLiteralTokenType(typ) {
		ds.Push(val)
		return nil
	}
	if HasTokenFlag(typ, TokenCompoundIdent) {
		return dispatchCompoundToken(p, t, rp, val)
	}
	if HasTokenFlag(typ, TokenIdent) {
		name, ok := val.(String)
		if !ok {
			return fmt.Errorf("interpret: expected String name, got %T", val)
		}
		w, _, found := FindNameInTask(t, rp, name)
		if !found {
			return fmt.Errorf("unknown word %q", name)
		}
		return dispatchExecWord(t, rp, w, inRunLoop)
	}
	return fmt.Errorf("interpret: unsupported token type %v", typ)
}

func checkInterpretWordAllowed(t *Task, w *Word) error {
	if t == nil || t.dc == nil || w == nil {
		return fmt.Errorf("checkInterpretWordAllowed: nil argument")
	}
	if t.dc.dict.IsBuilding() {
		return nil
	}
	if w.flags&WordFlagCompilerOnly != 0 {
		return fmt.Errorf("word %q is compiler-only", w.name)
	}
	if w.flags&WordFlagCompileOnly != 0 {
		return fmt.Errorf("word %q is compile-only", w.name)
	}
	return nil
}

func handleParsedTokenCompile(p *Parser, t *Task, rp *RunProvider, cs *CompilerState, val any, typ TokenType, inRunLoop bool, rf *RunFrame) error {
	if typ == TokenEOF {
		return nil
	}
	if typ == TokenNewline {
		return nil
	}
	if HasTokenFlag(typ, TokenComment) {
		return nil
	}
	if IsLiteralTokenType(typ) {
		if GetEmitting(cs) {
			emitBracketLiteral(cs, val)
			return nil
		}
		bootstrapEmitLiteral(cs, val)
		return nil
	}
	if HasTokenFlag(typ, TokenCompoundIdent) {
		if GetEmitting(cs) {
			return fmt.Errorf("compound token not supported in [[ … ]] emit mode")
		}
		return dispatchCompoundTokenCompile(p, t, rp, cs, val)
	}
	if HasTokenFlag(typ, TokenIdent) {
		name, ok := val.(String)
		if !ok {
			return fmt.Errorf("compile: expected String name, got %T", val)
		}
		if !GetEmitting(cs) {
			if handled, err := tryCompileSubwordRef(cs, string(name)); handled {
				return err
			}
		}
		w, _, found := FindNameInTask(t, rp, name)
		if !found {
			return fmt.Errorf("unknown word %q", name)
		}
		if GetEmitting(cs) {
			if w.ExecutesInEmitting() {
				return dispatchExecWord(t, rp, w, inRunLoop)
			}
			return emitDoEmitWord(cs, w)
		}
		if w.IsImmediate() {
			return dispatchExecWord(t, rp, w, inRunLoop)
		}
		if inRunLoop {
			if rf == nil {
				return fmt.Errorf("compile: nil RunFrame in run loop")
			}
			return w.CompileWord(rf)
		}
		return bootstrapEmitWordReference(cs, w)
	}
	return fmt.Errorf("compile: unsupported token type %v", typ)
}

func asCompoundPath(val any) (*compoundPathValue, bool) {
	cp, ok := val.(*compoundPathValue)
	if !ok || cp == nil {
		return nil, false
	}
	if cp.stackSelector >= 0 {
		return cp, true
	}
	return cp, cp.path != ""
}

// dispatchCompoundToken walks a compound path via the in-hand engine.
func dispatchCompoundToken(p *Parser, t *Task, rp *RunProvider, val any) error {
	cp, ok := asCompoundPath(val)
	if !ok {
		return fmt.Errorf("compound token not supported in session dispatch")
	}
	return walkCompoundSegments(p, t, rp, cp)
}

func dispatchCompoundTokenCompile(p *Parser, t *Task, rp *RunProvider, cs *CompilerState, val any) error {
	cp, ok := asCompoundPath(val)
	if !ok {
		return fmt.Errorf("compound token not supported in session dispatch")
	}
	if cp.stackSelector >= 0 {
		segs := compoundPathSegments(cp.path)
		if len(segs) == 0 {
			return fmt.Errorf("stack selector requires an operand or compound path")
		}
		if len(segs) == 1 {
			return compileSimplePrefixedOp(p, t, rp, cs, cp, segs[0])
		}
	}
	if cp.path == "" {
		return fmt.Errorf("compound token not supported in session dispatch")
	}
	bootstrapEmitCompoundLiteral(cs, compoundFrozenPathLiteral(cp))
	return nil
}

// walkCompoundSegments expands one compound path left-to-right (in-hand interpret engine).
// Uses a bare RunFrame (no ForceOp); Final methods detect that in mintFinalMethodCall and
// nest via runInterruptNest. Native member handlers run in-process (SR panics propagate).
func walkCompoundSegments(p *Parser, t *Task, rp *RunProvider, cp *compoundPathValue) error {
	if cp.stackSelector >= 0 {
		segs := compoundPathSegments(cp.path)
		if len(segs) == 0 {
			return fmt.Errorf("stack selector requires an operand or compound path")
		}
		if len(segs) == 1 {
			return dispatchSimplePrefixedOpInterpret(p, t, rp, cp.stackSelector, segs[0])
		}
	}
	segs := compoundPathSegments(cp.path)
	sel, rest := resolveCompoundWalkSelector(segs)
	if len(segs) > 0 {
		if _, ok := segmentToStackSelector(segs[0]); !ok && cp.stackSelector >= 0 {
			sel = cp.stackSelector
			rest = segs
		}
	}
	return walkCompoundPath(&compoundWalker{f: &RunFrame{Task: t, RP: rp}, sel: sel}, rest)
}

func executeInterpretWord(p *Parser, t *Task, rp *RunProvider, w *Word, xt XT) error {
	_ = p
	_ = xt
	return runExecWord(t, rp, w)
}

func executeInterpretWordBody(p *Parser, t *Task, rp *RunProvider, w *Word, xt XT) error {
	return executeInterpretWord(p, t, rp, w, xt)
}
