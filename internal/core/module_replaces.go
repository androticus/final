package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const replacesFileName = "_REPLACES_"

// replacesMap is identity (trimmed) → local filesystem root (trimmed).
type replacesMap map[string]string

// findReplacesFile walks from cwd upward looking for _REPLACES_.
// Returns absolute path to the file, or "" if none.
func findReplacesFile() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		p := filepath.Join(dir, replacesFileName)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// loadReplacesMap reads and parses _REPLACES_ JSON object (string→string).
// Missing file → empty map (not an error).
func loadReplacesMap() (replacesMap, string, error) {
	path := findReplacesFile()
	if path == "" {
		return replacesMap{}, "", nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("load _REPLACES_: %w", err)
	}
	var wire map[string]string
	if err := json.Unmarshal(raw, &wire); err != nil {
		return nil, "", fmt.Errorf("parse _REPLACES_ %s: %w", path, err)
	}
	out := make(replacesMap, len(wire))
	for k, v := range wire {
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k == "" || v == "" {
			continue
		}
		out[k] = v
	}
	return out, path, nil
}

// resolveModuleRoot picks the filesystem root for install.
// Order: opts "path" override → _REPLACES_[identity] → identity as path (cwd-relative or abs).
// Relative replace targets are joined to the directory containing _REPLACES_.
func resolveModuleRoot(identity string, opts *DynamicObject) (root string, viaReplace bool, err error) {
	identity = strings.TrimSpace(identity)
	if identity == "" {
		return "", false, fmt.Errorf("resolve module root: empty identity")
	}
	if opts != nil && bool(opts.HasField("path")) {
		p := moduleFileAsString(opts.Get("path"))
		p = strings.TrimSpace(p)
		if p == "" {
			return "", false, fmt.Errorf("resolve module root: options.path is empty")
		}
		return p, false, nil
	}
	rep, repPath, err := loadReplacesMap()
	if err != nil {
		return "", false, err
	}
	if mapped, ok := rep[identity]; ok {
		if filepath.IsAbs(mapped) {
			return mapped, true, nil
		}
		base := filepath.Dir(repPath)
		return filepath.Join(base, mapped), true, nil
	}
	return identity, false, nil
}
