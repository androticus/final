package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Path is a class-only catalog type for path-string algebra (no instances).
// Final: Path.join^, Path.split^, Path.clean^, Path.abs^, Path.basename^,
// Path.dirname^, Path.extname^. See docs/003-design-notes/156-file-info-iter.md.
type Path struct{}

var pathType *Type

func CreateRegisterPathType() {
	t := &Type{
		name:    "Path",
		getZero: ReferenceTypeZeroFn,
		checkCoerceType: func(value any) any {
			panic(fmt.Errorf("Path.checkCoerceType: Path is a class-only type (got %T)", value))
		},
	}
	registerPathMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	pathType = t
}

func registerPathMembers(t *Type) {
	// ( base:s name:s -- s )
	t.registerNativeClassMethod("join", func(f *RunFrame, _ *TypeObjPair) {
		name := string(f.RP.DataStack.Pop().(String))
		base := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(String(filepath.Join(base, name)))
	})

	// ( s -- list:StringList ) — segment split of one path (not PATH-env SplitList)
	t.registerNativeClassMethod("split", func(f *RunFrame, _ *TypeObjPair) {
		s := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(pathSplit(s))
	})

	// ( s -- s )
	t.registerNativeClassMethod("clean", func(f *RunFrame, _ *TypeObjPair) {
		s := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(String(filepath.Clean(s)))
	})

	// ( s -- s ) — may read cwd
	t.registerNativeClassMethodSec("abs", func(f *RunFrame, _ *TypeObjPair) {
		s := string(f.RP.DataStack.Pop().(String))
		abs, err := filepath.Abs(s)
		if err != nil {
			panic(fmt.Errorf("Path.abs: %w", err))
		}
		f.RP.DataStack.Push(String(abs))
	}, SR_FileSystemRead)

	// ( s -- s ) — leaf without extension (see pathBaseExt)
	t.registerNativeClassMethod("basename", func(f *RunFrame, _ *TypeObjPair) {
		s := string(f.RP.DataStack.Pop().(String))
		base, _ := pathBaseExt(s)
		f.RP.DataStack.Push(String(base))
	})

	// ( s -- s )
	t.registerNativeClassMethod("dirname", func(f *RunFrame, _ *TypeObjPair) {
		s := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(String(filepath.Dir(s)))
	})

	// ( s -- s ) — including leading ".", or ""; mirrors basename split
	t.registerNativeClassMethod("extname", func(f *RunFrame, _ *TypeObjPair) {
		s := string(f.RP.DataStack.Pop().(String))
		_, ext := pathBaseExt(s)
		if ext == "" {
			f.RP.DataStack.Push(String(""))
			return
		}
		f.RP.DataStack.Push(String("." + ext))
	})
}

func init() {
	CreateRegisterPathType()
}

// pathBaseExt strips any directory prefix, then splits the leaf into base + ext
// (ext without the dot). Leading-dot names (e.g. ".gitignore") are entirely base;
// an extension is never the basename. Same rules as File.getBaseExt (design 156).
func pathBaseExt(s string) (base, ext string) {
	leaf := filepath.Base(s)
	if leaf == "." || leaf == ".." || leaf == string(os.PathSeparator) {
		return leaf, ""
	}
	i := strings.LastIndex(leaf, ".")
	if i <= 0 {
		// no dot, or only a leading dot (dotfile)
		return leaf, ""
	}
	return leaf[:i], leaf[i+1:]
}

// pathSplit splits a single path into components (PurePath-style parts).
// Absolute Unix: first element is "/". Absolute Windows: volume+sep (e.g. `C:\`).
// Relative: components only. Empty input → empty list.
func pathSplit(s string) *StringList {
	out := NewStringList()
	if s == "" {
		return out
	}
	clean := filepath.Clean(s)
	vol := filepath.VolumeName(clean)
	rest := clean[len(vol):]
	sep := string(os.PathSeparator)

	isAbs := strings.HasPrefix(rest, sep)
	if vol != "" || isAbs {
		root := vol
		if isAbs {
			root = vol + sep
			rest = strings.TrimPrefix(rest, sep)
		}
		if root != "" {
			out.Append(root)
		}
	}

	if rest == "" || rest == "." {
		if out.Len() == 0 {
			out.Append(".")
		}
		return out
	}

	for _, part := range strings.Split(rest, sep) {
		if part != "" && part != "." {
			out.Append(part)
		}
	}
	return out
}
