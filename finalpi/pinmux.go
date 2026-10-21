package finalpi

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/androticus/final/internal/core"
)

// osPinMapLen is StringList length for BCM indices 0..27 inclusive.
const osPinMapLen = 28

var (
	pinmuxLineRE     = regexp.MustCompile(`^pin\s+(\d+)\s+\(gpio\d+\):\s*(.*)$`)
	pinmuxFunctionRE = regexp.MustCompile(`\bfunction\s+(\S+)`)
	pinmuxDeviceRE   = regexp.MustCompile(`\bdevice\s+(\S+)`)
	// pins file (sibling of pinmux-pins): "pin 14 (gpio14) 0:? function alt5 in hi; …"
	pinctrlPinsLineRE = regexp.MustCompile(`^pin\s+(\d+)\s+\(gpio\d+\)\s+.*\bfunction\s+(\S+)`)
)

// pinmuxPinsRead reads pinmux-pins bytes. Tests may override.
var pinmuxPinsRead = defaultPinmuxPinsRead

// pinctrlPinsRead reads the pinctrl "pins" dump (function/alt). Tests may override.
// Used to fill slots left empty by pinmux-pins (Z2 mini-UART: UNCLAIMED but alt5).
var pinctrlPinsRead = defaultPinctrlPinsRead

func defaultPinmuxPinsRead() ([]byte, error) {
	path, err := findPinmuxPinsPath()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

func defaultPinctrlPinsRead() ([]byte, error) {
	path, err := findPinmuxPinsPath()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(filepath.Dir(path), "pins"))
}

func findPinmuxPinsPath() (string, error) {
	const root = "/sys/kernel/debug/pinctrl"
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	labels := preferredGpioLabels(piSupport)
	if len(labels) == 0 {
		labels = []string{"pinctrl-rp1", "pinctrl-bcm2835"}
	}
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		name := ent.Name()
		for _, lab := range labels {
			if !strings.Contains(name, lab) {
				continue
			}
			p := filepath.Join(root, name, "pinmux-pins")
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p, nil
			}
		}
	}
	return "", fmt.Errorf("pinmux-pins not found under %s for %v", root, labels)
}

// parsePinmuxPins fills labels[0:osPinMapLen] from pinmux-pins text.
// Unmentioned / unclaimed / gpiochip-only BCMs stay "" (Reset does not skip those).
func parsePinmuxPins(data []byte, labels []string) {
	for i := range labels {
		labels[i] = ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		m := pinmuxLineRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		bcm, err := strconv.Atoi(m[1])
		if err != nil || bcm < 0 || bcm >= len(labels) {
			continue
		}
		labels[bcm] = pinmuxLabelFromRest(m[2])
	}
}

func pinmuxLabelFromRest(rest string) string {
	rest = strings.TrimSpace(rest)
	if rest == "" || rest == "UNCLAIMED" || strings.HasPrefix(rest, "(MUX UNCLAIMED)") {
		return ""
	}
	// bcm2835: "GPIO pinctrl-bcm2835:518" — gpiochip hold (Final/user), not a bus overlay.
	if strings.HasPrefix(rest, "GPIO ") || strings.HasPrefix(rest, "gpio ") {
		return ""
	}

	device := ""
	if dm := pinmuxDeviceRE.FindStringSubmatch(rest); dm != nil {
		device = dm[1]
	} else {
		// RP1: "1f00074000.i2c (GPIO UNCLAIMED) function i2c1 …"
		tok := rest
		if i := strings.IndexAny(rest, " \t"); i >= 0 {
			tok = rest[:i]
		}
		if strings.Contains(tok, ".") && !strings.HasPrefix(tok, "(") {
			device = tok
		}
	}

	function := ""
	if fm := pinmuxFunctionRE.FindStringSubmatch(rest); fm != nil {
		function = fm[1]
	}

	return pickPinmuxLabel(device, function)
}

func pickPinmuxLabel(device, function string) string {
	// RP1-style meaningful function names beat altN / gpio_*.
	if function != "" && !isAltFunction(function) && !isGPIOFunction(function) {
		return function
	}
	if device != "" {
		return friendlyDeviceLabel(device)
	}
	if function != "" && !isGPIOFunction(function) {
		return function
	}
	return ""
}

