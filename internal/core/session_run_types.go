package core

import (
	"errors"
	"fmt"
)

// RunStatus classifies programmatic Session.Run completion.
type RunStatus int

const (
	RunCompleted RunStatus = iota
	RunFaulted
	RunInterpretError
)

// InterpretError is an expected interpret/compile failure (unknown word, etc.).
type InterpretError struct {
	Message string
}

func (e *InterpretError) Error() string {
	if e == nil {
		return "<nil InterpretError>"
	}
	return e.Message
}

// RunResult is the completion value from Session.Run / RunHandle.Wait.
type RunResult struct {
	Status RunStatus
	Fault  *RunFault
	Err    error
	Output string // set when out was RunOutputCapture
}

// Unwrap returns Fault or Err for idiomatic Go error handling.
func (r RunResult) Unwrap() error {
	if r.Fault != nil {
		return r.Fault
	}
	return r.Err
}

var (
	ErrSessionClosed   = errors.New("session: closed")
	ErrEmptyScript     = errors.New("session: empty script")
	ErrNilReader       = errors.New("session: nil reader")
	ErrNilWriter       = errors.New("session: nil writer")
	ErrInvalidRunParam = errors.New("session: invalid run parameter")
)

func interpretError(err error) RunResult {
	if err == nil {
		return RunResult{Status: RunCompleted}
	}
	if fault, ok := err.(*RunFault); ok {
		return RunResult{Status: RunFaulted, Fault: fault}
	}
	return RunResult{
		Status: RunInterpretError,
		Err:    &InterpretError{Message: err.Error()},
	}
}

func faultResult(t *Task, rp *RunProvider, recovered any) RunResult {
	return RunResult{
		Status: RunFaulted,
		Fault:  captureRunFault(t, rp, recovered, rp.LastOp, rp.LastIP),
	}
}

func validateRunInput(in RunInput) error {
	switch v := in.(type) {
	case nil:
		return nil
	case RunInputString:
		if v.Script == "" {
			return ErrEmptyScript
		}
	case RunInputReader:
		if v.Source == nil {
			return ErrNilReader
		}
	case RunInputPipe:
		if v.vmReader == nil {
			return fmt.Errorf("session: invalid input pipe")
		}
	default:
		return ErrInvalidRunParam
	}
	return nil
}

func validateRunOutput(out RunOutput) error {
	if out == nil {
		return nil
	}
	switch v := out.(type) {
	case RunOutputCapture:
		return nil
	case RunOutputWriter:
		if v.Sink == nil {
			return ErrNilWriter
		}
	case RunOutputPipe:
		if v.vmWriter == nil {
			return fmt.Errorf("session: invalid output pipe")
		}
	default:
		return ErrInvalidRunParam
	}
	return nil
}

func runNeedsAsync(in RunInput, out RunOutput) bool {
	if _, ok := in.(RunInputPipe); ok {
		return true
	}
	if _, ok := out.(RunOutputPipe); ok {
		return true
	}
	return false
}
