package finaltui

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.design/x/clipboard"
)

const (
	pageMain = "main"
	pageOpen = "open"
)

func (ui *appUI) buildOpenPage() tview.Primitive {
	instr := tview.NewTextView().
		SetText("Load a .fnl, .fnlm, or .fnlp file (by suffix).").
		SetWrap(true)

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "(unknown)"
	}
	cwdView := tview.NewTextView().
		SetText("cwd: " + cwd).
		SetWrap(true)

	ui.pathField = tview.NewInputField().
		SetLabel("Path: ").
		SetFieldWidth(0).
		SetPlaceholder("path/to/file.fnl|.fnlm|.fnlp")
	ui.trackEditTarget(ui.pathField)
	ui.pathField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			ui.acceptOpen()
		}
	})
	// InputField has no SetClipboard; Ctrl-V would only hit tview's private buffer.
	// Wire OS clipboard paste when available (terminal bracketed paste works via PasteHandler).
	ui.pathField.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() != tcell.KeyCtrlV || !ui.clipboardOK {
			return ev
		}
		paste := normalizeOpenPath(string(clipboard.Read(clipboard.FmtText)))
		if paste == "" {
			return nil
		}
		// No public cursor API on InputField — append at end (typical for path paste).
		ui.pathField.SetText(normalizeOpenPath(ui.pathField.GetText() + paste))
		return nil
	})

	form := tview.NewForm().
		AddFormItem(ui.pathField).
		AddButton("Load", ui.acceptOpen).
		AddButton("Cancel", ui.dismissOpen)
	form.SetBorder(true).SetTitle(" Load File ")
	form.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			ui.dismissOpen()
			return nil
		}
		return ev
	})

	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(instr, 1, 0, false).
		AddItem(cwdView, 2, 0, false).
		AddItem(form, 0, 1, true)

	// Center a modest dialog over the background.
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(inner, 14, 0, true).
			AddItem(nil, 0, 1, false), 72, 0, true).
		AddItem(nil, 0, 1, false)
}

func (ui *appUI) showOpen() {
	if ui.busy.Load() {
		return
	}
	if name, _ := ui.pages.GetFrontPage(); name == pagePrompt || name == pageFields || name == pageAbout || name == pageBrowse {
		return
	}
	ui.pathField.SetText("")
	ui.pages.SwitchToPage(pageOpen)
	ui.app.SetFocus(ui.pathField)
}

func (ui *appUI) dismissOpen() {
	ui.pages.SwitchToPage(pageMain)
	ui.app.SetFocus(ui.prompt)
}

func (ui *appUI) acceptOpen() {
	if ui.busy.Load() {
		return
	}
	path := normalizeOpenPath(ui.pathField.GetText())
	ui.pathField.SetText(path)
	if path == "" {
		return
	}
	script, err := compilePathScript(path)
	if err != nil {
		ui.dismissOpen()
		ui.appendResponse(fmt.Sprintf("Error: %v\n", err))
		return
	}
	ui.dismissOpen()
	ui.sendScript(script, false)
}

// normalizeOpenPath trims whitespace and strips a single pair of surrounding
// quotes (Windows Explorer "Copy as path" yields `"C:\…\file.fnl"`).
func normalizeOpenPath(path string) string {
	path = strings.TrimSpace(strings.Trim(path, "\r\n"))
	if len(path) < 2 {
		return path
	}
	if (path[0] == '"' && path[len(path)-1] == '"') ||
		(path[0] == '\'' && path[len(path)-1] == '\'') {
		path = strings.TrimSpace(path[1 : len(path)-1])
	}
	return path
}
