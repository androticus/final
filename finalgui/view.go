package finalgui

import (
	"fyne.io/fyne/v2"
)

// chromeView is which upper dual pane is showing (Prompt stays in both).
type chromeView int

const (
	chromeViewREPL chromeView = iota // Response | History — menu "Prompt/History"
	chromeViewStreams                // Stdout | Stderr — menu "Stdout/Stderr"
)

// setChromeView swaps the upper dual pane. Prompt + Send stay. UI thread only.
func (g *FinalGui) setChromeView(v chromeView) {
	if g.upperHost == nil {
		g.chromeView = v
		g.syncViewMenuItems()
		return
	}
	if g.chromeView == v {
		g.syncViewMenuItems()
		return
	}
	g.chromeView = v
	var content fyne.CanvasObject
	switch v {
	case chromeViewStreams:
		content = g.streamPanes
		g.blitStdoutPane()
		g.blitStderrPane()
	default:
		content = g.replPanes
		// Refresh Response/History from standing logs (they keep appending while hidden).
		if g.response != nil && g.responseLog != nil {
			setReadOnlyText(g.response, g.responseLog.ToString())
		}
		if g.history != nil && g.historyLog != nil {
			setReadOnlyText(g.history, g.historyLog.ToString())
		}
	}
	g.upperHost.Objects = []fyne.CanvasObject{content}
	g.upperHost.Refresh()
	g.syncViewMenuItems()
}

func (g *FinalGui) syncViewMenuItems() {
	if g.viewREPLItem != nil {
		g.viewREPLItem.Disabled = g.chromeView == chromeViewREPL
	}
	if g.viewStreamsItem != nil {
		g.viewStreamsItem.Disabled = g.chromeView == chromeViewStreams
	}
}

func requestViewREPL() {
	FinalGui_.setChromeView(chromeViewREPL)
}

func requestViewStreams() {
	FinalGui_.setChromeView(chromeViewStreams)
}
