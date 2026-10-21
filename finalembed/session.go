// Package finalembed is the embedding API for hosting the Final language inside a Go
// application with native host types.
//
// Unlike finalgo (a minimal Send/Close surface), finalembed additionally lets the host
// register its own Go types as Final types via RegisterTypes, so scripts can construct
// and drive them. Host types register on this Session's DC; register once before Send.
// Creating additional Sessions that RegisterTypes the same Go type is not supported yet
// (process-global host-type map). Shared module registration comes later.
//
// Final compound types (List, Map, StringBuilder, strong numeric lists, …) and value
// types for native handlers are re-exported from types.go.
//
//	sess, _ := finalembed.NewSession()
//	defer sess.Close()
//	sess.RegisterTypes([]finalembed.TypeReg{
//		finalembed.NewTypeReg[*Widget]("Widget", registerWidget),
//	})
//	out, _ := sess.GetSender().Send(`Widget.new^ >ts t:.render^`, 0)
package finalembed

import (
	"github.com/androticus/final/internal/core"
)

// init resolves core type metadata only. Optional catalog pins (fyneui tranche3,
// finalpi tranche4) register in their package init(); NewSession boots/seals core
// afterward so those pins always land first. Eager InitGlobal in product package
// init is avoided for that reason.
func init() {
	core.ResolveBuiltInTypeRegistrations()
}

// Session is an embedding session backed by a private module dictionary.
type Session = core.ApiSession

type LogProvider = core.LogProvider

// SessionSender is the handle returned by Session.GetSender.
type SessionSender = core.ApiSessionSender

// LockSession is a Session whose only Send path is GetLockedSender (exclusive hold).
type LockSession = core.ApiLockSession

// LockedSender is the parked handle from LockSession.GetLockedSender; Unlock invalidates it.
type LockedSender = core.ApiLockSessionSender

// Sender is the Send + options interface implemented by both sender handles.
type Sender = core.Sender

// SecurityRestrictions is the disabled-set mask (bits set = themes denied).
type SecurityRestrictions = core.SecurityRestrictions

const (
	SR_ModuleInstall  = core.SR_ModuleInstall
	SR_ModuleImport   = core.SR_ModuleImport
	SR_FileRead       = core.SR_FileRead
	SR_FileWrite      = core.SR_FileWrite
	SR_FileSystemRead = core.SR_FileSystemRead
	SR_HttpDial       = core.SR_HttpDial
	SR_HttpListen     = core.SR_HttpListen
	SR_Threads        = core.SR_Threads
	SR_Db             = core.SR_Db
	SR_Crypto         = core.SR_Crypto
	SR_Mask           = core.SR_Mask
)

type Level = int

const (
	LogLevelOff   = core.LogLevelOff
	LogLevelFatal = core.LogLevelFatal
	LogLevelError = core.LogLevelError
	LogLevelWarn  = core.LogLevelWarn
	LogLevelInfo  = core.LogLevelInfo
	LogLevelDebug = core.LogLevelDebug
	LogLevelTrace = core.LogLevelTrace
)

// NewSession returns a session backed by a fresh local module (the embedder default).
// Register host types with RegisterTypes before Send. Stacks are ephemeral (new each Send).
// Send timeoutS is an advisory run budget in seconds (0 = none) for inward I/O —
// see 134-embedding-final.md.
func NewSession() (*Session, error) {
	return core.NewApiSession(&core.SessionOptions{Kind: core.SessionKindLocal})
}

// NewPersistentSession is like NewSession but reuses one RunProvider across Sends
// so data (and other) stacks persist — REPL-like behavior for final.
func NewPersistentSession() (*Session, error) {
	return core.NewApiSession(&core.SessionOptions{Kind: core.SessionKindLocal, Persist: true})
}

// SetSysLaunch fills the process Sys argv/params snapshots (product CLI).
// argv is a copy of os.Args; params is the Python-style program vector (empty when
// interactive with no script).
func SetSysLaunch(argv, params []string) {
	core.SetSysLaunch(argv, params)
}

// NewLockSession returns a lockable local-module session (ephemeral stacks).
func NewLockSession() (*LockSession, error) {
	return core.NewApiLockSession(&core.SessionOptions{Kind: core.SessionKindLocal})
}

var (
	ErrSessionClosed       = core.ErrSessionClosed
	ErrEmptyScript         = core.ErrEmptyScript
	ErrLockedSenderInvalid = core.ErrLockedSenderInvalid
)
