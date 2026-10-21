package finalgui

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/internal/core"
)

// responseMaxLines / historyMaxLines cap the REPL panes (terminal-style scrollback).
// standingMaxLines caps standing stdout/stderr LineBuffers (blit when View ▸ Stdout/Stderr).
const (
	responseMaxLines     = 320
	historyMaxLines      = 24
	standingMaxLines     = 160
	standingPollInterval = 250 * time.Millisecond
)

// FinalGui is the Go-native shell singleton (chrome widgets + prompt handler).
type FinalGui struct {
	typ            *finalembed.Type
	sess           *finalembed.Session
	win            fyne.Window
	chromeView     chromeView
	upperHost      *fyne.Container
	replPanes      fyne.CanvasObject
	streamPanes    fyne.CanvasObject
	viewREPLItem   *fyne.MenuItem
	viewStreamsItem *fyne.MenuItem
	prompt         *promptEntry
	response       *readOnlyPane
	history        *readOnlyPane
	stdoutPane     *readOnlyPane
	stderrPane     *readOnlyPane
	send           *widget.Button
	busyStatus     *canvas.Text // "idle" / "busy" under Send (TUI status semantics)
	busy           atomic.Bool
	responseLog    *core.LineBufferTextWriter
	historyLog     *core.LineBufferTextWriter
	stdoutLog      *core.LineBufferTextWriter
	stderrLog      *core.LineBufferTextWriter
	stdoutLastBlit atomic.Int64
	stderrLastBlit atomic.Int64
	standingStop   chan struct{}
	standingStopOnce sync.Once
	lastFileLoadPath string // last loaded file path; next Load starts in its directory
}

// FinalGui_ is the package singleton. Final sees the same instance after Init.
var FinalGui_ = &FinalGui{}

// FinalType implements Typed for host instances on the Final stack.
func (g *FinalGui) FinalType() *finalembed.Type { return g.typ }

func registerFinalGui(ty *finalembed.Type) {
	FinalGui_.typ = ty
	ty.RegisterNativeClassMethod("instance", func(f *finalembed.RunFrame, _ *finalembed.TypeObjPair) {
		f.RP.DataStack.Push(FinalGui_)
	})
	reg := func(name string, fn func(*FinalGui, *finalembed.RunFrame)) {
		ty.RegisterNativeMethod(name, func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
			fn(ths.Obj.(*FinalGui), f)
		})
	}
	// ( ev:Event -- ) — Event.params[0] may carry prompt/path strings
	reg("onSend", (*FinalGui).onSend)
	reg("onLoad", (*FinalGui).onLoad)
	reg("onQuit", (*FinalGui).onQuit)
	reg("onWords", (*FinalGui).onWords)
	reg("onAbout", (*FinalGui).onAbout)
}

func (g *FinalGui) onSend(f *finalembed.RunFrame) {
	src := takeEventStringParam(f)
	if src == "" {
		g.clearBusyFromWorker()
		return
	}
	g.runInterpret(f, src, true)
}

func (g *FinalGui) onLoad(f *finalembed.RunFrame) {
	path := takeEventStringParam(f)
	if path == "" {
		g.clearBusyFromWorker()
		return
	}
	script, err := compilePathScript(path)
	if err != nil {
		g.appendHistory(path)
		g.appendResponse("", err)
		g.clearBusyFromWorker()
		return
	}
	g.runInterpret(f, script, true)
}

func (g *FinalGui) onQuit(f *finalembed.RunFrame) {
	popEvent(f)
	if _, err := core.InterpretSource(f.Task, f.RP, "EventLoop.inst.close^"); err != nil {
		g.appendResponse("", err)
		return
	}
	Quit()
}

func (g *FinalGui) onWords(f *finalembed.RunFrame) {
	popEvent(f)
	g.runInterpret(f, "list", false)
}

func (g *FinalGui) onAbout(f *finalembed.RunFrame) {
	popEvent(f)
	msg := fmt.Sprintf("Final version: %s\n%s", Version(), Copyright())
	fyne.Do(func() {
		if g.win == nil {
			return
		}
		dialog.ShowInformation("About", msg, g.win)
	})
}

func (g *FinalGui) runInterpret(f *finalembed.RunFrame, src string, recordHistory bool) {
	out, err := func() (stdout string, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()
		return core.InterpretSource(f.Task, f.RP, src)
	}()
	if recordHistory {
		g.appendHistory(src)
	}
	g.replayStdout(out)
	if fault, ok := err.(*core.RunFault); ok && f != nil && f.RP != nil && f.RP.Stdout != nil {
		if !core.IsLocatedFaultMessage(fault.Message) {
			// Standing stdout only; appendResponse owns the result pane (avoid Hook double-print).
			f.RP.Stdout.WriteString(core.FormatRunFaultReport(fault))
			if fl, ok := f.RP.Stdout.(core.Flushable); ok {
				fl.Flush()
			}
		}
	}
	g.appendResponse(out, err)
	g.clearBusyFromWorker()
	if core.CheckResetCoreByeReq() {
		Quit()
		return
	}
	// Nested InterpretSource can leave OS/XS/… dirty (RS already unnested). Full
	// warmBootReset would also clear #run's foreach frame — escalate so
	// RunEventLoopSession re-enters the trampoline (185).
	if err != nil {
		core.AfterNestedInterpretFault()
	}
}

