package finaltui

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.design/x/clipboard"
)

// BrowserChrome configures title, instruction body, and primary action label.
type BrowserChrome struct {
	Title       string // e.g. " Load File "; empty → " Load File "
	Instruction string // optional line above the path row
	OpenLabel   string // e.g. "Load"; empty → "Load"
}

// Browser is a two-pane on-demand file browser (folders | files).
type Browser struct {
	Root tview.Primitive

	pathField *tview.InputField
	filter    *tview.DropDown
	folders   *navList
	files     *navList
	status    *tview.TextView

	dir       string // absolute current directory; "" = Windows drive picker
	finalOnly bool   // when true, Files pane shows only *.fnl / *.fnlm / *.fnlp

	OnOpenFile func(path string)
	OnCancel   func()
}

// NewBrowser builds the Browse UI starting at startDir (empty = cwd).
func NewBrowser(startDir string, chrome BrowserChrome) *Browser {
	if startDir == "" {
		startDir, _ = os.Getwd()
	}
	abs, err := filepath.Abs(startDir)
	if err != nil {
		abs = startDir
	}
	title := chrome.Title
	if title == "" {
		title = " Load File "
	}
	openLabel := chrome.OpenLabel
	if openLabel == "" {
		openLabel = "Load"
	}

	b := &Browser{dir: filepath.Clean(abs), finalOnly: true}

	b.pathField = tview.NewInputField().
		SetLabel("Path: ").
		SetFieldWidth(0)
	b.pathField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			b.commitPathField()
		}
	})
	b.pathField.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyCtrlV {
			paste := string(clipboard.Read(clipboard.FmtText))
			if paste != "" {
				b.pathField.SetText(b.pathField.GetText() + paste)
			}
			return nil
		}
		return ev
	})

	b.folders = newNavList(" Folders ")
	b.folders.onActivate = b.activateFolder
	b.folders.activateOnClick = true
	b.files = newNavList(" Files ")
	b.files.onActivate = b.activateFile
	b.status = tview.NewTextView().SetDynamicColors(true)

	// After folders/files exist: SetCurrentOption invokes the option callback → reload.
	b.filter = tview.NewDropDown().
		SetLabel("Show: ").
		SetFieldWidth(36).
		SetOptions([]string{"Final *.fnl *.fnlm *.fnlp", "All files"}, func(text string, index int) {
			b.finalOnly = index == 0
			b.reload()
		})
	b.filter.SetCurrentOption(0)

	copyBtn := tview.NewButton("Copy path").SetSelectedFunc(b.copyPath)
	openBtn := tview.NewButton(openLabel).SetSelectedFunc(b.openSelected)
	cancelBtn := tview.NewButton("Cancel").SetSelectedFunc(func() {
		if b.OnCancel != nil {
			b.OnCancel()
		}
	})
	openBtnWidth := len(openLabel) + 4
	if openBtnWidth < 10 {
		openBtnWidth = 10
	}
	if openBtnWidth > 20 {
		openBtnWidth = 20
	}

	pathRow := tview.NewFlex().
		AddItem(b.pathField, 0, 1, true).
		AddItem(b.filter, 40, 0, false).
		AddItem(copyBtn, 12, 0, false)

	panes := tview.NewFlex().
		AddItem(b.folders, 0, 1, false).
		AddItem(b.files, 0, 1, false)

	buttons := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(openBtn, openBtnWidth, 0, false).
		AddItem(nil, 2, 0, false).
		AddItem(cancelBtn, 10, 0, false).
		AddItem(nil, 0, 1, false)

	inner := tview.NewFlex().SetDirection(tview.FlexRow)
	if chrome.Instruction != "" {
		instr := tview.NewTextView().
			SetText(chrome.Instruction).
			SetWrap(true)
		inner.AddItem(instr, 1, 0, false)
	}
	inner.
		AddItem(pathRow, 1, 0, true).
		AddItem(panes, 0, 1, false).
		AddItem(b.status, 1, 0, false).
		AddItem(buttons, 1, 0, false)
	inner.SetBorder(true).SetTitle(title)
	inner.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			if b.OnCancel != nil {
				b.OnCancel()
			}
			return nil
		}
		return ev
	})

	b.Root = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(inner, 0, 4, true).
			AddItem(nil, 0, 1, false), 90, 0, true).
		AddItem(nil, 0, 1, false)

	b.reload()
	return b
}

