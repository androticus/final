package finalpi

import (
	"fmt"
	"os"
	"strings"

	"github.com/androticus/final/finalembed"
)

// SPI mode / word-width class constants (Final: SpiPeriph.mode0, .bitsPerWord8, …).
// maxHz stays a free Int — spidev accepts any positive rate and the driver rounds.
const (
	SpiMode0 = 0
	SpiMode1 = 1
	SpiMode2 = 2
	SpiMode3 = 3

	SpiBitsPerWord8  = 8
	SpiBitsPerWord16 = 16
)

// SpiPeriph is a GpioCtrl-managed SPI bus (/dev/spidevB.C for C in 0..numCS-1).
type SpiPeriph struct {
	PeripheralBase
	path        string // CE0 node, e.g. /dev/spidev0.0; other CEs derived
	ceBCMs      []int  // hardware CE0..CEmax-1 BCM numbers
	maxNumCS    int    // len(ceBCMs): spi0=2, spi1=3, spi5=2
	numCS       int    // 1..maxNumCS; CE0 in reqPins, CE1.. in optPins
	selCS       int    // modal CE for transfer/write/read; 0..numCS-1
	mode        uint8
	bitsPerWord uint8
	maxHz       uint32
	fds         []*os.File // index = CE; all opened on activate
}

func spiDevPath(envKey, def string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return def
}

func newSpiPeriph(name string, instance int, path string, mosi, miso, sclk int, ceBCMs []int) *SpiPeriph {
	if len(ceBCMs) < 1 {
		panic(fmt.Sprintf("newSpiPeriph %s: need at least CE0", name))
	}
	s := &SpiPeriph{
		PeripheralBase: PeripheralBase{
			name:     name,
			instance: instance,
			needsPi5: name == "spi1" || name == "spi5",
			reqPins: []*LogiPin{
				{Name: "mosi", BCM: mosi},
				{Name: "miso", BCM: miso},
				{Name: "sclk", BCM: sclk},
				{Name: "ce0", BCM: ceBCMs[0]},
			},
		},
		path:        path,
		ceBCMs:      append([]int(nil), ceBCMs...),
		maxNumCS:    len(ceBCMs),
		numCS:       1,
		selCS:       0,
		mode:        SpiMode0,
		bitsPerWord: SpiBitsPerWord8,
		maxHz:       1_000_000,
	}
	s.syncCSOptPins()
	return s
}

func (s *SpiPeriph) syncCSOptPins() {
	if s.numCS <= 1 {
		s.optPins = nil
		return
	}
	opts := make([]*LogiPin, 0, s.numCS-1)
	for i := 1; i < s.numCS; i++ {
		opts = append(opts, &LogiPin{Name: fmt.Sprintf("ce%d", i), BCM: s.ceBCMs[i]})
	}
	s.optPins = opts
}

func (s *SpiPeriph) pathForCE(ce int) string {
	base := s.path
	if i := strings.LastIndex(base, "."); i >= 0 {
		suf := base[i+1:]
		ok := len(suf) > 0
		for _, c := range suf {
			if c < '0' || c > '9' {
				ok = false
				break
			}
		}
		if ok {
			return fmt.Sprintf("%s.%d", base[:i], ce)
		}
	}
	return fmt.Sprintf("%s.%d", base, ce)
}

func (s *SpiPeriph) SetEnabled(enabled bool) error {
	return GpioCtrlInst().applySpiEnabled(s, enabled)
}

func (s *SpiPeriph) setMode(m int) {
	if s.enabled || s.active {
		panic(fmt.Errorf("%s.mode!: must set before enabled", s.name))
	}
	if m < SpiMode0 || m > SpiMode3 {
		panic(fmt.Errorf("%s.mode!: want 0..3, got %d", s.name, m))
	}
	s.mode = uint8(m)
}

func (s *SpiPeriph) setBitsPerWord(n int) {
	if s.enabled || s.active {
		panic(fmt.Errorf("%s.bitsPerWord!: must set before enabled", s.name))
	}
	if n != SpiBitsPerWord8 && n != SpiBitsPerWord16 {
		panic(fmt.Errorf("%s.bitsPerWord!: want 8 or 16, got %d", s.name, n))
	}
	s.bitsPerWord = uint8(n)
}

func (s *SpiPeriph) setMaxHz(hz int) {
	if s.enabled || s.active {
		panic(fmt.Errorf("%s.maxHz!: must set before enabled", s.name))
	}
	if hz <= 0 {
		panic(fmt.Errorf("%s.maxHz!: must be > 0", s.name))
	}
	s.maxHz = uint32(hz)
}

func (s *SpiPeriph) setNumCS(n int) {
	if s.enabled || s.active {
		panic(fmt.Errorf("%s.numCS!: must set before enabled", s.name))
	}
	if n < 1 || n > s.maxNumCS {
		panic(fmt.Errorf("%s.numCS!: want 1..%d, got %d", s.name, s.maxNumCS, n))
	}
	s.numCS = n
	if s.selCS >= s.numCS {
		s.selCS = 0
	}
	s.syncCSOptPins()
}

func (s *SpiPeriph) setSelCS(ce int) {
	if ce < 0 || ce >= s.numCS {
		panic(fmt.Errorf("%s.selCS!: want 0..%d, got %d", s.name, s.numCS-1, ce))
	}
	s.selCS = ce
}

