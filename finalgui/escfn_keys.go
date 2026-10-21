package finalgui

import (
	"fyne.io/fyne/v2"
	"github.com/androticus/final/internal/core"
)

// escFnCodeFromFyneKey maps Esc / F1–F12 to latch codes (0 / 1–12).
func escFnCodeFromFyneKey(name fyne.KeyName) (int, bool) {
	switch name {
	case fyne.KeyEscape:
		return core.EscFnEsc, true
	case fyne.KeyF1:
		return 1, true
	case fyne.KeyF2:
		return 2, true
	case fyne.KeyF3:
		return 3, true
	case fyne.KeyF4:
		return 4, true
	case fyne.KeyF5:
		return 5, true
	case fyne.KeyF6:
		return 6, true
	case fyne.KeyF7:
		return 7, true
	case fyne.KeyF8:
		return 8, true
	case fyne.KeyF9:
		return 9, true
	case fyne.KeyF10:
		return 10, true
	case fyne.KeyF11:
		return 11, true
	case fyne.KeyF12:
		return 12, true
	default:
		return 0, false
	}
}

// latchEscFnKey latches Esc/Fn if key matches. Returns true when consumed.
func latchEscFnKey(key *fyne.KeyEvent) bool {
	if key == nil {
		return false
	}
	n, ok := escFnCodeFromFyneKey(key.Name)
	if !ok {
		return false
	}
	core.LatchEscFn(n)
	return true
}
