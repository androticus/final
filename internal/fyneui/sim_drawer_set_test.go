package fyneui

import (
	"testing"

	"github.com/androticus/final/internal/core"
)

func TestSimDrawerSet_NewBindsHandleRender(t *testing.T) {
	d := NewAnimBlitDrawer([]*core.Mat2C{core.NewMat2C(1, 1)}, 10)
	c := NewSimDrawerCycler()
	set := NewSimDrawerSet(d, c)
	if set.Render() == nil || set.Render().MM.Name() != "handleRender" {
		t.Fatalf("render = %v", set.Render())
	}
	if set.Render().Parent.Obj != d {
		t.Fatal("render Parent should be drawer")
	}
}

func TestSimGame_EntitiesWithDrawerSets_SortsByZorder(t *testing.T) {
	sg := &SimGame{entities: NewSimEntityList()}
	mk := func(z int16) *SimEntity {
		e := NewSimEntity()
		e.zorder = z
		e.drawerSet = NewSimDrawerSet(
			NewAnimBlitDrawer([]*core.Mat2C{core.NewMat2C(1, 1)}, 10),
			NewSimDrawerCycler(),
		)
		sg.AddEntity(e)
		return e
	}
	low := mk(-1)
	mid := mk(0)
	hi := mk(2)
	sg.AddEntity(NewSimEntity()) // no drawerSet — omitted

	list := sg.EntitiesWithDrawerSets()
	if list.Len() != 3 {
		t.Fatalf("len=%d want 3", list.Len())
	}
	if list.At(0) != low || list.At(1) != mid || list.At(2) != hi {
		t.Fatalf("order wrong: %v %v %v", list.At(0), list.At(1), list.At(2))
	}
}
