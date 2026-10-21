package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
	bootstrapRootOnce sync.Once
	bootstrapRootDir  string

	cmdRootOnce   sync.Once
	cmdRootDir    string
	cmdRootWarned bool
)

// bootstrapRoot returns the module root that holds the bootstrap Final sources
// (internal/final/...). It is located from this file's compiled source path, so
// manifest paths resolve regardless of the process working directory: go test
// runs with the working dir set to each package's directory, and finalgo /
// cmd-final-ai trigger bootstrap at import time (before any TestMain can chdir).
// Returns "" if the root cannot be located; callers then try cmdRoot / cwd.
func bootstrapRoot() string {
	bootstrapRootOnce.Do(func() {
		_, file, _, ok := runtime.Caller(0)
		if !ok {
			return
		}
		dir := filepath.Dir(file)
		for {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				bootstrapRootDir = dir
				return
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				return
			}
			dir = parent
		}
	})
	return bootstrapRootDir
}

// cmdRoot returns the install / executable directory: Dir(EvalSymlinks(Executable())).
// Used when bootstrapRoot is unavailable (release packs without go.mod beside the binary).
// Returns "" if the executable path cannot be resolved.
func cmdRoot() string {
	cmdRootOnce.Do(func() {
		exe, err := os.Executable()
		if err != nil {
			return
		}
		resolved, err := filepath.EvalSymlinks(exe)
		if err != nil {
			resolved = exe
		}
		cmdRootDir = filepath.Dir(resolved)
	})
	return cmdRootDir
}

// resolveCompilePath maps a compile/manifest path to a filesystem path.
// Absolute paths are unchanged (then cleaned). Relative paths:
//  1. cmdRoot when it looks like a Final install/unpack (finaltui/final or internal/final)
//  2. else bootstrapRoot (module root via runtime.Caller — dev / go test)
//  3. else cmdRoot
//  4. else process cwd (one-shot stderr warning)
//
// Preferring an install-shaped cmdRoot avoids release binaries still resolving through
// a sibling source clone when Caller embeds the build-machine module path.
// No ~ expansion.
func resolveCompilePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("resolveCompilePath: empty path")
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	if root := cmdRoot(); root != "" && installLayoutPresent(root) {
		return filepath.Join(root, path), nil
	}
	if root := bootstrapRoot(); root != "" {
		return filepath.Join(root, path), nil
	}
	if root := cmdRoot(); root != "" {
		return filepath.Join(root, path), nil
	}
	if !cmdRootWarned {
		cmdRootWarned = true
		fmt.Fprintln(os.Stderr, "warning: compile path root unavailable (no go.mod / executable); using process working directory")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return abs, nil
}

// installLayoutPresent reports whether root looks like a Final install or unpack tree
// (sources next to the binary), as opposed to a bare go-build cache dir.
func installLayoutPresent(root string) bool {
	if root == "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(root, "finaltui", "final")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(root, "internal", "final")); err == nil {
		return true
	}
	return false
}

// RunBootstrapInit compiles core0, core1, and core2 into global GCD, then core3
// when EnableBootstrapTranche3 was called (fyneui binaries), then core4 when
// EnableBootstrapTranche4 was called (finalpi). Called once from
// installGlobalCoreVocabulary during InitGlobalExecutionEnvironment.
func RunBootstrapInit(t *Task, rp *RunProvider) {
	if err := compileBootstrapSource(t, rp, bootstrapInitScript0); err != nil {
		bootstrapWriteREPLError(rp, err)
		panic(err)
	}
	if err := compileBootstrapSource(t, rp, bootstrapInitScript1); err != nil {
		bootstrapWriteREPLError(rp, err)
		panic(err)
	}
	if err := runBootstrapTranche2(t, rp); err != nil {
		panic(err)
	}
	if tranche3Enabled {
		if err := runBootstrapTranche3(t, rp); err != nil {
			panic(err)
		}
	}
	if tranche4Enabled {
		if err := runBootstrapTranche4(t, rp); err != nil {
			panic(err)
		}
	}
}

// runBootstrapCoreCompile compiles core0/core1/core2 (and core3/4 if enabled)
// with a throwaway Task/RunProvider.
func runBootstrapCoreCompile() {
	t, _ := NewTask(false)
	rp := NewRunProvider()
	stderr, err := NewStderrTextWriter()
	if err != nil {
		panic(err)
	}
	rp.Stderr = stderr
	RunBootstrapInit(t, rp)
}

// parseBootstrapManifestLines returns repo-root-relative paths from a tranche manifest string.
// Blank lines and lines whose trimmed content starts with // are ignored.
func parseBootstrapManifestLines(manifest string) []string {
	var paths []string
	for _, line := range strings.Split(manifest, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		paths = append(paths, line)
	}
	return paths
}

func runBootstrapTranche2(t *Task, rp *RunProvider) error {
	return runBootstrapTranche2Manifest(t, rp, bootstrapTranche2Manifest)
}

func runBootstrapTranche3(t *Task, rp *RunProvider) error {
	return runBootstrapTranche2Manifest(t, rp, bootstrapTranche3Manifest)
}

func runBootstrapTranche4(t *Task, rp *RunProvider) error {
	return runBootstrapTranche2Manifest(t, rp, bootstrapTranche4Manifest)
}

func runBootstrapTranche2Manifest(t *Task, rp *RunProvider, manifest string) error {
	if err := compileManifestContent(t, rp, manifest); err != nil {
		bootstrapWriteREPLError(rp, err)
		return err
	}
	return nil
}

