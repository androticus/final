package core

import "fmt"

// Dyadic set:: ops share OpSetOp; subcode selects the operation.
// Stack: ( a:Set b:Set -- c:Set ) with b on TOS. Both operands must be the same
// sealed type (Set or StringSet); result is a new set of that type.

type SCSetOp = int

const (
	SC_set_union SCSetOp = iota
	SC_set_intersect
	SC_set_difference
	SC_set_symmetricDifference

	SCSetOpCount
)

type SetOpFn func(ds *Stack[any])

var setOpFns = [SCSetOpCount]SetOpFn{
	SC_set_union:               runSetUnion,
	SC_set_intersect:           runSetIntersect,
	SC_set_difference:          runSetDifference,
	SC_set_symmetricDifference: runSetSymmetricDifference,
}

func registerSetOps(d *[]*Word) {
	regSubcodeAxiom(d, "set::union", OpSetOp, SC_set_union, WordFlagAxiom)
	regSubcodeAxiom(d, "set::intersect", OpSetOp, SC_set_intersect, WordFlagAxiom)
	regSubcodeAxiom(d, "set::difference", OpSetOp, SC_set_difference, WordFlagAxiom)
	regSubcodeAxiom(d, "set::symmetricDifference", OpSetOp, SC_set_symmetricDifference, WordFlagAxiom)
}

func newEmptyLike[T comparable](s *Set[T]) *Set[T] {
	return &Set[T]{
		setType:            s.setType,
		coerceSetFromAnyFn: s.coerceSetFromAnyFn,
		items:              make(map[T]bool),
	}
}

func setUnion[T comparable](a, b *Set[T]) *Set[T] {
	out := newEmptyLike(a)
	for k := range a.items {
		out.items[k] = true
	}
	for k := range b.items {
		out.items[k] = true
	}
	return out
}

func setIntersect[T comparable](a, b *Set[T]) *Set[T] {
	out := newEmptyLike(a)
	for k := range a.items {
		if b.items[k] {
			out.items[k] = true
		}
	}
	return out
}

func setDifference[T comparable](a, b *Set[T]) *Set[T] {
	out := newEmptyLike(a)
	for k := range a.items {
		if !b.items[k] {
			out.items[k] = true
		}
	}
	return out
}

func setSymmetricDifference[T comparable](a, b *Set[T]) *Set[T] {
	out := newEmptyLike(a)
	for k := range a.items {
		if !b.items[k] {
			out.items[k] = true
		}
	}
	for k := range b.items {
		if !a.items[k] {
			out.items[k] = true
		}
	}
	return out
}

type setDyadicFn[T comparable] func(a, b *Set[T]) *Set[T]

func runSetDyadic(ds *Stack[any], name string, fnAny setDyadicFn[any], fnStr setDyadicFn[string]) {
	b := ds.Pop()
	a := ds.Pop()
	switch av := a.(type) {
	case *AnySet:
		bv, ok := b.(*AnySet)
		if !ok || bv == nil {
			panic(fmt.Errorf("%s: both operands must be Set, got %T and %T", name, a, b))
		}
		ds.Push(fnAny(av, bv))
	case *StringSet:
		bv, ok := b.(*StringSet)
		if !ok || bv == nil {
			panic(fmt.Errorf("%s: both operands must be StringSet, got %T and %T", name, a, b))
		}
		ds.Push(fnStr(av, bv))
	default:
		panic(fmt.Errorf("%s: want Set or StringSet, got %T", name, a))
	}
}

func runSetUnion(ds *Stack[any]) {
	runSetDyadic(ds, "set::union", setUnion[any], setUnion[string])
}

func runSetIntersect(ds *Stack[any]) {
	runSetDyadic(ds, "set::intersect", setIntersect[any], setIntersect[string])
}

func runSetDifference(ds *Stack[any]) {
	runSetDyadic(ds, "set::difference", setDifference[any], setDifference[string])
}

func runSetSymmetricDifference(ds *Stack[any]) {
	runSetDyadic(ds, "set::symmetricDifference", setSymmetricDifference[any], setSymmetricDifference[string])
}
