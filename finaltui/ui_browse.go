package finaltui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rivo/tview"
)

const pageBrowse = "browse"

func (ui *appUI) buildBrowsePage() tview.Primitive {
	b := ui.newLoadBrowser()
	ui.browser = b
	return b.Root
}

func (ui *appUI) showBrowseLoad() {
	if ui.busy.Load() {
		return
	}
	if name, _ := ui.pages.GetFrontPage(); name == pagePrompt || name == pageFields || name == pageAbout || name == pageOpen {
		return
	}
	b := ui.newLoadBrowser()
	ui.browser = b
	ui.pages.RemovePage(pageBrowse)
	ui.pages.AddPage(pageBrowse, b.Root, true, false)
	ui.pages.SwitchToPage(pageBrowse)
	ui.app.SetFocus(b.FocusPrimitive())
}

func (ui *appUI) newLoadBrowser() *Browser {
	b := NewBrowser(ui.chooserStartDir(), BrowserChrome{
		Title:       " Load File ",
		Instruction: "Load a .fnl, .fnlm, or .fnlp file (by suffix).",
		OpenLabel:   "Load",
	})
	b.OnCancel = ui.dismissBrowse
	b.OnOpenFile = ui.acceptBrowseLoad
	return b
}

func (ui *appUI) chooserStartDir() string {
	if ui.lastChooserPath != "" {
		dir := filepath.Dir(ui.lastChooserPath)
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			return dir
		}
	}
	start, _ := os.Getwd()
	return start
}

func (ui *appUI) dismissBrowse() {
	ui.pages.SwitchToPage(pageMain)
	ui.app.SetFocus(ui.prompt)
}

func (ui *appUI) acceptBrowseLoad(path string) {
	if ui.busy.Load() {
		return
	}
	script, err := compilePathScript(path)
	if err != nil {
		ui.dismissBrowse()
		ui.appendResponse(fmt.Sprintf("Error: %v\n", err))
		return
	}
	ui.lastChooserPath = path
	ui.dismissBrowse()
	ui.sendScript(script, false)
}
