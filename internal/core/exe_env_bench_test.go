package core

import "testing"

// BenchmarkBootstrapReset measures restoring the post-init global environment snapshot.
// Package init() already ran once before benchmarks start.
func BenchmarkBootstrapReset(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ResetGlobalExecutionEnvironmentForTest()
	}
}

// BenchmarkBootstrapColdBuild measures installGlobalCoreVocabulary on a freshly allocated
// DictionaryContext. Restores the process-global context after the benchmark completes.
func BenchmarkBootstrapColdBuild(b *testing.B) {
	ResetGlobalExecutionEnvironmentForTest()
	saved := globalDictionaryContext
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dc := NewDictionaryContext(
			"global",
			globalDictMaxSize, globalDictCap,
			globalCTMaxSize, globalCTCap,
			globalDTMaxSize, globalDTCap,
			GetTypeRegistrar(),
		)
		globalDictionaryContext = dc
		installGlobalCoreVocabulary(dc)
	}
	b.StopTimer()
	globalDictionaryContext = saved
	ResetGlobalExecutionEnvironmentForTest()
}
