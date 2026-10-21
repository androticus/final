package core

import "fmt"

// pushFetchResult pushes a fetched scalar onto DS (engine source/sink).
func pushFetchResult(f *RunFrame, value any) {
	if f == nil || f.RP == nil {
		panic("pushFetchResult: nil RunFrame")
	}
	f.RP.DataStack.Push(value)
}

// runNativeInvoke is the shared entry for OpNativeInvoke (receiver on context stack).
func runNativeInvoke(f *RunFrame, cs *Stack[any], ths *TypeObjPair, kind RuntimeDispatchKind) {
	if f == nil || f.RP == nil {
		panic("runNativeInvoke: nil RunFrame")
	}
	if ths == nil || ths.Typ == nil {
		panic("runNativeInvoke: nil ths")
	}
	ths.Typ.Invoke(f, cs, ths, kind)
}

// ReceiverLayoutKind selects how pairFromCS locates the dispatch subject on CS.
type ReceiverLayoutKind int

const (
	receiverLayoutTOS ReceiverLayoutKind = iota
	receiverLayoutBelowOperands
	receiverLayoutStackBottom
	receiverLayoutSkipRefsTOS
)

// ReceiverLayout is the opcode/path contract for CS → ths pairing (no Typed scan).
type ReceiverLayout struct {
	kind          ReceiverLayoutKind
	operandsBelow int
}

func ReceiverLayoutTOS() ReceiverLayout {
	return ReceiverLayout{kind: receiverLayoutTOS}
}

func ReceiverLayoutBelowOperands(n int) ReceiverLayout {
	return ReceiverLayout{kind: receiverLayoutBelowOperands, operandsBelow: n}
}

func ReceiverLayoutStackBottom() ReceiverLayout {
	return ReceiverLayout{kind: receiverLayoutStackBottom}
}

func ReceiverLayoutSkipRefsTOS() ReceiverLayout {
	return ReceiverLayout{kind: receiverLayoutSkipRefsTOS}
}

// liftValue pairs one stack cell via the runtime stub resolver.
func liftValue(rp *RunProvider, cell any) *TypeObjPair {
	if rp == nil {
		panic("liftValue: nil RunProvider")
	}
	return rp.Resolve(cell)
}

// pairFromCS pairs one dispatch subject on the armed context stack (call-time cs).
func pairFromCS(rp *RunProvider, cs *Stack[any], layout ReceiverLayout) *TypeObjPair {
	if rp == nil || cs == nil {
		panic("pairFromCS: nil RunProvider or context stack")
	}
	if cs.Len() < 1 {
		panic("pairFromCS: empty context stack")
	}
	switch layout.kind {
	case receiverLayoutTOS:
		return liftValue(rp, cs.Peek())
	case receiverLayoutBelowOperands:
		n := layout.operandsBelow
		if cs == rp.DataStack {
			if cs.Len() <= n {
				panic(fmt.Errorf("pairFromCS: need receiver below %d operand(s)", n))
			}
			return liftValue(rp, *cs.AddrAt(n))
		}
		return liftValue(rp, cs.Peek())
	case receiverLayoutStackBottom:
		if cs != rp.DataStack {
			panic("pairFromCS: stack-bottom layout requires DS context stack")
		}
		for tosIdx := cs.Len() - 1; tosIdx >= 0; tosIdx-- {
			switch (*cs.AddrAt(tosIdx)).(type) {
			case *PropertyRef, *MethodRef:
				continue
			default:
				return liftValue(rp, *cs.AddrAt(tosIdx))
			}
		}
		panic("pairFromCS: no receiver on context stack")
	case receiverLayoutSkipRefsTOS:
		for tosIdx := 0; tosIdx < cs.Len(); tosIdx++ {
			switch (*cs.AddrAt(tosIdx)).(type) {
			case *PropertyRef, *MethodRef:
				continue
			default:
				return liftValue(rp, *cs.AddrAt(tosIdx))
			}
		}
		panic("pairFromCS: no receiver on context stack")
	default:
		panic(fmt.Errorf("pairFromCS: unknown layout %d", layout.kind))
	}
}

