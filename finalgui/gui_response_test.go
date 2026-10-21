package finalgui

import (
	"fmt"
	"strings"
	"testing"
)

func TestAppendResponseCapsAtMaxLines(t *testing.T) {
	g := &FinalGui{}
	for i := 0; i < responseMaxLines+5; i++ {
		g.appendResponse(fmt.Sprintf("L%d", i), nil)
	}
	got := g.responseLog.ToString()
	if strings.Count(got, "\n") != responseMaxLines {
		t.Fatalf("newlines = %d, want %d\n%s", strings.Count(got, "\n"), responseMaxLines, got)
	}
	if !strings.Contains(got, fmt.Sprintf("L%d\n", responseMaxLines+4)) {
		t.Fatalf("missing newest line in %q", got)
	}
	if strings.Contains(got, "L0\n") {
		t.Fatalf("oldest line still present in %q", got)
	}
}

func TestAppendHistoryCapsAtMaxLines(t *testing.T) {
	g := &FinalGui{}
	for i := 0; i < historyMaxLines+5; i++ {
		g.appendHistory(fmt.Sprintf("H%d", i))
	}
	got := g.historyLog.ToString()
	if strings.Count(got, "\n") != historyMaxLines {
		t.Fatalf("newlines = %d, want %d\n%s", strings.Count(got, "\n"), historyMaxLines, got)
	}
	if !strings.Contains(got, fmt.Sprintf("H%d\n", historyMaxLines+4)) {
		t.Fatalf("missing newest line in %q", got)
	}
	if strings.Contains(got, "H0\n") {
		t.Fatalf("oldest line still present in %q", got)
	}
}
