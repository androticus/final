package finalpi

import (
	"fmt"
	"strings"
)

// pwmChipCand is one /sys/class/pwm/pwmchip* entry after EvalSymlinks.
type pwmChipCand struct {
	dir    string // e.g. /sys/class/pwm/pwmchip0
	target string // resolved symlink target
}

// pickPWMChipDir chooses a chip by platform profile (package piSupport).
// Pi 5 requires RP1 pwm@98000; Pi Zero 2 takes the first non-fan/cooling chip.
func pickPWMChipDir(support PiSupportType, cands []pwmChipCand) (string, error) {
	var nonFan []pwmChipCand
	for _, c := range cands {
		low := strings.ToLower(c.target)
		if strings.Contains(low, "fan") || strings.Contains(low, "cooling") {
			continue
		}
		nonFan = append(nonFan, c)
	}
	if support == PiSupportPi5 {
		for _, c := range nonFan {
			// RP1 PWM0 at 0x98000 — the block pwm-2chan uses for BCM 12/13.
			if strings.Contains(c.target, "98000") {
				return c.dir, nil
			}
		}
		return "", fmt.Errorf("PWM: RP1 pwm@98000 not found under /sys/class/pwm (need dtoverlay=pwm-2chan,pin=12,pin2=13,func=4,func2=4; not using pwm@9c000)")
	}
	// PiSupportPiZ2 (and None): classic pwm-2chan / first usable chip.
	if len(nonFan) > 0 {
		return nonFan[0].dir, nil
	}
	return "", fmt.Errorf("PWM: no usable pwmchip* under /sys/class/pwm (need dtoverlay=pwm-2chan,pin=12,pin2=13,func=4,func2=4?)")
}
