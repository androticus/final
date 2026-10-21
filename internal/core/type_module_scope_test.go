package core

import (
	"fmt"
	"testing"
)

// makeLocalType mints a user type with the given name in task's module via CreateTypeFromDef.
func makeLocalType(task *Task, name string) *Type {
	def := NewTypeDef()
	def.Set("name", String(name))
	return CreateTypeFromDef(task.dc, def)
}

// TestLocalModuleTypeDoesNotLeakToGlobal verifies a type defined in a local module
// lands in that module's registrar/dict only, never the global core catalog or dict.
func TestLocalModuleTypeDoesNotLeakToGlobal(t *testing.T) {
	ResetTestRuntime()
	task := CreateLocalTask()
	name := "LocalScoped_" + t.Name()

	typ := makeLocalType(task, name)

	if got := task.dc.tr.Get(name); got != typ {
		t.Fatalf("local registrar Get(%q) = %v, want %p", name, got, typ)
	}
	if got := GetTypeRegistrar().Get(name); got != nil {
		t.Fatalf("global registrar Get(%q) = %v, want nil (leaked)", name, got)
	}
	if w := task.dc.dict.GetByName(name); w == nil || !w.IsTypeWord() {
		t.Fatalf("local dict word for %q = %v, want a type word", name, w)
	}
	if w := globalDictionaryContext.dict.GetByName(name); w != nil {
		t.Fatalf("global dict word for %q = %v, want nil (leaked)", name, w)
	}
}

// TestLocalModuleParentFallbackResolvesCoreType verifies the module registrar still
// resolves core field types (Any) via parent fallback during registration.
func TestLocalModuleParentFallbackResolvesCoreType(t *testing.T) {
	ResetTestRuntime()
	task := CreateLocalTask()

	anyCore := GetTypeRegistrar().Get("Any")
	if anyCore == nil {
		t.Fatal("missing core Any type")
	}
	if got := task.dc.tr.Get("Any"); got != anyCore {
		t.Fatalf("local registrar Get(%q) = %v, want core %p", "Any", got, anyCore)
	}
}

// TestCoreTypeLiteralResolvesFromLocalTask verifies a core type word (absolute
// CoreModuleSlot VirtTypeIdx) resolves to the core *Type when executed on a local task.
func TestCoreTypeLiteralResolvesFromLocalTask(t *testing.T) {
	ResetTestRuntime()
	intT := GetTypeRegistrar().Get("Int")
	if intT == nil {
		t.Fatal("missing core Int type")
	}
	coreWord := globalDictionaryContext.dict.GetByName("Int")
	if coreWord == nil {
		t.Fatal("missing core Int type word")
	}

	task := CreateLocalTask()
	rp := NewTestRunProvider()
	if err := runExecWord(task, rp, coreWord); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(*Type)
	if !ok || got != intT {
		t.Fatalf("core Int literal on local task = %v, want %p", got, intT)
	}
}

// TestLocalTypeLiteralResolvesToOwningModule verifies a local type word resolves to
// the local *Type via its module-scoped VirtTypeIdx.
func TestLocalTypeLiteralResolvesToOwningModule(t *testing.T) {
	ResetTestRuntime()
	task := CreateLocalTask()
	name := "LocalDispatch_" + t.Name()
	typ := makeLocalType(task, name)

	w := task.dc.dict.GetByName(name)
	if w == nil {
		t.Fatalf("missing local dict word for %q", name)
	}
	wantSub := EncodeVirt(task.dc.moduleId, typ.TypeIdx())
	if w.subCode != wantSub {
		t.Fatalf("word %q subCode = %d, want VirtTypeIdx %d", name, w.subCode, wantSub)
	}

	rp := NewTestRunProvider()
	if err := runExecWord(task, rp, w); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(*Type)
	if !ok || got != typ {
		t.Fatalf("local type literal = %v, want %p", got, typ)
	}
}

// TestTwoLocalModulesReuseTypeName verifies distinct local sessions may register the
// same type name without a duplicate-name panic (separate catalogs).
func TestTwoLocalModulesReuseTypeName(t *testing.T) {
	ResetTestRuntime()
	name := "SharedName_" + t.Name()

	taskA := CreateLocalTask()
	taskB := CreateLocalTask()

	typA := makeLocalType(taskA, name)
	typB := makeLocalType(taskB, name)

	if typA == typB {
		t.Fatal("expected distinct *Type per module")
	}
	if got := taskA.dc.tr.Get(name); got != typA {
		t.Fatalf("module A Get(%q) = %v, want %p", name, got, typA)
	}
	if got := taskB.dc.tr.Get(name); got != typB {
		t.Fatalf("module B Get(%q) = %v, want %p", name, got, typB)
	}
	if got := GetTypeRegistrar().Get(name); got != nil {
		t.Fatalf("global Get(%q) = %v, want nil", name, got)
	}
}

// sanity: the module registrar assigns a local (0-based-ish) typeIdx independent of the
// growing core catalog, so two modules can produce the same local typeIdx for their types.
func TestLocalTypeIdxIsModuleLocal(t *testing.T) {
	ResetTestRuntime()
	taskA := CreateLocalTask()
	taskB := CreateLocalTask()
	typA := makeLocalType(taskA, fmt.Sprintf("IdxA_%s", t.Name()))
	typB := makeLocalType(taskB, fmt.Sprintf("IdxB_%s", t.Name()))
	if typA.TypeIdx() != typB.TypeIdx() {
		t.Fatalf("local typeIdx A=%d B=%d, want equal module-local indices", typA.TypeIdx(), typB.TypeIdx())
	}
}
