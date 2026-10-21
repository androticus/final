package core

import "fmt"

// Session dispatch helpers shared by the Go interpret loop and OpTransition handlers.

// dispatchSessionParsed routes one parsed token through the reentrant handler.
func dispatchSessionParsed(p *Parser, t *Task, rp *RunProvider, val any, typ TokenType) error {
	return handleParsedToken(p, t, rp, val, typ, false, nil)
}

func dispatchSessionParsedInRun(p *Parser, t *Task, rp *RunProvider, val any, typ TokenType, rf *RunFrame) error {
	return handleParsedToken(p, t, rp, val, typ, true, rf)
}

// dispatchInterpretParsed is a test/helper entry to the interpret path only.
func dispatchInterpretParsed(p *Parser, t *Task, rp *RunProvider, val any, typ TokenType) error {
	cs := t.compilerForTarget()
	if GetCompileState(cs) {
		return fmt.Errorf("dispatchInterpretParsed: compile state active")
	}
	return handleParsedTokenInterpret(p, t, rp, val, typ, false)
}

func popTokenType(ds *Stack[any]) (TokenType, error) {
	raw := ds.Pop()
	v, ok := raw.(Int)
	if !ok {
		return 0, fmt.Errorf("session: expected TokenType Int, got %T", raw)
	}
	return TokenType(v), nil
}

func dispatchSessionToken(p *Parser, t *Task, rp *RunProvider) error {
	ds := rp.DataStack
	if ds.Len() < 2 {
		return fmt.Errorf("session: expected value and TokenType on stack")
	}
	typ, err := popTokenType(ds)
	if err != nil {
		return err
	}
	val := ds.Pop()
	return dispatchSessionParsed(p, t, rp, val, typ)
}

// runWordInlineExec runs dictionary word w via execWord (OpExecWord).
func runWordInlineExec(t *Task, rp *RunProvider, w *Word) error {
	if w == nil {
		return fmt.Errorf("runWordInlineExec: nil word")
	}
	return runExecWord(t, rp, w)
}

// execInterpretWord runs w from token dispatch via execWord.
func execInterpretWord(t *Task, rp *RunProvider, w *Word, xt XT) error {
	_ = xt
	return runExecWord(t, rp, w)
}
