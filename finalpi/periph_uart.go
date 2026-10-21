package finalpi

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/internal/core"
)

var errTXTimeout = errors.New("TX timeout")

// UartParity is termios parity (Int in Final).
type UartParity int

const (
	UartParityNone UartParity = iota
	UartParityOdd
	UartParityEven
)

// UartPeriph is a GpioCtrl-managed UART (uart0/2/3/4; uart1 HAT ID omitted).
type UartPeriph struct {
	PeripheralBase
	path     string
	baud     int
	dataBits int
	stopBits int
	parity   UartParity
	flowHW   bool
	ctsBCM   int // hardwired CTS when flowHW (overlay ctsrts)
	rtsBCM   int // hardwired RTS when flowHW

	port   uartPort
	rxQ    *rxQueue
	rxWork *Worker

	txMu          sync.Mutex
	onWriteNotify core.HandlerMintEventFn
	txGen         atomic.Uint64 // bumped on release; cancels in-flight writeNotify mint

	rxMu        sync.Mutex
	onReadReady core.HandlerMintEventFn
	armed       bool          // readWhenReady one-shot armed
	rxReadyGen  atomic.Uint64 // bumped on release; cancels in-flight ready watch
}

// UartReader is a pull session over a UartPeriph RX queue (foreach-friendly).
type UartReader struct {
	u         *UartPeriph
	maxN      int // session byte budget; 0 = unlimited
	timeoutMs int // idle between bytes; default 100
	untilByte int // 0..255 delimiter, or -1 none

	n        int
	timedOut bool
	hitDelim bool
	ok       bool
	errStr   string
}

func uartDevPath(envKey, def string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return def
}

// uart0DefaultPath is the header UART node for the detected board.
// Pi Zero 2 keeps PL011 on Bluetooth; header is mini UART (/dev/ttyS0).
// Pi 5 uart0-pi5 → /dev/ttyAMA0. Do not use /dev/serial0 (Pi 5: debug JST).
func uart0DefaultPath(support PiSupportType) string {
	if support == PiSupportPiZ2 {
		return "/dev/ttyS0"
	}
	return "/dev/ttyAMA0"
}

// rejectUnsupportedUartBaud catches rates termios accepts but hardware ignores.
// Pi Zero 2 mini UART (ttyS*) keeps the prior rate for baud < 1200.
func rejectUnsupportedUartBaud(path string, baud int) error {
	base := filepath.Base(path)
	if strings.HasPrefix(base, "ttyS") && baud > 0 && baud < 1200 {
		return fmt.Errorf("UART %s: baud %d unsupported on mini UART (use >= 1200; kernel would keep prior rate)", path, baud)
	}
	return nil
}

func newUartPeriph(name string, instance int, path string, txBCM, rxBCM, ctsBCM, rtsBCM int) *UartPeriph {
	return &UartPeriph{
		PeripheralBase: PeripheralBase{
			name:     name,
			instance: instance,
			needsPi5: name == "uart2" || name == "uart3",
			reqPins: []*LogiPin{
				{Name: "tx", BCM: txBCM},
				{Name: "rx", BCM: rxBCM},
			},
		},
		path:     path,
		baud:     115200,
		dataBits: 8,
		stopBits: 1,
		parity:   UartParityNone,
		ctsBCM:   ctsBCM,
		rtsBCM:   rtsBCM,
		rxWork:   NewWorker(),
	}
}

func (u *UartPeriph) SetEnabled(enabled bool) error {
	return GpioCtrlInst().applyUartEnabled(u, enabled)
}

func (u *UartPeriph) setBaud(n int) {
	if u.active {
		panic(fmt.Errorf("%s.baud!: active", u.name))
	}
	if n <= 0 {
		panic(fmt.Errorf("%s.baud!: must be > 0", u.name))
	}
	u.baud = n
}

func (u *UartPeriph) setDataBits(n int) {
	if u.active {
		panic(fmt.Errorf("%s.dataBits!: active", u.name))
	}
	if n != 5 && n != 6 && n != 7 && n != 8 {
		panic(fmt.Errorf("%s.dataBits!: want 5..8, got %d", u.name, n))
	}
	u.dataBits = n
}

func (u *UartPeriph) setStopBits(n int) {
	if u.active {
		panic(fmt.Errorf("%s.stopBits!: active", u.name))
	}
	if n != 1 && n != 2 {
		panic(fmt.Errorf("%s.stopBits!: want 1 or 2, got %d", u.name, n))
	}
	u.stopBits = n
}

