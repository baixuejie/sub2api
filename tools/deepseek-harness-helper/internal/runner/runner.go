package runner

import (
	"context"
	"io"
)

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

type Runner interface {
	Run(ctx context.Context, executable string, args []string, dir string, stdout, stderr io.Writer) (Result, error)
}
