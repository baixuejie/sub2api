//go:build windows

package dsh

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

const (
	createNewProcessGroup = 0x00000200
	stillActive           = 259
)

func configureProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNewProcessGroup, HideWindow: true}
}

func terminateProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}

func terminatePID(pid int) error {
	if pid <= 0 {
		return nil
	}
	handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)
	return windows.TerminateProcess(handle, 1)
}

func processAlive(pid int) bool {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(handle)
	var exitCode uint32
	return windows.GetExitCodeProcess(handle, &exitCode) == nil && exitCode == stillActive
}

func samePath(a, b string) bool {
	left, err1 := canonicalWindowsPath(a)
	right, err2 := canonicalWindowsPath(b)
	return err1 == nil && err2 == nil && strings.EqualFold(left, right)
}

func canonicalWindowsPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		if os.IsNotExist(err) {
			return filepath.Clean(absolute), nil
		}
		return "", err
	}
	return filepath.Clean(resolved), nil
}
