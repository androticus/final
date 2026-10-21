package core

import "testing"

func TestValidatePlainIdentName(t *testing.T) {
	for _, name := range []string{"counter", "foo-bar", "math::sin", ">ts"} {
		if err := ValidatePlainIdentName(String(name)); err != nil {
			t.Fatalf("ValidatePlainIdentName(%q) = %v, want nil", name, err)
		}
	}
}

func TestValidatePlainIdentNameRejectsCompound(t *testing.T) {
	if err := ValidatePlainIdentName(String(".field")); err == nil {
		t.Fatal("expected error for compound path .field")
	}
}

func TestValidatePlainIdentNameRejectsNumber(t *testing.T) {
	if err := ValidatePlainIdentName(String("42")); err == nil {
		t.Fatal("expected error for numeric literal")
	}
}

func TestClassifyNameLiteralFindQ(t *testing.T) {
	typ, val, err := classifyNameLiteral("find?")
	if err != nil {
		t.Fatal(err)
	}
	if typ != TokenIdent {
		t.Fatalf("typ = %v, want TokenIdent", typ)
	}
	if val.(String) != "find?" {
		t.Fatalf("val = %v, want find?", val)
	}
}

func TestValidatePlainIdentNameAllowsTrailingQ(t *testing.T) {
	if err := ValidatePlainIdentName(String("find?")); err != nil {
		t.Fatalf("ValidatePlainIdentName(find?) = %v, want nil", err)
	}
}
