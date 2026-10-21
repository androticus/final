package fyneui

import (
	"testing"

	"github.com/androticus/final/internal/core"
)

func TestAnimBlitDrawer_NewCycler(t *testing.T) {
	f0 := core.NewMat2C(2, 2)
	f1 := core.NewMat2C(2, 2)
	d := NewAnimBlitDrawer([]*core.Mat2C{f0, f1}, 10)
	c := d.NewCycler()
	if c.phaseCount != 2 || c.phaseRate != 10 || c.cyclesToRun != 0 {
		t.Fatalf("cycler recipe: count=%d rate=%d run=%d", c.phaseCount, c.phaseRate, c.cyclesToRun)
	}
}

func TestAnimBlitDrawer_HandleRenderPhases(t *testing.T) {
	sg := &SimGame{entities: NewSimEntityList()}
	sg.pixmap = core.NewPixmap(40, 40)
	sg.AddEntity(NewSimEntity()) // ensure sg linkage path exists

	ink0 := core.NewColor(255, 0, 0, 255)
	ink1 := core.NewColor(0, 255, 0, 255)
	f0 := core.NewMat2C(1, 1)
	f0.Put(0, 0, ink0)
	f1 := core.NewMat2C(1, 1)
	f1.Put(0, 0, ink1)
	d := NewAnimBlitDrawer([]*core.Mat2C{f0, f1}, 10)

	e := NewSimEntity()
	e.SetAt(core.NewPoint(5, 7))
	sg.AddEntity(e)

	cyc := d.NewCycler()
	cyc.Advance(0)
	cyc.Advance(0) // latch
	d.HandleRender(e, cyc)
	got := sg.pixmap.ToMat2C().At_(7, 5)
	if got != ink0 {
		t.Fatalf("phase0: got %v want %v", got, ink0)
	}

	sg.pixmap.Clear()
	cyc.Advance(15) // phase 1
	d.HandleRender(e, cyc)
	got = sg.pixmap.ToMat2C().At_(7, 5)
	if got != ink1 {
		t.Fatalf("phase1: got %v want %v", got, ink1)
	}
}

func TestAnimBlitDrawer_FromPngAssets(t *testing.T) {
	// go test cwd is the package dir
	a, err := mat2CFromImageFile("../final/examples/gui/simgame/assets/alien-a.png")
	if err != nil {
		t.Fatal(err)
	}
	b, err := mat2CFromImageFile("../final/examples/gui/simgame/assets/alien-b.png")
	if err != nil {
		t.Fatal(err)
	}
	if a.Rows() != 24 || a.Cols() != 32 || b.Rows() != 24 || b.Cols() != 32 {
		t.Fatalf("size a=%dx%d b=%dx%d", a.Rows(), a.Cols(), b.Rows(), b.Cols())
	}
	d := NewAnimBlitDrawer([]*core.Mat2C{a, b}, 200)
	if len(d.frames) != 2 {
		t.Fatal("want 2 frames")
	}
}