func (u *UartPeriph) setParity(p UartParity) {
	if u.active {
		panic(fmt.Errorf("%s.parity!: active", u.name))
	}
	u.parity = p
}

func (u *UartPeriph) setFlowHW(v bool) {
	if u.enabled || u.active {
		panic(fmt.Errorf("%s.flowHW!: must set before enabled", u.name))
	}
	u.flowHW = v
	u.syncFlowOptPins()
}

// syncFlowOptPins puts the hardwired CTS/RTS pair on optPins iff flowHW.
// Validate claims OptPins when the UART is enabled — keeps Final surface to one Bool.
func (u *UartPeriph) syncFlowOptPins() {
	if u.flowHW {
		u.optPins = []*LogiPin{
			{Name: "cts", BCM: u.ctsBCM},
			{Name: "rts", BCM: u.rtsBCM},
		}
		return
	}
	u.optPins = nil
}

func (u *UartPeriph) activate() error {
	if u.active {
		return nil
	}
	port, err := openUartPort(u.path, u.baud, u.dataBits, u.stopBits, u.parity, u.flowHW)
	if err != nil {
		return err
	}
	u.port = port
	u.rxQ = newRxQueue(4096)
	u.startRX()
	u.active = true
	return nil
}

// activateWithPort is for host tests (pipePort over net.Pipe).
func (u *UartPeriph) activateWithPort(port uartPort) error {
	if u.active {
		return fmt.Errorf("%s: already active", u.name)
	}
	u.port = port
	u.rxQ = newRxQueue(4096)
	u.startRX()
	u.active = true
	u.enabled = true
	return nil
}

func (u *UartPeriph) startRX() {
	port := u.port
	q := u.rxQ
	u.rxWork.Go(func(cancel <-chan struct{}) {
		buf := make([]byte, 256)
		for {
			select {
			case <-cancel:
				q.close(nil)
				return
			default:
			}
			n, err := ReadWithTimeout(port, port, buf, 50*time.Millisecond)
			if n > 0 {
				q.push(buf[:n])
			}
			if err != nil {
				if os.IsTimeout(err) || isTimeoutErr(err) {
					continue
				}
				q.close(err)
				return
			}
		}
	})
}

func isTimeoutErr(err error) bool {
	type timeout interface{ Timeout() bool }
	if t, ok := err.(timeout); ok && t.Timeout() {
		return true
	}
	return false
}

func (u *UartPeriph) release() {
	u.txGen.Add(1)
	u.rxReadyGen.Add(1)
	if u.rxWork != nil {
		_ = u.rxWork.Stop(2 * time.Second)
	}
	if u.rxQ != nil {
		u.rxQ.close(nil)
		u.rxQ = nil
	}
	if u.port != nil {
		_ = u.port.Close()
		u.port = nil
	}
	u.active = false
	// enabled false / Reset: clear armed + both notify handlers (148.2).
	u.txMu.Lock()
	u.onWriteNotify = nil
	u.txMu.Unlock()
	u.rxMu.Lock()
	u.onReadReady = nil
	u.armed = false
	u.rxMu.Unlock()
}

func (u *UartPeriph) restoreResetDefaults() {
	u.release()
	u.enabled = false
	u.baud = 115200
	u.dataBits = 8
	u.stopBits = 1
	u.parity = UartParityNone
	u.flowHW = false
	u.syncFlowOptPins()
}

// uartBitsPerFrame is start + data + optional parity + stop.
func uartBitsPerFrame(dataBits, stopBits int, parity UartParity) int {
	n := 1 + dataBits + stopBits
	if parity != UartParityNone {
		n++
	}
	return n
}

// uartTxTimeout = nBytes * (bits/baud) * 1.15.
func uartTxTimeout(nBytes, baud, dataBits, stopBits int, parity UartParity) time.Duration {
	if nBytes <= 0 {
		return time.Millisecond
	}
	if baud <= 0 {
		return 2 * time.Second
	}
	bits := uartBitsPerFrame(dataBits, stopBits, parity)
	ns := float64(nBytes) * float64(bits) / float64(baud) * 1.15 * 1e9
	d := time.Duration(ns)
	if d < time.Millisecond {
		return time.Millisecond
	}
	return d
}

func (u *UartPeriph) txTimeout(nBytes int) time.Duration {
	return uartTxTimeout(nBytes, u.baud, u.dataBits, u.stopBits, u.parity)
}

