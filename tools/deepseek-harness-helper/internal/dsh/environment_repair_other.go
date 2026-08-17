//go:build !windows

package dsh

import (
	"context"
	"fmt"
	"io"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/runner"
)

func repairEnvironment(_ context.Context, _ runner.Runner, _, _ io.Writer, cause error) (Environment, error) {
	return Environment{}, fmt.Errorf("automatic Node.js repair is currently supported only on Windows: %w", cause)
}
