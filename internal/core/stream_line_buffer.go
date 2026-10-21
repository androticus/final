package core

// LineBufferTextWriter is a TextWriter hybrid over a StringLineBuffer, for easy
// access to ToString (last N complete lines).
type LineBufferTextWriter struct {
	TextWriter
	slb StringLineBufferer
}

func NewLineBufferTextWriter(maxLines int) *LineBufferTextWriter {
	slb := NewStringLineBuffer(maxLines)
	return &LineBufferTextWriter{
		TextWriter: TextWriter{p: &lineBufferWriterProvider{slb: slb}},
		slb:        slb,
	}
}

func NewLineBufferTextWriterSync(maxLines int) *LineBufferTextWriter {
	slb := NewStringLineBufferSync(maxLines)
	return &LineBufferTextWriter{
		TextWriter: TextWriter{p: &lineBufferWriterProvider{slb: slb}},
		slb:        slb,
	}
}

func (s *LineBufferTextWriter) ToString() string {
	return s.slb.ToString()
}

func (s *LineBufferTextWriter) GetStringLineBuffer() StringLineBufferer {
	return s.slb
}

func (s *LineBufferTextWriter) FinalType() *Type {
	return lineBufferTextWriterType
}

type lineBufferWriterProvider struct {
	slb StringLineBufferer
}

func (p *lineBufferWriterProvider) WriteRune(r rune) {
	p.slb.Append(string(r))
}

func (p *lineBufferWriterProvider) WriteLine(line string) {
	p.slb.AppendLine(line)
}

func (p *lineBufferWriterProvider) WriteString(s string) {
	p.slb.Append(s)
}

func registerLineBufferTextWriterMembers(t *Type) {
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		n := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewLineBufferTextWriter(n))
	})
	t.registerNativeClassMethod("newSync", func(f *RunFrame, _ *TypeObjPair) {
		n := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewLineBufferTextWriterSync(n))
	})
	t.registerNativeMethod("toString", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(String(ths.Obj.(*LineBufferTextWriter).ToString()))
	})
}

func CreateRegisterLineBufferTextWriterType() {
	t := &Type{
		name:            "LineBufferTextWriter",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[LineBufferTextWriter]("LineBufferTextWriter"),
	}
	t.SetParent(textWriteType)
	registerLineBufferTextWriterMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	lineBufferTextWriterType = t
}

var (
	lineBufferTextWriterType *Type
	_                        TextWrite         = (*LineBufferTextWriter)(nil)
	_                        TextWriteProvider = (*lineBufferWriterProvider)(nil)
	_                        Typed             = (*LineBufferTextWriter)(nil)
)
