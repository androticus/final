//go:build linux

package finalpi

import (
	"os"
	"strings"
)

// deviceTreeModelPaths is the read order for board model (overridable in tests).
var deviceTreeModelPaths = []string{
	"/proc/device-tree/model",
	"/sys/firmware/devicetree/base/model",
}

func readDeviceTreeModel() string {
	for _, p := range deviceTreeModelPaths {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		return strings.TrimRight(string(b), "\x00")
	}
	return ""
}
