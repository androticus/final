//go:build linux

package finalpi

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// spidev ioctls (include/uapi/linux/spi/spidev.h), little-endian arm64/amd64.
const (
	spiIOCMagic      = 'k'
	spiIOCWRMode     = 0x40016b01 // _IOW(k,1,u8)
	spiIOCWRBitsPW   = 0x40016b03 // _IOW(k,3,u8)
	spiIOCWRMaxSpeed = 0x40046b04 // _IOW(k,4,u32)
)

type spiIOCTransfer struct {
	TxBuf       uint64
	RxBuf       uint64
	Len         uint32
	SpeedHz     uint32
	DelayUsecs  uint16
	BitsPerWord uint8
	CSChange    uint8
	TxNBits     uint8
	RxNBits     uint8
	WordDelayUs uint8
	Pad         uint8
}

func spiIOCMessage(n uint) uintptr {
	const sizeOfTransfer = 32 // sizeof(struct spi_ioc_transfer) on 64-bit Linux
	size := n * sizeOfTransfer
	if size == 0 || size > 1<<13 {
		return 0
	}
	// _IOW('k', 0, char[size])
	return uintptr(0x40000000 | (size << 16) | (spiIOCMagic << 8) | 0)
}

func openSPI(path string, mode, bits uint8, maxHz uint32) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("SPI open %s: %w (dtparam=spi=on / dtoverlay=spi0-2cs? spi group?)", path, err)
	}
	if err := ioctlU8(f, spiIOCWRMode, mode); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("SPI mode: %w", err)
	}
	if err := ioctlU8(f, spiIOCWRBitsPW, bits); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("SPI bits: %w", err)
	}
	if err := ioctlU32(f, spiIOCWRMaxSpeed, maxHz); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("SPI maxHz: %w", err)
	}
	return f, nil
}

func ioctlU8(f *os.File, req uintptr, v uint8) error {
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), req, uintptr(unsafe.Pointer(&v)))
	if errno != 0 {
		return errno
	}
	return nil
}

func ioctlU32(f *os.File, req uintptr, v uint32) error {
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), req, uintptr(unsafe.Pointer(&v)))
	if errno != 0 {
		return errno
	}
	return nil
}

// spidevMaxChunk is the default spidev bufsiz (/sys/module/spidev/parameters/bufsiz).
// Larger transfers must be split; CS may toggle between chunks (OK for JD79661 / write-only panels).
const spidevMaxChunk = 4096

func spiTransfer(f *os.File, tx []byte, speedHz uint32, bits uint8) ([]byte, error) {
	n := len(tx)
	rx := make([]byte, n)
	if n == 0 {
		return rx, nil
	}
	req := spiIOCMessage(1)
	if req == 0 {
		return nil, fmt.Errorf("SPI: bad message size")
	}
	for off := 0; off < n; {
		m := n - off
		if m > spidevMaxChunk {
			m = spidevMaxChunk
		}
		chunkTx := tx[off : off+m]
		chunkRx := rx[off : off+m]
		tr := spiIOCTransfer{
			TxBuf:       uint64(uintptr(unsafe.Pointer(&chunkTx[0]))),
			RxBuf:       uint64(uintptr(unsafe.Pointer(&chunkRx[0]))),
			Len:         uint32(m),
			SpeedHz:     speedHz,
			BitsPerWord: bits,
		}
		_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), req, uintptr(unsafe.Pointer(&tr)))
		if errno != 0 {
			return nil, fmt.Errorf("SPI_IOC_MESSAGE: %w", errno)
		}
		off += m
	}
	return rx, nil
}
