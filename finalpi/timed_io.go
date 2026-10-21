package finalpi

import (
	"io"
	"time"
)

// ReadDeadliner is satisfied by types that support SetReadDeadline (e.g. net.Conn, serial).
type ReadDeadliner interface {
	SetReadDeadline(t time.Time) error
}

// ReadWithTimeout sets a read deadline then reads. timeout <= 0 clears the deadline
// (blocking read — prefer cancel via Worker for long waits).
func ReadWithTimeout(r io.Reader, d ReadDeadliner, buf []byte, timeout time.Duration) (int, error) {
	if d != nil {
		var deadline time.Time
		if timeout > 0 {
			deadline = time.Now().Add(timeout)
		}
		if err := d.SetReadDeadline(deadline); err != nil {
			return 0, err
		}
	}
	return r.Read(buf)
}
