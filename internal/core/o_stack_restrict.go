package core

import "fmt"

// oStackAllowsPrefixedWord reports whether w may be used with an o: stack prefix.
// Inst field/collection access (@ . @() etc.) is allowed; OStack layout mutators are not.
func oStackAllowsPrefixedWord(w *Word) bool {
	if w == nil || !w.RequiresStackContext() {
		return false
	}
	return !w.OSContextRestrict()
}

// checkOStackExecAllowed rejects o:-prefixed execution of inst-forbidden words.
func checkOStackExecAllowed(sel ContextStackIndex, w *Word) error {
	if sel != csOStack {
		return nil
	}
	if w == nil {
		return nil
	}
	if w.RequiresStackContext() && w.OSContextRestrict() {
		return fmt.Errorf("word %q is not allowed with o: prefix", w.name)
	}
	return nil
}
