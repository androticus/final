// parse.go - Enhanced Safety + UTF-8 + Normalization

package core

import (
	"fmt"
	"strings"
	"unicode"
)

// Token is a batch-parse snapshot (tests and debugging).
type Token struct {
	Type    TokenType
	Literal string
	Value   any
	Line    int
	Col     int
}

type LinePos struct {
	Line int
	Col  int
}

var (
	nullParser = NewParserFromTextRead(GetNullTextReader())
)

func GetNullParser() *Parser {
	return nullParser
}

// Parser produces a uniform token stream via Next for interpret and compile loops.
// Spec: docs/001-high-level/02-lexical-parsing.md
//
// Input is rune-only (TextRead.GetNextRune) with a small parser read-ahead buffer for peek.
// TokenEOF means end of the bound stream, not end of line.
type Parser struct {
	src TextRead

	buf       []rune
	streamEOF bool

	file string
	line int // 1-based line number (report site + current line)
	col  int // 1-based caret column for the last fault site (0 = none yet)

	// lineText is the text of the line currently being read (fault-report cache).
	// The live read column is derived as utf8.RuneCountInString(lineText)+1.
	lineText string

	inRaw    bool
	rawStart LinePos
	rawBuf   strings.Builder

	inBlock    bool
	blockStart LinePos
	blockBuf   strings.Builder

	pendingRawToken *Token
	errors          []string

	lastComment string
}

// NewParser returns a parser with no stream source until bindString or ParseLine.
func NewParser(_ any) *Parser {
	return &Parser{streamEOF: true, line: 1, col: 0}
}

// NewParserFromTextRead binds a text stream for incremental Next until stream EOF.
func NewParserFromTextRead(r TextRead) *Parser {
	if r == nil {
		panic("NewParserFromTextRead: nil TextRead")
	}
	return &Parser{src: r, line: 1, col: 0}
}

func (p *Parser) InputEOF() bool {
	return p.atStreamEOF()
}

// LastComment returns the elided text of the most recently lexed comment token.
func (p *Parser) LastComment() string {
	return p.lastComment
}

// GetClearLastComment returns LastComment and clears it to "".
func (p *Parser) GetClearLastComment() string {
	s := p.lastComment
	p.lastComment = ""
	return s
}

func (p *Parser) noteCommentValue(body string) {
	p.lastComment = body
}

func (p *Parser) Next() (value any, typ TokenType, err error) {
	tok, err := p.nextToken()
	if err != nil {
		return nil, TokenEOF, err
	}
	p.noteYieldedToken(tok)
	return tok.Value, tok.Type, nil
}

// noteYieldedToken records the caret column for tokens that can be fault sites.
// It must not assign tok.Line onto p.line: multi-line tokens (raw strings) store
// Line at their opener, while the live line counter has already advanced through
// the body. Rewinding would under-count later faults (e.g. compile-fault.fnl).
func (p *Parser) noteYieldedToken(tok Token) {
	if tok.Type == TokenEOF || tok.Type == TokenNewline || HasTokenFlag(tok.Type, TokenComment) {
		return
	}
	p.col = tok.Col
}

func (p *Parser) PushNextAsIdentifier(ds *Stack[any]) {
	for {
		val, typ, err := p.Next()
		if err != nil {
			panic(fmt.Errorf("PushNextAsIdentifier: %w", err))
		}
		if typ == TokenEOF {
			panic(fmt.Errorf("PushNextAsIdentifier: unexpected end of input"))
		}
		if typ == TokenNewline || HasTokenFlag(typ, TokenComment) {
			continue
		}
		if !IsPlainIdentToken(typ) {
			lit := tokenLiteralString(val)
			switch {
			case HasTokenFlag(typ, TokenCompoundIdent):
				panic(fmt.Errorf("PushNextAsIdentifier: expected simple identifier, got compound identifier %q", lit))
			default:
				panic(fmt.Errorf("PushNextAsIdentifier: expected simple identifier, got token type %d (%q)", typ, lit))
			}
		}
		name, ok := val.(String)
		if !ok {
			panic(fmt.Errorf("PushNextAsIdentifier: identifier token has type %T", val))
		}
		ds.Push(name)
		return
	}
}

func tokenLiteralString(val any) string {
	if val == nil {
		return ""
	}
	if s, ok := val.(String); ok {
		return string(s)
	}
	if l, ok := val.(*AnyList); ok {
		return fmt.Sprintf("List(len=%d)", l.Len())
	}
	return fmt.Sprint(val)
}

