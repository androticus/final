package fyneui

import (
	"slices"

	"github.com/androticus/final/internal/core"
)

type SimCrossing struct {
	entity   *SimEntity
	boundary BoundaryIndex
	at       core.Point
}

func NewSimCrossing(entity *SimEntity, boundary BoundaryIndex, at core.Point) *SimCrossing {
	return &SimCrossing{entity: entity, boundary: boundary, at: at}
}

func (c *SimCrossing) FinalType() *core.Type { return simCrossingType }

// SimCrossings is this step’s crossing list. Authors: foreach only.
type SimCrossings struct {
	items []*SimCrossing
}

func NewSimCrossings() *SimCrossings {
	return &SimCrossings{}
}

func (s *SimCrossings) FinalType() *core.Type { return simCrossingsType }

func (s *SimCrossings) Len() int {
	return len(s.items)
}

func (s *SimCrossings) Add(crossing *SimCrossing) {
	s.items = append(s.items, crossing)
}

// retainLive drops crossings whose entity is nil or not active.
func (s *SimCrossings) retainLive() {
	if s == nil {
		return
	}
	n := 0
	for _, c := range s.items {
		if c != nil && c.entity != nil && c.entity.active {
			s.items[n] = c
			n++
		}
	}
	s.items = s.items[:n]
}

type simCrossingsIterator struct {
	items []*SimCrossing
	index int
}

func (it *simCrossingsIterator) GetNext() bool {
	if it.index < (len(it.items) - 1) {
		it.index++
		return true
	}
	return false
}

func (it *simCrossingsIterator) Current() any {
	return it.items[it.index]
}

func (s *SimCrossings) GetIter() core.Iterator {
	return &simCrossingsIterator{items: slices.Clone(s.items), index: -1}
}

func registerSimCrossingMembers(t *core.Type) {
	t.RegisterProperty("entity", "SimEntity",
		func(inst any) any { return inst.(*SimCrossing).entity },
		nil,
	)
	t.RegisterProperty("boundary", "Int",
		func(inst any) any { return core.Int(inst.(*SimCrossing).boundary) },
		nil,
	)
	t.RegisterProperty("at", "Point",
		func(inst any) any { return inst.(*SimCrossing).at },
		nil,
	)
}

func registerSimCrossingsMembers(t *core.Type) {
	t.RegisterProperty("len", "Int",
		func(inst any) any { return core.Int(inst.(*SimCrossings).Len()) },
		nil,
	)
	// ( -- iter:Iterator )
	t.RegisterNativeMethod("getIter", func(f *core.RunFrame, ths *core.TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*SimCrossings).GetIter())
	})
}

func CreateRegisterSimCrossingType() {
	t := core.NewRegisteringType("SimCrossing",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[SimCrossing]("SimCrossing"),
	)
	registerSimCrossingMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	simCrossingType = t
}

func CreateRegisterSimCrossingsType() {
	t := core.NewRegisteringType("SimCrossings",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[SimCrossings]("SimCrossings"),
	)
	registerSimCrossingsMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	simCrossingsType = t
}

var (
	simCrossingType  *core.Type
	simCrossingsType *core.Type
	_                core.Typed    = (*SimCrossing)(nil)
	_                core.Typed    = (*SimCrossings)(nil)
	_                core.Iterable = (*SimCrossings)(nil)
	_                core.Iterator = (*simCrossingsIterator)(nil)
)

func init() {
	CreateRegisterSimCrossingType()
	CreateRegisterSimCrossingsType()
}
