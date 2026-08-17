package dsh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/runner"
)

var versionPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:[-+].*)?$`)

var MinimumNodeVersion = [3]int{22, 19, 0}

type Environment struct {
	NodePath    string
	NodeVersion string
	NPMPath     string
	NPMVersion  string
}

func EnsureEnvironment(ctx context.Context, run runner.Runner, stdout, stderr io.Writer) (Environment, error) {
	environment, err := CheckEnvironment(ctx, run)
	if err == nil {
		writeEnvironmentReady(stdout, environment)
		return environment, nil
	}
	if stdout != nil {
		_, _ = fmt.Fprintf(stdout, "Node.js environment requires repair: %v\n", err)
	}
	environment, err = repairEnvironment(ctx, run, stdout, stderr, err)
	if err != nil {
		return Environment{}, err
	}
	writeEnvironmentReady(stdout, environment)
	return environment, nil
}

func writeEnvironmentReady(output io.Writer, environment Environment) {
	if output == nil {
		return
	}
	_, _ = fmt.Fprintf(output, "Node.js %s and npm %s are ready.\n", environment.NodeVersion, environment.NPMVersion)
}

func CheckEnvironment(ctx context.Context, run runner.Runner) (Environment, error) {
	if run == nil {
		return Environment{}, errors.New("runner is required")
	}
	node, err := exec.LookPath("node")
	if err != nil {
		return Environment{}, errors.New("Node.js is not installed or not on PATH")
	}
	npm, err := exec.LookPath("npm")
	if err != nil {
		return Environment{}, errors.New("npm is not installed or not on PATH")
	}
	nodeResult, err := run.Run(ctx, node, []string{"--version"}, "", nil, nil)
	if err != nil {
		return Environment{}, fmt.Errorf("run node --version: %w", err)
	}
	nodeVersion := strings.TrimSpace(nodeResult.Stdout)
	if !AtLeastNode(nodeVersion, MinimumNodeVersion) {
		return Environment{}, fmt.Errorf("Node.js >= %d.%d.%d is required; found %s", MinimumNodeVersion[0], MinimumNodeVersion[1], MinimumNodeVersion[2], nodeVersion)
	}
	npmResult, err := run.Run(ctx, npm, []string{"--version"}, "", nil, nil)
	if err != nil {
		return Environment{}, fmt.Errorf("run npm --version: %w", err)
	}
	return Environment{NodePath: node, NodeVersion: nodeVersion, NPMPath: npm, NPMVersion: strings.TrimSpace(npmResult.Stdout)}, nil
}

func AtLeastNode(raw string, minimum [3]int) bool {
	matches := versionPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if matches == nil {
		return false
	}
	for i := 0; i < 3; i++ {
		value, err := strconv.Atoi(matches[i+1])
		if err != nil {
			return false
		}
		if value > minimum[i] {
			return true
		}
		if value < minimum[i] {
			return false
		}
	}
	return true
}
