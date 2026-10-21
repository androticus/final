package core

import (
	"fmt"
	"strings"
)

// Compound path segment vocabulary (fixed). Only these punctuation/affix cells may appear
// between name atoms inside a compound token; splitCompoundSegments rejects anything else.
//
// Lone ">" is not a compound segment (comparison operator). Other ">" words (>mark, >resolve,
// >ts, ts>) stay simple idents when whitespace-delimited.
//
// Native-class proxy identifiers follow host-language rules and are extremely unlikely to
// overlap Final's compound punctuation; no special overlap heuristics are required.
const (
	segDot        = "."
	segIndex      = "()"
	segAt         = "@"
	segBang       = "!"
	segAtIndex    = "@()"
	segBangIndex  = "!()"
	segAtBrace    = "@{}"
	segBangBrace  = "!{}"
	segQueryBrace = "?{}"
	segAtMulti    = "@[]"
	segBangMulti  = "![]"
	segTilde      = "~" // ref materialize suffix on .<name> ( .prop~ )
	segCaret      = "^" // method invoke (compound affix + standalone axiom)
)

// compoundPathNoSelector marks a compound path with no explicit stack-prefix (t:, …).
const compoundPathNoSelector ContextStackIndex = -1

// compoundPathValue is the frozen compound token (canonical path string + optional stack selector from d:/t:/x:/y: prefix).
type compoundPathValue struct {
	path          String            // space-separated segment cells from splitCompoundSegments
	stackSelector ContextStackIndex // compoundPathNoSelector when absent; else csDataStack, csTStack, csXStack, csYStack, …
}

func newCompoundPath(segs []string, selector ContextStackIndex) *compoundPathValue {
	return &compoundPathValue{
		path:          String(strings.Join(segs, " ")),
		stackSelector: selector,
	}
}

// peelEnabledCompoundStackSelector strips a leading d:/t:/x:/y: prefix from a compound literal.
// Lone selector (e.g. t:) yields empty rest; dispatch reports missing operand/path.
func peelEnabledCompoundStackSelector(literal string) (rest string, idx ContextStackIndex, ok bool, err error) {
	switch {
	case strings.HasPrefix(literal, "d:"):
		return literal[2:], csDataStack, true, nil
	case strings.HasPrefix(literal, "t:"):
		return literal[2:], csTStack, true, nil
	case strings.HasPrefix(literal, "x:"):
		return literal[2:], csXStack, true, nil
	case strings.HasPrefix(literal, "y:"):
		return literal[2:], csYStack, true, nil
	case strings.HasPrefix(literal, "z:"):
		return literal[2:], csZStack, true, nil
	case strings.HasPrefix(literal, "o:"):
		return literal[2:], csOStack, true, nil
	default:
		return literal, compoundPathNoSelector, false, nil
	}
}

func classifyCompoundLiteral(literal string, selector ContextStackIndex) (TokenType, any, error) {
	var segs []string
	var err error
	if hasCompoundPathPunct(literal) || strings.HasPrefix(literal, segDot) {
		segs, err = splitCompoundSegments(literal)
	} else {
		if err := validateCompoundNameSegment(literal); err != nil {
			return 0, nil, err
		}
		segs = []string{literal}
	}
	if err != nil {
		return 0, nil, err
	}
	if err := validateCompoundPathSegments(segs); err != nil {
		return 0, nil, err
	}
	return TokenCompoundIdent, newCompoundPath(segs, selector), nil
}

// validateCompoundPathSegments enforces path shape rules shared by lexer and walker.
func validateCompoundPathSegments(segs []string) error {
	for i, seg := range segs {
		if seg == segCaret && i != len(segs)-1 {
			return fmt.Errorf("compound: %q must be the terminal segment", segCaret)
		}
		if seg == segTilde && i != len(segs)-1 {
			return fmt.Errorf("compound: %q must be the terminal segment", segTilde)
		}
	}
	return nil
}

