package finaltui

import (
	"fmt"
	"os"

	"golang.org/x/term"

	"github.com/androticus/final/finalembed"
)

// LaunchMode is the product CLI mode after interpreter flags are parsed.
type LaunchMode int

const (
	LaunchInteractive LaunchMode = iota
	LaunchStdinScript
	LaunchFileScript
)

// LaunchFlags are interpreter flags consumed by Go flag (not placed in sys.params).
type LaunchFlags struct {
	Headless bool
	Stay     bool   // -i
	Manifest string // -m / --manifest
}

// LaunchSpec is the classified product launch (table-testable).
type LaunchSpec struct {
	Mode       LaunchMode
	Headless   bool
	Stay       bool
	Manifest   string
	ScriptPath string
	Params     []string
}

// StdinIsTTY reports whether process stdin is a terminal.
func StdinIsTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// ClassifyLaunch decides stdin-script vs script-file vs interactive from TTY and positionals.
func ClassifyLaunch(stdinIsTTY bool, flags LaunchFlags, positionals []string) (LaunchSpec, error) {
	first, hasFirst := "", false
	if len(positionals) > 0 {
		first = positionals[0]
		hasFirst = true
	}

	if !stdinIsTTY {
		if flags.Stay {
			return LaunchSpec{}, fmt.Errorf("final: -i cannot be used with stdin (pipe, redirect, or -)")
		}
		if flags.Manifest != "" {
			return LaunchSpec{}, fmt.Errorf("final: -m cannot be combined with a script file or stdin")
		}
		params := append([]string(nil), positionals...)
		if !hasFirst || first != "-" {
			params = append([]string{"-"}, params...)
		}
		return LaunchSpec{Mode: LaunchStdinScript, Params: params}, nil
	}

	if hasFirst && first != "-" {
		if flags.Manifest != "" {
			return LaunchSpec{}, fmt.Errorf("final: -m cannot be combined with a script file or stdin")
		}
		return LaunchSpec{
			Mode:       LaunchFileScript,
			Stay:       flags.Stay,
			ScriptPath: first,
			Params:     append([]string(nil), positionals...),
		}, nil
	}

	if hasFirst && first == "-" {
		if flags.Stay {
			return LaunchSpec{}, fmt.Errorf("final: -i cannot be used with stdin (pipe, redirect, or -)")
		}
		if flags.Manifest != "" {
			return LaunchSpec{}, fmt.Errorf("final: -m cannot be combined with a script file or stdin")
		}
		return LaunchSpec{Mode: LaunchStdinScript, Params: append([]string(nil), positionals...)}, nil
	}

	if flags.Stay {
		return LaunchSpec{}, fmt.Errorf("final: -i requires a script file on a terminal stdin")
	}
	return LaunchSpec{
		Mode:     LaunchInteractive,
		Headless: flags.Headless,
		Manifest: flags.Manifest,
	}, nil
}

// ApplySysLaunch writes argv (literal os.Args) and classified params onto the process Sys.
func ApplySysLaunch(argv []string, spec LaunchSpec) {
	finalembed.SetSysLaunch(argv, spec.Params)
}

// RunLaunch runs the classified mode. Call after Init. host.PlatformQual is used only
// for interactive RunHost (Pi packs).
func RunLaunch(sess *finalembed.Session, spec LaunchSpec, host HostOptions) error {
	switch spec.Mode {
	case LaunchStdinScript:
		return sess.InterpretReader(os.Stdin)
	case LaunchFileScript:
		f, err := os.Open(spec.ScriptPath)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := sess.InterpretReaderPath(f, spec.ScriptPath); err != nil {
			return err
		}
		if spec.Stay {
			return sess.RunREPL()
		}
		return nil
	default:
		host.Headless = spec.Headless
		host.UserManifest = spec.Manifest
		return RunHost(sess, host)
	}
}
