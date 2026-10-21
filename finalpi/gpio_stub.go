//go:build !linux

package finalpi

import "fmt"

func openGpioLine(bcm int, dir GpioDirection, pull GpioPull, outValue bool) (*Pin, error) {
	_ = dir
	_ = pull
	_ = outValue
	return nil, fmt.Errorf("GPIO gpio%d requires Linux (this binary has no gpiochip backend)", bcm)
}

func lineAsInput(l line, pull GpioPull) error {
	_ = l
	_ = pull
	return nil
}
