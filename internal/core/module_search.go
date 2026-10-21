package core

import (
	"fmt"
	"strings"
)

// splitModuleQualified splits "prefix\tail". ok is false if name has no backslash.
// Empty prefix or empty tail → ok false (treated as non-qualified / not found path).
func splitModuleQualified(name string) (prefix, tail string, ok bool) {
	i := strings.IndexByte(name, '\\')
	if i <= 0 || i >= len(name)-1 {
		return "", "", false
	}
	return name[:i], name[i+1:], true
}

// modulesListForFind returns Session.modules when available, else Task.currentSearchOrder.
func modulesListForFind(t *Task) *ModuleReferenceList {
	if t == nil {
		return nil
	}
	if t.session != nil && t.session.modules != nil {
		return t.session.modules
	}
	return t.currentSearchOrder
}

func findWordInModuleRefList(list *ModuleReferenceList, name string) (*Word, XT, bool) {
	if list == nil {
		return nil, 0, false
	}
	for i := 0; i < list.Len(); i++ {
		ref := list.At(i)
		if ref == nil || ref.module == nil || ref.module.dc == nil {
			continue
		}
		w, xt, ok := findNameInContext(ref.module.dc, String(name))
		if ok {
			return w, xt, true
		}
	}
	return nil, 0, false
}

func findWordQualified(t *Task, prefix, tail string) (*Word, XT, bool) {
	list := modulesListForFind(t)
	if list == nil {
		return nil, 0, false
	}
	ref := findModuleRefByPrefix(list, prefix)
	if ref == nil || ref.module == nil || ref.module.dc == nil {
		return nil, 0, false
	}
	return findNameInContext(ref.module.dc, String(tail))
}

// FindNameInTask resolves a bare or qualified (prefix\tail) name via Task.currentSearchOrder.
// Falls back to local DC only when no search order is set. On success, assertWordAllowed
// checks SR bits and WordFlagSystem.
func FindNameInTask(t *Task, rp *RunProvider, name String) (*Word, XT, bool) {
	if t == nil || t.dc == nil {
		panic("FindNameInTask: nil Task or DictionaryContext")
	}
	nameStr := string(name)
	var w *Word
	var xt XT
	var ok bool

	if prefix, tail, isQual := splitModuleQualified(nameStr); isQual {
		w, xt, ok = findWordQualified(t, prefix, tail)
	} else if t.currentSearchOrder != nil {
		w, xt, ok = findWordInModuleRefList(t.currentSearchOrder, nameStr)
		if !ok && rp != nil && rp.FindScope != nil {
			// Nested PushScope layers (tests / future) beyond the ModuleReference CSO.
			w, xt, ok = rp.FindScope.FindWordWithXT(nameStr)
		}
	} else if rp != nil && rp.FindScope != nil {
		w, xt, ok = rp.FindScope.FindWordWithXT(nameStr)
	} else {
		w, xt, ok = findNameInContext(t.dc, name)
	}
	if !ok {
		return nil, 0, false
	}
	assertWordAllowed(t, w)
	return w, xt, true
}

// WithTaskFindScope rebuilds rp.FindScope from Task.currentSearchOrder (front = innermost).
// Legacy fallback: global base + push local when distinct.
func (rp *RunProvider) WithTaskFindScope(t *Task) {
	if rp == nil {
		return
	}
	if t != nil && t.currentSearchOrder != nil && t.currentSearchOrder.Len() > 0 {
		n := t.currentSearchOrder.Len()
		baseRef := t.currentSearchOrder.At(n - 1)
		if baseRef == nil || baseRef.module == nil || baseRef.module.dc == nil {
			rp.FindScope = nil
			return
		}
		rp.ResetFindScopeBase(baseRef.module.dc)
		for i := n - 2; i >= 0; i-- {
			ref := t.currentSearchOrder.At(i)
			if ref != nil && ref.module != nil && ref.module.dc != nil {
				rp.FindScope.PushScope(ref.module.dc)
			}
		}
		return
	}
	var base Scope
	if globalDictionaryContext != nil {
		base = globalDictionaryContext
	} else if t != nil && t.dc != nil {
		base = t.dc
	} else {
		rp.FindScope = nil
		return
	}
	rp.ResetFindScopeBase(base)
	if t != nil && t.dc != nil && t.dc != globalDictionaryContext {
		rp.FindScope.PushScope(t.dc)
	}
}

func runUsing(t *Task, rp *RunProvider) {
	if t == nil {
		panic("using: nil Task")
	}
	sl := rp.DataStack.Pop().(*StringList)
	if sl.Len() == 0 {
		return
	}
	src := modulesListForFind(t)
	if src == nil || src.Len() == 0 {
		panic("using: no Session.modules / search order")
	}
	prior := t.currentSearchOrder
	if prior == nil {
		prior = src
	}
	front := make([]*ModuleReference, 0, sl.Len())
	seen := make(map[*ModuleReference]bool, sl.Len())
	for i := 0; i < sl.Len(); i++ {
		p := sl.At(i)
		ref := findModuleRefByPrefix(src, p)
		if ref == nil {
			panic(fmt.Errorf("using: unknown prefix %q", p))
		}
		if seen[ref] {
			continue
		}
		front = append(front, ref)
		seen[ref] = true
	}
	out := NewModuleReferenceList()
	for _, ref := range front {
		out.Append(ref)
	}
	for i := 0; i < prior.Len(); i++ {
		ref := prior.At(i)
		if ref == nil || seen[ref] {
			continue
		}
		out.Append(ref)
	}
	t.currentSearchOrder = out
}

func runUsingDefault(t *Task) {
	if t == nil {
		panic("using-default: nil Task")
	}
	if t.session == nil || t.session.modules == nil {
		panic("using-default: Task has no Session.modules")
	}
	t.currentSearchOrder = t.session.modules
}
