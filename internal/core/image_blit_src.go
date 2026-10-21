package core

import "fmt"

type blitSrcKind int

const (
	blitSrc1 blitSrcKind = iota
	blitSrc4
	blitSrc8
	blitSrcC
)

// BlitSrc pairs a mat with its palette (or Mat2c alone). Kind is fixed at
// construction; change depth by minting a new BlitSrc (no setMatPal).
type BlitSrc struct {
	kind blitSrcKind
	mat  Any
	pal  Any // nil for blitSrcC
}

func (s *BlitSrc) FinalType() *Type { return blitSrcType }

func (s *BlitSrc) Mat() Any { return s.mat }

func (s *BlitSrc) Pal() Any { return s.pal }

func (s *BlitSrc) rowsCols() (rows, cols int) {
	switch s.kind {
	case blitSrc1:
		m := s.mat.(*Mat2i1)
		return m.Rows(), m.Cols()
	case blitSrc4:
		m := s.mat.(*Mat2i4)
		return m.Rows(), m.Cols()
	case blitSrc8:
		m := s.mat.(*Mat2u8)
		return m.Rows(), m.Cols()
	case blitSrcC:
		m := s.mat.(*Mat2C)
		return m.Rows(), m.Cols()
	default:
		panic(fmt.Errorf("BlitSrc: unknown kind %d", s.kind))
	}
}

func NewBlitSrc1(pal *Palette1, mat *Mat2i1) *BlitSrc {
	if pal == nil || mat == nil {
		panic(fmt.Errorf("BlitSrc.new1: pal and mat required"))
	}
	return &BlitSrc{kind: blitSrc1, mat: mat, pal: pal}
}

func NewBlitSrc4(pal *Palette4, mat *Mat2i4) *BlitSrc {
	if pal == nil || mat == nil {
		panic(fmt.Errorf("BlitSrc.new4: pal and mat required"))
	}
	return &BlitSrc{kind: blitSrc4, mat: mat, pal: pal}
}

func NewBlitSrc8(pal *Palette8, mat *Mat2u8) *BlitSrc {
	if pal == nil || mat == nil {
		panic(fmt.Errorf("BlitSrc.new8: pal and mat required"))
	}
	return &BlitSrc{kind: blitSrc8, mat: mat, pal: pal}
}

func NewBlitSrcC(mat *Mat2C) *BlitSrc {
	if mat == nil {
		panic(fmt.Errorf("BlitSrc.newC: mat required"))
	}
	return &BlitSrc{kind: blitSrcC, mat: mat, pal: nil}
}

func (s *BlitSrc) SetMat(mat Any) {
	if mat == nil {
		panic(fmt.Errorf("BlitSrc.setMat: mat required"))
	}
	switch s.kind {
	case blitSrc1:
		if _, ok := mat.(*Mat2i1); !ok {
			panic(fmt.Errorf("BlitSrc.setMat: need Mat2i1, got %T", mat))
		}
	case blitSrc4:
		if _, ok := mat.(*Mat2i4); !ok {
			panic(fmt.Errorf("BlitSrc.setMat: need Mat2i4, got %T", mat))
		}
	case blitSrc8:
		if _, ok := mat.(*Mat2u8); !ok {
			panic(fmt.Errorf("BlitSrc.setMat: need Mat2u8, got %T", mat))
		}
	case blitSrcC:
		if _, ok := mat.(*Mat2C); !ok {
			panic(fmt.Errorf("BlitSrc.setMat: need Mat2c, got %T", mat))
		}
	default:
		panic(fmt.Errorf("BlitSrc.setMat: unknown kind %d", s.kind))
	}
	s.mat = mat
}

func (s *BlitSrc) SetPal(pal Any) {
	switch s.kind {
	case blitSrcC:
		panic(fmt.Errorf("BlitSrc.setPal: Mat2c source has no palette"))
	case blitSrc1:
		if pal == nil {
			panic(fmt.Errorf("BlitSrc.setPal: pal required"))
		}
		if _, ok := pal.(*Palette1); !ok {
			panic(fmt.Errorf("BlitSrc.setPal: need Palette1, got %T", pal))
		}
	case blitSrc4:
		if pal == nil {
			panic(fmt.Errorf("BlitSrc.setPal: pal required"))
		}
		if _, ok := pal.(*Palette4); !ok {
			panic(fmt.Errorf("BlitSrc.setPal: need Palette4, got %T", pal))
		}
	case blitSrc8:
		if pal == nil {
			panic(fmt.Errorf("BlitSrc.setPal: pal required"))
		}
		if _, ok := pal.(*Palette8); !ok {
			panic(fmt.Errorf("BlitSrc.setPal: need Palette8, got %T", pal))
		}
	default:
		panic(fmt.Errorf("BlitSrc.setPal: unknown kind %d", s.kind))
	}
	s.pal = pal
}

func registerBlitSrcMembers(t *Type) {
	t.registerProperty("mat", "Any",
		func(inst Any) Any { return inst.(*BlitSrc).Mat() },
		func(inst Any, val Any) { inst.(*BlitSrc).SetMat(val) },
	)
	t.registerProperty("pal", "Any",
		func(inst Any) Any { return inst.(*BlitSrc).Pal() },
		func(inst Any, val Any) { inst.(*BlitSrc).SetPal(val) },
	)
	// ( pal:Palette1 mat:Mat2i1 -- src:BlitSrc )
	t.registerNativeClassMethod("new1", func(f *RunFrame, _ *TypeObjPair) {
		mat := f.RP.DataStack.Pop().(*Mat2i1)
		pal := f.RP.DataStack.Pop().(*Palette1)
		f.RP.DataStack.Push(NewBlitSrc1(pal, mat))
	})
	// ( pal:Palette4 mat:Mat2i4 -- src:BlitSrc )
	t.registerNativeClassMethod("new4", func(f *RunFrame, _ *TypeObjPair) {
		mat := f.RP.DataStack.Pop().(*Mat2i4)
		pal := f.RP.DataStack.Pop().(*Palette4)
		f.RP.DataStack.Push(NewBlitSrc4(pal, mat))
	})
	// ( pal:Palette8 mat:Mat2u8 -- src:BlitSrc )
	t.registerNativeClassMethod("new8", func(f *RunFrame, _ *TypeObjPair) {
		mat := f.RP.DataStack.Pop().(*Mat2u8)
		pal := f.RP.DataStack.Pop().(*Palette8)
		f.RP.DataStack.Push(NewBlitSrc8(pal, mat))
	})
	// ( mat:Mat2c -- src:BlitSrc )
	t.registerNativeClassMethod("newC", func(f *RunFrame, _ *TypeObjPair) {
		mat := f.RP.DataStack.Pop().(*Mat2C)
		f.RP.DataStack.Push(NewBlitSrcC(mat))
	})
}

func CreateRegisterBlitSrcType() {
	t := NewRegisteringType("BlitSrc",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[BlitSrc]("BlitSrc"),
	)
	registerBlitSrcMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	blitSrcType = t
}

var (
	blitSrcType *Type
	_           Typed = (*BlitSrc)(nil)
)

func init() {
	CreateRegisterBlitSrcType()
}
