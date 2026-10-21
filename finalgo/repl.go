package finalgo

import "github.com/androticus/final/internal/core"

// REPLModule selects which dictionary backs the interactive terminal REPL.
type REPLModule int

const (
	// REPLModuleLocal uses a fresh per-run local module (cmd/final-go default).
	REPLModuleLocal REPLModule = iota
	// REPLModuleCore uses the sealed core module (cmd/final-go -core).
	REPLModuleCore
)

// NewREPLSession returns a persistent session for the given REPL module dictionary
// (stacks survive across batches — same as finalembed.NewPersistentSession).
// restrict is the initial Task.restrict mask (0 = unrestricted); typically a remapped
// uint32 SR bitmask from the final-go -restrict flag (prefer finalgo.SR_* named consts).
func NewREPLSession(module REPLModule, restrict int64) (*Session, error) {
	kind := core.SessionKindLocal
	if module == REPLModuleCore {
		kind = core.SessionKindCore
	}
	return core.NewApiSession(&core.SessionOptions{
		Kind:     kind,
		Persist:  true,
		Restrict: core.SecurityRestrictions(restrict),
	})
}

// RunInteractiveREPL creates a session for module and runs its REPL until stdin EOF.
// Bootstrap is compiled at package init via finalgo.init.
func RunInteractiveREPL(module REPLModule) error {
	sess, err := NewREPLSession(module, 0)
	if err != nil {
		return err
	}
	defer sess.Close()
	return sess.RunREPL()
}