func writeAll(port uartPort, data []byte) error {
	for len(data) > 0 {
		n, err := port.Write(data)
		if err != nil {
			return err
		}
		if n <= 0 {
			return fmt.Errorf("short write")
		}
		data = data[n:]
	}
	return nil
}

// runTXOp runs fn under timeout (serial has no write deadline).
func runTXOp(timeout time.Duration, fn func() error) error {
	if timeout <= 0 {
		timeout = time.Millisecond
	}
	done := make(chan error, 1)
	go func() { done <- fn() }()
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return errTXTimeout
	}
}

// writeBytes queues into the OS/driver (no tcdrain). Timeout covers the write path.
func (u *UartPeriph) writeBytes(data []byte) error {
	if !u.active || u.port == nil {
		return fmt.Errorf("%s.write: not active", u.name)
	}
	if len(data) == 0 {
		return nil
	}
	port := u.port
	payload := data
	timeout := u.txTimeout(len(payload))
	if err := runTXOp(timeout, func() error {
		if err := writeAll(port, payload); err != nil {
			return fmt.Errorf("%s.write: %w", u.name, err)
		}
		return nil
	}); err != nil {
		if errors.Is(err, errTXTimeout) {
			return fmt.Errorf("%s.write: TX timeout", u.name)
		}
		return err
	}
	return nil
}

func (u *UartPeriph) setWriteNotify(handler *core.MethodRef, index int) {
	u.txMu.Lock()
	defer u.txMu.Unlock()
	u.onWriteNotify = MintEvent(handler, index)
}

// writeNotify queues + drains on a worker, then mints onWriteNotify (param0=ok, param1=err if needed).
func (u *UartPeriph) writeNotify(data []byte) {
	if !u.active || u.port == nil {
		panic(fmt.Errorf("%s.writeNotify: not active", u.name))
	}
	u.txMu.Lock()
	mint := u.onWriteNotify
	u.txMu.Unlock()
	if mint == nil {
		panic(fmt.Errorf("%s.writeNotify: onWriteNotify not set (use setWriteNotify)", u.name))
	}
	if len(data) == 0 {
		mint(u, finalembed.Bool(true))
		return
	}
	port := u.port
	payload := append([]byte(nil), data...)
	timeout := u.txTimeout(len(payload))
	gen := u.txGen.Load()
	name := u.name
	go func() {
		err := runTXOp(timeout, func() error {
			if err := writeAll(port, payload); err != nil {
				return fmt.Errorf("%s.writeNotify: %w", name, err)
			}
			if err := port.Drain(); err != nil {
				return fmt.Errorf("%s.writeNotify drain: %w", name, err)
			}
			return nil
		})
		if u.txGen.Load() != gen {
			return
		}
		if err != nil {
			msg := err.Error()
			if errors.Is(err, errTXTimeout) {
				msg = name + ".writeNotify: TX timeout"
			}
			mint(u, finalembed.Bool(false), finalembed.String(msg))
			return
		}
		mint(u, finalembed.Bool(true))
	}()
}

func popUartWriteData(u *UartPeriph, v any, op string) []byte {
	switch x := v.(type) {
	case finalembed.Int:
		return []byte{byte(x)}
	case *finalembed.ByteList:
		data := make([]byte, x.Len())
		for i := 0; i < x.Len(); i++ {
			data[i] = x.At(i)
		}
		return data
	default:
		panic(fmt.Errorf("%s.%s: want Int or ByteList, got %T", u.name, op, v))
	}
}

func (u *UartPeriph) NewReader() *UartReader {
	return &UartReader{
		u:         u,
		maxN:      0,
		timeoutMs: 100,
		untilByte: -1,
		ok:        true,
	}
}

func (u *UartPeriph) setReadReady(handler *core.MethodRef, index int) {
	u.rxMu.Lock()
	defer u.rxMu.Unlock()
	u.onReadReady = MintEvent(handler, index)
}

// readWhenReady arms a one-shot onReadReady watch. Modal: nil = true; err = false.
// Already armed → nop (nil). Missing handler → panic. Not active → error.
func (u *UartPeriph) readWhenReady() error {
	if !u.active || u.rxQ == nil {
		return fmt.Errorf("%s.readWhenReady: not active", u.name)
	}
	u.rxMu.Lock()
	mint := u.onReadReady
	if mint == nil {
		u.rxMu.Unlock()
		panic(fmt.Errorf("%s.readWhenReady: onReadReady not set (use setReadReady)", u.name))
	}
	if u.armed {
		u.rxMu.Unlock()
		return nil
	}
	u.armed = true
	gen := u.rxReadyGen.Load()
	q := u.rxQ
	u.rxMu.Unlock()

	go u.watchReadReady(gen, q, mint)
	return nil
}

