//go:build linux

package finalpi

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/androticus/final/finalembed"
)

func requireUartPair(t *testing.T) {
	t.Helper()
	for _, path := range []string{"/dev/ttyAMA0", "/dev/ttyAMA2"} {
		f, err := os.OpenFile(path, os.O_RDWR, 0)
		if err != nil {
			t.Skipf("UART device %s unavailable (dtoverlay=uart0-pi5 / uart2-pi5?): %v", path, err)
		}
		_ = f.Close()
	}
}

// ensureUartPinmux restores overlay alt functions if a GPIO test stole the lines.
// uart0: GPIO14/15 a4; uart2: GPIO4/5 a2 (Pi 5 RP1).
func ensureUartPinmux(t *testing.T) {
	t.Helper()
	sets := [][]string{
		{"4", "a2"},
		{"5", "a2"},
		{"14", "a4"},
		{"15", "a4"},
	}
	for _, s := range sets {
		cmd := exec.Command("pinctrl", "set", s[0], s[1])
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Logf("pinctrl set %s %s: %v (%s)", s[0], s[1], err, out)
		}
	}
}

// Single-port write smoke (Pi 5 or Zero 2): open header UART, queue bytes, no loopback peer required.
func TestUartHardwareWriteSmoke(t *testing.T) {
	path := uart0DefaultPath(piSupport)
	if v := os.Getenv("FINAL_UART0"); v != "" {
		path = v
	}
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Skipf("UART %s unavailable: %v", path, err)
	}
	_ = f.Close()

	c := GpioCtrlInst()
	c.Reset()
	defer c.Reset()

	u := c.uart0
	u.setBaud(115200)
	if err := u.SetEnabled(true); err != nil {
		t.Fatalf("enable %s: %v", path, err)
	}
	payload := []byte("final-uart-tx\n")
	if err := u.writeBytes(payload); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := u.SetEnabled(false); err != nil {
		t.Fatalf("disable: %v", err)
	}
}

// Hardware loopback: breadboard uart0 TX↔uart2 RX and uart0 RX↔uart2 TX.
// Pi Zero 2 has no uart2 — skip (cannot loopback on a single header UART).
func TestUartHardwareLoopback0to2(t *testing.T) {
	if piSupport != PiSupportPi5 {
		t.Skipf("uart0↔uart2 loopback requires Pi 5 (piSupport=%d)", piSupport)
	}
	requireUartPair(t)
	ensureUartPinmux(t)
	c := GpioCtrlInst()
	c.Reset()
	defer c.Reset()

	c.uart0.setBaud(115200)
	c.uart2.setBaud(115200)
	if err := c.uart0.SetEnabled(true); err != nil {
		t.Fatalf("uart0 enable: %v", err)
	}
	if err := c.uart2.SetEnabled(true); err != nil {
		t.Fatalf("uart2 enable: %v", err)
	}

	payload := []byte("P6uart!\n")
	if err := c.uart0.writeBytes(payload); err != nil {
		t.Fatalf("uart0 write: %v", err)
	}

	r := c.uart2.NewReader()
	r.timeoutMs = 1000
	r.untilByte = int('\n')
	it := r.GetIter()
	var got []byte
	for it.GetNext() {
		got = append(got, byte(it.Current().(finalembed.Int)))
	}
	if string(got) != string(payload) {
		t.Fatalf("uart2 got %q want %q (n=%d timedOut=%v hitDelim=%v err=%q)",
			got, payload, r.n, r.timedOut, r.hitDelim, r.errStr)
	}

	// Reverse direction
	payload2 := []byte("rev\n")
	if err := c.uart2.writeBytes(payload2); err != nil {
		t.Fatalf("uart2 write: %v", err)
	}
	r2 := c.uart0.NewReader()
	r2.timeoutMs = 1000
	r2.untilByte = int('\n')
	it2 := r2.GetIter()
	got = got[:0]
	for it2.GetNext() {
		got = append(got, byte(it2.Current().(finalembed.Int)))
	}
	if string(got) != string(payload2) {
		t.Fatalf("uart0 got %q want %q", got, payload2)
	}

	// Reset must stop workers / close fds
	c.Reset()
	if c.uart0.Active() || c.uart2.Active() {
		t.Fatal("expected inactive after Reset")
	}
	time.Sleep(20 * time.Millisecond)
}
