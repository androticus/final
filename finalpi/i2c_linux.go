//go:build linux

package finalpi

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Linux i2c-dev I2C_RDWR (include/uapi/linux/i2c-dev.h).
const i2cRDWR = 0x0707

// I2C_M_RD
const i2cMRd = 0x0001

// Matches struct i2c_msg on 64-bit Linux (2-byte pad before pointer).
type i2cMsg struct {
	Addr  uint16
	Flags uint16
	Len   uint16
	_     uint16
	Buf   uintptr
}

type i2cRdwrIoctlData struct {
	Msgs  uintptr
	Nmsgs uint32
	_     uint32
}

func openI2C(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("I2C open %s: %w (dtparam=i2c_arm=on / dtoverlay=i2cN-pi5? i2c group?)", path, err)
	}
	return f, nil
}

func i2cWrite(f *os.File, addr uint16, data []byte) error {
	if len(data) == 0 {
		// Zero-length write still probes ACK on some devices; allow it.
		return i2cRdwr(f, []i2cMsg{{Addr: addr, Flags: 0, Len: 0, Buf: 0}})
	}
	return i2cRdwr(f, []i2cMsg{{
		Addr:  addr,
		Flags: 0,
		Len:   uint16(len(data)),
		Buf:   uintptr(unsafe.Pointer(&data[0])),
	}})
}

func i2cRead(f *os.File, addr uint16, n int) ([]byte, error) {
	buf := make([]byte, n)
	msg := i2cMsg{
		Addr:  addr,
		Flags: i2cMRd,
		Len:   uint16(n),
		Buf:   uintptr(unsafe.Pointer(&buf[0])),
	}
	if err := i2cRdwr(f, []i2cMsg{msg}); err != nil {
		return nil, err
	}
	return buf, nil
}

func i2cTx(f *os.File, addr uint16, w []byte, nRead int) ([]byte, error) {
	msgs := make([]i2cMsg, 0, 2)
	if len(w) > 0 {
		msgs = append(msgs, i2cMsg{
			Addr:  addr,
			Flags: 0,
			Len:   uint16(len(w)),
			Buf:   uintptr(unsafe.Pointer(&w[0])),
		})
	} else if nRead == 0 {
		msgs = append(msgs, i2cMsg{Addr: addr, Flags: 0, Len: 0, Buf: 0})
	}
	var rbuf []byte
	if nRead > 0 {
		rbuf = make([]byte, nRead)
		msgs = append(msgs, i2cMsg{
			Addr:  addr,
			Flags: i2cMRd,
			Len:   uint16(nRead),
			Buf:   uintptr(unsafe.Pointer(&rbuf[0])),
		})
	}
	if err := i2cRdwr(f, msgs); err != nil {
		return nil, err
	}
	return rbuf, nil
}

func i2cRdwr(f *os.File, msgs []i2cMsg) error {
	if len(msgs) == 0 {
		return nil
	}
	data := i2cRdwrIoctlData{
		Msgs:  uintptr(unsafe.Pointer(&msgs[0])),
		Nmsgs: uint32(len(msgs)),
	}
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), i2cRDWR, uintptr(unsafe.Pointer(&data)))
	if errno != 0 {
		return fmt.Errorf("I2C_RDWR: %w", errno)
	}
	return nil
}
