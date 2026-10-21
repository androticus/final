package core

import (
	"testing"
)

// sessionProbe is a host Go type registered into a session module via RegisterTypes.
// It binds to its dynamically created *Type (stamped by the `new` class method) so
// instance dispatch resolves through the Typed interface.
// Disposable Final name uses __SessProbe__ (double-underscore decoration, not CapCase).
type sessionProbe struct {
	n   int
	typ *Type
}

func (c *sessionProbe) FinalType() *Type { return c.typ }

func sessProbeTypeReg() TypeReg {
	return NewTypeReg[*sessionProbe]("__SessProbe__", func(ty *Type) {
		ty.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
			f.RP.DataStack.Push(&sessionProbe{typ: ths.Obj.(*Type)})
		})
		ty.registerNativeMethod("inc", func(f *RunFrame, ths *TypeObjPair) {
			ths.Obj.(*sessionProbe).n++
		})
		ty.registerNativeMethod("get", func(f *RunFrame, ths *TypeObjPair) {
			f.RP.DataStack.Push(Int(ths.Obj.(*sessionProbe).n))
		})
		ty.registerProperty("count", "Int",
			func(inst Any) Any { return Int(inst.(*sessionProbe).n) },
			nil,
		)
	})
}

func TestSessionRegisterTypesDispatch(t *testing.T) {
	ResetTestRuntime()
	session, err := NewSession(&SessionOptions{Kind: SessionKindLocal})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	if err := session.RegisterTypes([]TypeReg{sessProbeTypeReg()}); err != nil {
		t.Fatal(err)
	}

	// Drive the session's module task directly and inspect the data stack, avoiding
	// dependence on a stdout print-word surface.
	rp := NewTestRunProvider()
	tr, err := NewTextReaderFromString(`__SessProbe__.new^ >ts t:.inc^ t:.inc^ t:.inc^ t:.get^`)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	if err := interpretSession(session.task, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 3 {
		t.Fatalf("__SessProbe__.get after 3 inc = %v, want 3", got)
	}
}

func TestSessionRegisterTypesModuleScoped(t *testing.T) {
	ResetTestRuntime()
	session, err := NewSession(&SessionOptions{Kind: SessionKindLocal})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	if err := session.RegisterTypes([]TypeReg{sessProbeTypeReg()}); err != nil {
		t.Fatal(err)
	}

	dc := session.task.dc
	typ := dc.tr.Get("__SessProbe__")
	if typ == nil {
		t.Fatal("__SessProbe__ missing from session registrar")
	}
	if GetTypeRegistrar().Get("__SessProbe__") != nil {
		t.Fatal("__SessProbe__ leaked into global core registrar")
	}
	if w := dc.dict.GetByName("__SessProbe__"); w == nil || !w.IsTypeWord() {
		t.Fatalf("__SessProbe__ type word missing from session dict: %v", w)
	}
	if globalDictionaryContext.dict.GetByName("__SessProbe__") != nil {
		t.Fatal("__SessProbe__ type word leaked into core dict")
	}
	// property field type resolved via parent fallback to core Int
	if pm := typ.PropertyByName("count"); pm == nil {
		t.Fatal("count property not registered/resolved")
	}
}

func TestSessionRegisterTypesErrors(t *testing.T) {
	ResetTestRuntime()
	session, err := NewSession(&SessionOptions{Kind: SessionKindLocal})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	if err := session.RegisterTypes([]TypeReg{NewTypeReg[*sessionProbe]("", func(*Type) {})}); err == nil {
		t.Fatal("expected error for empty name")
	}
	if err := session.RegisterTypes([]TypeReg{NewTypeReg[*sessionProbe]("NoReg", nil)}); err == nil {
		t.Fatal("expected error for nil Reg")
	}
	dup := sessProbeTypeReg()
	if err := session.RegisterTypes([]TypeReg{dup, dup}); err == nil {
		t.Fatal("expected duplicate-name error")
	}
}