func (u *UartPeriph) watchReadReady(gen uint64, q *rxQueue, mint core.HandlerMintEventFn) {
	const poll = 10 * time.Millisecond
	for {
		if u.rxReadyGen.Load() != gen {
			return
		}
		if q != nil && q.hasData() {
			u.rxMu.Lock()
			if u.rxReadyGen.Load() != gen || !u.armed {
				u.rxMu.Unlock()
				return
			}
			u.armed = false
			m := u.onReadReady
			u.rxMu.Unlock()
			if m != nil {
				m(u)
			} else if mint != nil {
				mint(u)
			}
			return
		}
		time.Sleep(poll)
	}
}

// --- UartReader / Iterator ---

type uartReaderIter struct {
	r       *UartReader
	curr    finalembed.Int
	started bool
	done    bool
}

func (r *UartReader) GetIter() finalembed.Iterator {
	r.n = 0
	r.timedOut = false
	r.hitDelim = false
	r.ok = true
	r.errStr = ""
	return &uartReaderIter{r: r}
}

func (it *uartReaderIter) GetNext() bool {
	if it.done {
		return false
	}
	r := it.r
	if r.u == nil || !r.u.active || r.u.rxQ == nil {
		r.ok = false
		r.errStr = "uart reader: not active"
		it.done = true
		return false
	}
	if r.maxN > 0 && r.n >= r.maxN {
		it.done = true
		return false
	}
	idle := time.Duration(r.timeoutMs) * time.Millisecond
	if r.timeoutMs <= 0 {
		idle = 100 * time.Millisecond
	}
	// First byte may wait longer; idle timeout applies between bytes after start.
	b, ok, timedOut, err := r.u.rxQ.popByte(r.u.rxWork.cancelOrNil(), idle)
	if err != nil {
		r.ok = false
		r.errStr = err.Error()
		it.done = true
		return false
	}
	if timedOut {
		r.timedOut = true
		it.done = true
		return false
	}
	if !ok {
		it.done = true
		return false
	}
	it.curr = finalembed.Int(b)
	r.n++
	it.started = true
	if r.untilByte >= 0 && int(b) == r.untilByte {
		r.hitDelim = true
		it.done = true
		// include delimiter in stream (Current already set); next GetNext ends
		return true
	}
	return true
}

func (it *uartReaderIter) Current() any {
	return it.curr
}

func (w *Worker) cancelOrNil() <-chan struct{} {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cancel == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return w.cancel
}

