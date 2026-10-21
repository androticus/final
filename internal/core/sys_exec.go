package core

import (
	"context"
	"errors"
	"os/exec"
	"time"
)

const sysExecTimeoutErr = "timeout"

// sysExecContext returns a cancellable context. timeoutS <= 0 means no deadline.
func sysExecContext(timeoutS int) (context.Context, context.CancelFunc) {
	if timeoutS <= 0 {
		return context.WithCancel(context.Background())
	}
	return context.WithTimeout(context.Background(), time.Duration(timeoutS)*time.Second)
}

// runSysExecCombined runs program with argv args, merging stderr into the captured
// stdout text. Always returns any bytes already captured in out.
// Launch failure and timeout return err (modal false). Timeout err message is
// "timeout". A process that starts and exits nonzero returns its output, exit
// code, and nil err (modal true).
func runSysExecCombined(ctx context.Context, program string, args []string) (out string, code int, err error) {
	cmd := exec.CommandContext(ctx, program, args...)
	b, runErr := cmd.CombinedOutput()
	out = string(b)
	if runErr == nil {
		return out, 0, nil
	}
	if ctx.Err() == context.DeadlineExceeded || errors.Is(runErr, context.DeadlineExceeded) {
		return out, 0, errors.New(sysExecTimeoutErr)
	}
	var ee *exec.ExitError
	if errors.As(runErr, &ee) {
		return out, ee.ExitCode(), nil
	}
	return out, 0, runErr
}

// pushSysExecModal pushes ( out:s code:i true | out:s err:s false ).
func pushSysExecModal(ds *Stack[any], out string, code int, err error) {
	if err != nil {
		ds.Push(String(out))
		ds.Push(String(err.Error()))
		ds.Push(Bool(false))
		return
	}
	ds.Push(String(out))
	ds.Push(Int(code))
	ds.Push(Bool(true))
}
