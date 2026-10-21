package core

import (
	"fmt"
	"slices"
)

const DefaultDSStackSize = 32 // avoid nonsesensical small stack sizes

type Stack[T any] struct {
	name      string // for error messages
	items     []T
	nextIndex int // where the next item will be pushed, thus init=0
	maxSize   int
}

func NewStack[T any](name string, maxSize int) *Stack[T] {
	return &Stack[T]{
		name:    name,
		items:   make([]T, 0, DefaultDSStackSize),
		maxSize: maxSize, // no default, must be set; 0 panics on first push
	}
}

// NewAuxStack creates a new auxiliary stack with no initial capacity.
func NewAuxStack[T any](name string, maxSize int) *Stack[T] {
	return &Stack[T]{
		name:    name,
		items:   make([]T, 0),
		maxSize: maxSize, // no default, must be set; 0 panics on first push
	}
}

func (s *Stack[T]) Push(item T) {
	if s.nextIndex >= s.maxSize {
		panic(fmt.Errorf("Stack overflow: %s", s.name))
	}
	if s.nextIndex < len(s.items) {
		s.items[s.nextIndex] = item
	} else {
		s.items = append(s.items, item)
	}
	s.nextIndex++
}

func (s *Stack[T]) Pop() T {
	if s.nextIndex <= 0 {
		panic(fmt.Errorf("Stack underflow: %s", s.name))
	} else {
		s.nextIndex--
        val := s.items[s.nextIndex]
		// clear the value to avoid leaking memory
		var zero T
		s.items[s.nextIndex] = zero
		return val
	}
}

// Clear zeroes live cells and resets depth so GC does not pin values
// between TOS and high-water. Capacity is kept. Nil receiver is a no-op.
func (s *Stack[T]) Clear() {
	if s == nil {
		return
	}
	var zero T
	for i := 0; i < s.nextIndex; i++ {
		s.items[i] = zero
	}
	s.nextIndex = 0
}

func (s *Stack[T]) Peek() T {
	if s.nextIndex <= 0 {
		panic(fmt.Errorf("Stack underflow: %s", s.name))
	} else {
		return s.items[s.nextIndex-1]
	}
}

// Get returns the item at the given index, panics if out of range
// This isn't used much, but is useful for debugging and testing.
func (s *Stack[T]) Get(idx int) T {
	if idx < 0 || idx >= s.nextIndex {
		panic(fmt.Errorf("Stack index out of range: %s", s.name))
	}
	return s.items[idx]
}

// Len returns the number of items currently on the stack.
func (s *Stack[T]) Len() int {
	return s.nextIndex
}

// SnapshotItems returns a shallow clone of the live cells (independent backing array).
func (s *Stack[T]) SnapshotItems() []T {
	if s.nextIndex == 0 {
		return nil
	}
	out := make([]T, s.nextIndex)
	copy(out, s.items[:s.nextIndex])
	return out
}

// RestoreItems replaces live stack contents with snap (shallow copy into s).
func (s *Stack[T]) RestoreItems(snap []T) {
	n := len(snap)
	if n > s.maxSize {
		panic(fmt.Errorf("Stack restore overflow: %s", s.name))
	}
	if cap(s.items) < n {
		s.items = make([]T, n)
	} else {
		s.items = s.items[:n]
	}
	copy(s.items, snap)
	s.nextIndex = n
}

// Return the item at the given index, panics if out of range
// tosIdx is RELATIVE TO TOS aka Peek(), so 0 is TOS/Peek, 1 is TOS-1, etc.
func (s *Stack[T]) At(tosIdx int) T {
	// note: panic checks delegated to AddrAt()
	return *s.AddrAt(tosIdx)
}

// Return a pointer to the item at the given index
// tosIdx is RELATIVE TO TOS aka Peek(), so 0 is TOS/Peek, 1 is TOS-1, etc.
// IMPORTANT: you may NOT mutate the stack while holding a pointer to an item on it
func (s *Stack[T]) AddrAt(tosIdx int) *T {
	if tosIdx < 0 || tosIdx >= s.nextIndex {
		panic(fmt.Errorf("Stack underflow: %s", s.name))
	}
	stackIdx := s.nextIndex - 1 - tosIdx
	return &s.items[stackIdx]
}

// Remove the item at the given index, panics if out of range
// tosIdx is RELATIVE TO TOS aka Pop(), so 0 is TOS/Pop, 1 is TOS-1, etc.
func (s *Stack[T]) RemoveAt(tosIdx int) T {
	if tosIdx == 0 {
		return s.Pop()
	}
	if tosIdx < 0 || tosIdx >= s.nextIndex {
		panic(fmt.Errorf("Stack underflow: %s", s.name))
	}
	stackIdx := s.nextIndex - 1 - tosIdx
	item := s.items[stackIdx]
	s.items = slices.Delete(s.items, stackIdx, stackIdx+1)
	s.nextIndex--
	return item
}

type OrdinalMap[T any] struct {
	byOrdinal []T          // in order they were added
	byName    map[string]T // for efficient lookup by name
}

func NewOrdinalMap[T any]() *OrdinalMap[T] {
	return &OrdinalMap[T]{
		byOrdinal: make([]T, 0),
		byName:    make(map[string]T),
	}
}

func (m *OrdinalMap[T]) Len() int {
	return len(m.byOrdinal)
}

// GetByName returns the item by name, or false if not found
func (m *OrdinalMap[T]) GetByName(name string) (T, bool) {
	item, ok := m.byName[name]
	return item, ok
}

// Get returns the item by orddinal; panics if out of range
func (m *OrdinalMap[T]) GetByOrdinal(ordinal int) T {
	if ordinal < 0 || ordinal >= len(m.byOrdinal) {
		panic(fmt.Errorf("OrdinalMap.GetByOrdinal: out of range %d", ordinal))
	}
	return m.byOrdinal[ordinal]
}

// Add adds an item to the map by name; panics if the name is empty or the item already exists
func (m *OrdinalMap[T]) Add(name string, item T) {
	if name == "" {
		panic(fmt.Errorf("OrdinalMap: Attempt to add item with empty name"))
	}
	if _, ok := m.byName[name]; ok {
		panic(fmt.Errorf("OrdinalMap: item already exists for name: %s", name))
	} else {
		m.byOrdinal = append(m.byOrdinal, item)
		m.byName[name] = item
	}
}
