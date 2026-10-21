package core

import (
	"testing"
)

func TestNewPainterDefaultFont(t *testing.T) {
	p := NewPainter(NewSize(32, 32))
	f := p.Font()
	if f == nil {
		t.Fatal("Painter.font is nil")
	}
	if f.Family() != FontFamilyGo {
		t.Fatalf("family %q, want %q", f.Family(), FontFamilyGo)
	}
	if f.Size() != defaultFontSize {
		t.Fatalf("size %v, want %v", f.Size(), defaultFontSize)
	}
	if f.Align() != TextAlignCenterMiddle {
		t.Fatalf("align %#x, want centerMiddle %#x", f.Align(), TextAlignCenterMiddle)
	}
	pm := GetTypeRegistrar().Get("Painter").GetPropertyMetadata("font")
	if pm == nil || pm.set != nil {
		t.Fatal("Painter.font should be readonly const")
	}
}

func TestPainterTextPaintsPixels(t *testing.T) {
	p := NewPainter(NewSize(128, 64))
	p.Font().SetColor(NewColor(255, 0, 0, 255))
	p.Text("Hello", NewPoint(8, 40))
	if countPainted(painterRGBA(p)) == 0 {
		t.Fatal("text: no paint")
	}
}

func TestPainterTextSizeChangePaintsMore(t *testing.T) {
	small := NewPainter(NewSize(128, 64))
	small.Font().SetColor(NewColor(0, 0, 0, 255))
	small.Font().SetSize(10)
	small.Text("Ag", NewPoint(8, 40))
	nSmall := countPainted(painterRGBA(small))

	large := NewPainter(NewSize(128, 64))
	large.Font().SetColor(NewColor(0, 0, 0, 255))
	large.Font().SetSize(28)
	large.Text("Ag", NewPoint(8, 40))
	nLarge := countPainted(painterRGBA(large))
	if nLarge <= nSmall {
		t.Fatalf("larger size painted %d, small %d", nLarge, nSmall)
	}
}

func TestFontNamesAndConsts(t *testing.T) {
	typ := GetTypeRegistrar().Get("Font")
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, typ)
	typ.ClassMethodByName("names").invokeCall(f, ths)
	names := rp.DataStack.Pop().(*StringList)
	if names.Len() != len(fontFamilyNames) {
		t.Fatalf("Font.names len %d, want %d", names.Len(), len(fontFamilyNames))
	}
	for _, want := range fontFamilyNames {
		if !names.HasItemFromAny(want) {
			t.Fatalf("Font.names missing %q", want)
		}
		pm := typ.GetClassPropertyMetadata(want)
		if pm == nil || pm.get == nil {
			t.Fatalf("Font.%s class const missing", want)
		}
		got := pm.get(nil).(String)
		if string(got) != want {
			t.Fatalf("Font.%s = %q", want, got)
		}
	}
}

func TestFontFamilyRejectsUnknown(t *testing.T) {
	f := NewPainter(NewSize(8, 8)).Font()
	defer func() {
		if recover() == nil {
			t.Fatal("unknown family should panic")
		}
	}()
	f.SetFamily("Arial")
}

func TestFontSizeRejectsNonPositive(t *testing.T) {
	f := NewPainter(NewSize(8, 8)).Font()
	defer func() {
		if recover() == nil {
			t.Fatal("size 0 should panic")
		}
	}()
	f.SetSize(0)
}

func TestFontSizeNumericCoercion(t *testing.T) {
	typ := GetTypeRegistrar().Get("Font")
	pm := typ.GetPropertyMetadata("size")
	f := NewPainter(NewSize(8, 8)).Font()
	pm.set(f, Int(18))
	if f.Size() != 18 {
		t.Fatalf("size from Int %v", f.Size())
	}
	pm.set(f, Float(20.5))
	if f.Size() != 20.5 {
		t.Fatalf("size from Float %v", f.Size())
	}
	pm.set(f, Uint(16))
	if f.Size() != 16 {
		t.Fatalf("size from Uint %v", f.Size())
	}
}

func TestPainterTextNativeMethod(t *testing.T) {
	typ := GetTypeRegistrar().Get("Painter")
	p := NewPainter(NewSize(96, 48))
	p.Font().SetColor(NewColor(0, 0, 255, 255))
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, p)
	rp.DataStack.Push(String("Hi"))
	rp.DataStack.Push(NewPoint(6, 32))
	typ.MethodByName("text").invokeCall(f, ths)
	if countPainted(painterRGBA(p)) == 0 {
		t.Fatal("Painter.text native: no paint")
	}
}
