package fyneui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/androticus/final/internal/core"
)

func testMenuBar() *MenuBar {
	menus := core.NewAnyListWithCapacity(1)
	menus.Append(fyne.NewMenu("File", fyne.NewMenuItem("X", func() {})))
	return NewMenuBar(menus)
}

func TestWindowHideAndSetHideOnCloseRegistered(t *testing.T) {
	typ := core.GetTypeRegistrar().Get("Window")
	if typ == nil {
		t.Fatal("Window missing")
	}
	for _, name := range []string{"hide", "setHideOnClose"} {
		if typ.GetMethodMetadata(name) == nil {
			t.Fatalf("Window.%s missing", name)
		}
	}
}

func TestWindowSetMenuBarWrapsBorder(t *testing.T) {
	w := &Window{}
	bar := testMenuBar()
	w.setMenuBar(bar)
	body := canvas.NewRectangle(color.Black)
	tree := w.userContentTree(body)
	c, ok := tree.(*fyne.Container)
	if !ok {
		t.Fatalf("wrapped content is %T, want *fyne.Container", tree)
	}
	foundBar, foundBody := false, false
	for _, o := range c.Objects {
		if o == bar {
			foundBar = true
		}
		if o == body {
			foundBody = true
		}
	}
	if !foundBar || !foundBody {
		t.Fatalf("Border objects = %v, want MenuBar and body", c.Objects)
	}
}

func TestWindowSetContentWithoutMenuBarIsBare(t *testing.T) {
	w := &Window{}
	body := canvas.NewRectangle(color.Black)
	tree := w.userContentTree(body)
	if tree != body {
		t.Fatalf("without menu bar, setContent should be the user widget, got %T", tree)
	}
}

func TestWindowSetMenuBarAfterContentPanics(t *testing.T) {
	w := &Window{}
	w.userContentTree(canvas.NewRectangle(color.Black))
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	w.setMenuBar(testMenuBar())
}

func TestWindowSetMenuBarTwicePanics(t *testing.T) {
	w := &Window{}
	w.setMenuBar(testMenuBar())
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	w.setMenuBar(testMenuBar())
}

func TestWindowSetMenuBarNilPanics(t *testing.T) {
	w := &Window{}
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	w.setMenuBar(nil)
}