// markBusy is UI-thread: idle→busy before enqueueing interpret work (TUI setBusy).
func (g *FinalGui) markBusy() {
	g.busy.Store(true)
	g.applyBusyChrome(true)
}

// clearBusyFromWorker is EventLoop-thread: busy→idle after interpret (or early abort).
func (g *FinalGui) clearBusyFromWorker() {
	g.busy.Store(false)
	fyne.Do(func() {
		g.applyBusyChrome(false)
		if g.win != nil && g.prompt != nil {
			g.win.Canvas().Focus(g.prompt)
		}
	})
}

func (g *FinalGui) applyBusyChrome(busy bool) {
	if g.busyStatus != nil {
		if busy {
			g.busyStatus.Text = "busy"
			g.busyStatus.Color = color.NRGBA{R: 0xc0, G: 0, B: 0, A: 0xff}
		} else {
			g.busyStatus.Text = "idle"
			g.busyStatus.Color = color.NRGBA{R: 0, G: 0, B: 0, A: 0xff}
		}
		g.busyStatus.Refresh()
	}
	if g.prompt != nil {
		if busy {
			g.prompt.Disable()
		} else {
			g.prompt.Enable()
		}
	}
	if g.send != nil {
		if busy {
			g.send.Disable()
		} else {
			g.send.Enable()
		}
	}
}

// replayStdout copies captured Send `$` onto the standing session log (176).
// Fault reports are written separately onto standing stdout after Send.
func (g *FinalGui) replayStdout(out string) {
	if out == "" || g.stdoutLog == nil {
		return
	}
	g.stdoutLog.WriteString(out)
}

// takeEventStringParam pops the Event and returns Params[0] as a trimmed string.
func takeEventStringParam(f *finalembed.RunFrame) string {
	ev := popEvent(f)
	if ev == nil || ev.Params == nil || ev.Params.Len() == 0 {
		return ""
	}
	s, ok := ev.Params.At(0).(core.String)
	if !ok {
		return ""
	}
	return strings.TrimRight(string(s), "\r\n")
}

func popEvent(f *finalembed.RunFrame) *core.Event {
	if f == nil || f.RP == nil || f.RP.DataStack.Len() == 0 {
		return nil
	}
	ev, _ := f.RP.DataStack.Pop().(*core.Event)
	return ev
}

func (g *FinalGui) appendHistory(line string) {
	if line == "" {
		return
	}
	if g.historyLog == nil {
		g.historyLog = core.NewLineBufferTextWriter(historyMaxLines)
	}
	g.historyLog.WriteLine(line)
	full := g.historyLog.ToString()
	if g.history == nil {
		return
	}
	fyne.Do(func() {
		setReadOnlyText(g.history, full)
	})
}

// appendResponse appends interpret output to the Response pane (TUI convention):
// output text if any; otherwise "(ok)". Faults are the response when present.
// Kept to the last responseMaxLines complete lines via LineBufferTextWriter.
func (g *FinalGui) appendResponse(out string, err error) {
	var b strings.Builder
	if out != "" {
		b.WriteString(out)
		if !strings.HasSuffix(out, "\n") {
			b.WriteByte('\n')
		}
	}
	if err != nil {
		msg := err.Error()
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
	if g.responseLog == nil {
		g.responseLog = core.NewLineBufferTextWriter(responseMaxLines)
	}
	g.responseLog.WriteString(b.String())
	full := g.responseLog.ToString()
	if g.response == nil {
		return
	}
	fyne.Do(func() {
		setReadOnlyText(g.response, full)
	})
}

func (g *FinalGui) startStandingPoller() {
	if g.standingStop != nil {
		return
	}
	g.standingStopOnce = sync.Once{}
	g.standingStop = make(chan struct{})
	tick := time.NewTicker(standingPollInterval)
	stop := g.standingStop
	go func() {
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				g.pollStandingPanes()
			case <-stop:
				return
			}
		}
	}()
}

func (g *FinalGui) stopStandingPoller() {
	g.standingStopOnce.Do(func() {
		if g.standingStop != nil {
			close(g.standingStop)
			g.standingStop = nil
		}
	})
}

func (g *FinalGui) stdoutNeedsBlit() bool {
	if g.chromeView != chromeViewStreams || g.stdoutLog == nil || g.stdoutPane == nil {
		return false
	}
	tick := g.stdoutLog.GetStringLineBuffer().GetTick()
	return int64(tick) != g.stdoutLastBlit.Load()
}

