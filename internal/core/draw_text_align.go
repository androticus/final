package core

import (
	"fmt"
	"strings"
)

// TextAlign is a packed placement pose: one horizontal bit and one vertical bit.
// Final authors OR an H atom with a V atom (`TextAlign.left TextAlign.baseline |`)
// or assign a composed name (`TextAlign.centerMiddle`). Class consts return Int
// so `|` works. Default on Font is centerMiddle.
type TextAlign int

const (
	textAlignHMask TextAlign = 0x07
	textAlignVMask TextAlign = 0xF0

	TextAlignLeft   TextAlign = 0x01
	TextAlignCenter TextAlign = 0x02
	TextAlignRight  TextAlign = 0x04

	TextAlignTop      TextAlign = 0x10
	TextAlignMiddle   TextAlign = 0x20
	TextAlignBaseline TextAlign = 0x40
	TextAlignBottom   TextAlign = 0x80
)

const (
	TextAlignLeftTop        = TextAlignLeft | TextAlignTop
	TextAlignLeftMiddle     = TextAlignLeft | TextAlignMiddle
	TextAlignLeftBaseline   = TextAlignLeft | TextAlignBaseline
	TextAlignLeftBottom     = TextAlignLeft | TextAlignBottom
	TextAlignCenterTop      = TextAlignCenter | TextAlignTop
	TextAlignCenterMiddle   = TextAlignCenter | TextAlignMiddle
	TextAlignCenterBaseline = TextAlignCenter | TextAlignBaseline
	TextAlignCenterBottom   = TextAlignCenter | TextAlignBottom
	TextAlignRightTop       = TextAlignRight | TextAlignTop
	TextAlignRightMiddle    = TextAlignRight | TextAlignMiddle
	TextAlignRightBaseline  = TextAlignRight | TextAlignBaseline
	TextAlignRightBottom    = TextAlignRight | TextAlignBottom
)

func textAlignOneBit(x TextAlign) bool {
	return x != 0 && x&(x-1) == 0
}

func (a TextAlign) valid() bool {
	if a&^(textAlignHMask|textAlignVMask) != 0 {
		return false
	}
	return textAlignOneBit(a&textAlignHMask) && textAlignOneBit(a&textAlignVMask)
}

func (a TextAlign) horAnchor() float64 {
	switch a & textAlignHMask {
	case TextAlignLeft:
		return 0
	case TextAlignCenter:
		return 0.5
	case TextAlignRight:
		return 1
	default:
		panic(fmt.Errorf("TextAlign: invalid horizontal %#x", a))
	}
}

func registerTextAlignMembers(t *Type) {
	type taConst struct {
		name string
		val  TextAlign
	}
	consts := []taConst{
		{"left", TextAlignLeft},
		{"center", TextAlignCenter},
		{"right", TextAlignRight},
		{"top", TextAlignTop},
		{"middle", TextAlignMiddle},
		{"baseline", TextAlignBaseline},
		{"bottom", TextAlignBottom},
		{"leftTop", TextAlignLeftTop},
		{"leftMiddle", TextAlignLeftMiddle},
		{"leftBaseline", TextAlignLeftBaseline},
		{"leftBottom", TextAlignLeftBottom},
		{"centerTop", TextAlignCenterTop},
		{"centerMiddle", TextAlignCenterMiddle},
		{"centerBaseline", TextAlignCenterBaseline},
		{"centerBottom", TextAlignCenterBottom},
		{"rightTop", TextAlignRightTop},
		{"rightMiddle", TextAlignRightMiddle},
		{"rightBaseline", TextAlignRightBaseline},
		{"rightBottom", TextAlignRightBottom},
	}
	for _, c := range consts {
		v := Int(c.val)
		t.registerClassProperty(c.name, "Int",
			func(_ Any) Any { return v },
			nil,
		)
	}
}

func CreateRegisterTextAlignType() {
	t := &Type{
		name: "TextAlign",
		getZero: func() any {
			return TextAlign(0)
		},
		checkCoerceType: func(value any) any {
			switch x := value.(type) {
			case Int:
				return TextAlign(x)
			default:
				panic(fmt.Errorf("TextAlign.checkCoerceType: value %T is not TextAlign", value))
			}
		},
	}
	registerTextAlignMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	textAlignType = t
}

var textAlignType *Type

func init() {
	CreateRegisterTextAlignType()
}

// drawAlignedLines paints lines at at according to Font.align.
func (p *Painter) drawAlignedLines(lines []string, at Point) {
	n := len(lines)
	if n == 0 {
		return
	}
	a := p.ps.font.align
	dc := p.dc
	lineH := dc.FontHeight()
	m := p.ps.font.face.Metrics()
	ascent := float64(m.Ascent) / 64
	descent := float64(m.Descent) / 64
	boxAbove := ascent
	boxBelow := float64(n-1)*lineH + descent

	var firstBL float64
	switch a & textAlignVMask {
	case TextAlignTop:
		firstBL = at.Y + boxAbove
	case TextAlignMiddle:
		firstBL = at.Y - (boxBelow-boxAbove)/2
	case TextAlignBaseline:
		firstBL = at.Y
	case TextAlignBottom:
		firstBL = at.Y - boxBelow
	default:
		panic(fmt.Errorf("TextAlign: invalid vertical %#x", a))
	}

	ax := a.horAnchor()
	for i, line := range lines {
		w, _ := dc.MeasureString(line)
		x := at.X - ax*w
		y := firstBL + float64(i)*lineH
		dc.DrawString(line, x, y)
	}
}

// drawAlignedText paints s at at according to Font.align. \n is a hard line break.
func (p *Painter) drawAlignedText(s string, at Point) {
	p.drawAlignedLines(strings.Split(s, "\n"), at)
}

// drawWrappedText word-wraps s to width, then places the block like text.
func (p *Painter) drawWrappedText(s string, at Point, width float64) {
	if width <= 0 {
		panic(fmt.Errorf("Painter.textWrap: width must be positive, got %v", width))
	}
	p.drawAlignedLines(p.dc.WordWrap(s, width), at)
}
