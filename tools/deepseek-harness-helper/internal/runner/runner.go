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

// TerminalRunner executes a fixed command with the Helper's console handles
// attached directly. Some Windows tools reject captured pipe handles even when
// their output is forwarded to the same terminal.
type TerminalRunner interface {
	RunInTerminal(ctx context.Context, executable string, args []string, dir string) (Result, error)
}
