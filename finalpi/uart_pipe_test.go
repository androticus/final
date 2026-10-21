package finalpi

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/androticus/final/finalembed"
)

// pipePort adapts net.Pipe for host tests: same Read/Write/deadline as Conn; Drain is a no-op.
type pipePort struct{ net.Conn }

func (p pipePort) Drain() error { return nil }

func TestUartPipeLoopbackReadWrite(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()

	tx := newUartPeriph("uartPipeTx", 90, "pipe-a", 14, 15, 16, 17)
	rx := newUartPeriph("uartPipeRx", 91, "pipe-b", 4, 5, 6, 7)
	if err := tx.activateWithPort(pipePort{a}); err != nil {
		t.Fatal(err)
	}
	if err := rx.activateWithPort(pipePort{b}); err != nil {
		t.Fatal(err)
	}
	defer tx.release()
	defer rx.release()

	payload := []byte("hello\n")
	if err := tx.writeBytes(payload); err != nil {
		t.Fatal(err)
	}

	r := rx.NewReader()
	r.timeoutMs = 500
	r.untilByte = int('\n')
	it := r.GetIter()
	var got []byte
	for it.GetNext() {
		got = append(got, byte(it.Current().(finalembed.Int)))
	}
	if string(got) != "hello\n" {
		t.Fatalf("got %q want %q (n=%d timedOut=%v hitDelim=%v err=%q)",
			got, "hello\n", r.n, r.timedOut, r.hitDelim, r.errStr)
	}
	if !r.hitDelim {
		t.Fatal("expected hitDelim")
	}
	if r.timedOut {
		t.Fatal("unexpected timedOut")
	}
	if r.n != len(payload) {
		t.Fatalf("n=%d want %d", r.n, len(payload))
	}
}

func TestUartReaderIdleTimeout(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()

	rx := newUartPeriph("uartPipeIdle", 92, "pipe-idle", 4, 5, 6, 7)
	if err := rx.activateWithPort(pipePort{b}); err != nil {
		t.Fatal(err)
	}
	defer rx.release()
	_ = a // peer kept open; no bytes written

	r := rx.NewReader()
	r.timeoutMs = 80
	r.maxN = 8
	start := time.Now()
	it := r.GetIter()
	if it.GetNext() {
		t.Fatal("expected no bytes")
	}
	if !r.timedOut {
		t.Fatal("expected timedOut")
	}
	if time.Since(start) > 500*time.Millisecond {
		t.Fatalf("idle wait too long: %v", time.Since(start))
	}
}

func TestUartValidateConflictsWithGpio(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	defer c.Reset()

	c.gpio4.Enabled = true
	c.uart2.enabled = true
	err := c.Validate()
	if err == nil {
		t.Fatal("expected BCM conflict uart2 vs gpio4")
	}
	c.gpio4.Enabled = false
	c.uart2.enabled = false

	c.spi5.enabled = true
	c.pwm0.enabled = true
	err = c.Validate()
	if err == nil {
		t.Fatal("expected BCM conflict spi5 vs pwm0 (GPIO12)")
	}
	c.spi5.enabled = false
	c.pwm0.enabled = false

	c.gpio14.Enabled = true
	c.uart0.enabled = true
	err = c.Validate()
	if err == nil {
		t.Fatal("expected BCM conflict uart0 vs gpio14")
	}
	c.gpio14.Enabled = false
	c.uart0.enabled = false
}

func TestUartFlowHWOptPinsValidate(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	defer c.Reset()

	c.uart2.setFlowHW(true)
	opts := c.uart2.OptPins()
	if len(opts) != 2 || opts[0].BCM != 6 || opts[1].BCM != 7 {
		t.Fatalf("uart2 flowHW optPins=%v want CTS6 RTS7", opts)
	}
	c.uart2.enabled = true
	c.gpio6.Enabled = true
	if err := c.Validate(); err == nil {
		t.Fatal("expected conflict uart2 flowHW vs gpio6")
	}
	c.gpio6.Enabled = false
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}

	c.uart2.enabled = false
	c.uart2.setFlowHW(false)
	if len(c.uart2.OptPins()) != 0 {
		t.Fatal("flowHW false should clear optPins")
	}
	c.uart2.enabled = true
	c.gpio6.Enabled = true
	if err := c.Validate(); err != nil {
		t.Fatalf("without flowHW, gpio6 should be free: %v", err)
	}

	c.Reset()
	c.uart2.setFlowHW(true)
	c.uart2.enabled = true
	c.Reset()
	if c.uart2.flowHW || len(c.uart2.OptPins()) != 0 {
		t.Fatal("Reset should reset flowHW and optPins")
	}
}

func TestGpioCtrlSupportedPeriphInstances(t *testing.T) {
	c := GpioCtrlInst()
	if c.uart0 == nil || c.uart2 == nil || c.uart3 == nil {
		t.Fatal("missing uart instance")
	}
	if c.pwm0 == nil || c.pwm1 == nil {
		t.Fatal("missing pwm instance")
	}
	want0 := uart0DefaultPath(piSupport)
	if v := os.Getenv("FINAL_UART0"); v != "" {
		want0 = v
	}
	if c.uart0.path != want0 || c.uart2.path != "/dev/ttyAMA2" || c.uart3.path != "/dev/ttyAMA3" {
		t.Fatalf("uart paths: 0=%s (want %s) 2=%s 3=%s", c.uart0.path, want0, c.uart2.path, c.uart3.path)
	}
	if c.pwm0.bcm != 12 || c.pwm1.bcm != 13 {
		t.Fatalf("pwm0/1 bcm: %d %d", c.pwm0.bcm, c.pwm1.bcm)
	}
	// uart1 (HAT ID) and uart4 (spi5 clash) intentionally absent.
	for _, p := range c.periphs {
		switch p.Name() {
		case "uart1", "uart4":
			t.Fatalf("%s must not be in stub set", p.Name())
		}
	}
}

func TestUartSoftEnableMissingDevice(t *testing.T) {
	u := newUartPeriph("uartMissing", 93, "/dev/ttyAMA_does_not_exist_phase6", 14, 15, 16, 17)
	err := u.activate()
	if err == nil {
		u.release()
		t.Fatal("expected open error")
	}
}
