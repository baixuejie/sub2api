//go:build windows

package dsh

import (
	"errors"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

type startLock struct {
	file       *os.File
	overlapped windows.Overlapped
}

func acquireStartLock(path string, timeout time.Duration) (*startLock, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	lock := &startLock{file: file}
	deadline := time.Now().Add(timeout)
	for {
		err = windows.LockFileEx(windows.Handle(file.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &lock.overlapped)
		if err == nil {
			return lock, nil
		}
		if !errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			file.Close()
			return nil, err
		}
		if time.Now().After(deadline) {
			file.Close()
			return nil, errors.New("timeout acquiring DSH startup lock")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (lock *startLock) release() {
	if lock == nil || lock.file == nil {
		return
	}
	_ = windows.UnlockFileEx(windows.Handle(lock.file.Fd()), 0, 1, 0, &lock.overlapped)
	_ = lock.file.Close()
}
