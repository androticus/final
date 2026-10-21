package finalembed_test

import (
	"testing"
	"time"

	"github.com/androticus/final/finalembed"
)

func TestFinalembedTypesCompileAndConstruct(t *testing.T) {
	list := finalembed.NewAnyList()
	list.Append(finalembed.Int(1))

	strs := finalembed.NewStringList()
	strs.Append("alpha")

	m := finalembed.NewStringMap()
	m.Put(finalembed.String("k"), finalembed.String("v"))

	narrow := finalembed.NewFloatList()
	narrow.Append(3.14)

	obj := finalembed.NewDynamicObject()
	obj.Set("x", finalembed.Int(42))

	sb := finalembed.NewStringBuilder()
	sb.WriteString("ok")

	c := finalembed.NewColor(254, 128, 0, 255)
	r, _, _, a := c.RGBA()
	if r>>8 != 254 || a>>8 != 255 {
		t.Fatalf("Color.RGBA high bytes = %d, %d want 254, 255", r>>8, a>>8)
	}

	pt := finalembed.NewPoint(3, 7)
	if pt.X != 3 || pt.Y != 7 {
		t.Fatalf("Point = %+v, want {3 7}", pt)
	}
	sz := finalembed.NewSize(10, 20)
	if sz.W != 10 || sz.H != 20 {
		t.Fatalf("Size = %+v, want {10 20}", sz)
	}
	rc := finalembed.NewRect(pt, sz)
	if rc.TopLeft != pt || rc.Size != sz {
		t.Fatalf("Rect = %+v, want topLeft %+v size %+v", rc, pt, sz)
	}

	p := finalembed.NewPainter(sz)
	if p.Pixmap() == nil {
		t.Fatal("Painter.Pixmap is nil")
	}
	if p.Font() == nil {
		t.Fatal("Painter.Font is nil")
	}
	if p.Stroke() == nil || p.Fill() == nil {
		t.Fatal("Painter.Stroke/Fill is nil")
	}

	var dt finalembed.DateTime = time.Now()
	if dt.IsZero() {
		t.Fatal("DateTime zero unexpectedly")
	}

	var iv finalembed.Interval
	_ = iv
}