func receiverLayoutForTypeDispatch(kind int) ReceiverLayout {
	switch kind {
	case TypeDispatchKindInstanceItemSetByOrdinal:
		return ReceiverLayoutBelowOperands(2)
	case TypeDispatchKindInstanceItemGetByOrdinal:
		return ReceiverLayoutBelowOperands(1)
	case TypeDispatchKindInstanceMethod:
		return ReceiverLayoutStackBottom()
	default:
		return ReceiverLayoutBelowOperands(1)
	}
}

func receiverLayoutForNativeInvoke(kind RuntimeDispatchKind) ReceiverLayout {
	switch kind {
	case RuntimeDispatchKindInstanceSetByName,
		RuntimeDispatchKindClassSetByName,
		RuntimeDispatchKindInstanceItemsSetByKey,
		RuntimeDispatchKindInstanceItemsSetByOrdinal:
		return ReceiverLayoutBelowOperands(2)
	case RuntimeDispatchKindInstanceItemsSetByMultiIndex:
		// Default Mat2f rank-2 (![]: val + 2 indexes). Prefer runBangMulti for arity peek.
		return ReceiverLayoutBelowOperands(3)
	default:
		return ReceiverLayoutBelowOperands(1)
	}
}

// runNativeInvokeOpcode pairs the CS receiver then dispatches (OpNativeInvoke entry).
func runNativeInvokeOpcode(f *RunFrame, cs *Stack[any], kind RuntimeDispatchKind) {
	ths := pairFromCS(f.RP, cs, receiverLayoutForNativeInvoke(kind))
	runNativeInvoke(f, cs, ths, kind)
}

// runTypeMemberDispatchOpcode pairs the context-stack receiver then dispatches (OpTypeMemberDispatch entry).
func runTypeMemberDispatchOpcode(f *RunFrame, cs *Stack[any], kind int) {
	if f == nil || f.RP == nil {
		panic(fmt.Errorf("typeMemberDispatch: nil frame"))
	}
	ths := pairFromCS(f.RP, cs, receiverLayoutForTypeDispatch(kind))
	if ths.Typ == nil {
		panic(fmt.Errorf("typeMemberDispatch: nil ths.Typ"))
	}
	ds := f.RP.DataStack
	switch kind {
	case TypeDispatchKindInstanceItemGetByOrdinal, TypeDispatchKindInstanceItemSetByOrdinal:
		indexRaw := ds.Pop()
		index, ok := indexRaw.(Int)
		if !ok {
			panic(fmt.Errorf("typeMemberDispatch: index is %T, want Int", indexRaw))
		}
		ths.Typ.InvokeByKind(f, ths, kind, int(index))
	case TypeDispatchKindInstanceItemGetByKey, TypeDispatchKindInstanceItemSetByKey:
		rk, ok := runtimeKindForItemByKey(kind)
		if !ok {
			panic(fmt.Errorf("typeMemberDispatch: bad item kind %d", kind))
		}
		ths.Typ.Invoke(f, cs, ths, rk)
	default:
		name := popNameFromDS(ds)
		rk, ok := runtimeKindForTypeDispatch(kind)
		if !ok {
			panic(fmt.Errorf("typeMemberDispatch: unknown kind %d", kind))
		}
		ths.Typ.InvokeNamed(f, cs, ths, rk, name)
	}
}

