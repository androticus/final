package finaltui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/androticus/final/internal/core"
	"github.com/rivo/tview"
)

func TestAppendResponseCapsAtMaxLines(t *testing.T) {
	ui := &appUI{responseWriter: core.NewLineBufferTextWriter(responseMaxLines)}
	for i := 0; i < responseMaxLines+5; i++ {
		ui.appendResponse(fmt.Sprintf("L%d\n", i))
	}
	got := ui.responseWriter.ToString()
	if strings.Count(got, "\n") != responseMaxLines {
		t.Fatalf("newlines = %d, want %d\n%s", strings.Count(got, "\n"), responseMaxLines, got)
	}
	if !strings.Contains(got, fmt.Sprintf("L%d\n", responseMaxLines+4)) {
		t.Fatalf("missing newest line in %q", got)
	}
	if strings.Contains(got, "L0\n") {
		t.Fatalf("oldest line still present in %q", got)
	}
}

func TestAppendHistoryCapsAtMaxLines(t *testing.T) {
	ui := &appUI{historyWriter: core.NewLineBufferTextWriter(historyMaxLines)}
	for i := 0; i < historyMaxLines+5; i++ {
		ui.appendHistory(fmt.Sprintf("H%d", i))
	}
	got := ui.historyWriter.ToString()
	if strings.Count(got, "\n") != historyMaxLines {
		t.Fatalf("newlines = %d, want %d\n%s", strings.Count(got, "\n"), historyMaxLines, got)
	}
	if !strings.Contains(got, fmt.Sprintf("H%d\n", historyMaxLines+4)) {
		t.Fatalf("missing newest line in %q", got)
	}
	if strings.Contains(got, "H0\n") {
		t.Fatalf("oldest line still present in %q", got)
	}
}

func TestReplayStdoutAndViewSwitch(t *testing.T) {
	ui := &appUI{
		chromeView:     chromeViewREPL,
		responseWriter: core.NewLineBufferTextWriter(8),
		historyWriter:  core.NewLineBufferTextWriter(8),
		stdoutLog:      core.NewLineBufferTextWriterSync(standingMaxLines),
		stderrLog:      core.NewLineBufferTextWriterSync(standingMaxLines),
		response:       tview.NewTextArea(),
		history:        tview.NewTextArea(),
		stdoutPane:     tview.NewTextArea(),
		stderrPane:     tview.NewTextArea(),
		upperPages: tview.NewPages().
			AddPage("repl", tview.NewBox(), true, true).
			AddPage("streams", tview.NewBox(), true, false),
	}
	if standingMaxLines != 160 {
		t.Fatalf("standingMaxLines = %d, want 160", standingMaxLines)
	}
	ui.replayStdout("3\n")
	if got := ui.stdoutLog.ToString(); got != "3\n" {
		t.Fatalf("standing stdout = %q", got)
	}
	ui.appendResponse("resp\n")
	ui.setChromeView(chromeViewStreams)
	if ui.chromeView != chromeViewStreams {
		t.Fatal("expected streams view")
	}
	if got := ui.stdoutPane.GetText(); got != "3\n" {
		t.Fatalf("stdout pane = %q after switch", got)
	}
	ui.setChromeView(chromeViewREPL)
	if got := ui.response.GetText(); !strings.Contains(got, "resp") {
		t.Fatalf("response pane = %q after switch back", got)
	}
}

func TestStdoutNeedsBlitByView(t *testing.T) {
	ui := &appUI{
		stdoutLog:  core.NewLineBufferTextWriterSync(8),
		stdoutPane: tview.NewTextArea(),
		chromeView: chromeViewREPL,
	}
	ui.stdoutLog.WriteString("hello\n")
	if ui.stdoutNeedsBlit() {
		t.Fatal("REPL view should not need stdout blit")
	}
	ui.chromeView = chromeViewStreams
	if !ui.stdoutNeedsBlit() {
		t.Fatal("streams view with new tick should need blit")
	}
	ui.stdoutLastBlit.Store(int64(ui.stdoutLog.GetStringLineBuffer().GetTick()))
	if ui.stdoutNeedsBlit() {
		t.Fatal("matching lastBlit should not need blit")
	}
}

func TestStderrNeedsBlitByView(t *testing.T) {
	ui := &appUI{
		stderrLog:  core.NewLineBufferTextWriterSync(8),
		stderrPane: tview.NewTextArea(),
		chromeView: chromeViewREPL,
	}
	ui.stderrLog.WriteString("err\n")
	if ui.stderrNeedsBlit() {
		t.Fatal("REPL view should not need stderr blit")
	}
	ui.chromeView = chromeViewStreams
	if !ui.stderrNeedsBlit() {
		t.Fatal("streams view with new tick should need blit")
	}
	ui.stderrLastBlit.Store(int64(ui.stderrLog.GetStringLineBuffer().GetTick()))
	if ui.stderrNeedsBlit() {
		t.Fatal("matching lastBlit should not need blit")
	}
}

