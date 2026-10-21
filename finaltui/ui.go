package finaltui

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/internal/core"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.design/x/clipboard"
)

// responseMaxLines / historyMaxLines match final-gui REPL scrollback.
// standingMaxLines matches final-gui standing $ / $$ buffers.
const (
	responseMaxLines     = 160
	historyMaxLines      = 24
	standingMaxLines     = 160
	standingPollInterval = 250 * time.Millisecond
)

type chromeView int

const (
	chromeViewREPL chromeView = iota // Response | History — menu "Prompt/History"
	chromeViewStreams                // Stdout | Stderr — menu "Stdout/Stderr"
)

// Button look: steady blue; busy = grayer blue; no white flash on focus/press.
var (
	sendStyle = tcell.StyleDefault.
			Background(tcell.ColorBlue).
			Foreground(tcell.ColorWhite)
	sendBusyStyle = tcell.StyleDefault.
			Background(tcell.ColorNavy).
			Foreground(tcell.ColorLightGray)
)

// appUI is the final chrome (tview pages, menus, dialogs).
type appUI struct {
	app            *tview.Application
	sess           *finalembed.Session
	pages          *tview.Pages
	chromeView     chromeView
	upperPages     *tview.Pages
	response       *tview.TextArea
	history        *tview.TextArea
	stdoutPane     *tview.TextArea
	stderrPane     *tview.TextArea
	responseWriter *core.LineBufferTextWriter
	historyWriter  *core.LineBufferTextWriter
	stdoutLog      *core.LineBufferTextWriter
	stderrLog      *core.LineBufferTextWriter
	stdoutLastBlit atomic.Int64
	stderrLastBlit atomic.Int64
	standingStop   chan struct{}
	standingStopOnce sync.Once
	prompt         *tview.TextArea
	status         *tview.TextView
	sendBtn        *tview.Button
	pathField      *tview.InputField
	viewDD         *tview.DropDown
	sendFn         core.HandlerMintEventFn
	quitFn         core.HandlerMintEventFn

	promptForm    *tview.Flex
	promptBody    *tview.TextView
	promptOk      *tview.Button
	promptCancel  *tview.Button
	promptFocusOk bool

	fieldsRoot   *tview.Flex
	fieldsInstr  *tview.TextView
	fieldsForm   *tview.Form
	fieldsError  *tview.TextView
	fieldsOk     *tview.Button
	fieldsCancel *tview.Button
	fieldsSpecs  []FieldSpec
	fieldsCtrls  []fieldCtrl

	aboutOk *tview.Button
	browser *Browser

	// lastChooserPath is the last file accepted via Load. Cancel does not update it.
	// Used as the next Load chooser start directory's parent.
	lastChooserPath string

	// editTarget is the last TextArea/InputField focused for Edit menu actions
	// (menu DropDowns steal focus before Copy/Cut/Paste run).
	editTarget tview.Primitive

	// clipboardOK is false when golang.design/x/clipboard cannot init (typical
	// SSH / headless: no DISPLAY). In-app Copy/Cut/Paste no-op; use the
	// terminal’s paste (bracketed paste). See design note 146.
	clipboardOK bool

	busy atomic.Bool
}

