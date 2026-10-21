package core

import (
	"fmt"
	"strings"
)

// WordFlags are dictionary kind/behavior bits on Word.flags (uint32; separate from SR).
type WordFlags = uint32

const (
	// Kind bits — axiom vs colon; Defined accompanies Colon on var:/const:/defer: instances.
	// NOTE: the flags MUST remain idempotent, since they are used in bootstrap code.
	WordFlagAxiom   WordFlags = 0x0001 // axiom or worker word (OpCode dispatch)
	WordFlagColon   WordFlags = 0x0002 // colon `: … ;` word, or defined instance (see WordFlagDefined)
	WordFlagDefined WordFlags = 0x0008 // instance from var:/const:/defer: mint (also WordFlagColon)

	WordFlagImmediate            WordFlags = 0x0010 // executes immediately during compilation
	WordFlagCompileOnly          WordFlags = 0x0020 // may only be compiled, not executed interactively or via exec()
	WordFlagCompilerOnly         WordFlags = 0x0040 // may only be compiled by the compiler, eg Lit; not used in source code
	WordFlagHidden               WordFlags = 0x0080 // hides words not normally needed by the user
	WordFlagRequiresStackContext WordFlags = 0x0100 // requires a context stack selector prefix
	WordFlagType                 WordFlags = 0x0200 // is a type
	// 0x0400 reserved (was WordFlagDefiner / does>)
	WordFlagSystem             WordFlags = 0x0800 // never-user: Find/exec only when Task.systemMode (core compile)
	WordFlagExecutesInEmitting WordFlags = 0x1000 // executes during [[ … ]] emission (eg ]]) instead of OpDoEmit
	WordFlagDefer              WordFlags = 0x2000 // is a defer word, used to verify use of 'is'
	// WordFlagInterpretOnly: may be executed interactively / via exec, but not compiled into other words.
	// Used by forget: mark words (142-forget.md).
	WordFlagInterpretOnly     WordFlags = 0x4000
	WordFlagOSContextRestrict WordFlags = 0x8000 // forbidden with o: prefix (mutates OStack layout, not inst field access)
	// WordFlagFinalMethod marks an ephemeral building-stack scaffold for `: Type#…` / `##` tear-off.
	// Never committed to the dictionary — compile-time only.
	WordFlagFinalMethod WordFlags = 0x0004
	// Instances created by var: / const: / defer: mint.
	WordFlagDefinedWord WordFlags = WordFlagColon | WordFlagDefined
)

type Word struct {
	name     string   // the name of the word, as used in source test
	moduleId ModuleID // the module ID of the word (0=global, 0x7FFFFFFF=local, 1..0x7FFFFFFE=loaded module)
	ordIdx   int      // the ordinal index of the word in the dictionary
	flags    WordFlags
	srFlags  uint32 // SecurityRestrictions; no Final write path (Go + viral only)

	opCode  OpCode // >=0: Run dispatches this opcode; -1 when not opcode-driven
	dtIdx   int    // >=0: base index in owning DC.dt; -1 when no DT slot
	subCode int    // >=0: dispatch code (eg table of methods); -1 when none

	colonCode CTP // colon `: … ;` body CT entry; -1 if unused

	doc string // prelude doc comment captured at definition time

	// Compile-time only (187 subwords); cleared when `;` finalizes.
	subwords     []Subword // definition order; entry -1 until ;s
	subSkipMark  int       // CT index of skip-branch offset; -1 = none or resolved
	openSubStart int       // CT body start of open :s; -1 if none
}

// Subword is a named subroutine within a colon word (compile-time name → body CT index).
type Subword struct {
	name  string
	entry IP // CT index of the body; -1 from :s until ;s
}

func validateDefiningWordName(name string) error {
	if strings.ContainsRune(name, '\\') {
		return fmt.Errorf("word name %q may not contain '\\'", name)
	}
	return nil
}

func NewWord(name string, flags WordFlags) *Word {
	if err := validateDefiningWordName(name); err != nil {
		panic(err)
	}
	return &Word{
		name:         name,
		ordIdx:       -1,
		flags:        flags,
		opCode:       -1,
		dtIdx:        -1,
		subCode:      -1,
		colonCode:    -1,
		subSkipMark:  -1,
		openSubStart: -1,
	}
}

func NewAxiomWord(name string, opCode OpCode, flags WordFlags) *Word {
	return &Word{
		name:      name,
		flags:     WordFlagAxiom | flags,
		opCode:    opCode,
		dtIdx:     -1,
		subCode:   -1,
		colonCode: -1,
	}
}

func NewSubcodeAxiomWord(name string, opCode OpCode, subCode int, flags WordFlags) *Word {
	return &Word{
		name:      name,
		flags:     WordFlagAxiom | flags,
		opCode:    opCode,
		dtIdx:     -1,
		subCode:   subCode,
		colonCode: -1,
	}
}

