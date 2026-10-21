package core

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

// ApiSession and ApiLockSession get exported in finalembed / finalgo as Session and LockSession.

// Sender is the interface that wraps the Send method.
// It includes methods to Set Send options, and Send the message.
// It is provided by ApiSession.GetSender and ApiLockSession.GetLockedSender.
type Sender interface {
	SetSecurityRestrictions(restrictions SecurityRestrictions)
	Send(script string, timeoutS int) (string, error)
}

// SessionOptionsSetter is the session-level options surface (floor Restrict today).
type SessionOptionsSetter interface {
	SetSecurityRestrictions(restrictions SecurityRestrictions)
	InitConfig(jsonString string) error
	InitLogProvider(service, instanceId string, level int, provider LogProvider)
}

// ErrLockedSenderInvalid is returned when Send or Unlock is used after Unlock,
// or on a nil locked sender.
var ErrLockedSenderInvalid = fmt.Errorf("session: locked sender is invalid")

// ApiSessionBase is the base type for all API sessions.
// It contains the Session, and Session Option setters.
type ApiSessionBase struct {
	sess *Session
}

// RegisterTypes installs host types into this session's module DC.
// Call once after create and before Send. Same Go type registered on a second
// Session panics in the process-global host-type map (module share is later).
func (ss *ApiSessionBase) RegisterTypes(regs []TypeReg) error {
	if ss == nil || ss.sess == nil {
		return fmt.Errorf("RegisterTypes: nil Session")
	}
	return ss.sess.RegisterTypes(regs)
}

// Break requests cooperative stop of an in-flight VM run.
func (ss *ApiSessionBase) Break() {
	if ss == nil || ss.sess == nil {
		return
	}
	ss.sess.Break()
}

// Close stops the session worker and releases resources. Idempotent.
func (ss *ApiSessionBase) Close() error {
	if ss == nil || ss.sess == nil {
		return nil
	}
	return ss.sess.Close()
}

// Initializer Methods, all Sessions support these.

func (ss *ApiSessionBase) SetSecurityRestrictions(restrictions SecurityRestrictions) {
	if ss == nil || ss.sess == nil {
		return
	}
	ss.sess.SetSecurityRestrictions(restrictions)
}

func (ss *ApiSessionBase) InitConfig(jsonString string) error {
	if ss == nil || ss.sess == nil {
		return fmt.Errorf("InitConfig: nil Session")
	}
	return ss.sess.InitConfig(jsonString)
}

func (ss *ApiSessionBase) InitLogProvider(service, instanceId string, level int, provider LogProvider) {
	if ss == nil || ss.sess == nil {
		return
	}
	ss.sess.InitLogProvider(service, instanceId, Level(level), provider)
}

// SetStdout installs the standing stdout writer on a persistent Session.
// Nil w restores the null writer.
func (ss *ApiSessionBase) SetStdout(w TextWrite) {
	if ss == nil || ss.sess == nil {
		return
	}
	ss.sess.SetStdout(w)
}

// SetStderr installs the standing stderr writer on a persistent Session.
// Nil w restores the null writer. Stderr is for user $$ only.
func (ss *ApiSessionBase) SetStderr(w TextWrite) {
	if ss == nil || ss.sess == nil {
		return
	}
	ss.sess.SetStderr(w)
}

// end Initializer Methods

// RunREPL runs the interactive terminal REPL until stdin EOF or bye.
// Used by finalgo.NewREPLSession / cmd/final-go; not part of the embed Send path.
func (ss *ApiSessionBase) RunREPL() error {
	if ss == nil || ss.sess == nil {
		return fmt.Errorf("RunREPL: nil Session")
	}
	return ss.sess.RunREPL()
}

// RunEventLoop runs EventLoop#run until the event pipe closes (GUI/TUI worker root).
// rootWord is the trampoline colon; empty uses "guiRoot".
func (ss *ApiSessionBase) RunEventLoop(rootWord string) error {
	if ss == nil || ss.sess == nil {
		return fmt.Errorf("RunEventLoop: nil Session")
	}
	return ss.sess.RunEventLoop(rootWord)
}

// InterpretReader interprets Final source from r on process stdout until EOF.
// No REPL banner or (ok). Used by product script / stdin launch.
func (ss *ApiSessionBase) InterpretReader(r io.Reader) error {
	if ss == nil || ss.sess == nil {
		return fmt.Errorf("InterpretReader: nil Session")
	}
	return ss.sess.InterpretReader(r)
}

