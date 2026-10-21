package finalpi

import (
	"net"
	"testing"
	"time"
)

func TestReadWithTimeoutTimesOut(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 1)
		_, err := ReadWithTimeout(c1, c1, buf, 15*time.Millisecond)
		if err == nil {
			t.Error("expected timeout error")
		}
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("ReadWithTimeout did not return")
	}
}

func TestMintEventNil(t *testing.T) {
	if MintEvent(nil, 0) != nil {
		t.Fatal("nil handler should yield nil mint fn")
	}
}
