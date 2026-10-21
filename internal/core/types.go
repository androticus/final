package core

// This file contains the types that are used in the Final language
// Some are just aliases for Go types, some are new types that are specific to the Final language

//Execution Environment Types

type IP = int  // instruction pointer; index into the CodeTable
type CTP = int // code table pointer; index into the CodeTable
type DTP = int // data table pointer; index into the DataTable
// Final Simple Types
// All simple types are immutable

type Bool = bool
type Int = int
type Uint = uint
type Float = float64
type Rune = Int // a 32 bit unicode character; a synonym for *Int*
type String = string
type XT = Int // execution token: dictionary ordinal ([01-types.md](../../docs/00-high-level/01-types.md))
type Any = any

// NaturalStringer formats a data-stack cell for ds$ and similar debug output.
// Use this instead of fmt.Stringer so Go maps and other host types are never
// printed wholesale. Implementations must return a short summary only.
type NaturalStringer interface {
	NaturalString() string
}

// Final Compound Types
// All compound types are mutable
// List: see list.go. Map types below.
// Map: see map.go.
