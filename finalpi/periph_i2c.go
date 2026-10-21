package finalpi

import (
	"fmt"
	"os"

	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/internal/core"
)

// I2cPeriph is a GpioCtrl-managed I²C bus (/dev/i2c-N).
// Address is per-call on write/read/tx — not stored on the bus object.
type I2cPeriph struct {
	PeripheralBase
	bus  int
	path string
	fd   *os.File
}

func i2cDevPath(envKey string, bus int) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return fmt.Sprintf("/dev/i2c-%d", bus)
}

func newI2cPeriph(name string, bus, sdaBCM, sclBCM int) *I2cPeriph {
	return &I2cPeriph{
		PeripheralBase: PeripheralBase{
			name:     name,
			instance: bus,
			needsPi5: name == "i2c2" || name == "i2c3",
			reqPins: []*LogiPin{
				{Name: "sda", BCM: sdaBCM},
				{Name: "scl", BCM: sclBCM},
			},
		},
		bus:  bus,
		path: i2cDevPath(fmt.Sprintf("FINAL_I2C%d", bus), bus),
	}
}

func (b *I2cPeriph) SetEnabled(enabled bool) error {
	return GpioCtrlInst().applyI2cEnabled(b, enabled)
}

func (b *I2cPeriph) activate() error {
	if b.active {
		return nil
	}
	fd, err := openI2C(b.path)
	if err != nil {
		return err
	}
	b.fd = fd
	b.active = true
	return nil
}

func (b *I2cPeriph) release() {
	if b.fd != nil {
		_ = b.fd.Close()
		b.fd = nil
	}
	b.active = false
}

func (b *I2cPeriph) restoreResetDefaults() {
	b.release()
	b.enabled = false
}

func (b *I2cPeriph) requireActive() error {
	if !b.active || b.fd == nil {
		return fmt.Errorf("%s: not active", b.name)
	}
	return nil
}

func (b *I2cPeriph) Write(addr int, data []byte) error {
	if err := b.requireActive(); err != nil {
		return err
	}
	if addr < 0 || addr > 0x7f {
		return fmt.Errorf("%s.write: addr 0x%x out of 7-bit range", b.name, addr)
	}
	return i2cWrite(b.fd, uint16(addr), data)
}

func (b *I2cPeriph) Read(addr int, n int) ([]byte, error) {
	if err := b.requireActive(); err != nil {
		return nil, err
	}
	if addr < 0 || addr > 0x7f {
		return nil, fmt.Errorf("%s.read: addr 0x%x out of 7-bit range", b.name, addr)
	}
	if n < 0 {
		return nil, fmt.Errorf("%s.read: n < 0", b.name)
	}
	if n == 0 {
		return nil, nil
	}
	return i2cRead(b.fd, uint16(addr), n)
}

// Tx writes w then reads nRead with a repeated start (I2C_RDWR).
func (b *I2cPeriph) Tx(addr int, w []byte, nRead int) ([]byte, error) {
	if err := b.requireActive(); err != nil {
		return nil, err
	}
	if addr < 0 || addr > 0x7f {
		return nil, fmt.Errorf("%s.tx: addr 0x%x out of 7-bit range", b.name, addr)
	}
	if nRead < 0 {
		return nil, fmt.Errorf("%s.tx: nRead < 0", b.name)
	}
	return i2cTx(b.fd, uint16(addr), w, nRead)
}

func byteListToBytes(list *finalembed.ByteList) []byte {
	if list == nil {
		return nil
	}
	out := make([]byte, list.Len())
	for i := 0; i < list.Len(); i++ {
		out[i] = list.At(i)
	}
	return out
}

func bytesToByteList(b []byte) *finalembed.ByteList {
	return core.NewByteListFromBytes(b)
}

func registerI2cPeriph(ty *finalembed.Type) {
	ty.AddClassResolution("Peripheral")
	ty.RegisterProperty("bus", "Int",
		func(inst any) any { return finalembed.Int(inst.(*I2cPeriph).bus) },
		nil,
	)
	ty.RegisterProperty("path", "String",
		func(inst any) any { return finalembed.String(inst.(*I2cPeriph).path) },
		nil,
	)

	ty.RegisterNativeMethod("write", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		b := ths.Obj.(*I2cPeriph)
		data := byteListToBytes(f.RP.DataStack.Pop().(*finalembed.ByteList))
		addr := int(f.RP.DataStack.Pop().(finalembed.Int))
		if err := b.Write(addr, data); err != nil {
			panic(err)
		}
	})
	ty.RegisterNativeMethod("read", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		b := ths.Obj.(*I2cPeriph)
		n := int(f.RP.DataStack.Pop().(finalembed.Int))
		addr := int(f.RP.DataStack.Pop().(finalembed.Int))
		got, err := b.Read(addr, n)
		if err != nil {
			panic(err)
		}
		f.RP.DataStack.Push(bytesToByteList(got))
	})
	ty.RegisterNativeMethod("tx", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		b := ths.Obj.(*I2cPeriph)
		nRead := int(f.RP.DataStack.Pop().(finalembed.Int))
		w := byteListToBytes(f.RP.DataStack.Pop().(*finalembed.ByteList))
		addr := int(f.RP.DataStack.Pop().(finalembed.Int))
		got, err := b.Tx(addr, w, nRead)
		if err != nil {
			panic(err)
		}
		f.RP.DataStack.Push(bytesToByteList(got))
	})
}
