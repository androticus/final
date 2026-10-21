package core

import "testing"

func TestPalette8LenAndItems(t *testing.T) {
	ResetTestRuntime()
	p := NewPalette8FromSlice([]Color{NewColor(1, 2, 3, 255)})
	if p.Len() != 256 {
		t.Fatalf("len=%d, want 256", p.Len())
	}
	if p.FinalType() != palette8Type {
		t.Fatalf("FinalType = %v, want Palette8", p.FinalType())
	}
	if p.At(0) != NewColor(1, 2, 3, 255) {
		t.Fatalf("At(0)=%v", p.At(0))
	}
	if p.At(1) != (Color{}) {
		t.Fatalf("unset slot want zero, got %v", p.At(1))
	}
	ink := NewColor(9, 8, 7, 255)
	p.Put(255, ink)
	if p.At(255) != ink {
		t.Fatalf("At(255)=%v, want %v", p.At(255), ink)
	}
	im := palette8Type.GetItemsMetadata()
	im.SetByOrdinal(p, Int(10), ink)
	if got := im.GetByOrdinal(p, Int(10)).(Color); got != ink {
		t.Fatalf("@() = %v, want %v", got, ink)
	}
}

func TestNew332Palette8(t *testing.T) {
	ResetTestRuntime()
	p := New332Palette8(255)
	if p.Len() != 256 {
		t.Fatalf("len=%d, want 256", p.Len())
	}
	want := func(r, g, b int) Color { return NewColor(r, g, b, 255) }
	cases := []struct {
		i    int
		want Color
	}{
		{0x00, want(0, 0, 0)},
		{0xE0, want(255, 0, 0)},       // RRR.....
		{0x1C, want(0, 255, 0)},       // ...GGG..
		{0x03, want(0, 0, 255)},       // ......BB
		{0x20, want(36, 0, 0)},        // r=1
		{0x04, want(0, 36, 0)},        // g=1
		{0x01, want(0, 0, 85)},        // b=1
		{0xFF, want(255, 255, 255)},
	}
	for _, c := range cases {
		if got := p.At(c.i); got != c.want {
			t.Errorf("At(%#02x)=%v, want %v", c.i, got, c.want)
		}
	}
	p128 := New332Palette8(128)
	if got := p128.At(0xFF); got != NewColor(255, 255, 255, 128) {
		t.Fatalf("alpha 128 At(0xFF)=%v", got)
	}
}

func TestPalette4FromSliceAlpha(t *testing.T) {
	ResetTestRuntime()
	p := NewPalette4FromSlice(WebColors, 128)
	if p.Len() != 16 {
		t.Fatalf("len=%d, want 16", p.Len())
	}
	if p.FinalType() != palette4Type {
		t.Fatalf("FinalType = %v, want Palette4", p.FinalType())
	}
	src := WebColors[ColorRed]
	got := p.At(int(ColorRed))
	want := NewColor(int(src.R), int(src.G), int(src.B), 128)
	if got != want {
		t.Fatalf("At(red)=%v, want %v", got, want)
	}
}
