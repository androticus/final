package finalgui

import "testing"

func TestClampScale(t *testing.T) {
	if got := clampScale(0.5); got != 0.75 {
		t.Fatalf("clampScale(0.5) = %v, want 0.75", got)
	}
	if got := clampScale(3); got != 2.5 {
		t.Fatalf("clampScale(3) = %v, want 2.5", got)
	}
	if got := clampScale(1.5); got != 1.5 {
		t.Fatalf("clampScale(1.5) = %v, want 1.5", got)
	}
}

func TestNearestScalePreset(t *testing.T) {
	if got := nearestScalePreset(1.1); got != 1.0 {
		t.Fatalf("nearestScalePreset(1.1) = %v, want 1", got)
	}
	if got := nearestScalePreset(1.9); got != 2.0 {
		t.Fatalf("nearestScalePreset(1.9) = %v, want 2", got)
	}
}

func TestFormatParseScale(t *testing.T) {
	if got := formatScale(2); got != "2" {
		t.Fatalf("formatScale(2) = %q, want 2", got)
	}
	if got := formatScale(0.75); got != "0.75" {
		t.Fatalf("formatScale(0.75) = %q, want 0.75", got)
	}
	v, ok := parseScaleLabel("2")
	if !ok || v != 2 {
		t.Fatalf("parseScaleLabel(2) = %v %v", v, ok)
	}
}

func TestScaleLooksAnomalous(t *testing.T) {
	// Pi-like 4K, scale 1, default chrome ~720px wide
	if !scaleLooksAnomalous(1.0, 3840, 720) {
		t.Fatal("expected anomalous on 4K at scale 1")
	}
	if scaleLooksAnomalous(2.0, 3840, 1440) {
		t.Fatal("scale 2 on 4K should not invite")
	}
	// Ordinary 1080p: window not tiny enough, width < 3000
	if scaleLooksAnomalous(1.0, 1920, 720) {
		t.Fatal("1920x with 720 window (~37%) should not invite")
	}
	// Small absolute monitor but window still a small fraction
	if !scaleLooksAnomalous(1.0, 1280, 200) {
		t.Fatal("expected anomalous when window < 25% of monitor")
	}
}
