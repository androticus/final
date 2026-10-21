package finalgo

import (
	"github.com/androticus/final/internal/core"
)

// Sender runs one Final script batch synchronously and returns captured stdout.
// SetSecurityRestrictions is the per-sender options hook (floor vs extra merge later).
type Sender = core.Sender

// Session is a long-lived Final execution context backed by a private module dictionary.
// Send via GetSender().Send(script, timeoutS). timeoutS is an advisory run budget in
// seconds (0 = none) — not a hard guarantee the batch finishes in that time.
// Break() cooperatively stops an in-flight VM run (e.g. SIGINT during REPL).
type Session = core.ApiSession

// SessionSender is the handle returned by Session.GetSender.
type SessionSender = core.ApiSessionSender

// LockSession is a Session whose only Send path is GetLockedSender (exclusive hold).
type LockSession = core.ApiLockSession

// LockedSender is the parked handle from LockSession.GetLockedSender; Unlock invalidates it.
type LockedSender = core.ApiLockSessionSender

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

// NewSession returns a session backed by a fresh local module (the usual embedder default).
func NewSession() (*Session, error) {
	return core.NewApiSession(&core.SessionOptions{Kind: core.SessionKindLocal})
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
