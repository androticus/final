package finalgui

import (
	"fmt"
	"path/filepath"
	"strings"
)

// compilePathScript builds a Final interpret script from path suffix:
// .fnlm → compileFromManifest, .fnlp → loadProject, else → compileFromFile.
// Same rule as finaltui (host.compilePathScript).
func compilePathScript(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("empty path")
	}
	if strings.ContainsRune(path, '`') {
		return "", fmt.Errorf("path contains backtick; cannot wrap in Final raw string")
	}
	path = filepath.ToSlash(path)
	path = strings.ReplaceAll(path, "\\", "/")
	word := "compileFromFile"
	switch strings.ToLower(filepath.Ext(path)) {
	case ".fnlm":
		word = "compileFromManifest"
	case ".fnlp":
		word = "loadProject"
	}
	return "`" + path + "` " + word, nil
}
