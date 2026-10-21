package finalpi

import (
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/androticus/final/finalembed"
)

func TestUartTxTimeoutFormula(t *testing.T) {
	// 10 bytes, 8N1 @ 115200 → 10 bits/frame → 10*10/115200*1.15 s
	wantNs := float64(10) * 10.0 / 115200.0 * 1.15 * 1e9
	got := uartTxTimeout(10, 115200, 8, 1, UartParityNone)
	delta := float64(got) - wantNs
	if delta < -1e6 || delta > 1e6 { // ±1ms
		t.Fatalf("timeout=%v want ~%v ns", got, wantNs)
	}
	// parity adds one bit
	gotP := uartTxTimeout(10, 115200, 8, 1, UartParityEven)
	if gotP <= got {
		t.Fatalf("parity should lengthen timeout: none=%v even=%v", got, gotP)
	}
}

type drainCountingPort struct {
	pipePort
	drains atomic.Int32
}

func (p *drainCountingPort) Drain() error {
	p.drains.Add(1)
	return nil
}

func TestWriteDoesNotDrain(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()

	tx := newUartPeriph("uartTxNoDrain", 93, "pipe-a", 14, 15, 16, 17)
	port := &drainCountingPort{pipePort: pipePort{a}}
	if err := tx.activateWithPort(port); err != nil {
		t.Fatal(err)
	}
	defer tx.release()

	go func() {
		buf := make([]byte, 16)
		_, _ = b.Read(buf)
	}()
	if err := tx.writeBytes([]byte("hi")); err != nil {
		t.Fatal(err)
	}
	if port.drains.Load() != 0 {
		t.Fatalf("write drained %d times; want 0", port.drains.Load())
	}
}

func TestWriteNotifyDrainsAndMints(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()

	tx := newUartPeriph("uartTxNotify", 94, "pipe-a", 14, 15, 16, 17)
	port := &drainCountingPort{pipePort: pipePort{a}}
	if err := tx.activateWithPort(port); err != nil {
		t.Fatal(err)
	}
	defer tx.release()

	var gotOK atomic.Bool
	var gotErr atomic.Value
	tx.onWriteNotify = func(_ any, params ...any) {
		if len(params) < 1 {
			t.Error("expected ok param")
			return
		}
		ok := bool(params[0].(finalembed.Bool))
		gotOK.Store(ok)
		if !ok && len(params) > 1 {
			gotErr.Store(string(params[1].(finalembed.String)))
		}
	}

	go func() {
		buf := make([]byte, 16)
		_, _ = b.Read(buf)
	}()
	tx.writeNotify([]byte("ok"))

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if port.drains.Load() > 0 && gotOK.Load() {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if port.drains.Load() != 1 {
		t.Fatalf("writeNotify drains=%d want 1", port.drains.Load())
	}
	if !gotOK.Load() {
		t.Fatalf("expected ok notify; err=%v", gotErr.Load())
	}
}

func TestWriteNotifyRequiresHandler(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	tx := newUartPeriph("uartTxNoH", 95, "pipe-a", 14, 15, 16, 17)
	if err := tx.activateWithPort(pipePort{a}); err != nil {
		t.Fatal(err)
	}
	defer tx.release()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if !strings.Contains(fmt.Sprint(r), "onWriteNotify") {
			t.Fatalf("panic=%v", r)
		}
	}()
	tx.writeNotify([]byte("x"))
}
