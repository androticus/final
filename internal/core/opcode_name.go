package core

import "fmt"

// opCodeNames is indexed by OpCode ordinal; populated by registerOpCode during registerAxioms.
// Used for fault reports, tests, and switch-order validation — not for VM dispatch (Run uses int op values).
var opCodeNames []string

var opCodeByName map[string]OpCode

func resetOpCodeNames() {
	opCodeNames = make([]string, opAxiomLAST+1)
	opCodeByName = make(map[string]OpCode, opAxiomLAST+1)
}

// registerOpCode records the debug/lookup name for one axiom opcode (Go const name sans "Op" prefix).
// Called from regAxiom during registerAxioms only — one name per OpCode ordinal through opAxiomLAST.
func registerOpCode(op OpCode, goName string) {
	if op < 0 || op > opAxiomLAST {
		panic(fmt.Errorf("registerOpCode: opcode %d out of range 0..%d", op, opAxiomLAST))
	}
	if opCodeNames[op] != "" {
		panic(fmt.Errorf("registerOpCode: duplicate registration for opcode %d (%q, already %q)", op, goName, opCodeNames[op]))
	}
	opCodeNames[op] = goName
	opCodeByName[goName] = op
}

// OpCodeName returns the registered debug name for op, or a numeric fallback.
func OpCodeName(op OpCode) string {
	if op >= 0 && op < len(opCodeNames) && opCodeNames[op] != "" {
		return opCodeNames[op]
	}
	return fmt.Sprintf("OpCode(%d)", op)
}

// OpCodeByName resolves a registered opcode debug name (e.g. "AddI", "Stop") to its value.
func OpCodeByName(name string) (OpCode, bool) {
	op, ok := opCodeByName[name]
	return op, ok
}

// regAxiom registers one axiom dictionary entry and its OpCode debug name (Go const sans "Op").
func regAxiom(d *[]*Word, goName, finalName string, op OpCode, flags WordFlags) {
	registerOpCode(op, goName)
	*d = append(*d, NewAxiomWord(finalName, op, flags))
}

// regAxiomSec is regAxiom plus SecurityRestrictions on Word.srFlags (not WordFlags).
func regAxiomSec(d *[]*Word, goName, finalName string, op OpCode, flags WordFlags, sr SecurityRestrictions) {
	registerOpCode(op, goName)
	w := NewAxiomWord(finalName, op, flags)
	w.srFlags = uint32(sr)
	*d = append(*d, w)
}

func regSubcodeAxiom(d *[]*Word, finalName string, op OpCode, subCode int, flags WordFlags) {
	*d = append(*d, NewSubcodeAxiomWord(finalName, op, subCode, flags))
}
