// Programmatic embedder API for running Final scripts without the terminal REPL.
//
// Quick start:
//
//	session, err := core.NewSession(&core.SessionOptions{Kind: core.SessionKindLocal})
//	if err != nil {
//		panic(err)
//	}
//	defer session.Close()
//
//	out, err := session.Send("1 2 + .s", 0)
//	if err != nil {
//		// handle fault or interpret error
//	}
//	fmt.Println(out)
//
// Legacy RunScript* APIs remain for tests; embedders should use Send.
package core

import (
	"fmt"
	"io"
	"path/filepath"
	"sync"
)

// --- Candy API ---

// RunScript runs source with stdout discarded (status / fault only).
//
//	res := session.RunScript("1 2 +")
func (s *Session) RunScript(script string) RunResult {
	h, err := s.Run(RunInputString{Script: script}, nil)
	if err != nil {
		return interpretError(err)
	}
	return h.Wait()
}

// RunScriptWithOutput runs source and returns emitted stdout in res.Output.
//
//	res := session.RunScriptWithOutput(`1 2 .s`)
func (s *Session) RunScriptWithOutput(script string) RunResult {
	h, err := s.Run(RunInputString{Script: script}, RunOutputCapture{})
	if err != nil {
		return interpretError(err)
	}
	return h.Wait()
}

// --- Session ---

// SessionKind selects which module backs a Session or interactive REPL task.
type SessionKind int

const (
	// SessionKindCore uses the sealed core module (ModuleTable slot 1).
	SessionKindCore SessionKind = iota
	// SessionKindLocal registers a fresh per-session module (own DC).
	SessionKindLocal
)

// SessionOptions configures NewSession.
type SessionOptions struct {
	Kind     SessionKind
	Persist  bool                 // reuse one RunProvider across Send/Run (REPL-like)
	Restrict SecurityRestrictions // session floor; copied to Session.restrict (0 = unrestricted)
}

// Session owns one global Task and supports many Run invocations.
type Session struct {
	task       *Task
	statsID    uint64
	mu         sync.Mutex
	closed     bool
	persistent bool
	rp         *RunProvider // non-nil when persistent; reused across runOnce

	// modules is the default search order (ModuleReferences). For Local Sessions this is the
	// same list as the Session-local Module.requires ([local, core, …imports]).
	modules *ModuleReferenceList

	// restrict is the Session floor (disabled-set). Send ORs the sender extra onto
	// Task.restrict for that batch; already-running pips are not updated.
	restrict SecurityRestrictions

	sendMu     sync.Mutex
	workerReq  chan batchReq
	workerRes  chan batchRes
	workerDone chan struct{}

	// configJson and config are used to initialize Session.config from a JSON string.
	// this 2-step processis needed because JSON parsing is all done in Final code
	configJson string // JSON string for Session.config
	config     *DynamicObject

	// logger is the Session logger; provider starts as nil, which ignores all logs.
	// embbedders can set a provider via SetLogProvider.
	// Final code can set a provider via Session.setLogProvider
	logger *Logger
}

// NewSession returns a session backed by a Task for opts.Kind.
//
// SessionKindLocal is the preferred default for embedders; SessionKindCore uses the sealed core module.
// If Persist is true, one RunProvider (and its stacks) is reused across Send/Run batches
// (REPL-like). If false, each batch gets a fresh RunProvider (request/batch semantics).
// Restrict is the session floor (disabled-set). Copied to Session.restrict and the
// initial Task.restrict. Each Send writes Task.restrict = Session.restrict | extra.
// Use Send(script, timeoutS) for programmatic batches — timeoutS is an advisory run budget (seconds;
// 0 = none) passed to the task for inward I/O and native cooperative checks, not a hard wall-clock
// guarantee. Use Break() to cooperatively stop an in-flight VM run (REPL SIGINT, host cancel).
func NewSession(opts *SessionOptions) (*Session, error) {
	if opts == nil {
		return nil, fmt.Errorf("NewSession: nil SessionOptions")
	}
	// Embed-only hosts (finalembed without finalgo/finalgui) still need a boot.
	// Product binaries boot in package init; this Once is a no-op for them.
	ResolveBuiltInTypeRegistrations()
	InitGlobalExecutionEnvironment()
	if globalDictionaryContext == nil {
		return nil, fmt.Errorf("NewSession: runtime not initialized")
	}
	task, err := newSessionTask(opts.Kind)
	if err != nil {
		return nil, err
	}
	task.name = "session"
	s := &Session{
		task:       task,
		statsID:    allocSessionStatsID(),
		persistent: opts.Persist,
		modules:    task.currentSearchOrder,
		restrict:   opts.Restrict,
	}
	s.config = NewDynamicObject() // replaced by InitConfig call, else empty

	// no defaults; initialized eithervia InitLogger during Session creation.
	// or via Logger.Init during Final-basedSession initialization.
	s.logger = NewLogger()
	task.restrict = opts.Restrict
	task.session = s
	if opts.Persist {
		rp := NewRunProvider()
		rp.Stderr = GetNullTextWriter()
		s.rp = rp
	}
	startSessionWorker(s)
	return s, nil
}

