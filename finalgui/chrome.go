package finalgui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
	"github.com/androticus/final/internal/fyneui"
)

// promptEntry is a multiline Entry: Enter sends, Shift-Enter inserts a newline.
// Stock MultiLine+OnSubmitted is the inverse (Enter newline, Shift-Enter submit).
type promptEntry struct {
	widget.Entry
	onSend func()
}

func newPromptEntry(onSend func()) *promptEntry {
	e := &promptEntry{onSend: onSend}
	e.MultiLine = true
	e.Wrapping = fyne.TextWrapWord
	e.SetMinRowsVisible(3)
	e.SetPlaceHolder("Final prompt — Enter send, Shift-Enter newline")
	e.ExtendBaseWidget(e)
	return e
}

func (e *promptEntry) TypedKey(key *fyne.KeyEvent) {
	if latchEscFnKey(key) {
		return
	}
	if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter {
		shift := false
		if d, ok := fyne.CurrentApp().Driver().(desktop.Driver); ok {
			shift = d.CurrentKeyModifiers()&fyne.KeyModifierShift != 0
		}
		if !shift {
			if e.onSend != nil {
				e.onSend()
			}
			return
		}
		// Shift-Enter: parent inserts newline (OnSubmitted stays nil).
	}
	e.Entry.TypedKey(key)
}

// BuildChrome constructs the REPL window (menu, upper dual pane, prompt, Send + idle/busy).
// Call on the OS main thread before TheApp.Run().
func BuildChrome() fyne.Window {
	applyStoredScaleEnv() // Prefs → FYNE_SCALE before first window (if env unset)

	g := FinalGui_

	g.response = newReadOnlyPane("Response")
	g.history = newReadOnlyPane("History")
	g.stdoutPane = newReadOnlyPane("$, Errors (Stdout)")
	g.stderrPane = newReadOnlyPane("$$ (Stderr)")

	sendFn := core.GetEventLoop().CreateEventFn(methodRef("onSend"), 0)
	submit := func() {
		if g.busy.Load() {
			return
		}
		src := strings.TrimRight(g.prompt.Text, "\r\n")
		if src == "" {
			return
		}
		g.prompt.SetText("")
		g.markBusy()
		sendFn(nil, core.String(src))
	}

	g.prompt = newPromptEntry(submit)
	g.send = widget.NewButton("Send", submit)
	g.busyStatus = canvas.NewText("idle", color.NRGBA{R: 0, G: 0, B: 0, A: 0xff})
	g.busyStatus.TextSize = theme.TextSize()
	g.busyStatus.Alignment = fyne.TextAlignCenter

	g.stdoutLog = core.NewLineBufferTextWriterSync(standingMaxLines)
	g.stderrLog = core.NewLineBufferTextWriterSync(standingMaxLines)

	th := newReadOnlyTheme()
	fyneui.TheApp.Settings().SetTheme(th)

	replSplit := container.NewVSplit(g.response, g.history)
	replSplit.SetOffset(0.72)
	g.replPanes = container.NewThemeOverride(replSplit, th)

	streamSplit := container.NewVSplit(g.stdoutPane, g.stderrPane)
	streamSplit.SetOffset(0.5)
	g.streamPanes = container.NewThemeOverride(streamSplit, th)

	g.chromeView = chromeViewREPL
	g.upperHost = container.NewMax(g.replPanes)

	sendCol := container.NewVBox(g.send, container.NewCenter(g.busyStatus))
	promptRow := container.NewBorder(nil, nil, nil, sendCol, g.prompt)
	body := container.NewBorder(
		nil,
		promptRow,
		nil,
		nil,
		g.upperHost,
	)

	w := fyneui.TheApp.NewWindow("Final")
	bar := buildMenuBar(w)
	w.SetContent(container.NewBorder(bar, nil, nil, nil, body))
	w.SetCloseIntercept(func() {}) // swallow — Quit / bye exit intentionally
	w.Resize(fyne.NewSize(defaultChromeWidth, defaultChromeHeight))
	g.win = w
	installEscFnKeyLatch(w)
	g.startStandingPoller()
	fyneui.TheApp.Lifecycle().SetOnStarted(onAppStarted)
	return w
}

// installEscFnKeyLatch wires Esc / F1–F12 → LatchEscFn via TypedKey.
// Fyne only synthesizes desktop.CustomShortcut when a modifier is held, so
// bare Esc/Fn never reach Canvas.AddShortcut — TypedKey is the reliable path
// (prompt, read-only panes, and canvas when nothing is focused).
func installEscFnKeyLatch(w fyne.Window) {
	if w == nil {
		return
	}
	w.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		_ = latchEscFnKey(key)
	})
}

// readOnlyPane is a disabled multiline Entry: selectable + copy, not editable.
// TypedKey latches Esc/Fn so EventEscFn still works when the pane has focus.
type readOnlyPane struct {
	widget.Entry
}

// newReadOnlyPane is a disabled multiline Entry: selectable + copy, not editable.
// placeholder is Fyne's standard empty-state ghost text (same API as the prompt).
func newReadOnlyPane(placeholder string) *readOnlyPane {
	e := &readOnlyPane{}
	e.MultiLine = true
	e.Wrapping = fyne.TextWrapOff
	e.TextStyle = fyne.TextStyle{Monospace: true}
	e.SetPlaceHolder(placeholder)
	e.ExtendBaseWidget(e)
	e.Disable()
	return e
}

func (e *readOnlyPane) TypedKey(key *fyne.KeyEvent) {
	if latchEscFnKey(key) {
		return
	}
	e.Entry.TypedKey(key)
}
