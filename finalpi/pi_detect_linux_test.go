//go:build linux

package finalpi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDeviceTreeModelFromTempFiles(t *testing.T) {
	dir := t.TempDir()
	primary := filepath.Join(dir, "model")
	if err := os.WriteFile(primary, []byte("Raspberry Pi Zero 2 W Rev 1.0\x00"), 0o644); err != nil {
		t.Fatal(err)
	}
	prev := deviceTreeModelPaths
	deviceTreeModelPaths = []string{primary, filepath.Join(dir, "missing")}
	defer func() { deviceTreeModelPaths = prev }()

	got := readDeviceTreeModel()
	if got != "Raspberry Pi Zero 2 W Rev 1.0" {
		t.Fatalf("readDeviceTreeModel()=%q", got)
	}
	if mapDeviceTreeModel(got) != PiSupportPiZ2 {
		t.Fatalf("expected Pi2Z from temp model file")
	}
}
