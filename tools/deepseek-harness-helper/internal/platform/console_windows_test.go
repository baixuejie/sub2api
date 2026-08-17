//go:build windows

package platform

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	consoleChildEnvironment  = "SUB2API_HELPER_CONSOLE_TEST_CHILD"
	consoleResultEnvironment = "SUB2API_HELPER_CONSOLE_TEST_RESULT"
	consoleNVMEnvironment    = "SUB2API_HELPER_CONSOLE_TEST_NVM"
)

func TestEnsureInteractiveConsoleBindsCharacterDevice(t *testing.T) {
	resultFile := filepath.Join(t.TempDir(), "console-result.txt")
	if os.Getenv(consoleChildEnvironment) == "1" {
		if err := EnsureInteractiveConsole(); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stdout.Stat()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(os.Getenv(consoleResultEnvironment), []byte(info.Mode().String()), 0o600); err != nil {
			t.Fatal(err)
		}
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestEnsureInteractiveConsoleBindsCharacterDevice$")
	cmd.Env = append(os.Environ(),
		consoleChildEnvironment+"=1",
		consoleResultEnvironment+"="+resultFile,
	)
	var captured bytes.Buffer
	cmd.Stdout = &captured
	cmd.Stderr = &captured
	if err := cmd.Run(); err != nil {
		t.Fatalf("console child failed: %v\n%s", err, captured.String())
	}

	mode, err := os.ReadFile(resultFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mode), "c") {
		t.Fatalf("console stdout mode = %q, want character device", mode)
	}
}

func TestBoundConsoleSupportsNVMForWindows(t *testing.T) {
	nvmPath, err := exec.LookPath("nvm.exe")
	if err != nil {
		t.Skip("NVM for Windows is not installed")
	}
	resultFile := filepath.Join(t.TempDir(), "nvm-result.txt")
	if os.Getenv(consoleNVMEnvironment) == "1" {
		if err := EnsureInteractiveConsole(); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, nvmPath, "version")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("nvm version failed with bound console: %v", err)
		}
		if err := os.WriteFile(os.Getenv(consoleResultEnvironment), []byte("ok"), 0o600); err != nil {
			t.Fatal(err)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestBoundConsoleSupportsNVMForWindows$")
	cmd.Env = append(os.Environ(),
		consoleNVMEnvironment+"=1",
		consoleResultEnvironment+"="+resultFile,
	)
	var captured bytes.Buffer
	cmd.Stdout = &captured
	cmd.Stderr = &captured
	if err := cmd.Run(); err != nil {
		t.Fatalf("NVM console child failed: %v\n%s", err, captured.String())
	}
	result, err := os.ReadFile(resultFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != "ok" {
		t.Fatalf("NVM console result = %q, want ok", result)
	}
}