func (s *SpiPeriph) activate() error {
	if s.active {
		return nil
	}
	fds := make([]*os.File, s.numCS)
	for ce := 0; ce < s.numCS; ce++ {
		path := s.pathForCE(ce)
		fd, err := openSPI(path, s.mode, s.bitsPerWord, s.maxHz)
		if err != nil {
			for i := 0; i < ce; i++ {
				_ = fds[i].Close()
			}
			return err
		}
		fds[ce] = fd
	}
	s.fds = fds
	s.active = true
	return nil
}

func (s *SpiPeriph) release() {
	for _, fd := range s.fds {
		if fd != nil {
			_ = fd.Close()
		}
	}
	s.fds = nil
	s.active = false
}

func (s *SpiPeriph) restoreResetDefaults() {
	s.release()
	s.enabled = false
	s.mode = SpiMode0
	s.bitsPerWord = SpiBitsPerWord8
	s.maxHz = 1_000_000
	s.numCS = 1
	s.selCS = 0
	s.syncCSOptPins()
}

func (s *SpiPeriph) requireActive() error {
	if !s.active || len(s.fds) == 0 {
		return fmt.Errorf("%s: not active", s.name)
	}
	return nil
}

func (s *SpiPeriph) selectedFD() (*os.File, error) {
	if err := s.requireActive(); err != nil {
		return nil, err
	}
	if s.selCS < 0 || s.selCS >= len(s.fds) || s.fds[s.selCS] == nil {
		return nil, fmt.Errorf("%s: selCS %d not open", s.name, s.selCS)
	}
	return s.fds[s.selCS], nil
}

func (s *SpiPeriph) Transfer(tx []byte) ([]byte, error) {
	fd, err := s.selectedFD()
	if err != nil {
		return nil, err
	}
	return spiTransfer(fd, tx, s.maxHz, s.bitsPerWord)
}

func (s *SpiPeriph) Write(data []byte) error {
	_, err := s.Transfer(data)
	return err
}

func (s *SpiPeriph) Read(n int) ([]byte, error) {
	if n < 0 {
		return nil, fmt.Errorf("%s.read: n < 0", s.name)
	}
	tx := make([]byte, n)
	return s.Transfer(tx)
}

func registerSpiPeriph(ty *finalembed.Type) {
	ty.AddClassResolution("Peripheral")
	ty.RegisterProperty("path", "String",
		func(inst any) any { return finalembed.String(inst.(*SpiPeriph).path) },
		nil,
	)
	ty.RegisterProperty("mode", "Int",
		func(inst any) any { return finalembed.Int(inst.(*SpiPeriph).mode) },
		func(inst any, v any) { inst.(*SpiPeriph).setMode(int(v.(finalembed.Int))) },
	)
	ty.RegisterProperty("bitsPerWord", "Int",
		func(inst any) any { return finalembed.Int(inst.(*SpiPeriph).bitsPerWord) },
		func(inst any, v any) { inst.(*SpiPeriph).setBitsPerWord(int(v.(finalembed.Int))) },
	)
	ty.RegisterProperty("maxHz", "Int",
		func(inst any) any { return finalembed.Int(inst.(*SpiPeriph).maxHz) },
		func(inst any, v any) { inst.(*SpiPeriph).setMaxHz(int(v.(finalembed.Int))) },
	)
	ty.RegisterProperty("numCS", "Int",
		func(inst any) any { return finalembed.Int(inst.(*SpiPeriph).numCS) },
		func(inst any, v any) { inst.(*SpiPeriph).setNumCS(int(v.(finalembed.Int))) },
	)
	ty.RegisterProperty("selCS", "Int",
		func(inst any) any { return finalembed.Int(inst.(*SpiPeriph).selCS) },
		func(inst any, v any) { inst.(*SpiPeriph).setSelCS(int(v.(finalembed.Int))) },
	)

	ty.RegisterClassProperty("mode0", "Int", func(_ any) any { return finalembed.Int(SpiMode0) }, nil)
	ty.RegisterClassProperty("mode1", "Int", func(_ any) any { return finalembed.Int(SpiMode1) }, nil)
	ty.RegisterClassProperty("mode2", "Int", func(_ any) any { return finalembed.Int(SpiMode2) }, nil)
	ty.RegisterClassProperty("mode3", "Int", func(_ any) any { return finalembed.Int(SpiMode3) }, nil)
	ty.RegisterClassProperty("bitsPerWord8", "Int", func(_ any) any { return finalembed.Int(SpiBitsPerWord8) }, nil)
	ty.RegisterClassProperty("bitsPerWord16", "Int", func(_ any) any { return finalembed.Int(SpiBitsPerWord16) }, nil)

	ty.RegisterNativeMethod("transfer", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		s := ths.Obj.(*SpiPeriph)
		tx := byteListToBytes(f.RP.DataStack.Pop().(*finalembed.ByteList))
		rx, err := s.Transfer(tx)
		if err != nil {
			panic(err)
		}
		f.RP.DataStack.Push(bytesToByteList(rx))
	})
	ty.RegisterNativeMethod("write", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		s := ths.Obj.(*SpiPeriph)
		data := byteListToBytes(f.RP.DataStack.Pop().(*finalembed.ByteList))
		if err := s.Write(data); err != nil {
			panic(err)
		}
	})
	ty.RegisterNativeMethod("read", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		s := ths.Obj.(*SpiPeriph)
		n := int(f.RP.DataStack.Pop().(finalembed.Int))
		got, err := s.Read(n)
		if err != nil {
			panic(err)
		}
		f.RP.DataStack.Push(bytesToByteList(got))
	})
}
