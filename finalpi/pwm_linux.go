//go:build linux

package finalpi

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// BCM header PWM → channel index on the RP1 PWM0 block (pwm@98000).
// pwm0→12/ch0, pwm1→13/ch1 (stock pwm-2chan pinmux).
var bcmPWMChannel = map[int]int{
	12: 0,
	13: 1,
}

type pwmHandle struct {
	chipDir string // /sys/class/pwm/pwmchipN
	channel int
	pwmDir  string // .../pwmM after export
}

// openPWM exports the sysfs channel for bcm. Caller must Close.
// Pi 5: RP1 pwm@98000 (pwm-2chan on BCM 12/13). Pi Zero 2: first non-fan pwmchip*.
// Override: FINAL_PWMCHIP.
func openPWM(bcm int) (*pwmHandle, error) {
	ch, ok := bcmPWMChannel[bcm]
	if !ok {
		return nil, fmt.Errorf("PWM: BCM %d is not a header PWM pin (12/13)", bcm)
	}
	chipDir, err := resolvePWMChipDir()
	if err != nil {
		return nil, err
	}
	h := &pwmHandle{chipDir: chipDir, channel: ch}
	if err := h.export(); err != nil {
		return nil, err
	}
	return h, nil
}

func resolvePWMChipDir() (string, error) {
	if c := strings.TrimSpace(os.Getenv("FINAL_PWMCHIP")); c != "" {
		dir := filepath.Join("/sys/class/pwm", c)
		if st, err := os.Stat(dir); err != nil || !st.IsDir() {
			return "", fmt.Errorf("FINAL_PWMCHIP=%s not found under /sys/class/pwm", c)
		}
		return dir, nil
	}
	entries, err := os.ReadDir("/sys/class/pwm")
	if err != nil {
		return "", fmt.Errorf("PWM: no /sys/class/pwm (%w) — add dtoverlay=pwm-2chan,pin=12,pin2=13,func=4,func2=4?", err)
	}
	var cands []pwmChipCand
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "pwmchip") {
			continue
		}
		dir := filepath.Join("/sys/class/pwm", name)
		target, err := filepath.EvalSymlinks(dir)
		if err != nil {
			continue
		}
		cands = append(cands, pwmChipCand{dir: dir, target: target})
	}
	return pickPWMChipDir(piSupport, cands)
}

func (h *pwmHandle) export() error {
	pwmDir := filepath.Join(h.chipDir, "pwm"+strconv.Itoa(h.channel))
	if st, err := os.Stat(pwmDir); err == nil && st.IsDir() {
		h.pwmDir = pwmDir
		return h.waitWritable()
	}
	if err := os.WriteFile(filepath.Join(h.chipDir, "export"), []byte(strconv.Itoa(h.channel)), 0); err != nil {
		return fmt.Errorf("PWM export %s channel %d: %w (gpio group? overlay?)", h.chipDir, h.channel, err)
	}
	// sysfs may appear shortly after export; Pi OS udev then chgrp gpio (99-com.rules).
	deadline := time.Now().Add(2 * time.Second)
	for {
		if st, err := os.Stat(pwmDir); err == nil && st.IsDir() {
			h.pwmDir = pwmDir
			if err := h.waitWritableUntil(deadline); err != nil {
				_ = h.Close()
				return err
			}
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("PWM: %s did not appear after export", pwmDir)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (h *pwmHandle) waitWritable() error {
	return h.waitWritableUntil(time.Now().Add(2 * time.Second))
}

func (h *pwmHandle) waitWritableUntil(deadline time.Time) error {
	periodPath := filepath.Join(h.pwmDir, "period")
	for {
		f, err := os.OpenFile(periodPath, os.O_WRONLY, 0)
		if err == nil {
			_ = f.Close()
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("PWM: %s not writable after export (gpio group / udev 99-com.rules?): %w", periodPath, err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func (h *pwmHandle) writeAttr(name, val string) error {
	path := filepath.Join(h.pwmDir, name)
	if err := os.WriteFile(path, []byte(val), 0); err != nil {
		return fmt.Errorf("PWM write %s: %w", path, err)
	}
	return nil
}

func (h *pwmHandle) readAttr(name string) (string, error) {
	b, err := os.ReadFile(filepath.Join(h.pwmDir, name))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// applyPeriodDuty programs period/duty and leaves sysfs enable=1.
// Duty 0 idles without a second Final property.
func (h *pwmHandle) applyPeriodDuty(periodNs, dutyNs uint64) error {
	if periodNs == 0 {
		return fmt.Errorf("PWM: periodNs must be > 0 before activate")
	}
	if dutyNs > periodNs {
		return fmt.Errorf("PWM: dutyCycleNs %d > periodNs %d", dutyNs, periodNs)
	}
	// Safe order when shrinking period: clear duty, set period, restore duty, enable.
	if err := h.writeAttr("enable", "0"); err != nil {
		// ignore if already disabled / freshly exported
		_ = err
	}
	if err := h.writeAttr("duty_cycle", "0"); err != nil {
		return err
	}
	if err := h.writeAttr("period", strconv.FormatUint(periodNs, 10)); err != nil {
		return err
	}
	if err := h.writeAttr("duty_cycle", strconv.FormatUint(dutyNs, 10)); err != nil {
		return err
	}
	if err := h.writeAttr("enable", "1"); err != nil {
		return err
	}
	return nil
}

func (h *pwmHandle) setDuty(dutyNs uint64) error {
	return h.writeAttr("duty_cycle", strconv.FormatUint(dutyNs, 10))
}

func (h *pwmHandle) Close() error {
	if h == nil || h.pwmDir == "" {
		return nil
	}
	_ = h.writeAttr("enable", "0")
	_ = h.writeAttr("duty_cycle", "0")
	err := os.WriteFile(filepath.Join(h.chipDir, "unexport"), []byte(strconv.Itoa(h.channel)), 0)
	h.pwmDir = ""
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("PWM unexport: %w", err)
	}
	return nil
}