func TestBlitStdoutUpdatesLastBlit(t *testing.T) {
	ui := &appUI{
		stdoutLog:  core.NewLineBufferTextWriterSync(8),
		stdoutPane: tview.NewTextArea(),
		chromeView: chromeViewStreams,
	}
	ui.stdoutLog.WriteString("tick\n")
	ui.blitStdoutPane(true)
	if got := ui.stdoutPane.GetText(); got != "tick\n" {
		t.Fatalf("pane = %q", got)
	}
	if ui.stdoutNeedsBlit() {
		t.Fatal("after blit, should not need blit")
	}
	ui.stdoutLog.WriteString("more\n")
	if !ui.stdoutNeedsBlit() {
		t.Fatal("new write should need blit")
	}
}

func TestSetTextAreaStickBottomFollowsWhenRequested(t *testing.T) {
	ta := tview.NewTextArea()
	ta.SetRect(0, 0, 40, 10)
	if !setTextAreaStickBottom(ta, "a\nb\n", true) {
		t.Fatal("expected apply")
	}
	if got := ta.GetText(); got != "a\nb\n" {
		t.Fatalf("text = %q", got)
	}
	if !setTextAreaStickBottom(ta, "a\nb\nc\n", true) {
		t.Fatal("expected follow apply")
	}
	if got := ta.GetText(); got != "a\nb\nc\n" {
		t.Fatalf("text = %q", got)
	}
}

func TestSetTextAreaStickBottomSeedsBeforeLayout(t *testing.T) {
	ta := tview.NewTextArea()
	if !setTextAreaStickBottom(ta, "recent\n", true) {
		t.Fatal("expected apply")
	}
	if got := ta.GetText(); got != "recent\n" {
		t.Fatalf("text = %q", got)
	}
}

func TestTextAreaAtBottomCaretVisibility(t *testing.T) {
	ta := tview.NewTextArea()
	ta.SetRect(0, 0, 40, 5)
	ta.SetText("0\n1\n2\n3\n4\n5\n6\n7\n8\n9\n", true)
	if !textAreaAtBottom(ta) {
		t.Fatal("after SetText(true) should be at bottom")
	}
	ta.SetOffset(0, 0)
	if textAreaAtBottom(ta) {
		t.Fatal("scrolled to top with caret at end should not be at bottom")
	}
}

func TestSetChromeViewStreamsBlitsStandingLog(t *testing.T) {
	ui := &appUI{
		chromeView:     chromeViewREPL,
		stdoutLog:      core.NewLineBufferTextWriterSync(standingMaxLines),
		stderrLog:      core.NewLineBufferTextWriterSync(standingMaxLines),
		stdoutPane:     tview.NewTextArea(),
		stderrPane:     tview.NewTextArea(),
		response:       tview.NewTextArea(),
		history:        tview.NewTextArea(),
		responseWriter: core.NewLineBufferTextWriter(8),
		historyWriter:  core.NewLineBufferTextWriter(8),
		upperPages: tview.NewPages().
			AddPage("repl", tview.NewBox(), true, true).
			AddPage("streams", tview.NewBox(), true, false),
	}
	ui.stdoutLog.WriteString("from-send\n")
	ui.setChromeView(chromeViewStreams)
	if got := ui.stdoutPane.GetText(); got != "from-send\n" {
		t.Fatalf("stdout pane after switch = %q, want standing log", got)
	}
}

func TestSetTextAreaStickBottomPreservesOffsetWhenNotFollowing(t *testing.T) {
	ta := tview.NewTextArea()
	ta.SetText("0\n1\n2\n3\n4\n5\n6\n7\n8\n9\n", true)
	ta.SetOffset(0, 0)
	if !setTextAreaStickBottom(ta, "0\n1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n", false) {
		t.Fatal("expected apply")
	}
	gotOff, _ := ta.GetOffset()
	if gotOff != 0 {
		t.Fatalf("offset row = %d, want preserved 0", gotOff)
	}
	if got := ta.GetText(); !strings.Contains(got, "10\n") {
		t.Fatalf("text missing append: %q", got)
	}
}

func TestSetTextAreaStickBottomSkipsWhenSelecting(t *testing.T) {
	ta := tview.NewTextArea()
	ta.SetText("hello\n", true)
	ta.Select(0, 3)
	if setTextAreaStickBottom(ta, "hello\nworld\n", false) {
		t.Fatal("expected skip while selecting when not following")
	}
	if got := ta.GetText(); got != "hello\n" {
		t.Fatalf("text changed while selecting: %q", got)
	}
	if !setTextAreaStickBottom(ta, "hello\nworld\n", true) {
		t.Fatal("expected apply when following")
	}
	if got := ta.GetText(); got != "hello\nworld\n" {
		t.Fatalf("follow should apply: %q", got)
	}
}

func TestStreamFollowDefaultsWithoutApp(t *testing.T) {
	// Nil pane / no layout: treat as at bottom so first blit can follow.
	if !textAreaAtBottom(nil) {
		t.Fatal("nil pane should be at bottom")
	}
}
