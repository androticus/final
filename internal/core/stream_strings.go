package core

type stringReaderProvider struct {
	s   string
	off int
}

func (p *stringReaderProvider) GetNextRune() (rune, bool) {
	return readNextRuneFromString(p.s, &p.off, "string stream")
}

// NewTextReaderFromString returns a read-only text stream over s.
func NewTextReaderFromString(s string) (*TextReader, error) {
	return &TextReader{p: &stringReaderProvider{s: s}}, nil
}

var (
	_ TextReadProvider = (*stringReaderProvider)(nil)
)