// readCompoundStackArrowSegment rejects fused ">" shuffle cells in compound paths.
func readCompoundStackArrowSegment(literal string, i int) (seg string, next int, err error) {
	if i >= len(literal) || literal[i] != '>' {
		return "", i, fmt.Errorf("readCompoundStackArrowSegment: expected >, got %q", literal[i:])
	}
	return "", i, fmt.Errorf("invalid compound stack-arrow segment at %q (use d:/t:/x:/y:/z:/o: prefix)", literal[i:])
}

// readCompoundAffixSegment consumes @ or ! plus optional () / {} / [] lookahead (single segment cell).
func readCompoundAffixSegment(literal string, i int) (seg string, next int, err error) {
	if i >= len(literal) {
		return "", i, fmt.Errorf("readCompoundAffixSegment: empty input")
	}
	switch literal[i] {
	case '@':
		if i+len(segAtMulti) <= len(literal) && literal[i:i+len(segAtMulti)] == segAtMulti {
			return segAtMulti, i + len(segAtMulti), nil
		}
		if i+len(segAtIndex) <= len(literal) && literal[i:i+len(segAtIndex)] == segAtIndex {
			return segAtIndex, i + len(segAtIndex), nil
		}
		if i+len(segAtBrace) <= len(literal) && literal[i:i+len(segAtBrace)] == segAtBrace {
			return segAtBrace, i + len(segAtBrace), nil
		}
		return segAt, i + 1, nil
	case '!':
		if i+len(segBangMulti) <= len(literal) && literal[i:i+len(segBangMulti)] == segBangMulti {
			return segBangMulti, i + len(segBangMulti), nil
		}
		if i+len(segBangIndex) <= len(literal) && literal[i:i+len(segBangIndex)] == segBangIndex {
			return segBangIndex, i + len(segBangIndex), nil
		}
		if i+len(segBangBrace) <= len(literal) && literal[i:i+len(segBangBrace)] == segBangBrace {
			return segBangBrace, i + len(segBangBrace), nil
		}
		return segBang, i + 1, nil
	default:
		return "", i, fmt.Errorf("readCompoundAffixSegment: expected @ or !, got %q", literal[i:])
	}
}

// splitCompoundSegments splits a compound identifier into List cells (names, punctuation, affixes).
func splitCompoundSegments(literal string) ([]string, error) {
	if literal == "" {
		return nil, fmt.Errorf("empty compound")
	}
	var segs []string
	i := 0
	for i < len(literal) {
		switch {
		case literal[i] == '.':
			segs = append(segs, segDot)
			i++
		case i+1 < len(literal) && literal[i:i+2] == segIndex:
			segs = append(segs, segIndex)
			i += 2
		case i+len(segQueryBrace) <= len(literal) && literal[i:i+len(segQueryBrace)] == segQueryBrace:
			segs = append(segs, segQueryBrace)
			i += len(segQueryBrace)
		case literal[i] == '@' || literal[i] == '!':
			affix, next, err := readCompoundAffixSegment(literal, i)
			if err != nil {
				return nil, err
			}
			segs = append(segs, affix)
			i = next
		case literal[i] == '~':
			segs = append(segs, segTilde)
			i++
		case literal[i] == '^':
			segs = append(segs, segCaret)
			i++
		case literal[i] == '>':
			arrow, next, err := readCompoundStackArrowSegment(literal, i)
			if err != nil {
				return nil, err
			}
			segs = append(segs, arrow)
			i = next
		default:
			j := i
			for j < len(literal) {
				if literal[j] == '.' {
					break
				}
				if j+1 < len(literal) && literal[j:j+2] == segIndex {
					break
				}
				if j+len(segQueryBrace) <= len(literal) && literal[j:j+len(segQueryBrace)] == segQueryBrace {
					break
				}
				if literal[j] == '@' || literal[j] == '!' || literal[j] == '>' || literal[j] == '~' || literal[j] == '^' {
					break
				}
				j++
			}
			if j == i {
				return nil, fmt.Errorf("invalid compound character at %q", literal[i:])
			}
			atom := literal[i:j]
			if err := validateCompoundNameSegment(atom); err != nil {
				return nil, err
			}
			segs = append(segs, atom)
			i = j
		}
	}
	if len(segs) > 0 && segs[len(segs)-1] == segDot {
		return nil, fmt.Errorf("compound may not end with %q", segDot)
	}
	return segs, nil
}

