package core

import "fmt"

type TextReaderIter struct {
	r    *TextReader
	curr rune
}

func (i *TextReaderIter) GetNext() bool {
	curr, eof := i.r.GetNextRune()
	if eof {
		return false
	}
	i.curr = curr
	return true
}

func (i *TextReaderIter) Current() any {
	return int(i.curr)
}

func (r *TextReader) GetIter() Iterator {
	return &TextReaderIter{r: r}
}

func (r *TextReader) FinalType() *Type {
	return textReaderType
}

func (w *TextWriter) FinalType() *Type {
	return textWriterType
}

func (r *ByteReader) FinalType() *Type {
	return byteReaderType
}

func (w *ByteWriter) FinalType() *Type {
	return byteWriterType
}

func pushStreamOpenResult(f *RunFrame, val any, err error) {
	if err != nil {
		f.RP.DataStack.Push(String(err.Error()))
		f.RP.DataStack.Push(Bool(false))
		return
	}
	f.RP.DataStack.Push(val)
	f.RP.DataStack.Push(Bool(true))
}

// pushStreamWriteResult is the write-side modal shape: ( -- true | err:s false ).
func pushStreamWriteResult(f *RunFrame, err error) {
	if err != nil {
		f.RP.DataStack.Push(String(err.Error()))
		f.RP.DataStack.Push(Bool(false))
		return
	}
	f.RP.DataStack.Push(Bool(true))
}

func mustStreamOpen[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

func popByteOperand(ds *Stack[any]) byte {
	return byte(popIntegralInt(ds, "writeNextByte"))
}

func registerStreamReadWriteMembers(t *Type) {
	t.registerNativeMethod("close", func(f *RunFrame, ths *TypeObjPair) {
		if c, ok := ths.Obj.(Closeable); ok {
			c.Close()
		}
	})
	t.registerNativeMethod("flush", func(f *RunFrame, ths *TypeObjPair) {
		if fl, ok := ths.Obj.(Flushable); ok {
			fl.Flush()
		}
	})
}

// CreateRegisterStreamReadWriteType registers the shared Final parent for close/flush.
// Handlers probe Go Closeable / Flushable; absent capabilities are no-ops.
func CreateRegisterStreamReadWriteType() {
	t := &Type{
		name:    "StreamReadWrite",
		getZero: ReferenceTypeZeroFn,
		checkCoerceType: func(value any) any {
			if value == nil {
				return nil
			}
			switch value.(type) {
			case TextRead, TextWrite, ByteRead, ByteWrite:
				return value
			default:
				panic(fmt.Errorf(
					"StreamReadWrite.checkCoerceType: expected TextRead, TextWrite, ByteRead, or ByteWrite, got %T",
					value,
				))
			}
		},
	}
	registerStreamReadWriteMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	streamReadWriteType = t
}

func registerTextWriteMembers(t *Type) {
	t.registerNativeMethod("writeRune", func(f *RunFrame, ths *TypeObjPair) {
		r := rune(f.RP.DataStack.Pop().(Int))
		ths.Obj.(TextWrite).WriteRune(r)
	})
	t.registerNativeMethod("writeLine", func(f *RunFrame, ths *TypeObjPair) {
		line := string(f.RP.DataStack.Pop().(String))
		ths.Obj.(TextWrite).WriteLine(line)
	})
	t.registerNativeMethod("writeString", func(f *RunFrame, ths *TypeObjPair) {
		s := string(f.RP.DataStack.Pop().(String))
		ths.Obj.(TextWrite).WriteString(s)
	})
}

func CreateRegisterTextWriteType() {
	t := &Type{
		name:    "TextWrite",
		getZero: ReferenceTypeZeroFn,
		checkCoerceType: func(value any) any {
			if value == nil {
				return nil
			}
			if _, ok := value.(TextWrite); !ok {
				panic(fmt.Errorf("TextWrite.checkCoerceType: expected TextWrite, got %T", value))
			}
			return value
		},
	}
	t.SetParent(streamReadWriteType)
	registerTextWriteMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	textWriteType = t
}

func registerByteWriteMembers(t *Type) {
	t.registerNativeMethod("writeNextByte", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(ByteWrite).WriteNextByte(popByteOperand(f.RP.DataStack))
	})
	t.registerNativeMethod("writeBytes", func(f *RunFrame, ths *TypeObjPair) {
		list := f.RP.DataStack.Pop().(*ByteList)
		ths.Obj.(ByteWrite).WriteBytes(list.items)
	})
}

