package dsh

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/config"
	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/runner"
)

const SupportedVersion = "0.1.0-rc.6"

func Install(ctx context.Context, run runner.Runner, env Environment, paths config.Paths, version string) (string, error) {
	if version != SupportedVersion {
		return "", fmt.Errorf("unsupported dsh_version: expected %s", SupportedVersion)
	}
	if err := config.EnsurePrivateDir(paths.InstallDir); err != nil {
		return "", err
	}
	packageName := "@deepseek-ai/dsh@" + version
	args := []string{"install", "--prefix", paths.InstallDir, "--no-audit", "--no-fund", "--save-exact", packageName}
	if _, err := run.Run(ctx, env.NPMPath, args, paths.InstallDir, os.Stdout, os.Stderr); err != nil {
		return "", fmt.Errorf("npm install %s: %w", packageName, err)
	}
	bin := filepath.Join(paths.InstallDir, "node_modules", "@deepseek-ai", "dsh", "lib", "bin.js")
	if info, err := os.Stat(bin); err != nil || info.IsDir() {
		return "", errors.New("installed DSH CLI entry is missing")
	}
	verify, err := run.Run(ctx, env.NodePath, []string{bin, "--version"}, paths.InstallDir, nil, nil)
	if err != nil {
		return "", fmt.Errorf("verify installed DSH: %w", err)
	}
	if stringTrim(verify.Stdout) != version {
		return "", fmt.Errorf("installed DSH version mismatch: expected %s, got %s", version, stringTrim(verify.Stdout))
	}
	return bin, nil
}

func stringTrim(value string) string {
	for len(value) > 0 && (value[0] == ' ' || value[0] == '\r' || value[0] == '\n' || value[0] == '\t') {
		value = value[1:]
	}
	for len(value) > 0 {
		last := value[len(value)-1]
		if last != ' ' && last != '\r' && last != '\n' && last != '\t' {
			break
		}
		value = value[:len(value)-1]
	}
	return value
}
