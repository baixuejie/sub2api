//go:build windows

package dsh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/runner"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const nvmWingetPackageID = "CoreyButler.NVMforWindows"

func repairEnvironment(ctx context.Context, run runner.Runner, stdout, stderr io.Writer, cause error) (Environment, error) {
	if run == nil {
		return Environment{}, errors.New("runner is required")
	}
	if err := ensureNVMSymlinkAvailable(); err != nil {
		return Environment{}, err
	}
	nvmPath, err := ensureNVM(ctx, run, stdout, stderr, cause, findNVM)
	if err != nil {
		return Environment{}, err
	}
	refreshNVMEnvironment(nvmPath)

	if stdout != nil {
		_, _ = fmt.Fprintln(stdout, "Running: nvm install lts")
	}
	if _, err := runNVMInTerminal(ctx, run, nvmPath, []string{"install", "lts"}, stdout, stderr); err != nil {
		return Environment{}, fmt.Errorf("nvm install lts: %w", err)
	}
	if stdout != nil {
		_, _ = fmt.Fprintln(stdout, "Running: nvm use lts")
	}
	if _, err := runNVMInTerminal(ctx, run, nvmPath, []string{"use", "lts"}, stdout, stderr); err != nil {
		return Environment{}, fmt.Errorf("nvm use lts: %w", err)
	}
	refreshNVMEnvironment(nvmPath)
	environment, err := CheckEnvironment(ctx, run)
	if err != nil {
		return Environment{}, fmt.Errorf("verify Node.js after nvm use lts: %w", err)
	}
	return environment, nil
}

func ensureNVM(
	ctx context.Context,
	run runner.Runner,
	stdout, stderr io.Writer,
	cause error,
	finder func() (string, error),
) (string, error) {
	if nvmPath, err := finder(); err == nil {
		if stdout != nil {
			_, _ = fmt.Fprintln(stdout, "NVM for Windows is already installed; continuing with the existing installation.")
		}
		return nvmPath, nil
	}

	if stdout != nil {
		_, _ = fmt.Fprintln(stdout, "Node.js repair will install the official NVM for Windows package.")
		_, _ = fmt.Fprintln(stdout, "Windows may request elevation, and NVM will switch the active system Node.js version.")
	}
	wingetPath, err := exec.LookPath("winget.exe")
	if err != nil {
		return "", fmt.Errorf("repair Node.js after %v: winget is required to install NVM for Windows", cause)
	}
	if _, err := run.Run(ctx, wingetPath, nvmWingetInstallArgs(), "", stdout, stderr); err != nil {
		return "", fmt.Errorf("install latest NVM for Windows: %w", err)
	}
	nvmPath, err := finder()
	if err != nil {
		return "", errors.New("NVM for Windows installation completed but nvm.exe was not found; reopen Windows and retry")
	}
	return nvmPath, nil
}

func nvmWingetInstallArgs() []string {
	return []string{
		"install", "--id", nvmWingetPackageID, "--exact", "--source", "winget", "--silent",
		"--accept-package-agreements", "--accept-source-agreements", "--disable-interactivity",
	}
}

func runNVMInTerminal(
	ctx context.Context,
	run runner.Runner,
	executable string,
	args []string,
	stdout, stderr io.Writer,
) (runner.Result, error) {
	if terminal, ok := run.(runner.TerminalRunner); ok {
		return terminal.RunInTerminal(ctx, executable, args, "")
	}
	return run.Run(ctx, executable, args, "", stdout, stderr)
}

func findNVM() (string, error) {
	if path, err := exec.LookPath("nvm.exe"); err == nil {
		return filepath.Clean(path), nil
	}
	for _, candidate := range nvmCandidates() {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return filepath.Clean(candidate), nil
		}
	}
	return "", errors.New("nvm.exe was not found")
}

func nvmCandidates() []string {
	roots := []string{
		os.Getenv("NVM_HOME"),
		readWindowsEnvironment(registry.CURRENT_USER, `Environment`, "NVM_HOME"),
		readWindowsEnvironment(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, "NVM_HOME"),
		filepath.Join(os.Getenv("APPDATA"), "nvm"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "nvm"),
		filepath.Join(os.Getenv("ProgramFiles"), "nvm"),
	}
	seen := make(map[string]struct{}, len(roots))
	candidates := make([]string, 0, len(roots))
	for _, root := range roots {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		candidate := filepath.Join(root, "nvm.exe")
		key := strings.ToLower(candidate)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		candidates = append(candidates, candidate)
	}
	return candidates
}

func refreshNVMEnvironment(nvmPath string) {
	nvmHome := filepath.Dir(nvmPath)
	nvmSymlink := resolveNVMSymlink()
	_ = os.Setenv("NVM_HOME", nvmHome)
	if nvmSymlink != "" {
		_ = os.Setenv("NVM_SYMLINK", nvmSymlink)
	}
	pathParts := []string{nvmSymlink, nvmHome, os.Getenv("PATH")}
	_ = os.Setenv("PATH", strings.Join(nonEmpty(pathParts), string(os.PathListSeparator)))
}

func ensureNVMSymlinkAvailable() error {
	return validateNVMSymlink(resolveNVMSymlink())
}

func validateNVMSymlink(symlink string) error {
	if symlink == "" {
		return errors.New("NVM_SYMLINK could not be resolved")
	}
	info, err := os.Lstat(symlink)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect NVM_SYMLINK %s: %w", symlink, err)
	}
	path, pathErr := windows.UTF16PtrFromString(symlink)
	if pathErr != nil {
		return fmt.Errorf("inspect NVM_SYMLINK path %s: %w", symlink, pathErr)
	}
	attributes, attributeErr := windows.GetFileAttributes(path)
	if info.Mode()&os.ModeSymlink != 0 || (attributeErr == nil && attributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0) {
		return nil
	}
	if attributeErr != nil {
		return fmt.Errorf("inspect NVM_SYMLINK attributes %s: %w", symlink, attributeErr)
	}
	return fmt.Errorf(
		"NVM_SYMLINK %s is an existing physical directory; uninstall the standalone Node.js installation and retry so NVM can create its managed symlink",
		symlink,
	)
}

func resolveNVMSymlink() string {
	return firstNonEmpty(
		os.Getenv("NVM_SYMLINK"),
		readWindowsEnvironment(registry.CURRENT_USER, `Environment`, "NVM_SYMLINK"),
		readWindowsEnvironment(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, "NVM_SYMLINK"),
		filepath.Join(os.Getenv("ProgramFiles"), "nodejs"),
	)
}

func readWindowsEnvironment(root registry.Key, path, name string) string {
	key, err := registry.OpenKey(root, path, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer key.Close()
	value, _, err := key.GetStringValue(name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func nonEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}
