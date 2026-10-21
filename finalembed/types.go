package finalembed

import (
	"time"

	"github.com/androticus/final/internal/core"
)

// --- Registration API (host type handlers) ---

type (
	Type        = core.Type
	RunFrame    = core.RunFrame
	TypeObjPair = core.TypeObjPair
	TypeReg     = core.TypeReg
)

// NewTypeReg builds a TypeReg for host Go type T. Pass the concrete instance type,
// usually a pointer: NewTypeReg[*Widget]("Widget", registerWidget).
func NewTypeReg[T any](name string, reg func(t *Type)) TypeReg {
	return core.NewTypeReg[T](name, reg)
}

// --- Scalar value types ---

type (
	Bool   = core.Bool
	Int    = core.Int
	Uint   = core.Uint
	Float  = core.Float
	String = core.String
)

// DateTime is Final's sealed name for calendar/clock values; DS cells hold time.Time.
type DateTime = time.Time

// Interval is Final's sealed duration type (opaque wrapper over time.Duration).
type Interval = core.Interval

// Color is a sealed RGBA value type (r/g/b/a as Uint on the Final side).
type Color = core.Color

// Drawing adjuncts (geometry in Y-down drawing space; Point is float64).
type (
	Point       = core.Point
	Size        = core.Size
	Rect        = core.Rect
	Font        = core.Font
	Stroke      = core.Stroke
	Fill        = core.Fill
	LineCap     = core.LineCap
	LineStyle   = core.LineStyle
	FillStyle   = core.FillStyle
	PathBuilder = core.PathBuilder
)

// Painter strokes into a Pixmap (gg backend). Default NewPainter mints the buffer;
// NewPainterWithPixmap binds an existing Pixmap.
type Painter = core.Painter

// --- Collections (concrete sealed Final types only; no List[T]/Map[K,V] aliases) ---

type (
	AnyList    = core.AnyList
	StringList = core.StringList

	AnyMap    = core.AnyMap
	StringMap = core.StringMap

	AnyKeyValue    = core.AnyKeyValue
	StringKeyValue = core.StringKeyValue

	StringKeyValueList = core.StringKeyValueList
	DynamicObject      = core.DynamicObject
	StringBuilder      = core.StringBuilder
)

// --- Strong numeric lists (typed storage; Final Int/Uint/Float cells at the boundary) ---

type (
	Int8List    = core.Int8List
	Int16List   = core.Int16List
	Int32List   = core.Int32List
	IntList     = core.IntList // Go int
	Uint16List  = core.Uint16List
	Uint32List  = core.Uint32List
	UintList    = core.UintList // Go uint
	ByteList    = core.ByteList // unsigned 8-bit storage
	Float32List = core.Float32List
	FloatList   = core.FloatList // Go float64
)

// --- Constructors: general collections ---

func NewAnyList() *AnyList { return core.NewAnyList() }

func NewStringList() *StringList { return core.NewStringList() }

func NewAnyMap() *AnyMap { return core.NewAnyMap() }

func NewStringMap() *StringMap { return core.NewStringMap() }

func NewStringMapFromGoMap(m map[string]string) *StringMap {
	return core.NewStringMapFromGoMap(m)
}

func NewStringKeyValueList() *StringKeyValueList { return core.NewStringKeyValueList() }

func NewDynamicObject() *DynamicObject { return core.NewDynamicObject() }

func NewStringBuilder() *StringBuilder { return core.NewStringBuilder() }

func NewStringBuilderWithPretty(indent string) *StringBuilder {
	return core.NewStringBuilderWithPretty(indent)
}

func NewColor(r, g, b, a int) Color { return core.NewColor(r, g, b, a) }

func NewPoint(x, y int) Point { return core.NewPoint(x, y) }

func NewPointF(x, y float64) Point { return core.NewPointF(x, y) }

func NewSize(w, h int) Size { return core.NewSize(w, h) }

func NewSizeF(w, h float64) Size { return core.NewSizeF(w, h) }

func NewRect(topLeft Point, size Size) Rect { return core.NewRect(topLeft, size) }

func NewPainter(sz Size) *Painter { return core.NewPainter(sz) }

func NewPainterWithPixmap(pm *core.Pixmap) *Painter { return core.NewPainterWithPixmap(pm) }

// --- Constructors: strong numeric lists ---

func NewInt8List() *Int8List { return core.NewInt8List() }

func NewInt16List() *Int16List { return core.NewInt16List() }

func NewInt32List() *Int32List { return core.NewInt32List() }

func NewIntList() *IntList { return core.NewIntList() }

func NewByteList() *ByteList { return core.NewByteList() }

func NewUint16List() *Uint16List { return core.NewUint16List() }

func NewUint32List() *Uint32List { return core.NewUint32List() }

func NewUintList() *UintList { return core.NewUintList() }

func NewFloat32List() *Float32List { return core.NewFloat32List() }

func NewFloatList() *FloatList { return core.NewFloatList() }