func pushNextToken(p *Parser, ds *Stack[any]) error {
	val, typ, err := p.Next()
	if err != nil {
		return err
	}
	if typ == TokenEOF {
		ds.Push(Int(TokenEOF))
		return nil
	}
	if typ == TokenNewline {
		ds.Push(Any(nil))
		ds.Push(Int(TokenNewline))
		return nil
	}
	ds.Push(val)
	ds.Push(Int(typ))
	return nil
}

func (p *Parser) PushNext(ds *Stack[any]) error {
	return pushNextToken(p, ds)
}

func StreamEOFOnStack(ds *Stack[any]) bool {
	if ds.Len() == 0 {
		return false
	}
	v, ok := ds.Peek().(Int)
	return ok && TokenType(v) == TokenEOF
}

func PopStreamEOF(ds *Stack[any]) bool {
	if !StreamEOFOnStack(ds) {
		return false
	}
	ds.Pop()
	return true
}

func (p *Parser) ParseLine(line string) ([]Token, bool, error) {
	p.bindString(line)
	var tokens []Token
	for {
		tok, err := p.nextToken()
		if err != nil {
			return tokens, p.inRaw || p.inBlock, err
		}
		if tok.Type == TokenEOF {
			break
		}
		tokens = append(tokens, tok)
	}
	if len(p.errors) > 0 {
		return tokens, p.inRaw || p.inBlock, fmt.Errorf("%s", p.errors[0])
	}
	return tokens, p.inRaw || p.inBlock, nil
}

func (p *Parser) takePendingRawToken() (Token, bool) {
	if p.pendingRawToken == nil {
		return Token{}, false
	}
	tok := *p.pendingRawToken
	p.pendingRawToken = nil
	return tok, true
}

func (p *Parser) nextToken() (Token, error) {
	if tok, ok := p.takePendingRawToken(); ok {
		return tok, nil
	}

	for {
		if !p.hasRune() {
			// V1: report raise site (often EOF), not the opener line.
			if p.inRaw {
				return Token{}, p.faultAt(p.line, p.liveCol(), fmt.Errorf("unclosed raw string"))
			}
			if p.inBlock {
				return Token{}, p.faultAt(p.line, p.liveCol(), fmt.Errorf("unclosed block comment"))
			}
			return Token{Type: TokenEOF}, nil
		}

		if p.inBlock {
			if err := p.continueBlockComment(); err != nil {
				return Token{}, err
			}
			if !p.inBlock {
				return p.finishBlockCommentToken(), nil
			}
			if !p.hasRune() {
				continue
			}
		}
		if p.inRaw {
			if err := p.continueRawString(); err != nil {
				return Token{}, err
			}
			if tok, ok := p.takePendingRawToken(); ok {
				return tok, nil
			}
			continue
		}

		ch := p.peekRune()
		if IsDangerousRune(ch) {
			return Token{}, p.faultAt(p.line, p.liveCol(), errDangerousSourceRune(ch))
		}

		// TokenNewline only outside inRaw/inBlock — those modes absorb \n into content above.
		if ch == '\r' || ch == '\n' {
			startLine, startCol := p.line, p.liveCol()
			if _, err := p.takeSourceRune(); err != nil {
				return Token{}, err
			}
			return Token{Type: TokenNewline, Line: startLine, Col: startCol}, nil
		}

		if unicode.IsSpace(ch) {
			if err := p.skipHorizontalWhitespace(); err != nil {
				return Token{}, err
			}
			continue
		}

		startLine, startCol := p.line, p.liveCol()

		switch {
		case ch == '#' && startCol == 1:
			// Undocumented shebang form: '#' in column 1 is a whole-line comment.
			// Elided like // but must not enter lastComment / Word.doc.
			return p.parseHashColumn1CommentToken()
		case ch == '/' && p.peekNextRune() == '/':
			return p.parseLineCommentToken()
		case ch == '/' && p.peekNextRune() == '*':
			if err := p.startBlockComment(); err != nil {
				return Token{}, err
			}
			if !p.inBlock {
				return p.finishBlockCommentToken(), nil
			}
			continue
		case ch == '*' && p.peekNextRune() == '/' && !p.inBlock:
			return Token{}, p.faultAt(p.line, p.liveCol(), fmt.Errorf("stray */"))
		case ch == '"':
			return p.parseQuotedStringToken()
		case ch == '`':
			if err := p.startRawString(); err != nil {
				return Token{}, err
			}
			if tok, ok := p.takePendingRawToken(); ok {
				return tok, nil
			}
			continue
		case ch == '\'':
			return p.parseRuneToken()
		default:
			word, err := p.readWordUntilWhitespace()
			if err != nil {
				return Token{}, err
			}
			if word == "" {
				return Token{}, p.faultAt(startLine, startCol, fmt.Errorf("unexpected character %q", string(ch)))
			}
			return p.classifyWordToken(word, startLine, startCol)
		}
	}
}