// FocusPrimitive returns the initial focus target (folder list).
func (b *Browser) FocusPrimitive() tview.Primitive {
	return b.folders
}

func (b *Browser) setStatus(msg string) {
	b.status.SetText(msg)
}

func (b *Browser) reload() {
	b.folders.Clear()
	b.files.Clear()

	if b.dir == "" {
		b.pathField.SetText("")
		for _, r := range WindowsDriveRoots() {
			b.folders.AddItem(r, "", 0, nil)
		}
		b.setStatus("Select a drive")
		return
	}

	b.pathField.SetText(b.dir)
	listing := ListDir(b.dir)
	if listing.Err != nil {
		b.setStatus("[red]Error:[-] " + listing.Err.Error())
		if !IsFSRoot(b.dir) || runtime.GOOS == "windows" {
			b.folders.AddItem("..", "", 0, nil)
		}
		return
	}

	if !IsFSRoot(b.dir) || runtime.GOOS == "windows" {
		b.folders.AddItem("..", "", 0, nil)
	}
	for _, name := range listing.Folders {
		b.folders.AddItem(name+"/", "", 0, nil)
	}
	for _, name := range listing.Files {
		if !showBrowseFile(name, b.finalOnly) {
			continue
		}
		b.files.AddItem(name, "", 0, nil)
	}
	b.setStatus("")
}

// showBrowseFile reports whether name belongs in the Files pane for the filter.
// When finalOnly, only suffix-distinguished Final artifacts are shown (.fnl / .fnlm / .fnlp).
func showBrowseFile(name string, finalOnly bool) bool {
	if !finalOnly {
		return true
	}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".fnl", ".fnlm", ".fnlp":
		return true
	default:
		return false
	}
}

func (b *Browser) activateFolder(_ int, main string) {
	if main == ".." {
		b.goUp()
		return
	}
	name := trimTrailingSep(main)
	if runtime.GOOS == "windows" && looksLikeDriveRoot(name) {
		b.dir = filepath.Clean(name)
		b.reload()
		return
	}
	if b.dir == "" {
		return
	}
	b.dir = JoinUnder(b.dir, name)
	b.reload()
}

func (b *Browser) goUp() {
	if b.dir == "" {
		return
	}
	if IsFSRoot(b.dir) {
		if runtime.GOOS == "windows" {
			b.dir = ""
			b.reload()
		}
		return
	}
	b.dir = ParentDir(b.dir)
	b.reload()
}

func (b *Browser) activateFile(_ int, main string) {
	if b.dir == "" || main == "" {
		return
	}
	path := JoinUnder(b.dir, main)
	if b.OnOpenFile != nil {
		b.OnOpenFile(path)
	}
}

func (b *Browser) openSelected() {
	// Browse is for opening a file: always prefer the Files pane current
	// item. (Clicking the button steals focus, so HasFocus is unreliable.)
	if b.files.GetItemCount() > 0 {
		idx := b.files.GetCurrentItem()
		if idx >= 0 {
			main, _ := b.files.GetItemText(idx)
			if main != "" {
				b.activateFile(idx, main)
				return
			}
		}
	}
	b.setStatus("[yellow]No file selected[-]")
}

func (b *Browser) copyPath() {
	path := b.selectedPath()
	if path == "" {
		b.setStatus("[yellow]Nothing to copy[-]")
		return
	}
	clipboard.Write(clipboard.FmtText, []byte(path))
	b.setStatus("Copied: " + path)
}

