package core

import (
	"testing"
)

func TestStringLineBufferNoNewlineDoesNotCount(t *testing.T) {
	b := NewStringLineBuffer(3)
	b.Append("ab")
	b.Append("c")
	if got := b.ToString(); got != "abc" {
		t.Fatalf("ToString() = %q, want %q", got, "abc")
	}
	if b.lines != 0 {
		t.Fatalf("lines = %d, want 0", b.lines)
	}
}

func TestStringLineBufferGlueThenNewline(t *testing.T) {
	b := NewStringLineBuffer(2)
	b.Append("hel")
	b.Append("lo\n")
	if got := b.ToString(); got != "hello\n" {
		t.Fatalf("ToString() = %q, want %q", got, "hello\n")
	}
	if b.lines != 1 {
		t.Fatalf("lines = %d, want 1", b.lines)
	}
}

func TestStringLineBufferDropsOldestCompleteLines(t *testing.T) {
	b := NewStringLineBuffer(2)
	b.Append("a\nb\n")
	b.Append("c\n")
	if got := b.ToString(); got != "b\nc\n" {
		t.Fatalf("ToString() = %q, want %q", got, "b\nc\n")
	}
	if b.lines != 2 {
		t.Fatalf("lines = %d, want 2", b.lines)
	}
}

func TestStringLineBufferInputTallerThanNReplaces(t *testing.T) {
	b := NewStringLineBuffer(2)
	b.Append("keep\n")
	b.Append("a\nb\nc\n")
	if got := b.ToString(); got != "b\nc\n" {
		t.Fatalf("ToString() = %q, want %q", got, "b\nc\n")
	}
	if b.lines != 2 {
		t.Fatalf("lines = %d, want 2", b.lines)
	}
}

func TestStringLineBufferMaxOne(t *testing.T) {
	b := NewStringLineBuffer(1)
	b.Append("a\nb\n")
	if got := b.ToString(); got != "b\n" {
		t.Fatalf("ToString() = %q, want %q", got, "b\n")
	}
	if b.lines != 1 {
		t.Fatalf("lines = %d, want 1", b.lines)
	}
}

func TestStringLineBufferKeepsIncompleteSuffixWhenDropping(t *testing.T) {
	b := NewStringLineBuffer(1)
	b.Append("a\npartial")
	b.Append("x\n")
	if got := b.ToString(); got != "partialx\n" {
		t.Fatalf("ToString() = %q, want %q", got, "partialx\n")
	}
	if b.lines != 1 {
		t.Fatalf("lines = %d, want 1", b.lines)
	}
}

func TestStringLineBufferEmptyAppend(t *testing.T) {
	b := NewStringLineBuffer(1)
	b.Append("a\n")
	b.Append("")
	if got := b.ToString(); got != "a\n" {
		t.Fatalf("ToString() = %q, want %q", got, "a\n")
	}
}

func TestStringLineBufferGetTick(t *testing.T) {
	b := NewStringLineBuffer(4)
	if b.GetTick() != 0 {
		t.Fatalf("tick = %d, want 0", b.GetTick())
	}
	b.Append("")
	if b.GetTick() != 0 {
		t.Fatalf("empty append tick = %d, want 0", b.GetTick())
	}
	b.Append("ab")
	if b.GetTick() != 1 {
		t.Fatalf("tick = %d, want 1", b.GetTick())
	}
	b.Append("c\n")
	if b.GetTick() != 2 {
		t.Fatalf("tick = %d, want 2", b.GetTick())
	}
}