func NewReservedVarWord(name string, dtIdx int) *Word {
	return &Word{
		name:      name,
		flags:     WordFlagAxiom | WordFlagRequiresStackContext,
		opCode:    OpBootstrapDoVar,
		dtIdx:     dtIdx,
		subCode:   -1,
		colonCode: -1,
	}
}

func (w *Word) IsAxiom() bool {
	return w != nil && (w.flags&WordFlagAxiom) != 0
}

func (w *Word) IsColon() bool {
	return w != nil && (w.flags&WordFlagColon) != 0
}

func (w *Word) IsDefined() bool {
	return w != nil && (w.flags&WordFlagDefined) != 0
}

func (w *Word) IsImmediate() bool {
	return w != nil && (w.flags&WordFlagImmediate) != 0
}

func (w *Word) ExecutesInEmitting() bool {
	return w != nil && (w.flags&WordFlagExecutesInEmitting) != 0
}

func (w *Word) IsTypeWord() bool {
	return w != nil && (w.flags&WordFlagType) != 0
}

func (w *Word) IsSystem() bool {
	return w != nil && (w.flags&WordFlagSystem) != 0
}

// IsUserVisible is false for Hidden and System words, and for words whose SR flags
// intersect the Task's current disabled-set. Listing (not Find/exec) never shows System.
func (w *Word) IsUserVisible(t *Task) bool {
	var restrict SecurityRestrictions
	if t != nil {
		restrict = t.restrict
	}
	return w.isUserVisible(restrict)
}

func (w *Word) isUserVisible(restrict SecurityRestrictions) bool {
	if w == nil {
		return false
	}
	if w.flags&(WordFlagHidden|WordFlagSystem) != 0 {
		return false
	}
	return w.SRFlags()&restrict == 0
}

func (w *Word) RequiresStackContext() bool {
	return w != nil && w.flags&WordFlagRequiresStackContext != 0
}

func (w *Word) OSContextRestrict() bool {
	return w != nil && w.flags&WordFlagOSContextRestrict != 0
}

// quit loop lifecycle handlers — single source of truth for run vs emit.

// CompileWord emits a reference to w into the open colon body.
func (w *Word) CompileWord(rf *RunFrame) error {
	if w == nil || rf == nil || rf.Task == nil {
		return fmt.Errorf("CompileWord: nil argument")
	}
	cs := rf.Task.compilerForTarget()
	if cs == nil || !GetCompileState(cs) {
		return fmt.Errorf("CompileWord: not in compile state")
	}
	return bootstrapEmitWordReference(cs, w)
}

// NewTypeWord registers a dictionary entry for sealed type typ (WordFlagType axiom).
// subCode is the VirtTypeIdx: EncodeVirt(owning module base, local typeIdx). OpPushType
// demuxes it (like a word XT) to the owning module's DC + local catalog index, so a type
// literal resolves to its owning module regardless of the executing DC.
func NewTypeWord(typ *Type, moduleId ModuleID) *Word {
	if typ == nil {
		panic("NewTypeWord: nil type")
	}
	if typ.Name() == "" {
		panic("NewTypeWord: empty type name")
	}
	return &Word{
		name:      typ.Name(),
		flags:     WordFlagType | WordFlagAxiom,
		srFlags:   uint32(typ.restrict),
		opCode:    OpPushType,
		subCode:   EncodeVirt(moduleId, typ.TypeIdx()),
		ordIdx:    -1,
		dtIdx:     -1,
		colonCode: -1,
	}
}

func (w *Word) NaturalString() string {
	if w == nil {
		return "Word(nil)"
	}
	return "Word: " + w.name
}

// XT returns the execution token (dictionary ordinal) for w in this dictionary.
// VirtAddr high half is the owning module table slot (w.moduleId); resolve via ModuleTable at run.
func (w *Word) XT() XT {
	if w == nil || w.ordIdx < 0 {
		return 0
	}
	return XT(EncodeVirt(w.moduleId, w.ordIdx))
}

func (w *Word) FinalType() *Type { return GetTypeRegistrar().Get("Word") }

// CreateRegisterWordFlagsType registers WordFlags as a synthetic sealed type whose class
// properties expose dictionary flag constants (name → Uint value).
func CreateRegisterWordFlagsType() {
	t := &Type{
		name: "WordFlags",
		getZero: func() any {
			return WordFlags(0)
		},
		checkCoerceType: func(value any) any {
			u, ok := integralToUint32(value)
			if !ok {
				panic(fmt.Errorf("WordFlags.checkCoerceType: value %T is not WordFlags", value))
			}
			return WordFlags(u)
		},
	}
	registerWordFlagsClassConstants(t)
	GetTypeRegistrar().RegisterCurrentType(t)
}

