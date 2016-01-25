package shell

import (
	"bytes"
	"time"
	"errors"
	"os/exec"
	"github.com/cretz/systrument/context"
	"github.com/cretz/systrument/util"
	"io"
)

var (
	ErrTimeout = errors.New("Timed out")
)

func Run(ctx *context.Context, name string, args ...string) error {
	return WrapCommandOutput(ctx, exec.Command(name, args...)).Run()
}

func Output(ctx *context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	var b bytes.Buffer
	cmd.Stdout = &b
	err := WrapCommandOutput(ctx, cmd).Run()
	return b.Bytes(), err
}

func CombinedOutput(ctx *context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	err := WrapCommandOutput(ctx, cmd).Run()
	return b.Bytes(), err
}

func WrapCommandOutput(ctx *context.Context, cmd *exec.Cmd) *exec.Cmd {
	// If we are verbose, we want to wrap stdout/stderr to log writes
	if ctx.DebugEnabled() {
		debugOutWriter := util.NewDebugLogWriter("SHELL OUT:", ctx)
		if cmd.Stdout != nil {
			cmd.Stdout = io.MultiWriter(cmd.Stdout, debugOutWriter)
		} else {
			cmd.Stdout = debugOutWriter
		}
		debugErrWriter := util.NewDebugLogWriter("SHELL ERR:", ctx)
		if cmd.Stderr != nil {
			cmd.Stderr = io.MultiWriter(cmd.Stderr, debugErrWriter)
		} else {
			cmd.Stderr = debugErrWriter
		}
	}
	return cmd
}

func WaitTimeout(cmd *exec.Cmd, dur time.Duration) error {
	c := make(chan error, 1)
	go func() { c <- cmd.Wait() }()
	select {
	case err := <- c: return err
	case <-time.After(dur):
		cmd.Process.Kill()
		return ErrTimeout
	}
}
