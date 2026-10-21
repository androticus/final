package core

import (
	"fmt"
	"strings"
	"testing"
	"unicode"
)

func TestIsDangerousRune(t *testing.T) {
	dangerous := []rune{
		0x00, 0x08, 0x1B, 0x7F,
		0x200E, 0x200F, 0x061C,
		0x202A, 0x202E,
		0x2066, 0x2069,
		0x200B, 0x200C, 0x200D, 0x2060, 0xFEFF,
		0xFE00, 0xFE0F, 0xE0100, 0xE01EF,
		0xE0001, 0xE007F,
		0x3164, 0xFFA0,
		0x00AD, 0x2061, 0x2064,
	}
	for _, r := range dangerous {
		if !IsDangerousRune(r) {
			t.Errorf("IsDangerousRune(%U) = false, want true", r)
		}
		if !ContainsDangerousUnicode(string(r)) {
			t.Errorf("ContainsDangerousUnicode(%U) = false, want true", r)
		}
	}
	safe := []rune{'\t', '\n', '\r', ' ', 'A', 'é', '日', '🙂', '_', '+'}
	for _, r := range safe {
		if IsDangerousRune(r) {
			t.Errorf("IsDangerousRune(%U) = true, want false", r)
		}
	}
}

func TestDangerousRuneSourceHardFail(t *testing.T) {
	zwsp := string(rune(0x200B))
	cases := []string{
		"foo" + zwsp + "bar",
		"// hello" + zwsp,
		"/* hello" + zwsp + " */",
		"`raw" + zwsp + "text`",
		"\"str" + zwsp + "ing\"",
	}
	for _, src := range cases {
		p := NewParser(nil)
		_, _, err := p.ParseLine(src)
		if err == nil {
			t.Fatalf("ParseLine(%q): want dangerous-rune error", src)
		}
		msg := err.Error()
		if !strings.Contains(msg, "U+200B") {
			t.Fatalf("error %q: want U+200B", msg)
		}
		if !strings.Contains(msg, `\x`) || !strings.Contains(msg, `\u`) || !strings.Contains(msg, `\U`) {
			t.Fatalf("error %q: want \\x / \\u / \\U hint", msg)
		}
	}
}

func TestHangulFillerInIdentIsSourceError(t *testing.T) {
	src := "foo" + string(rune(0x3164)) + "bar"
	p := NewParser(nil)
	_, _, err := p.ParseLine(src)
	if err == nil {
		t.Fatal("want source error for Hangul filler in ident")
	}
	if !strings.Contains(err.Error(), "U+3164") {
		t.Fatalf("error %q: want U+3164", err.Error())
	}
}

func TestQuotedStringUnicodeEscapes(t *testing.T) {
	tokens := parseLineTokens(t, `"\u200B" "\U000E0100" "\x08"`)
	if tokens[0].Value != String("\u200b") {
		t.Fatalf("\\u200B value = %q", tokens[0].Value)
	}
	if tokens[1].Value != String(string(rune(0xE0100))) {
		t.Fatalf("\\U000E0100 value = %q (%U)", tokens[1].Value, []rune(string(tokens[1].Value.(String)))[0])
	}
	if tokens[2].Value != String("\x08") {
		t.Fatalf("\\x08 value = %q", tokens[2].Value)
	}
}

func TestRuneLiteralUnicodeEscapes(t *testing.T) {
	tokens := parseLineTokens(t, `'\u200B' '\U000E0100' '\x08'`)
	if tokens[0].Value != Rune(0x200B) {
		t.Fatalf("\\u200B rune = %v", tokens[0].Value)
	}
	if tokens[1].Value != Rune(0xE0100) {
		t.Fatalf("\\U000E0100 rune = %v", tokens[1].Value)
	}
	if tokens[2].Value != Rune(0x08) {
		t.Fatalf("\\x08 rune = %v", tokens[2].Value)
	}
}

func TestQuotedStringInvalidEscapeErrors(t *testing.T) {
	parseLineError(t, `"\q"`)
	parseLineError(t, `"\xGG"`)
	parseLineError(t, `"\u12"`)
	parseLineError(t, `"\U12345678"`) // not a valid code point
}

func TestUnicodeIdentsAndCompoundMembers(t *testing.T) {
	tokens := parseLineTokens(t, "café 🙂name t:.aé🙂日")
	if !IsPlainIdentToken(tokens[0].Type) || tokens[0].Value != String("café") {
		t.Fatalf("token0 = %+v, want ident café", tokens[0])
	}
	if !IsPlainIdentToken(tokens[1].Type) || tokens[1].Value != String("🙂name") {
		t.Fatalf("token1 = %+v, want ident 🙂name", tokens[1])
	}
	if !HasTokenFlag(tokens[2].Type, TokenCompoundIdent) {
		t.Fatalf("token2 type = %v, want compound", tokens[2].Type)
	}
	expectCompoundPath(t, tokens[2].Value, ".", "aé🙂日")
}

func TestValidatePlainIdentNameUnicode(t *testing.T) {
	if err := ValidatePlainIdentName(String("aé🙂日")); err != nil {
		t.Fatalf("ValidatePlainIdentName(aé🙂日) = %v, want nil", err)
	}
	if err := ValidatePlainIdentName(String(`foo\bar`)); err == nil {
		t.Fatal("ValidatePlainIdentName should reject backslash when creating names")
	}
}

func TestIdentSpellingPrintable(t *testing.T) {
	if unicode.IsPrint('🙂') && validateIdentSpelling("🙂", false) != nil {
		t.Fatal("emoji should be a legal ident rune")
	}
	if err := validateIdentSpelling("foo\"bar", false); err == nil {
		t.Fatal("want error for quote in ident")
	}
}

func TestErrDangerousSourceRuneMessage(t *testing.T) {
	msg := errDangerousSourceRune(0x200B).Error()
	want := fmt.Sprintf("dangerous character %U", rune(0x200B))
	if !strings.Contains(msg, want) {
		t.Fatalf("msg %q: want %q", msg, want)
	}
}
