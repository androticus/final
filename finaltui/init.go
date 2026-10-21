package finaltui

import (
	"fmt"
	"path/filepath"

	"github.com/androticus/final/finalembed"
)

const defaultSendTimeoutS = 30

// Package Final sources are listed in package.fnlm (same comment/blank rules as
// other .fnlm files). Paths inside that manifest are relative to its directory.
// Resolve via cmd-root / bootstrap root like compileFromManifest elsewhere.
const packageFinalManifest = "finaltui/final/package.fnlm"

// TuiRootWord is the EventLoop trampoline compiled by Init.
// Headed TUI RunTUI sits in RunEventLoop(TuiRootWord).
const TuiRootWord = "tuiRoot"

// Init registers Tui with the session, mints Final var Tui_, and compiles package Final sources.
// Does not start any UI — cmd owns chrome / RunHost.
func Init(sess *finalembed.Session) error {
	if sess == nil {
		return fmt.Errorf("finaltui.Init: nil session")
	}
	if err := sess.RegisterTypes(TypeRegs()); err != nil {
		return err
	}
	// Bind singleton into the session dictionary as Tui_.
	if _, err := sess.GetSender().Send(`Tui.instance^ var: Tui_`, defaultSendTimeoutS); err != nil {
		return fmt.Errorf("finaltui.Init: mint Tui_: %w", err)
	}
	boot := filepath.ToSlash(packageFinalManifest)
	script := fmt.Sprintf("`%s` compileFromManifest", boot)
	if _, err := sess.GetSender().Send(script, defaultSendTimeoutS); err != nil {
		return fmt.Errorf("finaltui.Init: compile %s: %w", boot, err)
	}
	return nil
}
