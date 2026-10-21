package core

// In Final, fetching and storing values, and operations (methods) exposed
// by compound types, including native Go types, are done via the Ref interface family.

// Note: unlike Forth, which uses raw memory addresses, the addresses of Go values
// cannot be exposed directly, because they can change, such as when a slice grows.
//
// Refs are transient: bare @, !, and ^ consume the ref in the axiom prologue.
// Read-modify-write: dup @ … swap ! (Forth-style ref on TOS for store).

type ValueRef interface {
	Fetch() any
	Store(v any) // panics if readonly
}
