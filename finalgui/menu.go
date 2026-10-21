package finalgui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/androticus/final/internal/core"
	"github.com/androticus/final/internal/fyneui"
)

// buildShellMenus returns File / Edit / Run / View / Help.
// Load opens the file dialog on the UI thread, then enqueues onLoad.
// Break calls Session.Break on the UI thread (must not wait on EventLoop).
// Cut/Copy/Paste also run on the UI thread (clipboard; do not enqueue).
// View items swap the upper dual pane on the UI thread.
// Other items enqueue FinalGui MethodRefs.
func buildShellMenus(w fyne.Window) []*fyne.Menu {
	el := core.GetEventLoop()
	enqueue := func(name string) func() {
		fn := el.CreateEventFn(methodRef(name), 0)
		return func() { fn(nil) }
	}

	load := fyne.NewMenuItem("Load…", openLoadDialog)
	assignDefaultShortcut(load, fyne.KeyL)
	bindWindowShortcut(w, load)

	quit := fyne.NewMenuItem("Quit", requestQuit)
	assignDefaultShortcut(quit, fyne.KeyQ)
	bindWindowShortcut(w, quit)

	// Cut/Copy/Paste run on the UI thread (same as Break). Enqueue → EventLoop →
	// fyne.Do(TypedShortcut) deadlocks on macOS: clipboard read needs the run
	// loop, which is blocked in the deferred Do.
	cutItem := fyne.NewMenuItem("Cut", requestCut)
	cutItem.Shortcut = &fyne.ShortcutCut{}
	copyItem := fyne.NewMenuItem("Copy", requestCopy)
	copyItem.Shortcut = &fyne.ShortcutCopy{}
	pasteItem := fyne.NewMenuItem("Paste", requestPaste)
	pasteItem.Shortcut = &fyne.ShortcutPaste{}

	brk := fyne.NewMenuItem("Break", requestBreak)
	assignDefaultShortcut(brk, fyne.KeyB)
	bindWindowShortcut(w, brk)

	g := FinalGui_
	g.viewREPLItem = fyne.NewMenuItem("Prompt/History", requestViewREPL)
	g.viewStreamsItem = fyne.NewMenuItem("Stdout/Stderr", requestViewStreams)
	g.syncViewMenuItems()
	scaleItem := fyne.NewMenuItem("Scale…", openScaleDialog)

	file := fyne.NewMenu("File",
		load,
		fyne.NewMenuItemSeparator(),
		quit,
	)
	edit := fyne.NewMenu("Edit", cutItem, copyItem, pasteItem)
	run := fyne.NewMenu("Run", brk)
	view := fyne.NewMenu("View", g.viewREPLItem, g.viewStreamsItem, fyne.NewMenuItemSeparator(), scaleItem)
	words := fyne.NewMenuItem("Words…", func() {
		if FinalGui_.busy.Load() {
			return
		}
		FinalGui_.markBusy()
		enqueue("onWords")()
	})
	help := fyne.NewMenu("Help",
		words,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("About…", enqueue("onAbout")),
	)
	return []*fyne.Menu{file, edit, run, view, help}
}

// buildMenuBar is the Phase 0 stand-in for Window.setMenuBar: in-window MenuBar
// from the same menus Final will pass to MenuBar.new^.
func buildMenuBar(w fyne.Window) *fyneui.MenuBar {
	menus := buildShellMenus(w)
	list := core.NewAnyListWithCapacity(len(menus))
	for _, m := range menus {
		list.Append(m)
	}
	return fyneui.NewMenuBar(list)
}

// requestQuit enqueues FinalGui#onQuit (File ▸ Quit and darwin native Quit / Cmd-Q).
func requestQuit() {
	fn := core.GetEventLoop().CreateEventFn(methodRef("onQuit"), 0)
	fn(nil)
}

func assignDefaultShortcut(item *fyne.MenuItem, key fyne.KeyName) {
	item.Shortcut = &desktop.CustomShortcut{
		KeyName:  key,
		Modifier: fyne.KeyModifierShortcutDefault,
	}
}

// bindWindowShortcut registers a canvas shortcut so the combo fires while the
// prompt Entry is focused. Skip Cut/Copy/Paste — requestCut/Copy/Paste from the
// menu item; Entry TypedShortcut handles the chord (a canvas bind would double-fire).
func bindWindowShortcut(w fyne.Window, item *fyne.MenuItem) {
	if w == nil || item == nil || item.Shortcut == nil || item.Action == nil {
		return
	}
	action := item.Action
	w.Canvas().AddShortcut(item.Shortcut, func(fyne.Shortcut) { action() })
}

// requestBreak arms Session.Break from the UI thread. Do not enqueue through
// EventLoop: the worker may be inside interpret, which is what Break interrupts.
func requestBreak() {
	if g := FinalGui_; g.sess != nil {
		g.sess.Break()
	}
}

func requestCut()   { applyEditShortcut(&fyne.ShortcutCut{}) }
func requestCopy()  { applyEditShortcut(&fyne.ShortcutCopy{}) }
func requestPaste() { applyEditShortcut(&fyne.ShortcutPaste{}) }

// applyEditShortcut dispatches Cut/Copy/Paste on the UI thread to the focused
// widget (prompt, or Response/History/Stdout/Stderr for Copy). Must not be fyne.Do'd from
// EventLoop — glfw clipboard on darwin needs the nested run loop of pollEvents.
func applyEditShortcut(s fyne.Shortcut) {
	g := FinalGui_
	if g.win == nil {
		return
	}
	app := fyne.CurrentApp()
	if app == nil {
		return
	}
	clip := app.Clipboard()
	switch t := s.(type) {
	case *fyne.ShortcutCut:
		t.Clipboard = clip
	case *fyne.ShortcutCopy:
		t.Clipboard = clip
	case *fyne.ShortcutPaste:
		t.Clipboard = clip
	}

	target := g.win.Canvas().Focused()
	if target == nil {
		target = g.prompt
	}
	if target == nil {
		return
	}
	if d, ok := target.(fyne.Disableable); ok && d.Disabled() {
		if _, isCopy := s.(*fyne.ShortcutCopy); !isCopy {
			return
		}
	}
	if sc, ok := target.(fyne.Shortcutable); ok {
		sc.TypedShortcut(s)
		return
	}
	if g.prompt != nil {
		g.prompt.TypedShortcut(s)
	}
}
