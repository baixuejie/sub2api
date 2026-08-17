package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/bootstrap"
	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/config"
	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/platform"
)

var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "deepseek-harness-helper:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 1 && (args[0] == "--version" || args[0] == "version") {
		fmt.Println(version)
		return nil
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
	ctx, timeout := context.WithTimeout(ctx, 10*time.Minute)
	defer timeout()
	workflow := bootstrap.Workflow{Client: bootstrap.NewClient(), Paths: paths, Output: os.Stdout, WarningOutput: os.Stderr}
	harnessURL, err := workflow.Run(ctx, args[0])
	if err != nil {
		return err
	}
	return openBrowser(harnessURL)
}

func installHelper() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}
	installedPath, err := platform.InstallSelf(executable, paths.DataDir)
	if err != nil {
		return err
	}
	fmt.Println("DeepSeek Harness Helper installed:", installedPath)
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
		return nil
	}
	cmd := exec.Command(path, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}
