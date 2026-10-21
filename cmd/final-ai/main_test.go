package main

import "testing"

func TestResolveAPIKey_flagWins(t *testing.T) {
	t.Setenv("FINAL_AI_API_KEY", "from-env")
	if got := resolveAPIKey("  from-flag  "); got != "from-flag" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveAPIKey_envFallback(t *testing.T) {
	t.Setenv("FINAL_AI_API_KEY", "  from-env  ")
	if got := resolveAPIKey(""); got != "from-env" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveAPIKey_empty(t *testing.T) {
	t.Setenv("FINAL_AI_API_KEY", "")
	if got := resolveAPIKey(""); got != "" {
		t.Fatalf("got %q", got)
	}
}
