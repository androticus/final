package core

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	chdirRepoRootForTests()
	ResolveBuiltInTypeRegistrations()
	InitGlobalExecutionEnvironment()
	InitTestRuntime()
	os.Exit(m.Run())
}

func TestOpStop(t *testing.T) {
	ResetTestRuntime()
	RunOps(NewTestTask(), NewTestRunProvider(), OpStop)
}

func TestOpNop(t *testing.T) {
	ResetTestRuntime()
	RunOps(NewTestTask(), NewTestRunProvider(), OpNop, OpStop)
}
