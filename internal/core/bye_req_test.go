package core

import "testing"

func TestByeReqInDT(t *testing.T) {
	dc := globalDictionaryContext
	if GetByeReq(dc) {
		t.Fatal("fresh byeReq should be false")
	}
	SetByeReq(dc, true)
	if !GetByeReq(dc) {
		t.Fatal("byeReq should be true")
	}
	SetByeReq(dc, false)
	if GetByeReq(dc) {
		t.Fatal("byeReq should be false")
	}
}

func TestCheckResetByeReq(t *testing.T) {
	dc := globalDictionaryContext
	SetByeReq(dc, false)
	if CheckResetByeReq(dc) {
		t.Fatal("CheckResetByeReq on false should return false")
	}
	if GetByeReq(dc) {
		t.Fatal("slot should stay false")
	}

	SetByeReq(dc, true)
	if !CheckResetByeReq(dc) {
		t.Fatal("CheckResetByeReq on true should return true")
	}
	if GetByeReq(dc) {
		t.Fatal("slot should be cleared")
	}
	if CheckResetByeReq(dc) {
		t.Fatal("second CheckResetByeReq should return false")
	}
}

func TestResetTestRuntimeRestoresByeReqSlot(t *testing.T) {
	SetByeReq(globalDictionaryContext, true)
	ResetTestRuntime()
	if GetByeReq(globalDictionaryContext) {
		t.Fatal("ResetTestRuntime should restore byeReq slot to false")
	}
}
