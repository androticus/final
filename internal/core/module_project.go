package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const projectFileExt = ".fnlp"

// projectManifestWire is the on-disk .fnlp JSON shape (same family as _MODULE_).
type projectManifestWire struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Final       string            `json:"final"` // optional; unchecked until semver
	Version     string            `json:"version"`
	Requires    []json.RawMessage `json:"requires"`
	Files       []string          `json:"files"`
}

// projectRequireWire is one requires[] entry (version ignored for selection).
type projectRequireWire struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Prefix  string `json:"prefix"`
}

func readProjectFile(path string) (wire projectManifestWire, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return wire, fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		return wire, fmt.Errorf("parse %s: %w", path, err)
	}
	wire.Name = strings.TrimSpace(wire.Name)
	wire.Version = strings.TrimSpace(wire.Version)
	wire.Final = strings.TrimSpace(wire.Final)
	return wire, nil
}

func vetProjectManifest(wire *projectManifestWire) error {
	if wire.Name == "" {
		return fmt.Errorf("loadProject: project missing name")
	}
	if wire.Version == "" {
		return fmt.Errorf("loadProject: project missing version")
	}
	// final optional; no version check this release
	for i, f := range wire.Files {
		f = strings.TrimSpace(f)
		if f == "" {
			return fmt.Errorf("loadProject: empty path in files[%d]", i)
		}
		wire.Files[i] = f
	}
	return nil
}

func parseProjectRequire(raw json.RawMessage) (projectRequireWire, error) {
	var req projectRequireWire
	if err := json.Unmarshal(raw, &req); err != nil {
		return req, fmt.Errorf("loadProject: requires entry: %w", err)
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Version = strings.TrimSpace(req.Version)
	req.Prefix = strings.TrimSpace(req.Prefix)
	if req.Name == "" {
		return req, fmt.Errorf("loadProject: requires entry missing name")
	}
	return req, nil
}

// resolveProjectFilePath joins projectDir with a project files entry.
// Absolute entries stay absolute; relative may use .. above the project directory.
func resolveProjectFilePath(projectDir, entry string) string {
	entry = strings.TrimSpace(entry)
	if filepath.IsAbs(entry) {
		return filepath.Clean(entry)
	}
	return filepath.Clean(filepath.Join(projectDir, entry))
}

// runLoadProject is OpLoadProject: ( projectPath:s -- )
// projectPath is the path to a .fnlp file; relative files resolve from its directory.
func runLoadProject(t *Task, rp *RunProvider) {
	assertCallerDCOpen(t, "loadProject")
	if t.session == nil || t.session.modules == nil {
		panic("loadProject: Task has no Session.modules")
	}
	pathArg := string(rp.DataStack.Pop().(String))
	loadProject(t, rp, pathArg)
}

// loadProject installs+imports requires, then compiles project files into the current Local Session.
func loadProject(t *Task, rp *RunProvider, pathArg string) {
	pathArg = strings.TrimSpace(pathArg)
	if pathArg == "" {
		panic("loadProject: empty project path")
	}
	if !strings.EqualFold(filepath.Ext(pathArg), projectFileExt) {
		panic(fmt.Errorf("loadProject: path must end in %s, got %q", projectFileExt, pathArg))
	}
	st, err := os.Stat(pathArg)
	if err != nil || st.IsDir() {
		panic(fmt.Errorf("loadProject: %q is not a project file", pathArg))
	}
	projectDir := filepath.Dir(pathArg)
	wire, err := readProjectFile(pathArg)
	if err != nil {
		panic(err)
	}
	if err := vetProjectManifest(&wire); err != nil {
		panic(err)
	}

	emptyOpts := NewDynamicObject()
	for _, raw := range wire.Requires {
		req, err := parseProjectRequire(raw)
		if err != nil {
			panic(err)
		}
		installLeafModule(t, req.Name, emptyOpts)
		identity, _ := ParseModuleRef(req.Name)
		importModule(t, req.Prefix, identity)
	}

	compileW := globalDictionaryContext.dict.GetByName("compileFromFile")
	if compileW == nil {
		panic("loadProject: missing compileFromFile")
	}
	for _, entry := range wire.Files {
		path := resolveProjectFilePath(projectDir, entry)
		if st, err := os.Stat(path); err != nil || st.IsDir() {
			panic(fmt.Errorf("loadProject: project file %q not found", path))
		}
		rp.DataStack.Push(String(path))
		if err := runWordInlineExec(t, rp, compileW); err != nil {
			panic(fmt.Errorf("loadProject: compileFromFile %s: %w", path, err))
		}
		if rp.Stdout != nil {
			rp.Stdout.WriteString("Compiled " + path + ".\n")
		}
	}
	if rp.Stdout != nil {
		rp.Stdout.WriteString("zzPROJ_DONE_zz\n")
	}
}