func CreateRegisterByteWriteType() {
	t := &Type{
		name:    "ByteWrite",
		getZero: ReferenceTypeZeroFn,
		checkCoerceType: func(value any) any {
			if value == nil {
				return nil
			}
			if _, ok := value.(ByteWrite); !ok {
				panic(fmt.Errorf("ByteWrite.checkCoerceType: expected ByteWrite, got %T", value))
			}
			return value
		},
	}
	t.SetParent(streamReadWriteType)
	registerByteWriteMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	byteWriteType = t
}

func registerTextReadMembers(t *Type) {
	t.registerNativeMethod("getNextRune", func(f *RunFrame, ths *TypeObjPair) {
		r, eof := ths.Obj.(TextRead).GetNextRune()
		if eof {
			f.RP.DataStack.Push(false)
			return
		}
		f.RP.DataStack.Push(Int(r))
		f.RP.DataStack.Push(true)
	})
}

func CreateRegisterTextReadType() {
	t := &Type{
		name:    "TextRead",
		getZero: ReferenceTypeZeroFn,
		checkCoerceType: func(value any) any {
			if value == nil {
				return nil
			}
			if _, ok := value.(TextRead); !ok {
				panic(fmt.Errorf("TextRead.checkCoerceType: expected TextRead, got %T", value))
			}
			return value
		},
	}
	t.SetParent(streamReadWriteType)
	registerTextReadMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	textReadType = t
}

func registerByteReadMembers(t *Type) {
	t.registerNativeMethod("getNextByte", func(f *RunFrame, ths *TypeObjPair) {
		b, eof := ths.Obj.(ByteRead).GetNextByte()
		if eof {
			f.RP.DataStack.Push(false)
			return
		}
		f.RP.DataStack.Push(Int(b))
		f.RP.DataStack.Push(true)
	})
	t.registerNativeMethod("readBytes", func(f *RunFrame, ths *TypeObjPair) {
		n := popIntegralInt(f.RP.DataStack, "readBytes")
		bs := ths.Obj.(ByteRead).ReadBytes(n)
		f.RP.DataStack.Push(bytesToByteList(bs))
	})
}

func CreateRegisterByteReadType() {
	t := &Type{
		name:    "ByteRead",
		getZero: ReferenceTypeZeroFn,
		checkCoerceType: func(value any) any {
			if value == nil {
				return nil
			}
			if _, ok := value.(ByteRead); !ok {
				panic(fmt.Errorf("ByteRead.checkCoerceType: expected ByteRead, got %T", value))
			}
			return value
		},
	}
	t.SetParent(streamReadWriteType)
	registerByteReadMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	byteReadType = t
}

func registerTextReaderMembers(t *Type) {
	// readLine is on the TextReader shell (GetNextRune loop); not on Go TextRead.
	t.registerNativeMethod("readLine", func(f *RunFrame, ths *TypeObjPair) {
		line, eof := ths.Obj.(*TextReader).ReadLine()
		if eof {
			f.RP.DataStack.Push(false)
			return
		}
		f.RP.DataStack.Push(String(line))
		f.RP.DataStack.Push(true)
	})
	// ( path:s -- r ) — panics on error
	t.registerNativeClassMethodSec("newFromFile", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(mustStreamOpen(NewFileTextReader(path)))
	}, SR_FileRead)
	// ( path:s -- r true | err:s false )
	t.registerNativeClassMethodSec("newFromFile?", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		rd, err := NewFileTextReader(path)
		pushStreamOpenResult(f, rd, err)
	}, SR_FileRead)
	t.registerNativeClassMethod("newFromString", func(f *RunFrame, _ *TypeObjPair) {
		s := string(f.RP.DataStack.Pop().(String))
		rd, err := NewTextReaderFromString(s)
		pushStreamOpenResult(f, rd, err)
	})
}

func registerTextWriterMembers(t *Type) {
	// ( path:s -- w ) — panics on error
	t.registerNativeClassMethodSec("newFromFileAppendOrCreate", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(mustStreamOpen(NewFileTextWriterAppendOrCreate(path)))
	}, SR_FileWrite)
	// ( path:s -- w true | err:s false )
	t.registerNativeClassMethodSec("newFromFileAppendOrCreate?", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		wr, err := NewFileTextWriterAppendOrCreate(path)
		pushStreamOpenResult(f, wr, err)
	}, SR_FileWrite)
	// ( path:s -- w ) — panics on error
	t.registerNativeClassMethodSec("newFromFileTruncateOrCreate", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(mustStreamOpen(NewFileTextWriterTruncateOrCreate(path)))
	}, SR_FileWrite)
	// ( path:s -- w true | err:s false )
	t.registerNativeClassMethodSec("newFromFileTruncateOrCreate?", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		wr, err := NewFileTextWriterTruncateOrCreate(path)
		pushStreamOpenResult(f, wr, err)
	}, SR_FileWrite)
}

