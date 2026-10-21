package core

import "fmt"

// ReturnStackType is the return stack: value cells of ColonFrameMark.
// Peek/At/PushNew return *ColonFrameMark into the backing slice, never a heap frame.
// Go GC scans items[:len(items)] (high-water), not nextIndex; Toss zeros the whole cell.
type ReturnStackType struct {
	name      string
	items     []ColonFrameMark
	nextIndex int
	maxSize   int
}

func NewReturnStack(name string, maxSize int) *ReturnStackType {
	return &ReturnStackType{
		name:    name,
		items:   make([]ColonFrameMark, 0, DefaultDSStackSize),
		maxSize: maxSize,
	}
}

func (rs *ReturnStackType) Len() int {
	return rs.nextIndex
}

func (rs *ReturnStackType) Peek() *ColonFrameMark {
	if rs.nextIndex <= 0 {
		panic(fmt.Errorf("Stack underflow: %s", rs.name))
	}
	return &rs.items[rs.nextIndex-1]
}

// At returns a pointer to the cell at TOS-relative index (0 = Peek).
func (rs *ReturnStackType) At(tosIdx int) *ColonFrameMark {
	if tosIdx < 0 || tosIdx >= rs.nextIndex {
		panic(fmt.Errorf("Stack underflow: %s", rs.name))
	}
	return &rs.items[rs.nextIndex-1-tosIdx]
}

// PushNew writes a frame into the next cell (value assign, not heap) and returns its address.
// Assigning ColonFrameMark{lastIP} zeros lastDC and adjunct pointers so a recycled cell cannot reuse them.
func (rs *ReturnStackType) PushNew(lastIP int) *ColonFrameMark {
	if rs.nextIndex >= rs.maxSize {
		panic(fmt.Errorf("Stack overflow: %s", rs.name))
	}
	cell := ColonFrameMark{lastIP: lastIP}
	if rs.nextIndex < len(rs.items) {
		rs.items[rs.nextIndex] = cell
	} else {
		rs.items = append(rs.items, cell)
	}
	rs.nextIndex++
	return &rs.items[rs.nextIndex-1]
}

// Toss zeros the TOS cell so GC can collect lastDC and adjunct stacks (slice len still includes the cell), then drops depth.
func (rs *ReturnStackType) Toss() {
	*rs.Peek() = ColonFrameMark{}
	rs.nextIndex--
}

// Clear zeroes live cells and resets depth. Capacity and high-water len are kept. Nil receiver is a no-op.
func (rs *ReturnStackType) Clear() {
	if rs == nil {
		return
	}
	var zero ColonFrameMark
	for i := 0; i < rs.nextIndex; i++ {
		rs.items[i] = zero
	}
	rs.nextIndex = 0
}
