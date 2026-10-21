package core

// Scope is one searchable layer in a FindScope chain (innermost = TOS on Scopes stack).
type Scope interface {
	FindWordByName(name string) *Word
}

// FindScope walks an ordered stack of Scope layers (index 0 = outer/base, TOS = innermost).
// Bound to RunProvider; rebuilt or pushed per run/task as nested dicts come online.
type FindScope struct {
	Scopes *Stack[Scope]
}

// Compile-time: DictionaryContext is the primary Scope implementation at inception.
var _ Scope = (*DictionaryContext)(nil)

// NewFindScope creates a scope chain whose outermost (base) layer is baseScope.
func NewFindScope(baseScope Scope) *FindScope {
	if baseScope == nil {
		panic("NewFindScope: nil baseScope")
	}
	fs := &FindScope{
		Scopes: NewStack[Scope]("findScopes", 16),
	}
	fs.Scopes.Push(baseScope)
	return fs
}

// PushScope adds an inner (higher-precedence) search layer (local before global when pushed last).
func (fs *FindScope) PushScope(s Scope) {
	if fs == nil || fs.Scopes == nil {
		panic("FindScope.PushScope: nil receiver")
	}
	if s == nil {
		panic("FindScope.PushScope: nil Scope")
	}
	fs.Scopes.Push(s)
}

func (fs *FindScope) FindWordByName(name string) *Word {
	if fs == nil {
		return nil
	}
	for i := fs.Scopes.Len() - 1; i >= 0; i-- {
		scope := fs.Scopes.Get(i)
		if scope == nil {
			continue
		}
		if w := scope.FindWordByName(name); w != nil {
			return w
		}
	}
	return nil
}

// FindWordWithXT resolves name innermost-first and returns XT from the owning dictionary layer.
func (fs *FindScope) FindWordWithXT(name string) (*Word, XT, bool) {
	if fs == nil {
		return nil, 0, false
	}
	for i := fs.Scopes.Len() - 1; i >= 0; i-- {
		scope := fs.Scopes.Get(i)
		if scope == nil {
			continue
		}
		w := scope.FindWordByName(name)
		if w == nil {
			continue
		}
		dc, ok := scope.(*DictionaryContext)
		if !ok {
			return w, 0, false
		}
		xt, ok := dc.WordXT(w)
		if !ok {
			return w, 0, false
		}
		return w, xt, true
	}
	return nil, 0, false
}

// ResetFindScopeBase replaces the chain with a single base layer (typically global GCD).
func (rp *RunProvider) ResetFindScopeBase(base Scope) {
	if rp == nil {
		panic("ResetFindScopeBase: nil RunProvider")
	}
	if base == nil {
		rp.FindScope = nil
		return
	}
	rp.FindScope = NewFindScope(base)
}

// WithTaskFindScope rebuilds rp.FindScope from Task.currentSearchOrder (front = innermost).
// Legacy fallback: global base + push local when distinct.
// Implementation: module_search.go

