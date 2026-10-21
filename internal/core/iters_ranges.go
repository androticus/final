package core

import "fmt"

// Iterator and range implementation

type Iterator interface {
	GetNext() bool
	Current() any
}

type Iterable interface {
	GetIter() Iterator
}

type Range struct {
	beg int
	end int
}

type RangeBy struct {
	beg Any
	end Any
	by  Any
}

type RangeIterator struct {
	beg  int
	end  int
	curr int
}

type RangeByIterator struct {
	beg  Any
	end  Any
	by   Any
	curr Any
}

// NullIter is a dummy iterator that always returns false and panics on Current.
// Can be used by any provider. Useful for nil-safe iteration.
type NullIter struct{}

func (ni *NullIter) GetNext() bool {
	return false
}

func (ni *NullIter) Current() any {
	panic(fmt.Errorf("NullIter.Current: called"))
}

func assertRangeValid(beg, end int) {
	empty := beg == end+1
	if !empty && beg > end {
		panic(fmt.Errorf("range is invalid: requires beg <= end (or empty beg == end+1)"))
	}
}

func assertRangeByValid(beg, end, by Any) {
	if !IsNumeric(beg) || !IsNumeric(end) || !IsNumeric(by) {
		panic(fmt.Errorf("rangeby requires numeric beg/end/by, got %T %T %T", beg, end, by))
	}
	switch by.(type) {
	case Int:
		if _, ok := beg.(Int); !ok {
			panic(fmt.Errorf("rangeby requires same numeric type, got beg %T by Int", beg))
		}
		if _, ok := end.(Int); !ok {
			panic(fmt.Errorf("rangeby requires same numeric type, got end %T by Int", end))
		}
		b, e, s := beg.(Int), end.(Int), by.(Int)
		if s == 0 {
			panic(fmt.Errorf("range by must not be 0"))
		}
		empty := b == e+s
		if s > 0 {
			if !empty && b > e {
				panic(fmt.Errorf("range is invalid: positive by requires beg <= end (or empty beg == end+by)"))
			}
			return
		}
		if empty {
			return
		}
		if b < e {
			panic(fmt.Errorf("range is invalid: negative by requires beg >= end (or empty beg == end+by)"))
		}
	case Uint:
		if _, ok := beg.(Uint); !ok {
			panic(fmt.Errorf("rangeby requires same numeric type, got beg %T by Uint", beg))
		}
		if _, ok := end.(Uint); !ok {
			panic(fmt.Errorf("rangeby requires same numeric type, got end %T by Uint", end))
		}
		b, e, s := beg.(Uint), end.(Uint), by.(Uint)
		if s == 0 {
			panic(fmt.Errorf("range by must not be 0"))
		}
		empty := b == e+s
		if !empty && b > e {
			panic(fmt.Errorf("range is invalid: positive by requires beg <= end (or empty beg == end+by)"))
		}
	case Float:
		if _, ok := beg.(Float); !ok {
			panic(fmt.Errorf("rangeby requires same numeric type, got beg %T by Float", beg))
		}
		if _, ok := end.(Float); !ok {
			panic(fmt.Errorf("rangeby requires same numeric type, got end %T by Float", end))
		}
		b, e, s := beg.(Float), end.(Float), by.(Float)
		if s == 0 {
			panic(fmt.Errorf("range by must not be 0"))
		}
		empty := b == e+s
		if s > 0 {
			if !empty && b > e {
				panic(fmt.Errorf("range is invalid: positive by requires beg <= end (or empty beg == end+by)"))
			}
			return
		}
		if empty {
			return
		}
		if b < e {
			panic(fmt.Errorf("range is invalid: negative by requires beg >= end (or empty beg == end+by)"))
		}
	default:
		panic(fmt.Errorf("rangeby requires Int, Uint, or Float, got %T", by))
	}
}

func NewRange(rp *RunProvider) *Range {
	ds := rp.DataStack
	end := int(ds.Pop().(Int))
	beg := int(ds.Pop().(Int))
	assertRangeValid(beg, end)
	return &Range{end: end, beg: beg}
}

