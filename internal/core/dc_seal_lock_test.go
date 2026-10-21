package core

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestSealedDCRejectsColon(t *testing.T) {
	ResetTestRuntime()
	dc := globalDictionaryContext
	dc.Seal()
	defer dc.Unseal()

	task := NewTestTask()
	rp := NewTestRunProvider()
	tr, err := NewTextReaderFromString(`: sealedFail 1 ;`)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	err = interpretSession(task, rp)
	if err == nil {
		t.Fatal("expected error compiling into sealed DC")
	}
	if !strings.Contains(fmt.Sprint(err), "sealed") {
		t.Fatalf("err = %v, want sealed", err)
	}
}

func TestUnsealAllowsColon(t *testing.T) {
	ResetTestRuntime()
	dc := globalDictionaryContext
	dc.Seal()
	dc.Unseal()

	task := NewTestTask()
	rp := NewTestRunProvider()
	tr, err := NewTextReaderFromString(`: unsealOk 42 ;`)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	if err := interpretSession(task, rp); err != nil {
		t.Fatalf("interpret after Unseal: %v", err)
	}
	if dc.dict.GetByName("unsealOk") == nil {
		t.Fatal("missing unsealOk after Unseal")
	}
}

func TestLockedDCRejectsColonWhilePipAlive(t *testing.T) {
	ResetTestRuntime()
	parent := CreateLocalTask()
	rp := NewTestRunProvider()
	sb := NewStringBuilder()
	rp.Stdout = sb
	tr, err := NewTextReaderFromString(`: hangPip drop 200 delayMs ;`)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	rp.WithTaskFindScope(parent)
	if err := interpretSession(parent, rp); err != nil {
		t.Fatalf("compile hangPip: %v", err)
	}
	w := parent.dc.dict.GetByName("hangPip")
	if w == nil {
		t.Fatal("missing hangPip")
	}

	mm := pipType.ClassMethodByName("newPipParams")
	f, ths := memberRunFrame(rp, pipType)
	mm.invokeCall(f, ths)
	pp := rp.DataStack.Pop().(*AtomicDynamicObject)

	pip := spawnPip(parent, w.XT(), pp)
	if !parent.hasPip(pip.ID()) {
		t.Fatal("pip should be registered")
	}
	if !parent.dc.IsLockedOrSealed() {
		t.Fatal("DC should be locked while pip registered")
	}

	tr2, err := NewTextReaderFromString(`: lockedFail 1 ;`)
	if err != nil {
		t.Fatal(err)
	}
	rp2 := NewTestRunProvider()
	rp2.SetParser(NewParserFromTextRead(tr2))
	rp2.WithTaskFindScope(parent)
	err = interpretSession(parent, rp2)
	if err == nil {
		t.Fatal("expected error compiling into locked DC")
	}
	if !strings.Contains(fmt.Sprint(err), "locked") {
		t.Fatalf("err = %v, want locked", err)
	}

	waitPipUnregistered(t, parent, pip.ID())
	deadline := time.Now().Add(2 * time.Second)
	for parent.dc.IsLockedOrSealed() {
		if time.Now().After(deadline) {
			t.Fatal("DC still locked after pip exit")
		}
		time.Sleep(time.Millisecond)
	}
}
