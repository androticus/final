package finalpi

// preferredGpioLabels is the label match order for the header gpiochip.
// Only the label known for that profile — no cross-family speculative fallbacks.
func preferredGpioLabels(support PiSupportType) []string {
	switch support {
	case PiSupportPiZ2:
		return []string{"pinctrl-bcm2835"}
	case PiSupportPi5:
		return []string{"pinctrl-rp1"}
	default:
		return nil
	}
}
