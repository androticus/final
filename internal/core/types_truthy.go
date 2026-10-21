package core

import "fmt"

// IsTruthy reports whether v is a Truthy pseudo-interface member (membership only).
func IsTruthy(v any) bool {
	switch v.(type) {
	case Bool, Int, Uint, Float, String:
		return true
	default:
		return false
	}
}

// AnyToTruthy converts a Truthy member to Bool. Panics if v is not a Truthy member.
func AnyToTruthy(v any) Bool {
	switch x := v.(type) {
	case Bool:
		return x
	case Int:
		return x != 0
	case Uint:
		return x != 0
	case Float:
		return x != 0
	case String:
		return x != ""
	default:
		panic(fmt.Errorf("value is not Truthy: %T", v))
	}
}

// IsNumeric reports whether v is a Numeric pseudo-interface member.
func IsNumeric(v any) bool {
	switch v.(type) {
	case Int, Uint, Float:
		return true
	default:
		return false
	}
}

// IsIntegral reports whether v is an Integral pseudo-interface member.
func IsIntegral(v any) bool {
	switch v.(type) {
	case Int, Uint:
		return true
	default:
		return false
	}
}
