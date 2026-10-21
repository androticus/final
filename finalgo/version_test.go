package finalgo_test

import (
	"strings"
	"testing"

	"github.com/androticus/final/finalgo"
)

func TestVersionAndCopyright(t *testing.T) {
	v := finalgo.Version()
	if v == "" {
		t.Fatal("empty Version")
	}
	c := finalgo.Copyright()
	if !strings.Contains(c, "Brad Aisa") {
		t.Fatalf("Copyright = %q", c)
	}
	about := "Final version: " + v + "\n" + c
	if !strings.HasPrefix(about, "Final version: ") {
		t.Fatalf("about = %q", about)
	}
}
