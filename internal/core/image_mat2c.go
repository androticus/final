package core

import (
	"fmt"
	"image"
	"image/color"
	"os"
)

// Mat2C is a row-major matrix of Color cells. Final type name remains Mat2c.
// Implements image.Image (RGBA / premul model, same as Pixmap) for codec interop; matrix @()/@[] use At_.
type Mat2C struct {
	*Mat2[Color]
}

func NewMat2C(rows, cols int) *Mat2C {
	return &Mat2C{newMat2Typed[Color](mat2cType, rows, cols)}
}

func (m *Mat2C) Bounds() image.Rectangle {
	return image.Rect(0, 0, m.Cols(), m.Rows())
}

func (m *Mat2C) ColorModel() color.Model {
	return color.RGBAModel
}

// At implements image.Image (x col, y row). Returns premul color.RGBA like Pixmap.
func (m *Mat2C) At(x, y int) color.Color {
	return colorToPremulRGBA(m.At_(y, x))
}

// NewMat2CFromFile decodes an image file (any format registered with image.Decode) into a Mat2C.
func NewMat2CFromFile(path string) (*Mat2C, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return Mat2CFromImage(img), nil
}

// Mat2CFromImage copies img into a new Mat2C (NRGBA cells).
func Mat2CFromImage(img image.Image) *Mat2C {
	b := img.Bounds()
	rows, cols := b.Dy(), b.Dx()
	m := NewMat2C(rows, cols)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			nrgba := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			m.Put(y-b.Min.Y, x-b.Min.X, NewColor(int(nrgba.R), int(nrgba.G), int(nrgba.B), int(nrgba.A)))
		}
	}
	return m
}

// Crop returns a new Mat2C that is a copy of the rectangle [row0,row0+rows)×[col0,col0+cols).
func (m *Mat2C) Crop(row0, col0, rows, cols int) *Mat2C {
	if m == nil {
		panic("Mat2c.crop: mat is nil")
	}
	if row0 < 0 || col0 < 0 || rows <= 0 || cols <= 0 {
		panic(fmt.Errorf("Mat2c.crop: bad rect row0=%d col0=%d rows=%d cols=%d", row0, col0, rows, cols))
	}
	if row0+rows > m.Rows() || col0+cols > m.Cols() {
		panic(fmt.Errorf("Mat2c.crop: rect [%d,%d)+%dx%d outside %dx%d",
			row0, col0, rows, cols, m.Rows(), m.Cols()))
	}
	out := NewMat2C(rows, cols)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			out.Put(r, c, m.At_(row0+r, col0+c))
		}
	}
	return out
}

func registerMat2CMembers(t *Type) {
	t.registerProperty("rows", "Int",
		func(inst Any) Any { return Int(inst.(*Mat2C).rows) },
		nil,
	)
	t.registerProperty("cols", "Int",
		func(inst Any) Any { return Int(inst.(*Mat2C).cols) },
		nil,
	)
	t.registerProperty("len", "Int",
		func(inst Any) Any { return Int(inst.(*Mat2C).Len()) },
		nil,
	)
	t.registerItemsData2("Color",
		func(inst Any, idx Any) any {
			return inst.(*Mat2C).AtFlat(int(idx.(Int)))
		},
		func(inst Any, idx Any, value Any) {
			inst.(*Mat2C).PutFlat(int(idx.(Int)), value.(Color))
		},
		nil, nil, nil,
		func(inst Any, idxs ...Any) any {
			if len(idxs) != 2 {
				panic(fmt.Errorf("Mat2c @[]: need 2 indexes, got %d", len(idxs)))
			}
			return inst.(*Mat2C).At_(int(idxs[0].(Int)), int(idxs[1].(Int)))
		},
		func(inst Any, value Any, idxs ...Any) {
			if len(idxs) != 2 {
				panic(fmt.Errorf("Mat2c ![]: need 2 indexes, got %d", len(idxs)))
			}
			inst.(*Mat2C).Put(int(idxs[0].(Int)), int(idxs[1].(Int)), value.(Color))
		},
		2,
	)
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		cols := int(f.RP.DataStack.Pop().(Int))
		rows := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewMat2C(rows, cols))
	})
	// ( path:s -- m:Mat2c )
	t.registerNativeClassMethodSec("newFromFile", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		m, err := NewMat2CFromFile(path)
		if err != nil {
			Throw(err.Error(), f)
			return
		}
		f.RP.DataStack.Push(m)
	}, SR_FileRead)
	// ( row0:i col0:i rows:i cols:i -- m:Mat2c )
	t.registerNativeMethod("crop", func(f *RunFrame, ths *TypeObjPair) {
		cols := int(f.RP.DataStack.Pop().(Int))
		rows := int(f.RP.DataStack.Pop().(Int))
		col0 := int(f.RP.DataStack.Pop().(Int))
		row0 := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(ths.Obj.(*Mat2C).Crop(row0, col0, rows, cols))
	})
	t.registerNativeMethod("getRowIter", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*Mat2C).GetRowIter())
	})
	t.registerNativeMethod("getColIter", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*Mat2C).GetColIter())
	})
	t.registerNativeMethod("wrap", func(f *RunFrame, ths *TypeObjPair) {
		mc := ths.Obj.(*Mat2C)
		wrapMat2dOnFrame(f, mc.rows, mc.cols)
	})
}

func registerMat2CType() {
	t := &Type{
		name:            "Mat2c",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Mat2C]("Mat2c"),
	}
	mat2cType = t
	registerMat2CMembers(t)
	ApplyMixin(t, imageType, func(t *Type) {
		registerImageCodecMembers[*Mat2C](t)
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	t.SetParent(matProtoType)
}

var (
	_ Typed       = (*Mat2C)(nil)
	_ Iterable    = (*Mat2C)(nil)
	_ image.Image = (*Mat2C)(nil)
)