func validateCompoundNameSegment(atom string) error {
	if atom == "" {
		return fmt.Errorf("empty name segment")
	}
	if strings.ContainsRune(atom, ':') {
		parts := strings.Split(atom, "::")
		for _, p := range parts {
			if p == "" {
				return fmt.Errorf("invalid :: segment in %q", atom)
			}
			if err := validateIdentSpelling(p, false); err != nil {
				return fmt.Errorf("invalid name in :: segment %q", atom)
			}
		}
		return nil
	}
	if err := validateIdentSpelling(atom, false); err != nil {
		return fmt.Errorf("invalid compound name segment %q: %w", atom, err)
	}
	return nil
}

func hasCompoundPathPunct(literal string) bool {
	if strings.Contains(literal, ".") {
		return true
	}
	if strings.Contains(literal, segIndex) {
		return true
	}
	if strings.Contains(literal, segQueryBrace) {
		return true
	}
	for _, r := range literal {
		switch r {
		case '@', '!', '~', '^':
			return true
		}
	}
	return false
}

// CompoundPathForTest builds a compound path value for tests (no stack prefix).
func CompoundPathForTest(segs []string) *compoundPathValue {
	return newCompoundPath(segs, compoundPathNoSelector)
}

// CompoundPathWithSelectorForTest builds a t:-style compound path for tests.
func CompoundPathWithSelectorForTest(segs []string, selector ContextStackIndex) *compoundPathValue {
	return newCompoundPath(segs, selector)
}

func classifyNameLiteral(literal string) (TokenType, any, error) {
	switch literal {
	case "true":
		return LiteralTokenType(TokenBool), Bool(true), nil
	case "false":
		return LiteralTokenType(TokenBool), Bool(false), nil
	case "null":
		return LiteralTokenType(TokenNull), Any(nil), nil
	}

	if literal == segIndex {
		return TokenIdent, String(literal), nil
	}
	if literal == segAt || literal == segBang || literal == segTilde || literal == segCaret ||
		literal == segAtIndex || literal == segBangIndex || literal == segAtBrace || literal == segBangBrace || literal == segQueryBrace ||
		literal == ">ts" || literal == "ts>" {
		return TokenIdent, String(literal), nil
	}

	if rest, selIdx, peeled, err := peelEnabledCompoundStackSelector(literal); err != nil {
		return 0, nil, err
	} else if peeled {
		if rest == "" {
			return TokenCompoundIdent, newCompoundPath(nil, selIdx), nil
		}
		typ, val, err := classifyCompoundLiteral(rest, selIdx)
		if err == nil {
			return typ, val, nil
		}
		if hasCompoundPathPunct(rest) || strings.HasPrefix(rest, segDot) {
			return 0, nil, err
		}
		return TokenCompoundIdent, newCompoundPath([]string{rest}, selIdx), nil
	}

	if hasCompoundPathPunct(literal) || strings.HasPrefix(literal, segDot) {
		return classifyCompoundLiteral(literal, compoundPathNoSelector)
	}

	if strings.Contains(literal, "::") {
		if err := validateCompoundNameSegment(literal); err != nil {
			return 0, nil, err
		}
	}

	if err := validateIdentSpelling(literal, true); err != nil {
		return 0, nil, err
	}
	return TokenIdent, String(literal), nil
}
