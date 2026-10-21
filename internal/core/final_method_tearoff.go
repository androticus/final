package core

import (
	"fmt"
	"strings"
)

// parseFinalMethodHeader parses Type#method or Type##method.
// ok=false means the name has no `#` (ordinary colon word).
// err != nil means `#`/`##` present but malformed.
func parseFinalMethodHeader(name string) (typeName, methodName string, isClass bool, ok bool, err error) {
	if !strings.Contains(name, "#") {
		return "", "", false, false, nil
	}
	if strings.Contains(name, "##") {
		parts := strings.Split(name, "##")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", false, false, fmt.Errorf("invalid Final method name %q: want Type##method", name)
		}
		if strings.Contains(parts[0], "#") || strings.Contains(parts[1], "#") {
			return "", "", false, false, fmt.Errorf("invalid Final method name %q: extra #", name)
		}
		return parts[0], parts[1], true, true, nil
	}
	parts := strings.Split(name, "#")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false, false, fmt.Errorf("invalid Final method name %q: want Type#method", name)
	}
	return parts[0], parts[1], false, true, nil
}

// prepareFinalMethodTearOff validates ownership and flat namespace before opening a Final method colon.
// Writes a parent-hide warning to rp.Stderr when applicable.
func prepareFinalMethodTearOff(cs *CompilerState, rp *RunProvider, fullName, typeName, methodName string, isClass bool) error {
	if cs == nil || cs.DC == nil || cs.DC.tr == nil {
		return fmt.Errorf("%s: no type registrar on compile DC", fullName)
	}
	typ := cs.DC.tr.GetLocal(typeName)
	if typ == nil {
		return fmt.Errorf("%s: type %q not owned by this dictionary context (local types only)", fullName, typeName)
	}
	if typ.ownMemberNameTaken(methodName) {
		return fmt.Errorf("%s: member %q already exists on %s", fullName, methodName, typeName)
	}
	if parent := typ.GetParent(); parent != nil {
		hidden := false
		if isClass {
			hidden = parent.GetClassMethodMetadata(methodName) != nil || parent.GetClassPropertyMetadata(methodName) != nil
		} else {
			hidden = parent.GetMethodMetadata(methodName) != nil || parent.GetPropertyMetadata(methodName) != nil
		}
		if !hidden {
			// Cross-facet inherit hide: any inherited member with that name.
			hidden = parent.GetMethodMetadata(methodName) != nil ||
				parent.GetPropertyMetadata(methodName) != nil ||
				parent.GetClassMethodMetadata(methodName) != nil ||
				parent.GetClassPropertyMetadata(methodName) != nil
		}
		if hidden && rp != nil && rp.Stderr != nil {
			rp.Stderr.WriteLine(fmt.Sprintf(
				"WARNING: Final method %s hides inherited member %q", fullName, methodName))
			flushTextWrite(rp.Stderr)
		}
	}
	return nil
}

func bootstrapFinalizeFinalMethodTearOff(cs *CompilerState, w *Word) error {
	typeName, methodName, isClass, ok, err := parseFinalMethodHeader(w.name)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("Final method tear-off: %q missing #", w.name)
	}
	typ := cs.DC.tr.GetLocal(typeName)
	if typ == nil {
		return fmt.Errorf("%s: type %q not owned by this dictionary context", w.name, typeName)
	}
	if w.colonCode < 0 {
		return fmt.Errorf("%s: no colonCode", w.name)
	}
	typ.registerFinalMethodMeta(methodName, w.moduleId, int(w.colonCode), w.doc, isClass)
	return nil
}
