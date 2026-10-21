package core

import (
	"fmt"
	"image/color"
)

// Color holds straight 8-bit channels (R,G,B,A) for authors and matching premul
// bytes (rp,gp,bp,ap) for sinks. Value type — channels are immutable after mint.
type Color struct {
	R, G, B, A     uint8
	rp, gp, bp, ap uint8
}

func (c Color) RGBA() (r, g, b, a uint32) {
	r = uint32(c.rp)
	r |= r << 8
	g = uint32(c.gp)
	g |= g << 8
	b = uint32(c.bp)
	b |= b << 8
	a = uint32(c.ap)
	a |= a << 8
	return r, g, b, a
}

func (c Color) FinalType() *Type { return colorType }

// NewColor creates a Color from 0–255 straight (non-premultiplied) channels.
func NewColor(r, g, b, a int) Color {
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 || a < 0 || a > 255 {
		panic(fmt.Errorf("NewColor: color values must be between 0 and 255: r=%d, g=%d, b=%d, a=%d", r, g, b, a))
	}
	return newColorInternal(r, g, b, a)
}

func NewColorFromFloat(r, g, b, a float64) Color {
	if r < 0 || r > 1 || g < 0 || g > 1 || b < 0 || b > 1 || a < 0 || a > 1 {
		panic(fmt.Errorf("NewColorFromFloat: color values must be between 0.0 and 1.0: r=%f, g=%f, b=%f, a=%f", r, g, b, a))
	}
	return newColorInternal(int(r*255), int(g*255), int(b*255), int(a*255))
}

func newColorInternal(r, g, b, a int) Color {
	r_ := uint8(r)
	g_ := uint8(g)
	b_ := uint8(b)
	a_ := uint8(a)
	var rp, gp, bp, ap uint8
	switch {
	case a_ == 0:
		// Straight R,G,B,A preserved; premul bytes stay zero (nothing to composite).
	case a_ == 0xff:
		rp = r_
		gp = g_
		bp = b_
		ap = a_
	default:
		// Widen before multiply: uint8(200)*uint8(128) wraps to 0.
		rp = uint8(uint16(r_) * uint16(a_) / 255)
		gp = uint8(uint16(g_) * uint16(a_) / 255)
		bp = uint8(uint16(b_) * uint16(a_) / 255)
		ap = a_
	}
	return Color{R: r_, G: g_, B: b_, A: a_, rp: rp, gp: gp, bp: bp, ap: ap}
}

// Named-color indices.
// THIS MUST ONLY CONTAIN 16 VALUES
const (
	ColorBlack = iota
	ColorSilver
	ColorGray
	ColorWhite
	ColorMaroon
	ColorRed
	ColorPurple
	ColorFuchsia
	ColorGreen
	ColorLime
	ColorOlive
	ColorYellow
	ColorNavy
	ColorBlue
	ColorTeal
	ColorAqua

	ColorCount // not a color; length of WebColors
)

// WebColors is indexed by the Color* constants above.
// Values are straight (non-premultiplied) RGBA.
// THIS MUST ONLY CONTAIN 16 VALUES
var WebColors = []Color{
	ColorBlack:   NewColor(0, 0, 0, 255),
	ColorSilver:  NewColor(192, 192, 192, 255),
	ColorGray:    NewColor(128, 128, 128, 255),
	ColorWhite:   NewColor(255, 255, 255, 255),
	ColorMaroon:  NewColor(128, 0, 0, 255),
	ColorRed:     NewColor(255, 0, 0, 255),
	ColorPurple:  NewColor(128, 0, 128, 255),
	ColorFuchsia: NewColor(255, 0, 255, 255), // aka Magenta
	ColorGreen:   NewColor(0, 128, 0, 255),
	ColorLime:    NewColor(0, 255, 0, 255),
	ColorOlive:   NewColor(128, 128, 0, 255),
	ColorYellow:  NewColor(255, 255, 0, 255),
	ColorNavy:    NewColor(0, 0, 128, 255),
	ColorBlue:    NewColor(0, 0, 255, 255),
	ColorTeal:    NewColor(0, 128, 128, 255),
	ColorAqua:    NewColor(0, 255, 255, 255), // aka Cyan
}

var (
	colorTransparent = NewColor(0, 0, 0, 0)
)

