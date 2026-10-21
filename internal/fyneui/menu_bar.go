package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

var (
	_ fyne.Widget       = (*MenuBar)(nil)
	_ desktop.Hoverable = (*menuBarLabel)(nil)
	_ fyne.Tappable     = (*menuBarLabel)(nil)
)

// MenuBar is an in-window menu strip (flat labels, popup menus).
// Use instead of fyne.MainMenu / Window.SetMainMenu.
type MenuBar struct {
	widget.BaseWidget
	labels []*menuBarLabel
	row    *fyne.Container
}

type menuBarLabel struct {
	widget.BaseWidget
	menu   *fyne.Menu
	parent *MenuBar
	hover  bool
	active bool
}

func NewMenuBar(menus *core.AnyList) *MenuBar {
	items := menusFromList(menus)
	labels := make([]*menuBarLabel, len(items))
	objs := make([]fyne.CanvasObject, len(items))
	bar := &MenuBar{}
	bar.ExtendBaseWidget(bar)
	for i, menu := range items {
		lbl := &menuBarLabel{menu: menu, parent: bar}
		lbl.ExtendBaseWidget(lbl)
		labels[i] = lbl
		objs[i] = lbl
	}
	bar.labels = labels
	bar.row = container.NewHBox(objs...)
	return bar
}

func (b *MenuBar) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	divider := canvas.NewRectangle(theme.Color(theme.ColorNameSeparator))
	objects := []fyne.CanvasObject{bg, b.row, divider}
	return &menuBarRenderer{
		bar:        b,
		background: bg,
		divider:    divider,
		objects:    objects,
	}
}

type menuBarRenderer struct {
	bar        *MenuBar
	background *canvas.Rectangle
	divider    *canvas.Rectangle
	objects    []fyne.CanvasObject
}

func menuBarDividerThickness() float32 {
	return theme.SeparatorThicknessSize()
}

func (r *menuBarRenderer) Layout(size fyne.Size) {
	thick := menuBarDividerThickness()
	rowH := size.Height - thick
	if rowH < 0 {
		rowH = 0
	}
	r.background.Resize(size)
	r.bar.row.Resize(fyne.NewSize(size.Width, rowH))
	r.bar.row.Move(fyne.NewPos(0, 0))
	r.divider.Resize(fyne.NewSize(size.Width, thick))
	r.divider.Move(fyne.NewPos(0, rowH))
}

func (r *menuBarRenderer) MinSize() fyne.Size {
	row := r.bar.row.MinSize()
	return fyne.NewSize(row.Width, row.Height+menuBarDividerThickness())
}

func (r *menuBarRenderer) Refresh() {
	r.background.FillColor = theme.Color(theme.ColorNameBackground)
	r.divider.FillColor = theme.Color(theme.ColorNameSeparator)
	r.background.Refresh()
	r.divider.Refresh()
	r.bar.row.Refresh()
}

func (r *menuBarRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *menuBarRenderer) Destroy() {}

func (l *menuBarLabel) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameHover))
	bg.CornerRadius = theme.SelectionRadiusSize()
	bg.Hide()
	text := canvas.NewText(l.menu.Label, theme.Color(theme.ColorNameForeground))
	return &menuBarLabelRenderer{
		label:      l,
		background: bg,
		text:       text,
		objects:    []fyne.CanvasObject{bg, text},
	}
}

type menuBarLabelRenderer struct {
	label      *menuBarLabel
	background *canvas.Rectangle
	text       *canvas.Text
	objects    []fyne.CanvasObject
}

func (r *menuBarLabelRenderer) Layout(size fyne.Size) {
	pad := menuBarLabelPadding()
	r.text.TextSize = theme.TextSize()
	r.text.Color = theme.Color(theme.ColorNameForeground)
	r.text.Resize(r.text.MinSize())
	r.text.Move(fyne.NewPos(pad.Width/2, pad.Height/2))
	r.background.Resize(size)
}

func (r *menuBarLabelRenderer) MinSize() fyne.Size {
	return r.text.MinSize().Add(menuBarLabelPadding())
}

func (r *menuBarLabelRenderer) Refresh() {
	if r.label.active || r.label.hover {
		if r.label.active {
			r.background.FillColor = theme.Color(theme.ColorNameFocus)
		} else {
			r.background.FillColor = theme.Color(theme.ColorNameHover)
		}
		r.background.Show()
	} else {
		r.background.Hide()
	}
	r.background.Refresh()
	r.text.Refresh()
}

func (r *menuBarLabelRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *menuBarLabelRenderer) Destroy() {}

func menuBarLabelPadding() fyne.Size {
	return fyne.NewSize(theme.InnerPadding()*2, theme.InnerPadding())
}

func (l *menuBarLabel) MouseIn(*desktop.MouseEvent) {
	l.hover = true
	l.Refresh()
}

func (l *menuBarLabel) MouseOut() {
	l.hover = false
	l.Refresh()
}

func (l *menuBarLabel) MouseMoved(*desktop.MouseEvent) {}

func (l *menuBarLabel) Tapped(*fyne.PointEvent) {
	l.showMenu()
}

func (l *menuBarLabel) showMenu() {
	c := fyne.CurrentApp().Driver().CanvasForObject(l)
	if c == nil {
		return
	}
	for _, peer := range l.parent.labels {
		if peer != l {
			peer.active = false
			peer.Refresh()
		}
	}
	l.active = true
	l.Refresh()
	pop := widget.NewPopUpMenu(l.menu, c)
	if pop == nil {
		return
	}
	pop.OnDismiss = func() {
		l.active = false
		l.Refresh()
		pop.Hide()
	}
	pop.ShowAtRelativePosition(fyne.NewPos(0, l.Size().Height), l)
}

func registerMenuBarMembers(t *core.Type) {
	// ( menus:List -- bar:MenuBar )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		menus := f.RP.DataStack.Pop().(*core.AnyList)
		f.RP.DataStack.Push(NewMenuBar(menus))
	})
}

func CreateRegisterMenuBarType() {
	t := core.NewRegisteringType("MenuBar",
		core.MakeHostTypeZero[*MenuBar](),
		core.MakeHostTypeCheckCoerce[*MenuBar]("MenuBar"),
	)
	registerMenuBarMembers(t)
	applyCanvasMixin[*MenuBar](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*MenuBar](t)
}

func init() {
	CreateRegisterMenuBarType()
}