func NewRangeBy(rp *RunProvider) *RangeBy {
	ds := rp.DataStack
	by := ds.Pop()
	end := ds.Pop()
	beg := ds.Pop()
	assertRangeByValid(beg, end, by)
	return &RangeBy{by: by, end: end, beg: beg}
}

func (r *Range) FinalType() *Type { return GetTypeRegistrar().Get("Range") }

func (rb *RangeBy) FinalType() *Type { return GetTypeRegistrar().Get("RangeBy") }

func (ri *RangeIterator) GetNext() bool {
	if ri.curr == ri.end {
		return false
	}
	ri.curr++
	return true
}

func (ri *RangeIterator) Current() any {
	return Int(ri.curr)
}

func (ri *RangeByIterator) GetNext() bool {
	switch by := ri.by.(type) {
	case Int:
		curr := ri.curr.(Int)
		end := ri.end.(Int)
		next := curr + by
		if by > 0 {
			if next > end {
				return false
			}
		} else if next < end {
			return false
		}
		ri.curr = next
		return true
	case Uint:
		curr := ri.curr.(Uint)
		end := ri.end.(Uint)
		next := curr + by
		if next > end {
			return false
		}
		ri.curr = next
		return true
	case Float:
		curr := ri.curr.(Float)
		end := ri.end.(Float)
		next := curr + by
		if by > 0 {
			if next > end {
				return false
			}
		} else if next < end {
			return false
		}
		ri.curr = next
		return true
	default:
		panic(fmt.Errorf("RangeByIterator.GetNext: unsupported by type %T", ri.by))
	}
}

func (ri *RangeByIterator) Current() any {
	return ri.curr
}

func (r *Range) GetIter() Iterator {
	return &RangeIterator{
		beg:  r.beg,
		end:  r.end,
		curr: r.beg - 1,
	}
}

func (rb *RangeBy) GetIter() Iterator {
	var curr Any
	switch by := rb.by.(type) {
	case Int:
		curr = rb.beg.(Int) - by
	case Uint:
		curr = rb.beg.(Uint) - by
	case Float:
		curr = rb.beg.(Float) - by
	default:
		panic(fmt.Errorf("RangeBy.GetIter: unsupported by type %T", rb.by))
	}
	return &RangeByIterator{
		beg:  rb.beg,
		end:  rb.end,
		by:   rb.by,
		curr: curr,
	}
}

func registerRangeMembers(t *Type) {
	t.registerProperty("beg", "Int",
		func(inst Any) Any { return Int(inst.(*Range).beg) },
		nil,
	)
	t.registerProperty("end", "Int",
		func(inst Any) Any { return Int(inst.(*Range).end) },
		nil,
	)
}

func registerRangeByMembers(t *Type) {
	t.registerProperty("beg", "Any",
		func(inst Any) Any { return inst.(*RangeBy).beg },
		nil,
	)
	t.registerProperty("end", "Any",
		func(inst Any) Any { return inst.(*RangeBy).end },
		nil,
	)
	t.registerProperty("by", "Any",
		func(inst Any) Any { return inst.(*RangeBy).by },
		nil,
	)
}

func CreateRegisterRangeType() {
	t := &Type{
		name:            "Range",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[*Range]("Range"),
	}
	registerRangeMembers(t)
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewRange(f.RP))
	})
	GetTypeRegistrar().RegisterCurrentType(t)
}

func CreateRegisterRangeByType() {
	t := &Type{
		name:            "RangeBy",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[*RangeBy]("RangeBy"),
	}
	registerRangeByMembers(t)
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewRangeBy(f.RP))
	})
	GetTypeRegistrar().RegisterCurrentType(t)
}

func init() {
	CreateRegisterRangeType()
	CreateRegisterRangeByType()
}

// Validate Interfaces

var (
	_ Iterable = (*Range)(nil)
	_ Iterable = (*RangeBy)(nil)
	_ Iterator = (*RangeIterator)(nil)
	_ Iterator = (*RangeByIterator)(nil)
	_ Typed    = (*Range)(nil)
	_ Typed    = (*RangeBy)(nil)
)
