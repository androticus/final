package finalgo

import "github.com/androticus/final/internal/core"

func init() {
	// Do not InitGlobalExecutionEnvironment here: optional core pins (fyneui tranche3,
	// finalpi tranche4) register in their package init() and must run before Seal.
	// NewSession / NewPersistentSession boots core after all package inits.
	core.ResolveBuiltInTypeRegistrations()
}

// Version returns the library version string.
func Version() string {
	return core.Version()
}

// Copyright returns the product copyright line (About / banners).
func Copyright() string {
	return core.Copyright
}
