package finaltui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/androticus/final/finalgo"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.design/x/clipboard"
)

const pageAbout = "about"

var (
	menuItemStyle = tcell.StyleDefault.
			Foreground(tcell.ColorBlack).
			Background(tcell.ColorLightGray)
	menuSelectedStyle = tcell.StyleDefault.
				Foreground(tcell.ColorWhite).
				Background(tcell.ColorBlue)
)

// buildMenuBar returns File / Edit / Run / View / Help dropdowns (tview has no MenuBar primitive).
func (ui *appUI) buildMenuBar() *tview.Flex {
	file := ui.newMenuDropDown("File", 10,
		menuItem{label: "Load…", shortcut: "^L", action: ui.showBrowseLoad},
		menuSeparator,
		menuItem{label: "Quit", shortcut: "^Q", action: ui.requestQuit},
	)
	edit := ui.newMenuDropDown("Edit", 8,
		menuItem{label: "Cut", shortcut: "^X", action: ui.editCut},
		menuItem{label: "Copy", shortcut: "^C", action: ui.editCopy},
		menuItem{label: "Paste", shortcut: "^V", action: ui.editPaste},
	)
	run := ui.newMenuDropDown("Run", 8,
		menuItem{label: "Break", shortcut: "^B", action: ui.requestBreak},
	)
	ui.viewDD = ui.newMenuDropDown("View", 18, ui.viewMenuItems()...)
	help := ui.newMenuDropDown("Help", 8,
		menuItem{label: "Words…", action: ui.helpWords},
		menuSeparator,
		menuItem{label: "About…", action: ui.showAbout},
	)

	bar := tview.NewFlex().
		AddItem(file, 12, 0, false).
		AddItem(edit, 10, 0, false).
		AddItem(run, 10, 0, false).
		AddItem(ui.viewDD, 20, 0, false).
		AddItem(help, 10, 0, false).
		AddItem(nil, 0, 1, false)
	return bar
}

func (ui *appUI) viewMenuItems() []menuItem {
	// Active view is a non-action row (tview DropDown has no Disabled); other switches.
	repl := menuItem{label: "Prompt/History", action: func() { ui.setChromeView(chromeViewREPL) }}
	streams := menuItem{label: "Stdout/Stderr", action: func() { ui.setChromeView(chromeViewStreams) }}
	if ui.chromeView == chromeViewREPL {
		repl.separator = true
		repl.action = nil
	} else {
		streams.separator = true
		streams.action = nil
	}
	return []menuItem{repl, streams}
}

func (ui *appUI) refreshViewMenu() {
	if ui.viewDD == nil {
		return
	}
	ui.viewDD.SetOptions([]string{}, nil)
	items := ui.viewMenuItems()
	labelWidth := menuShortcutLabelWidth(items)
	for _, it := range items {
		it := it
		ui.viewDD.AddOption(menuItemText(it, labelWidth), func() {
			if it.separator || it.action == nil {
				return
			}
			it.action()
		})
	}
	ui.viewDD.SetCurrentOption(-1)
}

// menuSeparator is a non-action row (DropDown has no real separator primitive).
var menuSeparator = menuItem{label: "────────", separator: true}

type menuItem struct {
	label     string
	shortcut  string // e.g. "^L"; shown after two spaces (DropDown has no right-align API)
	separator bool
	action    func()
}

// menuItemText pads labels (among shortcut rows) so shortcuts column-align in monospace.
// Width is in runes so "Load…" (ellipsis) matches visible columns, not UTF-8 byte length.
func menuItemText(it menuItem, labelWidth int) string {
	if it.separator || it.shortcut == "" {
		return it.label
	}
	label := it.label
	if pad := labelWidth - utf8.RuneCountInString(label); pad > 0 {
		label += strings.Repeat(" ", pad)
	}
	return label + "  " + it.shortcut
}

func menuShortcutLabelWidth(items []menuItem) int {
	w := 0
	for _, it := range items {
		if it.separator || it.shortcut == "" {
			continue
		}
		if n := utf8.RuneCountInString(it.label); n > w {
			w = n
		}
	}
	return w
}

func (ui *appUI) helpWords() {
	ui.sendScript("list", false)
}

func (ui *appUI) newMenuDropDown(title string, fieldWidth int, items ...menuItem) *tview.DropDown {
	dd := tview.NewDropDown().
		SetLabel("").
		SetFieldWidth(fieldWidth).
		SetTextOptions("", "", "", "", title).
		SetListStyles(menuItemStyle, menuSelectedStyle).
		SetFieldStyle(menuItemStyle).
		SetFocusedStyle(menuSelectedStyle)

	// tview List always has a current index (opens on item 0). Hide the highlight
	// until the user presses Down/Up, then show selection on the first item.
	pending := true
	menuListPending(dd, true)

	labelWidth := menuShortcutLabelWidth(items)
	for _, it := range items {
		it := it
		dd.AddOption(menuItemText(it, labelWidth), func() {
			dd.SetCurrentOption(-1)
			pending = true
			menuListPending(dd, true)
			if it.separator || it.action == nil {
				return
			}
			it.action()
		})
	}

	wasOpen := false
	dd.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		open := dd.IsOpen()
		if open && !wasOpen {
			// Keyboard or mouse just opened — start with no visible selection.
			pending = true
			menuListPending(dd, true)
		}
		if !open && wasOpen {
			pending = true
			menuListPending(dd, true)
		}
		wasOpen = open
		return x, y, width, height
	})

	dd.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if !dd.IsOpen() {
			switch ev.Key() {
			case tcell.KeyDown:
				// Open without advancing past item 0 (DropDown would forward Down → item 1).
				pending = true
				menuListPending(dd, true)
				return tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
			case tcell.KeyEnter:
				pending = true
				menuListPending(dd, true)
				return ev
			default:
				return ev
			}
		}

		if pending {
			switch ev.Key() {
			case tcell.KeyDown:
				pending = false
				menuListPending(dd, false)
				return nil // stay on first item, now visibly selected
			case tcell.KeyUp:
				pending = false
				menuListPending(dd, false)
				return nil // reveal highlight on first item
			case tcell.KeyEnter:
				pending = false
				menuListPending(dd, false)
				return ev // activate first item
			case tcell.KeyEscape:
				pending = true
				menuListPending(dd, true)
				return ev
			}
		}
		return ev
	})
	return dd
}

