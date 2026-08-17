package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/bootstrap"
	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/config"
	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/platform"
)

var version = "dev"

func main() {
	args := os.Args[1:]
	err := run(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "deepseek-harness-helper:", err)
	}
	if shouldWaitForExit(args, err) {
		platform.WaitForExitPrompt()
	}
	if err != nil {
		os.Exit(1)
	}
}

func shouldWaitForExit(args []string, runErr error) bool {
	if len(args) == 0 {
		return true
	}
	return runErr != nil && len(args) == 1 && strings.HasPrefix(strings.ToLower(strings.TrimSpace(args[0])), "sub2api-harness://")
}

func run(args []string) error {
	if len(args) == 1 && (args[0] == "--version" || args[0] == "version") {
		fmt.Println(version)
		return nil
	}
	if needsInteractiveConsole(args) {
		if err := platform.EnsureInteractiveConsole(); err != nil {
			return fmt.Errorf("prepare installation console: %w", err)
		}
	}
	if len(args) == 0 || (len(args) == 1 && args[0] == "register-protocol") {
		return installHelper()
	}
	if len(args) != 1 {
		return errors.New("usage: deepseek-harness-helper <sub2api-harness://bootstrap?...> | register-protocol | --version")
	}
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	ctx, timeout := context.WithTimeout(ctx, 30*time.Minute)
	defer timeout()
	workflow := bootstrap.Workflow{Client: bootstrap.NewClient(), Paths: paths, Output: os.Stdout, WarningOutput: os.Stderr, HelperVersion: version}
	harnessURL, err := workflow.Run(ctx, args[0])
	if err != nil {
		return err
	}
	if harnessURL == "" {
		return nil
	}
	return openBrowser(harnessURL)
}

func needsInteractiveConsole(args []string) bool {
	if len(args) == 0 {
		return true
	}
	return len(args) == 1 && strings.HasPrefix(strings.ToLower(strings.TrimSpace(args[0])), "sub2api-harness://")
}

func installHelper() error {
	fmt.Fprintln(os.Stdout, "[1/3] Resolving the per-user installation directory...")
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "[2/3] Copying Helper and registering the sub2api-harness protocol...")
	installedPath, err := platform.InstallSelf(executable, paths.DataDir)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "[3/3] Installation completed successfully.")
	fmt.Fprintln(os.Stdout, "Installed path:", installedPath)
	fmt.Fprintln(os.Stdout, "Return to Sub2API and start the selected local tool.")
	return nil
}

func openBrowser(rawURL string) error {
	var command string
	var args []string
	switch runtime.GOOS {
	case "windows":
		command = "rundll32.exe"
		args = []string{"url.dll,FileProtocolHandler", rawURL}
	case "darwin":
		command = "open"
		args = []string{rawURL}
	default:
		command = "xdg-open"
		args = []string{rawURL}
	}
	path, err := exec.LookPath(command)
	if err != nil {
		return fmt.Errorf("open browser for %s: %w; open the URL manually", rawURL, err)
	}
	cmd := exec.Command(path, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}