func (b *Browser) selectedPath() string {
	if b.dir == "" {
		if b.folders.GetItemCount() == 0 {
			return ""
		}
		main, _ := b.folders.GetItemText(b.folders.GetCurrentItem())
		return main
	}
	if b.files.HasFocus() && b.files.GetItemCount() > 0 {
		main, _ := b.files.GetItemText(b.files.GetCurrentItem())
		if main != "" {
			return JoinUnder(b.dir, main)
		}
	}
	if b.folders.HasFocus() && b.folders.GetItemCount() > 0 {
		main, _ := b.folders.GetItemText(b.folders.GetCurrentItem())
		if main == ".." {
			return ParentDir(b.dir)
		}
		name := trimTrailingSep(main)
		if name != "" {
			return JoinUnder(b.dir, name)
		}
	}
	return b.dir
}

func (b *Browser) commitPathField() {
	dir, file, err := ResolveBrowsePath(b.pathField.GetText())
	if err != nil {
		b.setStatus("[red]Error:[-] " + err.Error())
		return
	}
	b.dir = dir
	b.reload()
	if file != "" {
		for i := 0; i < b.files.GetItemCount(); i++ {
			main, _ := b.files.GetItemText(i)
			if main == file {
				b.files.SetCurrentItem(i)
				break
			}
		}
	}
}

func trimTrailingSep(s string) string {
	return strings.TrimRight(s, `/\`)
}

func looksLikeDriveRoot(s string) bool {
	// C: or C:\
	if len(s) >= 2 && s[1] == ':' {
		return true
	}
	return false
}

// navList: single-click selects; Enter / double-click activates.
// When activateOnClick (folder pane), single-click also activates; double-click
// is ignored so a physical double-click does not descend twice (tview fires
// Click or DoubleClick, not both).
type navList struct {
	*tview.List
	onActivate      func(index int, mainText string)
	activateOnClick bool
}

func (l *navList) activateIndex(index int) {
	if index < 0 || index >= l.GetItemCount() || l.onActivate == nil {
		return
	}
	main, _ := l.GetItemText(index)
	l.onActivate(index, main)
}

func newNavList(title string) *navList {
	l := &navList{List: tview.NewList().ShowSecondaryText(false)}
	l.SetBorder(true).SetTitle(title)
	l.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEnter {
			l.activateIndex(l.GetCurrentItem())
			return nil
		}
		return ev
	})
	return l
}

func (l *navList) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
	return l.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !l.InRect(event.Position()) {
			return false, nil
		}
		switch action {
		case tview.MouseLeftClick:
			setFocus(l)
			if index := mouseIndexAt(l.List, event); index >= 0 {
				l.SetCurrentItem(index)
				if l.activateOnClick {
					l.activateIndex(index)
				}
			}
			return true, nil
		case tview.MouseLeftDoubleClick:
			setFocus(l)
			if index := mouseIndexAt(l.List, event); index >= 0 {
				l.SetCurrentItem(index)
				if !l.activateOnClick {
					l.activateIndex(index)
				}
			}
			return true, nil
		case tview.MouseScrollUp:
			// List clamps offset to keep the current item visible, so
			// offset-only scroll stops at the selection. Move selection instead.
			idx := l.GetCurrentItem()
			if idx > 0 {
				l.SetCurrentItem(idx - 1)
			}
			return true, nil
		case tview.MouseScrollDown:
			idx := l.GetCurrentItem()
			if idx < l.GetItemCount()-1 {
				l.SetCurrentItem(idx + 1)
			}
			return true, nil
		}
		return false, nil
	})
}

func mouseIndexAt(list *tview.List, event *tcell.EventMouse) int {
	_, y := event.Position()
	_, rectY, _, height := list.GetInnerRect()
	if height <= 0 {
		return -1
	}
	row := y - rectY
	off, _ := list.GetOffset()
	index := row + off
	if index < 0 || index >= list.GetItemCount() {
		return -1
	}
	return index
}
