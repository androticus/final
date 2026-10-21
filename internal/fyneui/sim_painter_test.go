package fyneui

import (
	"testing"

	"github.com/androticus/final/internal/core"
)

func TestSimGame_PainterJIT(t *testing.T) {
	sg := &SimGame{}
	defer func() {
		if recover() == nil {
			t.Fatal("Painter() with nil pixmap should panic")
		}
	}()
	_ = sg.Painter()
}

func TestSimGame_PainterBindsPixmapAndCaches(t *testing.T) {
	sg := &SimGame{}
	pm := core.NewPainter(core.NewSize(32, 24)).Pixmap()
	sg.SetPixmap(pm)
	p1 := sg.Painter()
	if p1 == nil {
		t.Fatal("Painter() nil")
	}
	if p1.Pixmap() != pm {
		t.Fatal("Painter not bound to sim pixmap")
	}
	p2 := sg.Painter()
	if p1 != p2 {
		t.Fatal("Painter() should cache")
	}
	pm2 := core.NewPainter(core.NewSize(16, 16)).Pixmap()
	sg.SetPixmap(pm2)
	p3 := sg.Painter()
	if p3 == p1 {
		t.Fatal("Painter cache should clear when pixmap replaced")
	}
	if p3.Pixmap() != pm2 {
		t.Fatal("new Painter should bind replaced pixmap")
	}
}
