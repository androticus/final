package core

type nullTextProvider struct{}

func (nullTextProvider) GetNextRune() (r rune, eof bool) { return 0, true }

func (nullTextProvider) WriteRune(r rune) {}

func (nullTextProvider) WriteLine(line string) {}

func (nullTextProvider) WriteString(s string) {}

var (
	nullTextProviderInstance = nullTextProvider{}
	nullTextReader           = &TextReader{p: nullTextProviderInstance}
	nullTextWriter           = &TextWriter{p: nullTextProviderInstance}
	nullTextReaderWriter     = &TextReaderWriter{p: nullTextProviderInstance}
)

// GetNullTextReader returns the shared null stdin (reads always EOF).
func GetNullTextReader() *TextReader { return nullTextReader }

// GetNullTextWriter returns the shared null stdout (writes discarded).
func GetNullTextWriter() *TextWriter { return nullTextWriter }

// GetNullTextReaderWriter returns the shared null read/write stream.
func GetNullTextReaderWriter() *TextReaderWriter { return nullTextReaderWriter }

var _ TextReadWriteProvider = (*nullTextProvider)(nil)