// InterpretReaderPath is InterpretReader with an optional source path for fault reports.
func (ss *ApiSessionBase) InterpretReaderPath(r io.Reader, path string) error {
	if ss == nil || ss.sess == nil {
		return fmt.Errorf("InterpretReaderPath: nil Session")
	}
	return ss.sess.InterpretReaderPath(r, path)
}

func newApiSessionBase(opts *SessionOptions) (*ApiSessionBase, error) {
	s, err := NewSession(opts)
	if err != nil {
		return nil, err
	}
	return &ApiSessionBase{sess: s}, nil
}

// ApiSession is the host-facing one-shot Session: GetSender for each dispatch.
type ApiSession struct {
	*ApiSessionBase
}

// NewApiSession returns an API Session for opts (Kind, Persist, Restrict).
func NewApiSession(opts *SessionOptions) (*ApiSession, error) {
	base, err := newApiSessionBase(opts)
	if err != nil {
		return nil, err
	}
	return &ApiSession{ApiSessionBase: base}, nil
}

// ApiSessionSender is a fresh Send handle from ApiSession.GetSender.
type ApiSessionSender struct {
	*ApiSession
	restrict SecurityRestrictions
}

func (ss *ApiSession) GetSender() *ApiSessionSender {
	if ss == nil {
		return nil
	}
	return &ApiSessionSender{ApiSession: ss}
}

func (ss *ApiSessionSender) SetSecurityRestrictions(restrictions SecurityRestrictions) {
	if ss == nil {
		return
	}
	ss.restrict = restrictions
}

func (ss *ApiSessionSender) Send(script string, timeoutS int) (string, error) {
	if ss == nil || ss.ApiSession == nil || ss.sess == nil {
		return "", fmt.Errorf("Send: nil Session")
	}
	return ss.sess.send(script, timeoutS, ss.restrict)
}

// ApiLockSession is the host-facing exclusive Session: GetLockedSender holds
// until Unlock on that sender.
type ApiLockSession struct {
	mu sync.Mutex
	*ApiSessionBase
}

// NewApiLockSession returns an API lockable Session for opts.
func NewApiLockSession(opts *SessionOptions) (*ApiLockSession, error) {
	base, err := newApiSessionBase(opts)
	if err != nil {
		return nil, err
	}
	return &ApiLockSession{ApiSessionBase: base}, nil
}

// GetLockedSender acquires the exclusive hold and returns a sticky sender.
// A second call blocks until the current sender Unlock()s.
func (ls *ApiLockSession) GetLockedSender() *ApiLockSessionSender {
	if ls == nil {
		return nil
	}
	ls.mu.Lock()
	return &ApiLockSessionSender{ApiLockSession: ls}
}

// ApiLockSessionSender is the parked Send handle for a lock hold.
type ApiLockSessionSender struct {
	*ApiLockSession
	restrict SecurityRestrictions
	unlocked atomic.Bool
}

func (ls *ApiLockSessionSender) SetSecurityRestrictions(restrictions SecurityRestrictions) {
	if ls == nil || ls.unlocked.Load() || ls.ApiLockSession == nil {
		return
	}
	ls.restrict = restrictions
}

func (ls *ApiLockSessionSender) Send(script string, timeoutS int) (string, error) {
	if ls == nil || ls.unlocked.Load() || ls.ApiLockSession == nil || ls.sess == nil {
		return "", ErrLockedSenderInvalid
	}
	return ls.sess.send(script, timeoutS, ls.restrict)
}

// Unlock releases the LockSession so GetLockedSender may proceed, and invalidates
// this sender. It is an error to call again after unlocking.
func (ls *ApiLockSessionSender) Unlock() error {
	if ls == nil || !ls.unlocked.CompareAndSwap(false, true) {
		return ErrLockedSenderInvalid
	}
	parent := ls.ApiLockSession
	ls.ApiLockSession = nil
	if parent != nil {
		parent.mu.Unlock()
	}
	return nil
}

var (
	_ Sender               = (*ApiSessionSender)(nil)
	_ Sender               = (*ApiLockSessionSender)(nil)
	_ SessionOptionsSetter = (*ApiSessionBase)(nil)
	_ SessionOptionsSetter = (*ApiSession)(nil)
	_ SessionOptionsSetter = (*ApiLockSession)(nil)
)