func isAltFunction(fn string) bool {
	if len(fn) < 4 || !strings.HasPrefix(fn, "alt") {
		return false
	}
	for _, c := range fn[3:] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func isGPIOFunction(fn string) bool {
	return fn == "gpio" || strings.HasPrefix(fn, "gpio_")
}

// mergeAltFunctionLabels fills empty label slots from the pinctrl "pins" dump.
// Classic bcm2835 mini-UART (enable_uart=1 + BT on PL011) leaves gpio14/15 as
// UNCLAIMED in pinmux-pins while "pins" shows function alt5 — without this,
// pinMap misses a working uart. Never surfaces bare altN (148.5).
func mergeAltFunctionLabels(pinsData []byte, labels []string) {
	for _, line := range strings.Split(string(pinsData), "\n") {
		line = strings.TrimSpace(line)
		m := pinctrlPinsLineRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		bcm, err := strconv.Atoi(m[1])
		if err != nil || bcm < 0 || bcm >= len(labels) {
			continue
		}
		if labels[bcm] != "" {
			continue
		}
		if lab := labelFromBCMAlt(bcm, m[2]); lab != "" {
			labels[bcm] = lab
		}
	}
}

// labelFromBCMAlt maps known BCM + altN muxes to a friendly periph label.
// Returns "" for gpio / unknown alts (do not emit bare "alt5").
func labelFromBCMAlt(bcm int, fn string) string {
	if !isAltFunction(fn) {
		return ""
	}
	switch bcm {
	case 14, 15:
		// ALT0 = PL011 TX/RX; ALT5 = mini UART (ttyS0) on Zero 2 header.
		if fn == "alt0" || fn == "alt5" {
			return "uart"
		}
	case 16, 17:
		// Optional UART CTS/RTS (ALT3 PL011 / ALT5 mini UART).
		if fn == "alt3" || fn == "alt5" {
			return "uart"
		}
	}
	return ""
}

// friendlyDeviceLabel: 3f20c000.pwm → pwm; 1f00030000.serial → uart.
func friendlyDeviceLabel(device string) string {
	base := device
	if i := strings.LastIndex(device, "."); i >= 0 && i+1 < len(device) {
		base = device[i+1:]
	}
	switch base {
	case "serial":
		return "uart"
	default:
		return base
	}
}

func labelMatchesPeriph(periphName, label string) bool {
	if label == "" {
		return false
	}
	if label == periphName {
		return true
	}
	// Classic bcm2835 often yields bare "pwm"/"i2c"/"spi"/"uart".
	if strings.HasPrefix(periphName, "pwm") && (label == "pwm" || strings.HasPrefix(label, "pwm")) {
		return true
	}
	if strings.HasPrefix(periphName, "i2c") && (label == "i2c" || strings.HasPrefix(label, "i2c")) {
		return true
	}
	if strings.HasPrefix(periphName, "spi") && (label == "spi" || strings.HasPrefix(label, "spi")) {
		return true
	}
	if strings.HasPrefix(periphName, "uart") && (label == "uart" || label == "serial" || strings.HasPrefix(label, "uart")) {
		return true
	}
	if periphName == "i2s" && strings.Contains(label, "i2s") {
		return true
	}
	return false
}

func (c *GpioCtrl) ensureOsPinMap() {
	if c.osPinMap == nil {
		c.osPinMap = core.NewStringListWithLen(osPinMapLen)
	}
}

// reenumerateOsPinMap reads pinmux-pins (+ pins alt fallback) and replaces osPinMap (BCM 0..27).
// Fail soft: unreadable → all "".
func (c *GpioCtrl) reenumerateOsPinMap() *core.StringList {
	c.ensureOsPinMap()
	buf := make([]string, osPinMapLen)
	data, err := pinmuxPinsRead()
	if err == nil {
		parsePinmuxPins(data, buf)
	}
	if pins, err := pinctrlPinsRead(); err == nil {
		mergeAltFunctionLabels(pins, buf)
	}
	for i, s := range buf {
		c.osPinMap.Put(i, s)
	}
	return c.osPinMap
}

func (c *GpioCtrl) osMappedForBCM(bcm int) string {
	c.reenumerateOsPinMap()
	if bcm < 0 || bcm >= c.osPinMap.Len() {
		return ""
	}
	return c.osPinMap.At(bcm)
}

func (c *GpioCtrl) osMappedForPeriph(p Peripheral) string {
	c.reenumerateOsPinMap()
	req := p.ReqPins()
	if len(req) == 0 {
		return ""
	}
	for _, lp := range req {
		if lp == nil || lp.BCM < 0 || lp.BCM >= c.osPinMap.Len() {
			return ""
		}
		lab := c.osPinMap.At(lp.BCM)
		if !labelMatchesPeriph(p.Name(), lab) {
			return ""
		}
	}
	return p.Name()
}
