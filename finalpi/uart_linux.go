//go:build linux

package finalpi

import (
	"fmt"
	"time"

	"go.bug.st/serial"
)

type serialPortWrap struct {
	p serial.Port
}

func (w *serialPortWrap) Read(b []byte) (int, error)  { return w.p.Read(b) }
func (w *serialPortWrap) Write(b []byte) (int, error) { return w.p.Write(b) }
func (w *serialPortWrap) Close() error                { return w.p.Close() }
func (w *serialPortWrap) Drain() error                { return w.p.Drain() }

func (w *serialPortWrap) SetReadDeadline(t time.Time) error {
	if t.IsZero() {
		return w.p.SetReadTimeout(serial.NoTimeout)
	}
	d := time.Until(t)
	if d < 0 {
		d = 0
	}
	return w.p.SetReadTimeout(d)
}

func openUartPort(path string, baud int, dataBits int, stopBits int, parity UartParity, flowHW bool) (uartPort, error) {
	if err := rejectUnsupportedUartBaud(path, baud); err != nil {
		return nil, err
	}
	mode := &serial.Mode{
		BaudRate: baud,
		DataBits: dataBits,
	}
	switch stopBits {
	case 2:
		mode.StopBits = serial.TwoStopBits
	default:
		mode.StopBits = serial.OneStopBit
	}
	switch parity {
	case UartParityOdd:
		mode.Parity = serial.OddParity
	case UartParityEven:
		mode.Parity = serial.EvenParity
	default:
		mode.Parity = serial.NoParity
	}
	p, err := serial.Open(path, mode)
	if err != nil {
		return nil, fmt.Errorf("UART open %s: %w (overlay / enable_uart? dialout group?)", path, err)
	}
	_ = p.SetReadTimeout(serial.NoTimeout)
	if flowHW {
		// HW RTS/CTS needs overlay ctsrts=on; Mode has no separate toggle here.
		_ = flowHW
	}
	return &serialPortWrap{p: p}, nil
}