func registerColorMembers(t *Type) {
	t.registerProperty("r", "Int",
		func(inst Any) Any { return Int(inst.(Color).R) },
		nil,
	)
	t.registerProperty("g", "Int",
		func(inst Any) Any { return Int(inst.(Color).G) },
		nil,
	)
	t.registerProperty("b", "Int",
		func(inst Any) Any { return Int(inst.(Color).B) },
		nil,
	)
	t.registerProperty("a", "Int",
		func(inst Any) Any { return Int(inst.(Color).A) },
		nil,
	)

	colorClassConst(t, "black", ColorBlack)
	colorClassConst(t, "silver", ColorSilver)
	colorClassConst(t, "gray", ColorGray)
	colorClassConst(t, "white", ColorWhite)
	colorClassConst(t, "maroon", ColorMaroon)
	colorClassConst(t, "red", ColorRed)
	colorClassConst(t, "purple", ColorPurple)
	colorClassConst(t, "fuchsia", ColorFuchsia)
	colorClassConst(t, "green", ColorGreen)
	colorClassConst(t, "lime", ColorLime)
	colorClassConst(t, "olive", ColorOlive)
	colorClassConst(t, "yellow", ColorYellow)
	colorClassConst(t, "navy", ColorNavy)
	colorClassConst(t, "blue", ColorBlue)
	colorClassConst(t, "teal", ColorTeal)
	colorClassConst(t, "aqua", ColorAqua)

	// CMY aliases (same values as aqua / fuchsia)
	colorClassConst(t, "cyan", ColorAqua)
	colorClassConst(t, "magenta", ColorFuchsia)
	t.registerClassProperty("transparent", "Color",
		func(_ Any) Any { return colorTransparent },
		nil,
	)
	t.registerClassProperty("mask", "Color",
		func(_ Any) Any { return colorTransparent },
		nil,
	)

	// ( r:Int g:Int b:Int -- color:Color )
	t.registerNativeClassMethod("rgb", func(f *RunFrame, _ *TypeObjPair) {
		b := f.RP.DataStack.Pop().(Int)
		g := f.RP.DataStack.Pop().(Int)
		r := f.RP.DataStack.Pop().(Int)
		f.RP.DataStack.Push(NewColor(r, g, b, 255))
	})
	// ( r:Int g:Int b:Int a:Int -- color:Color )
	t.registerNativeClassMethod("rgba", func(f *RunFrame, _ *TypeObjPair) {
		a := f.RP.DataStack.Pop().(Int)
		b := f.RP.DataStack.Pop().(Int)
		g := f.RP.DataStack.Pop().(Int)
		r := f.RP.DataStack.Pop().(Int)
		f.RP.DataStack.Push(NewColor(r, g, b, a))
	})
	// ( packed:Int|String -- color:Color )  0xRRGGBB, alpha 255
	t.registerNativeClassMethod("hex", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(colorFromHex(f.RP.DataStack))
	})
	// ( packed:Int|String -- color:Color )  0xRRGGBBAA
	t.registerNativeClassMethod("hexa", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(colorFromHexA(f.RP.DataStack))
	})
	// ( r:Float g:Float b:Float -- color:Color )
	t.registerNativeClassMethod("rgbf", func(f *RunFrame, _ *TypeObjPair) {
		b := f.RP.DataStack.Pop().(Float)
		g := f.RP.DataStack.Pop().(Float)
		r := f.RP.DataStack.Pop().(Float)
		f.RP.DataStack.Push(NewColorFromFloat(r, g, b, 1.0))
	})
	// ( r:Float g:Float b:Float a:Float -- color:Color )
	t.registerNativeClassMethod("rgbaf", func(f *RunFrame, _ *TypeObjPair) {
		a := f.RP.DataStack.Pop().(Float)
		b := f.RP.DataStack.Pop().(Float)
		g := f.RP.DataStack.Pop().(Float)
		r := f.RP.DataStack.Pop().(Float)
		f.RP.DataStack.Push(NewColorFromFloat(r, g, b, a))
	})
}

// colorFromHex unpacks a 0xRRGGBB packed int or hex string (alpha 255).
func colorFromHex(ds *Stack[Any]) Color {
	switch v := ds.Pop().(type) {
	case Int:
		u := uint32(v)
		return NewColor(int(u>>16)&0xff, int(u>>8)&0xff, int(u)&0xff, 255)
	case String:
		u := parseHexColorDigits(string(v), 6, "Color.hex")
		return NewColor(int(u>>16)&0xff, int(u>>8)&0xff, int(u)&0xff, 255)
	default:
		panic("expected Int or String")
	}
}

// colorFromHexA unpacks a 0xRRGGBBAA packed int or hex string.
func colorFromHexA(ds *Stack[Any]) Color {
	switch v := ds.Pop().(type) {
	case Int:
		u := uint32(v)
		return NewColor(int(u>>24), int(u>>16)&0xff, int(u>>8)&0xff, int(u)&0xff)
	case String:
		u := parseHexColorDigits(string(v), 8, "Color.hexa")
		return NewColor(int(u>>24), int(u>>16)&0xff, int(u>>8)&0xff, int(u)&0xff)
	default:
		panic("expected Int or String")
	}
}

// parseHexColorDigits reads wantDigits hex digits; optional 0x|0X prefix.
func parseHexColorDigits(s string, wantDigits int, where string) uint32 {
	if len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		s = s[2:]
	}
	if len(s) != wantDigits {
		panic(fmt.Errorf("%s: need %d hex digits, got %q", where, wantDigits, s))
	}
	var v uint32
	for i := 0; i < len(s); i++ {
		c := s[i]
		var d uint32
		switch {
		case c >= '0' && c <= '9':
			d = uint32(c - '0')
		case c >= 'a' && c <= 'f':
			d = uint32(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			d = uint32(c - 'A' + 10)
		default:
			panic(fmt.Errorf("%s: invalid hex digit %q in %q", where, c, s))
		}
		v = v<<4 | d
	}
	return v
}

func colorClassConst(t *Type, name string, i int) {
	c := WebColors[i]
	t.registerClassProperty(name, "Color",
		func(_ Any) Any { return c },
		nil,
	)
}

func CreateRegisterColorType() {
	t := &Type{
		name:            "Color",
		getZero:         MakeHostTypeZero[Color](),
		checkCoerceType: MakeHostTypeCheckCoerce[Color]("Color"),
	}
	registerColorMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	colorType = t
}

var (
	colorType *Type
)

func init() {
	CreateRegisterColorType()
}

var (
	_ Typed       = (*Color)(nil)
	_ color.Color = (*Color)(nil)
	_ color.Color = Color{}
)
