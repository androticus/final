package core

import "testing"

func TestCoreModuleRegistered(t *testing.T) {
	InitTestRuntime()
	if !coreModuleRegistered {
		t.Fatal("core module not registered")
	}
	m := ModuleTable.moduleByName("core")
	if m == nil {
		t.Fatal("core module missing from table")
	}
	if m.moduleId != CoreModuleSlot {
		t.Fatalf("core moduleId = %d, want %d", m.moduleId, CoreModuleSlot)
	}
	if m.dc != globalDictionaryContext {
		t.Fatal("core dc != globalDictionaryContext")
	}
	if globalDictionaryContext.moduleId != CoreModuleSlot {
		t.Fatalf("globalDictionaryContext.moduleId = %d, want %d", globalDictionaryContext.moduleId, CoreModuleSlot)
	}
}

func TestNewTaskCompileTargetIsCoreModule(t *testing.T) {
	InitTestRuntime()
	task := NewTestTask()
	if task.compileModID != CoreModuleSlot {
		t.Fatalf("compileModID = %d, want %d", task.compileModID, CoreModuleSlot)
	}
	if task.dc != globalDictionaryContext {
		t.Fatal("task.dc != globalDictionaryContext")
	}
}

func TestRegisterCoreModuleIdempotent(t *testing.T) {
	InitTestRuntime()
	RegisterCoreModule(globalDictionaryContext)
	if ModuleTable.moduleByName("core") == nil {
		t.Fatal("core module missing after idempotent register")
	}
}