func newAppUI(sess *finalembed.Session) (*appUI, error) {
	useDoubleBordersAlways()

	ui := &appUI{
		app:            tview.NewApplication(),
		sess:           sess,
		clipboardOK:    true,
		chromeView:     chromeViewREPL,
		responseWriter: core.NewLineBufferTextWriter(responseMaxLines),
		historyWriter:  core.NewLineBufferTextWriter(historyMaxLines),
		stdoutLog:      core.NewLineBufferTextWriterSync(standingMaxLines),
		stderrLog:      core.NewLineBufferTextWriterSync(standingMaxLines),
	}
	if err := clipboard.Init(); err != nil {
		ui.clipboardOK = false
	}

	ui.response = ui.newReadOnlyArea(" Response ")
	ui.history = ui.newReadOnlyArea(" History ")
	ui.stdoutPane = ui.newReadOnlyArea(" $, Errors (Stdout) ")
	ui.stderrPane = ui.newReadOnlyArea(" $$ (Stderr) ")
	ui.trackEditTarget(ui.response)
	ui.trackEditTarget(ui.history)
	ui.trackEditTarget(ui.stdoutPane)
	ui.trackEditTarget(ui.stderrPane)

	// ">" is a fixed gutter (not a placeholder), so typing starts after it.
	ui.prompt = tview.NewTextArea().SetWrap(true)
	ui.wireOSClipboard(ui.prompt)
	ui.prompt.SetInputCapture(ui.promptCapture)
	ui.trackEditTarget(ui.prompt)
	ui.editTarget = ui.prompt

	promptGutter := tview.NewTextView().
		SetText(">").
		SetTextAlign(tview.AlignLeft)
	promptBody := tview.NewFlex().
		AddItem(promptGutter, 2, 0, false).
		AddItem(ui.prompt, 0, 1, true)
	promptBody.SetBorder(true).SetTitle(" Prompt ")

	ui.sendBtn = tview.NewButton("Send").
		SetStyle(sendStyle).
		SetActivatedStyle(sendStyle). // no invert / white flash when focused or pressed
		SetDisabledStyle(sendBusyStyle).
		SetSelectedFunc(ui.submit)

	ui.status = tview.NewTextView().
		SetDynamicColors(true).
		SetText(shortcutLegend(false, ui.clipboardOK))

	// Fixed-height Send, vertically centered beside the Prompt box.
	sendCol := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(ui.sendBtn, 3, 0, false).
		AddItem(nil, 0, 1, false)

	promptRow := tview.NewFlex().
		AddItem(promptBody, 0, 1, true).
		AddItem(sendCol, 12, 0, false)

	replUpper := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ui.response, 0, 1, false).
		AddItem(ui.history, 7, 0, false)
	streamsUpper := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ui.stdoutPane, 0, 1, false).
		AddItem(ui.stderrPane, 0, 1, false)
	ui.upperPages = tview.NewPages().
		AddPage("repl", replUpper, true, true).
		AddPage("streams", streamsUpper, true, false)

	// History: ~5 content lines (height 7 with border); stole ~5 from Response + 1 from Prompt.
	main := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ui.buildMenuBar(), 1, 0, false).
		AddItem(ui.upperPages, 0, 1, false).
		AddItem(promptRow, 7, 0, true).
		AddItem(ui.status, 1, 0, false)

	ui.pages = tview.NewPages().
		AddPage(pageMain, main, true, true).
		AddPage(pageOpen, ui.buildOpenPage(), true, false).
		AddPage(pageBrowse, ui.buildBrowsePage(), true, false).
		AddPage(pagePrompt, ui.buildPromptPage(), true, false).
		AddPage(pageFields, ui.buildFieldsPage(), true, false).
		AddPage(pageAbout, ui.buildAboutPage(), true, false)

	ui.app.SetRoot(ui.pages, true).EnableMouse(true).EnablePaste(true)
	// App-level capture runs before the focused primitive (tview Application).
	ui.app.SetInputCapture(ui.globalCapture)
	ui.startUIReqDrain()
	return ui, nil
}

func (ui *appUI) setChromeView(v chromeView) {
	if ui.chromeView == v {
		return
	}
	ui.chromeView = v
	switch v {
	case chromeViewStreams:
		ui.upperPages.SwitchToPage("streams")
		// ForceDraw lays out the streams TextAreas (updates lastWidth/lastHeight).
		// Safe during DropDown handling; Application.Draw → QueueUpdate would deadlock.
		if ui.app != nil {
			ui.app.ForceDraw()
		}
		ui.blitStdoutPane(true)
		ui.blitStderrPane(true)
	default:
		ui.upperPages.SwitchToPage("repl")
		ui.refreshREPLPanes()
	}
	ui.refreshViewMenu()
}

func (ui *appUI) refreshREPLPanes() {
	if ui.response != nil && ui.responseWriter != nil {
		ui.response.SetText(ui.responseWriter.ToString(), true)
	}
	if ui.history != nil && ui.historyWriter != nil {
		ui.history.SetText(ui.historyWriter.ToString(), true)
	}
}

func (ui *appUI) refreshStreamPanes() {
	ui.blitStdoutPane(textAreaAtBottom(ui.stdoutPane))
	ui.blitStderrPane(textAreaAtBottom(ui.stderrPane))
}

