package fyneui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

func sliderNumeric(v any) float64 {
	switch x := v.(type) {
	case core.Float:
		return float64(x)
	case core.Int:
		return float64(x)
	case core.Uint:
		return float64(x)
	default:
		panic(fmt.Errorf("Slider: expected numeric, got %T", v))
	}
}

func NewSlider(min, max, step float64) *widget.Slider {
	sl := widget.NewSlider(min, max)
	sl.Step = step
	return sl
}

func registerSliderMembers(t *core.Type) {
	t.RegisterProperty("min", "Float",
		func(inst any) any {
			return core.Float(inst.(*widget.Slider).Min)
		},
		nil,
	)
	t.RegisterProperty("max", "Float",
		func(inst any) any {
			return core.Float(inst.(*widget.Slider).Max)
		},
		nil,
	)
	t.RegisterProperty("value", "Float",
		func(inst any) any {
			return core.Float(inst.(*widget.Slider).Value)
		},
		func(inst any, value any) {
			sl := inst.(*widget.Slider)
			val := (sliderNumeric(value))
			fyne.Do(func() {
				sl.Value = val
			})
		},
	)
	t.RegisterProperty("step", "Float",
		func(inst any) any {
			return core.Float(inst.(*widget.Slider).Step)
		},
		nil,
	)

	// ( handler:MethodRef index:Int -- )
	t.RegisterNativeMethod("setOnChanged", func(f *core.RunFrame, ths *core.TypeObjPair) {
		index := int(f.RP.DataStack.Pop().(core.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		sl := ths.Obj.(*widget.Slider)
		eventFn := core.GetEventLoop().CreateEventFn(handler, index)
		fyne.Do(func() {
			sl.OnChanged = func(v float64) {
				eventFn(sl, core.Float(v))
			}
		})
	})

	// ( handler:MethodRef index:Int -- )
	t.RegisterNativeMethod("setOnChangeEnded", func(f *core.RunFrame, ths *core.TypeObjPair) {
		index := int(f.RP.DataStack.Pop().(core.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		sl := ths.Obj.(*widget.Slider)
		eventFn := core.GetEventLoop().CreateEventFn(handler, index)
		fyne.Do(func() {
			sl.OnChangeEnded = func(v float64) {
				eventFn(sl, core.Float(v))
			}
		})
	})

	// ( w:n h:n -- )  Int/Uint/Float → fyne float32
	t.RegisterNativeMethod("resize", func(f *core.RunFrame, ths *core.TypeObjPair) {
		h := float32(core.CheckCoerceNumericToFloat(f.RP.DataStack.Pop()))
		w := float32(core.CheckCoerceNumericToFloat(f.RP.DataStack.Pop()))
		sl := ths.Obj.(*widget.Slider)
		fyne.Do(func() {
			sl.Resize(fyne.NewSize(w, h))
		})
	})

	// ( min:n max:n step:n -- slider:Slider )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		step := sliderNumeric(f.RP.DataStack.Pop())
		max := sliderNumeric(f.RP.DataStack.Pop())
		min := sliderNumeric(f.RP.DataStack.Pop())
		f.RP.DataStack.Push(NewSlider(min, max, step))
	})
}

func CreateRegisterSliderType() {
	t := core.NewRegisteringType("Slider",
		core.MakeHostTypeZero[*widget.Slider](),
		core.MakeHostTypeCheckCoerce[*widget.Slider]("Slider"),
	)
	registerSliderMembers(t)
	applyCanvasMixin[*widget.Slider](t)
	applyDisableableMixin[*widget.Slider](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*widget.Slider](t)
}

func init() {
	CreateRegisterSliderType()
}
