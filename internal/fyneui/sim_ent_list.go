package fyneui

import (
	"slices"
	"sync"

	"github.com/androticus/final/internal/core"
)

const DefaultEntityListSize = 32

// SimEntityList holds live entities by pointer.
// Threadsafe for add/remove.
// Threadsafe iteration via a snapshot iterator.
type SimEntityList struct {
	mutex sync.RWMutex
	items []*SimEntity
}

// SimEntityListIterator walks a snapshot of entity pointers (order at GetIter).
type SimEntityListIterator struct {
	items []*SimEntity
	index int
}

func (it *SimEntityListIterator) GetNext() bool {
	if it.index < (len(it.items) - 1) {
		it.index++
		return true
	}
	return false
}

func (it *SimEntityListIterator) Current() any {
	return it.items[it.index]
}

// NewSimEntityList creates a new SimEntityList.
func NewSimEntityList() *SimEntityList {
	return &SimEntityList{
		items: make([]*SimEntity, 0, DefaultEntityListSize),
	}
}

func (el *SimEntityList) FinalType() *core.Type { return simEntityListType }

// add appends entity and marks active. Prefer SimGame.AddEntity so sg is set.
func (el *SimEntityList) add(entity *SimEntity) {
	el.mutex.Lock()
	defer el.mutex.Unlock()
	el.items = append(el.items, entity)
	entity.active = true
}

func (el *SimEntityList) Remove(entity *SimEntity) {
	el.mutex.Lock()
	defer el.mutex.Unlock()
	entity.active = false
	entity.sg = nil
	for i, item := range el.items {
		if item == entity {
			el.items = append(el.items[:i], el.items[i+1:]...)
			break
		}
	}
}

func (el *SimEntityList) Clear() {
	el.mutex.Lock()
	defer el.mutex.Unlock()
	for _, entity := range el.items {
		entity.active = false
		entity.sg = nil
	}
	el.items = make([]*SimEntity, 0, DefaultEntityListSize)
}

// Snapshot returns a copy of the live entity pointers (for MotionHit / Go callers).
func (el *SimEntityList) Snapshot() []*SimEntity {
	el.mutex.RLock()
	defer el.mutex.RUnlock()
	return slices.Clone(el.items)
}

// GetIter snapshots items under the read lock; foreach is stable against later add/remove.
func (el *SimEntityList) GetIter() core.Iterator {
	return &SimEntityListIterator{
		items: el.Snapshot(),
		index: -1,
	}
}

func registerSimEntityListMembers(t *core.Type) {
	// ( entity:SimEntity -- )
	t.RegisterNativeMethod("clear", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimEntityList).Clear()
	})
	// ( entity:SimEntity -- )
	t.RegisterNativeMethod("remove", func(f *core.RunFrame, ths *core.TypeObjPair) {
		entity := f.RP.DataStack.Pop().(*SimEntity)
		ths.Obj.(*SimEntityList).Remove(entity)
	})
	// ( -- iter:Iterator )
	t.RegisterNativeMethod("getIter", func(f *core.RunFrame, ths *core.TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*SimEntityList).GetIter())
	})
}

func CreateRegisterSimEntityListType() {
	t := core.NewRegisteringType("SimEntityList",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[SimEntityList]("SimEntityList"),
	)
	registerSimEntityListMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	simEntityListType = t
}

var (
	simEntityListType *core.Type
	_                 core.Typed    = (*SimEntityList)(nil)
	_                 core.Iterable = (*SimEntityList)(nil)
	_                 core.Iterator = (*SimEntityListIterator)(nil)
)

func init() {
	CreateRegisterSimEntityListType()
}
