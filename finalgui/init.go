package finalgui

import (
	"fmt"
	"path/filepath"

	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/internal/core"

	_ "github.com/androticus/final/internal/fyneui" // App/Window/Label/… on the global catalog before boot
)

const defaultSendTimeoutS = 30

const packageFinalManifest = "finalgui/final/package.fnlm"

const guiRootWord = "guiRoot"

// init pins fyneui (blank import above) then resolves built-in type metadata.
// Do not InitGlobalExecutionEnvironment here: optional pins (fyneui tranche3, finalpi
// tranche4) must finish package init before Seal. NewSession boots core afterward.
func init() {
	core.ResolveBuiltInTypeRegistrations()
}

// Version returns the library version string (cmd/final-gui; avoids importing finalgo).
func Version() string {
	return core.Version()
}

// Init registers FinalGui and compiles package Final (guiRoot trampoline).
// Does not start Fyne or EventLoop.
func Init(sess *finalembed.Session) error {
	if sess == nil {
		return fmt.Errorf("finalgui.Init: nil session")
	}
	FinalGui_.sess = sess
	if err := sess.RegisterTypes(TypeRegs()); err != nil {
		return err
	}
	boot := filepath.ToSlash(packageFinalManifest)
	script := fmt.Sprintf("`%s` compileFromManifest", boot)
	if _, err := sess.GetSender().Send(script, defaultSendTimeoutS); err != nil {
		return fmt.Errorf("finalgui.Init: compile %s: %w", boot, err)
	}
	return nil
}

// TypeRegs returns host type registrations for this package (call from Init).
func TypeRegs() []finalembed.TypeReg {
	return []finalembed.TypeReg{
		finalembed.NewTypeReg[*FinalGui]("FinalGui", registerFinalGui),
	}
}
