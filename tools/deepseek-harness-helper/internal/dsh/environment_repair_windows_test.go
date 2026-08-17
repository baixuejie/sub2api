//go:build windows

package dsh

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/runner"
)

type terminalRecordingRunner struct {
	normalCalls   int
	terminalCalls int
}

func (r *terminalRecordingRunner) Run(context.Context, string, []string, string, io.Writer, io.Writer) (runner.Result, error) {
	r.normalCalls++
	return runner.Result{}, nil
}

func (r *terminalRecordingRunner) RunInTerminal(context.Context, string, []string, string) (runner.Result, error) {
	r.terminalCalls++
	return runner.Result{}, nil
}

func TestNVMCandidatesDoNotContainEmptyRoots(t *testing.T) {
	t.Parallel()
	for _, candidate := range nvmCandidates() {
		if candidate == "" {
			t.Fatal("nvmCandidates() returned an empty path")
		}
	}
}

func TestNVMCommandsUseTerminalRunner(t *testing.T) {
	t.Parallel()
	run := &terminalRecordingRunner{}
	if _, err := runNVMInTerminal(context.Background(), run, `C:\nvm\nvm.exe`, []string{"install", "lts"}, io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	if run.terminalCalls != 1 || run.normalCalls != 0 {
		t.Fatalf("terminal calls = %d, normal calls = %d", run.terminalCalls, run.normalCalls)
	}
}

func TestNVMWingetInstallUsesOfficialPackage(t *testing.T) {
	t.Parallel()
	want := []string{
		"install", "--id", "CoreyButler.NVMforWindows", "--exact", "--source", "winget", "--silent",
		"--accept-package-agreements", "--accept-source-agreements", "--disable-interactivity",
	}
	if got := nvmWingetInstallArgs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("nvmWingetInstallArgs() = %#v, want %#v", got, want)
	}
}

func TestValidateNVMSymlinkRejectsPhysicalDirectory(t *testing.T) {
	t.Parallel()
	physical := filepath.Join(t.TempDir(), "nodejs")
	if err := os.Mkdir(physical, 0o700); err != nil {
		t.Fatal(err)
	}
	err := validateNVMSymlink(physical)
	if err == nil || !strings.Contains(err.Error(), "existing physical directory") {
		t.Fatalf("validateNVMSymlink() error = %v", err)
	}
	if err := validateNVMSymlink(filepath.Join(t.TempDir(), "missing")); err != nil {
		t.Fatalf("validateNVMSymlink(missing) error = %v", err)
	}
}
