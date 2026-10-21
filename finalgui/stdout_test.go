package finalgui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

func TestReplayStdoutIntoLog(t *testing.T) {
	g := &FinalGui{stdoutLog: core.NewLineBufferTextWriterSync(8)}
	g.replayStdout("")
	if got := g.stdoutLog.ToString(); got != "" {
		t.Fatalf("empty replay = %q", got)
	}
	g.replayStdout("3")
	if got := g.stdoutLog.ToString(); got != "3" {
		t.Fatalf("replay = %q, want 3", got)
	}
	if g.stdoutLog.GetStringLineBuffer().GetTick() == 0 {
		t.Fatal("tick did not advance")
	}
}

func TestStdoutNeedsBlitByView(t *testing.T) {
	g := &FinalGui{
		stdoutLog:  core.NewLineBufferTextWriterSync(8),
		stdoutPane: newReadOnlyPane(""),
		chromeView: chromeViewREPL,
	}
	g.stdoutLog.WriteString("hello\n")
	if g.stdoutNeedsBlit() {
		t.Fatal("REPL view should not need stdout blit")
	}
	g.chromeView = chromeViewStreams
	if !g.stdoutNeedsBlit() {
		t.Fatal("streams view with new tick should need blit")
	}
	g.stdoutLastBlit.Store(int64(g.stdoutLog.GetStringLineBuffer().GetTick()))
	if g.stdoutNeedsBlit() {
		t.Fatal("matching lastBlit should not need blit")
	}
}

func TestStderrNeedsBlitByView(t *testing.T) {
	g := &FinalGui{
		stderrLog:  core.NewLineBufferTextWriterSync(8),
		stderrPane: newReadOnlyPane(""),
		chromeView: chromeViewREPL,
	}
	g.stderrLog.WriteString("err\n")
	if g.stderrNeedsBlit() {
		t.Fatal("REPL view should not need stderr blit")
	}
	g.chromeView = chromeViewStreams
	if !g.stderrNeedsBlit() {
		t.Fatal("streams view with new tick should need blit")
	}
	g.stderrLastBlit.Store(int64(g.stderrLog.GetStringLineBuffer().GetTick()))
	if g.stderrNeedsBlit() {
		t.Fatal("matching lastBlit should not need blit")
	}
}

func TestStandingLogCapsAtMaxLines(t *testing.T) {
	g := &FinalGui{
		stdoutLog: core.NewLineBufferTextWriterSync(standingMaxLines),
		stderrLog: core.NewLineBufferTextWriterSync(standingMaxLines),
	}
	if standingMaxLines != 160 {
		t.Fatalf("standingMaxLines = %d, want 160", standingMaxLines)
	}
	g.replayStdout("a\nb\nc\nd\n")
	if got := g.stdoutLog.ToString(); got != "a\nb\nc\nd\n" {
		t.Fatalf("stdout log = %q", got)
	}
	small := &FinalGui{stdoutLog: core.NewLineBufferTextWriterSync(3)}
	small.replayStdout("a\nb\nc\nd\n")
	if got := small.stdoutLog.ToString(); got != "b\nc\nd\n" {
		t.Fatalf("capped log = %q, want %q", got, "b\nc\nd\n")
	}
}

func TestSetChromeViewSwapsUpperHost(t *testing.T) {
	g := &FinalGui{
		replPanes:   widget.NewLabel("repl"),
		streamPanes: widget.NewLabel("streams"),
	}
	g.upperHost = &fyne.Container{Objects: []fyne.CanvasObject{g.replPanes}}
	g.viewREPLItem = &fyne.MenuItem{}
	g.viewStreamsItem = &fyne.MenuItem{}
	g.setChromeView(chromeViewREPL)
	if !g.viewREPLItem.Disabled || g.viewStreamsItem.Disabled {
		t.Fatal("REPL view: Prompt/History should be disabled, Stdout/Stderr enabled")
	}
	g.setChromeView(chromeViewStreams)
	if g.chromeView != chromeViewStreams || g.upperHost.Objects[0] != g.streamPanes {
		t.Fatal("streams view should show streamPanes")
	}
	if g.viewREPLItem.Disabled || !g.viewStreamsItem.Disabled {
		t.Fatal("streams view: Stdout/Stderr should be disabled, Prompt/History enabled")
	}
	g.setChromeView(chromeViewREPL)
	if g.upperHost.Objects[0] != g.replPanes {
		t.Fatal("REPL view should show replPanes")
	}
}

func TestSetReadOnlyTextStickFollowsWhenRequested(t *testing.T) {
	e := newReadOnlyPane("")
	e.Enable()
	e.SetText("a\nb\n")
	e.CursorRow = 0
	e.Disable()
	if !setReadOnlyTextStick(e, "a\nb\nc\n", true) {
		t.Fatal("expected apply")
	}
	if e.CursorRow != 3 {
		t.Fatalf("CursorRow = %d, want 3 (end)", e.CursorRow)
	}
}

func TestSetReadOnlyTextStickPreservesCaretWhenNotFollowing(t *testing.T) {
	e := newReadOnlyPane("")
	e.Enable()
	e.SetText("a\nb\nc\n")
	e.CursorRow = 1
	e.CursorColumn = 0
	e.Disable()
	if !setReadOnlyTextStick(e, "a\nb\nc\nd\n", false) {
		t.Fatal("expected apply")
	}
	if e.CursorRow != 1 {
		t.Fatalf("CursorRow = %d, want preserved 1", e.CursorRow)
	}
	if e.Text != "a\nb\nc\nd\n" {
		t.Fatalf("Text = %q", e.Text)
	}
}

func TestSetReadOnlyTextStickSkipsWhenSelecting(t *testing.T) {
	e := newReadOnlyPane("")
	e.Enable()
	e.SetText("hello")
	// Fyne selection is internal; select via keyboard API fields through DoubleTapped/Select.
	// Drive selection by setting cursor then extending — use Select() if available.
	e.TypedShortcut(&fyne.ShortcutSelectAll{})
	if e.SelectedText() == "" {
		t.Skip("could not establish selection on MultiLineEntry in this Fyne version")
	}
	e.Disable()
	if setReadOnlyTextStick(e, "hello\nworld", true) {
		t.Fatal("expected skip while selecting")
	}
	if e.Text != "hello" {
		t.Fatalf("Text changed while selecting: %q", e.Text)
	}
}
