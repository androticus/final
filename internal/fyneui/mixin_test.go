package fyneui

import (
	"testing"

	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

func mixinNamed(t *core.Type, name string) bool {
	for _, m := range t.GetMixinTypes() {
		if m != nil && m.Name() == name {
			return true
		}
	}
	return false
}

func TestCanvasMixinOnLabel(t *testing.T) {
	typ := core.GetTypeRegistrar().Get("Label")
	if typ == nil {
		t.Fatal("Label missing")
	}
	if !mixinNamed(typ, "Canvas") {
		t.Fatal("Label missing Canvas mixin")
	}
	if mixinNamed(typ, "Disableable") {
		t.Fatal("Label must not have Disableable mixin")
	}
	vis := typ.GetPropertyMetadata("visible")
	if vis == nil || vis.MixinType() == nil || vis.MixinType().Name() != "Canvas" {
		t.Fatalf("Label.visible mixinType = %v, want Canvas", vis)
	}
	ref := typ.GetMethodMetadata("refresh")
	if ref == nil || ref.MixinType() == nil || ref.MixinType().Name() != "Canvas" {
		t.Fatalf("Label.refresh mixinType = %v, want Canvas", ref)
	}
	if typ.GetPropertyMetadata("disabled") != nil {
		t.Fatal("Label must not have disabled")
	}
}

func TestDisableableMixinOnButton(t *testing.T) {
	typ := core.GetTypeRegistrar().Get("Button")
	if typ == nil {
		t.Fatal("Button missing")
	}
	if !mixinNamed(typ, "Canvas") || !mixinNamed(typ, "Disableable") {
		t.Fatalf("Button mixins = %v, want Canvas and Disableable", typ.GetMixinTypes())
	}
	dis := typ.GetPropertyMetadata("disabled")
	if dis == nil || dis.MixinType() == nil || dis.MixinType().Name() != "Disableable" {
		t.Fatalf("Button.disabled mixinType = %v, want Disableable", dis)
	}
}

func TestEntryDisabledFromDisableableMixin(t *testing.T) {
	typ := core.GetTypeRegistrar().Get("Entry")
	if typ == nil {
		t.Fatal("Entry missing")
	}
	dis := typ.GetPropertyMetadata("disabled")
	if dis == nil || dis.MixinType() == nil || dis.MixinType().Name() != "Disableable" {
		t.Fatalf("Entry.disabled mixinType = %v, want Disableable", dis)
	}
}

func TestCanvasImageRefreshFromCanvasMixin(t *testing.T) {
	typ := core.GetTypeRegistrar().Get("CanvasImage")
	if typ == nil {
		t.Fatal("CanvasImage missing")
	}
	ref := typ.GetMethodMetadata("refresh")
	if ref == nil || ref.MixinType() == nil || ref.MixinType().Name() != "Canvas" {
		t.Fatalf("CanvasImage.refresh mixinType = %v, want Canvas", ref)
	}
}

func TestContainerCanvasMixin(t *testing.T) {
	typ := core.GetTypeRegistrar().Get("Container")
	if typ == nil {
		t.Fatal("Container missing")
	}
	if !mixinNamed(typ, "Canvas") {
		t.Fatal("Container missing Canvas mixin")
	}
	if typ.GetMethodMetadata("refresh") == nil {
		t.Fatal("Container missing refresh")
	}
}

func TestClassOnlyAndNonWidgetsSkipCanvas(t *testing.T) {
	for _, name := range []string{"Box", "Grid", "Border", "Icon", "MenuItem", "ListRef"} {
		typ := core.GetTypeRegistrar().Get(name)
		if typ == nil {
			t.Fatalf("%s missing", name)
		}
		if mixinNamed(typ, "Canvas") {
			t.Fatalf("%s must not have Canvas mixin", name)
		}
		if typ.GetPropertyMetadata("visible") != nil {
			t.Fatalf("%s must not have visible", name)
		}
	}
}

func TestCanvasCatalogHasMembers(t *testing.T) {
	typ := core.GetTypeRegistrar().Get("Canvas")
	if typ == nil {
		t.Fatal("Canvas catalog type missing")
	}
	if typ.GetPropertyMetadata("visible") == nil || typ.GetMethodMetadata("refresh") == nil {
		t.Fatal("Canvas catalog missing visible/refresh")
	}
	if canvasType.GetMethodMetadata("refresh").MixinType() != nil {
		t.Fatal("Canvas catalog refresh must not be stamped as mixin")
	}
}

func TestDisableableClients(t *testing.T) {
	yes := []string{"Button", "Entry", "Check", "RadioGroup", "Select", "Slider", "Form"}
	no := []string{"Label", "Separator", "Activity", "RichText", "Accordion", "ListView", "TableView", "TreeView", "Container"}
	for _, name := range yes {
		typ := core.GetTypeRegistrar().Get(name)
		if typ == nil || !mixinNamed(typ, "Disableable") {
			t.Fatalf("%s missing Disableable mixin", name)
		}
	}
	for _, name := range no {
		typ := core.GetTypeRegistrar().Get(name)
		if typ == nil {
			t.Fatalf("%s missing", name)
		}
		if mixinNamed(typ, "Disableable") {
			t.Fatalf("%s must not have Disableable mixin", name)
		}
	}
}

func TestLabelVisibleGetWithoutWindow(t *testing.T) {
	lbl := widget.NewLabel("x")
	if !lbl.Visible() {
		t.Fatal("new Label should be visible")
	}
}
