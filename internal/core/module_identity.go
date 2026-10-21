package core

import (
	"strings"
)

// ParseModuleRef splits a Go-style module reference: name or name@version.
// Trimmed. Empty version (omit @, or trailing @) means "latest".
// Phase 2 carries version but does not select among versions.
func ParseModuleRef(ref string) (name, version string) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", ""
	}
	name, ver, ok := strings.Cut(ref, "@")
	name = strings.TrimSpace(name)
	if !ok {
		return name, ""
	}
	return name, strings.TrimSpace(ver)
}

// FormatModuleRef builds name or name@version (no @ when version is empty).
func FormatModuleRef(name, version string) string {
	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)
	if version == "" {
		return name
	}
	return name + "@" + version
}

// moduleDefaultPrefix is the last path segment of a module identity name.
func moduleDefaultPrefix(name string) string {
	name = strings.TrimSpace(name)
	name = strings.TrimRight(name, "/")
	if name == "" {
		return ""
	}
	if i := strings.LastIndex(name, "/"); i >= 0 {
		return name[i+1:]
	}
	return name
}