// compileManifestContent compiles each non-comment / non-blank path line in manifest text.
// Paths are resolved via resolveCompilePath (bootstrap/cmd root) — used by
// embedded tranche manifests. Disk .fnlm files use compileBootstrapManifestFile instead.
func compileManifestContent(t *Task, rp *RunProvider, manifest string) error {
	return compileManifestContentRelativeTo(t, rp, manifest, "")
}

// compileManifestContentRelativeTo is like compileManifestContent, but when baseDir is
// non-empty, relative path lines are joined to that directory and opened directly
// (movable project collections: paths relative to the .fnlm). They must not go through
// resolveCompilePath again — that would re-apply bootstrap/cmd root and can concatenate
// the module path twice.
func compileManifestContentRelativeTo(t *Task, rp *RunProvider, manifest, baseDir string) error {
	for _, path := range parseBootstrapManifestLines(manifest) {
		if baseDir != "" {
			filePath := path
			if !filepath.IsAbs(path) {
				filePath = filepath.Join(baseDir, path)
			}
			filePath = filepath.Clean(filePath)
			if err := compileBootstrapFileAt(t, rp, filePath); err != nil {
				return fmt.Errorf("%s: %w", filePath, err)
			}
			continue
		}
		if err := compileBootstrapFile(t, rp, path); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}
	return nil
}

// compileBootstrapManifestFile reads a .fnlm (or other) manifest file and compiles listed sources.
// Relative paths inside the file are resolved against the manifest's directory.
func compileBootstrapManifestFile(t *Task, rp *RunProvider, path string) error {
	resolved, err := resolveCompilePath(path)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return err
	}
	return compileManifestContentRelativeTo(t, rp, string(data), filepath.Dir(resolved))
}

// compileWithTransaction wraps one compile unit in begin/endTransaction (all-or-nothing).
func compileWithTransaction(t *Task, rp *RunProvider, run func() error) (err error) {
	if t == nil || rp == nil {
		panic("compileWithTransaction: nil Task or RunProvider")
	}
	cs := t.compilerForTarget()
	cs.beginTransaction()
	defer func() {
		if r := recover(); r != nil {
			rbErr := cs.tryRollbackTransaction()
			compileErr := fmt.Errorf("compile panic: %v", r)
			if rp.Parser != nil {
				compileErr = locateFault(compileErr, rp.Parser)
				drainParserToEOF(rp.Parser)
			}
			if rbErr != nil {
				err = fmt.Errorf("%s; %v", compileErr, rbErr)
				return
			}
			err = compileErr
		}
	}()
	if err = run(); err != nil {
		rbErr := cs.tryRollbackTransaction()
		err = locateFault(err, rp.Parser)
		drainParserToEOF(rp.Parser)
		if rbErr != nil {
			return fmt.Errorf("%w; %v", err, rbErr)
		}
		return err
	}
	cs.endTransaction(true)
	return nil
}

// withSystemMode runs fn with t.systemMode true (core/bootstrap compile).
func withSystemMode(t *Task, fn func() error) error {
	if t == nil {
		return fn()
	}
	prev := t.systemMode
	t.systemMode = true
	defer func() { t.systemMode = prev }()
	return fn()
}

// compileBootstrapSource parses and compiles one Final source string into global GCD.
func compileBootstrapSource(t *Task, rp *RunProvider, script string) error {
	script = strings.TrimSpace(script)
	if script == "" {
		return nil
	}
	if t == nil || rp == nil {
		panic("compileBootstrapSource: nil Task or RunProvider")
	}
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader(script)),
		label: "bootstrap-compile",
	}})
	return withSystemMode(t, func() error {
		return nestParser(rp, p, func() error {
			return compileWithTransaction(t, rp, func() error {
				return interpretSession(t, rp)
			})
		})
	})
}

// compileBootstrapFile parses and compiles one Final source file into global GCD.
// Relative paths go through resolveCompilePath. The compile unit is wrapped in
// compileWithTransaction (rollback on error).
func compileBootstrapFile(t *Task, rp *RunProvider, path string) error {
	resolved, err := resolveCompilePath(path)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return compileBootstrapFileAt(t, rp, resolved)
}

// compileBootstrapFileAt compiles an already-resolved filesystem path (no root join).
// If line 1 is exactly "forget: name" and that interpret-only mark exists, it is
// run before beginTransaction so a reload can recreate the mark.
func compileBootstrapFileAt(t *Task, rp *RunProvider, resolved string) error {
	resolved = filepath.Clean(resolved)
	absPath, err := filepath.Abs(resolved)
	if err != nil {
		return fmt.Errorf("%s: %w", resolved, err)
	}
	if err := maybeRunLeadingForgetMark(t, rp, resolved); err != nil {
		return err
	}
	tr, err := NewFileTextReader(resolved)
	if err != nil {
		return err
	}
	defer tr.Close()
	p := NewParserFromTextRead(tr)
	p.setParserLocation(absPath, 1, 1)
	return nestParser(rp, p, func() error {
		return compileWithTransaction(t, rp, func() error {
			return interpretSession(t, rp)
		})
	})
}

// compileBootstrapReader parses and compiles one Final source from r into global GCD.
func compileBootstrapReader(t *Task, rp *RunProvider, r io.Reader, label string) error {
	if r == nil {
		return fmt.Errorf("compileBootstrapReader: nil reader")
	}
	if label == "" {
		label = "bootstrap-compile"
	}
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(r),
		label: label,
	}})
	return nestParser(rp, p, func() error {
		return compileWithTransaction(t, rp, func() error {
			return interpretSession(t, rp)
		})
	})
}

func runBootstrapInitScript(t *Task, rp *RunProvider, script string) {
	if err := compileBootstrapSource(t, rp, script); err != nil {
		panic(err)
	}
}
