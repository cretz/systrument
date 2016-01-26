package shell

import (
	"bytes"
	"errors"
	"github.com/cretz/systrument/context"
	"github.com/cretz/systrument/util"
	"io"
	"os/exec"
	"regexp"
	"time"
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
		AppendStdoutWriter(cmd, util.NewDebugLogWriter("SHELL OUT:", ctx))
		AppendStderrWriter(cmd, util.NewDebugLogWriter("SHELL ERR:", ctx))
	}
	return cmd
}

func WaitTimeout(cmd *exec.Cmd, dur time.Duration) error {
	c := make(chan error, 1)
	go func() { c <- cmd.Wait() }()
	select {
	case err := <-c:
		return err
	case <-time.After(dur):
		cmd.Process.Kill()
		return ErrTimeout
	}
}

func AppendStdoutWriter(cmd *exec.Cmd, writer io.Writer) {
	if cmd.Stdout == nil {
		cmd.Stdout = writer
	} else {
		cmd.Stdout = io.MultiWriter(cmd.Stdout, writer)
	}
}

func AppendStderrWriter(cmd *exec.Cmd, writer io.Writer) {
	if cmd.Stderr == nil {
		cmd.Stderr = writer
	} else {
		cmd.Stderr = io.MultiWriter(cmd.Stderr, writer)
	}
}

var SudoPasswordPromptMatch = regexp.MustCompile("\\[sudo\\] password for .*:")

func SudoCommand(password string, name string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command("sudo", append([]string{"-S", name}, args...)...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	sudoHandler := util.NewExpectListener(stdin, SudoPasswordPromptMatch, password+"\n")
	AppendStdoutWriter(cmd, sudoHandler)
	AppendStderrWriter(cmd, sudoHandler)
	return cmd, nil
}
