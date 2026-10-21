package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMat2C_Crop(t *testing.T) {
	m := NewMat2C(4, 6)
	ink := NewColor(1, 2, 3, 255)
	m.Put(2, 3, ink)
	c := m.Crop(1, 2, 2, 3)
	if c.Rows() != 2 || c.Cols() != 3 {
		t.Fatalf("crop shape %dx%d", c.Rows(), c.Cols())
	}
	if got := c.At_(1, 1); got != ink {
		t.Fatalf("crop cell = %v want %v", got, ink)
	}
}

func TestMat2C_FromFileAlien(t *testing.T) {
	path := filepath.Join("..", "final", "examples", "gui", "simgame", "assets", "alien-a.png")
	if _, err := os.Stat(path); err != nil {
		t.Skip(err)
	}
	m, err := NewMat2CFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.Rows() != 24 || m.Cols() != 32 {
		t.Fatalf("size %dx%d", m.Rows(), m.Cols())
	}
}

func TestAnyMat_NilCellAndIter(t *testing.T) {
	ResetTestRuntime()
	m := NewAnyMat(2, 2)
	m.Put(0, 0, String("a"))
	m.Put(0, 1, nil)
	if m.At_(0, 1) != nil {
		t.Fatal("nil cell")
	}
	if m.FinalType() != anyMatType || anyMatType.Name() != "Mat" {
		t.Fatalf("FinalType %v name %q", m.FinalType(), anyMatType.Name())
	}
	it := m.GetIter().(*Mat2Iter[any])
	if !it.GetNext() {
		t.Fatal("first")
	}
	if it.Current() != String("a") {
		t.Fatalf("current=%v", it.Current())
	}
	if !it.GetNext() {
		t.Fatal("second")
	}
	if it.Current() != nil {
		t.Fatalf("want nil, got %v", it.Current())
	}
}