// Invoke runs one accessor op on ths; unsupported kind or Inst panics.
func (t *Type) Invoke(f *RunFrame, cs *Stack[any], ths *TypeObjPair, kind RuntimeDispatchKind) {
	if t == nil {
		panic("Type.Invoke: nil type")
	}
	switch kind {
	case RuntimeDispatchKindInstanceItemsGetByOrdinal:
		t.invokeItemsGetByOrdinal(f, ths)
	case RuntimeDispatchKindInstanceItemsSetByOrdinal:
		t.invokeItemsSetByOrdinal(f, ths)
	case RuntimeDispatchKindInstanceItemsGetByKey:
		t.invokeItemsGetByKey(f, ths)
	case RuntimeDispatchKindInstanceItemsSetByKey:
		t.invokeItemsSetByKey(f, ths)
	case RuntimeDispatchKindInstanceItemsHasByKey:
		t.invokeItemsHasByKey(f, ths)
	case RuntimeDispatchKindInstanceItemsGetByMultiIndex:
		t.invokeItemsGetByMultiIndex(f, ths)
	case RuntimeDispatchKindInstanceItemsSetByMultiIndex:
		t.invokeItemsSetByMultiIndex(f, ths)
	case RuntimeDispatchKindInstanceGetByName:
		t.invokeInstanceGetByName(f, ths, popNameFromDS(f.RP.DataStack))
	case RuntimeDispatchKindInstanceSetByName:
		t.invokeInstanceSetByName(f, ths, popNameFromDS(f.RP.DataStack))
	case RuntimeDispatchKindInstanceMethodByName:
		t.invokeInstanceMethodByName(f, ths, popNameFromDS(f.RP.DataStack))
	case RuntimeDispatchKindClassGetByName:
		t.invokeClassGetByName(f, ths, popNameFromDS(f.RP.DataStack))
	case RuntimeDispatchKindClassSetByName:
		t.invokeClassSetByName(f, ths, popNameFromDS(f.RP.DataStack))
	case RuntimeDispatchKindClassMethodByName:
		t.invokeClassMethodByName(f, ths, popNameFromDS(f.RP.DataStack))
	default:
		panic(fmt.Errorf("%s.Invoke: unknown kind %d", t.name, kind))
	}
}

// InvokeNamed runs one by-name member op; name is already popped from DS.
func (t *Type) InvokeNamed(f *RunFrame, cs *Stack[any], ths *TypeObjPair, kind RuntimeDispatchKind, name string) {
	if t == nil {
		panic("Type.InvokeNamed: nil type")
	}
	switch kind {
	case RuntimeDispatchKindInstanceGetByName:
		t.invokeInstanceGetByName(f, ths, name)
	case RuntimeDispatchKindInstanceSetByName:
		t.invokeInstanceSetByName(f, ths, name)
	case RuntimeDispatchKindInstanceMethodByName:
		t.invokeInstanceMethodByName(f, ths, name)
	case RuntimeDispatchKindClassGetByName:
		t.invokeClassGetByName(f, ths, name)
	case RuntimeDispatchKindClassSetByName:
		t.invokeClassSetByName(f, ths, name)
	case RuntimeDispatchKindClassMethodByName:
		t.invokeClassMethodByName(f, ths, name)
	default:
		panic(fmt.Errorf("%s.InvokeNamed: unknown kind %d", t.name, kind))
	}
}

// resolveMethodRefParent returns the baked-in dispatch couplet on ref.
func resolveMethodRefParent(f *RunFrame, ref *MethodRef) *TypeObjPair {
	if f == nil {
		panic("resolveMethodRefParent: nil frame")
	}
	if ref == nil || ref.Parent == nil {
		panic("resolveMethodRefParent: nil ref or Parent")
	}
	return ref.Parent
}

// runFetch implements OpFetch: cs( ValueRef -- value ); consumes the ref.
func runFetch(f *RunFrame, cs *Stack[any]) {
	if f == nil || f.RP == nil {
		panic("runFetch: nil frame")
	}
	if cs == nil {
		panic("runFetch: nil context stack")
	}
	if cs.Len() < 1 {
		panic("@: empty context stack")
	}
	ref := cs.Pop().(ValueRef)
	if pr, ok := ref.(*PropertyRef); ok && pr.PM != nil {
		assertMemberRestrict(f, pr.PM.restrict)
	}
	f.RP.DataStack.Push(ref.Fetch())
}

// runStore implements OpStore: cs( value ValueRef -- ); Forth-style ( n addr -- ) with ref on CS TOS.
// When CS aliases DS, stack is [… value ref] with ref on top. R-M-W: dup @ … swap !
func runStore(f *RunFrame, cs *Stack[any]) {
	if f == nil || f.RP == nil {
		panic("runStore: nil frame")
	}
	if cs == nil {
		panic("runStore: nil context stack")
	}
	ds := f.RP.DataStack
	if ds.Len() < 1 {
		panic("!: empty data stack")
	}
	if cs.Len() < 1 {
		panic("!: empty context stack")
	}
	ref := cs.Pop().(ValueRef)
	if pr, ok := ref.(*PropertyRef); ok && pr.PM != nil {
		assertMemberRestrict(f, pr.PM.restrict)
	}
	val := ds.Pop()
	ref.Store(val)
}

