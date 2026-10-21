package core

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// Rune stream access for Parser: pull-only TextRead with a small read-ahead buffer.

func (p *Parser) atStreamEOF() bool {
	return p.streamEOF && len(p.buf) == 0
}

func (p *Parser) ensureBuffered(n int) {
	for len(p.buf) < n && !p.streamEOF {
		if p.src == nil {
			p.streamEOF = true
			return
		}
		r, eof := p.src.GetNextRune()
		if eof {
			p.streamEOF = true
			return
		}
		p.buf = append(p.buf, r)
	}
}

func (p *Parser) hasRune() bool {
	p.ensureBuffered(1)
	return len(p.buf) > 0
}

func (p *Parser) peekRune() rune {
	p.ensureBuffered(1)
	if len(p.buf) == 0 {
		return 0
	}
	return p.buf[0]
}

func (p *Parser) peekNextRune() rune {
	p.ensureBuffered(2)
	if len(p.buf) < 2 {
		return 0
	}
	return p.buf[1]
}

// liveCol is the 1-based column of the next rune to read on the current line.
func (p *Parser) liveCol() int {
	return utf8.RuneCountInString(p.lineText) + 1
}

func (p *Parser) advanceRune() rune {
	p.ensureBuffered(1)
	if len(p.buf) == 0 {
		return 0
	}
	ch := p.buf[0]
	p.buf = p.buf[1:]
	switch ch {
	case '\r':
		p.line++
		p.lineText = ""
		p.ensureBuffered(1)
		if len(p.buf) > 0 && p.buf[0] == '\n' {
			p.buf = p.buf[1:]
		}
	case '\n':
		p.line++
		p.lineText = ""
	default:
		p.lineText += string(ch)
	}
	return ch
}

func (p *Parser) takeSourceRune() (rune, error) {
	if !p.hasRune() {
		return 0, fmt.Errorf("unexpected end of input")
	}
	ch := p.peekRune()
	if IsDangerousRune(ch) {
		return 0, p.faultAt(p.line, p.liveCol(), errDangerousSourceRune(ch))
	}
	return p.advanceRune(), nil
}

func (p *Parser) skipHorizontalWhitespace() error {
	for p.hasRune() {
		ch := p.peekRune()
		if ch == '\n' || ch == '\r' {
			return nil
		}
		if ch == ' ' || ch == '\t' || (unicode.IsSpace(ch) && ch != '\n' && ch != '\r') {
			if _, err := p.takeSourceRune(); err != nil {
				return err
			}
			continue
		}
		return nil
	}
	return nil
}

func (p *Parser) drainToEndOfLine() {
	for {
		tok, err := p.nextToken()
		if err != nil {
			continue
		}
		if tok.Type == TokenNewline || tok.Type == TokenEOF {
			return
		}
	}
}

// drainToStreamEOF abandons unread input without updating line/col/lineText
// (those remain the fault-report site).
func (p *Parser) drainToStreamEOF() {
	for {
		p.ensureBuffered(1)
		if len(p.buf) == 0 {
			break
		}
		p.buf = p.buf[1:]
	}
	p.buf = p.buf[:0]
	p.streamEOF = true
}

func (p *Parser) bindString(s string) {
	tr, err := NewTextReaderFromString(s)
	if err != nil {
		panic(err)
	}
	p.src = tr
	p.buf = p.buf[:0]
	p.streamEOF = false
	p.file = ""
	p.line = 1
	p.col = 0
	p.lineText = ""
	p.inRaw = false
	p.inBlock = false
	p.rawBuf.Reset()
	p.blockBuf.Reset()
	p.pendingRawToken = nil
	p.errors = p.errors[:0]
	p.lastComment = ""
}
