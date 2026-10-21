//go:build !linux

package finalpi

import (
	"fmt"
	"os"
)

func openI2C(path string) (*os.File, error) {
	return nil, fmt.Errorf("I2C %s requires Linux (this binary has no i2c-dev backend)", path)
}

func i2cWrite(_ *os.File, _ uint16, _ []byte) error {
	return fmt.Errorf("I2C: no backend")
}

func i2cRead(_ *os.File, _ uint16, _ int) ([]byte, error) {
	return nil, fmt.Errorf("I2C: no backend")
}

func i2cTx(_ *os.File, _ uint16, _ []byte, _ int) ([]byte, error) {
	return nil, fmt.Errorf("I2C: no backend")
}
