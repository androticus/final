package fyneui

import (
	"testing"

	"github.com/androticus/final/internal/core"
)

func TestBlitDrawer_HandleRender(t *testing.T) {
	sg := &SimGame{entities: NewSimEntityList()}
	sg.pixmap = core.NewPixmap(40, 40)
	ink := core.NewColor(0, 0, 255, 255)
	mat := core.NewMat2C(1, 1)
	mat.Put(0, 0, ink)
	d := NewBlitDrawer(mat)
	e := NewSimEntity()
	e.SetAt(core.NewPoint(3, 4))
	sg.AddEntity(e)
	d.HandleRender(e, d.NewCycler())
	got := sg.pixmap.ToMat2C().At_(4, 3)
	if got != ink {
		t.Fatalf("got %v want %v", got, ink)
	}
}

func TestBlitDrawer_FromPngAsset(t *testing.T) {
	m, err := mat2CFromImageFile("../final/examples/gui/simgame/assets/cannon.png")
	if err != nil {
		t.Fatal(err)
	}
	d := NewBlitDrawer(m)
	if d.mat.Rows() != 36 || d.mat.Cols() != 48 {
		t.Fatalf("size %dx%d", d.mat.Rows(), d.mat.Cols())
	}
	c := d.NewCycler()
	if c.phaseCount != 0 {
		t.Fatal("idle cycler expected")
	}
}