func registerWordFlagsClassConstants(t *Type) {
	for _, c := range []struct {
		name string
		val  WordFlags
	}{
		{"Axiom", WordFlagAxiom},
		{"Colon", WordFlagColon},
		{"FinalMethod", WordFlagFinalMethod},
		{"Defined", WordFlagDefined},
		{"Immediate", WordFlagImmediate},
		{"CompileOnly", WordFlagCompileOnly},
		{"CompilerOnly", WordFlagCompilerOnly},
		{"Hidden", WordFlagHidden},
		{"RequiresStackContext", WordFlagRequiresStackContext},
		{"OSContextRestrict", WordFlagOSContextRestrict},
		{"Type", WordFlagType},
		{"System", WordFlagSystem},
		{"ExecutesInEmitting", WordFlagExecutesInEmitting},
		{"DefinedWord", WordFlagDefinedWord},
		{"Defer", WordFlagDefer},
		{"InterpretOnly", WordFlagInterpretOnly},
	} {
		v := Uint(c.val)
		name := c.name
		t.registerClassProperty(name, "Uint", func(_ Any) Any {
			return v
		}, nil)
	}
}

func CreateRegisterWordType() {
	t := &Type{
		name: "Word",
		getZero: func() any {
			return nil
		},
		checkCoerceType: func(value any) any {
			w, ok := value.(*Word)
			if !ok {
				panic(fmt.Errorf("Word.checkCoerceType: value %T is not *Word", value))
			}
			return w
		},
	}

	// Stored properties
	t.registerProperty("name", "String",
		func(inst Any) Any { return inst.(*Word).name },
		func(inst Any, value Any) { inst.(*Word).name = value.(String) },
	)
	t.registerProperty("ordIdx", "Int",
		func(inst Any) Any { return inst.(*Word).ordIdx },
		func(inst Any, value Any) { inst.(*Word).ordIdx = int(value.(Int)) },
	)
	t.registerProperty("flags", "Uint",
		func(inst Any) Any { return Uint(inst.(*Word).flags) },
		func(inst Any, value Any) {
			u, ok := integralToUint32(value)
			if !ok {
				panic(fmt.Errorf("Word.flags!: expected Int or Uint, got %T", value))
			}
			inst.(*Word).flags = WordFlags(u)
		},
	)
	t.registerProperty("opCode", "Int",
		func(inst Any) Any { return inst.(*Word).opCode },
		func(inst Any, value Any) {
			inst.(*Word).opCode = OpCode(value.(Int))
		},
	)
	t.registerProperty("dtIdx", "Int",
		func(inst Any) Any { return inst.(*Word).dtIdx },
		func(inst Any, value Any) {
			inst.(*Word).dtIdx = int(value.(Int))
		},
	)
	t.registerProperty("colonCode", "Int",
		func(inst Any) Any { return inst.(*Word).colonCode },
		func(inst Any, value Any) {
			inst.(*Word).colonCode = CTP(value.(Int))
		},
	)
	t.registerProperty("doc", "String",
		func(inst Any) Any { return String(inst.(*Word).doc) },
		nil,
	)

	// Computed properties — always after stored fields.
	t.registerProperty("isAxiom", "Bool",
		func(inst Any) Any { return Bool(inst.(*Word).IsAxiom()) },
		nil,
	)
	t.registerProperty("isColon", "Bool",
		func(inst Any) Any { return Bool(inst.(*Word).IsColon()) },
		nil,
	)
	t.registerProperty("isDefined", "Bool",
		func(inst Any) Any { return Bool(inst.(*Word).IsDefined()) },
		nil,
	)
	t.registerProperty("isType", "Bool",
		func(inst Any) Any { return Bool(inst.(*Word).IsTypeWord()) },
		nil,
	)
	t.registerProperty("xt", "Int",
		func(inst Any) Any { return Int(inst.(*Word).XT()) },
		nil,
	)
	t.registerNativeMethod("compileWord", func(f *RunFrame, ths *TypeObjPair) {
		if err := ths.Obj.(*Word).CompileWord(f); err != nil {
			panic(err)
		}
	})
	// ( -- … ) veneer over exec / OpExec — see 130-worker-word-pattern.md
	t.registerNativeMethod("exec", func(f *RunFrame, ths *TypeObjPair) {
		f.SetExec(ths.Obj.(*Word).XT())
	})
	t.registerNativeMethod("setDocFromLastComment",
		func(f *RunFrame, ths *TypeObjPair) {
			w := ths.Obj.(*Word)
			w.doc = f.RP.Parser.GetClearLastComment()
		},
	)
	// ( flags:iu name:String -- w:*Word ) — NewWord shell; not dict.Add (WordFlags only; no SR)
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		name := string(f.RP.DataStack.Pop().(String))
		u, ok := integralToUint32(f.RP.DataStack.Pop())
		if !ok {
			panic(fmt.Errorf("Word.new: flags expected Int or Uint"))
		}
		f.RP.DataStack.Push(NewWord(name, WordFlags(u)))
	})
	GetTypeRegistrar().RegisterCurrentType(t)
}
