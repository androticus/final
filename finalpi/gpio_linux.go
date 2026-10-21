//go:build linux

package finalpi

import (
	"fmt"
	"os"
	"strings"

	"github.com/warthog618/go-gpiocdev"
)

// openGpioLine requests one gpiochip line with direction + pull (activate backend).
func openGpioLine(bcm int, dir GpioDirection, pull GpioPull, outValue bool) (*Pin, error) {
	chip, err := resolveChipName()
	if err != nil {
		return nil, err
	}
	opts := []gpiocdev.LineReqOption{
		gpiocdev.WithConsumer("final"),
	}
	switch pull {
	case GpioPullUp:
		opts = append(opts, gpiocdev.WithPullUp)
	case GpioPullDown:
		opts = append(opts, gpiocdev.WithPullDown)
	default:
		opts = append(opts, gpiocdev.WithBiasDisabled)
	}
	if dir == GpioOut {
		v := 0
		if outValue {
			v = 1
		}
		opts = append(opts, gpiocdev.AsOutput(v))
	} else {
		opts = append(opts, gpiocdev.AsInput)
	}
	l, err := gpiocdev.RequestLine(chip, bcm, opts...)
	if err != nil {
		return nil, fmt.Errorf("request %s line %d: %w (need gpio group / udev?)", chip, bcm, err)
	}
	level := 0
	if outValue {
		level = 1
	}
	return &Pin{bcm: bcm, output: dir == GpioOut, level: level, line: l}, nil
}

func lineAsInput(l line, pull GpioPull) error {
	gl, ok := l.(*gpiocdev.Line)
	if !ok {
		return nil
	}
	opts := []gpiocdev.LineConfigOption{gpiocdev.AsInput}
	switch pull {
	case GpioPullUp:
		opts = append(opts, gpiocdev.WithPullUp)
	case GpioPullDown:
		opts = append(opts, gpiocdev.WithPullDown)
	default:
		opts = append(opts, gpiocdev.WithBiasDisabled)
	}
	return gl.Reconfigure(opts...)
}

// resolveChipName picks the user-facing gpiochip.
// Override with FINAL_GPIOCHIP (e.g. gpiochip0 or gpiochip4) when auto-detect is wrong.
// Pi 5: pinctrl-rp1 only; Pi Zero 2: pinctrl-bcm2835 only; else gpiochip0 / first chip.
func resolveChipName() (string, error) {
	if c := strings.TrimSpace(os.Getenv("FINAL_GPIOCHIP")); c != "" {
		return c, nil
	}
	preferred := preferredGpioLabels(piSupport)
	chips := gpiocdev.Chips()
	for _, want := range preferred {
		for _, name := range chips {
			c, err := gpiocdev.NewChip(name)
			if err != nil {
				continue
			}
			label := c.Label
			_ = c.Close()
			if label == want {
				return name, nil
			}
		}
	}
	for _, name := range chips {
		if name == "gpiochip0" {
			return name, nil
		}
	}
	if len(chips) > 0 {
		return chips[0], nil
	}
	return "", fmt.Errorf("no /dev/gpiochip* found")
}