func registerByteReaderMembers(t *Type) {
	// getNextByte / readBytes inherited from ByteRead.
	// ( path:s -- r ) — panics on error
	t.registerNativeClassMethodSec("newFromFile", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(mustStreamOpen(NewFileByteReader(path)))
	}, SR_FileRead)
	// ( path:s -- r true | err:s false )
	t.registerNativeClassMethodSec("newFromFile?", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		rd, err := NewFileByteReader(path)
		pushStreamOpenResult(f, rd, err)
	}, SR_FileRead)
	t.registerNativeClassMethod("newFromBytes", func(f *RunFrame, _ *TypeObjPair) {
		list := f.RP.DataStack.Pop().(*ByteList)
		rd, err := NewByteReaderFromBytes(list)
		pushStreamOpenResult(f, rd, err)
	})
}

func registerByteWriterMembers(t *Type) {
	// ( path:s -- w ) — panics on error
	t.registerNativeClassMethodSec("newFromFileAppendOrCreate", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(mustStreamOpen(NewFileByteWriterAppendOrCreate(path)))
	}, SR_FileWrite)
	// ( path:s -- w true | err:s false )
	t.registerNativeClassMethodSec("newFromFileAppendOrCreate?", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		wr, err := NewFileByteWriterAppendOrCreate(path)
		pushStreamOpenResult(f, wr, err)
	}, SR_FileWrite)
	// ( path:s -- w ) — panics on error
	t.registerNativeClassMethodSec("newFromFileTruncateOrCreate", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(mustStreamOpen(NewFileByteWriterTruncateOrCreate(path)))
	}, SR_FileWrite)
	// ( path:s -- w true | err:s false )
	t.registerNativeClassMethodSec("newFromFileTruncateOrCreate?", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		wr, err := NewFileByteWriterTruncateOrCreate(path)
		pushStreamOpenResult(f, wr, err)
	}, SR_FileWrite)
	t.registerNativeClassMethod("newFromBytes", func(f *RunFrame, _ *TypeObjPair) {
		list := f.RP.DataStack.Pop().(*ByteList)
		wr, err := NewByteWriterFromBytes(list)
		pushStreamOpenResult(f, wr, err)
	})
}

func CreateRegisterTextReaderType() {
	t := &Type{
		name:            "TextReader",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[TextReader]("TextReader"),
	}
	t.SetParent(textReadType)
	registerTextReaderMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	textReaderType = t
}

func CreateRegisterTextWriterType() {
	t := &Type{
		name:            "TextWriter",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[TextWriter]("TextWriter"),
	}
	registerTextWriterMembers(t)
	t.SetParent(textWriteType)
	GetTypeRegistrar().RegisterCurrentType(t)
	textWriterType = t
}

func CreateRegisterByteReaderType() {
	t := &Type{
		name:            "ByteReader",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[ByteReader]("ByteReader"),
	}
	t.SetParent(byteReadType)
	registerByteReaderMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	byteReaderType = t
}

func CreateRegisterByteWriterType() {
	t := &Type{
		name:            "ByteWriter",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[ByteWriter]("ByteWriter"),
	}
	registerByteWriterMembers(t)
	t.SetParent(byteWriteType)
	GetTypeRegistrar().RegisterCurrentType(t)
	byteWriterType = t
}

// Stream Final types: parents before children (least-dependent → most-dependent).
// SetParent uses package type vars — no registrar name lookup.
func init() {
	CreateRegisterStreamReadWriteType()
	CreateRegisterTextWriteType()
	CreateRegisterByteWriteType()
	CreateRegisterTextReadType()
	CreateRegisterByteReadType()
	CreateRegisterTextReaderType()
	CreateRegisterTextWriterType()
	CreateRegisterByteReaderType()
	CreateRegisterByteWriterType()
	CreateRegisterLineBufferTextWriterType()
	CreateRegisterStringBuilderType() // after StreamReadWrite; SetParent(streamReadWriteType)
}

var (
	streamReadWriteType      *Type
	textWriteType            *Type
	byteWriteType            *Type
	textReadType             *Type
	byteReadType             *Type
	textReaderType           *Type
	textWriterType           *Type
	byteReaderType           *Type
	byteWriterType           *Type
)
