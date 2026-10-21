package core

import "testing"

func TestInitGlobalExecutionEnvironmentRunsOnce(t *testing.T) {
	before := len(globalDictionaryContext.dict.byOrdinal)
	InitGlobalExecutionEnvironment()
	InitGlobalExecutionEnvironment()
	if len(globalDictionaryContext.dict.byOrdinal) != before {
		t.Fatalf("dict len = %d after second Init, want unchanged %d", len(globalDictionaryContext.dict.byOrdinal), before)
	}
}

func TestResetTestRuntimeReRegistersBootstrapWords(t *testing.T) {
	ResetTestRuntime()
	if globalDictionaryContext.dict.GetByName(bootWordColdBoot) == nil {
		t.Fatal("ColdBoot missing after ResetTestRuntime")
	}
	if globalDictionaryContext.dict.GetByName(":") == nil {
		t.Fatal(": missing after ResetTestRuntime")
	}
	if globalDictionaryContext.dict.GetByName("if") == nil {
		t.Fatal("if missing after ResetTestRuntime (core bootstrap not in snapshot)")
	}
	if globalDictionaryContext.dict.GetByName(">json") == nil {
		t.Fatal(">json missing after ResetTestRuntime (core2 bootstrap not in snapshot)")
	}
}