func newSessionTask(kind SessionKind) (*Task, error) {
	switch kind {
	case SessionKindCore:
		// Dev: compile into core (final-go -core). Production default is SessionKindLocal.
		if globalDictionaryContext != nil {
			globalDictionaryContext.Unseal()
		}
		return CreateCoreTask(), nil
	case SessionKindLocal:
		return CreateLocalTask(), nil
	default:
		return nil, fmt.Errorf("NewSession: unknown SessionKind %d", kind)
	}
}

// Session Options Setters

func (s *Session) SetSecurityRestrictions(restrictions SecurityRestrictions) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.restrict = restrictions
	s.mu.Unlock()
}

// Initialize the Session.config from a JSON string.
// This call invokes a Send on the VM, in order to execute Final code.
// This happens "below the radar" of Embedder code and locks.
func (s *Session) InitConfig(jsonString string) error {
	if s == nil {
		return fmt.Errorf("InitConfig: nil Session")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configJson = jsonString
	cmd := "sess.initConfig^"
	_, err := s.Send(cmd, 0)
	return err
}

func (s *Session) InitLogProvider(service, instanceId string, level Level, provider LogProvider) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.logger.Init(service, instanceId, level, provider)
	s.mu.Unlock()
}

// SetLogProvider sets the LogProvider for the Session.
// Can be nil to ignore all logs.
func (s *Session) SetLogProvider(provider LogProvider) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.logger.SetLogProvider(provider)
	s.mu.Unlock()
}

// SetStdout installs the standing stdout writer for a persistent Session.
// Nil w restores the null writer. No-op when the Session is not persistent.
func (s *Session) SetStdout(w TextWrite) {
	if s == nil || s.rp == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if w == nil {
		s.rp.Stdout = GetNullTextWriter()
		return
	}
	s.rp.Stdout = w
}

// SetStderr installs the standing stderr writer for a persistent Session.
// Nil w restores the null writer. No-op when the Session is not persistent.
// Stderr is for user $$ only; system fault reports use stdout.
func (s *Session) SetStderr(w TextWrite) {
	if s == nil || s.rp == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if w == nil {
		s.rp.Stderr = GetNullTextWriter()
		return
	}
	s.rp.Stderr = w
}

// Close marks the session closed and removes its stats entry. Idempotent.
// Blocks until any in-flight Send completes, then stops the VM worker.
func (s *Session) Close() error {
	if s == nil {
		return nil
	}
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	unregisterSessionStats(s.statsID)
	s.mu.Unlock()
	close(s.workerReq)
	<-s.workerDone
	return nil
}

// StatsID returns the metrics registry id (0 if none).
func (s *Session) StatsID() uint64 {
	if s == nil {
		return 0
	}
	return s.statsID
}

// --- RunHandle ---

// RunHandle is completion-only; pipe client ends come from NewRunInputPipe / NewRunOutputPipe.
type RunHandle struct {
	done chan RunResult
}

// Wait blocks until the run finishes and returns the result.
func (h RunHandle) Wait() RunResult {
	if h.done == nil {
		return RunResult{Status: RunCompleted}
	}
	return <-h.done
}

// --- Flexible Run I/O types ---

// RunInput selects program source for Run. nil = graceful no-op (immediate EOF).
type RunInput interface {
	runInput()
}

// RunOutput selects stdout handling. nil = discard stdout.
type RunOutput interface {
	runOutput()
}

type RunInputString struct{ Script string }
// RunInputReader reads Final source from Source. Path is optional; when set,
// the parser file location is that path (absolute when possible) for fault reports.
type RunInputReader struct {
	Source io.Reader
	Path   string
}
type RunInputPipe struct{ vmReader io.ReadCloser }

type RunOutputCapture struct{}
type RunOutputWriter struct{ Sink io.Writer }
type RunOutputPipe struct{ vmWriter io.WriteCloser }

func (RunInputString) runInput()    {}
func (RunInputReader) runInput()    {}
func (RunInputPipe) runInput()      {}
func (RunOutputCapture) runOutput() {}
func (RunOutputWriter) runOutput()  {}
func (RunOutputPipe) runOutput()    {}

// NewRunInputPipe creates a paired stdin pipe. Returns the RunInput param and client WriteCloser.
func NewRunInputPipe() (RunInputPipe, io.WriteCloser) {
	vmR, clientW := io.Pipe()
	return RunInputPipe{vmReader: vmR}, clientW
}

