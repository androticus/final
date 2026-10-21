package core

import (
	"image/color"
	"testing"
)

func colorFromHexOnStack(v Any) Color {
	ds := NewStack[Any]("test", 4)
	ds.Push(v)
	return colorFromHex(ds)
}

func colorFromHexAOnStack(v Any) Color {
	ds := NewStack[Any]("test", 4)
	ds.Push(v)
	return colorFromHexA(ds)
}

func TestColorCyanMagentaClassConstsAreColors(t *testing.T) {
	// Regression: cyan/magenta aliases must return WebColors entries, not iota indices
	// (ColorAqua=15 / ColorFuchsia=7), which broke lists like ( Color.cyan Color.magenta … ).
	for _, tc := range []struct {
		name string
		want Color
	}{
		{"cyan", WebColors[ColorAqua]},
		{"magenta", WebColors[ColorFuchsia]},
		{"aqua", WebColors[ColorAqua]},
		{"fuchsia", WebColors[ColorFuchsia]},
	} {
		pm := colorType.GetClassPropertyMetadata(tc.name)
		if pm == nil {
			t.Fatalf("missing Color.%s", tc.name)
		}
		got := pm.get(colorType)
		c, ok := got.(Color)
		if !ok {
			t.Fatalf("Color.%s => %T %#v, want Color", tc.name, got, got)
		}
		if c != tc.want {
			t.Fatalf("Color.%s = %#v, want %#v", tc.name, c, tc.want)
		}
	}
}

func TestColorFromPremulRGBAPreservesStoreBytes(t *testing.T) {
	px := color.RGBA{R: 100, G: 50, B: 25, A: 128}
	c := colorFromPremulRGBA(px)
	if c.rp != 100 || c.gp != 50 || c.bp != 25 || c.ap != 128 {
		t.Fatalf("premul = (%d,%d,%d,%d), want store bytes", c.rp, c.gp, c.bp, c.ap)
	}
	if c.A != 128 || c.R != uint8(uint16(100)*255/128) {
		t.Fatalf("straight = (%d,_,_,%d)", c.R, c.A)
	}
	opaque := colorFromPremulRGBA(color.RGBA{R: 10, G: 20, B: 30, A: 255})
	if opaque.R != 10 || opaque.G != 20 || opaque.B != 30 || opaque.A != 255 {
		t.Fatalf("opaque straight = (%d,%d,%d,%d)", opaque.R, opaque.G, opaque.B, opaque.A)
	}
	if opaque.rp != 10 || opaque.ap != 255 {
		t.Fatalf("opaque premul mutated")
	}
	clear := colorFromPremulRGBA(color.RGBA{})
	if clear.R != 0 || clear.A != 0 || clear.rp != 0 {
		t.Fatalf("A=0 = %#v", clear)
	}
}

func TestColorFromHex(t *testing.T) {
	c := colorFromHexOnStack(Int(0xFF8000))
	if c.R != 255 || c.G != 128 || c.B != 0 || c.A != 255 {
		t.Fatalf("hex 0xFF8000 = rgba(%d,%d,%d,%d), want (255,128,0,255)", c.R, c.G, c.B, c.A)
	}
	blue := colorFromHexOnStack(Int(0x0000FF))
	if blue.R != 0 || blue.G != 0 || blue.B != 255 || blue.A != 255 {
		t.Fatalf("hex 0x0000FF = rgba(%d,%d,%d,%d), want (0,0,255,255)", blue.R, blue.G, blue.B, blue.A)
	}
	c = colorFromHexOnStack(String("FF8000"))
	if c.R != 255 || c.G != 128 || c.B != 0 || c.A != 255 {
		t.Fatalf("hex \"FF8000\" = rgba(%d,%d,%d,%d)", c.R, c.G, c.B, c.A)
	}
	c = colorFromHexOnStack(String("0x0000ff"))
	if c.R != 0 || c.G != 0 || c.B != 255 || c.A != 255 {
		t.Fatalf("hex \"0x0000ff\" = rgba(%d,%d,%d,%d)", c.R, c.G, c.B, c.A)
	}
}

