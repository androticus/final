package core

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const maxLastLineOutput = 80

// locatedError is a compile/interpret fault already formatted with file/line/caret.
// Nested locateFault / formatFault leave it unchanged.
type locatedError struct {
	msg string
}

func (e *locatedError) Error() string {
	if e == nil {
		return "<nil locatedError>"
	}
	return e.msg
}

// IsLocatedFaultMessage reports whether s is a formatFault caret report.
func IsLocatedFaultMessage(s string) bool {
	return strings.HasPrefix(s, "Error compiling")
}

func formatFault(err error, p *Parser) string {
	if err == nil {
		return ""
	}
	if le, ok := err.(*locatedError); ok {
		return le.Error()
	}
	if IsLocatedFaultMessage(err.Error()) {
		return err.Error()
	}
	if p == nil || !p.hasParserLocation() {
		return err.Error()
	}
	return formatFaultAt(p.file, p.line, p.lineText, p.col, err)
}

// locateFault wraps err as a locatedError when the parser has a fault site.
func locateFault(err error, p *Parser) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(*locatedError); ok {
		return err
	}
	// Preserve RunFault (and any already-formatted caret text in its Message).
	if fault, ok := err.(*RunFault); ok {
		return fault
	}
	if IsLocatedFaultMessage(err.Error()) {
		return err
	}
	if p == nil || !p.hasParserLocation() {
		return err
	}
	return &locatedError{msg: formatFault(err, p)}
}

func formatFaultAt(file string, line int, lineText string, col int, err error) string {
	if err == nil {
		return ""
	}
	if col <= 0 {
		return err.Error()
	}
	if lineText == "" {
		col = 1
	}
	dispLine, dispCol, elided := elideLastLineForFault(lineText, col)
	caretPad := dispCol - 1
	if caretPad < 0 {
		caretPad = 0
	}
	var b strings.Builder
	if file != "" {
		fmt.Fprintf(&b, "Error compiling file %q\nline %d; column %d :\n", file, line, col)
	} else {
		fmt.Fprintf(&b, "Error compiling; line %d; column %d :\n", line, col)
	}
	b.WriteString("> ")
	if elided {
		b.WriteString("... ")
	}
	b.WriteString(dispLine)
	b.WriteByte('\n')
	b.WriteString("> ")
	if elided {
		b.WriteString("... ")
	}
	b.WriteString(strings.Repeat(" ", caretPad))
	b.WriteByte('^')
	b.WriteByte('\n')
	b.WriteString("Error: ")
	b.WriteString(err.Error())
	return b.String()
}

// elideLastLineForFault keeps the last maxLastLineOutput runes of lineText.
// Returns display text (tabs → space), caret column within that text, and whether elided.
func elideLastLineForFault(lineText string, col int) (disp string, dispCol int, elided bool) {
	var b strings.Builder
	b.Grow(len(lineText))
	for _, r := range lineText {
		if r == '\t' {
			r = ' '
		}
		b.WriteRune(r)
	}
	normalized := b.String()
	n := utf8.RuneCountInString(normalized)
	if n <= maxLastLineOutput {
		dispCol = col
		if dispCol < 1 {
			dispCol = 1
		}
		if n == 0 {
			dispCol = 1
		} else if dispCol > n+1 {
			dispCol = n + 1
		}
		return normalized, dispCol, false
	}
	disp = lastNRunes(normalized, maxLastLineOutput)
	skip := n - maxLastLineOutput
	dispCol = col - skip
	if dispCol < 1 {
		dispCol = 1
	}
	kept := utf8.RuneCountInString(disp)
	if dispCol > kept+1 {
		dispCol = kept + 1
	}
	return disp, dispCol, true
}

// lastNRunes returns the last n Unicode runes of s (or all of s if shorter).
func lastNRunes(s string, n int) string {
	if n <= 0 || s == "" {
		return ""
	}
	end := len(s)
	for i := 0; i < n && end > 0; i++ {
		_, size := utf8.DecodeLastRuneInString(s[:end])
		end -= size
	}
	return s[end:]
}
