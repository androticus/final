package core

import "fmt"

// ContainsDangerousUnicode reports whether s contains any of the Unicode
// characters commonly abused for Trojan Source, invisible payloads,
// or display-vs-logic attacks (Bidi controls, zero-width/format chars,
// variation selectors, tags, Hangul fillers, etc.).
//
// Designed for a zero-tolerance policy: raw occurrences are forbidden;
// legitimate needs must use \u / \U escapes.
func ContainsDangerousUnicode(s string) bool {
	for _, r := range s {
		if IsDangerousRune(r) {
			return true
		}
	}
	return false
}

func IsDangerousRune(r rune) bool {
	switch {
	// C0 controls that can overwrite display (Backspace, etc.)
	case r < 0x20 && r != '\t' && r != '\n' && r != '\r':
		return true
	case r == 0x7F: // DEL
		return true

	// Bidi marks
	case r == 0x200E || r == 0x200F: // LRM, RLM
		return true
	case r == 0x061C: // Arabic Letter Mark
		return true

	// Bidi embeddings / overrides
	case r >= 0x202A && r <= 0x202E: // LRE, RLE, PDF, LRO, RLO
		return true

	// Bidi isolates
	case r >= 0x2066 && r <= 0x2069: // LRI, RLI, FSI, PDI
		return true

	// Zero-width / format
	case r == 0x200B || r == 0x200C || r == 0x200D: // ZWSP, ZWNJ, ZWJ
		return true
	case r == 0x2060: // Word Joiner
		return true
	case r == 0xFEFF: // ZWNBSP / BOM (reject even as leading BOM for simplicity)
		return true

	// Variation selectors (Glassworm-style)
	case r >= 0xFE00 && r <= 0xFE0F: // VS1–VS16
		return true
	case r >= 0xE0100 && r <= 0xE01EF: // VS17–VS256
		return true

	// Tag characters (ASCII smuggling)
	case r >= 0xE0001 && r <= 0xE007F:
		return true

	// Hangul fillers (invisible letter-like)
	case r == 0x3164 || r == 0xFFA0:
		return true

	// Other common default-ignorable / format chars sometimes abused
	case r == 0x00AD: // Soft Hyphen
		return true
	case r >= 0x2061 && r <= 0x2064: // Invisible math operators
		return true
	}

	return false
}

func errDangerousSourceRune(r rune) error {
	return fmt.Errorf("dangerous character %U not allowed in source; include it in a quoted string or rune literal using \\x, \\u, or \\U escapes", r)
}
