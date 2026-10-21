package finalpi

import (
	"strings"
	"testing"
)

func TestPickPWMChipDirPi5Requires98000(t *testing.T) {
	cands := []pwmChipCand{
		{dir: "/sys/class/pwm/pwmchip0", target: "/sys/devices/.../fan/pwm"},
		{dir: "/sys/class/pwm/pwmchip1", target: "/sys/devices/.../pwm@9c000"},
		{dir: "/sys/class/pwm/pwmchip2", target: "/sys/devices/.../pwm@98000"},
	}
	got, err := pickPWMChipDir(PiSupportPi5, cands)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/sys/class/pwm/pwmchip2" {
		t.Fatalf("got %s want pwmchip2 (98000)", got)
	}
	_, err = pickPWMChipDir(PiSupportPi5, cands[:2])
	if err == nil || !strings.Contains(err.Error(), "98000") {
		t.Fatalf("expected 98000 missing error, got %v", err)
	}
}

func TestPickPWMChipDirPi2ZFirstNonFan(t *testing.T) {
	cands := []pwmChipCand{
		{dir: "/sys/class/pwm/pwmchip0", target: "/sys/devices/.../hwmon/fan/pwm"},
		{dir: "/sys/class/pwm/pwmchip1", target: "/sys/devices/platform/soc/fe20c000.pwm/pwm/pwmchip1"},
		{dir: "/sys/class/pwm/pwmchip2", target: "/sys/devices/.../pwm@98000"},
	}
	got, err := pickPWMChipDir(PiSupportPiZ2, cands)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/sys/class/pwm/pwmchip1" {
		t.Fatalf("got %s want first non-fan (pwmchip1)", got)
	}
}

func TestPreferredGpioLabelsModal(t *testing.T) {
	z2 := preferredGpioLabels(PiSupportPiZ2)
	if len(z2) != 1 || z2[0] != "pinctrl-bcm2835" {
		t.Fatalf("Pi2Z labels=%v want [pinctrl-bcm2835]", z2)
	}
	pi5 := preferredGpioLabels(PiSupportPi5)
	if len(pi5) != 1 || pi5[0] != "pinctrl-rp1" {
		t.Fatalf("Pi5 labels=%v want [pinctrl-rp1]", pi5)
	}
	if got := preferredGpioLabels(PiSupportNone); len(got) != 0 {
		t.Fatalf("None labels=%v want empty", got)
	}
}
