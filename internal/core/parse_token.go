package core

import (
	"fmt"
	"unicode"
)

// TokenType classifies tokens from Parser.Next. Flags are OR-combined (manual hex slots).
type TokenType int

const (
	TokenEOF           TokenType = 0
	TokenIdent         TokenType = 0x0001
	TokenCompoundIdent TokenType = 0x0002
	TokenInt           TokenType = 0x0004
	TokenUint          TokenType = 0x0008
	TokenFloat         TokenType = 0x0010
	TokenBool          TokenType = 0x0020
	TokenString        TokenType = 0x0040
	TokenRawString     TokenType = 0x0080
	TokenRune          TokenType = 0x0100
	TokenLiteral       TokenType = 0x0200 // meta: OR with concrete literal kinds when useful
	TokenNull          TokenType = 0x0800
	TokenComment       TokenType = 0x1000
	TokenNewline       TokenType = 0x2000
)

// HasTokenFlag reports whether all bits in flag are set in typ.
func HasTokenFlag(typ, flag TokenType) bool {
	return typ&flag == flag
}

// IsPlainIdentToken is a simple dictionary name (not compound).
func IsPlainIdentToken(typ TokenType) bool {
	return HasTokenFlag(typ, TokenIdent) &&
		!HasTokenFlag(typ, TokenCompoundIdent)
}

// ValidatePlainIdentName reports whether name satisfies plain-identifier lexical rules.
func ValidatePlainIdentName(name String) error {
	word := string(name)
	if looksLikeNumberStart(word) {
		return fmt.Errorf("validatePlainIdent: expected simple identifier, got number %q", name)
	}
	typ, _, err := classifyNameLiteral(word)
	if err != nil {
		return fmt.Errorf("validatePlainIdent: %w", err)
	}
	if !IsPlainIdentToken(typ) {
		return fmt.Errorf("validatePlainIdent: expected simple identifier, got %q", name)
	}
	if err := validateIdentSpelling(word, false); err != nil {
		return fmt.Errorf("validatePlainIdent: %w", err)
	}
	return nil
}

func isIdentForbiddenPunct(r rune) bool {
	switch r {
	case '.', '@', '!', '^', '~', '"', '`', '\'':
		return true
	default:
		return false
	}
}

// validateIdentSpelling checks printable non-whitespace ident runes.
// allowBackslash is true for lexer lookup forms (prefix\tail); false when creating names.
func validateIdentSpelling(s string, allowBackslash bool) error {
	if s == "" {
		return fmt.Errorf("empty identifier")
	}
	first := true
	for _, r := range s {
		if first {
			if unicode.IsDigit(r) {
				return fmt.Errorf("identifier may not start with a digit")
			}
			first = false
		}
		if r == '\\' {
			if allowBackslash {
				continue
			}
			return fmt.Errorf("identifier may not contain \\")
		}
		if !unicode.IsPrint(r) || unicode.IsSpace(r) || isIdentForbiddenPunct(r) {
			return fmt.Errorf("invalid identifier character %U", r)
		}
	}
	return nil
}

// LiteralTokenType ORs TokenLiteral with a concrete literal kind for parser tokens.
func LiteralTokenType(concrete TokenType) TokenType {
	return concrete | TokenLiteral
}

// IsLiteralTokenType reports whether typ is a literal value token.
func IsLiteralTokenType(typ TokenType) bool {
	return HasTokenFlag(typ, TokenLiteral)
}