// textAreaAtBottom reports whether ta is pinned for live-tail: caret at end and
// still visible in the viewport. Mouse-wheel scroll without focus leaves the caret
// at end but moves offset — then the caret is below the viewport and we stop
// following so the poller does not yank the view.
func textAreaAtBottom(ta *tview.TextArea) bool {
	if ta == nil {
		return true
	}
	_, _, _, h := ta.GetInnerRect()
	if h <= 0 {
		return true
	}
	if ta.HasSelection() {
		return false
	}
	_, start, end := ta.GetSelection()
	if end < ta.GetTextLength() || start != end {
		return false
	}
	_, _, curRow, _ := ta.GetCursor()
	offRow, _ := ta.GetOffset()
	return curRow >= offRow && curRow < offRow+h
}

func (ui *appUI) startStandingPoller() {
	if ui.standingStop != nil {
		return
	}
	ui.standingStopOnce = sync.Once{}
	ui.standingStop = make(chan struct{})
	tick := time.NewTicker(standingPollInterval)
	stop := ui.standingStop
	go func() {
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				ui.pollStandingPanes()
			case <-stop:
				return
			}
		}
	}()
}

func (ui *appUI) stopStandingPoller() {
	ui.standingStopOnce.Do(func() {
		if ui.standingStop != nil {
			close(ui.standingStop)
			ui.standingStop = nil
		}
	})
}

func (ui *appUI) stdoutNeedsBlit() bool {
	if ui.chromeView != chromeViewStreams || ui.stdoutLog == nil || ui.stdoutPane == nil {
		return false
	}
	tick := ui.stdoutLog.GetStringLineBuffer().GetTick()
	return int64(tick) != ui.stdoutLastBlit.Load()
}

func (ui *appUI) stderrNeedsBlit() bool {
	if ui.chromeView != chromeViewStreams || ui.stderrLog == nil || ui.stderrPane == nil {
		return false
	}
	tick := ui.stderrLog.GetStringLineBuffer().GetTick()
	return int64(tick) != ui.stderrLastBlit.Load()
}

func (ui *appUI) pollStandingPanes() {
	if ui.app == nil {
		return
	}
	needOut := ui.stdoutPane != nil && ui.stdoutNeedsBlit()
	needErr := ui.stderrPane != nil && ui.stderrNeedsBlit()
	if !needOut && !needErr {
		return
	}
	ui.app.QueueUpdateDraw(func() {
		if ui.stdoutNeedsBlit() {
			ui.blitStdoutPane(textAreaAtBottom(ui.stdoutPane))
		}
		if ui.stderrNeedsBlit() {
			ui.blitStderrPane(textAreaAtBottom(ui.stderrPane))
		}
	})
}

// blitStdoutPane copies the standing log into the stdout TextArea. Call on the UI thread.
// follow true → scroll to end; false → preserve offset.
func (ui *appUI) blitStdoutPane(follow bool) {
	if ui.stdoutLog == nil || ui.stdoutPane == nil {
		return
	}
	text := ui.stdoutLog.ToString()
	tick := ui.stdoutLog.GetStringLineBuffer().GetTick()
	if !setTextAreaStickBottom(ui.stdoutPane, text, follow) {
		return // selection held; leave lastBlit so the poller retries
	}
	ui.stdoutLastBlit.Store(int64(tick))
}

// blitStderrPane copies the standing stderr log into the stderr TextArea. Call on the UI thread.
func (ui *appUI) blitStderrPane(follow bool) {
	if ui.stderrLog == nil || ui.stderrPane == nil {
		return
	}
	text := ui.stderrLog.ToString()
	tick := ui.stderrLog.GetStringLineBuffer().GetTick()
	if !setTextAreaStickBottom(ui.stderrPane, text, follow) {
		return
	}
	ui.stderrLastBlit.Store(int64(tick))
}

// setTextAreaStickBottom updates ta to text. follow scrolls to end via SetText(true)
// once the area has a layout size (call ForceDraw after SwitchToPage first).
// follow false preserves offset; returns false if selecting (poller retries).
func setTextAreaStickBottom(ta *tview.TextArea, text string, follow bool) bool {
	if ta == nil {
		return true
	}
	if follow {
		_, _, w, h := ta.GetInnerRect()
		if w > 0 && h > 0 {
			ta.SetText(text, true)
			return true
		}
		ta.SetText(text, false)
		return true
	}
	if ta.HasSelection() {
		return false
	}
	offRow, offCol := ta.GetOffset()
	ta.SetText(text, false)
	ta.SetOffset(offRow, offCol)
	// Keep caret at end (offsets unchanged) so scrolling back to the bottom
	// makes textAreaAtBottom true and live-tail resumes.
	n := ta.GetTextLength()
	ta.Select(n, n)
	return true
}

