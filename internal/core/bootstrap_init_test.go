package core

import "testing"

func TestBootstrapGlobalColdEntryIsCTZero(t *testing.T) {
	ResetTestRuntime()
	v := BootstrapGlobalColdEntry()
	mod, off := DecodeVirt(v)
	if mod != ModuleIDThis || off != ReservedCTBootColdEntry {
		t.Fatalf("entry = mod %d off %d, want global CT[%d]", mod, off, ReservedCTBootColdEntry)
	}
}

func TestRunBootstrapInitScriptUsesColdBoot(t *testing.T) {
	ResetTestRuntime()
	tk, _ := NewTask(false)
	rp := NewTestRunProvider()
	runBootstrapInitScript(tk, rp, "1 2 +\n")
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 3 {
		t.Fatalf("got %v, want 3", got)
	}
}

func TestTranche3DisabledByDefault(t *testing.T) {
	if tranche3Enabled {
		t.Fatal("core tests must not enable fyne core3")
	}
}

func TestTranche4DisabledByDefault(t *testing.T) {
	if tranche4Enabled {
		t.Fatal("core tests must not enable pi core4")
	}
}
