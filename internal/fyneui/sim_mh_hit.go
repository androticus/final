package fyneui

import (
	"slices"

	"github.com/androticus/final/internal/core"
)

// SimHit is one pairwise contact (freeze-for-game or info). at is WC contact.
type SimHit struct {
	entity1 *SimEntity
	entity2 *SimEntity
	at      core.Point
	normal  core.Point
}

func NewSimHit(entity1 *SimEntity, entity2 *SimEntity, at core.Point, normal core.Point) *SimHit {
	return &SimHit{entity1: entity1, entity2: entity2, at: at, normal: normal}
}

func (h *SimHit) FinalType() *core.Type { return simHitType }

// SimHits is this step’s hit list. Authors: foreach only (not a general collection).
type SimHits struct {
	hits []*SimHit
}

func NewSimHits() *SimHits {
	return &SimHits{}
}

func (h *SimHits) FinalType() *core.Type { return simHitsType }

func (h *SimHits) Len() int {
	return len(h.hits)
}

func (h *SimHits) Add(hit *SimHit) {
	h.hits = append(h.hits, hit)
}

// retainLive drops hits whose either entity is nil or not active.
func (h *SimHits) retainLive() {
	if h == nil {
		return
	}
	n := 0
	for _, hit := range h.hits {
		if hit != nil &&
			hit.entity1 != nil && hit.entity1.active &&
			hit.entity2 != nil && hit.entity2.active {
			h.hits[n] = hit
			n++
		}
	}
	h.hits = h.hits[:n]
}

type simHitsIterator struct {
	items []*SimHit
	index int
}

func (it *simHitsIterator) GetNext() bool {
	if it.index < (len(it.items) - 1) {
		it.index++
		return true
	}
	return false
}

func (it *simHitsIterator) Current() any {
	return it.items[it.index]
}

func (h *SimHits) GetIter() core.Iterator {
	return &simHitsIterator{items: slices.Clone(h.hits), index: -1}
}

func registerSimHitMembers(t *core.Type) {
	t.RegisterProperty("entity1", "SimEntity",
		func(inst any) any { return inst.(*SimHit).entity1 },
		nil,
	)
	t.RegisterProperty("entity2", "SimEntity",
		func(inst any) any { return inst.(*SimHit).entity2 },
		nil,
	)
	t.RegisterProperty("at", "Point",
		func(inst any) any { return inst.(*SimHit).at },
		nil,
	)
	t.RegisterProperty("normal", "Point",
		func(inst any) any { return inst.(*SimHit).normal },
		nil,
	)
}

func registerSimHitsMembers(t *core.Type) {
	t.RegisterProperty("len", "Int",
		func(inst any) any { return core.Int(inst.(*SimHits).Len()) },
		nil,
	)
	// ( -- iter:Iterator )
	t.RegisterNativeMethod("getIter", func(f *core.RunFrame, ths *core.TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*SimHits).GetIter())
	})
}

func CreateRegisterSimHitType() {
	t := core.NewRegisteringType("SimHit",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[SimHit]("SimHit"),
	)
	registerSimHitMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	simHitType = t
}

func CreateRegisterSimHitsType() {
	t := core.NewRegisteringType("SimHits",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[SimHits]("SimHits"),
	)
	registerSimHitsMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	simHitsType = t
}

var (
	simHitType  *core.Type
	simHitsType *core.Type
	_           core.Typed    = (*SimHit)(nil)
	_           core.Typed    = (*SimHits)(nil)
	_           core.Iterable = (*SimHits)(nil)
	_           core.Iterator = (*simHitsIterator)(nil)
)

func init() {
	CreateRegisterSimHitType()
	CreateRegisterSimHitsType()
}
