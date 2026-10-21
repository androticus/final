package core

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (p *Parser) parseRuneToken() (Token, error) {
	startLine, startCol := p.line, p.liveCol()
	if _, err := p.takeSourceRune(); err != nil {
		return Token{}, err
	}
	if !p.hasRune() {
		return Token{}, fmt.Errorf("unclosed rune literal at line %d", startLine)
	}

	var val rune
	ch := p.peekRune()

	if ch == '\\' {
		if _, err := p.takeSourceRune(); err != nil {
			return Token{}, err
		}
		r, b, usedByte, err := p.decodeGoEscape()
		if err != nil {
			return Token{}, fmt.Errorf("rune escape at line %d: %w", startLine, err)
		}
		if usedByte {
			val = rune(b)
		} else {
			val = r
		}
	} else {
		var err error
		val, err = p.takeSourceRune()
		if err != nil {
			return Token{}, err
		}
		if val == '\'' {
			return Token{}, fmt.Errorf("empty rune literal at line %d", startLine)
		}
	}

	if !p.hasRune() || p.peekRune() != '\'' {
		return Token{}, fmt.Errorf("multi-rune or unclosed rune literal at line %d", startLine)
	}
	if _, err := p.takeSourceRune(); err != nil {
		return Token{}, err
	}

	if !utf8.ValidRune(val) {
		return Token{}, fmt.Errorf("invalid rune code point at line %d", startLine)
	}
	return Token{Type: LiteralTokenType(TokenRune), Value: Rune(val), Line: startLine, Col: startCol}, nil
}

// decodeGoEscape consumes the character after a backslash and returns the decoded value.
// \xHH yields a single byte (usedByte true); \u/\U yield a Unicode rune.
func (p *Parser) decodeGoEscape() (r rune, b byte, usedByte bool, err error) {
	if !p.hasRune() {
		return 0, 0, false, fmt.Errorf("unclosed escape")
	}
	esc, err := p.takeSourceRune()
	if err != nil {
		return 0, 0, false, err
	}
	switch esc {
	case 'a':
		return '\a', 0, false, nil
	case 'b':
		return '\b', 0, false, nil
	case 'f':
		return '\f', 0, false, nil
	case 'n':
		return '\n', 0, false, nil
	case 'r':
		return '\r', 0, false, nil
	case 't':
		return '\t', 0, false, nil
	case 'v':
		return '\v', 0, false, nil
	case '\\', '\'', '"':
		return esc, 0, false, nil
	case 'x':
		u, err := p.readHexDigits(2)
		if err != nil {
			return 0, 0, false, fmt.Errorf("\\x escape: %w", err)
		}
		return 0, byte(u), true, nil
	case 'u':
		u, err := p.readHexDigits(4)
		if err != nil {
			return 0, 0, false, fmt.Errorf("\\u escape: %w", err)
		}
		val := rune(u)
		if !utf8.ValidRune(val) {
			return 0, 0, false, fmt.Errorf("invalid \\u code point %U", val)
		}
		return val, 0, false, nil
	case 'U':
		u, err := p.readHexDigits(8)
		if err != nil {
			return 0, 0, false, fmt.Errorf("\\U escape: %w", err)
		}
		val := rune(u)
		if !utf8.ValidRune(val) {
			return 0, 0, false, fmt.Errorf("invalid \\U code point")
		}
		return val, 0, false, nil
	default:
		return 0, 0, false, fmt.Errorf("invalid escape \\%c", esc)
	}
}

func (p *Parser) readHexDigits(n int) (uint64, error) {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		if !p.hasRune() {
			return 0, fmt.Errorf("want %d hex digits", n)
		}
		r := p.peekRune()
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return 0, fmt.Errorf("invalid hex %q", sb.String())
		}
		ch, err := p.takeSourceRune()
		if err != nil {
			return 0, err
		}
		sb.WriteRune(ch)
	}
	s := sb.String()
	u, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return 0, err
	}
	return u, nil
}
