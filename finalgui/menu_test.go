package finalgui

import (
	"testing"

	"fyne.io/fyne/v2"
	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/internal/core"
)

func TestBuildShellMenusShortcuts(t *testing.T) {
	core.ResetTestRuntime()
	sess, err := finalembed.NewPersistentSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	if err := Init(sess); err != nil {
		t.Fatalf("Init: %v", err)
	}

	menus := buildShellMenus(nil)
	if got, want := len(menus), 5; got != want {
		t.Fatalf("menu count = %d, want %d", got, want)
	}
	labels := []string{"File", "Edit", "Run", "View", "Help"}
	for i, want := range labels {
		if menus[i].Label != want {
			t.Fatalf("menu[%d] = %q, want %q", i, menus[i].Label, want)
		}
	}

	file := menus[0]
	load := file.Items[0]
	if load.Label != "Load…" {
		t.Fatalf("File[0] = %q, want Load…", load.Label)
	}
	assertShortcutKey(t, "Load", load.Shortcut, fyne.KeyL)

	quit := file.Items[len(file.Items)-1]
	if quit.Label != "Quit" {
		t.Fatalf("File last = %q, want Quit", quit.Label)
	}
	assertShortcutKey(t, "Quit", quit.Shortcut, fyne.KeyQ)

	edit := menus[1]
	assertShortcutKey(t, "Cut", edit.Items[0].Shortcut, fyne.KeyX)
	assertShortcutKey(t, "Copy", edit.Items[1].Shortcut, fyne.KeyC)
	assertShortcutKey(t, "Paste", edit.Items[2].Shortcut, fyne.KeyV)

	run := menus[2]
	if len(run.Items) != 1 || run.Items[0].Label != "Break" {
		t.Fatalf("Run = %+v, want [Break]", itemLabels(run.Items))
	}
	assertShortcutKey(t, "Break", run.Items[0].Shortcut, fyne.KeyB)

	view := menus[3]
	if len(view.Items) != 4 ||
		view.Items[0].Label != "Prompt/History" ||
		view.Items[1].Label != "Stdout/Stderr" ||
		view.Items[3].Label != "Scale…" {
		t.Fatalf("View = %+v, want [Prompt/History Stdout/Stderr — Scale…]", itemLabels(view.Items))
	}
	if !view.Items[0].Disabled || view.Items[1].Disabled {
		t.Fatal("default View: Prompt/History disabled, Stdout/Stderr enabled")
	}

	// Edit actions must be UI-thread (nil window is a no-op, not an EventLoop send).
	edit.Items[0].Action()
	edit.Items[1].Action()
	edit.Items[2].Action()

	if buildMenuBar(nil) == nil {
		t.Fatal("buildMenuBar returned nil")
	}
}

func TestRequestEditShortcutNilWindow(t *testing.T) {
	prevWin, prevPrompt := FinalGui_.win, FinalGui_.prompt
	FinalGui_.win, FinalGui_.prompt = nil, nil
	t.Cleanup(func() {
		FinalGui_.win, FinalGui_.prompt = prevWin, prevPrompt
	})
	requestCut()
	requestCopy()
	requestPaste()
}

func TestRequestBreakIdle(t *testing.T) {
	core.ResetTestRuntime()
	sess, err := finalembed.NewPersistentSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	if err := Init(sess); err != nil {
		t.Fatalf("Init: %v", err)
	}
	requestBreak() // idle Session.Break is a no-op
}

func assertShortcutKey(t *testing.T, name string, s fyne.Shortcut, want fyne.KeyName) {
	t.Helper()
	ks, ok := s.(fyne.KeyboardShortcut)
	if !ok {
		t.Fatalf("%s shortcut type %T, want KeyboardShortcut", name, s)
	}
	if ks.Key() != want {
		t.Fatalf("%s key = %q, want %q", name, ks.Key(), want)
	}
	if ks.Mod() != fyne.KeyModifierShortcutDefault {
		t.Fatalf("%s modifier = %v, want KeyModifierShortcutDefault", name, ks.Mod())
	}
}

func itemLabels(items []*fyne.MenuItem) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.Label
	}
	return out
}