func (g *FinalGui) stderrNeedsBlit() bool {
	if g.chromeView != chromeViewStreams || g.stderrLog == nil || g.stderrPane == nil {
		return false
	}
	tick := g.stderrLog.GetStringLineBuffer().GetTick()
	return int64(tick) != g.stderrLastBlit.Load()
}

func (g *FinalGui) pollStandingPanes() {
	needOut := g.stdoutPane != nil && g.stdoutNeedsBlit()
	needErr := g.stderrPane != nil && g.stderrNeedsBlit()
	if !needOut && !needErr {
		return
	}
	fyne.Do(func() {
		if g.stdoutNeedsBlit() {
			g.blitStdoutPane()
		}
		if g.stderrNeedsBlit() {
			g.blitStderrPane()
		}
	})
}

// blitStdoutPane copies the standing log into the stdout Entry. Call on the UI thread.
func (g *FinalGui) blitStdoutPane() {
	if g.stdoutLog == nil || g.stdoutPane == nil {
		return
	}
	text := g.stdoutLog.ToString()
	tick := g.stdoutLog.GetStringLineBuffer().GetTick()
	follow := !g.streamPaneFocused(g.stdoutPane)
	if !setReadOnlyTextStick(g.stdoutPane, text, follow) {
		return
	}
	g.stdoutLastBlit.Store(int64(tick))
}

// blitStderrPane copies the standing stderr log into the stderr Entry. Call on the UI thread.
func (g *FinalGui) blitStderrPane() {
	if g.stderrLog == nil || g.stderrPane == nil {
		return
	}
	text := g.stderrLog.ToString()
	tick := g.stderrLog.GetStringLineBuffer().GetTick()
	follow := !g.streamPaneFocused(g.stderrPane)
	if !setReadOnlyTextStick(g.stderrPane, text, follow) {
		return
	}
	g.stderrLastBlit.Store(int64(tick))
}

func (g *FinalGui) streamPaneFocused(e *readOnlyPane) bool {
	if g == nil || g.win == nil || e == nil {
		return false
	}
	return g.win.Canvas().Focused() == e
}

// setReadOnlyText updates a disabled multiline Entry (Enable/SetText/Disable so the
// view refreshes and selection/copy stay available). Always scrolls to the end —
// used for Response/History where updates are submit-driven, not polled.
func setReadOnlyText(e *readOnlyPane, text string) {
	_ = setReadOnlyTextStick(e, text, true)
}

// setReadOnlyTextStick updates a disabled multiline Entry. If the user has a
// selection, returns false and leaves the widget unchanged (poller retries).
// When follow is true (pane not focused — watching from the prompt), scroll to
// end. When follow is false (pane focused), restore CursorRow/Column so
// scroll-back and reading stay put while text still updates.
func setReadOnlyTextStick(e *readOnlyPane, text string, follow bool) bool {
	if e == nil {
		return true
	}
	if e.SelectedText() != "" {
		return false
	}
	oldRow, oldCol := e.CursorRow, e.CursorColumn

	e.Enable()
	e.SetText(text)
	newLast := strings.Count(text, "\n")
	if follow {
		e.CursorRow = newLast
		e.CursorColumn = 0
	} else {
		if oldRow > newLast {
			oldRow = newLast
		}
		e.CursorRow = oldRow
		e.CursorColumn = oldCol
	}
	e.Refresh()
	e.Disable()
	return true
}

func methodRef(name string) *core.MethodRef {
	if FinalGui_.typ == nil {
		panic("finalgui: FinalGui type not registered")
	}
	mm := FinalGui_.typ.GetMethodMetadata(name)
	if mm == nil {
		panic("finalgui: missing FinalGui#" + name)
	}
	return &core.MethodRef{
		Parent: core.NewTypeObjPair(FinalGui_),
		MM:     mm,
	}
}

func (g *FinalGui) loadStartDir() string {
	if g.lastFileLoadPath != "" {
		dir := filepath.Dir(g.lastFileLoadPath)
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			return dir
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}

// openLoadDialog runs on the Fyne UI thread: file picker, then enqueue onLoad with path.
func openLoadDialog() {
	g := FinalGui_
	if g.win == nil {
		return
	}
	loadFn := core.GetEventLoop().CreateEventFn(methodRef("onLoad"), 0)
	d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		uri := reader.URI()
		_ = reader.Close()
		if uri == nil {
			return
		}
		path := uri.Path()
		if path == "" {
			return
		}
		g.lastFileLoadPath = path
		if g.busy.Load() {
			return
		}
		g.markBusy()
		loadFn(nil, core.String(path))
	}, g.win)
	d.SetView(dialog.ListView)
	if dir := g.loadStartDir(); dir != "" {
		if u, err := storage.ListerForURI(storage.NewFileURI(dir)); err == nil {
			d.SetLocation(u)
		}
	}
	d.Show()
	min := d.MinSize()
	d.Resize(fyne.NewSize(min.Width*1.5, min.Height*2))
}