func runAtIndex(f *RunFrame, cs *Stack[any]) {
	runCollectionOp(f, cs, RuntimeDispatchKindInstanceItemsGetByOrdinal, 1)
}

func runBangIndex(f *RunFrame, cs *Stack[any]) {
	runCollectionOp(f, cs, RuntimeDispatchKindInstanceItemsSetByOrdinal, 2)
}

func runAtBrace(f *RunFrame, cs *Stack[any]) {
	runCollectionOp(f, cs, RuntimeDispatchKindInstanceItemsGetByKey, 1)
}

func runBangBrace(f *RunFrame, cs *Stack[any]) {
	runCollectionOp(f, cs, RuntimeDispatchKindInstanceItemsSetByKey, 2)
}

func runQueryBrace(f *RunFrame, cs *Stack[any]) {
	runCollectionOp(f, cs, RuntimeDispatchKindInstanceItemsHasByKey, 1)
}

func runAtMulti(f *RunFrame, cs *Stack[any]) {
	n := peekMultiIndexOperandCount(f, cs, false)
	runCollectionOp(f, cs, RuntimeDispatchKindInstanceItemsGetByMultiIndex, n)
}

func runBangMulti(f *RunFrame, cs *Stack[any]) {
	n := peekMultiIndexOperandCount(f, cs, true)
	runCollectionOp(f, cs, RuntimeDispatchKindInstanceItemsSetByMultiIndex, n)
}

// peekMultiIndexOperandCount finds a Typed receiver with multi-index metadata and returns
// DS operand count (arity for get; arity+1 for set).
func peekMultiIndexOperandCount(f *RunFrame, cs *Stack[any], isSet bool) int {
	const maxArity = 4
	for arity := 1; arity <= maxArity; arity++ {
		nOp := arity
		if isSet {
			nOp = arity + 1
		}
		raw, ok := peekReceiverRaw(f, cs, nOp)
		if !ok {
			continue
		}
		typed, ok := raw.(Typed)
		if !ok || typed == nil {
			continue
		}
		im := typed.FinalType().GetItemsMetadata()
		if im == nil || im.multiIndexArity != arity {
			continue
		}
		if isSet {
			if im.setByMultiIndex == nil {
				continue
			}
		} else if im.getByMultiIndex == nil {
			continue
		}
		return nOp
	}
	name := "@[]"
	if isSet {
		name = "![]"
	}
	panic(fmt.Errorf("%s: no multi-index receiver (arity 1..%d)", name, maxArity))
}

func peekReceiverRaw(f *RunFrame, cs *Stack[any], nOperands int) (any, bool) {
	if f == nil || f.RP == nil || cs == nil || cs.Len() < 1 {
		return nil, false
	}
	if cs == f.RP.DataStack {
		if cs.Len() <= nOperands {
			return nil, false
		}
		return cs.At(nOperands), true
	}
	return cs.Peek(), true
}

// runCollectionOp dispatches aggregate access: source on context stack, operands on DS.
func runCollectionOp(f *RunFrame, cs *Stack[any], kind RuntimeDispatchKind, dsOperands int) {
	if f == nil || f.RP == nil {
		panic("runCollectionOp: nil frame")
	}
	if cs == nil {
		panic("runCollectionOp: nil context stack")
	}
	name := collectionOpName(kind)
	if cs == f.RP.DataStack {
		if cs.Len() <= dsOperands {
			panic(fmt.Errorf("%s: need receiver below %d operand(s)", name, dsOperands))
		}
	} else if cs.Len() < 1 {
		panic(fmt.Errorf("%s: empty context stack", name))
	}
	layout := ReceiverLayoutTOS()
	if cs == f.RP.DataStack {
		layout = ReceiverLayoutBelowOperands(dsOperands)
	}
	ths := pairFromCS(f.RP, cs, layout)
	ths.Typ.Invoke(f, cs, ths, kind)
}

