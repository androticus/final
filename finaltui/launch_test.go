package finaltui

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/androticus/final/finalembed"
)

func TestClassifyLaunch(t *testing.T) {
	cases := []struct {
		name         string
		tty          bool
		flags        LaunchFlags
		positionals  []string
		wantMode     LaunchMode
		wantStay     bool
		wantParams   string
		wantPath     string
		wantHeadless bool
		wantErr      string
	}{
		{
			name:     "tty interactive tui",
			tty:      true,
			wantMode: LaunchInteractive,
		},
		{
			name:         "tty interactive headless",
			tty:          true,
			flags:        LaunchFlags{Headless: true},
			wantMode:     LaunchInteractive,
			wantHeadless: true,
		},
		{
			name:        "tty file",
			tty:         true,
			positionals: []string{"script.fnl", "a", "b"},
			wantMode:    LaunchFileScript,
			wantPath:    "script.fnl",
			wantParams:  "script.fnl a b",
		},
		{
			name:        "tty file -i myscript foo bar",
			tty:         true,
			flags:       LaunchFlags{Stay: true, Headless: true},
			positionals: []string{"myscript.fnl", "foo", "bar"},
			wantMode:    LaunchFileScript,
			wantStay:    true,
			wantPath:    "myscript.fnl",
			wantParams:  "myscript.fnl foo bar",
		},
		{
			name:        "tty explicit dash",
			tty:         true,
			positionals: []string{"-", "foo"},
			wantMode:    LaunchStdinScript,
			wantParams:  "- foo",
		},
		{
			name:        "pipe synthesizes dash",
			positionals: []string{"suspect.fnl"},
			wantMode:    LaunchStdinScript,
			wantParams:  "- suspect.fnl",
		},
		{
			name:        "pipe typed dash not doubled",
			positionals: []string{"-", "foo", "bar"},
			wantMode:    LaunchStdinScript,
			wantParams:  "- foo bar",
		},
		{
			name:       "pipe no positionals",
			wantMode:   LaunchStdinScript,
			wantParams: "-",
		},
		{
			name:    "pipe -i error",
			flags:   LaunchFlags{Stay: true},
			wantErr: "-i cannot be used with stdin",
		},
		{
			name:        "tty -i with dash error",
			tty:         true,
			flags:       LaunchFlags{Stay: true},
			positionals: []string{"-"},
			wantErr:     "-i cannot be used with stdin",
		},
		{
			name:    "tty -i no file error",
			tty:     true,
			flags:   LaunchFlags{Stay: true},
			wantErr: "-i requires a script file",
		},
		{
			name:        "pipe -m error",
			flags:       LaunchFlags{Manifest: "app.fnlm"},
			positionals: []string{"x"},
			wantErr:     "-m cannot be combined",
		},
		{
			name:        "tty -m with file error",
			tty:         true,
			flags:       LaunchFlags{Manifest: "app.fnlm"},
			positionals: []string{"script.fnl"},
			wantErr:     "-m cannot be combined",
		},
		{
			name:     "tty -m interactive ok",
			tty:      true,
			flags:    LaunchFlags{Manifest: "app.fnlm"},
			wantMode: LaunchInteractive,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, err := ClassifyLaunch(tc.tty, tc.flags, tc.positionals)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if spec.Mode != tc.wantMode {
				t.Fatalf("mode = %v, want %v", spec.Mode, tc.wantMode)
			}
			if spec.Stay != tc.wantStay {
				t.Fatalf("stay = %v, want %v", spec.Stay, tc.wantStay)
			}
			if spec.Headless != tc.wantHeadless {
				t.Fatalf("headless = %v, want %v", spec.Headless, tc.wantHeadless)
			}
			if spec.ScriptPath != tc.wantPath {
				t.Fatalf("path = %q, want %q", spec.ScriptPath, tc.wantPath)
			}
			got := strings.Join(spec.Params, " ")
			if got != tc.wantParams {
				t.Fatalf("params = %q, want %q", got, tc.wantParams)
			}
		})
	}
}

func sendSysJSON(t *testing.T, script string) []string {
	t.Helper()
	sess, err := finalembed.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	out, err := sess.GetSender().Send(script, 0)
	if err != nil {
		t.Fatalf("Send(%q): %v", script, err)
	}
	var got []string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &got); err != nil {
		t.Fatalf("json %q: %v", out, err)
	}
	return got
}

func TestApplySysLaunchPipeDoesNotSynthesizeArgvDash(t *testing.T) {
	t.Cleanup(func() { ApplySysLaunch(nil, LaunchSpec{}) })
	spec, err := ClassifyLaunch(false, LaunchFlags{}, []string{"suspect.fnl"})
	if err != nil {
		t.Fatal(err)
	}
	argv := []string{"final", "suspect.fnl"}
	ApplySysLaunch(argv, spec)
	if got := sendSysJSON(t, "sys.params >json $"); strings.Join(got, " ") != "- suspect.fnl" {
		t.Fatalf("params = %v", got)
	}
	if got := sendSysJSON(t, "sys.argv >json $"); strings.Join(got, " ") != "final suspect.fnl" {
		t.Fatalf("argv = %v, want no synthesized dash", got)
	}
}

func TestApplySysLaunchTypedDashOnPipe(t *testing.T) {
	t.Cleanup(func() { ApplySysLaunch(nil, LaunchSpec{}) })
	spec, err := ClassifyLaunch(false, LaunchFlags{}, []string{"-", "foo", "bar"})
	if err != nil {
		t.Fatal(err)
	}
	argv := []string{"final", "-", "foo", "bar"}
	ApplySysLaunch(argv, spec)
	if got := sendSysJSON(t, "sys.params >json $"); strings.Join(got, " ") != "- foo bar" {
		t.Fatalf("params = %v", got)
	}
	if got := sendSysJSON(t, "sys.argv >json $"); strings.Join(got, " ") != "final - foo bar" {
		t.Fatalf("argv = %v", got)
	}
}

func TestApplySysLaunchTTYFileIKeepsFlagsInArgvOnly(t *testing.T) {
	t.Cleanup(func() { ApplySysLaunch(nil, LaunchSpec{}) })
	spec, err := ClassifyLaunch(true, LaunchFlags{Headless: true, Stay: true}, []string{"myscript.fnl", "foo", "bar"})
	if err != nil {
		t.Fatal(err)
	}
	argv := []string{"final", "--headless", "-i", "myscript.fnl", "foo", "bar"}
	ApplySysLaunch(argv, spec)
	if got := sendSysJSON(t, "sys.params >json $"); strings.Join(got, " ") != "myscript.fnl foo bar" {
		t.Fatalf("params = %v", got)
	}
	if got := sendSysJSON(t, "sys.argv >json $"); strings.Join(got, " ") != strings.Join(argv, " ") {
		t.Fatalf("argv = %v, want full os.Args", got)
	}
}