func (ui *appUI) isReadOnlyPane(ta *tview.TextArea) bool {
	return ta == ui.response || ta == ui.history || ta == ui.stdoutPane || ta == ui.stderrPane
}

func (ui *appUI) trackEditTarget(p tview.Primitive) {
	switch t := p.(type) {
	case *tview.TextArea:
		t.SetFocusFunc(func() { ui.editTarget = t })
	case *tview.InputField:
		t.SetFocusFunc(func() { ui.editTarget = t })
	}
}

func (ui *appUI) newReadOnlyArea(title string) *tview.TextArea {
	ta := tview.NewTextArea().SetWrap(true)
	ta.SetBorder(true).SetTitle(title)
	ui.wireOSClipboard(ta)
	ta.SetInputCapture(readOnlyCapture)
	return ta
}

func readOnlyCapture(ev *tcell.EventKey) *tcell.EventKey {
	switch ev.Key() {
	case tcell.KeyRune, tcell.KeyEnter, tcell.KeyCtrlV, tcell.KeyCtrlX,
		tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyDelete,
		tcell.KeyTab, tcell.KeyBacktab: // never insert \t into read-only panes
		return nil
	}
	return ev
}

// useDoubleBordersAlways makes focused and unfocused boxes use the same double-line
// runes (tview defaults to light single-line when unfocused).
func useDoubleBordersAlways() {
	tview.Borders.Horizontal = tview.Borders.HorizontalFocus
	tview.Borders.Vertical = tview.Borders.VerticalFocus
	tview.Borders.TopLeft = tview.Borders.TopLeftFocus
	tview.Borders.TopRight = tview.Borders.TopRightFocus
	tview.Borders.BottomLeft = tview.Borders.BottomLeftFocus
	tview.Borders.BottomRight = tview.Borders.BottomRightFocus
}

func (ui *appUI) wireOSClipboard(ta *tview.TextArea) {
	if !ui.clipboardOK {
		return
	}
	ta.SetClipboard(
		func(text string) {
			clipboard.Write(clipboard.FmtText, []byte(text))
		},
		func() string {
			return string(clipboard.Read(clipboard.FmtText))
		},
	)
}

func shortcutLegend(busy, clipboardOK bool) string {
	busyNote := "idle"
	if busy {
		busyNote = "[red]BUSY[-]  ^B break"
	}
	if !clipboardOK {
		return fmt.Sprintf(" %s  |  ^B break  ^Q quit  ^L load  Enter send  ^J newline  |  clipboard off — use terminal paste ", busyNote)
	}
	return fmt.Sprintf(" %s  |  ^B break  ^Q quit  ^L load  Enter send  ^J newline  ^X cut  ^C copy  ^V paste ", busyNote)
}

func (ui *appUI) setBusy(busy bool) {
	ui.busy.Store(busy)
	ui.prompt.SetDisabled(busy)
	ui.sendBtn.SetDisabled(busy)
	ui.sendBtn.SetLabel("Send") // keep lettering; style alone shows busy
	ui.status.SetText(shortcutLegend(busy, ui.clipboardOK))
}

func (ui *appUI) appendResponse(s string) {
	ui.responseWriter.WriteString(s)
	if ui.response == nil || ui.chromeView != chromeViewREPL {
		return
	}
	ui.response.SetText(ui.responseWriter.ToString(), true)
}

// appendHistory records a sent script. WriteLine matches final-gui (trailing
// newline per tranche; empty skipped).
func (ui *appUI) appendHistory(script string) {
	if script == "" {
		return
	}
	ui.historyWriter.WriteLine(script)
	if ui.history == nil || ui.chromeView != chromeViewREPL {
		return
	}
	ui.history.SetText(ui.historyWriter.ToString(), true)
}

