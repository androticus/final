package core

import "fmt"

// GetByeReq reads the one-shot leave-quit flag from DT slot ReservedDTSlotByeReq.
func GetByeReq(dc *DictionaryContext) Bool {
	if dc == nil || dc.dt == nil {
		return false
	}
	if dc.dt.Len() <= ReservedDTSlotByeReq {
		return false
	}
	v := dc.dt.Read(ReservedDTSlotByeReq)
	b, ok := v.(Bool)
	if !ok {
		panic(fmt.Errorf("GetByeReq: DT[%d] is %T, want Bool", ReservedDTSlotByeReq, v))
	}
	return b
}

// SetByeReq writes the leave-quit flag to DT slot ReservedDTSlotByeReq.
func SetByeReq(dc *DictionaryContext, v Bool) {
	if dc == nil || dc.dt == nil {
		panic("SetByeReq: nil DictionaryContext or DT")
	}
	if dc.dt.Len() <= ReservedDTSlotByeReq {
		panic(fmt.Errorf("SetByeReq: DT has no reserved slot %d", ReservedDTSlotByeReq))
	}
	dc.dt.Write(ReservedDTSlotByeReq, v)
}

// CheckResetByeReq returns the previous byeReq value and clears the slot to false.
func CheckResetByeReq(dc *DictionaryContext) Bool {
	was := GetByeReq(dc)
	if was {
		SetByeReq(dc, false)
	}
	return was
}

// CheckResetCoreByeReq is like CheckResetByeReq on the sealed core module DT.
// Bootstrap byeReq (OpBootstrapDoVar) always lives there, even when the active
// task compiles into a local module.
func CheckResetCoreByeReq() Bool {
	if globalDictionaryContext == nil {
		return false
	}
	return CheckResetByeReq(globalDictionaryContext)
}

// ClearCoreByeReq clears the bootstrap leave-quit flag on the core module DT.
func ClearCoreByeReq() {
	if globalDictionaryContext != nil {
		SetByeReq(globalDictionaryContext, false)
	}
}
