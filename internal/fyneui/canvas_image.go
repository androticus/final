package fyneui

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

// CanvasImage wraps canvas.Image and exposes configurable mouse delegates.
// Note: we don't implement Tappable and TappableSecondary; use Mouseable
type CanvasImage struct {
	widget.BaseWidget

	Image *canvas.Image

	CursorShape desktop.StandardCursor

	// Delegate Functors
	// from desktop.Mouseable
	OnMouseDown func(ev *desktop.MouseEvent)
	OnMouseUp   func(ev *desktop.MouseEvent)
	// from desktop.Hoverable
	OnMouseIn    func(ev *desktop.MouseEvent)
	OnMouseOut   func()
	OnMouseMoved func(ev *desktop.MouseEvent)
	// from fyne.Scrollable
	OnMouseScrolled func(ev *fyne.ScrollEvent)
}

func NewCanvasImage(src image.Image) *CanvasImage {
	img := canvas.NewImageFromImage(src)
	// Stretch + pixmap-pixel MinSize: 1 pixmap pixel = 1 Fyne DIP.
	// ImageFillOriginal uses physical pixels (minSize /= canvas scale) and
	// letterboxes in a wider cell — HiDPI / RDP shows a tiny centered pixmap.
	img.FillMode = canvas.ImageFillStretch
	img.ScaleMode = canvas.ImageScalePixels
	if src != nil {
		b := src.Bounds()
		img.SetMinSize(fyne.NewSize(float32(b.Dx()), float32(b.Dy())))
	}
	ci := &CanvasImage{Image: img}
	ci.ExtendBaseWidget(ci)
	return ci
}

func (ci *CanvasImage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ci.Image)
}

func (ci *CanvasImage) MinSize() fyne.Size {
	if ci.Image != nil && ci.Image.Image != nil {
		b := ci.Image.Image.Bounds()
		return fyne.NewSize(float32(b.Dx()), float32(b.Dy()))
	}
	return ci.BaseWidget.MinSize()
}

func (ci *CanvasImage) FinalType() *core.Type {
	return canvasImageType
}

func (ci *CanvasImage) SetCursorShape(shape desktop.StandardCursor) {
	if shape == ci.CursorShape {
		return
	}
	if int(shape) < int(desktop.DefaultCursor) || int(shape) > int(desktop.HiddenCursor) {
		return
	}
	ci.CursorShape = shape
	ci.Refresh()
}

func (ci *CanvasImage) Refresh() {
	fyne.Do(func() {
		ci.Image.Refresh()
	})
}

func mintMouseEventFn(ci *CanvasImage, handler *core.MethodRef, index int) func(*desktop.MouseEvent) {
	if handler == nil {
		return nil
	}
	eventFn := core.GetEventLoop().CreateEventFn(handler, index)
	return func(ev *desktop.MouseEvent) {
		eventFn(ci, ev)
	}
}

func mintMouseOutFn(ci *CanvasImage, handler *core.MethodRef, index int) func() {
	if handler == nil {
		return nil
	}
	eventFn := core.GetEventLoop().CreateEventFn(handler, index)
	return func() {
		eventFn(ci)
	}
}

func mintScrollEventFn(ci *CanvasImage, handler *core.MethodRef, index int) func(*fyne.ScrollEvent) {
	if handler == nil {
		return nil
	}
	eventFn := core.GetEventLoop().CreateEventFn(handler, index)
	return func(ev *fyne.ScrollEvent) {
		eventFn(ci, ev)
	}
}

// SetOnMouseDownUp installs down/up handlers. A nil MethodRef leaves that slot unset.
// Down and up may share a MethodRef (both deliver MouseEvent in Event.params[0]).
func (ci *CanvasImage) SetOnMouseDownUp(
	downHandler *core.MethodRef, downIndex int,
	upHandler *core.MethodRef, upIndex int) {
	downFn := mintMouseEventFn(ci, downHandler, downIndex)
	upFn := mintMouseEventFn(ci, upHandler, upIndex)
	fyne.Do(func() {
		ci.OnMouseDown = downFn
		ci.OnMouseUp = upFn
	})
}

// SetOnMouseInMovedOut installs hover handlers. A nil MethodRef leaves that slot unset.
// In and moved may share a MethodRef (MouseEvent). Out has no event payload.
func (ci *CanvasImage) SetOnMouseInMovedOut(
	inHandler *core.MethodRef, inIndex int,
	movedHandler *core.MethodRef, movedIndex int,
	outHandler *core.MethodRef, outIndex int) {
	inFn := mintMouseEventFn(ci, inHandler, inIndex)
	movedFn := mintMouseEventFn(ci, movedHandler, movedIndex)
	outFn := mintMouseOutFn(ci, outHandler, outIndex)
	fyne.Do(func() {
		ci.OnMouseIn = inFn
		ci.OnMouseMoved = movedFn
		ci.OnMouseOut = outFn
	})
}

// SetOnMouseScrolled installs the scroll handler. A nil MethodRef leaves it unset.
func (ci *CanvasImage) SetOnMouseScrolled(
	scrolledHandler *core.MethodRef, scrolledIndex int) {
	scrolledFn := mintScrollEventFn(ci, scrolledHandler, scrolledIndex)
	fyne.Do(func() {
		ci.OnMouseScrolled = scrolledFn
	})
}

// desktop.Cursorable
func (ci *CanvasImage) Cursor() desktop.Cursor {
	return ci.CursorShape
}

// --- desktop.Mouseable ---

