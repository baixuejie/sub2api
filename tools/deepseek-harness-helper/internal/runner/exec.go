package runner

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
)

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, executable string, args []string, dir string, stdout, stderr io.Writer) (Result, error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = dir
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(&outBuf, writerOrDiscard(stdout))
	cmd.Stderr = io.MultiWriter(&errBuf, writerOrDiscard(stderr))
	err := cmd.Run()
	result := Result{Stdout: outBuf.String(), Stderr: errBuf.String()}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		return result, err
	}
	return result, err
}

func (ExecRunner) RunInTerminal(ctx context.Context, executable string, args []string, dir string) (Result, error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	result := Result{}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		return result, err
	}
	return result, err
}

func writerOrDiscard(w io.Writer) io.Writer {
	if w == nil {
		return io.Discard
	}
	return w
}
