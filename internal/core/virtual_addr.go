package core

// Virtual addressing (UVAS): 64-bit packing for CT addresses (VirtAddr / IP) and
// dictionary XTs (VirtXT). Same EncodeVirt/DecodeVirt bit layout; namespace is which
// API consumes the value. DT slots are never VirtAddr — plain indices into the
// companion CT's DictionaryContext.dt only (see docs/001-high-level/07-…md).
//
// EncodeVirt and DecodeVirt are pure — no Task required.
// Demux of VirtAddr/VirtXT needs Task / execDC (and ModuleTable for slots ≥ 1).
//
// Stored VirtAddr/VirtXT values obey the dependency DAG: more-dependent contexts
// may reference upstream only. Canonical spec:
// docs/001-high-level/07-dictionary-contexts-and-virtual-addressing.md.

// ModuleID identifies a dictionary context for module-manager, compile-target, and
// VirtAddr/VirtXT high-half packing. Absolute table slots are ≥ 1 (core at 1).
type ModuleID int

const (
	// ModuleIDThis is the compile/run "this" sentinel (not a module-table row).
	// Packed as VirtAddr/VirtXT high half 0 for same-module refs.
	ModuleIDThis ModuleID = 0
)

// VirtualAddr packs moduleId (b63–b32) and offset (b31–b0) into one int.
// Go type is int (same width as CT cells); bit layout is still 64-bit logical.
type VirtualAddr = int

// EncodeVirt packs moduleId and physical idx into a VirtualAddr.
// The stored low half is int(moduleId)+idx (moduleId alone when idx is 0).
func EncodeVirt(moduleId ModuleID, idx int) VirtualAddr {
	b := int(moduleId)
	return b<<32 | (b + idx)
}

// EncodeExecIP packs a demuxed run-loop IP with execDC's module slot for Task.ip.
// Nil dc uses ModuleIDThis (same as a session entry that has not demuxed yet).
func EncodeExecIP(dc *DictionaryContext, ip IP) VirtualAddr {
	if dc == nil {
		return EncodeVirt(ModuleIDThis, int(ip))
	}
	return EncodeVirt(dc.moduleId, int(ip))
}

// DecodeVirt unpacks a VirtualAddr into moduleId and the packed low half
// (callers subtract int(moduleId) to recover the physical index when moduleId ≠ 0).
func DecodeVirt(v VirtualAddr) (moduleId ModuleID, offset int) {
	return ModuleID(uint32(v >> 32)), int(uint32(v))
}
