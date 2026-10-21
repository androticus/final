package core

import (
	"fmt"
	"strings"
)

// StringBuilder accumulates UTF-8 text for stream writers and ToString retrieval.
type StringBuilder struct {
	b            strings.Builder
	pretty       bool   // if true, add newlines and indentation for pretty printing
	indentString string // the string to use for indentation
	nestingLevel int    // the current nesting level, 0==outermost (no indent)
}

func NewStringBuilder() *StringBuilder {
	return &StringBuilder{}
}

func NewStringBuilderWithPretty(indent string) *StringBuilder {
	return &StringBuilder{
		pretty:       true,
		indentString: indent,
	}
}

func (sb *StringBuilder) FinalType() *Type {
	return stringBuilderType
}

func (sb *StringBuilder) Indent() {
	if !sb.pretty || sb.nestingLevel == 0 {
		return
	}
	sb.b.WriteString(strings.Repeat(sb.indentString, sb.nestingLevel))
}

func (sb *StringBuilder) WriteRune(r rune) {
	if _, err := sb.b.WriteRune(r); err != nil {
		panic(fmt.Errorf("StringBuilder.WriteRune: %w", err))
	}
}

func (sb *StringBuilder) WriteLine(line string) {
	sb.WriteString(line)
	sb.WriteRune('\n')
}

func (sb *StringBuilder) WriteString(s string) {
	sb.b.WriteString(s)
}

func (sb *StringBuilder) WriteStringPretty(s string) {
	sb.Indent()
	sb.b.WriteString(s)
}

func (sb *StringBuilder) WriteSpacePretty() {
	if sb.pretty {
		sb.b.WriteRune(' ')
	}
}

func (sb *StringBuilder) WriteWhitespacePretty() {
	if sb.pretty {
		sb.b.WriteRune('\n')
	} else {
		sb.b.WriteRune(' ')
	}
}

func (sb *StringBuilder) Nest() {
	/*if sb.pretty {
		sb.b.WriteRune('\n')
	}*/
	sb.nestingLevel++
}

func (sb *StringBuilder) Unnest() {
	if sb.nestingLevel > 0 {
		sb.nestingLevel--
	}
}

func (sb *StringBuilder) ToString() string {
	return sb.b.String()
}

func registerStringBuilderMembers(t *Type) {
	t.registerNativeMethod("writeRune", func(f *RunFrame, ths *TypeObjPair) {
		r := rune(f.RP.DataStack.Pop().(Int))
		ths.Obj.(*StringBuilder).WriteRune(r)
	})
	t.registerNativeMethod("writeLine", func(f *RunFrame, ths *TypeObjPair) {
		s := string(f.RP.DataStack.Pop().(String))
		ths.Obj.(*StringBuilder).WriteLine(s)
	})
	t.registerNativeMethod("writeString", func(f *RunFrame, ths *TypeObjPair) {
		s := string(f.RP.DataStack.Pop().(String))
		ths.Obj.(*StringBuilder).WriteString(s)
	})
	t.registerNativeMethod("writeStringPretty", func(f *RunFrame, ths *TypeObjPair) {
		s := string(f.RP.DataStack.Pop().(String))
		ths.Obj.(*StringBuilder).WriteStringPretty(s)
	})
	t.registerNativeMethod("writeSpacePretty", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*StringBuilder).WriteSpacePretty()
	})
	t.registerNativeMethod("writeWhitespacePretty", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*StringBuilder).WriteWhitespacePretty()
	})
	t.registerNativeMethod("nest", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*StringBuilder).Nest()
	})
	t.registerNativeMethod("unnest", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*StringBuilder).Unnest()
	})
	t.registerNativeMethod("toString", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(String(ths.Obj.(*StringBuilder).ToString()))
	})
	t.registerProperty("pretty", "Bool",
		func(inst Any) Any { return Bool(inst.(*StringBuilder).pretty) },
		nil,
	)
	t.registerProperty("indentString", "String",
		func(inst Any) Any { return String(inst.(*StringBuilder).indentString) },
		nil,
	)
}

func CreateRegisterStringBuilderType() {
	t := &Type{
		name:            "StringBuilder",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[StringBuilder]("StringBuilder"),
	}
	t.SetParent(streamReadWriteType)
	registerStringBuilderMembers(t)
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewStringBuilder())
	})
	t.registerNativeClassMethod("newWithPretty", func(f *RunFrame, _ *TypeObjPair) {
		indent := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(NewStringBuilderWithPretty(indent))
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	stringBuilderType = t
}

var stringBuilderType *Type

var _ TextWrite = (*StringBuilder)(nil)

// Registration is driven from stream_types.go init (after StreamReadWrite)
// so SetParent can use streamReadWriteType by instance.
