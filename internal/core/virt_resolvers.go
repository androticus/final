package core

import "fmt"

// Virt*Resolver types implement Task-scoped virtual table access for VirtAddr values.
// EncodeVirt/DecodeVirt pack ModuleID in the high half — there is no separate ModuleBase type.
//
// Write/Append must respect dependency DAG: stored VirtAddrs only point upstream (see 07-doc).
//
// Current stub: ModuleIDThis / CoreModuleSlot reach globalDictionaryContext.
// Absolute module slots ≥ 2 are not implemented here yet (run-loop demux uses ModuleTable).
type VirtDictResolver struct {
	localDict *Dictionary // unused until a non-process module segment exists
}

type VirtCTResolver struct {
	localCT *CodeTable // unused until a non-process module segment exists
}

type VirtDTResolver struct {
	localDT *DataTable // unused until a non-process module segment exists
}

func unpackVirtAddr(v VirtualAddr) (moduleId ModuleID, offset int) {
	return DecodeVirt(v)
}

func (r *VirtDictResolver) dictForModID(modID ModuleID) *Dictionary {
	switch modID {
	case ModuleIDThis, CoreModuleSlot:
		if globalDictionaryContext == nil {
			panic(fmt.Errorf("VirtDictResolver: global execution environment not initialized"))
		}
		return globalDictionaryContext.dict
	default:
		panic(fmt.Errorf("VirtDictResolver: unsupported ModuleID %d", modID))
	}
}

func (r *VirtDictResolver) dictForVirt(virtAddr VirtualAddr) (*Dictionary, int) {
	moduleId, offset := unpackVirtAddr(virtAddr)
	switch moduleId {
	case ModuleIDThis:
		if globalDictionaryContext == nil {
			panic(fmt.Errorf("VirtDictResolver: global execution environment not initialized"))
		}
		return globalDictionaryContext.dict, offset
	default:
		panic(fmt.Errorf("VirtDictResolver: module ModuleID %d not implemented", moduleId))
	}
}

func (r *VirtDictResolver) GetByName(modID ModuleID, name string) *Word {
	return r.dictForModID(modID).GetByName(name)
}

func (r *VirtDictResolver) GetByOrdinal(virtAddr VirtualAddr) *Word {
	d, idx := r.dictForVirt(virtAddr)
	return d.GetByOrdinal(idx)
}

func (r *VirtDictResolver) Append(modID ModuleID, word *Word) VirtualAddr {
	d := r.dictForModID(modID)
	idx := len(d.byOrdinal)
	d.Add(word)
	return EncodeVirt(modID, idx)
}

func (r *VirtCTResolver) ctForVirt(virtAddr VirtualAddr) (*CodeTable, int) {
	moduleId, offset := unpackVirtAddr(virtAddr)
	switch moduleId {
	case ModuleIDThis:
		if globalDictionaryContext == nil {
			panic(fmt.Errorf("VirtCTResolver: global execution environment not initialized"))
		}
		return globalDictionaryContext.ct, offset
	default:
		panic(fmt.Errorf("VirtCTResolver: module ModuleID %d not implemented", moduleId))
	}
}

func (r *VirtCTResolver) ctForModID(modID ModuleID) *CodeTable {
	switch modID {
	case ModuleIDThis, CoreModuleSlot:
		if globalDictionaryContext == nil {
			panic(fmt.Errorf("VirtCTResolver: global execution environment not initialized"))
		}
		return globalDictionaryContext.ct
	default:
		panic(fmt.Errorf("VirtCTResolver: unsupported ModuleID %d", modID))
	}
}

func (r *VirtCTResolver) Read(virtAddr VirtualAddr) int {
	ct, idx := r.ctForVirt(virtAddr)
	return ct.Read(idx)
}

func (r *VirtCTResolver) Write(virtAddr VirtualAddr, val int) {
	ct, idx := r.ctForVirt(virtAddr)
	ct.Write(idx, val)
}

func (r *VirtCTResolver) Append(modID ModuleID, val int) VirtualAddr {
	idx := r.ctForModID(modID).Append(val)
	return EncodeVirt(modID, idx)
}

func (r *VirtDTResolver) dtForVirt(virtAddr VirtualAddr) (*DataTable, int) {
	moduleId, offset := unpackVirtAddr(virtAddr)
	switch moduleId {
	case ModuleIDThis:
		if globalDictionaryContext == nil {
			panic(fmt.Errorf("VirtDTResolver: global execution environment not initialized"))
		}
		return globalDictionaryContext.dt, offset
	default:
		panic(fmt.Errorf("VirtDTResolver: module ModuleID %d not implemented", moduleId))
	}
}

func (r *VirtDTResolver) dtForModID(modID ModuleID) *DataTable {
	switch modID {
	case ModuleIDThis, CoreModuleSlot:
		if globalDictionaryContext == nil {
			panic(fmt.Errorf("VirtDTResolver: global execution environment not initialized"))
		}
		return globalDictionaryContext.dt
	default:
		panic(fmt.Errorf("VirtDTResolver: unsupported ModuleID %d", modID))
	}
}

func (r *VirtDTResolver) Read(virtAddr VirtualAddr) any {
	dt, idx := r.dtForVirt(virtAddr)
	return dt.Read(idx)
}

func (r *VirtDTResolver) Write(virtAddr VirtualAddr, val any) {
	dt, idx := r.dtForVirt(virtAddr)
	dt.Write(idx, val)
}

func (r *VirtDTResolver) Append(modID ModuleID, val any) VirtualAddr {
	idx := r.dtForModID(modID).Append(val)
	return EncodeVirt(modID, idx)
}