func TestColorFromHexA(t *testing.T) {
	c := colorFromHexAOnStack(Int(0xFF8000CC))
	if c.R != 255 || c.G != 128 || c.B != 0 || c.A != 204 {
		t.Fatalf("hexa 0xFF8000CC = rgba(%d,%d,%d,%d), want (255,128,0,204)", c.R, c.G, c.B, c.A)
	}
	clear := colorFromHexAOnStack(Int(0x00FF0000))
	if clear.R != 0 || clear.G != 255 || clear.B != 0 || clear.A != 0 {
		t.Fatalf("hexa 0x00FF0000 = rgba(%d,%d,%d,%d), want straight (0,255,0,0)", clear.R, clear.G, clear.B, clear.A)
	}
	if clear.rp != 0 || clear.gp != 0 || clear.bp != 0 || clear.ap != 0 {
		t.Fatalf("hexa A=0 premul = (%d,%d,%d,%d), want zeros", clear.rp, clear.gp, clear.bp, clear.ap)
	}
	c = colorFromHexAOnStack(String("FF8000CC"))
	if c.A != 204 {
		t.Fatalf("hexa \"FF8000CC\" a=%d", c.A)
	}
	c = colorFromHexAOnStack(String("0X00ff0000"))
	if c.G != 255 || c.A != 0 {
		t.Fatalf("hexa \"0X00ff0000\" = (%d,%d,%d,%d)", c.R, c.G, c.B, c.A)
	}
}

func TestNewColorAZeroPreservesStraightRGB(t *testing.T) {
	c := NewColor(200, 100, 50, 0)
	if c.R != 200 || c.G != 100 || c.B != 50 || c.A != 0 {
		t.Fatalf("straight = (%d,%d,%d,%d), want (200,100,50,0)", c.R, c.G, c.B, c.A)
	}
	if c.rp != 0 || c.gp != 0 || c.bp != 0 || c.ap != 0 {
		t.Fatalf("premul = (%d,%d,%d,%d), want zeros", c.rp, c.gp, c.bp, c.ap)
	}
}

func TestColorHexClassMethods(t *testing.T) {
	typ := GetTypeRegistrar().Get("Color")
	if typ == nil {
		t.Fatal("Color type not registered")
	}

	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, typ)
	mm := typ.ClassMethodByName("hex")
	if mm == nil {
		t.Fatal("missing Color.hex")
	}
	rp.DataStack.Push(Int(0x336699))
	mm.invokeCall(f, ths)
	c := rp.DataStack.Pop().(Color)
	if c.R != 0x33 || c.G != 0x66 || c.B != 0x99 || c.A != 255 {
		t.Fatalf("Color.hex 0x336699 = rgba(%d,%d,%d,%d)", c.R, c.G, c.B, c.A)
	}
	rp.DataStack.Push(String("336699"))
	mm.invokeCall(f, ths)
	c = rp.DataStack.Pop().(Color)
	if c.R != 0x33 || c.G != 0x66 || c.B != 0x99 || c.A != 255 {
		t.Fatalf("Color.hex \"336699\" = rgba(%d,%d,%d,%d)", c.R, c.G, c.B, c.A)
	}

	mm = typ.ClassMethodByName("hexa")
	if mm == nil {
		t.Fatal("missing Color.hexa")
	}
	rp.DataStack.Push(Int(0x33669980))
	mm.invokeCall(f, ths)
	c = rp.DataStack.Pop().(Color)
	if c.R != 0x33 || c.G != 0x66 || c.B != 0x99 || c.A != 0x80 {
		t.Fatalf("Color.hexa 0x33669980 = rgba(%d,%d,%d,%d)", c.R, c.G, c.B, c.A)
	}
	rp.DataStack.Push(String("0x33669980"))
	mm.invokeCall(f, ths)
	c = rp.DataStack.Pop().(Color)
	if c.A != 0x80 {
		t.Fatalf("Color.hexa \"0x33669980\" a=%d", c.A)
	}
}

func TestNewColorPremulOpaque(t *testing.T) {
	c := NewColor(200, 100, 50, 255)
	if c.rp != 200 || c.gp != 100 || c.bp != 50 || c.ap != 255 {
		t.Fatalf("opaque premul = (%d,%d,%d,%d), want (200,100,50,255)", c.rp, c.gp, c.bp, c.ap)
	}
	r, g, b, a := c.RGBA()
	wantR, wantG, wantB, wantA := color.RGBA{R: 200, G: 100, B: 50, A: 255}.RGBA()
	if r != wantR || g != wantG || b != wantB || a != wantA {
		t.Fatalf("RGBA() opaque = %x,%x,%x,%x, want %x,%x,%x,%x", r, g, b, a, wantR, wantG, wantB, wantA)
	}
}

func TestNewColorPremulTranslucent(t *testing.T) {
	c := NewColor(200, 100, 50, 128)
	wantRp := uint8(200 * 128 / 255)
	if c.rp != wantRp || c.ap != 128 {
		t.Fatalf("translucent premul rp=%d ap=%d, want rp=%d ap=128", c.rp, c.ap, wantRp)
	}
}

func TestColorRGBAConvertsLikeRGBAModel(t *testing.T) {
	c := NewColor(200, 100, 50, 128)
	got := color.RGBAModel.Convert(c)
	want := color.RGBA{R: c.rp, G: c.gp, B: c.bp, A: c.ap}
	if got != want {
		t.Fatalf("RGBAModel.Convert = %v, want %v", got, want)
	}
}
