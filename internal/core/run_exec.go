package core

import "fmt"

// popExecXT pops an execution token from DS (VirtualAddr or Int-encoded XT).
func popExecXT(ds *Stack[any]) VirtualAddr {
	switch v := ds.Pop().(type) {
	case Int:
		return VirtualAddr(v)
	default:
		panic(fmt.Errorf("exec: TOS is %T, want VirtualAddr or Int XT", v))
	}
}
