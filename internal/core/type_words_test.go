package core

import (
	"fmt"
	"testing"
)

func assertCoreTypeWordsInDictionary(t *testing.T) {
	t.Helper()
	verified := 0
	for _, w := range typeWords {
		got := globalDictionaryContext.dict.GetByName(w.name)
		if got == nil {
			continue // test-defined type on global DC; absent from post-reset snapshot
		}
		typ := GetTypeRegistrar().Get(w.name)
		if typ == nil {
			t.Fatalf("typeWords entry %q missing from registrar", w.name)
		}
		if got != w {
			t.Fatalf("dictionary word for type %q = %p, want %p", w.name, got, w)
		}
		if !w.IsTypeWord() {
			t.Fatalf("word %q is not a type word", w.name)
		}
		wantSub := EncodeVirt(CoreModuleSlot, typ.TypeIdx())
		if w.subCode != wantSub {
			t.Fatalf("word %q subCode = %d, want VirtTypeIdx %d", w.name, w.subCode, wantSub)
		}
		verified++
	}
	if verified == 0 {
		t.Fatal("no core type words verified in dictionary")
	}
	if globalDictionaryContext.dict.GetByName("Int") == nil {
		t.Fatal("Int type word missing after reset")
	}
}

func TestTypeWordsRegisteredAfterInit(t *testing.T) {
	ResetTestRuntime()
	intT := GetTypeRegistrar().Get("Int")
	if intT == nil {
		t.Fatal("missing Int type in registrar")
	}
	w := globalDictionaryContext.dict.GetByName("Int")
	if w == nil {
		t.Fatal("missing Int type word in dictionary")
	}
	if !w.IsTypeWord() || !w.IsAxiom() {
		t.Fatalf("Int word flags = %v, want WordFlagType|WordFlagAxiom", w.flags)
	}
	if w.opCode != OpPushType {
		t.Fatalf("Int opCode = %s, want PushType", OpCodeName(w.opCode))
	}
	wantSub := EncodeVirt(CoreModuleSlot, intT.TypeIdx())
	if w.subCode != wantSub {
		t.Fatalf("Int subCode = %d, want VirtTypeIdx %d", w.subCode, wantSub)
	}
}

func TestOpPushTypeViaRunSubCode(t *testing.T) {
	ResetTestRuntime()
	intT := GetTypeRegistrar().Get("Int")
	if intT == nil {
		t.Fatal("missing Int type")
	}
	task := NewTestTask()
	rp := NewTestRunProvider()
	RunOpsWithSubCode(task, rp, intT.TypeIdx(), OpPushType, OpStop)
	got, ok := rp.DataStack.Pop().(*Type)
	if !ok || got != intT {
		t.Fatalf("got %v, want %p", got, intT)
	}
}

func TestTypeWordInterpretExec(t *testing.T) {
	ResetTestRuntime()
	intT := GetTypeRegistrar().Get("Int")
	w := globalDictionaryContext.dict.GetByName("Int")
	if w == nil {
		t.Fatal("missing Int type word")
	}
	task := NewTestTask()
	rp := NewTestRunProvider()
	if err := runExecWord(task, rp, w); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(*Type)
	if !ok || got != intT {
		t.Fatalf("got %v, want %p", got, intT)
	}
}

func TestTypeWordCompileReference(t *testing.T) {
	ResetTestRuntime()
	intT := GetTypeRegistrar().Get("Int")
	w := globalDictionaryContext.dict.GetByName("Int")
	if w == nil {
		t.Fatal("missing Int type word")
	}
	task := NewTestTask()
	cs := task.compilerForTarget()
	bootstrapCreate(cs, "typeprobe")
	emitAt := cs.ct().Len()
	if err := bootstrapEmitWordReference(cs, w); err != nil {
		t.Fatal(err)
	}
	ct := globalDictionaryContext.ct
	if ct.Read(emitAt) != int(OpDoSubCode) {
		t.Fatalf("CT[%d] = %d, want OpDoSubCode", emitAt, ct.Read(emitAt))
	}
	wantSub := EncodeVirt(CoreModuleSlot, intT.TypeIdx())
	if ct.Read(emitAt+1) != wantSub {
		t.Fatalf("CT[%d] = %d, want VirtTypeIdx %d", emitAt+1, ct.Read(emitAt+1), wantSub)
	}
	if ct.Read(emitAt+2) != int(OpPushType) {
		t.Fatalf("CT[%d] = %d, want OpPushType", emitAt+2, ct.Read(emitAt+2))
	}
}

func TestTypeWordInterpretViaSubCodeFromDS(t *testing.T) {
	ResetTestRuntime()
	intT := GetTypeRegistrar().Get("Int")
	task := NewTestTask()
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(intT.TypeIdx()))
	RunOps(task, rp, OpSubCodeFromDS, OpPushType, OpStop)
	got, ok := rp.DataStack.Pop().(*Type)
	if !ok || got != intT {
		t.Fatalf("got %v, want %p", got, intT)
	}
}

func TestEveryRegisteredTypeHasDictionaryWord(t *testing.T) {
	ResetTestRuntime()
	assertCoreTypeWordsInDictionary(t)
}

func TestCreateTypeFromDefRegistersTypeWord(t *testing.T) {
	ResetTestRuntime()
	cases := []struct {
		name string
		task *Task
	}{
		{"core", NewTestTask()},
		{"local", CreateLocalTask()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			typeName := fmt.Sprintf("DynType_%s_%s", tc.name, t.Name())
			def := NewTypeDef()
			def.Set("name", String(typeName))
			typ := CreateTypeFromDef(tc.task.dc, def)
			w := tc.task.dc.dict.GetByName(typeName)
			if w == nil {
				t.Fatalf("missing dictionary word for type %q", typeName)
			}
			if !w.IsTypeWord() {
				t.Fatalf("word %q is not a type word", typeName)
			}
			wantSub := EncodeVirt(tc.task.dc.moduleId, typ.TypeIdx())
			if w.subCode != wantSub {
				t.Fatalf("word %q subCode = %d, want VirtTypeIdx %d", typeName, w.subCode, wantSub)
			}
			rp := NewTestRunProvider()
			if err := runExecWord(tc.task, rp, w); err != nil {
				t.Fatal(err)
			}
			got, ok := rp.DataStack.Pop().(*Type)
			if !ok || got != typ {
				t.Fatalf("runExecWord got %v, want %p", got, typ)
			}
		})
	}
}
