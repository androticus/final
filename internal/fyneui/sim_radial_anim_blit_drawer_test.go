package fyneui

import (
	"testing"

	"github.com/androticus/final/internal/core"
)

func TestRadialAnimBlit_Phase0Intact(t *testing.T) {
	sg := &SimGame{entities: NewSimEntityList()}
	sg.pixmap = core.NewPixmap(40, 40)

	ink := core.NewColor(0, 200, 0, 255)
	tile := core.NewMat2C(2, 2)
	tile.Put(0, 0, ink)

	mats := core.NewAnyMat(1, 1)
	mats.Put(0, 0, tile)
	vels := core.NewAnyMat(1, 1)
	vels.Put(0, 0, core.NewPoint(3, 4))

	d := NewRadialAnimBlitDrawer(mats, vels, 5)
	e := NewSimEntity()
	e.SetAt(core.NewPoint(10, 12))
	sg.AddEntity(e)

	cyc := d.NewCycler()
	cyc.phaseCount = 8
	cyc.cyclesToRun = 1
	cyc.Advance(0)
	cyc.Advance(0) // latch; phase 0
	d.HandleRender(e, cyc)

	got := sg.pixmap.ToMat2C().At_(12, 10)
	if got != ink {
		t.Fatalf("phase0: got %v want %v", got, ink)
	}
}

func TestRadialAnimBlit_PhaseMoves(t *testing.T) {
	sg := &SimGame{entities: NewSimEntityList()}
	sg.pixmap = core.NewPixmap(40, 40)

	ink := core.NewColor(0, 200, 0, 255)
	tile := core.NewMat2C(1, 1)
	tile.Put(0, 0, ink)

	mats := core.NewAnyMat(1, 1)
	mats.Put(0, 0, tile)
	vels := core.NewAnyMat(1, 1)
	vels.Put(0, 0, core.NewPoint(2, 0))

	d := NewRadialAnimBlitDrawer(mats, vels, 1)
	e := NewSimEntity()
	e.SetAt(core.NewPoint(5, 5))
	sg.AddEntity(e)

	cyc := d.NewCycler()
	cyc.phaseCount = 10
	cyc.cyclesToRun = 1
	cyc.Advance(0)
	cyc.Advance(3) // phase 3 after rate=1
	d.HandleRender(e, cyc)

	got := sg.pixmap.ToMat2C().At_(5, 5+2*3)
	if got != ink {
		t.Fatalf("phase3: got %v want %v at x=%d", got, ink, 5+6)
	}
}
