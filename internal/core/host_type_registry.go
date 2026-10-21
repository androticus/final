package core

import (
	"fmt"
	"reflect"
)

// hostTypeRegistry maps a host Go type (reflect.Type) to its registered Final *Type.
//
// Embedders register plain Go structs (e.g. *Lissajous) that do not implement Typed.
// When such a value reaches NewTypeObjPair on the data stack, there is no FinalType()
// to consult, so we resolve its concrete Go type here as a last resort before panicking.
//
// No locking: embedding is app-global and registers all host types once on startup
// (single session, before any run), so there is never a concurrent writer or reader.
var hostTypeRegistry = map[reflect.Type]*Type{}

// registerHostGoType records goType -> t. Panics on a duplicate goType, since a single
// Go type must map to exactly one Final type in an app-global embedding.
func registerHostGoType(goType reflect.Type, t *Type) {
	if goType == nil || t == nil {
		return
	}
	if existing, ok := hostTypeRegistry[goType]; ok {
		panic(fmt.Errorf("registerHostGoType: %s already registered as %q", goType, existing.name))
	}
	hostTypeRegistry[goType] = t
}

// RegisterHostType records T → t in the host Go-type map used by NewTypeObjPair
// when the value does not implement Typed (same map as Session.RegisterTypes).
func RegisterHostType[T any](t *Type) {
	registerHostGoType(reflect.TypeFor[T](), t)
}

// lookupHostType resolves obj's concrete Go type to a registered Final *Type, or nil.
func lookupHostType(obj any) *Type {
	if obj == nil {
		return nil
	}
	return hostTypeRegistry[reflect.TypeOf(obj)]
}
