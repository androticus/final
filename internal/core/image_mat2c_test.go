package core

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"
)

func TestMat2CImageRGBAModel(t *testing.T) {
	m := NewMat2C(1, 2)
	ink := NewColor(10, 20, 30, 255)
	m.PutFlat(1, ink)
	if m.ColorModel() != color.RGBAModel {
		t.Fatal("ColorModel want RGBAModel")
	}
	want := colorToPremulRGBA(ink)
	if got := m.At(1, 0); got != want {
		t.Fatalf("image At(1,0)=%v, want %v", got, want)
	}
	if got := m.At_(0, 1); got != ink {
		t.Fatalf("matrix At_(0,1)=%v, want %v", got, ink)
	}
}

func TestMat2CImagePngEncode(t *testing.T) {
	m := NewMat2C(2, 2)
	m.Put(0, 0, NewColor(255, 0, 0, 255))
	m.Put(0, 1, NewColor(0, 255, 0, 255))
	m.Put(1, 0, NewColor(0, 0, 255, 255))
	m.Put(1, 1, NewColor(255, 255, 0, 255))
	var buf bytes.Buffer
	if err := png.Encode(&buf, m); err != nil {
		t.Fatal(err)
	}
	dec, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if dec.Bounds().Dx() != 2 || dec.Bounds().Dy() != 2 {
		t.Fatalf("decoded bounds %v", dec.Bounds())
	}
	got := color.RGBAModel.Convert(dec.At(1, 0))
	want := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	if got != want {
		t.Fatalf("decoded (0,1)=%v, want %v", got, want)
	}
}