// menuListPending makes the list highlight invisible (pending) or normal (armed).
func menuListPending(dd *tview.DropDown, pending bool) {
	if pending {
		dd.SetListStyles(menuItemStyle, menuItemStyle)
		return
	}
	dd.SetListStyles(menuItemStyle, menuSelectedStyle)
}

func (ui *appUI) buildAboutPage() tview.Primitive {
	body := fmt.Sprintf("Final version: %s\n%s", finalgo.Version(), finalgo.Copyright())
	tv := tview.NewTextView().
		SetText(body).
		SetTextAlign(tview.AlignLeft).
		SetWrap(true)
	tv.SetBorder(true).SetTitle(" About ")

	ui.aboutOk = tview.NewButton("Ok").SetSelectedFunc(ui.dismissAbout)
	ui.aboutOk.SetStyle(sendStyle).SetActivatedStyle(sendStyle)

	form := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tv, 0, 1, false).
		AddItem(ui.aboutOk, 1, 0, true)
	form.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyEnter {
			ui.dismissAbout()
			return nil
		}
		return ev
	})

	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 8, 0, true).
			AddItem(nil, 0, 1, false), 50, 0, true).
		AddItem(nil, 0, 1, false)
}

func (ui *appUI) showAbout() {
	if name, _ := ui.pages.GetFrontPage(); name == pagePrompt || name == pageFields || name == pageOpen || name == pageBrowse {
		return
	}
	ui.pages.SwitchToPage(pageAbout)
	ui.app.SetFocus(ui.aboutOk)
}

func (ui *appUI) dismissAbout() {
	ui.pages.SwitchToPage(pageMain)
	ui.app.SetFocus(ui.prompt)
}

// editCopy copies the selection (or focused InputField text) to the OS clipboard.
func (ui *appUI) editCopy() {
	if !ui.clipboardOK {
		ui.focusAfterEditMenu()
		return
	}
	switch f := ui.editTarget.(type) {
	case *tview.TextArea:
		text, start, end := f.GetSelection()
		if start == end {
			return
		}
		clipboard.Write(clipboard.FmtText, []byte(text))
	case *tview.InputField:
		text := f.GetText()
		if text == "" {
			return
		}
		clipboard.Write(clipboard.FmtText, []byte(text))
	}
	ui.focusAfterEditMenu()
}

// editCut cuts from editable targets only (prompt / InputField; not read-only panes).
func (ui *appUI) editCut() {
	if !ui.clipboardOK || ui.busy.Load() {
		ui.focusAfterEditMenu()
		return
	}
	switch f := ui.editTarget.(type) {
	case *tview.TextArea:
		if ui.isReadOnlyPane(f) {
			return
		}
		text, start, end := f.GetSelection()
		if start == end {
			return
		}
		clipboard.Write(clipboard.FmtText, []byte(text))
		f.Replace(start, end, "")
	case *tview.InputField:
		text := f.GetText()
		if text == "" {
			return
		}
		clipboard.Write(clipboard.FmtText, []byte(text))
		f.SetText("")
	}
	ui.focusAfterEditMenu()
}

// editPaste pastes into editable targets only.
func (ui *appUI) editPaste() {
	if !ui.clipboardOK || ui.busy.Load() {
		ui.focusAfterEditMenu()
		return
	}
	paste := string(clipboard.Read(clipboard.FmtText))
	if paste == "" {
		return
	}
	switch f := ui.editTarget.(type) {
	case *tview.TextArea:
		if ui.isReadOnlyPane(f) {
			return
		}
		_, start, end := f.GetSelection()
		f.Replace(start, end, paste)
		scrollCursorIntoView(f)
	case *tview.InputField:
		if f == ui.pathField {
			paste = normalizeOpenPath(paste)
		}
		f.SetText(f.GetText() + paste)
	}
	ui.focusAfterEditMenu()
}

func (ui *appUI) focusAfterEditMenu() {
	name, _ := ui.pages.GetFrontPage()
	switch name {
	case pageMain, pageOpen, pageFields:
		if ui.editTarget != nil {
			ui.app.SetFocus(ui.editTarget)
			return
		}
		ui.app.SetFocus(ui.prompt)
	}
}
