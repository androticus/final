package core

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

/*
Provides information about a File or Directory.
The information is obtained from the operating system.
Path is the parent directory of the entry (no trailing separator
except for a filesystem root). Name is the final path element.
Path.join(Path, Name) is the full endpoint. Size / ModTime /
IsDir / IsSymlink are host metadata. Base/ext are not fields —
use File.getBaseExt on Name (see design note 156).
*/
type FileInfo struct {
	Path      string
	Name      string
	Size      int64
	ModTime   time.Time
	IsDir     bool
	IsSymlink bool
}

func NewFileInfo(fi os.FileInfo) *FileInfo {
	return &FileInfo{
		Name:      fi.Name(),
		Size:      fi.Size(),
		ModTime:   fi.ModTime(),
		IsDir:     fi.IsDir(),
		IsSymlink: (fi.Mode() & os.ModeSymlink) != 0,
	}
}

// fileInfoFromPath builds FileInfo for an endpoint path (Lstat: isSymlink is about the path itself).
// Path = parent directory; Name = final element.
func fileInfoFromPath(endpoint string) (*FileInfo, error) {
	fi, err := os.Lstat(endpoint)
	if err != nil {
		return nil, err
	}
	return &FileInfo{
		Path:      filepath.Dir(endpoint),
		Name:      filepath.Base(endpoint),
		Size:      fi.Size(),
		ModTime:   fi.ModTime(),
		IsDir:     fi.IsDir(),
		IsSymlink: fi.Mode()&os.ModeSymlink != 0,
	}, nil
}

// fileInfoFromDirEntry builds FileInfo for an entry under parentDir.
// parentDir is the directory containing the entry (FileInfo.Path).
func fileInfoFromDirEntry(parentDir string, e fs.DirEntry) *FileInfo {
	out := &FileInfo{
		Path:      parentDir,
		Name:      e.Name(),
		IsDir:     e.IsDir(),
		IsSymlink: e.Type()&fs.ModeSymlink != 0,
	}
	if info, err := e.Info(); err == nil {
		out.Size = info.Size()
		out.ModTime = info.ModTime()
	}
	return out
}

func (fi *FileInfo) FinalType() *Type {
	return GetTypeRegistrar().Get("FileInfo")
}

func registerFileInfoMembers(t *Type) {
	t.registerProperty("path", "String",
		func(inst Any) Any { return String(inst.(*FileInfo).Path) },
		nil,
	)
	t.registerProperty("name", "String",
		func(inst Any) Any { return String(inst.(*FileInfo).Name) },
		nil,
	)
	t.registerProperty("size", "Int",
		func(inst Any) Any { return Int(inst.(*FileInfo).Size) },
		nil,
	)
	// added for Final-32 compatibility
	t.registerProperty("sizeHigh", "Uint",
		func(inst Any) Any { return Uint(uint32(inst.(*FileInfo).Size >> 32)) },
		nil,
	)
	// added for Final-32 compatibility
	t.registerProperty("sizeLow", "Uint",
		func(inst Any) Any { return Uint(uint32(inst.(*FileInfo).Size & 0xffffffff)) },
		nil,
	)
	t.registerProperty("modTime", "DateTime",
		func(inst Any) Any { return inst.(*FileInfo).ModTime },
		nil,
	)
	t.registerProperty("isDir", "Bool",
		func(inst Any) Any { return Bool(inst.(*FileInfo).IsDir) },
		nil,
	)
	t.registerProperty("isSymlink", "Bool",
		func(inst Any) Any { return Bool(inst.(*FileInfo).IsSymlink) },
		nil,
	)
}

func CreateRegisterFileInfoType() {
	t := &Type{
		name:            "FileInfo",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[FileInfo]("FileInfo"),
	}
	registerFileInfoMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
}

func init() {
	CreateRegisterFileInfoType()
}

var _ Typed = (*FileInfo)(nil)
