package core

import (
	"fmt"
	"reflect"
)

// TypeReg encapsulates registration of one native host type into a session module.
// Its fields are unexported: construct it only via NewTypeReg[T], which derives the Go
// type, zero, and coerce functors from T.
//
// name is the Final (CapCase) type name. goType is the concrete Go type of instances
// (usually a pointer, e.g. *Lissajous); it lets RegisterTypes resolve stack values of
// that Go type back to this Final type at runtime without the host type implementing
// Typed. zero and checkCoerce are the type's getZero/checkCoerceType functors. reg
// populates the freshly created *Type with instance methods, class methods, and
// properties (copy the core registration patterns, e.g. StringBuilder).
//
// Types are registered in slice order; any dependent field type (a core type or an
// earlier entry in the same slice) already exists at registration time, so no cross-type
// deferred resolution is needed.
type TypeReg struct {
	name        string
	goType      reflect.Type
	zero        func() any
	checkCoerce func(value any) any
	reg         func(t *Type)
}

// NewTypeReg builds a TypeReg for host Go type T, deriving goType via reflect.TypeFor[T],
// a T-derived zero via MakeHostTypeZero[T], and a strongly typed coerce via
// MakeHostTypeCheckCoerce[T]. Pass the concrete instance type: NewTypeReg[*Lissajous]
// ("Lissajous", reg) for a reference type, or NewTypeReg[Point]("Point", reg) for a value type.
func NewTypeReg[T any](name string, reg func(t *Type)) TypeReg {
	return TypeReg{
		name:        name,
		goType:      reflect.TypeFor[T](),
		zero:        MakeHostTypeZero[T](),
		checkCoerce: MakeHostTypeCheckCoerce[T](name),
		reg:         reg,
	}
}

// RegisterTypes registers each TypeReg into this session's module (its own TypeRegistrar
// and dictionary), installing a type-proxy word per type. Call once after NewSession and
// before any Send/Run that uses the types. It serializes with Send on the session lock.
func (s *Session) RegisterTypes(regs []TypeReg) error {
	if s == nil {
		return fmt.Errorf("RegisterTypes: nil Session")
	}
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrSessionClosed
	}
	s.mu.Unlock()

	dc := s.task.dc
	if dc == nil || dc.tr == nil || dc.dict == nil {
		return fmt.Errorf("RegisterTypes: session task has no module context")
	}
	for _, r := range regs {
		if r.name == "" {
			return fmt.Errorf("RegisterTypes: empty type name (use NewTypeReg)")
		}
		if r.reg == nil {
			return fmt.Errorf("RegisterTypes: nil reg for type %q (use NewTypeReg)", r.name)
		}
		if dc.tr.Get(r.name) != nil {
			return fmt.Errorf("RegisterTypes: type name %q already registered", r.name)
		}
		t := &Type{
			name:            r.name,
			registrar:       dc.tr,
			getZero:         r.zero,
			checkCoerceType: r.checkCoerce,
		}
		r.reg(t)
		dc.tr.RegisterCurrentType(t)
		dc.tr.ResolveRegistrations()
		addTypeDictionaryWord(dc, t)
		registerHostGoType(r.goType, t)
	}
	return nil
}
