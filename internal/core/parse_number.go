package core

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// looksLikeNumberStart reports whether a whitespace-delimited word begins as a numeric literal
// (02-lexical-parsing.md number-start rules).
func looksLikeNumberStart(lit string) bool {
	if lit == "" {
		return false
	}
	r, size := utf8.DecodeRuneInString(lit)
	if unicode.IsDigit(r) {
		return true
	}
	if r == '-' {
		rest := lit[size:]
		if rest == "" {
			return false
		}
		next, nextSize := utf8.DecodeRuneInString(rest)
		if unicode.IsDigit(next) {
			return true
		}
		if next == '.' && len(rest) > nextSize {
			third, _ := utf8.DecodeRuneInString(rest[nextSize:])
			return unicode.IsDigit(third)
		}
		return false
	}
	if r == '.' && len(lit) > size {
		next, _ := utf8.DecodeRuneInString(lit[size:])
		return unicode.IsDigit(next)
	}
	return false
}

func parseNumberLiteral(lit string) (TokenType, any, error) {
	if lit == "" {
		return 0, nil, fmt.Errorf("invalid number %q", lit)
	}

	negative := false
	body := lit
	if strings.HasPrefix(lit, "-") {
		negative = true
		body = lit[1:]
		if body == "" {
			return 0, nil, fmt.Errorf("invalid number %q", lit)
		}
	}

	// Hex digits include e/E; 0x/0b/0o prefixes are integrals only (02-lexical-parsing.md).
	prefixed := strings.HasPrefix(body, "0x") || strings.HasPrefix(body, "0X") ||
		strings.HasPrefix(body, "0b") || strings.HasPrefix(body, "0B") ||
		strings.HasPrefix(body, "0o")
	if !prefixed && strings.ContainsAny(lit, ".eE") {
		f, err := strconv.ParseFloat(lit, 64)
		if err != nil {
			return 0, nil, fmt.Errorf("invalid float %q", lit)
		}
		return LiteralTokenType(TokenFloat), Float(f), nil
	}

	if strings.HasSuffix(lit, "u") || strings.HasSuffix(lit, "U") {
		numPart := lit[:len(lit)-1]
		if negative || strings.HasPrefix(numPart, "-") {
			return 0, nil, fmt.Errorf("Uint cannot be negative: %q", lit)
		}
		u, err := parseIntegralLiteral(numPart)
		if err != nil {
			return 0, nil, err
		}
		return LiteralTokenType(TokenUint), Uint(u), nil
	}

	if negative {
		i, err := strconv.ParseInt(lit, 0, 64)
		if err != nil {
			return 0, nil, fmt.Errorf("invalid int %q", lit)
		}
		return LiteralTokenType(TokenInt), Int(i), nil
	}

	u, err := parseIntegralLiteral(lit)
	if err != nil {
		return 0, nil, err
	}
	return LiteralTokenType(TokenInt), Int(int64(u)), nil
}

func isBaseDigit(r rune, baseMarker rune) bool {
	switch baseMarker {
	case 'x', 'X':
		return unicode.IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
	case 'b', 'B':
		return r == '0' || r == '1'
	case 'o':
		return r >= '0' && r <= '7'
	default:
		return unicode.IsDigit(r)
	}
}

func parseIntegralLiteral(s string) (uint64, error) {
	s = strings.ReplaceAll(s, "_", "")
	if s == "" {
		return 0, fmt.Errorf("invalid int %q", s)
	}
	var base int
	var body string
	switch {
	case strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X"):
		base = 16
		body = s[2:]
	case strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B"):
		base = 2
		body = s[2:]
	case strings.HasPrefix(s, "0o"):
		base = 8
		body = s[2:]
	default:
		base = 10
		body = s
	}
	if body == "" {
		return 0, fmt.Errorf("invalid int %q", s)
	}
	u, err := strconv.ParseUint(body, base, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid int %q", s)
	}
	return u, nil
}
