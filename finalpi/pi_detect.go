package finalpi

import (
	"os"
	"strings"
)

const envPiSupport = "FINAL_PI_SUPPORT"

// detectPiSupport maps a device-tree model string and optional FINAL_PI_SUPPORT
// override to a PiSupportType. Env wins when set to none|pi2z|pi5 (case-insensitive).
func detectPiSupport(model, envOverride string) PiSupportType {
	if s, ok := parsePiSupportOverride(envOverride); ok {
		return s
	}
	return mapDeviceTreeModel(model)
}

func parsePiSupportOverride(env string) (PiSupportType, bool) {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "":
		return 0, false
	case "none":
		return PiSupportNone, true
	case "pi2z":
		return PiSupportPiZ2, true
	case "pi5":
		return PiSupportPi5, true
	default:
		return 0, false
	}
}

// mapDeviceTreeModel recognizes known Raspberry Pi board model strings.
// Unknown / non-Pi models → PiSupportNone.
func mapDeviceTreeModel(model string) PiSupportType {
	m := strings.TrimSpace(strings.TrimRight(model, "\x00"))
	switch {
	case strings.Contains(m, "Raspberry Pi 5"):
		return PiSupportPi5
	case strings.Contains(m, "Raspberry Pi Zero 2"):
		return PiSupportPiZ2
	default:
		return PiSupportNone
	}
}

// initPiSupport sets the package-global piSupport from device-tree + env.
// Must run at package init (before gpio.fnl reads GpioCtrl.piSupport), not in newGpioCtrl.
func initPiSupport() {
	piSupport = detectPiSupport(readDeviceTreeModel(), os.Getenv(envPiSupport))
}
