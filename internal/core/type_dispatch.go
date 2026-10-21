package core

import "fmt"

func runTypeMemberDispatch(f *RunFrame, kind int, cs *Stack[any]) {
	runTypeMemberDispatchOpcode(f, cs, kind)
}

func runtimeKindForTypeDispatch(kind int) (RuntimeDispatchKind, bool) {
	switch kind {
	case TypeDispatchKindInstanceGet:
		return RuntimeDispatchKindInstanceGetByName, true
	case TypeDispatchKindInstanceSet:
		return RuntimeDispatchKindInstanceSetByName, true
	case TypeDispatchKindInstanceMethod:
		return RuntimeDispatchKindInstanceMethodByName, true
	case TypeDispatchKindClassGet:
		return RuntimeDispatchKindClassGetByName, true
	case TypeDispatchKindClassSet:
		return RuntimeDispatchKindClassSetByName, true
	case TypeDispatchKindClassMethod:
		return RuntimeDispatchKindClassMethodByName, true
	default:
		return 0, false
	}
}

func runtimeKindForItemByKey(kind int) (RuntimeDispatchKind, bool) {
	switch kind {
	case TypeDispatchKindInstanceItemGetByKey:
		return RuntimeDispatchKindInstanceItemsGetByKey, true
	case TypeDispatchKindInstanceItemSetByKey:
		return RuntimeDispatchKindInstanceItemsSetByKey, true
	default:
		return 0, false
	}
}

// InvokeByKind dispatches item access by ordinal index on DS.
func (t *Type) InvokeByKind(f *RunFrame, ths *TypeObjPair, kind int, index int) {
	if t == nil {
		panic(fmt.Errorf("Type.InvokeByKind: nil type"))
	}
	switch kind {
	case TypeDispatchKindInstanceItemGetByOrdinal:
		im := t.requireItemsMetadata()
		pushFetchResult(f, im.GetByOrdinal(ths.Obj, Int(index)))
	case TypeDispatchKindInstanceItemSetByOrdinal:
		im := t.requireItemsMetadata()
		ds := f.RP.DataStack
		if ds.Len() < 1 {
			panic(fmt.Errorf("%s: setByOrdinal: need value on DS", t.name))
		}
		val := ds.Pop()
		im.SetByOrdinal(ths.Obj, Int(index), val)
	default:
		panic(fmt.Errorf("%s.InvokeByKind: kind %d use by-name dispatch", t.name, kind))
	}
}

func (t *Type) requireItemsMetadata() *ItemsMetadata {
	im := t.GetItemsMetadata()
	if im == nil {
		panic(fmt.Errorf("%s: no itemsMetadata", t.name))
	}
	return im
}
