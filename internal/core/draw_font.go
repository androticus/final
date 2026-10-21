package core

import (
	"fmt"
	"sync"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gobolditalic"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/gofont/gosmallcaps"
)

const (
	FontFamilyGo           = "go"
	FontFamilyGoBold       = "goBold"
	FontFamilyGoItalic     = "goItalic"
	FontFamilyGoBoldItalic = "goBoldItalic"
	FontFamilyGoMono       = "goMono"
	FontFamilyGoSmallCaps  = "goSmallCaps"

	defaultFontSize = 12
)

var fontFamilyNames = []string{
	FontFamilyGo,
	FontFamilyGoBold,
	FontFamilyGoItalic,
	FontFamilyGoBoldItalic,
	FontFamilyGoMono,
	FontFamilyGoSmallCaps,
}

type parsedTTF struct {
	once sync.Once
	font *truetype.Font
	err  error
}

var parsedByFamily = map[string]*parsedTTF{
	FontFamilyGo:           {},
	FontFamilyGoBold:       {},
	FontFamilyGoItalic:     {},
	FontFamilyGoBoldItalic: {},
	FontFamilyGoMono:       {},
	FontFamilyGoSmallCaps:  {},
}

func ttfBytesForFamily(family string) []byte {
	switch family {
	case FontFamilyGo:
		return goregular.TTF
	case FontFamilyGoBold:
		return gobold.TTF
	case FontFamilyGoItalic:
		return goitalic.TTF
	case FontFamilyGoBoldItalic:
		return gobolditalic.TTF
	case FontFamilyGoMono:
		return gomono.TTF
	case FontFamilyGoSmallCaps:
		return gosmallcaps.TTF
	default:
		panic(fmt.Errorf("Font.family: unknown %q", family))
	}
}

func parsedTTFForFamily(family string) *truetype.Font {
	p, ok := parsedByFamily[family]
	if !ok {
		panic(fmt.Errorf("Font.family: unknown %q", family))
	}
	p.once.Do(func() {
		p.font, p.err = truetype.Parse(ttfBytesForFamily(family))
	})
	if p.err != nil {
		panic(fmt.Errorf("Font: parse %s: %w", family, p.err))
	}
	return p.font
}

// Font is family + size (points) + glyph color + placement. One lives on Painter; mutate it, do not replace.
type Font struct {
	p      *Painter
	family string
	size   float64
	face   font.Face
	color  Color
	align  TextAlign
}

func (f *Font) FinalType() *Type { return fontType }

func newFont(p *Painter) *Font {
	f := &Font{
		p:      p,
		family: FontFamilyGo,
		size:   defaultFontSize,
		color:  WebColors[ColorBlack],
		align:  TextAlignCenterMiddle,
	}
	f.rebuildFace()
	return f
}

func (f *Font) Family() string { return f.family }

func (f *Font) Size() float64 { return f.size }

func (f *Font) Face() font.Face { return f.face }

func (f *Font) Color() Color { return f.color }

func (f *Font) SetColor(c Color) { f.color = c }

func (f *Font) Align() TextAlign { return f.align }

func (f *Font) SetAlign(a TextAlign) {
	if !a.valid() {
		panic(fmt.Errorf("Font.align: need one horizontal and one vertical flag, got %#x", a))
	}
	f.align = a
}

func (f *Font) setFamilyInternal(family string) {
	if _, ok := parsedByFamily[family]; !ok {
		panic(fmt.Errorf("Font.family: unknown %q", family))
	}
	if f.family == family {
		return
	}
	f.family = family
}

func (f *Font) SetFamily(family string) {
	f.setFamilyInternal(family)
	f.rebuildFace()
}

func (f *Font) setSizeInternal(size float64) {
	if size <= 0 {
		panic(fmt.Errorf("Font.size: must be positive, got %v", size))
	}
	if f.size == size {
		return
	}
	f.size = size
}

func (f *Font) SetSize(size float64) {
	f.setSizeInternal(size)
	f.rebuildFace()
}

func (f *Font) SetFamilyAndSize(family string, size float64) {
	f.setFamilyInternal(family)
	f.setSizeInternal(size)
	f.rebuildFace()
}

func (f *Font) rebuildFace() {
	f.face = truetype.NewFace(parsedTTFForFamily(f.family), &truetype.Options{Size: f.size})
	f.p.dc.SetFontFace(f.face)
}

func fontFamilyConst(t *Type, name string) {
	n := String(name)
	t.registerClassProperty(name, "String",
		func(_ Any) Any { return n },
		nil,
	)
}

func registerFontMembers(t *Type) {
	t.registerProperty("family", "String",
		func(inst Any) Any { return String(inst.(*Font).family) },
		func(inst Any, value Any) { inst.(*Font).SetFamily(string(value.(String))) },
	)
	t.registerProperty("size", "Float",
		func(inst Any) Any { return Float(inst.(*Font).size) },
		func(inst Any, value Any) {
			inst.(*Font).SetSize(float64(checkCoerceNumericToFloat(value)))
		},
	)
	t.registerProperty("color", "Color",
		func(inst Any) Any { return inst.(*Font).color },
		func(inst Any, value Any) { inst.(*Font).SetColor(value.(Color)) },
	)
	t.registerProperty("align", "TextAlign",
		func(inst Any) Any { return Int(inst.(*Font).align) },
		func(inst Any, value Any) { inst.(*Font).SetAlign(value.(TextAlign)) },
	)
	t.registerNativeMethod("setFamilyAndSize", func(f *RunFrame, ths *TypeObjPair) {
		inst := ths.Obj.(*Font)
		size := float64(checkCoerceNumericToFloat(f.RP.DataStack.Pop()))
		family := string(f.RP.DataStack.Pop().(String))
		inst.SetFamilyAndSize(family, size)
	})

	fontFamilyConst(t, FontFamilyGo)
	fontFamilyConst(t, FontFamilyGoBold)
	fontFamilyConst(t, FontFamilyGoItalic)
	fontFamilyConst(t, FontFamilyGoBoldItalic)
	fontFamilyConst(t, FontFamilyGoMono)
	fontFamilyConst(t, FontFamilyGoSmallCaps)
	// ( -- names:StringList )
	t.registerNativeClassMethod("names", func(f *RunFrame, _ *TypeObjPair) {
		cp := append([]string(nil), fontFamilyNames...)
		f.RP.DataStack.Push(NewStringListFromSlice(cp))
	})
}

func CreateRegisterFontType() {
	t := NewRegisteringType("Font",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[Font]("Font"),
	)
	registerFontMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	fontType = t
}

var (
	fontType *Type
	_        Typed = (*Font)(nil)
)

func init() {
	CreateRegisterFontType()
}