func collectionOpName(kind RuntimeDispatchKind) string {
	switch kind {
	case RuntimeDispatchKindInstanceItemsGetByOrdinal:
		return "@()"
	case RuntimeDispatchKindInstanceItemsSetByOrdinal:
		return "!()"
	case RuntimeDispatchKindInstanceItemsGetByKey:
		return "@{}"
	case RuntimeDispatchKindInstanceItemsSetByKey:
		return "!{}"
	case RuntimeDispatchKindInstanceItemsHasByKey:
		return "?{}"
	case RuntimeDispatchKindInstanceItemsGetByMultiIndex:
		return "@[]"
	case RuntimeDispatchKindInstanceItemsSetByMultiIndex:
		return "![]"
	default:
		return "collection op"
	}
}

// runMethodInvoke implements OpMethodInvoke: cs( MethodRef -- … ); consumes MethodRef before dispatch.
func runMethodInvoke(f *RunFrame, cs *Stack[any]) {
	if f == nil || f.RP == nil {
		panic("runMethodInvoke: nil frame")
	}
	if cs == nil {
		panic("runMethodInvoke: nil context stack")
	}
	if cs.Len() < 1 {
		panic("^: empty context stack")
	}
	top := cs.Pop()
	mr, ok := top.(*MethodRef)
	if !ok {
		if _, ok := top.(*PropertyRef); ok {
			return
		}
		panic(fmt.Errorf("^: TOS is %T, want MethodRef", top))
	}
	ths := resolveMethodRefParent(f, mr)
	mr.MM.invokeCall(f, ths)
}

func (t *Type) invokeInstanceGetByName(f *RunFrame, ths *TypeObjPair, name string) {
	pm := t.GetPropertyMetadata(name)
	if pm == nil {
		panic(fmt.Errorf("%s: unknown property %q", t.name, name))
	}
	pm.invokeGet(f, ths)
}

func (t *Type) invokeInstanceSetByName(f *RunFrame, ths *TypeObjPair, name string) {
	if t.GetMethodMetadata(name) != nil {
		panic(fmt.Errorf("%s: cannot store method %q", t.name, name))
	}
	pm := t.GetPropertyMetadata(name)
	if pm == nil {
		panic(fmt.Errorf("%s: unknown property %q", t.name, name))
	}
	if pm.set == nil {
		panic(fmt.Errorf("%s: read-only property %q", t.name, name))
	}
	pm.invokeSet(f, ths)
}

func (t *Type) invokeInstanceMethodByName(f *RunFrame, ths *TypeObjPair, name string) {
	mm := ths.Typ.GetMethodMetadata(name)
	if mm == nil {
		panic(fmt.Errorf("%s: unknown method %q", t.name, name))
	}
	mm.invokeCall(f, ths)
}

func classDispatchHost(ths *TypeObjPair) *Type {
	if ths == nil {
		return nil
	}
	if lit, ok := ths.Obj.(*Type); ok {
		return lit
	}
	return nil
}

func (t *Type) invokeClassGetByName(f *RunFrame, ths *TypeObjPair, name string) {
	host := classDispatchHost(ths)
	if host == nil {
		panic(fmt.Errorf("%s: class dispatch requires type literal Inst", t.name))
	}
	pm := host.GetClassPropertyMetadata(name)
	if pm == nil {
		panic(fmt.Errorf("%s: unknown class property %q", t.name, name))
	}
	pm.invokeGet(f, ths)
}

func (t *Type) invokeClassSetByName(f *RunFrame, ths *TypeObjPair, name string) {
	host := classDispatchHost(ths)
	if host == nil {
		panic(fmt.Errorf("%s: class dispatch requires type literal Inst", t.name))
	}
	if host.HasClassMethod(name) {
		panic(fmt.Errorf("%s: cannot store class method %q", t.name, name))
	}
	pm := host.GetClassPropertyMetadata(name)
	if pm == nil {
		panic(fmt.Errorf("%s: unknown class property %q", t.name, name))
	}
	if pm.set == nil {
		panic(fmt.Errorf("%s: read-only class property %q", t.name, name))
	}
	pm.invokeSet(f, ths)
}

func (t *Type) invokeClassMethodByName(f *RunFrame, ths *TypeObjPair, name string) {
	host := classDispatchHost(ths)
	if host == nil {
		panic(fmt.Errorf("%s: class dispatch requires type literal Inst", t.name))
	}
	mm := host.GetClassMethodMetadata(name)
	if mm == nil {
		panic(fmt.Errorf("%s: unknown class method %q", t.name, name))
	}
	mm.invokeCall(f, ths)
}