// replayStdout copies captured Send `$` onto the standing session log (GUI parity).
func (ui *appUI) replayStdout(out string) {
	if out == "" || ui.stdoutLog == nil {
		return
	}
	ui.stdoutLog.WriteString(out)
	if ui.chromeView == chromeViewStreams {
		ui.refreshStreamPanes()
	}
}

func (ui *appUI) promptCapture(ev *tcell.EventKey) *tcell.EventKey {
	if ui.busy.Load() {
		return nil
	}
	// Newline: Shift-Enter when the terminal reports ModShift (CSI-u / Kitty protocol).
	// Many terminals (incl. macOS Terminal, cmd.exe) send Shift-Enter as plain Enter —
	// then Ctrl-J is the portable newline (LF ≠ Enter's CR).
	if isNewlineKey(ev) {
		insertNewline(ui.prompt)
		return nil
	}
	if ev.Key() == tcell.KeyEnter {
		ui.submit()
		return nil
	}
	// No focus-cycle Tab on main chrome; TextArea would insert \t — Final uses \t escape.
	// PromptDialog / FieldDialog will use Form Tab later; leave app-level Tab alone for Forms.
	if ev.Key() == tcell.KeyTab || ev.Key() == tcell.KeyBacktab {
		return nil
	}
	return ev
}

// isNewlineKey reports Shift-Enter (when distinguished) or Ctrl-J.
func isNewlineKey(ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyEnter && ev.Modifiers()&tcell.ModShift != 0 {
		return true
	}
	if ev.Key() == tcell.KeyCtrlJ {
		return true
	}
	return false
}

func insertNewline(ta *tview.TextArea) {
	_, start, end := ta.GetSelection()
	ta.Replace(start, end, tview.NewLine)
	// Replace documents that scroll offsets are unchanged; KeyEnter uses findCursor(clamp).
	scrollCursorIntoView(ta)
}

// scrollCursorIntoView adjusts TextArea row offset so the cursor row is visible.
func scrollCursorIntoView(ta *tview.TextArea) {
	_, _, _, h := ta.GetInnerRect()
	if h <= 0 {
		return
	}
	_, _, cursorRow, _ := ta.GetCursor()
	offRow, offCol := ta.GetOffset()
	if cursorRow < offRow {
		ta.SetOffset(cursorRow, offCol)
		return
	}
	if cursorRow >= offRow+h {
		ta.SetOffset(cursorRow-h+1, offCol)
	}
}

func (ui *appUI) requestBreak() {
	ui.sess.Break()
	ui.cancelDialogIfOpen()
}

func (ui *appUI) globalCapture(ev *tcell.EventKey) *tcell.EventKey {
	if n, ok := escFnCodeFromTcell(ev); ok {
		core.LatchEscFn(n)
		// Esc also dismisses menus/dialogs — latch and pass through.
		if n == core.EscFnEsc {
			return ev
		}
		return nil // F1–F12: consume (TextArea ignores them anyway)
	}
	switch ev.Key() {
	case tcell.KeyCtrlB:
		ui.requestBreak()
		return nil
	case tcell.KeyCtrlL:
		ui.showBrowseLoad()
		return nil
	case tcell.KeyCtrlC:
		// Copy. Must consume (return nil): tview Application Stops if the
		// original KeyCtrlC event object is returned unchanged.
		ui.editCopy()
		return nil
	case tcell.KeyCtrlQ:
		ui.requestQuit()
		return nil
	}
	return ev
}

// escFnCodeFromTcell maps Esc / F1–F12 to latch codes (0 / 1–12).
func escFnCodeFromTcell(ev *tcell.EventKey) (int, bool) {
	switch ev.Key() {
	case tcell.KeyEscape:
		return core.EscFnEsc, true
	case tcell.KeyF1:
		return 1, true
	case tcell.KeyF2:
		return 2, true
	case tcell.KeyF3:
		return 3, true
	case tcell.KeyF4:
		return 4, true
	case tcell.KeyF5:
		return 5, true
	case tcell.KeyF6:
		return 6, true
	case tcell.KeyF7:
		return 7, true
	case tcell.KeyF8:
		return 8, true
	case tcell.KeyF9:
		return 9, true
	case tcell.KeyF10:
		return 10, true
	case tcell.KeyF11:
		return 11, true
	case tcell.KeyF12:
		return 12, true
	default:
		return 0, false
	}
}

