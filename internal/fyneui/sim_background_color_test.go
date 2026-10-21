package fyneui

import (
	"testing"

	"github.com/androticus/final/internal/core"
)

func TestSimGame_BackgroundColorDefaultAndSet(t *testing.T) {
	sg := &SimGame{backgroundColor: core.WebColors[core.ColorBlack]}
	wantBlack := core.NewColor(0, 0, 0, 255)
	if sg.backgroundColor != wantBlack {
		t.Fatalf("default want %v, got %v", wantBlack, sg.backgroundColor)
	}
	navy := core.WebColors[core.ColorNavy]
	sg.backgroundColor = navy
	if sg.backgroundColor != navy {
		t.Fatal("set failed")
	}
}

func TestSimGame_BackgroundColorPropertyRegistered(t *testing.T) {
	if simGameType.GetPropertyMetadata("backgroundColor") == nil {
		t.Fatal("missing backgroundColor property")
	}
}
