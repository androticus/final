package core

// ResolveBuiltInTypeRegistrations patches bootstrap-queued metadata type refs.
// Call once after all init()-time CreateRegister* functions, before InitGlobalExecutionEnvironment.
// Also instantiates the process Sys singleton (needs resolved StringList).
func ResolveBuiltInTypeRegistrations() {
	GetTypeRegistrar().ResolveRegistrations()
	initGlobalSys()
}
