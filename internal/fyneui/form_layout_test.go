package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

func formLayoutNewFrame(t *testing.T, list *core.AnyList) *core.RunFrame {
	t.Helper()
	typ := core.GetTypeRegistrar().Get("FormLayout")
	if typ == nil {
		t.Fatal("FormLayout missing")
	}
	mm := typ.GetClassMethodMetadata("new")
	if mm == nil {
		t.Fatal("FormLayout.new missing")
	}
	f := &core.RunFrame{RP: core.NewTestRunProvider()}
	f.RP.DataStack.Push(list)
	mm.Invoke(f, core.NewTypeObjPair(typ))
	return f
}

func TestFormLayoutTypeRegistered(t *testing.T) {
	typ := core.GetTypeRegistrar().Get("FormLayout")
	if typ == nil {
		t.Fatal("FormLayout missing")
	}
	if typ.GetClassMethodMetadata("new") == nil {
		t.Fatal("FormLayout.new missing")
	}
}

func TestFormLayoutNewEven(t *testing.T) {
	list := core.NewAnyList()
	list.Append(widget.NewLabel("Name"))
	list.Append(widget.NewEntry())
	list.Append(widget.NewLabel("Age"))
	list.Append(widget.NewEntry())

	f := formLayoutNewFrame(t, list)
	c, ok := f.RP.DataStack.Pop().(*fyne.Container)
	if !ok || c == nil {
		t.Fatalf("got %T, want *fyne.Container", f.RP.DataStack.Peek())
	}
	if len(c.Objects) != 4 {
		t.Fatalf("Objects=%d, want 4", len(c.Objects))
	}
}

func TestFormLayoutNewOddPanics(t *testing.T) {
	list := core.NewAnyList()
	list.Append(widget.NewLabel("Name"))
	list.Append(widget.NewEntry())
	list.Append(widget.NewLabel("orphan"))

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on odd length")
		}
	}()
	_ = formLayoutNewFrame(t, list)
}
