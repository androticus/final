package finalgui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/internal/core"
	"github.com/androticus/final/internal/fyneui"
)

// Run starts EventLoop#run as the session worker root. Blocks until the event pipe
// closes. Call from a worker goroutine after BuildChrome; OS main must be in TheApp.Run.
func Run(sess *finalembed.Session) error {
	if sess == nil {
		return fmt.Errorf("finalgui.Run: nil session")
	}
	if g := FinalGui_; g.stdoutLog != nil {
		sess.SetStdout(g.stdoutLog)
	}
	if g := FinalGui_; g.stderrLog != nil {
		sess.SetStderr(g.stderrLog)
	}
	core.RunFaultReportHook = func(fault *core.RunFault) {
		FinalGui_.appendResponse("", fault)
	}
	defer func() { core.RunFaultReportHook = nil }()
	return sess.RunEventLoop(guiRootWord)
}

// Stop requests EventLoop#run to exit (closed flag + wakeup). Idempotent.
// Does not Close Fyne windows — TheApp.Quit (UI thread) tears them down.
func Stop() {
	core.RunFaultReportHook = nil
	FinalGui_.stopStandingPoller()
	core.GetEventLoop().Close()
}

// Quit stops the EventLoop worker and requests Fyne App.Quit (intentional exit).
func Quit() {
	Stop()
	fyne.Do(func() {
		fyneui.TheApp.Quit()
	})
}

// Copyright returns the product copyright line (About dialog).
func Copyright() string {
	return core.Copyright
}
