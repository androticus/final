package finalpi

import (
	"io"
	"time"
)

// uartPort is the activate-time Linux tty handle (go.bug.st/serial).
// Read deadline via SetReadTimeout; TX completion via Drain (tcdrain).
// No write deadline — TX timeouts use a goroutine + timer around Write/Drain.
type uartPort interface {
	io.ReadWriteCloser
	SetReadDeadline(t time.Time) error
	Drain() error // writeNotify only; plain write does not drain
}

// rxQueue fans bytes from the RX worker to UartReader (Phase 5 spine).
type rxQueue struct {
	ch     chan byte
	errCh  chan error
	closed chan struct{}
}

func newRxQueue(cap int) *rxQueue {
	if cap < 64 {
		cap = 256
	}
	return &rxQueue{
		ch:     make(chan byte, cap),
		errCh:  make(chan error, 1),
		closed: make(chan struct{}),
	}
}

func (q *rxQueue) push(b []byte) {
	for _, c := range b {
		select {
		case <-q.closed:
			return
		case q.ch <- c:
		}
	}
}

// hasData reports whether at least one byte is queued (non-blocking).
func (q *rxQueue) hasData() bool {
	return len(q.ch) > 0
}

func (q *rxQueue) close(err error) {
	select {
	case <-q.closed:
		return
	default:
		close(q.closed)
	}
	if err != nil {
		select {
		case q.errCh <- err:
		default:
		}
	}
}

func (q *rxQueue) popByte(cancel <-chan struct{}, idle time.Duration) (b byte, ok bool, timedOut bool, err error) {
	if idle <= 0 {
		idle = time.Hour // treat non-positive as "long" rather than busy-spin
	}
	t := time.NewTimer(idle)
	defer t.Stop()
	select {
	case <-cancel:
		return 0, false, false, errCanceled
	case <-q.closed:
		select {
		case b := <-q.ch:
			return b, true, false, nil
		default:
		}
		select {
		case err := <-q.errCh:
			return 0, false, false, err
		default:
			return 0, false, false, nil
		}
	case b := <-q.ch:
		return b, true, false, nil
	case <-t.C:
		return 0, false, true, nil
	}
}

var errCanceled = errString("uart: canceled")

type errString string

func (e errString) Error() string { return string(e) }
