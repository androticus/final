package core

import (
	"bufio"
	"strings"
	"testing"
)

func TestParenOpsCompiled(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(`
: plist ( 1 2 "three" ) ;
: psl ( "a" "b" s) ;
: pdyn { 1 "one" 2 "two" } ;
plist
psl
pdyn
`)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	if rp.ReturnStack.Len() != 0 {
		t.Fatalf("RS len = %d, want 0", rp.ReturnStack.Len())
	}
	dyn, ok := rp.DataStack.Pop().(*DynamicObject)
	if !ok {
		t.Fatalf("pdyn = %T, want *DynamicObject", rp.DataStack.Peek())
	}
	if dyn.Get("one") != Int(1) || dyn.Get("two") != Int(2) {
		t.Fatalf("pdyn fields = one:%v two:%v", dyn.Get("one"), dyn.Get("two"))
	}
	sl, ok := rp.DataStack.Pop().(*StringList)
	if !ok || sl.Len() != 2 {
		t.Fatalf("psl = %v, want StringList len 2", rp.DataStack.Peek())
	}
	lst, ok := rp.DataStack.Pop().(*List[any])
	if !ok || lst.Len() != 3 {
		t.Fatalf("plist = %v, want List len 3", rp.DataStack.Peek())
	}
}

func TestParenOpsInterpreted(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(`
( 10 20 30 )
{ 1 "x" 2 "y" }
`)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	if rp.ReturnStack.Len() != 0 {
		t.Fatalf("RS len = %d, want 0 after interpret floor cleanup", rp.ReturnStack.Len())
	}
	dyn, ok := rp.DataStack.Pop().(*DynamicObject)
	if !ok {
		t.Fatalf("TOS = %T, want *DynamicObject", rp.DataStack.Peek())
	}
	if dyn.Get("x") != Int(1) || dyn.Get("y") != Int(2) {
		t.Fatalf("dyn fields x:%v y:%v", dyn.Get("x"), dyn.Get("y"))
	}
	lst, ok := rp.DataStack.Pop().(*List[any])
	if !ok || lst.Len() != 3 {
		t.Fatalf("list = %v, want List len 3", rp.DataStack.Peek())
	}
}

func TestParenOpsContextMarkNotOnRS(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	ensureInnermostColonFrameMark(rp.ReturnStack)
	RunOps(task, rp, OpParen, OpStop)
	cm := rp.ReturnStack.Peek()
	if _, ok := cm.GetContextMark(); !ok {
		t.Fatal("expected ContextMark on ColonFrameMark slot after OpParen")
	}
}

func TestForeachWithParenCompiled(t *testing.T) {
	// nesting-snafu paren arm: j inside ( ) must see outer foreach (ContextMark not on RS)
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(`
: snafuParen
  0
  0 0 2 range foreach
    0 0 0 range foreach
      ( j ) drop
      j +
    /foreach
  /foreach
;
snafuParen
`)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	// 0 0 2 range → 0,1; inner once each adding j → 0+1 = 2 (proves j survives ( ))
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 2 {
		t.Fatalf("snafuParen = %v, want 2", got)
	}
}

func TestNestedParenOpsCompiled(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(`
: nested
  ( 1 ( 2 3 ) 4 )
;
nested
`)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	lst, ok := rp.DataStack.Pop().(*List[any])
	if !ok || lst.Len() != 3 {
		t.Fatalf("nested = %v, want List len 3", rp.DataStack.Peek())
	}
	if lst.At(0) != Int(1) || lst.At(2) != Int(4) {
		t.Fatalf("outer cells = %v %v, want 1 … 4", lst.At(0), lst.At(2))
	}
	inner, ok := lst.At(1).(*List[any])
	if !ok || inner.Len() != 2 || inner.At(0) != Int(2) || inner.At(1) != Int(3) {
		t.Fatalf("inner = %v, want List [2 3]", lst.At(1))
	}
}

func TestNestedBraceListCompiled(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(`
: tree
  { ( 10 20 ) "nums" 3 "n" }
;
tree
`)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	dyn, ok := rp.DataStack.Pop().(*DynamicObject)
	if !ok {
		t.Fatalf("tree = %T, want *DynamicObject", rp.DataStack.Peek())
	}
	if dyn.Get("n") != Int(3) {
		t.Fatalf("n = %v, want 3", dyn.Get("n"))
	}
	nums, ok := dyn.Get("nums").(*List[any])
	if !ok || nums.Len() != 2 || nums.At(0) != Int(10) || nums.At(1) != Int(20) {
		t.Fatalf("nums = %v, want List [10 20]", dyn.Get("nums"))
	}
}