func (ci *CanvasImage) MouseDown(ev *desktop.MouseEvent) {
	if ci.OnMouseDown != nil {
		ci.OnMouseDown(ev)
	}
}

func (ci *CanvasImage) MouseUp(ev *desktop.MouseEvent) {
	if ci.OnMouseUp != nil {
		ci.OnMouseUp(ev)
	}
}

// --- desktop.Hoverable ---

func (ci *CanvasImage) MouseIn(ev *desktop.MouseEvent) {
	if ci.OnMouseIn != nil {
		ci.OnMouseIn(ev)
	}
}

func (ci *CanvasImage) MouseOut() {
	if ci.OnMouseOut != nil {
		ci.OnMouseOut()
	}
}

func (ci *CanvasImage) MouseMoved(ev *desktop.MouseEvent) {
	if ci.OnMouseMoved != nil {
		ci.OnMouseMoved(ev)
	}
}

// --- fyne.Scrollable ---

func (ci *CanvasImage) Scrolled(ev *fyne.ScrollEvent) {
	if ci.OnMouseScrolled != nil {
		ci.OnMouseScrolled(ev)
	}
}

func canvasImageIntConst(t *core.Type, name string, v int) {
	t.RegisterClassProperty(name, "Int",
		func(_ any) any { return core.Int(v) },
		nil,
	)
}

func registerCanvasImageMembers(t *core.Type) {
	canvasImageIntConst(t, "cursorDefault", int(desktop.DefaultCursor))
	canvasImageIntConst(t, "cursorText", int(desktop.TextCursor))
	canvasImageIntConst(t, "cursorCrosshair", int(desktop.CrosshairCursor))
	canvasImageIntConst(t, "cursorPointer", int(desktop.PointerCursor))
	canvasImageIntConst(t, "cursorHResize", int(desktop.HResizeCursor))
	canvasImageIntConst(t, "cursorVResize", int(desktop.VResizeCursor))
	canvasImageIntConst(t, "cursorNESWResize", int(desktop.NESWResizeCursor))
	canvasImageIntConst(t, "cursorNWSEResize", int(desktop.NWSEResizeCursor))
	canvasImageIntConst(t, "cursorHidden", int(desktop.HiddenCursor))

	t.RegisterProperty("cursorShape", "Int",
		func(inst any) any {
			return core.Int(int(inst.(*CanvasImage).CursorShape))
		},
		func(inst any, value any) {
			ci := inst.(*CanvasImage)
			shape := desktop.StandardCursor(value.(core.Int))
			ci.SetCursorShape(shape)
		},
	)

	// ( down:MethodRef downIdx:Int up:MethodRef upIdx:Int -- )
	// each MethodRef may be nil to skip that event
	t.RegisterNativeMethod("setOnMouseDownUp", func(f *core.RunFrame, ths *core.TypeObjPair) {
		upIndex := int(f.RP.DataStack.Pop().(core.Int))
		upHandler, _ := f.RP.DataStack.Pop().(*core.MethodRef)
		downIndex := int(f.RP.DataStack.Pop().(core.Int))
		downHandler, _ := f.RP.DataStack.Pop().(*core.MethodRef)
		ths.Obj.(*CanvasImage).SetOnMouseDownUp(downHandler, downIndex, upHandler, upIndex)
	})

	// ( in:MethodRef inIdx:Int moved:MethodRef movedIdx:Int out:MethodRef outIdx:Int -- )
	// each MethodRef may be nil to skip that event
	t.RegisterNativeMethod("setOnMouseInMovedOut", func(f *core.RunFrame, ths *core.TypeObjPair) {
		outIndex := int(f.RP.DataStack.Pop().(core.Int))
		outHandler, _ := f.RP.DataStack.Pop().(*core.MethodRef)
		movedIndex := int(f.RP.DataStack.Pop().(core.Int))
		movedHandler, _ := f.RP.DataStack.Pop().(*core.MethodRef)
		inIndex := int(f.RP.DataStack.Pop().(core.Int))
		inHandler, _ := f.RP.DataStack.Pop().(*core.MethodRef)
		ths.Obj.(*CanvasImage).SetOnMouseInMovedOut(
			inHandler, inIndex, movedHandler, movedIndex, outHandler, outIndex)
	})

	// ( scrolled:MethodRef scrolledIdx:Int -- )
	// MethodRef may be nil to skip
	t.RegisterNativeMethod("setOnMouseScrolled", func(f *core.RunFrame, ths *core.TypeObjPair) {
		scrolledIndex := int(f.RP.DataStack.Pop().(core.Int))
		scrolledHandler, _ := f.RP.DataStack.Pop().(*core.MethodRef)
		ths.Obj.(*CanvasImage).SetOnMouseScrolled(scrolledHandler, scrolledIndex)
	})

	// ( img:Any -- img:CanvasImage )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		src := f.RP.DataStack.Pop().(image.Image)
		f.RP.DataStack.Push(NewCanvasImage(src))
	})
}

func CreateRegisterCanvasImageType() {
	t := core.NewRegisteringType("CanvasImage",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[CanvasImage]("CanvasImage"),
	)
	registerCanvasImageMembers(t)
	applyCanvasMixin[*CanvasImage](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	canvasImageType = t
}

// Interface Compliance Assertions
var (
	canvasImageType *core.Type
	_               desktop.Mouseable = (*CanvasImage)(nil)
	_               desktop.Hoverable = (*CanvasImage)(nil)
	_               fyne.Scrollable   = (*CanvasImage)(nil)
	_               core.Typed        = (*CanvasImage)(nil)
)

func init() {
	CreateRegisterCanvasImageType()
}
