package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/androticus/final/internal/core"
)

func TestCanvasImageFromPainterPixmap(t *testing.T) {
	p := core.NewPainter(core.NewSize(16, 16))
	p.Stroke().SetColor(core.NewColor(255, 0, 0, 255))
	p.Line(core.NewPoint(0, 8), core.NewPoint(15, 8))
	img := NewCanvasImage(p.Pixmap().Image())
	if img == nil || img.Image == nil {
		t.Fatal("CanvasImage missing image")
	}
	if img.Image.Image.Bounds() != p.Pixmap().Image().Bounds() {
		t.Fatalf("CanvasImage bounds %v, want %v", img.Image.Image.Bounds(), p.Pixmap().Image().Bounds())
	}
}

func TestCanvasImageMinSizeIsPixmapPixels(t *testing.T) {
	p := core.NewPainter(core.NewSize(64, 48))
	img := NewCanvasImage(p.Pixmap().Image())
	if img.Image.FillMode != canvas.ImageFillStretch {
		t.Fatalf("FillMode %v, want Stretch", img.Image.FillMode)
	}
	if img.Image.ScaleMode != canvas.ImageScalePixels {
		t.Fatalf("ScaleMode %v, want Pixels", img.Image.ScaleMode)
	}
	want := fyne.NewSize(64, 48)
	if got := img.MinSize(); got != want {
		t.Fatalf("MinSize %v, want %v", got, want)
	}
	if got := img.Image.MinSize(); got != want {
		t.Fatalf("inner MinSize %v, want %v", got, want)
	}
}

func TestCanvasImageNilMethodRefDoesNotMint(t *testing.T) {
	if mintMouseEventFn(nil, nil, 0) != nil {
		t.Fatal("nil down/up/in/moved handler should not mint")
	}
	if mintMouseOutFn(nil, nil, 0) != nil {
		t.Fatal("nil out handler should not mint")
	}
	if mintScrollEventFn(nil, nil, 0) != nil {
		t.Fatal("nil scrolled handler should not mint")
	}
}
