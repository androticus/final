package core

import (
	"fmt"
	"io/fs"
	"iter"
	"os"
	"path/filepath"
)

// Walk flag bits for Dir.iter / FileSystemIter (combinable with or / |).
// Final names: Dir.IterDirs, Dir.IterFiles, Dir.IterDeep, Dir.IterAll.
const (
	DirIterDirs  = 1 << iota // include directories
	DirIterFiles             // include files (non-dirs)
	DirIterDeep              // recursive depth-first
	DirIterAll   = DirIterDirs | DirIterFiles | DirIterDeep
)

// Dir is a directory root path (not an OS open-directory handle).
type Dir struct {
	path string
}

func NewDir(path string) *Dir {
	return &Dir{path: path}
}

func (d *Dir) Path() string {
	return d.path
}

func (d *Dir) FinalType() *Type {
	return GetTypeRegistrar().Get("Dir")
}

// FileSystemIter is a pull iterator of *FileInfo under a Dir root.
// Technique: Go iter.Pull over a push seq (see docs/005-background/05-go-path-iter.md).
type FileSystemIter struct {
	next    func() (*FileInfo, bool)
	stop    func()
	current *FileInfo
	ok      bool
}

// NewFileSystemIter creates a pull-style iterator rooted at dir.
// dir must name a directory; panics otherwise.
func NewFileSystemIter(dir *Dir, flags int) *FileSystemIter {
	if dir == nil {
		panic(fmt.Errorf("FileSystemIter: nil Dir"))
	}
	root := dir.path
	info, err := os.Stat(root)
	if err != nil {
		panic(fmt.Errorf("FileSystemIter: stat %q: %w", root, err))
	}
	if !info.IsDir() {
		panic(fmt.Errorf("FileSystemIter: %q is not a directory", root))
	}

	seq := makeFileSystemWalkSeq(root, flags)
	next, stop := iter.Pull(seq)
	return &FileSystemIter{
		next: next,
		stop: stop,
	}
}

func (it *FileSystemIter) GetNext() bool {
	if it.next == nil {
		return false
	}
	it.current, it.ok = it.next()
	if !it.ok {
		it.Close()
	}
	return it.ok
}

func (it *FileSystemIter) Current() any {
	return it.current
}

func (it *FileSystemIter) Close() {
	if it.stop != nil {
		it.stop()
		it.stop = nil
		it.next = nil
	}
}

func makeFileSystemWalkSeq(root string, flags int) iter.Seq[*FileInfo] {
	return func(yield func(*FileInfo) bool) {
		wantDirs := flags&DirIterDirs != 0
		wantFiles := flags&DirIterFiles != 0
		deep := flags&DirIterDeep != 0

		if !deep {
			entries, err := os.ReadDir(root)
			if err != nil {
				return
			}
			for _, e := range entries {
				if shouldYieldDirEntry(e, wantDirs, wantFiles) {
					if !yield(fileInfoFromDirEntry(root, e)) {
						return
					}
				}
			}
			return
		}

		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // skip; continue walk
			}
			if path == root {
				return nil // do not yield the root
			}
			if shouldYieldDirEntry(d, wantDirs, wantFiles) {
				parent := filepath.Dir(path)
				if !yield(fileInfoFromDirEntry(parent, d)) {
					return filepath.SkipAll
				}
			}
			return nil
		})
	}
}

func shouldYieldDirEntry(d fs.DirEntry, wantDirs, wantFiles bool) bool {
	if d.IsDir() {
		return wantDirs
	}
	return wantFiles
}

func registerDirMembers(t *Type) {
	t.registerClassProperty("IterDirs", "Int", func(_ Any) Any { return Int(DirIterDirs) }, nil)
	t.registerClassProperty("IterFiles", "Int", func(_ Any) Any { return Int(DirIterFiles) }, nil)
	t.registerClassProperty("IterDeep", "Int", func(_ Any) Any { return Int(DirIterDeep) }, nil)
	t.registerClassProperty("IterAll", "Int", func(_ Any) Any { return Int(DirIterAll) }, nil)

	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(NewDir(path))
	})

	// ( -- dir:Dir )
	t.registerNativeClassMethodSec("cwd", func(f *RunFrame, _ *TypeObjPair) {
		wd, err := os.Getwd()
		if err != nil {
			panic(fmt.Errorf("Dir.cwd: %w", err))
		}
		f.RP.DataStack.Push(NewDir(wd))
	}, SR_FileSystemRead)

	// ( s -- true | err:s false )
	t.registerNativeClassMethodSec("setCwd", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		pushStreamWriteResult(f, os.Chdir(path))
	}, SR_FileSystemRead)

	// ( s -- b )
	t.registerNativeClassMethodSec("exists?", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(Bool(FilePathExists(path)))
	}, SR_FileSystemRead)

	// ( s -- fi:FileInfo true | err:s false )
	t.registerNativeClassMethodSec("info", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		fi, err := fileInfoFromPath(path)
		pushStreamOpenResult(f, fi, err)
	}, SR_FileSystemRead)

	// ( s -- true | err:s false ) — MkdirAll; idempotent if already a directory
	t.registerNativeClassMethodSec("mkdir", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		pushStreamWriteResult(f, dirMkdirAll(path))
	}, SR_FileWrite)

	t.registerProperty("string", "String",
		func(inst Any) Any { return String(inst.(*Dir).path) },
		nil,
	)

	t.registerNativeMethodSec("iter", func(f *RunFrame, ths *TypeObjPair) {
		flags := int(f.RP.DataStack.Pop().(Int))
		dir := ths.Obj.(*Dir)
		f.RP.DataStack.Push(NewFileSystemIter(dir, flags))
	}, SR_FileSystemRead)
}

// dirMkdirAll creates path and parents. Already a directory → nil. Exists as file → error.
func dirMkdirAll(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return nil
		}
		return fmt.Errorf("Dir.mkdir: %q exists and is not a directory", path)
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("Dir.mkdir: %w", err)
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("Dir.mkdir: %w", err)
	}
	return nil
}

func CreateRegisterDirType() {
	t := &Type{
		name:            "Dir",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Dir]("Dir"),
	}
	registerDirMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
}

func init() {
	CreateRegisterDirType()
}

var (
	_ Typed    = (*Dir)(nil)
	_ Iterator = (*FileSystemIter)(nil)
)
