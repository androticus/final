package core

import "fmt"

// FindName resolves a bare name in global GCD (REPL v1). Returns word and XT, or nil if missing.
func FindName(name String) (*Word, XT, bool) {
	if globalDictionaryContext == nil {
		panic("FindName: global execution environment not initialized")
	}
	return findNameInContext(globalDictionaryContext, name)
}

// FindNameInRun resolves a dictionary word via rp.FindScope (L0 word-only); falls back to global GCD when unset.
// Prefer FindNameInTask when a Task is available (SR enforcement).
func FindNameInRun(rp *RunProvider, name String) (*Word, XT, bool) {
	if rp == nil {
		panic("FindNameInRun: nil RunProvider")
	}
	if rp.FindScope != nil {
		return rp.FindScope.FindWordWithXT(string(name))
	}
	return FindName(name)
}

// findNameInContext resolves a bare name in dc (no SR check — use FindNameInTask for enforcement).
func findNameInContext(dc *DictionaryContext, name String) (*Word, XT, bool) {
	if dc == nil {
		panic("findNameInContext: nil DictionaryContext")
	}
	w := dc.dict.GetByName(string(name))
	if w == nil {
		return nil, 0, false
	}
	xt, ok := dc.dict.XTOf(w)
	if !ok {
		return nil, 0, false
	}
	return w, xt, true
}

// FindNameInTask — see module_search.go (CSO walk + prefix\tail).

// runGetNextToken reads the next token via Parser into DS (FinalGetNextToken / TrNextToken handler).
// Parser errors panic; no replError recovery (downstream steps assume a real token or stream EOF).
func runGetNextToken(t *Task, rp *RunProvider) {
	if err := pushNextToken(rp.Parser, rp.DataStack); err != nil {
		panic(fmt.Errorf("getNextToken: %w", err))
	}
}

// nestParser installs p as the run-scoped lexer for fn, then restores the previous parser.
// The saved parser is a Go local (not TS). Does not trim TS; leftover cells are the nested fn's.
// On error, formats a located fault from p before restore, and copies the fault site onto saved.
func nestParser(rp *RunProvider, p *Parser, fn func() error) (err error) {
	saved := rp.Parser
	rp.SetParser(p)
	defer func() {
		if err != nil && p != nil {
			err = locateFault(err, p)
			if saved != nil {
				saved.file = p.file
				saved.line = p.line
				saved.col = p.col
				saved.lineText = p.lineText
			}
		}
		rp.SetParser(saved)
	}()
	err = fn()
	return err
}

// drainParserToEOF abandons the remainder of the current compile source after rollback.
func drainParserToEOF(p *Parser) {
	if p == nil {
		return
	}
	p.drainToStreamEOF()
}

// runInterpretLoop is the Go-hosted REPL driver: one session parser on stdin until stream EOF.
// Errors are reported and the loop continues after draining to TokenNewline (no line ack).
func runInterpretLoop(t *Task, rp *RunProvider) {
	if rp == nil || rp.Stdin == nil {
		panic("runInterpretLoop: nil RunProvider or Stdin")
	}
	p := NewParserFromTextRead(rp.Stdin)
	rp.SetParser(p)
	rp.WithTaskFindScope(t)
	for {
		if CheckResetCoreByeReq() {
			break
		}
		if err := pushNextToken(rp.Parser, rp.DataStack); err != nil {
			bootstrapWriteREPLError(rp, err)
			skipParserLine(rp.Parser)
			continue
		}
		if PopStreamEOF(rp.DataStack) {
			break
		}
		// Panics from interpret (e.g. AssertSR on Find during compound walk) happen
		// outside runExecWord's doneChan recover — capture here like RunFault returns.
		err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = captureRunFault(t, rp, r, rp.LastOp, rp.LastIP)
				}
			}()
			return dispatchSessionToken(rp.Parser, t, rp)
		}()
		if err != nil {
			if fault, ok := err.(*RunFault); ok {
				warmBootReset(t, rp)
				if IsLocatedFaultMessage(fault.Message) {
					bootstrapWriteREPLError(rp, &locatedError{msg: fault.Message})
				} else {
					WriteRunFaultReport(rp, fault)
				}
			} else {
				bootstrapWriteREPLError(rp, err)
			}
			if t.dc != nil {
				SetCompileState(t.compilerForTarget(), false)
			}
			skipParserLine(rp.Parser)
		}
	}
}

// InterpretSource interprets src on the current task (no Session.Send).
// Stdout during the batch is captured and returned. Faults and interpret errors
// are returned; the caller prints them. Does not nest EventLoop#run.
func InterpretSource(t *Task, rp *RunProvider, src string) (stdout string, err error) {
	if t == nil || rp == nil {
		return "", fmt.Errorf("InterpretSource: nil Task or RunProvider")
	}
	tr, err := NewTextReaderFromString(src)
	if err != nil {
		return "", err
	}
	p := NewParserFromTextRead(tr)
	savedOut := rp.Stdout
	sb := NewStringBuilder()
	rp.Stdout = sb
	defer func() { rp.Stdout = savedOut }()
	err = nestParser(rp, p, func() error {
		return interpretSession(t, rp)
	})
	return sb.ToString(), err
}

// interpretSession runs the core token loop until stream EOF (errors propagate).
// rp.Parser must be bound before call.
func interpretSession(t *Task, rp *RunProvider) error {
	if rp == nil || rp.Parser == nil {
		panic("interpretSession: nil RunProvider or Parser")
	}
	rp.WithTaskFindScope(t)
	rs := rp.ReturnStack
	hadColon := rs.Len() > 0
	if !hadColon {
		// Ephemeral floor so interpret-time ( { [ can park ContextMark on the frame.
		ensureInnermostColonFrameMark(rs)
	}
	for {
		if err := pushNextToken(rp.Parser, rp.DataStack); err != nil {
			return err
		}
		if PopStreamEOF(rp.DataStack) {
			break
		}
		if err := dispatchSessionToken(rp.Parser, t, rp); err != nil {
			return err
		}
	}
	if t.dc != nil && GetCompileState(t.compilerForTarget()) {
		return fmt.Errorf("unfinished colon definition")
	}
	if !hadColon {
		popInterpretFloorIfAlone(rs)
	}
	return nil
}

// popInterpretFloorIfAlone removes a lone BootStop floor ColonFrameMark left by ensureInnermost.
func popInterpretFloorIfAlone(rs *ReturnStackType) {
	if rs == nil || rs.Len() != 1 {
		return
	}
	cm := rs.Peek()
	if cm.lastIP != int(ReservedCTBootStop) {
		return
	}
	rs.Toss()
}

// skipParserLine abandons the rest of the current input line after an interpret error.
func skipParserLine(p *Parser) {
	if p == nil {
		return
	}
	p.drainToEndOfLine()
}
