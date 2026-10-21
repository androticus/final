package core

import (
	"os"
	"runtime"
	"sync"

	"golang.org/x/term"
)

type Sys struct {
	argv   *StringList
	params *StringList
	isTty  Bool
}

func NewSys() *Sys {
	return &Sys{
		argv:   NewStringList(),
		params: NewStringListWithCapacity(0), // rarely used, so don't pre-allocate
		isTty:  Bool(term.IsTerminal(int(os.Stdin.Fd()))),
	}
}

func (s *Sys) FinalType() *Type {
	return sysType
}

func registerSysMembers(t *Type) {
	t.registerProperty("argv", "StringList",
		func(inst Any) Any { return inst.(*Sys).argv },
		nil,
	)
	t.registerProperty("params", "StringList",
		func(inst Any) Any { return inst.(*Sys).params },
		nil,
	)
	t.registerProperty("isTty", "Bool",
		func(inst Any) Any { return inst.(*Sys).isTty },
		nil,
	)
	t.registerProperty("coreCount", "Int",
		func(inst Any) Any { return Int(runtime.NumCPU()) },
		nil,
	)
	// ( program:s args:StringList timeoutS:i -- out:s code:i true | out:s err:s false )
	// stderr merged into captured text. Nonzero exit is still modal true (check code).
	// timeoutS: whole seconds; 0 = no deadline. Timeout → err "timeout" + partial out.
	t.registerNativeMethodSec("exec?", func(f *RunFrame, _ *TypeObjPair) {
		ds := f.RP.DataStack
		timeoutS := int(ds.Pop().(Int))
		argsList := ds.Pop().(*StringList)
		program := string(ds.Pop().(String))
		args := make([]string, argsList.Len())
		for i := 0; i < argsList.Len(); i++ {
			args[i] = argsList.At(i)
		}
		ctx, cancel := sysExecContext(timeoutS)
		defer cancel()
		out, code, err := runSysExecCombined(ctx, program, args)
		pushSysExecModal(ds, out, code, err)
	}, SR_System)
}

func CreateRegisterSysType() {
	t := &Type{
		name:            "Sys",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Sys]("Sys"),
	}
	registerSysMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	sysType = t
}

var (
	sys     *Sys // process singleton; filled by initGlobalSys after type resolution
	sysType *Type

	initSysOnce sync.Once
)

func initGlobalSys() {
	initSysOnce.Do(func() {
		sys = NewSys()
	})
}

// SetSysLaunch replaces the process Sys argv/params snapshots.
// argv is a copy of os.Args (may include interpreter flags). params is the
// Python-style program vector; empty for interactive TUI/REPL with no script.
func SetSysLaunch(argv, params []string) {
	if sys == nil {
		panic("SetSysLaunch: Sys not initialized")
	}
	sys.argv = NewStringListFromSlice(append([]string(nil), argv...))
	sys.params = NewStringListFromSlice(append([]string(nil), params...))
}

func init() {
	CreateRegisterSysType()
}

var _ Typed = (*Sys)(nil)
