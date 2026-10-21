//go:build linux

package finalpi

import (
	"testing"
	"unsafe"
)

func TestSpiIOCTransferSize(t *testing.T) {
	if unsafe.Sizeof(spiIOCTransfer{}) != 32 {
		t.Fatalf("spiIOCTransfer size=%d want 32 (SPI_IOC_MESSAGE encoding)", unsafe.Sizeof(spiIOCTransfer{}))
	}
}