// NewRunOutputPipe creates a paired stdout pipe. Returns the RunOutput param and client ReadCloser.
func NewRunOutputPipe() (RunOutputPipe, io.ReadCloser) {
	clientR, vmW := io.Pipe()
	return RunOutputPipe{vmWriter: vmW}, clientR
}

// Run executes one interpret session with the given input and output params.
func (s *Session) Run(in RunInput, out RunOutput) (RunHandle, error) {
	if s == nil {
		return RunHandle{}, fmt.Errorf("session: nil Session")
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return RunHandle{}, ErrSessionClosed
	}
	s.mu.Unlock()

	if err := validateRunInput(in); err != nil {
		return RunHandle{}, err
	}
	if err := validateRunOutput(out); err != nil {
		return RunHandle{}, err
	}

	done := make(chan RunResult, 1)
	if runNeedsAsync(in, out) {
		go func() {
			done <- s.runOnce(in, out)
		}()
		return RunHandle{done: done}, nil
	}
	done <- s.runOnce(in, out)
	return RunHandle{done: done}, nil
}

func (s *Session) runOnce(in RunInput, out RunOutput) RunResult {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return RunResult{Status: RunInterpretError, Err: ErrSessionClosed}
	}
	s.mu.Unlock()

	recordSessionRun(s.statsID)

	rp, sb, cleanup, err := s.bindRunIO(in, out)
	if err != nil {
		return interpretError(err)
	}
	defer cleanup()

	var result RunResult
	func() {
		defer func() {
			if r := recover(); r != nil {
				result = faultResult(s.task, rp, r)
			}
		}()
		if rp.Stdin == nil {
			result = RunResult{Status: RunCompleted}
			return
		}
		p := NewParserFromTextRead(rp.Stdin)
		if path := runInputFilePath(in); path != "" {
			abs := path
			if a, err := filepath.Abs(path); err == nil {
				abs = a
			}
			p.setParserLocation(abs, 1, 1)
		}
		rp.SetParser(p)
		if err := interpretSession(s.task, rp); err != nil {
			result = interpretError(locateFault(err, rp.Parser))
			return
		}
		result = RunResult{Status: RunCompleted}
	}()
	if sb != nil {
		result.Output = sb.ToString()
	}
	return result
}

func runInputFilePath(in RunInput) string {
	if v, ok := in.(RunInputReader); ok {
		return v.Path
	}
	return ""
}

type runIOCleanup func()

// bindRunIO wires stdin/stdout onto a RunProvider. Ephemeral sessions get a fresh
// provider; persistent sessions reuse s.rp (stacks survive across batches).
func (s *Session) bindRunIO(in RunInput, out RunOutput) (*RunProvider, *StringBuilder, runIOCleanup, error) {
	var rp *RunProvider
	if s.persistent {
		rp = s.rp
	} else {
		rp = NewRunProvider()
		rp.Stderr = GetNullTextWriter()
	}

	var sb *StringBuilder
	var closers []io.Closer
	prevStdout := rp.Stdout
	prevStderr := rp.Stderr

	cleanup := func() {
		rp.Stdout = prevStdout
		rp.Stderr = prevStderr
		for _, c := range closers {
			_ = c.Close()
		}
	}

	// stdin
	switch v := in.(type) {
	case nil:
		rp.Stdin = GetNullTextReader()
	case RunInputString:
		tr, err := NewTextReaderFromString(v.Script)
		if err != nil {
			return nil, nil, cleanup, err
		}
		rp.Stdin = tr
	case RunInputReader:
		tr, err := NewTextReaderFromReader(v.Source)
		if err != nil {
			return nil, nil, cleanup, err
		}
		rp.Stdin = tr
	case RunInputPipe:
		tr, err := NewTextReaderFromReader(v.vmReader)
		if err != nil {
			return nil, nil, cleanup, err
		}
		rp.Stdin = tr
		closers = append(closers, v.vmReader)
	default:
		panic(fmt.Errorf("bindRunIO: invalid RunInput %T (validateRunInput should have rejected)", in))
	}

	// stdout
	switch v := out.(type) {
	case nil:
		rp.Stdout = GetNullTextWriter()
	case RunOutputCapture:
		sb = NewStringBuilder()
		rp.Stdout = sb
	case RunOutputWriter:
		tw, err := NewTextWriterFromWriter(v.Sink)
		if err != nil {
			return nil, nil, cleanup, err
		}
		rp.Stdout = tw
	case RunOutputPipe:
		tw, err := NewTextWriterFromWriter(v.vmWriter)
		if err != nil {
			return nil, nil, cleanup, err
		}
		rp.Stdout = tw
		closers = append(closers, v.vmWriter)
	default:
		panic(fmt.Errorf("bindRunIO: invalid RunOutput %T (validateRunOutput should have rejected)", out))
	}

	return rp, sb, cleanup, nil
}
