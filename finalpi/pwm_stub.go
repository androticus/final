//go:build !linux

package finalpi

import "fmt"

type pwmHandle struct{}

func openPWM(bcm int) (*pwmHandle, error) {
	return nil, fmt.Errorf("PWM bcm%d requires Linux (this binary has no pwmchip sysfs backend)", bcm)
}

func (h *pwmHandle) applyPeriodDuty(periodNs, dutyNs uint64) error {
	_ = periodNs
	_ = dutyNs
	return fmt.Errorf("PWM: no sysfs backend")
}

func (h *pwmHandle) setDuty(dutyNs uint64) error {
	_ = dutyNs
	return fmt.Errorf("PWM: no sysfs backend")
}

func (h *pwmHandle) Close() error { return nil }

func (h *pwmHandle) readAttr(name string) (string, error) {
	_ = name
	return "", fmt.Errorf("PWM: no sysfs backend")
}