func (ui *appUI) submit() {
	script := strings.TrimSpace(ui.prompt.GetText())
	if script == "" {
		return
	}
	ui.sendScript(script, true)
}

// sendScript mints an onSend Event (EventLoop worker interprets). Never Session.Send
// after the loop starts — sendMu is held by RunEventLoop.
// clearPrompt is true for Prompt Send; false for Open/Load (prompt left alone).
func (ui *appUI) sendScript(script string, clearPrompt bool) {
	if ui.busy.Load() {
		return
	}
	if strings.TrimSpace(script) == "" {
		return
	}
	if ui.sendFn == nil {
		return
	}
	ui.setBusy(true)
	if clearPrompt {
		ui.prompt.SetText("", true)
	}
	ui.sendFn(nil, core.String(script))
}

func (ui *appUI) requestQuit() {
	if ui.quitFn == nil {
		ui.app.Stop()
		return
	}
	ui.quitFn(nil)
}

// runInterpret runs src on the EventLoop task (InterpretSource). Call from onSend.
func (ui *appUI) runInterpret(f *finalembed.RunFrame, src string, recordHistory bool) {
	out, err := func() (stdout string, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()
		return core.InterpretSource(f.Task, f.RP, src)
	}()

	ui.app.QueueUpdateDraw(func() {
		if recordHistory {
			ui.appendHistory(src)
		}
		ui.replayStdout(out)
		var b strings.Builder
		if out != "" {
			b.WriteString(out)
			if !strings.HasSuffix(out, "\n") {
				b.WriteByte('\n')
			}
		}
		if err != nil {
			msg := err.Error()
			if fault, ok := err.(*core.RunFault); ok {
				if !core.IsLocatedFaultMessage(fault.Message) && ui.stdoutLog != nil {
					ui.stdoutLog.WriteString(core.FormatRunFaultReport(fault))
					if ui.chromeView == chromeViewStreams {
						ui.refreshStreamPanes()
					}
				}
			}
			if core.IsLocatedFaultMessage(msg) {
				b.WriteString(msg)
				if !strings.HasSuffix(msg, "\n") {
					b.WriteByte('\n')
				}
			} else if fault, ok := err.(*core.RunFault); ok {
				b.WriteString(core.FormatRunFaultReport(fault))
			} else {
				b.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		}
		if b.Len() == 0 {
			b.WriteString("(ok)\n")
		}
		ui.appendResponse(b.String())
		ui.setBusy(false)
		ui.app.SetFocus(ui.prompt)
	})

	if core.CheckResetCoreByeReq() {
		ui.requestQuit()
		return
	}
	if err != nil {
		core.AfterNestedInterpretFault()
	}
}

func (ui *appUI) Run() error {
	return ui.app.Run()
}

// RunTUI runs headed tview with EventLoop as the worker root (GUI-shaped invert).
// LoadHostManifests must already have run (still via Send, before this).
func RunTUI(sess *finalembed.Session) error {
	ui, err := newAppUI(sess)
	if err != nil {
		return err
	}
	Tui_.ArmUI()
	Tui_.bindChrome(ui)
	defer Tui_.unbindChrome()

	if ui.stdoutLog != nil {
		sess.SetStdout(ui.stdoutLog)
	}
	if ui.stderrLog != nil {
		sess.SetStderr(ui.stderrLog)
	}
	ui.startStandingPoller()
	defer ui.stopStandingPoller()

	core.RunFaultReportHook = func(fault *core.RunFault) {
		ui.app.QueueUpdateDraw(func() {
			ui.appendResponse(core.FormatRunFaultReport(fault))
		})
	}
	defer func() { core.RunFaultReportHook = nil }()

	el := core.GetEventLoop()
	ui.sendFn = el.CreateEventFn(core.NewMethodRef(Tui_, "onSend"), 0)
	ui.quitFn = el.CreateEventFn(core.NewMethodRef(Tui_, "onQuit"), 0)

	workerDone := make(chan error, 1)
	go func() {
		workerDone <- sess.RunEventLoop(TuiRootWord)
	}()

	runErr := ui.Run()

	// Ensure the worker exits if the UI stopped without onQuit (e.g. SIGTERM path).
	el.Close()
	<-workerDone
	return runErr
}
