//go:build !linux

package finalpi

import "fmt"

func openUartPort(path string, baud int, dataBits int, stopBits int, parity UartParity, flowHW bool) (uartPort, error) {
	_ = baud
	_ = dataBits
	_ = stopBits
	_ = parity
	_ = flowHW
	return nil, fmt.Errorf("UART %s requires Linux (this binary has no tty backend)", path)
}
