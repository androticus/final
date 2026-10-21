package core

import (
	"fmt"
	"strings"
)

// runImport is OpImport: ( syn:s name:s -- )
func runImport(t *Task, rp *RunProvider) {
	assertCallerDCOpen(t, "import")
	if t.session == nil || t.session.modules == nil {
		panic("import: Task has no Session.modules")
	}
	nameArg := string(rp.DataStack.Pop().(String))
	syn := string(rp.DataStack.Pop().(String))
	importModule(t, syn, nameArg)
}

// importModule attaches an already-installed module to the Session (first-attach wins).
func importModule(t *Task, syn, nameArg string) {
	if t == nil {
		panic("operation not allowed")
	}
	AssertSR(SR_ModuleImport, t.restrict)
	identity, _ := ParseModuleRef(nameArg)
	identity = strings.TrimSpace(identity)
	if identity == "" {
		panic("import: empty module name")
	}
	mod := ModuleTable.moduleByName(identity)
	if mod == nil {
		panic(fmt.Errorf("import: module %q is not installed", identity))
	}
	syn = strings.TrimSpace(syn)
	if syn == "" {
		syn = mod.prefix
	}
	if syn == "" {
		panic(fmt.Errorf("import: module %q has empty prefix", identity))
	}
	if syn == "local" || syn == "core" {
		panic(fmt.Errorf("import: prefix %q is reserved", syn))
	}

	sess := t.session
	if existing := findAttachedModule(sess.modules, mod); existing != nil {
		if existing.prefix == syn {
			return // exact re-import — nop
		}
		panic(fmt.Errorf("import: module %q already attached as %q; cannot re-attach as %q", identity, existing.prefix, syn))
	}
	if conflict := findModuleRefByPrefix(sess.modules, syn); conflict != nil {
		panic(fmt.Errorf("import: prefix %q already in use", syn))
	}

	ref := NewModuleReference(mod, syn)
	sess.modules.Append(ref)
	if t.currentSearchOrder != nil && t.currentSearchOrder != sess.modules {
		t.currentSearchOrder.Append(ref)
	}
}

func findAttachedModule(list *ModuleReferenceList, mod *Module) *ModuleReference {
	if list == nil || mod == nil {
		return nil
	}
	for i := 0; i < list.Len(); i++ {
		ref := list.At(i)
		if ref != nil && ref.module == mod {
			return ref
		}
	}
	return nil
}
