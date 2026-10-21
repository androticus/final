package finaltui

import (
	"runtime"
	"testing"
)

func TestPlatformManifestPath(t *testing.T) {
	want := "extras/platform/" + runtime.GOOS + "/tview.fnlm"
	if got := PlatformManifestPath(false, ""); got != want {
		t.Fatalf("headed: got %q want %q", got, want)
	}
	want = "extras/platform/" + runtime.GOOS + "/headless.fnlm"
	if got := PlatformManifestPath(true, ""); got != want {
		t.Fatalf("headless: got %q want %q", got, want)
	}
	want = "extras/platform/linux/pi/tview.fnlm"
	if got := PlatformManifestPath(false, "pi"); got != want {
		t.Fatalf("qual pi: got %q want %q", got, want)
	}
}

func TestCompilePathScript(t *testing.T) {
	s, err := compilePathScript("extras/platform/darwin/tview.fnlm")
	if err != nil {
		t.Fatal(err)
	}
	if s != "`extras/platform/darwin/tview.fnlm` compileFromManifest" {
		t.Fatalf("got %q", s)
	}
	s, err = compilePathScript("foo/bar.fnl")
	if err != nil {
		t.Fatal(err)
	}
	if s != "`foo/bar.fnl` compileFromFile" {
		t.Fatalf("got %q", s)
	}
	s, err = compilePathScript("apps/demo.fnlp")
	if err != nil {
		t.Fatal(err)
	}
	if s != "`apps/demo.fnlp` loadProject" {
		t.Fatalf("got %q", s)
	}
	if _, err := compilePathScript("bad`path.fnl"); err == nil {
		t.Fatal("expected backtick error")
	}
}
