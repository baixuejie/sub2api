//go:build !windows

package dsh

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

func configureProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func terminateProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	time.Sleep(500 * time.Millisecond)
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}

func terminatePID(pid int) error {
	if pid <= 0 {
		return nil
	}
	_ = syscall.Kill(-pid, syscall.SIGTERM)
	time.Sleep(500 * time.Millisecond)
	if !processAlive(pid) {
		return nil
	}
	return syscall.Kill(-pid, syscall.SIGKILL)
}

func processAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	return err == nil && process.Signal(syscall.Signal(0)) == nil
}

func samePath(a, b string) bool {
	left, err1 := filepath.Abs(a)
	right, err2 := filepath.Abs(b)
	return err1 == nil && err2 == nil && left == right
}
