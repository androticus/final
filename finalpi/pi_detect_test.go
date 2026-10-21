package finalpi

import (
	"os"
	"testing"
)

func TestMapDeviceTreeModel(t *testing.T) {
	cases := []struct {
		model string
		want  PiSupportType
	}{
		{"Raspberry Pi 5 Model B Rev 1.1", PiSupportPi5},
		{"Raspberry Pi 5 Model B Rev 1.0\x00", PiSupportPi5},
		{"Raspberry Pi Zero 2 W Rev 1.0", PiSupportPiZ2},
		{"Raspberry Pi Zero 2 W Rev 1.0\x00", PiSupportPiZ2},
		{"Raspberry Pi 4 Model B Rev 1.5", PiSupportNone},
		{"Raspberry Pi 3 Model B Plus Rev 1.3", PiSupportNone},
		{"", PiSupportNone},
		{"Some other board", PiSupportNone},
	}
	for _, tc := range cases {
		if got := mapDeviceTreeModel(tc.model); got != tc.want {
			t.Errorf("mapDeviceTreeModel(%q)=%v want %v", tc.model, got, tc.want)
		}
	}
}

func TestDetectPiSupportEnvOverride(t *testing.T) {
	cases := []struct {
		env   string
		model string
		want  PiSupportType
	}{
		{"pi5", "Raspberry Pi Zero 2 W Rev 1.0", PiSupportPi5},
		{"pi2z", "Raspberry Pi 5 Model B Rev 1.1", PiSupportPiZ2},
		{"none", "Raspberry Pi 5 Model B Rev 1.1", PiSupportNone},
		{"PI5", "", PiSupportPi5},
		{"", "Raspberry Pi 5 Model B Rev 1.1", PiSupportPi5},
		{"", "Raspberry Pi Zero 2 W Rev 1.0", PiSupportPiZ2},
		{"bogus", "Raspberry Pi 5 Model B Rev 1.1", PiSupportPi5}, // invalid env → model
	}
	for _, tc := range cases {
		if got := detectPiSupport(tc.model, tc.env); got != tc.want {
			t.Errorf("detectPiSupport(%q, %q)=%v want %v", tc.model, tc.env, got, tc.want)
		}
	}
}

func TestInitPiSupportMatchesHostOrEnv(t *testing.T) {
	// Package init already ran; compare to current env + live/stub model.
	want := detectPiSupport(readDeviceTreeModel(), os.Getenv(envPiSupport))
	if piSupport != want {
		t.Fatalf("package piSupport=%v want %v (model=%q env=%q)",
			piSupport, want, readDeviceTreeModel(), os.Getenv(envPiSupport))
	}
}

func TestUart0DefaultPath(t *testing.T) {
	if got := uart0DefaultPath(PiSupportPiZ2); got != "/dev/ttyS0" {
		t.Fatalf("pi2z path: %s", got)
	}
	if got := uart0DefaultPath(PiSupportPi5); got != "/dev/ttyAMA0" {
		t.Fatalf("pi5 path: %s", got)
	}
	if got := uart0DefaultPath(PiSupportNone); got != "/dev/ttyAMA0" {
		t.Fatalf("none path: %s", got)
	}
}

func TestRejectUnsupportedUartBaudMiniUART(t *testing.T) {
	if err := rejectUnsupportedUartBaud("/dev/ttyS0", 250); err == nil {
		t.Fatal("expected reject baud 250 on ttyS0")
	}
	if err := rejectUnsupportedUartBaud("/dev/ttyS0", 1200); err != nil {
		t.Fatalf("1200 should be ok: %v", err)
	}
	if err := rejectUnsupportedUartBaud("/dev/ttyAMA0", 250); err != nil {
		t.Fatalf("PL011 250 should be ok: %v", err)
	}
}
