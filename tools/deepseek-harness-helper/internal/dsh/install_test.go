package dsh

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/config"
	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/runner"
)

type installRunner struct {
	calls [][]string
}

func (r *installRunner) Run(_ context.Context, executable string, args []string, _ string, _, _ io.Writer) (runner.Result, error) {
	r.calls = append(r.calls, append([]string{executable}, args...))
	if len(args) > 0 && args[0] == "install" {
		prefix := args[2]
		bin := filepath.Join(prefix, "node_modules", "@deepseek-ai", "dsh", "lib", "bin.js")
		if err := os.MkdirAll(filepath.Dir(bin), 0o700); err != nil {
			return runner.Result{}, err
		}
		if err := os.WriteFile(bin, []byte("test"), 0o600); err != nil {
			return runner.Result{}, err
		}
		return runner.Result{}, nil
	}
	return runner.Result{Stdout: "0.1.0-rc.6\n"}, nil
}

func TestInstallUsesFixedArguments(t *testing.T) {
	t.Parallel()
	paths := config.PathsFor(t.TempDir())
	run := &installRunner{}
	_, err := Install(context.Background(), run, Environment{NodePath: "node", NPMPath: "npm"}, paths, "0.1.0-rc.6", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"npm", "install", "--prefix", paths.InstallDir, "--no-audit", "--no-fund", "--save-exact", "@deepseek-ai/dsh@0.1.0-rc.6"}
	if !reflect.DeepEqual(run.calls[0], want) {
		t.Fatalf("npm argv = %#v, want %#v", run.calls[0], want)
	}
	for _, version := range []string{"0.1.0-rc.7", "latest; calc.exe"} {
		if _, err := Install(context.Background(), run, Environment{NodePath: "node", NPMPath: "npm"}, paths, version, nil, nil); err == nil {
			t.Fatalf("expected unsupported version rejection for %q", version)
		}
	}
}

func TestInstallSkipsMatchingPrivateRuntime(t *testing.T) {
	t.Parallel()
	paths := config.PathsFor(t.TempDir())
	bin := filepath.Join(paths.InstallDir, "node_modules", "@deepseek-ai", "dsh", "lib", "bin.js")
	if err := os.MkdirAll(filepath.Dir(bin), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bin, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	run := &installRunner{}
	var output bytes.Buffer
	got, err := Install(
		context.Background(), run, Environment{NodePath: "node", NPMPath: "npm"}, paths,
		SupportedVersion, &output, &output,
	)
	if err != nil {
		t.Fatal(err)
	}
	if got != bin {
		t.Fatalf("Install() = %q, want %q", got, bin)
	}
	if len(run.calls) != 1 || run.calls[0][0] != "node" {
		t.Fatalf("calls = %#v, want one version check", run.calls)
	}
	if !strings.Contains(output.String(), "already installed; skipping npm install") {
		t.Fatalf("output = %q", output.String())
	}
}
