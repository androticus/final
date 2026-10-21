package core

import "image/color"

type PaletteBase struct {
	colors []Color
}

func (p *PaletteBase) Len() int {
	return len(p.colors)
}

func (p *PaletteBase) At(index int) Color {
	// let runtime panic if index is out of bounds
	// this routine is a blit hot path
	return p.colors[index]
}

func (p *PaletteBase) Put(index int, c Color) {
	p.colors[index] = c
}

// premulRGBA is the palette in Pixmap store form (once per blit).
func (p *PaletteBase) premulRGBA() []color.RGBA {
	out := make([]color.RGBA, len(p.colors))
	for i, c := range p.colors {
		out[i] = colorToPremulRGBA(c)
	}
	return out
}

type Palette1 struct {
	PaletteBase
}

func (p *Palette1) FinalType() *Type { return palette1Type }

func NewEmptyPalette1() *Palette1 {
	return &Palette1{PaletteBase: PaletteBase{colors: make([]Color, 2)}}
}

func NewPalette1(color0, color1 Color) *Palette1 {
	return &Palette1{PaletteBase: PaletteBase{colors: []Color{color0, color1}}}
}

type Palette4 struct {
	PaletteBase
}

func (p *Palette4) FinalType() *Type { return palette4Type }

func NewEmptyPalette4() *Palette4 {
	return &Palette4{PaletteBase: PaletteBase{colors: make([]Color, 16)}}
}

func NewPalette4FromList(colors *AnyList) *Palette4 {
	p := NewEmptyPalette4()
	for i := 0; i < colors.Len(); i++ {
		p.colors[i] = colors.At(i).(Color)
	}
	return p
}

// NewPalette4FromSlice copies RGB from colors into a 16-slot palette, forcing
// every filled slot to straight alpha a (0..255). Slots past len(colors) stay zero.
func NewPalette4FromSlice(colors []Color, a int) *Palette4 {
	p := NewEmptyPalette4()
	n := len(colors)
	if n > len(p.colors) {
		n = len(p.colors)
	}
	for i := 0; i < n; i++ {
		c := colors[i]
		p.colors[i] = NewColor(int(c.R), int(c.G), int(c.B), a)
	}
	return p
}

type Palette8 struct {
	PaletteBase
}

func (p *Palette8) FinalType() *Type { return palette8Type }

func NewEmptyPalette8() *Palette8 {
	return &Palette8{PaletteBase: PaletteBase{colors: make([]Color, 256)}}
}

func NewPalette8FromSlice(colors []Color) *Palette8 {
	p := NewEmptyPalette8()
	copy(p.colors, colors)
	return p
}

// New332Palette8 returns a 256-color 3-3-2 RGB palette packed RRRGGGBB
// (red in the high bits). Channel ramps are linear 8-bit sRGB:
// 3-bit {0,36,73,109,146,182,219,255}, 2-bit {0,85,170,255}.
// Every slot uses straight alpha a (0..255).
func New332Palette8(a int) *Palette8 {
	rgb3 := [...]int{0, 36, 73, 109, 146, 182, 219, 255}
	rgb2 := [...]int{0, 85, 170, 255}
	p := NewEmptyPalette8()
	for i := 0; i < 256; i++ {
		p.colors[i] = NewColor(rgb3[(i>>5)&7], rgb3[(i>>2)&7], rgb2[i&3], a)
	}
	return p
}

func registerPalette1Members(t *Type) {
	t.registerProperty("len", "Int",
		func(inst Any) Any { return Int(inst.(*Palette1).Len()) },
		nil,
	)
	t.registerItemsData("Color",
		func(inst Any, idx Any) any {
			return inst.(*Palette1).At(int(idx.(Int)))
		},
		func(inst Any, idx Any, value Any) {
			inst.(*Palette1).Put(int(idx.(Int)), value.(Color))
		},
		nil, nil, nil,
	)
	// ( -- pal:Palette1 )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewEmptyPalette1())
	})
	// ( c0:Color c1:Color -- pal:Palette1 )
	t.registerNativeClassMethod("newWithColors", func(f *RunFrame, _ *TypeObjPair) {
		c1 := f.RP.DataStack.Pop().(Color)
		c0 := f.RP.DataStack.Pop().(Color)
		f.RP.DataStack.Push(NewPalette1(c0, c1))
	})
}

func registerPalette4Members(t *Type) {
	t.registerProperty("len", "Int",
		func(inst Any) Any { return Int(inst.(*Palette4).Len()) },
		nil,
	)
	t.registerItemsData("Color",
		func(inst Any, idx Any) any {
			return inst.(*Palette4).At(int(idx.(Int)))
		},
		func(inst Any, idx Any, value Any) {
			inst.(*Palette4).Put(int(idx.(Int)), value.(Color))
		},
		nil, nil, nil,
	)
	// ( -- pal:Palette4 )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewEmptyPalette4())
	})
	// ( colors:List -- pal:Palette4 )
	t.registerNativeClassMethod("newFromList", func(f *RunFrame, _ *TypeObjPair) {
		colors := f.RP.DataStack.Pop().(*AnyList)
		f.RP.DataStack.Push(NewPalette4FromList(colors))
	})
	// ( a:Int -- pal:Palette4 )
	t.registerNativeClassMethod("newWebPalette", func(f *RunFrame, _ *TypeObjPair) {
		a := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewPalette4FromSlice(WebColors, a))
	})
}

func registerPalette8Members(t *Type) {
	t.registerProperty("len", "Int",
		func(inst Any) Any { return Int(inst.(*Palette8).Len()) },
		nil,
	)
	t.registerItemsData("Color",
		func(inst Any, idx Any) any {
			return inst.(*Palette8).At(int(idx.(Int)))
		},
		func(inst Any, idx Any, value Any) {
			inst.(*Palette8).Put(int(idx.(Int)), value.(Color))
		},
		nil, nil, nil,
	)
	// ( -- pal:Palette8 )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewEmptyPalette8())
	})
	// ( a:Int -- pal:Palette8 )
	t.registerNativeClassMethod("new332Palette", func(f *RunFrame, _ *TypeObjPair) {
		a := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(New332Palette8(a))
	})
}

func CreateRegisterPalette1Type() {
	t := NewRegisteringType("Palette1",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[Palette1]("Palette1"),
	)
	registerPalette1Members(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	palette1Type = t
}

func CreateRegisterPalette4Type() {
	t := NewRegisteringType("Palette4",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[Palette4]("Palette4"),
	)
	registerPalette4Members(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	palette4Type = t
}

func CreateRegisterPalette8Type() {
	t := NewRegisteringType("Palette8",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[Palette8]("Palette8"),
	)
	registerPalette8Members(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	palette8Type = t
}

var (
	palette1Type *Type
	palette4Type *Type
	palette8Type *Type
	_            Typed = (*Palette1)(nil)
	_            Typed = (*Palette4)(nil)
	_            Typed = (*Palette8)(nil)
)

func init() {
	CreateRegisterPalette1Type()
	CreateRegisterPalette4Type()
	CreateRegisterPalette8Type()
}
