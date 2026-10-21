package finaltui

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/androticus/final/finalembed"
)

// HostOptions controls post-Init loads and which UI loop to run.
type HostOptions struct {
	Headless     bool   // true → stdin/stdout REPL; false → tview
	UserManifest string // optional -m path (.fnl / .fnlm / .fnlp); empty → skip
	PlatformQual string // optional qualifier under GOOS (e.g. "pi" → extras/platform/linux/pi/)
}

// PlatformManifestPath returns the built-in .fnlm path for this host (slash-separated).
// Qualifier "pi" is always under linux/pi (Pi foundation), even when the binary is
// built/run on another GOOS for bring-up.
func PlatformManifestPath(headless bool, qual string) string {
	name := "tview.fnlm"
	if headless {
		name = "headless.fnlm"
	}
	goos := runtime.GOOS
	if qual == "pi" {
		goos = "linux"
	}
	parts := []string{"extras", "platform", goos}
	if qual != "" {
		parts = append(parts, qual)
	}
	parts = append(parts, name)
	return filepath.ToSlash(filepath.Join(parts...))
}

// compilePathScript builds a Final Send script from path suffix:
// .fnlm → compileFromManifest, .fnlp → loadProject, else → compileFromFile.
func compilePathScript(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("empty path")
	}
	if strings.ContainsRune(path, '`') {
		return "", fmt.Errorf("path contains backtick; cannot wrap in Final raw string")
	}
	path = filepath.ToSlash(path)
	// Explorer / pasted Windows paths may keep '\' even on non-Windows hosts.
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

// LoadPlatformManifest compiles the built-in platform .fnlm for this GOOS (+ optional qualifier).
func LoadPlatformManifest(sess *finalembed.Session, headless bool, qual string) error {
	path := PlatformManifestPath(headless, qual)
	script, err := compilePathScript(path)
	if err != nil {
		return err
	}
	if _, err := sess.GetSender().Send(script, defaultSendTimeoutS); err != nil {
		return fmt.Errorf("platform manifest %s: %w", path, err)
	}
	return nil
}

// LoadUserPath loads a user -m path by suffix (.fnlm / .fnlp / .fnl).
// Non-empty captured stdout from the load (e.g. smoke SMOKE_OK via $) is echoed to
// process stdout — Send itself does not print; without this, -m loads look silent.
func LoadUserPath(sess *finalembed.Session, path string) error {
	script, err := compilePathScript(path)
	if err != nil {
		return err
	}
	out, err := sess.GetSender().Send(script, defaultSendTimeoutS)
	if err != nil {
		return fmt.Errorf("user load %s: %w", path, err)
	}
	if out != "" {
		fmt.Fprint(os.Stdout, out)
	}
	return nil
}

// LoadHostManifests compiles platform .fnlm then optional user -m. Call after Init.
func LoadHostManifests(sess *finalembed.Session, opts HostOptions) error {
	if sess == nil {
		return fmt.Errorf("finaltui.LoadHostManifests: nil session")
	}
	if err := LoadPlatformManifest(sess, opts.Headless, opts.PlatformQual); err != nil {
		return err
	}
	if opts.UserManifest != "" {
		if err := LoadUserPath(sess, opts.UserManifest); err != nil {
			return err
		}
	}
	return nil
}

// RunHost loads platform (+ optional -m) manifests, then tview or stdin REPL.
// Call after Init.
func RunHost(sess *finalembed.Session, opts HostOptions) error {
	if err := LoadHostManifests(sess, opts); err != nil {
		return err
	}
	if opts.Headless {
		return sess.RunREPL()
	}
	return RunTUI(sess)
}
