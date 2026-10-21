//go:build !linux

package finalpi

import (
	"fmt"
	"os"
)

func openSPI(path string, mode, bits uint8, maxHz uint32) (*os.File, error) {
	_ = mode
	_ = bits
	_ = maxHz
	return nil, fmt.Errorf("SPI %s requires Linux (this binary has no spidev backend)", path)
}

func spiTransfer(_ *os.File, _ []byte, _ uint32, _ uint8) ([]byte, error) {
	return nil, fmt.Errorf("SPI: no backend")
}
