package core

import (
	"image"
	"testing"
)

func TestTextAlignClassConsts(t *testing.T) {
	typ := GetTypeRegistrar().Get("TextAlign")
	if typ == nil {
		t.Fatal("TextAlign type missing")
	}
	get := func(name string) Int {
		pm := typ.GetClassPropertyMetadata(name)
		if pm == nil || pm.get == nil {
			t.Fatalf("TextAlign.%s missing", name)
		}
		return pm.get(nil).(Int)
	}
	if get("left") != Int(TextAlignLeft) || get("centerMiddle") != Int(TextAlignCenterMiddle) {
		t.Fatal("atom / composed const mismatch")
	}
	if get("left")|get("middle") != get("leftMiddle") {
		t.Fatal("left | middle should equal leftMiddle")
	}
	if get("right")|get("baseline") != get("rightBaseline") {
		t.Fatal("right | baseline should equal rightBaseline")
	}
}

func TestFontAlignRejectsIncomplete(t *testing.T) {
	f := NewPainter(NewSize(8, 8)).Font()
	defer func() {
		if recover() == nil {
			t.Fatal("bare left should panic")
		}
	}()
	f.SetAlign(TextAlignLeft)
}

func TestFontAlignRejectsTwoHorizontal(t *testing.T) {
	f := NewPainter(NewSize(8, 8)).Font()
	defer func() {
		if recover() == nil {
			t.Fatal("left|center|middle should panic")
		}
	}()
	f.SetAlign(TextAlignLeft | TextAlignCenter | TextAlignMiddle)
}

func TestFontAlignNativeProperty(t *testing.T) {
	typ := GetTypeRegistrar().Get("Font")
	pm := typ.GetPropertyMetadata("align")
	if pm == nil || pm.set == nil {
		t.Fatal("Font.align missing")
	}
	f := NewPainter(NewSize(8, 8)).Font()
	pm.SetValue(f, Int(TextAlignLeftBaseline))
	if f.Align() != TextAlignLeftBaseline {
		t.Fatalf("align %#x", f.Align())
	}
	if pm.GetValue(f).(Int) != Int(TextAlignLeftBaseline) {
		t.Fatal("align get")
	}
}

func paintedBBox(img *image.RGBA) image.Rectangle {
	r := image.Rectangle{}
	found := false
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.RGBAAt(x, y).A == 0 {
				continue
			}
			if !found {
				r = image.Rect(x, y, x+1, y+1)
				found = true
				continue
			}
			if x < r.Min.X {
				r.Min.X = x
			}
			if y < r.Min.Y {
				r.Min.Y = y
			}
			if x+1 > r.Max.X {
				r.Max.X = x + 1
			}
			if y+1 > r.Max.Y {
				r.Max.Y = y + 1
			}
		}
	}
	return r
}

func TestPainterTextAlignShifts(t *testing.T) {
	ink := NewColor(0, 0, 0, 255)
	at := NewPoint(64, 32)
	paint := func(a TextAlign) image.Rectangle {
		p := NewPainter(NewSize(128, 64))
		p.Font().SetColor(ink)
		p.Font().SetSize(24)
		p.Font().SetAlign(a)
		p.Text("A", at)
		r := paintedBBox(painterRGBA(p))
		if r.Empty() {
			t.Fatalf("align %#x: no paint", a)
		}
		return r
	}
	left := paint(TextAlignLeftBaseline)
	right := paint(TextAlignRightBaseline)
	if left.Min.X <= right.Min.X {
		t.Fatalf("left minX %d should be > right minX %d", left.Min.X, right.Min.X)
	}
	top := paint(TextAlignCenterTop)
	bot := paint(TextAlignCenterBottom)
	if top.Min.Y <= bot.Min.Y {
		t.Fatalf("top-aligned ink should sit lower (minY %d vs bottom minY %d)", top.Min.Y, bot.Min.Y)
	}
	if top.Min.Y < 28 {
		t.Fatalf("top minY %d should be at or below origin 32", top.Min.Y)
	}
	if bot.Max.Y > 34 {
		t.Fatalf("bottom maxY %d should be at or above origin 32", bot.Max.Y)
	}
}

func TestPainterTextMultilinePaintsTaller(t *testing.T) {
	ink := NewColor(0, 0, 0, 255)
	at := NewPoint(64, 32)
	paint := func(s string) image.Rectangle {
		p := NewPainter(NewSize(128, 80))
		p.Font().SetColor(ink)
		p.Font().SetSize(18)
		p.Font().SetAlign(TextAlignCenterMiddle)
		p.Text(s, at)
		r := paintedBBox(painterRGBA(p))
		if r.Empty() {
			t.Fatalf("%q: no paint", s)
		}
		return r
	}
	one := paint("A")
	two := paint("A\nA")
	if two.Dy() <= one.Dy() {
		t.Fatalf("multiline height %d, single %d", two.Dy(), one.Dy())
	}
}

func TestPainterTextWrapWrapsAndPaints(t *testing.T) {
	ink := NewColor(0, 0, 0, 255)
	at := NewPoint(16, 24)
	p := NewPainter(NewSize(160, 96))
	p.Font().SetColor(ink)
	p.Font().SetSize(16)
	p.Font().SetAlign(TextAlignLeftBaseline)
	p.TextWrap("word wrap sample text here", at, 70)
	r := paintedBBox(painterRGBA(p))
	if r.Empty() {
		t.Fatal("textWrap: no paint")
	}
	if r.Dy() < 20 {
		t.Fatalf("textWrap should be taller than one line, dy=%d", r.Dy())
	}
}

func TestPainterTextWrapRejectsNonPositiveWidth(t *testing.T) {
	p := NewPainter(NewSize(32, 32))
	defer func() {
		if recover() == nil {
			t.Fatal("width 0 should panic")
		}
	}()
	p.TextWrap("x", NewPoint(8, 8), 0)
}

func TestPainterTextWrapNativeMethod(t *testing.T) {
	typ := GetTypeRegistrar().Get("Painter")
	p := NewPainter(NewSize(160, 64))
	p.Font().SetColor(NewColor(0, 0, 0, 255))
	p.Font().SetAlign(TextAlignLeftBaseline)
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, p)
	rp.DataStack.Push(String("wrap this phrase please"))
	rp.DataStack.Push(NewPoint(8, 20))
	rp.DataStack.Push(Int(60))
	typ.MethodByName("textWrap").invokeCall(f, ths)
	if countPainted(painterRGBA(p)) == 0 {
		t.Fatal("Painter.textWrap native: no paint")
	}
}

func TestPainterPushPopRestoresAlign(t *testing.T) {
	p := NewPainter(NewSize(16, 16))
	if p.Font().Align() != TextAlignCenterMiddle {
		t.Fatal("default")
	}
	p.Push()
	p.Font().SetAlign(TextAlignLeftBaseline)
	p.Pop()
	if p.Font().Align() != TextAlignCenterMiddle {
		t.Fatalf("after pop %#x", p.Font().Align())
	}
}
