package fyneui

import (
	"image"
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/androticus/final/internal/core"
)

func TestSimWindowContentSizeFloorsToPixSize(t *testing.T) {
	body := canvas.NewRectangle(color.Black)
	sz := simWindowContentSize(body, core.NewSize(768, 576))
	if sz.Width < 768 || sz.Height < 576 {
		t.Fatalf("size %v, want at least 768×576", sz)
	}
}

func TestSimWindowContentSizeIncludesBottomBar(t *testing.T) {
	img := NewCanvasImage(image.NewRGBA(image.Rect(0, 0, 768, 576)))
	bar := canvas.NewRectangle(color.White)
	bar.SetMinSize(fyne.NewSize(400, 48))
	content := container.NewBorder(nil, bar, nil, nil, img)
	sz := simWindowContentSize(content, core.NewSize(768, 576))
	if sz.Width < 768 {
		t.Fatalf("width %v, want ≥ 768", sz.Width)
	}
	if sz.Height < 576+48 {
		t.Fatalf("height %v, want pixmap+bar", sz.Height)
	}
}
