package fyneui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/androticus/final/internal/core"
)

func menusFromList(list *core.AnyList) []*fyne.Menu {
	n := list.Len()
	menus := make([]*fyne.Menu, n)
	for i := 0; i < n; i++ {
		item := list.At(i)
		menu, ok := item.(*fyne.Menu)
		if !ok {
			panic(fmt.Errorf("MenuBar: item %d is %T, want *fyne.Menu", i, item))
		}
		menus[i] = menu
	}
	return menus
}

func menuItemsFromList(list *core.AnyList) []*fyne.MenuItem {
	n := list.Len()
	items := make([]*fyne.MenuItem, n)
	for i := 0; i < n; i++ {
		item := list.At(i)
		mi, ok := item.(*fyne.MenuItem)
		if !ok {
			panic(fmt.Errorf("Menu: item %d is %T, want *fyne.MenuItem", i, item))
		}
		items[i] = mi
	}
	return items
}

func NewMenu(label string, items *core.AnyList) *fyne.Menu {
	return fyne.NewMenu(label, menuItemsFromList(items)...)
}

func NewMenuItem(handler *core.MethodRef, index int, label string) *fyne.MenuItem {
	var item *fyne.MenuItem
	action := func() {}
	if handler != nil {
		eventFn := core.GetEventLoop().CreateEventFn(handler, index)
		action = func() {
			eventFn(item)
		}
	}
	item = fyne.NewMenuItem(label, action)
	return item
}

func registerMenuMembers(t *core.Type) {
	// ( menuItems:List label:String -- menu:Menu )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		label := string(f.RP.DataStack.Pop().(core.String))
		items := f.RP.DataStack.Pop().(*core.AnyList)
		f.RP.DataStack.Push(NewMenu(label, items))
	})
}

func registerMenuItemMembers(t *core.Type) {
	t.RegisterProperty("enabled", "Bool",
		func(inst any) any {
			return core.Bool(!inst.(*fyne.MenuItem).Disabled)
		},
		func(inst any, value any) {
			item := inst.(*fyne.MenuItem)
			enabled := bool(value.(core.Bool))
			fyne.Do(func() {
				item.Disabled = !enabled
			})
		},
	)
	t.RegisterProperty("checked", "Bool",
		func(inst any) any {
			return core.Bool(inst.(*fyne.MenuItem).Checked)
		},
		func(inst any, value any) {
			item := inst.(*fyne.MenuItem)
			checked := bool(value.(core.Bool))
			fyne.Do(func() {
				item.Checked = checked
			})
		},
	)

	// ( key:String -- )
	// Cmd+key on macOS, Ctrl+key elsewhere. key is a fyne.KeyName string (e.g. "L", "F1").
	t.RegisterNativeMethod("setDefaultShortcut", func(f *core.RunFrame, ths *core.TypeObjPair) {
		key := fyne.KeyName(f.RP.DataStack.Pop().(core.String))
		item := ths.Obj.(*fyne.MenuItem)
		fyne.Do(func() {
			item.Shortcut = &desktop.CustomShortcut{
				KeyName:  key,
				Modifier: fyne.KeyModifierShortcutDefault,
			}
		})
	})

	// ( handler:MethodRef index:Int label:String -- menuItem:MenuItem )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		label := string(f.RP.DataStack.Pop().(core.String))
		index := int(f.RP.DataStack.Pop().(core.Int))
		handler, _ := f.RP.DataStack.Pop().(*core.MethodRef)
		f.RP.DataStack.Push(NewMenuItem(handler, index, label))
	})
	// ( -- menuItem:MenuItem )
	t.RegisterNativeClassMethod("newSeparator", func(f *core.RunFrame, _ *core.TypeObjPair) {
		f.RP.DataStack.Push(fyne.NewMenuItemSeparator())
	})
}

func CreateRegisterMenuType() {
	t := core.NewRegisteringType("Menu",
		core.MakeHostTypeZero[*fyne.Menu](),
		core.MakeHostTypeCheckCoerce[*fyne.Menu]("Menu"),
	)
	registerMenuMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*fyne.Menu](t)
}

func CreateRegisterMenuItemType() {
	t := core.NewRegisteringType("MenuItem",
		core.MakeHostTypeZero[*fyne.MenuItem](),
		core.MakeHostTypeCheckCoerce[*fyne.MenuItem]("MenuItem"),
	)
	registerMenuItemMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*fyne.MenuItem](t)
}

func init() {
	CreateRegisterMenuType()
	CreateRegisterMenuItemType()
}
