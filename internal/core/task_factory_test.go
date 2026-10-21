package core

import (
	"strings"
	"testing"
)

func TestCreateCoreTaskUsesModuleTableCore(t *testing.T) {
	InitTestRuntime()
	coreMod := ModuleTable.moduleByName("core")
	if coreMod == nil {
		t.Fatal("core module missing")
	}

	task := CreateCoreTask()
	if task.compileModID != CoreModuleSlot {
		t.Fatalf("compileModID = %d, want %d", task.compileModID, CoreModuleSlot)
	}
	if task.dc != coreMod.dc {
		t.Fatal("task.dc != ModuleTable core dc")
	}
	if task.dc != globalDictionaryContext {
		t.Fatal("core task.dc != globalDictionaryContext")
	}
	if !task.systemMode {
		t.Fatal("CreateCoreTask must set systemMode")
	}
}

func TestCreateLocalTaskRegistersModule(t *testing.T) {
	InitTestRuntime()
	before := ModuleTable.modules.Len()

	task := CreateLocalTask()
	if ModuleTable.modules.Len() != before+1 {
		t.Fatalf("module table len = %d, want %d", ModuleTable.modules.Len(), before+1)
	}
	if !strings.HasPrefix(task.name, "local:") {
		t.Fatalf("task.name = %q, want local:<int64> prefix", task.name)
	}
	mod := ModuleTable.moduleByName(task.name)
	if mod == nil {
		t.Fatalf("module %q not in table", task.name)
	}
	if mod.moduleId != task.compileModID {
		t.Fatalf("mod.moduleId = %d, task.compileModID = %d", mod.moduleId, task.compileModID)
	}
	if task.dc != mod.dc {
		t.Fatal("task.dc != registered module dc")
	}
	if task.dc == globalDictionaryContext {
		t.Fatal("local task must not use core dc")
	}
	if task.dc.dt.Len() != 0 {
		t.Fatalf("local DC dt len = %d, want 0 (no bootstrap reserved slots)", task.dc.dt.Len())
	}
	if task.dc.moduleId != mod.moduleId {
		t.Fatalf("dc.moduleId = %d, want %d", task.dc.moduleId, mod.moduleId)
	}
	if task.systemMode {
		t.Fatal("CreateLocalTask must not set systemMode")
	}
}

func TestCreateLocalTaskConcurrentDistinctModules(t *testing.T) {
	InitTestRuntime()
	const n = 8
	ids := make(chan ModuleID, n)
	for range n {
		go func() {
			task := CreateLocalTask()
			ids <- task.compileModID
		}()
	}
	seen := make(map[ModuleID]struct{}, n)
	for range n {
		id := <-ids
		if id <= CoreModuleSlot {
			t.Fatalf("local moduleId = %d, want > %d", id, CoreModuleSlot)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate local moduleId %d", id)
		}
		seen[id] = struct{}{}
	}
}
