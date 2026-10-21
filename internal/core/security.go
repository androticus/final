package core

import "fmt"

// SecurityRestrictions are themed capability bits stored on Word.srFlags (uint32).
// Public Final names live on the synthetic type SR — not on WordFlags.
// Spec: docs/003-design-notes/143-security-restrictions.md
type SecurityRestrictions = uint32

// SR theme bits occupy a dedicated uint32 field (remapped from the former high
// WordFlags tranche: newBit = oldBit - 32). Groups leave spare bits for later themes.
// IMPORTANT: numeric values changed once for 32-bit portability (final-32); use named
// SR_* / Final SR.* constants — do not hardcode old high-bit hex masks.
const (
	// Currently unused in the remapped layout:
	// bits 2, 4..11 (and other spares noted per group)

	// Module group (was bits 32..33)
	SR_ModuleInstall SecurityRestrictions = 0x1 // 1 << 0; was 1 << 32
	SR_ModuleImport  SecurityRestrictions = 0x2 // 1 << 1; was 1 << 33

	// System group (was bit 35)
	SR_System SecurityRestrictions = 0x8 // 1 << 3; was 1 << 35

	// File group — remapped bits 12..16 (was 44..48; 3 used, 2 spare)
	SR_FileRead       SecurityRestrictions = 0x1000 // 1 << 12; was 1 << 44
	SR_FileWrite      SecurityRestrictions = 0x2000 // 1 << 13; was 1 << 45
	SR_FileSystemRead SecurityRestrictions = 0x4000 // 1 << 14; was 1 << 46
	// bits 15..16 reserved (file family)

	// Http group — remapped bits 17..20 (was 49..52; 2 used, 2 spare)
	SR_HttpDial   SecurityRestrictions = 0x20000 // 1 << 17; was 1 << 49
	SR_HttpListen SecurityRestrictions = 0x40000 // 1 << 18; was 1 << 50
	// bits 19..20 reserved (http family)

	// Threads group — remapped bits 21..22 (was 53..54; 1 used, 1 spare)
	SR_Threads SecurityRestrictions = 0x200000 // 1 << 21; was 1 << 53
	// bit 22 reserved (concurrency family)

	// Db group — remapped bits 23..24 (was 55..56; 1 used, 1 spare)
	SR_Db SecurityRestrictions = 0x800000 // 1 << 23; was 1 << 55
	// bit 24 reserved (db family)

	// Crypto group — remapped bits 25..26 (was 57..58; 1 used, 1 spare)
	SR_Crypto SecurityRestrictions = 0x2000000 // 1 << 25; was 1 << 57
	// bit 26 reserved (crypto family)

	// bits 27..31 available for future groups

	// SR_Mask is the full uint32 SecurityRestrictions capacity.
	SR_Mask SecurityRestrictions = 0xFFFFFFFF
)

// SRFlags returns the SecurityRestrictions on w (Word.srFlags).
func (w *Word) SRFlags() SecurityRestrictions {
	if w == nil {
		return 0
	}
	return SecurityRestrictions(w.srFlags)
}

// accumulateSRFrom ORs from's SR bits into into (viral colon percolation).
func accumulateSRFrom(into, from *Word) {
	if into == nil || from == nil {
		return
	}
	into.srFlags |= from.srFlags
}

// AssertSR panics with a fixed message if target intersects the disabled-set mask.
// Intentionally opaque — no word names or bit details (security).
func AssertSR(target, restricts SecurityRestrictions) {
	if target&restricts != 0 {
		panic("operation not allowed")
	}
}

// assertSystemAccess denies WordFlagSystem unless t.systemMode (core compile).
// Fail-closed if the word is System and task is missing. Not viral; not an SR bit.
func assertSystemAccess(t *Task, w *Word) {
	if w == nil || !w.IsSystem() {
		return
	}
	if t == nil || !t.systemMode {
		panic("operation not allowed")
	}
}

// assertWordAllowed is Find/exec security: SR disabled-set, then System privilege.
func assertWordAllowed(t *Task, w *Word) {
	if w == nil {
		return
	}
	if t != nil {
		AssertSR(w.SRFlags(), t.restrict)
	}
	assertSystemAccess(t, w)
}

// assertMemberRestrict denies a nonzero member restrict against the task mask.
// Fail-closed if restrict is set but frame/task is missing.
func assertMemberRestrict(f *RunFrame, restrict SecurityRestrictions) {
	if restrict == 0 {
		return
	}
	if f == nil || f.Task == nil {
		panic("operation not allowed")
	}
	AssertSR(restrict, f.Task.restrict)
}

// CreateRegisterSRType registers SR as a synthetic sealed type whose class properties
// expose SecurityRestrictions constants (Final: SR.FileRead, …). System vocabulary.
func CreateRegisterSRType() {
	t := &Type{
		name: "SR",
		getZero: func() any {
			return SecurityRestrictions(0)
		},
		checkCoerceType: func(value any) any {
			u, ok := integralToUint32(value)
			if !ok {
				panic(fmt.Errorf("SR.checkCoerceType: value %T is not SecurityRestrictions", value))
			}
			return SecurityRestrictions(u)
		},
	}
	registerSRClassConstants(t)
	GetTypeRegistrar().RegisterCurrentType(t)
}

func registerSRClassConstants(t *Type) {
	for _, c := range []struct {
		name string
		val  SecurityRestrictions
	}{
		{"ModuleInstall", SR_ModuleInstall},
		{"ModuleImport", SR_ModuleImport},
		{"FileRead", SR_FileRead},
		{"FileWrite", SR_FileWrite},
		{"FileSystemRead", SR_FileSystemRead},
		{"HttpDial", SR_HttpDial},
		{"HttpListen", SR_HttpListen},
		{"Threads", SR_Threads},
		{"Db", SR_Db},
		{"Crypto", SR_Crypto},
		{"Mask", SR_Mask},
	} {
		v := Uint(c.val)
		name := c.name
		t.registerClassProperty(name, "Uint", func(_ Any) Any {
			return v
		}, nil)
	}
}

func init() {
	CreateRegisterSRType()
}