func (p *Parser) readWordUntilWhitespace() (string, error) {
	var sb strings.Builder
	for p.hasRune() && !unicode.IsSpace(p.peekRune()) {
		ch, err := p.takeSourceRune()
		if err != nil {
			return "", err
		}
		sb.WriteRune(ch)
	}
	return sb.String(), nil
}

func (p *Parser) classifyWordToken(word string, line, col int) (Token, error) {
	if looksLikeNumberStart(word) {
		typ, val, err := parseNumberLiteral(word)
		if err != nil {
			p.addError(err)
			return Token{}, err
		}
		return Token{Type: typ, Value: val, Literal: word, Line: line, Col: col}, nil
	}
	typ, val, err := classifyNameLiteral(word)
	if err != nil {
		p.addError(err)
		return Token{}, err
	}
	return Token{Type: typ, Value: val, Literal: word, Line: line, Col: col}, nil
}

func (p *Parser) parseHashColumn1CommentToken() (Token, error) {
	startLine, startCol := p.line, p.liveCol()
	if _, err := p.takeSourceRune(); err != nil {
		return Token{}, err
	}
	var literal strings.Builder
	literal.WriteByte('#')
	for p.hasRune() && p.peekRune() != '\n' && unicode.IsSpace(p.peekRune()) {
		ch, err := p.takeSourceRune()
		if err != nil {
			return Token{}, err
		}
		literal.WriteRune(ch)
	}
	var body strings.Builder
	for p.hasRune() && p.peekRune() != '\n' {
		ch, err := p.takeSourceRune()
		if err != nil {
			return Token{}, err
		}
		literal.WriteRune(ch)
		body.WriteRune(ch)
	}
	val := body.String()
	return Token{
		Type:    TokenComment,
		Value:   String(val),
		Literal: literal.String(),
		Line:    startLine,
		Col:     startCol,
	}, nil
}

func (p *Parser) parseLineCommentToken() (Token, error) {
	startLine, startCol := p.line, p.liveCol()
	if _, err := p.takeSourceRune(); err != nil {
		return Token{}, err
	}
	if _, err := p.takeSourceRune(); err != nil {
		return Token{}, err
	}
	var literal strings.Builder
	literal.WriteString("//")
	for p.hasRune() && p.peekRune() != '\n' && unicode.IsSpace(p.peekRune()) {
		ch, err := p.takeSourceRune()
		if err != nil {
			return Token{}, err
		}
		literal.WriteRune(ch)
	}
	var body strings.Builder
	for p.hasRune() && p.peekRune() != '\n' {
		ch, err := p.takeSourceRune()
		if err != nil {
			return Token{}, err
		}
		literal.WriteRune(ch)
		body.WriteRune(ch)
	}
	lit := literal.String()
	val := body.String()
	p.noteCommentValue(val)
	return Token{
		Type:    TokenComment,
		Value:   String(val),
		Literal: lit,
		Line:    startLine,
		Col:     startCol,
	}, nil
}

func (p *Parser) startBlockComment() error {
	startLine, startCol := p.line, p.liveCol()
	if _, err := p.takeSourceRune(); err != nil {
		return err
	}
	if _, err := p.takeSourceRune(); err != nil {
		return err
	}
	p.inBlock = true
	p.blockStart = LinePos{startLine, startCol}
	p.blockBuf.Reset()
	p.blockBuf.WriteString("/*")
	return p.continueBlockComment()
}