func (t *Type) invokeItemsGetByOrdinal(f *RunFrame, ths *TypeObjPair) {
	im := t.requireItemsMetadata()
	ds := f.RP.DataStack
	if ds.Len() < 1 {
		panic(fmt.Errorf("%s: getByOrdinal: empty DS", t.name))
	}
	idx := ds.Pop()
	pushFetchResult(f, im.GetByOrdinal(ths.Obj, idx))
}

func (t *Type) invokeItemsSetByOrdinal(f *RunFrame, ths *TypeObjPair) {
	im := t.requireItemsMetadata()
	ds := f.RP.DataStack
	if ds.Len() < 2 {
		panic(fmt.Errorf("%s: setByOrdinal: need value and index on DS", t.name))
	}
	idx := ds.Pop()
	val := ds.Pop()
	im.SetByOrdinal(ths.Obj, idx, val)
}

func (t *Type) invokeItemsGetByMultiIndex(f *RunFrame, ths *TypeObjPair) {
	im := t.requireItemsMetadata()
	if im.getByMultiIndex == nil || im.multiIndexArity < 1 {
		panic(fmt.Errorf("%s: no @[] accessor", t.name))
	}
	ds := f.RP.DataStack
	n := im.multiIndexArity
	if ds.Len() < n {
		panic(fmt.Errorf("%s: @[]: need %d index(es) on DS", t.name, n))
	}
	// Pop TOS-first into idxs[n-1]…idxs[0] so source order (r c …) matches [r][c].
	idxs := make([]any, n)
	for i := n - 1; i >= 0; i-- {
		idxs[i] = ds.Pop()
	}
	pushFetchResult(f, im.GetByMultiIndex(ths.Obj, idxs...))
}

func (t *Type) invokeItemsSetByMultiIndex(f *RunFrame, ths *TypeObjPair) {
	im := t.requireItemsMetadata()
	if im.setByMultiIndex == nil || im.multiIndexArity < 1 {
		panic(fmt.Errorf("%s: no ![] accessor", t.name))
	}
	ds := f.RP.DataStack
	n := im.multiIndexArity
	if ds.Len() < n+1 {
		panic(fmt.Errorf("%s: ![]: need value and %d index(es) on DS", t.name, n))
	}
	idxs := make([]any, n)
	for i := n - 1; i >= 0; i-- {
		idxs[i] = ds.Pop()
	}
	val := ds.Pop()
	im.SetByMultiIndex(ths.Obj, val, idxs...)
}

func (t *Type) invokeItemsGetByKey(f *RunFrame, ths *TypeObjPair) {
	im := t.requireItemsMetadata()
	ds := f.RP.DataStack
	if ds.Len() < 1 {
		panic(fmt.Errorf("%s: getByKey: empty DS", t.name))
	}
	key := ds.Pop()
	pushFetchResult(f, im.GetByKey(ths.Obj, key))
}

func (t *Type) invokeItemsSetByKey(f *RunFrame, ths *TypeObjPair) {
	im := t.requireItemsMetadata()
	ds := f.RP.DataStack
	if ds.Len() < 2 {
		panic(fmt.Errorf("%s: setByKey: need value and key on DS", t.name))
	}
	key := ds.Pop()
	val := ds.Pop()
	im.SetByKey(ths.Obj, key, val)
}

func (t *Type) invokeItemsHasByKey(f *RunFrame, ths *TypeObjPair) {
	im := t.requireItemsMetadata()
	ds := f.RP.DataStack
	if ds.Len() < 1 {
		panic(fmt.Errorf("%s: hasByKey: empty DS", t.name))
	}
	key := ds.Pop()
	pushFetchResult(f, im.HasByKey(ths.Obj, key))
}

func popNameFromDS(ds *Stack[any]) string {
	return popNameFromStack(ds)
}

func popNameFromStack(s *Stack[any]) string {
	raw := s.Pop()
	name, ok := raw.(String)
	if !ok {
		panic(fmt.Errorf("expected String name on stack, got %T", raw))
	}
	return string(name)
}
