package core

import (
	"image"
	"testing"
)

func testDS() *Stack[any] {
	return NewStack[any]("ds", 16)
}

func TestPopPointObject(t *testing.T) {
	ds := testDS()
	ds.Push(NewPoint(3, 7))
	p := popPoint(ds)
	if p.X != 3 || p.Y != 7 {
		t.Fatalf("popPoint(object) = %v, want (3, 7)", p)
	}
	if ds.Len() != 0 {
		t.Fatalf("ds leftover %d", ds.Len())
	}
}

func TestPopPointTwoInts(t *testing.T) {
	ds := testDS()
	ds.Push(Int(3))
	ds.Push(Int(7))
	p := popPoint(ds)
	if p.X != 3 || p.Y != 7 {
		t.Fatalf("popPoint(ints) = %v, want (3, 7)", p)
	}
	if ds.Len() != 0 {
		t.Fatalf("ds leftover %d", ds.Len())
	}
}

func TestPopSizeObject(t *testing.T) {
	ds := testDS()
	ds.Push(NewSize(10, 20))
	s := popSize(ds)
	if s.W != 10 || s.H != 20 {
		t.Fatalf("popSize(object) = %v, want 10x20", s)
	}
	if ds.Len() != 0 {
		t.Fatalf("ds leftover %d", ds.Len())
	}
}

func TestPopSizeTwoInts(t *testing.T) {
	ds := testDS()
	ds.Push(Int(10))
	ds.Push(Int(20))
	s := popSize(ds)
	if s.W != 10 || s.H != 20 {
		t.Fatalf("popSize(ints) = %v, want 10x20", s)
	}
	if ds.Len() != 0 {
		t.Fatalf("ds leftover %d", ds.Len())
	}
}

func TestPopSizeFloats(t *testing.T) {
	ds := testDS()
	ds.Push(Float(10.5))
	ds.Push(Float(20.25))
	s := popSize(ds)
	if s.W != 10.5 || s.H != 20.25 {
		t.Fatalf("popSize(floats) = %v, want 10.5x20.25", s)
	}
}

func TestPopSizeUint(t *testing.T) {
	ds := testDS()
	ds.Push(Uint(4))
	ds.Push(Uint(5))
	s := popSize(ds)
	if s.W != 4 || s.H != 5 {
		t.Fatalf("popSize(uints) = %v, want 4x5", s)
	}
}

func TestPointToImagePoint(t *testing.T) {
	got := NewPoint(4, 9).ToImagePoint()
	want := image.Point{X: 4, Y: 9}
	if got != want {
		t.Fatalf("ToImagePoint = %v, want %v", got, want)
	}
}

func TestRectToImageRect(t *testing.T) {
	r := NewRect(NewPoint(10, 20), NewSize(30, 40))
	got := r.ToImageRect()
	want := image.Rect(10, 20, 40, 60)
	if got != want {
		t.Fatalf("ToImageRect = %v, want %v", got, want)
	}
}

func TestRectToImageRectRejectsFractionalSize(t *testing.T) {
	r := NewRect(NewPoint(0, 0), NewSizeF(10.5, 20))
	defer func() {
		if recover() == nil {
			t.Fatal("ToImageRect fractional width should panic")
		}
	}()
	_ = r.ToImageRect()
}

func TestRequireIntegralDim(t *testing.T) {
	if got := RequireIntegralDim("t", 10); got != 10 {
		t.Fatalf("RequireIntegralDim(10) = %d", got)
	}
	if got := RequireIntegralDim("t", 10.0); got != 10 {
		t.Fatalf("RequireIntegralDim(10.0) = %d", got)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("RequireIntegralDim(10.5) should panic")
		}
	}()
	_ = RequireIntegralDim("t", 10.5)
}

func TestRectNewPopsPointAndSizeSlots(t *testing.T) {
	ds := testDS()
	ds.Push(Int(10))
	ds.Push(Int(20))
	ds.Push(Int(30))
	ds.Push(Int(40))
	sz := popSize(ds)
	tl := popPoint(ds)
	r := NewRect(tl, sz)
	if r.TopLeft.X != 10 || r.TopLeft.Y != 20 || r.Size.W != 30 || r.Size.H != 40 {
		t.Fatalf("Rect from four ints = %v", r)
	}
}

func TestDrawAdjunctTypesRegistered(t *testing.T) {
	for _, name := range []string{"Point", "Size", "Rect"} {
		if GetTypeRegistrar().Get(name) == nil {
			t.Fatalf("%s type not registered", name)
		}
	}
	if NewPoint(1, 2).FinalType() != pointType {
		t.Fatal("Point.FinalType mismatch")
	}
	if NewSize(1, 2).FinalType() != sizeType {
		t.Fatal("Size.FinalType mismatch")
	}
	if NewRect(NewPoint(0, 0), NewSize(1, 1)).FinalType() != rectType {
		t.Fatal("Rect.FinalType mismatch")
	}
}

func TestSizeWiHiRound(t *testing.T) {
	s := NewSizeF(10.4, 20.6)
	if s.Wi() != 10 || s.Hi() != 21 {
		t.Fatalf("Wi/Hi = %d/%d, want 10/21", s.Wi(), s.Hi())
	}
}
