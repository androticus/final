package core

import "testing"

func TestParseModuleRef(t *testing.T) {
	cases := []struct {
		in, name, ver string
	}{
		{"", "", ""},
		{"  modules/csv  ", "modules/csv", ""},
		{"modules/csv@", "modules/csv", ""},
		{"modules/csv@v0.1.0", "modules/csv", "v0.1.0"},
		{"  github.com/a/b@ v1.2.3 ", "github.com/a/b", "v1.2.3"},
	}
	for _, tc := range cases {
		n, v := ParseModuleRef(tc.in)
		if n != tc.name || v != tc.ver {
			t.Fatalf("ParseModuleRef(%q) = (%q,%q), want (%q,%q)", tc.in, n, v, tc.name, tc.ver)
		}
	}
}

func TestFormatModuleRef(t *testing.T) {
	if got := FormatModuleRef("a/b", ""); got != "a/b" {
		t.Fatalf("FormatModuleRef empty ver = %q", got)
	}
	if got := FormatModuleRef("a/b", "v1"); got != "a/b@v1" {
		t.Fatalf("FormatModuleRef with ver = %q", got)
	}
}

func TestModuleDefaultPrefix(t *testing.T) {
	if got := moduleDefaultPrefix("github.com/androticus/final/modules/csv"); got != "csv" {
		t.Fatalf("prefix = %q, want csv", got)
	}
	if got := moduleDefaultPrefix("csv"); got != "csv" {
		t.Fatalf("prefix = %q, want csv", got)
	}
}
