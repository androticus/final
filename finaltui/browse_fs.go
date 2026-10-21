package finaltui

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// DirListing is one on-demand ReadDir of a single directory (no recursion).
type DirListing struct {
	Dir     string   // absolute cleaned path
	Folders []string // base names only, sorted
	Files   []string // base names only, sorted
	Err     error    // set if ReadDir failed; Folders/Files may be empty
}

// ListDir reads only dir. Hides Unix dotfiles (names starting with '.').
// On Windows, when atRoots is true, Folders are drive roots and Files is empty.
func ListDir(dir string) DirListing {
	dir = filepath.Clean(dir)
	out := DirListing{Dir: dir}
	entries, err := os.ReadDir(dir)
	if err != nil {
		out.Err = err
		return out
	}
	for _, e := range entries {
		name := e.Name()
		if name == "." || name == ".." {
			continue
		}
		if strings.HasPrefix(name, ".") {
			continue
		}
		if e.IsDir() {
			out.Folders = append(out.Folders, name)
		} else {
			out.Files = append(out.Files, name)
		}
	}
	sort.Strings(out.Folders)
	sort.Strings(out.Files)
	return out
}

// ParentDir returns the parent of dir, or dir itself at filesystem root.
func ParentDir(dir string) string {
	dir = filepath.Clean(dir)
	parent := filepath.Dir(dir)
	if parent == dir {
		return dir
	}
	return parent
}

// IsFSRoot reports whether dir is a filesystem root (no further ..).
func IsFSRoot(dir string) bool {
	dir = filepath.Clean(dir)
	return filepath.Dir(dir) == dir
}

// WindowsDriveRoots returns existing drive letter roots (e.g. C:\). Empty on non-Windows.
func WindowsDriveRoots() []string {
	if runtime.GOOS != "windows" {
		return nil
	}
	var roots []string
	for c := 'A'; c <= 'Z'; c++ {
		p := string(c) + `:\`
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			roots = append(roots, p)
		}
	}
	return roots
}

// JoinUnder joins base and name safely (name must be a single path element).
func JoinUnder(base, name string) string {
	return filepath.Join(filepath.Clean(base), name)
}

// ResolveBrowsePath interprets a path field commit: returns a directory to list
// and an optional file basename to highlight in that directory.
func ResolveBrowsePath(path string) (dir string, highlightFile string, err error) {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "\r\n")
	if len(path) >= 2 {
		if (path[0] == '"' && path[len(path)-1] == '"') ||
			(path[0] == '\'' && path[len(path)-1] == '\'') {
			path = strings.TrimSpace(path[1 : len(path)-1])
		}
	}
	if path == "" {
		return "", "", os.ErrInvalid
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", "", err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return "", "", err
	}
	if st.IsDir() {
		return abs, "", nil
	}
	return filepath.Dir(abs), filepath.Base(abs), nil
}
