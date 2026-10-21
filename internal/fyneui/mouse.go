package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/androticus/final/internal/core"
)

func eventPoint(p fyne.Position) core.Point {
	return core.NewPointF(float64(p.X), float64(p.Y))
}

func eventDeltaPoint(d fyne.Delta) core.Point {
	return core.NewPointF(float64(d.DX), float64(d.DY))
}

func mouseIntConst(t *core.Type, name string, v int) {
	t.RegisterClassProperty(name, "Int",
		func(_ any) any { return core.Int(v) },
		nil,
	)
}

func registerMouseEventMembers(t *core.Type) {
	mouseIntConst(t, "btnPrimary", int(desktop.MouseButtonPrimary))
	mouseIntConst(t, "btnSecondary", int(desktop.MouseButtonSecondary))
	mouseIntConst(t, "btnTertiary", int(desktop.MouseButtonTertiary))
	mouseIntConst(t, "keyShift", int(fyne.KeyModifierShift))
	mouseIntConst(t, "keyControl", int(fyne.KeyModifierControl))
	mouseIntConst(t, "keyAlt", int(fyne.KeyModifierAlt))
	mouseIntConst(t, "keySuper", int(fyne.KeyModifierSuper))

	t.RegisterProperty("position", "Point",
		func(inst any) any { return eventPoint(inst.(*desktop.MouseEvent).Position) },
		nil,
	)
	t.RegisterProperty("absolutePosition", "Point",
		func(inst any) any { return eventPoint(inst.(*desktop.MouseEvent).AbsolutePosition) },
		nil,
	)
	t.RegisterProperty("button", "Int",
		func(inst any) any { return core.Int(inst.(*desktop.MouseEvent).Button) },
		nil,
	)
	t.RegisterProperty("modifier", "Int",
		func(inst any) any { return core.Int(inst.(*desktop.MouseEvent).Modifier) },
		nil,
	)
}

func registerScrollEventMembers(t *core.Type) {
	t.RegisterProperty("position", "Point",
		func(inst any) any { return eventPoint(inst.(*fyne.ScrollEvent).Position) },
		nil,
	)
	t.RegisterProperty("absolutePosition", "Point",
		func(inst any) any { return eventPoint(inst.(*fyne.ScrollEvent).AbsolutePosition) },
		nil,
	)
	t.RegisterProperty("scrolled", "Point",
		func(inst any) any { return eventDeltaPoint(inst.(*fyne.ScrollEvent).Scrolled) },
		nil,
	)
}

func CreateRegisterMouseEventType() {
	t := core.NewRegisteringType("MouseEvent",
		core.MakeHostTypeZero[*desktop.MouseEvent](),
		core.MakeHostTypeCheckCoerce[*desktop.MouseEvent]("MouseEvent"),
	)
	registerMouseEventMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*desktop.MouseEvent](t)
	mouseEventType = t
}

func CreateRegisterScrollEventType() {
	t := core.NewRegisteringType("ScrollEvent",
		core.MakeHostTypeZero[*fyne.ScrollEvent](),
		core.MakeHostTypeCheckCoerce[*fyne.ScrollEvent]("ScrollEvent"),
	)
	registerScrollEventMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*fyne.ScrollEvent](t)
	scrollEventType = t
}

var (
	mouseEventType  *core.Type
	scrollEventType *core.Type
)

func init() {
	CreateRegisterMouseEventType()
	CreateRegisterScrollEventType()
}
