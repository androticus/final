package finalembed

import (
	"time"

	"github.com/androticus/final/internal/core"
)

// ToCell normalizes v to a DS-safe cell before a property get, method result,
// Iterator.Current(), or DataStack push from a native handler.
//
// Final arithmetic and polymorphic ops expect Int, Uint, and Float — not raw Go
// int32/float32/etc. Use ToCell (or explicit Int/Float/… casts) at every boundary.
func ToCell(v any) any {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case bool:
		return Bool(x)
	case string:
		return String(x)
	case int:
		return Int(x)
	case int8:
		return Int(x)
	case int16:
		return Int(x)
	case int32:
		return Int(x)
	case int64:
		return Int(x)
	case uint:
		return Uint(x)
	case uint8:
		return Uint(x)
	case uint16:
		return Uint(x)
	case uint32:
		return Uint(x)
	case uint64:
		return Uint(x)
	case float32:
		return Float(x)
	case float64:
		return Float(x)
	case time.Time:
		return x
	case Interval:
		return x
	case Color:
		return x
	default:
		if _, ok := v.(core.Typed); ok {
			return v
		}
		return v
	}
}

// NewAnyKeyValue builds a KeyValue pair for map- or field-like Iterator.Current()
// results. Keys and values are normalized via ToCell.
func NewAnyKeyValue(key, value any) *AnyKeyValue {
	return &AnyKeyValue{
		Key:   ToCell(key),
		Value: ToCell(value),
	}
}
