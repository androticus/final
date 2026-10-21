package finaltui

import "testing"

func TestNormalizeOpenPath(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{`C:\proj\fib.fnl`, `C:\proj\fib.fnl`},
		{`"C:\proj\fib.fnl"`, `C:\proj\fib.fnl`},
		{`'C:\proj\fib.fnl'`, `C:\proj\fib.fnl`},
		{`  "C:\proj\fib.fnl"  `, `C:\proj\fib.fnl`},
		{`"C:\proj\fib.fnl"` + "\r\n", `C:\proj\fib.fnl`},
		{``, ``},
		{`""`, ``},
	}
	for _, c := range cases {
		if got := normalizeOpenPath(c.in); got != c.want {
			t.Errorf("normalizeOpenPath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCompilePathScriptOpenFlows(t *testing.T) {
	got, err := compilePathScript(`examples/threading/tick-tick-boom.fnl`)
	if err != nil {
		t.Fatal(err)
	}
	want := "`examples/threading/tick-tick-boom.fnl` compileFromFile"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}

	path := normalizeOpenPath(`"C:\proj\fib.fnl"`)
	got, err = compilePathScript(path)
	if err != nil {
		t.Fatal(err)
	}
	want = "`C:/proj/fib.fnl` compileFromFile"
	if got != want {
		t.Fatalf("quoted path: got %q want %q", got, want)
	}

	if _, err := compilePathScript("bad`path"); err == nil {
		t.Fatal("expected error for backtick in path")
	}
}
