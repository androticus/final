package core

import "testing"

func TestIsTruthy(t *testing.T) {
	members := []any{Bool(true), Bool(false), Int(0), Int(1), Uint(0), Float(0), String(""), String("x")}
	for _, v := range members {
		if !IsTruthy(v) {
			t.Fatalf("%T %#v should be Truthy member", v, v)
		}
	}
	non := []any{nil, &AnyList{}, NewAnyMap(), struct{}{}}
	for _, v := range non {
		if IsTruthy(v) {
			t.Fatalf("%T should not be Truthy member", v)
		}
	}
}

func TestAnyToTruthy(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want Bool
	}{
		{"bool_true", Bool(true), true},
		{"bool_false", Bool(false), false},
		{"int_nonzero", Int(3), true},
		{"int_zero", Int(0), false},
		{"uint_nonzero", Uint(1), true},
		{"uint_zero", Uint(0), false},
		{"float_nonzero", Float(0.5), true},
		{"float_zero", Float(0), false},
		{"string_nonempty", String("x"), true},
		{"string_empty", String(""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AnyToTruthy(tt.in); got != tt.want {
				t.Fatalf("AnyToTruthy(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestAnyToTruthyNonMemberPanics(t *testing.T) {
	// Note: Bool/Int/Uint/Float/String are type aliases to Go builtins (types.go),
	// so native Go bool/int/etc. are indistinguishable and cannot be rejected here.
	cases := []struct {
		name string
		in   any
	}{
		{"nil", nil},
		{"list", &AnyList{}},
		{"map", NewAnyMap()},
		{"struct", struct{}{}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			expectPanicSubstr(t, "Truthy", func() { AnyToTruthy(tt.in) })
		})
	}
}

func TestIsNumeric(t *testing.T) {
	if !IsNumeric(Int(1)) || !IsNumeric(Uint(1)) || !IsNumeric(Float(1)) {
		t.Fatal("expected numeric members")
	}
	if IsNumeric(Bool(true)) || IsNumeric(String("1")) {
		t.Fatal("expected non-numeric values to return false")
	}
}

func TestIsIntegral(t *testing.T) {
	if !IsIntegral(Int(1)) || !IsIntegral(Uint(1)) {
		t.Fatal("expected integral members")
	}
	if IsIntegral(Float(1)) || IsIntegral(Bool(true)) {
		t.Fatal("expected non-integral values to return false")
	}
}
