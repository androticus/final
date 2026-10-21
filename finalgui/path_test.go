package finalgui

import "testing"

func TestCompilePathScript(t *testing.T) {
	got, err := compilePathScript(`examples/threading/tick-tick-boom.fnl`)
	if err != nil {
		t.Fatal(err)
	}
	want := "`examples/threading/tick-tick-boom.fnl` compileFromFile"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}

	got, err = compilePathScript(`C:\proj\fib.fnl`)
	if err != nil {
		t.Fatal(err)
	}
	want = "`C:/proj/fib.fnl` compileFromFile"
	if got != want {
		t.Fatalf("windows path: got %q want %q", got, want)
	}

	got, err = compilePathScript("pkg/package.fnlm")
	if err != nil {
		t.Fatal(err)
	}
	want = "`pkg/package.fnlm` compileFromManifest"
	if got != want {
		t.Fatalf("fnlm: got %q want %q", got, want)
	}

	got, err = compilePathScript("proj.fnlp")
	if err != nil {
		t.Fatal(err)
	}
	want = "`proj.fnlp` loadProject"
	if got != want {
		t.Fatalf("fnlp: got %q want %q", got, want)
	}

	if _, err := compilePathScript("bad`path"); err == nil {
		t.Fatal("expected error for backtick in path")
	}
	if _, err := compilePathScript("  "); err == nil {
		t.Fatal("expected error for empty path")
	}
}
