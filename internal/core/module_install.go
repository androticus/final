package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const moduleFileName = "_MODULE_"

// moduleManifestWire is the on-disk _MODULE_ JSON shape used by leaf install vetting.
type moduleManifestWire struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Final       string            `json:"final"`
	Version     string            `json:"version"`
	Requires    []json.RawMessage `json:"requires"`
	Files       []string          `json:"files"`
}

func readModuleManifest(root string) (wire moduleManifestWire, jsonText string, err error) {
	path := filepath.Join(root, moduleFileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		return wire, "", fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		return wire, "", fmt.Errorf("parse %s: %w", path, err)
	}
	wire.Name = strings.TrimSpace(wire.Name)
	wire.Version = strings.TrimSpace(wire.Version)
	wire.Final = strings.TrimSpace(wire.Final)
	return wire, string(raw), nil
}

func vetLeafModuleManifest(wire moduleManifestWire) error {
	if wire.Name == "" {
		return fmt.Errorf("install: _MODULE_ missing name")
	}
	if wire.Version == "" {
		return fmt.Errorf("install: _MODULE_ missing version")
	}
	if wire.Final == "" {
		return fmt.Errorf("install: _MODULE_ missing final")
	}
	if len(wire.Requires) > 0 {
		return fmt.Errorf("install: module %q has requires; Phase 2 supports leaf modules only", wire.Name)
	}
	if len(wire.Files) == 0 {
		return fmt.Errorf("install: _MODULE_ files list is empty")
	}
	for _, f := range wire.Files {
		f = strings.TrimSpace(f)
		if f == "" {
			return fmt.Errorf("install: empty path in files")
		}
		if err := vetModuleSourcePath(f); err != nil {
			return err
		}
	}
	return nil
}

func vetModuleSourcePath(p string) error {
	if filepath.IsAbs(p) {
		return fmt.Errorf("install: files entry must be relative to module root, got absolute %q", p)
	}
	clean := filepath.Clean(p)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("install: files entry %q escapes module root", p)
	}
	if !filepath.IsLocal(clean) {
		return fmt.Errorf("install: files entry %q is not local to module root", p)
	}
	return nil
}

func assertCallerDCOpen(t *Task, op string) {
	if t == nil || t.dc == nil {
		panic(fmt.Errorf("%s: nil Task or DictionaryContext", op))
	}
	if t.dc.IsLockedOrSealed() {
		reason := "sealed"
		if !t.dc.sealed {
			reason = "locked"
		}
		panic(fmt.Errorf("%s: local DictionaryContext is %s", op, reason))
	}
}

// runInstall is OpInstall: ( name:s opts:DynamicObject -- )
func runInstall(t *Task, rp *RunProvider) {
	assertCallerDCOpen(t, "install")
	opts := rp.DataStack.Pop().(*DynamicObject)
	nameArg := string(rp.DataStack.Pop().(String))
	installLeafModule(t, nameArg, opts)
}

// installLeafModule resolves, compiles, seals, and registers a no-deps module.
// Already-present FQN → nop. Requested @version is ignored for selection (Phase 2).
func installLeafModule(caller *Task, nameArg string, opts *DynamicObject) {
	if caller == nil {
		panic("operation not allowed")
	}
	AssertSR(SR_ModuleInstall, caller.restrict)
	identity, _ := ParseModuleRef(nameArg)
	if identity == "" {
		panic("install: empty module name")
	}
	root, viaReplace, err := resolveModuleRoot(identity, opts)
	if err != nil {
		panic(err)
	}
	st, err := os.Stat(root)
	if err != nil || !st.IsDir() {
		panic(fmt.Errorf("install: module root %q is not a directory", root))
	}
	wire, jsonText, err := readModuleManifest(root)
	if err != nil {
		panic(err)
	}
	if err := vetLeafModuleManifest(wire); err != nil {
		panic(err)
	}
	if viaReplace && wire.Name != identity {
		panic(fmt.Errorf("install: _MODULE_ name %q does not match replace identity %q", wire.Name, identity))
	}
	if existing := ModuleTable.moduleByName(wire.Name); existing != nil {
		return // already installed — nop
	}

	// Compile sessions stay unrestricted: module source may emit any op.
	// Runtime denial is the client Session's SR + exec / viral flags.
	compileSess, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		panic(fmt.Errorf("install: NewSession: %w", err))
	}
	abandoned := true
	defer func() {
		if abandoned {
			if mod := ModuleTable.moduleByID(compileSess.task.compileModID); mod != nil {
				ModuleTable.abandonModule(mod)
			}
		}
		_ = compileSess.Close()
	}()

	task := compileSess.task
	rp := compileSess.rp
	rp.WithTaskFindScope(task)
	w := globalDictionaryContext.dict.GetByName("compileModule")
	if w == nil {
		panic("install: missing compileModule")
	}
	rp.DataStack.Push(String(root))
	rp.DataStack.Push(String(jsonText))
	if err := runWordInlineExec(task, rp, w); err != nil {
		panic(fmt.Errorf("install: compileModule: %w", err))
	}

	mod := ModuleTable.moduleByID(task.compileModID)
	if mod == nil || mod.dc == nil {
		panic(fmt.Errorf("install: compile Session module id %d missing", task.compileModID))
	}
	mod.dc.Seal()
	prefix := moduleDefaultPrefix(wire.Name)
	ModuleTable.finishInstall(mod, wire.Name, wire.Version, prefix)
	abandoned = false
}