func (p *Parser) continueBlockComment() error {
	for p.hasRune() {
		if p.peekRune() == '/' && p.peekNextRune() == '/' {
			p.blockBuf.WriteString("//")
			if _, err := p.takeSourceRune(); err != nil {
				return err
			}
			if _, err := p.takeSourceRune(); err != nil {
				return err
			}
			for p.hasRune() && p.peekRune() != '\n' {
				ch, err := p.takeSourceRune()
				if err != nil {
					return err
				}
				p.blockBuf.WriteRune(ch)
			}
			if p.hasRune() && p.peekRune() == '\n' {
				ch, err := p.takeSourceRune()
				if err != nil {
					return err
				}
				p.blockBuf.WriteRune(ch)
			}
			continue
		}
		if p.peekRune() == '*' && p.peekNextRune() == '/' {
			p.blockBuf.WriteString("*/")
			if _, err := p.takeSourceRune(); err != nil {
				return err
			}
			if _, err := p.takeSourceRune(); err != nil {
				return err
			}
			p.inBlock = false
			return nil
		}
		ch, err := p.takeSourceRune()
		if err != nil {
			return err
		}
		p.blockBuf.WriteRune(ch)
	}
	return nil
}

func (p *Parser) finishBlockCommentToken() Token {
	lit := p.blockBuf.String()
	val := elideBlockCommentBody(lit)
	p.noteCommentValue(val)
	return Token{
		Type:    TokenComment,
		Value:   String(val),
		Literal: lit,
		Line:    p.blockStart.Line,
		Col:     p.blockStart.Col,
	}
}

func elideBlockCommentBody(raw string) string {
	if !strings.HasPrefix(raw, "/*") {
		return strings.TrimSpace(raw)
	}
	inner := raw[2:]
	if i := strings.LastIndex(inner, "*/"); i >= 0 {
		inner = inner[:i]
	}
	return strings.TrimSpace(inner)
}

func (p *Parser) startRawString() error {
	if _, err := p.takeSourceRune(); err != nil {
		return err
	}
	p.inRaw = true
	p.rawStart = LinePos{p.line, p.liveCol()}
	p.rawBuf.Reset()
	return p.continueRawString()
}

func (p *Parser) continueRawString() error {
	for p.hasRune() {
		ch, err := p.takeSourceRune()
		if err != nil {
			return err
		}
		if ch == '`' {
			p.inRaw = false
			tok := Token{
				Type:  LiteralTokenType(TokenRawString),
				Value: String(p.rawBuf.String()),
				Line:  p.rawStart.Line,
				Col:   p.rawStart.Col,
			}
			p.pendingRawToken = &tok
			return nil
		}
		if ch == '\r' || ch == '\n' {
			p.rawBuf.WriteRune('\n')
		} else {
			p.rawBuf.WriteRune(ch)
		}
	}
	p.rawBuf.WriteRune('\n')
	return nil
}

func (p *Parser) parseQuotedStringToken() (Token, error) {
	startLine, startCol := p.line, p.liveCol()
	if _, err := p.takeSourceRune(); err != nil {
		return Token{}, err
	}

	var sb strings.Builder
	for p.hasRune() {
		ch, err := p.takeSourceRune()
		if err != nil {
			return Token{}, err
		}
		if ch == '"' {
			return Token{
				Type:  LiteralTokenType(TokenString),
				Value: String(sb.String()),
				Line:  startLine,
				Col:   startCol,
			}, nil
		}
		if ch == '\\' {
			r, b, usedByte, err := p.decodeGoEscape()
			if err != nil {
				return Token{}, p.faultAt(startLine, startCol, err)
			}
			if usedByte {
				sb.WriteByte(b)
			} else {
				sb.WriteRune(r)
			}
			continue
		}
		sb.WriteRune(ch)
	}
	return Token{}, p.faultAt(p.line, p.liveCol(), fmt.Errorf("unclosed string"))
}

func (p *Parser) setParserLocation(file string, line, col int) {
	p.file = file
	p.line = line
	p.col = 0 // no fault site until a token is yielded or faultAt
	p.lineText = ""
	_ = col
}

func (p *Parser) clearParserLocation() {
	p.file = ""
	p.line = 1
	p.col = 0
	p.lineText = ""
}

func (p *Parser) hasParserLocation() bool {
	return p.col > 0
}

func (p *Parser) faultAt(line, col int, err error) error {
	p.line = line
	if p.lineText == "" {
		p.col = 1
	} else if col > 0 {
		p.col = col
	} else {
		p.col = 1
	}
	return err
}

func (p *Parser) addError(err error) {
	p.errors = append(p.errors, formatFault(err, p))
}

func (p *Parser) errorf(format string, args ...any) {
	p.addError(fmt.Errorf(format, args...))
}

func peekStackTokenType(ds *Stack[any]) (TokenType, bool) {
	if ds.Len() == 0 {
		return 0, false
	}
	v, ok := ds.Peek().(Int)
	if !ok {
		return 0, false
	}
	return TokenType(v), true
}
