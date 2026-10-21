package core

import "fmt"

// SFM is a small, fast, insertion-order preserving map with string keys.
// Name lookup scans items while len <= sfmMaxIteratedFields; past that, a
// namesMap is filled once and used for name→ordinal thereafter.
//
// Type metadata tables use PropertyMetadataMap / MethodMetadataMap monomorphs.
// They are not Final types: Final may foreach via Iterable, but member lookup/doc uses Type methods.

// Benchmarking by Grok, with analysis of Go runtime, suggests this is
// the optimal number of iterated fields before map-based lookup is used.
const sfmMaxIteratedFields = 64

type SmallFastAnyMap = SFM[any]

type PropertyMetadataMap = SFM[*PropertyMetadata]
type MethodMetadataMap = SFM[*MethodMetadata]

type SFM[V any] struct {
	items    []Item[V]
	namesMap map[string]int // nil until len(items) > sfmMaxIteratedFields
}

type Item[V any] struct {
	Name  string
	Value V
}

// NewSFM creates a new SFM with suggested capacity.
func NewSFM[V any](cap int) *SFM[V] {
	if cap < 0 {
		cap = 0
	}
	return &SFM[V]{items: make([]Item[V], 0, cap)}
}

func (m *SFM[V]) Len() int {
	if m == nil {
		return 0
	}
	return len(m.items)
}

// NOTE: you must NOT mutate the SFM while holding a pointer to an item
func (m *SFM[V]) ItemAt(i int) *Item[V] {
	if i < 0 || i >= len(m.items) {
		panic(fmt.Errorf("SFM.ItemAt: index %d out of range for len %d", i, len(m.items)))
	}
	return &m.items[i]
}

func (m *SFM[V]) lookup(name string) (int, bool) {
	if m == nil {
		return 0, false
	}
	if m.namesMap != nil {
		i, ok := m.namesMap[name]
		return i, ok
	}
	for i := range m.items {
		if m.items[i].Name == name {
			return i, true
		}
	}
	return 0, false
}

func (m *SFM[V]) maybeSpill(name string) {
	if len(m.items) <= sfmMaxIteratedFields {
		return
	}
	if m.namesMap == nil {
		m.namesMap = make(map[string]int, len(m.items))
		for i := range m.items {
			m.namesMap[m.items[i].Name] = i
		}
		return
	}
	m.namesMap[name] = len(m.items) - 1
}

// Add appends name→value. Panics on empty name or duplicate key (add-only profile).
func (m *SFM[V]) Add(name string, value V) {
	if name == "" {
		panic(fmt.Errorf("SFM.Add: empty name"))
	}
	if _, ok := m.lookup(name); ok {
		panic(fmt.Errorf("SFM.Add: duplicate name %q", name))
	}
	m.items = append(m.items, Item[V]{Name: name, Value: value})
	m.maybeSpill(name)
}

func (m *SFM[V]) Set(name string, value V) {
	if i, ok := m.lookup(name); ok {
		m.items[i].Value = value
		return
	}
	m.items = append(m.items, Item[V]{Name: name, Value: value})
	m.maybeSpill(name)
}

// Get returns the value for the given key.
// Panics if the key is not found.
// Use HasKey to check if the key exists.
func (m *SFM[V]) Get(name string) V {
	return m.items[m.IndexOf(name)].Value
}

// IndexOf returns the insertion-order index of name.
// Panics if the key is not found.
func (m *SFM[V]) IndexOf(name string) int {
	if i, ok := m.lookup(name); ok {
		return i
	}
	panic(fmt.Errorf("SFM.IndexOf: key %q not found", name))
}

func (m *SFM[V]) HasKey(name string) bool {
	_, ok := m.lookup(name)
	return ok
}

// Iteration (Go Iterable — used by foreach; no Final type on SFM)

type SFMIterator[V any] struct {
	m     *SFM[V]
	index int
}

func (it *SFMIterator[V]) GetNext() bool {
	if it.m == nil {
		return false
	}
	if it.index < it.m.Len()-1 {
		it.index++
		return true
	}
	return false
}

func (it *SFMIterator[V]) Current() any {
	item := it.m.ItemAt(it.index)
	return &AnyKeyValue{
		Key:   String(item.Name),
		Value: item.Value,
	}
}

func (m *SFM[V]) GetIter() Iterator {
	if m == nil {
		return &NullIter{}
	}
	return &SFMIterator[V]{m: m, index: -1}
}

var (
	_ Iterable = (*PropertyMetadataMap)(nil)
	_ Iterator = (*SFMIterator[*PropertyMetadata])(nil)
	_ Iterable = (*MethodMetadataMap)(nil)
	_ Iterator = (*SFMIterator[*MethodMetadata])(nil)
)
