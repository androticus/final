package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

type ListRef struct {
	list *core.StringList
}

func NewListRef(list *core.StringList) *ListRef {
	return &ListRef{list: list}
}

func (lr *ListRef) Len() int {
	if lr.list == nil {
		return 0
	}
	return lr.list.Len()
}

func (lr *ListRef) At(index int) string {
	if lr.list == nil {
		return ""
	}
	return lr.list.At(index)
}

func (lr *ListRef) GetList() *core.StringList {
	return lr.list
}

func (lr *ListRef) SetList(list *core.StringList) {
	lr.list = list
}

func NewListView(lr *ListRef) *widget.List {
	return widget.NewList(
		func() int { return lr.Len() },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(lr.At(id))
		},
	)
}

func registerListRefMembers(t *core.Type) {
	t.RegisterProperty("list", "StringList",
		func(inst any) any {
			return inst.(*ListRef).GetList()
		},
		func(inst any, value any) {
			inst.(*ListRef).SetList(value.(*core.StringList))
		},
	)

	// ( list:StringList -- ref:ListRef )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		list := f.RP.DataStack.Pop().(*core.StringList)
		f.RP.DataStack.Push(NewListRef(list))
	})
}

func CreateRegisterListRefType() {
	t := core.NewRegisteringType("ListRef",
		core.MakeHostTypeZero[*ListRef](),
		core.MakeHostTypeCheckCoerce[*ListRef]("ListRef"),
	)
	registerListRefMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*ListRef](t)
}

func registerListViewMembers(t *core.Type) {
	// ( handler:MethodRef index:Int -- )
	t.RegisterNativeMethod("setOnSelected", func(f *core.RunFrame, ths *core.TypeObjPair) {
		index := int(f.RP.DataStack.Pop().(core.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		lv := ths.Obj.(*widget.List)
		eventFn := core.GetEventLoop().CreateEventFn(handler, index)
		fyne.Do(func() {
			lv.OnSelected = func(id widget.ListItemID) {
				eventFn(lv, core.Int(id))
			}
		})
	})

	// ( ref:ListRef -- w:ListView )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ref := f.RP.DataStack.Pop().(*ListRef)
		f.RP.DataStack.Push(NewListView(ref))
	})
}

func CreateRegisterListViewType() {
	t := core.NewRegisteringType("ListView",
		core.MakeHostTypeZero[*widget.List](),
		core.MakeHostTypeCheckCoerce[*widget.List]("ListView"),
	)
	registerListViewMembers(t)
	applyCanvasMixin[*widget.List](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*widget.List](t)
}

func init() {
	CreateRegisterListRefType()
	CreateRegisterListViewType()
}