func registerUartPeriph(ty *finalembed.Type) {
	ty.AddClassResolution("Peripheral")
	ty.RegisterProperty("path", "String",
		func(inst any) any { return finalembed.String(inst.(*UartPeriph).path) },
		nil,
	)
	ty.RegisterProperty("baud", "Int",
		func(inst any) any { return finalembed.Int(inst.(*UartPeriph).baud) },
		func(inst any, v any) { inst.(*UartPeriph).setBaud(int(v.(finalembed.Int))) },
	)
	ty.RegisterProperty("dataBits", "Int",
		func(inst any) any { return finalembed.Int(inst.(*UartPeriph).dataBits) },
		func(inst any, v any) { inst.(*UartPeriph).setDataBits(int(v.(finalembed.Int))) },
	)
	ty.RegisterProperty("stopBits", "Int",
		func(inst any) any { return finalembed.Int(inst.(*UartPeriph).stopBits) },
		func(inst any, v any) { inst.(*UartPeriph).setStopBits(int(v.(finalembed.Int))) },
	)
	ty.RegisterProperty("parity", "Int",
		func(inst any) any { return finalembed.Int(inst.(*UartPeriph).parity) },
		func(inst any, v any) { inst.(*UartPeriph).setParity(UartParity(v.(finalembed.Int))) },
	)
	ty.RegisterProperty("flowHW", "Bool",
		func(inst any) any { return finalembed.Bool(inst.(*UartPeriph).flowHW) },
		func(inst any, v any) { inst.(*UartPeriph).setFlowHW(bool(v.(finalembed.Bool))) },
	)
	ty.RegisterClassProperty("parityNone", "Int", func(_ any) any { return finalembed.Int(UartParityNone) }, nil)
	ty.RegisterClassProperty("parityOdd", "Int", func(_ any) any { return finalembed.Int(UartParityOdd) }, nil)
	ty.RegisterClassProperty("parityEven", "Int", func(_ any) any { return finalembed.Int(UartParityEven) }, nil)

	// ( value:Int|data:ByteList -- true | err:s false )
	ty.RegisterNativeMethod("write", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		u := ths.Obj.(*UartPeriph)
		if !u.active || u.port == nil {
			panic(fmt.Errorf("%s.write: not active", u.name))
		}
		data := popUartWriteData(u, f.RP.DataStack.Pop(), "write")
		ds := f.RP.DataStack
		if err := u.writeBytes(data); err != nil {
			ds.Push(finalembed.String(err.Error()))
			ds.Push(finalembed.Bool(false))
			return
		}
		ds.Push(finalembed.Bool(true))
	})
	// ( value:Int|data:ByteList -- ) completion via onWriteNotify Event
	ty.RegisterNativeMethod("writeNotify", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		u := ths.Obj.(*UartPeriph)
		data := popUartWriteData(u, f.RP.DataStack.Pop(), "writeNotify")
		u.writeNotify(data)
	})
	// ( handler:MethodRef index:Int -- ) Event params: ok:b [err:s]
	ty.RegisterNativeMethod("setWriteNotify", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		index := int(f.RP.DataStack.Pop().(finalembed.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		ths.Obj.(*UartPeriph).setWriteNotify(handler, index)
	})
	// ( handler:MethodRef index:Int -- ) Event: schedule signal to newReader
	ty.RegisterNativeMethod("setReadReady", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		index := int(f.RP.DataStack.Pop().(finalembed.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		ths.Obj.(*UartPeriph).setReadReady(handler, index)
	})
	// ( -- true | err:s false ) arm one-shot onReadReady (nop if already armed)
	ty.RegisterNativeMethod("readWhenReady", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		u := ths.Obj.(*UartPeriph)
		ds := f.RP.DataStack
		if err := u.readWhenReady(); err != nil {
			ds.Push(finalembed.String(err.Error()))
			ds.Push(finalembed.Bool(false))
			return
		}
		ds.Push(finalembed.Bool(true))
	})
	ty.RegisterNativeMethod("newReader", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*UartPeriph).NewReader())
	})
}

func registerUartReader(ty *finalembed.Type) {
	ty.RegisterProperty("maxN", "Int",
		func(inst any) any { return finalembed.Int(inst.(*UartReader).maxN) },
		func(inst any, v any) { inst.(*UartReader).maxN = int(v.(finalembed.Int)) },
	)
	ty.RegisterProperty("timeout", "Int",
		func(inst any) any { return finalembed.Int(inst.(*UartReader).timeoutMs) },
		func(inst any, v any) { inst.(*UartReader).timeoutMs = int(v.(finalembed.Int)) },
	)
	ty.RegisterProperty("untilByte", "Int",
		func(inst any) any { return finalembed.Int(inst.(*UartReader).untilByte) },
		func(inst any, v any) { inst.(*UartReader).untilByte = int(v.(finalembed.Int)) },
	)
	ty.RegisterProperty("n", "Int",
		func(inst any) any { return finalembed.Int(inst.(*UartReader).n) },
		nil,
	)
	ty.RegisterProperty("timedOut", "Bool",
		func(inst any) any { return finalembed.Bool(inst.(*UartReader).timedOut) },
		nil,
	)
	ty.RegisterProperty("hitDelim", "Bool",
		func(inst any) any { return finalembed.Bool(inst.(*UartReader).hitDelim) },
		nil,
	)
	ty.RegisterProperty("ok", "Bool",
		func(inst any) any { return finalembed.Bool(inst.(*UartReader).ok) },
		nil,
	)
	ty.RegisterProperty("err", "String",
		func(inst any) any { return finalembed.String(inst.(*UartReader).errStr) },
		nil,
	)
	// foreach path for TypeObjPair receivers
	ty.RegisterNativeMethod("_tryGetIter", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
		ds := f.RP.DataStack
		if ds.Len() > 0 && ds.Peek() == any(ths.Obj) {
			ds.Pop()
		}
		ds.Push(ths.Obj.(*UartReader).GetIter())
		ds.Push(finalembed.Bool(true))
	})
}
